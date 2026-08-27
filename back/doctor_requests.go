package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

func GetDoctorRequests(doctorID int, status, date, search string) ([]Request, error) {
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
		query += " AND r.status=$" + strconv.Itoa(n)
		args = append(args, status)
		n++
	}
	if date != "" {
		query += " AND r.appointment_date=$" + strconv.Itoa(n)
		args = append(args, date)
		n++
	}
	if strings.TrimSpace(search) != "" {
		query += " AND (p.name ILIKE $" + strconv.Itoa(n) + " OR p.phone ILIKE $" + strconv.Itoa(n) + " OR CAST(r.id AS TEXT) ILIKE $" + strconv.Itoa(n) + ")"
		args = append(args, "%"+strings.TrimSpace(search)+"%")
		n++
	}
	query += ` ORDER BY r.appointment_date NULLS LAST, r.appointment_time NULLS LAST, r.created_at DESC`
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Request, 0)
	for rows.Next() {
		var x Request
		if err := rows.Scan(&x.ID, &x.PatientID, &x.Name, &x.Phone, &x.Comment, &x.Services, &x.Price, &x.Source, &x.DoctorID, &x.DoctorName, &x.AppointmentDate, &x.AppointmentTime, &x.AppointmentComment, &x.Status, &x.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, x)
	}
	return result, rows.Err()
}

func GetDoctorRequest(doctorID, requestID int) (Request, error) {
	var x Request
	err := DB.QueryRow(`
SELECT r.id, r.patient_id, p.name, p.phone, p.comment, r.services, r.price, r.source,
       r.doctor_id, COALESCE(d.name,''), COALESCE(TO_CHAR(r.appointment_date,'YYYY-MM-DD'),''),
       COALESCE(TO_CHAR(r.appointment_time,'HH24:MI'),''), COALESCE(r.appointment_comment,''), r.status, r.created_at
FROM requests r JOIN patients p ON p.id=r.patient_id JOIN doctors d ON d.id=r.doctor_id
WHERE r.id=$1 AND r.doctor_id=$2`, requestID, doctorID).
		Scan(&x.ID, &x.PatientID, &x.Name, &x.Phone, &x.Comment, &x.Services, &x.Price, &x.Source, &x.DoctorID, &x.DoctorName, &x.AppointmentDate, &x.AppointmentTime, &x.AppointmentComment, &x.Status, &x.CreatedAt)
	return x, err
}

func UpdateDoctorRequestStatus(doctorID, requestID int, status string) (Request, error) {
	allowed := map[string]bool{"Подтверждена": true, "Отменена": true, "Завершена": true}
	if !allowed[status] {
		return Request{}, errors.New("недопустимый статус")
	}
	var current string
	err := DB.QueryRow(`SELECT status FROM requests WHERE id=$1 AND doctor_id=$2`, requestID, doctorID).Scan(&current)
	if err != nil {
		return Request{}, err
	}
	if status == "Подтверждена" && current != "Новая" {
		return Request{}, errors.New("подтвердить можно только новую заявку")
	}
	if status == "Завершена" {
		return Request{}, errors.New("завершение приёма выполняется через медицинскую карту")
	}
	if status == "Отменена" && current == "Завершена" {
		return Request{}, errors.New("завершённую заявку нельзя отменить")
	}
	_, err = DB.Exec(`UPDATE requests SET status=$1 WHERE id=$2 AND doctor_id=$3`, status, requestID, doctorID)
	if err != nil {
		return Request{}, err
	}
	return GetDoctorRequest(doctorID, requestID)
}

