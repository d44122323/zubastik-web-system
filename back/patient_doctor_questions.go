package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type DoctorQuestion struct {
	ID            int        `json:"id"`
	AppointmentID int        `json:"appointmentId"`
	DoctorID      int        `json:"doctorId"`
	DoctorName    string     `json:"doctor"`
	PatientID     int        `json:"patientId"`
	PatientName   string     `json:"patient"`
	Service       string     `json:"service"`
	Question      string     `json:"question"`
	Answer        string     `json:"answer"`
	CreatedAt     time.Time  `json:"createdAt"`
	AnsweredAt    *time.Time `json:"answeredAt,omitempty"`
}

func InitDoctorQuestionsSchema() error {
	_, err := DB.Exec(`
CREATE TABLE IF NOT EXISTS doctor_questions (
    id SERIAL PRIMARY KEY,
    appointment_id INTEGER NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
    patient_id INTEGER NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    doctor_id INTEGER NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    question TEXT NOT NULL,
    answer TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    answered_at TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_doctor_questions_patient ON doctor_questions(patient_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_doctor_questions_doctor ON doctor_questions(doctor_id, created_at DESC);
`)
	return err
}

func PatientDoctorQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	u, err := currentPatientUser(r)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	switch r.Method {
	case http.MethodGet:
		rows, err := DB.Query(`
SELECT q.id,q.appointment_id,q.doctor_id,COALESCE(d.name,''),q.patient_id,COALESCE(p.name,''),COALESCE(r.services,''),q.question,q.answer,q.created_at,q.answered_at
FROM doctor_questions q
JOIN doctors d ON d.id=q.doctor_id
JOIN patients p ON p.id=q.patient_id
JOIN requests r ON r.id=q.appointment_id
WHERE q.patient_id=$1 ORDER BY q.created_at DESC`, u.PatientID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		out := []DoctorQuestion{}
		for rows.Next() {
			var q DoctorQuestion
			if err := rows.Scan(&q.ID, &q.AppointmentID, &q.DoctorID, &q.DoctorName, &q.PatientID, &q.PatientName, &q.Service, &q.Question, &q.Answer, &q.CreatedAt, &q.AnsweredAt); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			out = append(out, q)
		}
		writeJSON(w, out)
	case http.MethodPost:
		var body struct {
			AppointmentID int    `json:"appointmentId"`
			Question      string `json:"question"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", 400)
			return
		}
		body.Question = strings.TrimSpace(body.Question)
		if body.AppointmentID <= 0 || body.Question == "" {
			http.Error(w, "Укажите вопрос и посещение", 400)
			return
		}
		if len([]rune(body.Question)) > 2000 {
			http.Error(w, "Вопрос слишком длинный", 400)
			return
		}
		var doctorID int
		var status string
		var service string
		err = DB.QueryRow(`SELECT doctor_id,status,COALESCE(services,'') FROM requests WHERE id=$1 AND patient_id=$2`, body.AppointmentID, u.PatientID).Scan(&doctorID, &status, &service)
		if err == sql.ErrNoRows {
			http.Error(w, "Посещение не найдено", 404)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if doctorID <= 0 || status != "Завершена" {
			http.Error(w, "Задать вопрос врачу можно после завершённого приёма", 409)
			return
		}
		var id int
		err = DB.QueryRow(`INSERT INTO doctor_questions(appointment_id,patient_id,doctor_id,question) VALUES($1,$2,$3,$4) RETURNING id`, body.AppointmentID, u.PatientID, doctorID, body.Question).Scan(&id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, map[string]any{"status": "ok", "id": id, "service": service})
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func DoctorQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	u, err := DoctorRequestGuard(r)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	switch r.Method {
	case http.MethodGet:
		rows, err := DB.Query(`
SELECT q.id,q.appointment_id,q.doctor_id,COALESCE(d.name,''),q.patient_id,COALESCE(p.name,''),COALESCE(r.services,''),q.question,q.answer,q.created_at,q.answered_at
FROM doctor_questions q
JOIN doctors d ON d.id=q.doctor_id
JOIN patients p ON p.id=q.patient_id
JOIN requests r ON r.id=q.appointment_id
WHERE q.doctor_id=$1 ORDER BY q.created_at DESC`, u.DoctorID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		out := []DoctorQuestion{}
		for rows.Next() {
			var q DoctorQuestion
			if err := rows.Scan(&q.ID, &q.AppointmentID, &q.DoctorID, &q.DoctorName, &q.PatientID, &q.PatientName, &q.Service, &q.Question, &q.Answer, &q.CreatedAt, &q.AnsweredAt); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			out = append(out, q)
		}
		writeJSON(w, out)
	case http.MethodPut:
		id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/doctor/questions/"))
		if err != nil || id <= 0 {
			http.Error(w, "invalid id", 400)
			return
		}
		var body struct {
			Answer string `json:"answer"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", 400)
			return
		}
		body.Answer = strings.TrimSpace(body.Answer)
		if body.Answer == "" {
			http.Error(w, "Введите ответ", 400)
			return
		}
		if len([]rune(body.Answer)) > 4000 {
			http.Error(w, "Ответ слишком длинный", 400)
			return
		}
		var patientID int
		err = DB.QueryRow(`UPDATE doctor_questions SET answer=$1,answered_at=NOW() WHERE id=$2 AND doctor_id=$3 RETURNING patient_id`, body.Answer, id, u.DoctorID).Scan(&patientID)
		if err == sql.ErrNoRows {
			http.Error(w, "Вопрос не найден", 404)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		_, _ = DB.Exec(`INSERT INTO patient_notifications(patient_id,title,message) VALUES($1,$2,$3)`, patientID, "Врач ответил на ваш вопрос", "Врач ответил на вопрос в личном кабинете.")
		writeJSON(w, map[string]string{"status": "ok"})
	default:
		http.Error(w, "method not allowed", 405)
	}
}
