package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func PatientRescheduleAppointmentHandler(w http.ResponseWriter, r *http.Request) {
	u, err := currentPatientUser(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/patient/appointments/")
	path = strings.TrimSuffix(path, "/reschedule")
	id, err := strconv.Atoi(strings.Trim(path, "/"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var body struct {
		Action string `json:"action"`
		Date   string `json:"date"`
		Time   string `json:"time"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.Action != "reschedule" {
		http.Error(w, "неизвестное действие", http.StatusBadRequest)
		return
	}

	var doctorID int
	var status string
	if err := DB.QueryRow(`SELECT COALESCE(doctor_id,0),status FROM requests WHERE id=$1 AND patient_id=$2`, id, u.PatientID).Scan(&doctorID, &status); err != nil {
		http.Error(w, "Запись не найдена", http.StatusNotFound)
		return
	}
	if doctorID <= 0 {
		http.Error(w, "У записи не назначен врач", http.StatusConflict)
		return
	}
	if status == "Завершена" || status == "Отменена" || status == "Отменено" || status == "Отменено пациентом" {
		http.Error(w, "Эту запись нельзя перенести", http.StatusConflict)
		return
	}
	if err := ValidateAppointment(doctorID, body.Date, body.Time); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	res, err := DB.Exec(`UPDATE requests SET appointment_date=$1::date,appointment_time=$2::time WHERE id=$3 AND patient_id=$4 AND status NOT IN ('Завершена','Отменена','Отменено','Отменено пациентом')`, body.Date, body.Time, id, u.PatientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, "Запись не удалось перенести", http.StatusConflict)
		return
	}
	_, _ = DB.Exec(`INSERT INTO patient_notifications(patient_id,title,message) VALUES($1,$2,$3)`, u.PatientID, "Запись перенесена", "Ваша запись перенесена на "+body.Date+" в "+body.Time+".")
	_, _ = DB.Exec(`DELETE FROM telegram_sent WHERE request_id=$1`, id)
	go telegramRescheduleNotification(LoadConfig(), id)
	writeJSON(w, map[string]any{"status": "ok", "date": body.Date, "time": body.Time, "doctorId": doctorID})
}

func patientBookingMessage(status string) string {
	switch status {
	case "Подтверждена":
		return "Ваша запись подтверждена."
	case "Отменено пациентом":
		return "Вы отменили запись."
	case "Отменена", "Отменено":
		return "Ваша запись отменена клиникой."
	case "Завершена":
		return "Приём завершён."
	default:
		return "Статус записи: " + status
	}
}