func DoctorRequestsHandler(w http.ResponseWriter, r *http.Request) {
	u, err := DoctorRequestGuard(r)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	switch r.Method {
	case http.MethodGet:
		q := r.URL.Query()
		list, err := GetDoctorRequests(u.DoctorID, q.Get("status"), q.Get("date"), q.Get("search"))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, list)
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func DoctorRequestByIDHandler(w http.ResponseWriter, r *http.Request) {
	u, err := DoctorRequestGuard(r)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/doctor/requests/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", 400)
		return
	}
	if r.Method == http.MethodGet {
		x, err := GetDoctorRequest(u.DoctorID, id)
		if err == sql.ErrNoRows {
			http.Error(w, "not found", 404)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, x)
		return
	}
	if r.Method == http.MethodPut {
		var body struct {
			Status     string `json:"status"`
			ServiceIDs []int  `json:"service_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", 400)
			return
		}
		if body.ServiceIDs != nil {
			var currentStatus string
			if err := DB.QueryRow(`SELECT status FROM requests WHERE id=$1 AND doctor_id=$2`, id, u.DoctorID).Scan(&currentStatus); err != nil {
				http.Error(w, "not found", 404)
				return
			}
			services, price, err := ResolveServices(body.ServiceIDs)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			if _, err = DB.Exec(`UPDATE requests SET services=$1, price=$2 WHERE id=$3 AND doctor_id=$4`, services, price, id, u.DoctorID); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			if body.Status == "" {
				x, err := GetDoctorRequest(u.DoctorID, id)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				writeJSON(w, x)
				return
			}
		}
		if body.Status == "" {
			body.Status = "Новая"
		}
		x, err := UpdateDoctorRequestStatus(u.DoctorID, id, body.Status)
		if err != nil {
			http.Error(w, err.Error(), 409)
			return
		}
		if body.Status != "" {
			go telegramStatusNotification(LoadConfig(), id, body.Status)
		}
		writeJSON(w, x)
		return
	}
	http.Error(w, "method not allowed", 405)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}

func GetDoctorPatients(doctorID int, search string) ([]Patient, error) {
	query := `SELECT p.id,p.name,p.phone,COUNT(r.id),COALESCE(SUM(r.price),0),
       COALESCE(MAX(r.appointment_date::timestamp + r.appointment_time), MAX(r.created_at)),
       COALESCE((SELECT services FROM requests rr
                 WHERE rr.patient_id=p.id AND rr.doctor_id=$1
                 ORDER BY rr.appointment_date DESC NULLS LAST, rr.appointment_time DESC NULLS LAST, rr.created_at DESC
                 LIMIT 1),'')
FROM patients p JOIN requests r ON r.patient_id=p.id AND r.doctor_id=$1 WHERE 1=1`
	args := []any{doctorID}
	if strings.TrimSpace(search) != "" {
		query += ` AND (p.name ILIKE $2 OR p.phone ILIKE $2)`
		args = append(args, "%"+strings.TrimSpace(search)+"%")
	}
	query += ` GROUP BY p.id,p.name,p.phone ORDER BY MAX(r.created_at) DESC`
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Patient, 0)
	for rows.Next() {
		var p Patient
		if err := rows.Scan(&p.ID, &p.Name, &p.Phone, &p.Visits, &p.TotalPrice, &p.LastVisit, &p.LastService); err != nil {
			return nil, err
		}
		if p.LastVisit.Valid {
			p.LastVisitStr = p.LastVisit.Time.Format("2006-01-02T15:04:05")
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func GetDoctorPatient(doctorID, patientID int) (PatientDetails, error) {
	var p PatientDetails
	err := DB.QueryRow(`SELECT p.id,p.name,p.phone,p.comment,COUNT(r.id),COALESCE(SUM(r.price),0),
       COALESCE(MAX(r.appointment_date::timestamp + r.appointment_time), MAX(r.created_at), NOW())
FROM patients p JOIN requests r ON r.patient_id=p.id AND r.doctor_id=$1
WHERE p.id=$2 GROUP BY p.id`, doctorID, patientID).Scan(&p.ID, &p.Name, &p.Phone, &p.Comment, &p.Visits, &p.TotalPrice, &p.LastVisit)
	if err != nil {
		return p, err
	}
	rows, err := DB.Query(`SELECT COALESCE(appointment_date::timestamp + appointment_time, created_at), services, price
FROM requests WHERE doctor_id=$1 AND patient_id=$2
ORDER BY appointment_date DESC NULLS LAST, appointment_time DESC NULLS LAST, created_at DESC`, doctorID, patientID)
	if err != nil {
		return p, err
	}
	defer rows.Close()
	for rows.Next() {
		var v Visit
		if err := rows.Scan(&v.Date, &v.Service, &v.Price); err != nil {
			return p, err
		}
		p.History = append(p.History, v)
	}
	if err := rows.Err(); err != nil {
		return p, err
	}
	medicalRows, err := DB.Query(`
        SELECT mr.id, mr.appointment_id, COALESCE(TO_CHAR(r.appointment_date,'YYYY-MM-DD'),''),
               COALESCE(TO_CHAR(r.appointment_time,'HH24:MI'),''), mr.complaints, mr.diagnosis,
               mr.treatment, mr.recommendations, mr.created_at
        FROM medical_records mr
        JOIN requests r ON r.id=mr.appointment_id
        WHERE mr.doctor_id=$1 AND mr.patient_id=$2
        ORDER BY r.appointment_date DESC NULLS LAST, r.appointment_time DESC NULLS LAST, mr.created_at DESC`, doctorID, patientID)
	if err != nil {
		return p, err
	}
	defer medicalRows.Close()
	for medicalRows.Next() {
		var m PatientMedicalRecord
		if err := medicalRows.Scan(&m.ID, &m.AppointmentID, &m.AppointmentDate, &m.AppointmentTime, &m.Complaints, &m.Diagnosis, &m.Treatment, &m.Recommendations, &m.CreatedAt); err != nil {
			return p, err
		}
		m.Files, err = getMedicalFiles(doctorID, m.ID)
		if err != nil {
			return p, err
		}
		p.MedicalRecords = append(p.MedicalRecords, m)
	}
	return p, medicalRows.Err()
}

func DoctorPatientsHandler(w http.ResponseWriter, r *http.Request) {
	u, err := DoctorRequestGuard(r)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 405)
		return
	}
	list, err := GetDoctorPatients(u.DoctorID, r.URL.Query().Get("search"))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, list)
}
func DoctorPatientByIDHandler(w http.ResponseWriter, r *http.Request) {
	u, err := DoctorRequestGuard(r)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/doctor/patients/"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", 400)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 405)
		return
	}
	p, err := GetDoctorPatient(u.DoctorID, id)
	if err == sql.ErrNoRows {
		http.Error(w, "not found", 404)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, p)
}
