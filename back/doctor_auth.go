package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	doctorSessionCookie = "doctor_session"
	pbkdf2Iterations    = 120000
	sessionTTL          = 24 * time.Hour
)

type DoctorAccountPayload struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	IsActive *bool  `json:"isActive"`
}

func GetDoctorAccount(doctorID int) (map[string]any, error) {
	var login string
	var active bool
	err := DB.QueryRow(`SELECT login, is_active FROM users WHERE doctor_id=$1 AND role='DOCTOR'`, doctorID).Scan(&login, &active)
	if err == sql.ErrNoRows {
		return map[string]any{"exists": false, "doctorId": doctorID}, nil
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{"exists": true, "doctorId": doctorID, "login": login, "isActive": active}, nil
}

func SaveDoctorAccount(doctorID int, payload DoctorAccountPayload) error {
	login := strings.TrimSpace(payload.Login)
	if login == "" {
		return errors.New("login is required")
	}
	if payload.Password == "" {
		var exists bool
		err := DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE doctor_id=$1 AND role='DOCTOR')`, doctorID).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("password is required for a new account")
		}
	}

	var existingID int
	err := DB.QueryRow(`SELECT id FROM users WHERE doctor_id=$1 AND role='DOCTOR'`, doctorID).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	active := true
	if payload.IsActive != nil {
		active = *payload.IsActive
	}

	if err == sql.ErrNoRows {
		hash, hashErr := HashPassword(payload.Password)
		if hashErr != nil {
			return hashErr
		}
		_, err = DB.Exec(`INSERT INTO users (login,password_hash,role,doctor_id,is_active) VALUES ($1,$2,'DOCTOR',$3,$4)`, login, hash, doctorID, active)
		return err
	}

	if payload.Password != "" {
		hash, hashErr := HashPassword(payload.Password)
		if hashErr != nil {
			return hashErr
		}
		_, err = DB.Exec(`UPDATE users SET login=$1,password_hash=$2,is_active=$3,updated_at=NOW() WHERE id=$4`, login, hash, active, existingID)
	} else {
		_, err = DB.Exec(`UPDATE users SET login=$1,is_active=$2,updated_at=NOW() WHERE id=$3`, login, active, existingID)
	}
	return err
}

type DoctorUser struct {
	ID       int
	Login    string
	DoctorID int
	Doctor   Doctor
}

func InitDoctorAuthTables() error {
	_, err := DB.Exec(`
CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
	login VARCHAR(100) NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	role VARCHAR(30) NOT NULL CHECK (role IN ('DOCTOR')),
	doctor_id INTEGER NOT NULL UNIQUE REFERENCES doctors(id) ON DELETE CASCADE,
	patient_id INTEGER,
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS auth_sessions (
	id BIGSERIAL PRIMARY KEY,
	token_hash CHAR(64) NOT NULL UNIQUE,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	role VARCHAR(30) NOT NULL,
	expires_at TIMESTAMP NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auth_sessions_token_hash ON auth_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_user_id ON auth_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires_at ON auth_sessions(expires_at);
`)
	if err != nil {
		return err
	}

	// Demo accounts are created only when no account exists for the doctor.
	// Change these passwords before using the system outside a local/demo environment.
	for doctorID := 1; doctorID <= 7; doctorID++ {
		login := fmt.Sprintf("doctor%d", doctorID)
		password := fmt.Sprintf("Doctor%d!2026", doctorID)

		var exists bool
		if err := DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE doctor_id=$1)`, doctorID).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}

		hash, err := HashPassword(password)
		if err != nil {
			return err
		}
		if _, err := DB.Exec(`
INSERT INTO users (login,password_hash,role,doctor_id,is_active)
VALUES ($1,$2,'DOCTOR',$3,TRUE)
`, login, hash, doctorID); err != nil {
			return err
		}
	}
	return nil
}

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password is empty")
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	dk := pbkdf2SHA256([]byte(password), salt, pbkdf2Iterations, 32)
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s",
		pbkdf2Iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(dk),
	), nil
}

func CheckPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}
	var iterations int
	if _, err := fmt.Sscanf(parts[1], "%d", &iterations); err != nil || iterations < 1 {
		return false
	}
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[2])
	expected, err2 := base64.RawStdEncoding.DecodeString(parts[3])
	if err1 != nil || err2 != nil || len(expected) == 0 {
		return false
	}
	actual := pbkdf2SHA256([]byte(password), salt, iterations, len(expected))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

