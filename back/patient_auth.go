package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const patientSessionCookie = "patient_session"

type PatientRegisterPayload struct {
	Name           string `json:"name"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	BirthDate      string `json:"birthDate"`
	CaptchaToken   string `json:"captchaToken"`
	CaptchaAnswer  string `json:"captchaAnswer"`
	Password       string `json:"password"`
	PrivacyConsent bool   `json:"privacyConsent"`
}

type PatientUser struct {
	ID        int
	PatientID int
	Patient   Patient
}

func InitPatientAuthSchema() error {
	_, err := DB.Exec(`
ALTER TABLE patients ADD COLUMN IF NOT EXISTS email VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE patients ADD COLUMN IF NOT EXISTS birth_date DATE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS privacy_consent_at TIMESTAMP;
CREATE UNIQUE INDEX IF NOT EXISTS ux_patients_email_nonempty
ON patients(email) WHERE email <> '';

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ALTER COLUMN doctor_id DROP NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('ADMIN','DOCTOR','PATIENT'));
CREATE UNIQUE INDEX IF NOT EXISTS ux_users_patient_id ON users(patient_id) WHERE patient_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS patient_payments (
	id SERIAL PRIMARY KEY,
	patient_id INTEGER NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
	request_id INTEGER REFERENCES requests(id) ON DELETE SET NULL,
	amount INTEGER NOT NULL DEFAULT 0,
	status VARCHAR(50) NOT NULL DEFAULT 'К оплате',
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	paid_at TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_patient_payments_patient ON patient_payments(patient_id);
CREATE UNIQUE INDEX IF NOT EXISTS ux_patient_paid_request ON patient_payments(patient_id, request_id) WHERE status = 'Оплачено' AND request_id IS NOT NULL;
CREATE TABLE IF NOT EXISTS patient_notifications (
	id SERIAL PRIMARY KEY,
	patient_id INTEGER NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
	title VARCHAR(255) NOT NULL,
	message TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	is_read BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_patient_notifications_patient ON patient_notifications(patient_id, created_at DESC);
`)
	return err
}

func PatientRegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var p PatientRegisterPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	if !p.PrivacyConsent {
		http.Error(w, "Необходимо согласиться на обработку персональных данных и с Политикой конфиденциальности", 400)
		return
	}
	if err := enforceRateLimit("register:"+clientIP(r), 5, time.Hour); err != nil {
		http.Error(w, err.Error(), http.StatusTooManyRequests)
		return
	}
	if err := validateCaptcha(r, p.CaptchaToken, p.CaptchaAnswer); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	p.Name = strings.TrimSpace(p.Name)
	p.Phone = strings.TrimSpace(p.Phone)
	p.Email = strings.TrimSpace(strings.ToLower(p.Email))
	if p.Name == "" || p.Phone == "" || len(p.Password) < 6 {
		http.Error(w, "Заполните имя, телефон и пароль минимум из 6 символов", 400)
		return
	}
	var patientID int
	err := DB.QueryRow(`SELECT id FROM patients WHERE phone=$1`, p.Phone).Scan(&patientID)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "database error", 500)
		return
	}
	if err == sql.ErrNoRows {
		// If the email already belongs to another patient, registration is rejected.
		if p.Email != "" {
			var emailPatientID int
			errEmail := DB.QueryRow(`SELECT id FROM patients WHERE LOWER(email)=LOWER($1)`, p.Email).Scan(&emailPatientID)
			if errEmail == nil {
				http.Error(w, "Этот email уже используется", 409)
				return
			}
			if errEmail != sql.ErrNoRows {
				http.Error(w, "database error", 500)
				return
			}
		}

		if p.BirthDate != "" {
			err = DB.QueryRow(`INSERT INTO patients(name,phone,comment,email,birth_date) VALUES($1,$2,'',$3,$4) RETURNING id`, p.Name, p.Phone, p.Email, p.BirthDate).Scan(&patientID)
		} else {
			err = DB.QueryRow(`INSERT INTO patients(name,phone,comment,email) VALUES($1,$2,'',$3) RETURNING id`, p.Name, p.Phone, p.Email).Scan(&patientID)
		}
		if err != nil {
			http.Error(w, "Не удалось создать пациента", 500)
			return
		}
	} else {
		// Existing guest patient may be converted into an account.
		if p.Email != "" {
			var emailPatientID int
			errEmail := DB.QueryRow(`SELECT id FROM patients WHERE LOWER(email)=LOWER($1)`, p.Email).Scan(&emailPatientID)
			if errEmail == nil && emailPatientID != patientID {
				http.Error(w, "Этот email уже используется", 409)
				return
			}
			if errEmail != nil && errEmail != sql.ErrNoRows {
				http.Error(w, "database error", 500)
				return
			}
		}
		_, _ = DB.Exec(`UPDATE patients SET name=$1,email=CASE WHEN $2='' THEN email ELSE $2 END,birth_date=CASE WHEN $3='' THEN birth_date ELSE $3::date END WHERE id=$4`, p.Name, p.Email, p.BirthDate, patientID)
	}
	var existsAccount bool
	if err := DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE patient_id=$1 AND role='PATIENT')`, patientID).Scan(&existsAccount); err != nil {
		http.Error(w, "database error", 500)
		return
	}
	if existsAccount {
		http.Error(w, "Для этого пациента аккаунт уже существует", 409)
		return
	}
	hash, err := HashPassword(p.Password)
	if err != nil {
		http.Error(w, "Не удалось создать пароль", 500)
		return
	}
	login := "patient_" + strconv.Itoa(patientID)
	if _, err = DB.Exec(`INSERT INTO users(login,password_hash,role,patient_id,is_active,privacy_consent_at) VALUES($1,$2,'PATIENT',$3,TRUE,NOW())`, login, hash, patientID); err != nil {
		http.Error(w, "Не удалось создать аккаунт", 500)
		return
	}
	token, err := createPatientSession(patientID)
	if err != nil {
		http.Error(w, "Не удалось создать сессию", 500)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: patientSessionCookie, Value: token, Path: "/", HttpOnly: true, Secure: isProduction(), SameSite: http.SameSiteLaxMode, MaxAge: int(sessionTTL / time.Second)})
	writeJSON(w, map[string]any{"status": "ok", "patientId": patientID})
}

func PatientLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var d LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	u, err := authenticatePatient(d.Login, d.Password)
	if err != nil {
		http.Error(w, "Неверный телефон/email или пароль", 401)
		return
	}
	token, err := createPatientSession(u.PatientID)
	if err != nil {
		http.Error(w, "Не удалось создать сессию", 500)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: patientSessionCookie, Value: token, Path: "/", HttpOnly: true, Secure: isProduction(), SameSite: http.SameSiteLaxMode, MaxAge: int(sessionTTL / time.Second)})
	writeJSON(w, map[string]any{"status": "ok", "patientId": u.PatientID})
}

func PatientLogoutHandler(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(patientSessionCookie); err == nil {
		_, _ = DB.Exec(`DELETE FROM auth_sessions WHERE token_hash=$1`, hashSessionToken(c.Value))
	}
	http.SetCookie(w, &http.Cookie{Name: patientSessionCookie, Value: "", Path: "/", HttpOnly: true, MaxAge: -1, SameSite: http.SameSiteLaxMode})
	w.WriteHeader(204)
}

func authenticatePatient(identifier, password string) (PatientUser, error) {
	var u PatientUser
	var hash string
	var active bool
	err := DB.QueryRow(`
SELECT u.id,u.patient_id,u.password_hash,u.is_active
FROM users u JOIN patients p ON p.id=u.patient_id
WHERE u.role='PATIENT' AND u.is_active=TRUE
AND (u.login=$1 OR p.phone=$1 OR LOWER(p.email)=LOWER($1))`, strings.TrimSpace(identifier)).Scan(&u.ID, &u.PatientID, &hash, &active)
	if err != nil || !active || !CheckPassword(password, hash) {
		return u, errors.New("invalid credentials")
	}
	err = DB.QueryRow(`SELECT id,name,phone,comment FROM patients WHERE id=$1`, u.PatientID).Scan(&u.Patient.ID, &u.Patient.Name, &u.Patient.Phone, &u.Patient.Comment)
	if err != nil {
		return u, err
	}
	return u, nil
}

func createPatientSession(patientID int) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	_, err := DB.Exec(`INSERT INTO auth_sessions(token_hash,user_id,role,expires_at) SELECT $1,id,'PATIENT',$2 FROM users WHERE patient_id=$3 AND role='PATIENT'`, hashSessionToken(token), time.Now().Add(sessionTTL), patientID)
	return token, err
}

func currentPatientUser(r *http.Request) (PatientUser, error) {
	c, err := r.Cookie(patientSessionCookie)
	if err != nil || c.Value == "" {
		return PatientUser{}, errors.New("unauthorized")
	}
	var u PatientUser
	var active bool
	err = DB.QueryRow(`
SELECT u.id,u.patient_id,u.is_active
FROM auth_sessions s JOIN users u ON u.id=s.user_id JOIN patients p ON p.id=u.patient_id
WHERE s.token_hash=$1 AND s.role='PATIENT' AND s.expires_at>NOW() AND u.role='PATIENT' AND u.is_active=TRUE`, hashSessionToken(c.Value)).Scan(&u.ID, &u.PatientID, &active)
	if err != nil || !active {
		return PatientUser{}, errors.New("unauthorized")
	}
	return u, nil
}

func PatientAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := currentPatientUser(r); err != nil {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.Error(w, "unauthorized", 401)
				return
			}
			http.Redirect(w, r, "/?account=patient", 303)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func PatientMeHandler(w http.ResponseWriter, r *http.Request) {
	u, err := currentPatientUser(r)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	var name, phone, comment, email string
	var birthDate *time.Time
	err = DB.QueryRow(`SELECT name,phone,comment,email,birth_date FROM patients WHERE id=$1`, u.PatientID).Scan(&name, &phone, &comment, &email, &birthDate)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	var bd string
	if birthDate != nil {
		bd = birthDate.Format("2006-01-02")
	}
	writeJSON(w, map[string]any{"id": u.PatientID, "name": name, "phone": phone, "email": email, "birthDate": bd, "comment": comment})
}
