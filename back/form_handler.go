package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func FormHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		patient, authErr := currentPatientUser(r)
		if authErr != nil {
			http.Redirect(w, r, "/?account=patient&record=1&notice=Для+отправки+заявки+войдите+или+зарегистрируйтесь", http.StatusSeeOther)
			return
		}
		if err := enforceRateLimit(fmt.Sprintf("request:%d:%s", patient.PatientID, clientIP(r)), 10, 10*time.Minute); err != nil {
			http.Error(w, err.Error(), http.StatusTooManyRequests)
			return
		}
		if err := validateCaptcha(r, r.FormValue("captcha_token"), r.FormValue("captcha_answer")); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data := FormData{
			Name:            r.FormValue("name"),
			Phone:           r.FormValue("phone"),
			Comment:         r.FormValue("comment"),
			Services:        r.FormValue("services"),
			Source:          r.FormValue("source"),
			DoctorID:        0,
			AppointmentDate: r.FormValue("appointment_date"),
			AppointmentTime: r.FormValue("appointment_time"),
			CaptchaToken:    r.FormValue("captcha_token"),
			CaptchaAnswer:   r.FormValue("captcha_answer"),
		}
		if doctorID := r.FormValue("doctor_id"); doctorID != "" {
			data.DoctorID, _ = strconv.Atoi(doctorID)
		}
		if data.DoctorID > 0 && data.AppointmentDate != "" && data.AppointmentTime != "" {
			if err := ValidateAppointment(data.DoctorID, data.AppointmentDate, data.AppointmentTime); err != nil {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
		}
		if data.Source == "" {
			data.Source = "Главная"
		}
		price := r.FormValue("price")
		if price != "" {
			fmt.Sscanf(price, "%d", &data.Price)
		}
		if err := ValidateForm(data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		patientID := patient.PatientID
		var accountName, accountPhone string
		if err := DB.QueryRow(`SELECT name,phone FROM patients WHERE id=$1`, patientID).Scan(&accountName, &accountPhone); err != nil {
			http.Error(w, "Не удалось получить данные пациента", http.StatusInternalServerError)
			return
		}
		data.Name = accountName
		data.Phone = accountPhone
		if err := SaveRequest(patientID, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		SendBitrix(cfg, data)
		SendTelegram(cfg, data)
		SendEmail(cfg, data)
		http.Redirect(
			w,
			r,
			"/success.html",
			http.StatusSeeOther,
		)
	}
}