func pbkdf2SHA256(password, salt []byte, iterations, keyLen int) []byte {
	const hashLen = 32
	blocks := (keyLen + hashLen - 1) / hashLen
	out := make([]byte, 0, blocks*hashLen)

	for block := 1; block <= blocks; block++ {
		mac := hmac.New(sha256.New, password)
		mac.Write(salt)
		mac.Write([]byte{byte(block >> 24), byte(block >> 16), byte(block >> 8), byte(block)})
		u := mac.Sum(nil)
		t := append([]byte(nil), u...)

		for i := 1; i < iterations; i++ {
			mac = hmac.New(sha256.New, password)
			mac.Write(u)
			u = mac.Sum(nil)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		out = append(out, t...)
	}
	return out[:keyLen]
}

func DoctorLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	user, err := authenticateDoctor(data.Login, data.Password)
	if err != nil {
		http.Error(w, "Неверный логин или пароль", http.StatusUnauthorized)
		return
	}

	token, err := createDoctorSession(user.ID)
	if err != nil {
		http.Error(w, "Не удалось создать сессию", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     doctorSessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionTTL / time.Second),
		Secure:   isProduction(),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"doctor": map[string]any{
			"id":   user.Doctor.ID,
			"name": user.Doctor.Name,
		},
	})
}

func DoctorLogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if cookie, err := r.Cookie(doctorSessionCookie); err == nil {
		_, _ = DB.Exec(`DELETE FROM auth_sessions WHERE token_hash=$1`, hashSessionToken(cookie.Value))
	}
	http.SetCookie(w, &http.Cookie{Name: doctorSessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	w.WriteHeader(http.StatusNoContent)
}

func authenticateDoctor(login, password string) (DoctorUser, error) {
	var u DoctorUser
	var doctorID int
	var hash string
	var active bool

	err := DB.QueryRow(`
SELECT id, password_hash, doctor_id, is_active
FROM users
WHERE login=$1 AND role='DOCTOR'
`, strings.TrimSpace(login)).Scan(&u.ID, &hash, &doctorID, &active)
	if err != nil || !active {
		return u, errors.New("invalid credentials")
	}
	// Backward compatibility: older doctor accounts may contain the legacy
	// plain-text password in password_hash. Accept it once and immediately
	// replace it with a PBKDF2 hash.
	if !CheckPassword(password, hash) {
		if subtle.ConstantTimeCompare([]byte(hash), []byte(password)) != 1 {
			return u, errors.New("invalid credentials")
		}
		if upgraded, hashErr := HashPassword(password); hashErr == nil {
			_, _ = DB.Exec(`UPDATE users SET password_hash=$1, updated_at=NOW() WHERE id=$2`, upgraded, u.ID)
		}
	}

	u.DoctorID = doctorID
	u.Doctor, err = GetDoctor(doctorID)
	if err != nil || !u.Doctor.IsActive {
		return DoctorUser{}, errors.New("doctor is inactive")
	}
	return u, nil
}

func createDoctorSession(userID int) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	_, err := DB.Exec(`
INSERT INTO auth_sessions (token_hash,user_id,role,expires_at)
VALUES ($1,$2,'DOCTOR',$3)
`, hashSessionToken(token), userID, time.Now().Add(sessionTTL))
	if err != nil {
		return "", err
	}
	return token, nil
}

func hashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum[:])
}

func currentDoctorUser(r *http.Request) (DoctorUser, error) {
	cookie, err := r.Cookie(doctorSessionCookie)
	if err != nil || cookie.Value == "" {
		return DoctorUser{}, errors.New("unauthorized")
	}

	var u DoctorUser
	var doctorID int
	var active bool
	err = DB.QueryRow(`
SELECT u.id, u.doctor_id, u.is_active
FROM auth_sessions s
JOIN users u ON u.id=s.user_id
JOIN doctors d ON d.id=u.doctor_id
WHERE s.token_hash=$1
  AND s.role='DOCTOR'
  AND s.expires_at > NOW()
  AND u.role='DOCTOR'
  AND u.is_active=TRUE
  AND d.is_active=TRUE
`, hashSessionToken(cookie.Value)).Scan(&u.ID, &doctorID, &active)
	if err != nil || !active {
		return DoctorUser{}, errors.New("unauthorized")
	}

	u.DoctorID = doctorID
	u.Doctor, err = GetDoctor(doctorID)
	if err != nil {
		return DoctorUser{}, err
	}
	return u, nil
}

func DoctorAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := currentDoctorUser(r); err != nil {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/?account=doctor", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func DoctorMeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	u, err := currentDoctorUser(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":       u.Doctor.ID,
		"doctorId": u.Doctor.ID,
		"name":     u.Doctor.Name,
		"position": u.Doctor.Position,
		"role":     "DOCTOR",
	})
}

func DoctorRequestGuard(r *http.Request) (DoctorUser, error) {
	return currentDoctorUser(r)
}
