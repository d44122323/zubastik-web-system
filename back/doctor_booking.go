package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func InitDoctorBookingSchema() error {
	_, err := DB.Exec(`
        ALTER TABLE requests ADD COLUMN IF NOT EXISTS appointment_comment TEXT;
    `)
	return err
}

func CreateDoctorAppointment(doctorID, patientID int, service, dateStr, timeStr, comment string, price int) (Request, error) {
	if doctorID <= 0 || patientID <= 0 {
		return Request{}, errors.New("некорректный врач или пациент")
	}
	service = strings.TrimSpace(service)
	if service == "" {
		return Request{}, errors.New("укажите услугу")
	}
	dateStr = strings.TrimSpace(dateStr)
	timeStr = strings.TrimSpace(timeStr)
	if err := ValidateAppointment(doctorID, dateStr, timeStr); err != nil {
		return Request{}, err
	}

	var exists bool
	if err := DB.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM requests
            WHERE patient_id=$1 AND doctor_id=$2
        )`, patientID, doctorID).Scan(&exists); err != nil {
		return Request{}, err
	}
	if !exists {
		return Request{}, errors.New("этот пациент не относится к текущему врачу")
	}

	if price < 0 {
		price = 0
	}
	var id int
	err := DB.QueryRow(`
        INSERT INTO requests
        (patient_id, services, price, source, doctor_id, appointment_date, appointment_time, appointment_comment, status, created_at)
        VALUES ($1,$2,$3,'doctor',$4,$5,$6,$7,'Подтверждена',NOW())
        RETURNING id
    `, patientID, service, price, doctorID, dateStr, timeStr, strings.TrimSpace(comment)).Scan(&id)
	if err != nil {
		return Request{}, err
	}
	return GetDoctorRequest(doctorID, id)
}

func DoctorCreateAppointmentHandler(w http.ResponseWriter, r *http.Request) {
	u, err := DoctorRequestGuard(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		PatientID int    `json:"patient_id"`
		Service   string `json:"service"`
		Date      string `json:"date"`
		Time      string `json:"time"`
		Comment   string `json:"comment"`
		Price     int    `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	result, err := CreateDoctorAppointment(u.DoctorID, body.PatientID, body.Service, body.Date, body.Time, body.Comment, body.Price)
	if err != nil {
		code := http.StatusConflict
		if strings.Contains(err.Error(), "пациент") || strings.Contains(err.Error(), "услуг") {
			code = http.StatusBadRequest
		}
		http.Error(w, err.Error(), code)
		return
	}
	writeJSON(w, result)
}

func GetDoctorPatientForBooking(doctorID, patientID int) (Patient, error) {
	var p Patient
	err := DB.QueryRow(`
        SELECT p.id,p.name,p.phone,COUNT(r.id),COALESCE(SUM(r.price),0),MAX(r.created_at),
               COALESCE((SELECT services FROM requests rr WHERE rr.patient_id=p.id AND rr.doctor_id=$1 ORDER BY rr.created_at DESC LIMIT 1),'')
        FROM patients p
        JOIN requests r ON r.patient_id=p.id AND r.doctor_id=$1
        WHERE p.id=$2
        GROUP BY p.id,p.name,p.phone
    `, doctorID, patientID).Scan(&p.ID, &p.Name, &p.Phone, &p.Visits, &p.TotalPrice, &p.LastVisit, &p.LastService)
	if err == sql.ErrNoRows {
		return p, fmt.Errorf("пациент не найден")
	}
	return p, err
}

func GetDoctorAppointments(doctorID int, period, status, date, search string) ([]Request, error) {
	query := `
SELECT r.id, r.patient_id, p.name, p.phone, p.comment, r.services, r.price, r.source,
       r.doctor_id, COALESCE(d.name,''), COALESCE(TO_CHAR(r.appointment_date,'YYYY-MM-DD'),''),
       COALESCE(TO_CHAR(r.appointment_time,'HH24:MI'),''), COALESCE(r.appointment_comment,''), r.status, r.created_at
FROM requests r
JOIN patients p ON p.id=r.patient_id
JOIN doctors d ON d.id=r.doctor_id
WHERE r.doctor_id=$1`
	args := []any{doctorID}
	n := 2
	if status != "" {
		query += " AND r.status=$" + fmt.Sprint(n)
		args = append(args, status)
		n++
	}
	if date != "" {
		query += " AND r.appointment_date=$" + fmt.Sprint(n)
		args = append(args, date)
		n++
	}
	today := time.Now().Format("2006-01-02")
	if period == "today" {
		query += " AND r.appointment_date=$" + fmt.Sprint(n)
		args = append(args, today)
		n++
	} else if period == "upcoming" {
		query += " AND r.appointment_date >= $" + fmt.Sprint(n)
		args = append(args, today)
		n++
	}
	if strings.TrimSpace(search) != "" {
		query += " AND (p.name ILIKE $" + fmt.Sprint(n) + " OR p.phone ILIKE $" + fmt.Sprint(n) + " OR CAST(r.id AS TEXT) ILIKE $" + fmt.Sprint(n) + ")"
		args = append(args, "%"+strings.TrimSpace(search)+"%")
		n++
	}
	query += ` ORDER BY r.appointment_date NULLS LAST, r.appointment_time NULLS LAST, r.created_at DESC`
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Request, 0)
	for rows.Next() {
		var x Request
		if err := rows.Scan(&x.ID, &x.PatientID, &x.Name, &x.Phone, &x.Comment, &x.Services, &x.Price, &x.Source, &x.DoctorID, &x.DoctorName, &x.AppointmentDate, &x.AppointmentTime, &x.AppointmentComment, &x.Status, &x.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func DoctorAppointmentsHandler(w http.ResponseWriter, r *http.Request) {
	u, err := DoctorRequestGuard(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	items, err := GetDoctorAppointments(u.DoctorID, r.URL.Query().Get("period"), r.URL.Query().Get("status"), r.URL.Query().Get("date"), r.URL.Query().Get("search"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, items)
}
