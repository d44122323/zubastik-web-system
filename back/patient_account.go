package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func PatientProfileHandler(w http.ResponseWriter, r *http.Request) {
	u, err := currentPatientUser(r)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	if r.Method == http.MethodGet {
		PatientMeHandler(w, r)
		return
	}
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", 405)
		return
	}
	var p struct {
		Name      string `json:"name"`
		Phone     string `json:"phone"`
		Email     string `json:"email"`
		BirthDate string `json:"birthDate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	p.Name = strings.TrimSpace(p.Name)
	p.Phone = strings.TrimSpace(p.Phone)
	p.Email = strings.TrimSpace(strings.ToLower(p.Email))
	if p.Name == "" || p.Phone == "" {
		http.Error(w, "Имя и телефон обязательны", 400)
		return
	}
	var duplicate int
	if p.Email != "" {
		err = DB.QueryRow(`SELECT id FROM patients WHERE email=$1 AND id<>$2`, p.Email, u.PatientID).Scan(&duplicate)
		if err == nil {
			http.Error(w, "Этот email уже используется", 409)
			return
		}
	}
	_, err = DB.Exec(`UPDATE patients SET name=$1,phone=$2,email=$3,birth_date=CASE WHEN $4='' THEN NULL ELSE $4::date END WHERE id=$5`, p.Name, p.Phone, p.Email, p.BirthDate, u.PatientID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	PatientMeHandler(w, r)
}

func PatientAppointmentsHandler(w http.ResponseWriter, r *http.Request) {
	u, err := currentPatientUser(r)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	if r.Method == http.MethodGet {
		rows, err := DB.Query(`SELECT r.id,r.services,r.price,r.status,COALESCE(r.appointment_date::text,''),COALESCE(TO_CHAR(r.appointment_time,'HH24:MI'),''),COALESCE(d.id,0),COALESCE(d.name,'Врач не назначен'),COALESCE(r.appointment_comment,'') FROM requests r LEFT JOIN doctors d ON d.id=r.doctor_id WHERE r.patient_id=$1 ORDER BY r.appointment_date DESC NULLS LAST,r.appointment_time DESC NULLS LAST,r.created_at DESC`, u.PatientID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		out := []map[string]any{}
		for rows.Next() {
			var id, price, did int
			var service, status, date, t, doctor, comment string
			if err := rows.Scan(&id, &service, &price, &status, &date, &t, &did, &doctor, &comment); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			out = append(out, map[string]any{"id": id, "service": service, "price": price, "status": status, "date": date, "time": t, "doctorId": did, "doctor": doctor, "comment": comment})
		}
		writeJSON(w, out)
		return
	}
	if r.Method == http.MethodPut {
		id, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/patient/appointments/"))
		var body struct {
			Action string `json:"action"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", 400)
			return
		}
		if body.Action != "cancel" {
			http.Error(w, "Поддерживается отмена записи", 400)
			return
		}
		res, err := DB.Exec(`UPDATE requests SET status='Отменено пациентом' WHERE id=$1 AND patient_id=$2 AND status NOT IN ('Завершена','Отменена','Отменено','Отменено пациентом')`, id, u.PatientID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			http.Error(w, "Запись нельзя отменить", 409)
			return
		}
		_, _ = DB.Exec(`INSERT INTO patient_notifications(patient_id,title,message) VALUES($1,$2,$3)`, u.PatientID, "Запись отменена", "Вы отменили свою запись.")
		writeJSON(w, map[string]string{"status": "ok"})
		return
	}
	http.Error(w, "method not allowed", 405)
}

func PatientMedicalHandler(w http.ResponseWriter, r *http.Request) {
	u, err := currentPatientUser(r)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	rows, err := DB.Query(`SELECT mr.id,mr.appointment_id,COALESCE(r.appointment_date::text,''),COALESCE(TO_CHAR(r.appointment_time,'HH24:MI'),''),COALESCE(d.name,''),mr.complaints,mr.diagnosis,mr.treatment,mr.recommendations,mr.created_at FROM medical_records mr JOIN requests r ON r.id=mr.appointment_id LEFT JOIN doctors d ON d.id=mr.doctor_id WHERE mr.patient_id=$1 ORDER BY r.appointment_date DESC NULLS LAST,r.appointment_time DESC NULLS LAST,mr.created_at DESC`, u.PatientID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, aid int
		var date, t, doctor, complaints, diagnosis, treatment, recs string
		var created interface{}
		if err := rows.Scan(&id, &aid, &date, &t, &doctor, &complaints, &diagnosis, &treatment, &recs, &created); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		out = append(out, map[string]any{"id": id, "appointmentId": aid, "date": date, "time": t, "doctor": doctor, "complaints": complaints, "diagnosis": diagnosis, "treatment": treatment, "recommendations": recs})
	}
	writeJSON(w, out)
}

func PatientDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	u, err := currentPatientUser(r)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	rows, err := DB.Query(`SELECT f.id,f.file_name,f.file_type,f.file_size,COALESCE(r.appointment_date::text,''),COALESCE(TO_CHAR(r.appointment_time,'HH24:MI'),''),mr.id FROM medical_files f JOIN medical_records mr ON mr.id=f.medical_record_id JOIN requests r ON r.id=mr.appointment_id WHERE mr.patient_id=$1 ORDER BY f.created_at DESC`, u.PatientID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, mrid, size int
		var name, typ, date, t string
		if err := rows.Scan(&id, &name, &typ, &size, &date, &t, &mrid); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		out = append(out, map[string]any{"id": id, "name": name, "type": typ, "size": size, "date": date, "time": t, "recordId": mrid})
	}
	writeJSON(w, out)
}

func PatientMedicalFileHandler(w http.ResponseWriter, r *http.Request) {
	u, err := currentPatientUser(r)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/patient/files/"))
	if err != nil {
		http.Error(w, "invalid id", 400)
		return
	}
	var name, path, typ string
	err = DB.QueryRow(`SELECT f.file_name,f.file_path,f.file_type FROM medical_files f JOIN medical_records mr ON mr.id=f.medical_record_id WHERE f.id=$1 AND mr.patient_id=$2`, id, u.PatientID).Scan(&name, &path, &typ)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		http.Error(w, "file unavailable", 404)
		return
	}
	defer f.Close()
	if typ != "" {
		w.Header().Set("Content-Type", typ)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Header().Set("Content-Disposition", `inline; filename="`+strings.ReplaceAll(name, `"`, ``)+`"`)
	http.ServeContent(w, r, name, time.Time{}, f)
}

func PatientNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	u, err := currentPatientUser(r)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	if r.Method == http.MethodPut {
		if _, err := DB.Exec(`UPDATE patient_notifications SET is_read=TRUE WHERE patient_id=$1`, u.PatientID); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, map[string]any{"status": "ok"})
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 405)
		return
	}
	rows, err := DB.Query(`
SELECT id,title,message,created_at,is_read FROM patient_notifications WHERE patient_id=$1
UNION ALL
SELECT r.id,'Статус записи',
       'Статус записи: ' || r.status || CASE WHEN r.appointment_date IS NOT NULL THEN ' · ' || r.appointment_date::text || ' ' || COALESCE(TO_CHAR(r.appointment_time,'HH24:MI'),'') ELSE '' END,
       r.created_at,TRUE
FROM requests r WHERE r.patient_id=$1
ORDER BY created_at DESC LIMIT 30`, u.PatientID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int
		var title, message string
		var created interface{}
		var read bool
		if err := rows.Scan(&id, &title, &message, &created, &read); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		out = append(out, map[string]any{"id": id, "title": title, "text": message, "createdAt": created, "read": read})
	}
	writeJSON(w, out)
}

func PatientPaymentsHandler(w http.ResponseWriter, r *http.Request) {
	u, err := currentPatientUser(r)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	if r.Method == http.MethodGet {
		rows, err := DB.Query(`SELECT pp.id,pp.request_id,pp.amount,pp.status,COALESCE(r.services,''),COALESCE(r.appointment_date::text,''),COALESCE(TO_CHAR(r.appointment_time,'HH24:MI'),'') FROM patient_payments pp LEFT JOIN requests r ON r.id=pp.request_id WHERE pp.patient_id=$1 ORDER BY pp.created_at DESC`, u.PatientID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		out := []map[string]any{}
		for rows.Next() {
			var id, reqid, amount int
			var status, service, date, t string
			if err := rows.Scan(&id, &reqid, &amount, &status, &service, &date, &t); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			out = append(out, map[string]any{"id": id, "requestId": reqid, "amount": amount, "status": status, "service": service, "date": date, "time": t})
		}
		writeJSON(w, out)
		return
	}
	if r.Method == http.MethodPost {
		var body struct {
			RequestID int `json:"requestId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", 400)
			return
		}
		var amount int
		var exists bool
		err = DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM requests WHERE id=$1 AND patient_id=$2),COALESCE((SELECT price FROM requests WHERE id=$1 AND patient_id=$2),0)`, body.RequestID, u.PatientID).Scan(&exists, &amount)
		if err != nil || !exists {
			http.Error(w, "Запись не найдена", 404)
			return
		}
		var paid bool
		if err := DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM patient_payments WHERE patient_id=$1 AND request_id=$2 AND status='Оплачено')`, u.PatientID, body.RequestID).Scan(&paid); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if paid {
			writeJSON(w, map[string]any{"status": "already_paid", "amount": amount})
			return
		}
		_, err = DB.Exec(`INSERT INTO patient_payments(patient_id,request_id,amount,status,paid_at) VALUES($1,$2,$3,'Оплачено',NOW())`, u.PatientID, body.RequestID, amount)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		_, _ = DB.Exec(`INSERT INTO patient_notifications(patient_id,title,message) VALUES($1,$2,$3)`, u.PatientID, "Оплата", "Оплата по записи успешно выполнена в демонстрационном режиме.")
		writeJSON(w, map[string]any{"status": "ok", "amount": amount})
		return
	}
	http.Error(w, "method not allowed", 405)
}
