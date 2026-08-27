package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func UpdateRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if !isAdminRequest(r) {
		http.Error(w, "unauthorized", 401)
		return
	}
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/requests/"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid ID", 400)
		return
	}
	var body AppointmentUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}
	if err := UpdateAppointment(id, body); err != nil {
		http.Error(w, err.Error(), 409)
		return
	}
	request, err := GetRequestByID(id)
	if err != nil {
		http.Error(w, "Request not found", 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(request)
	if body.Status != "" {
		go telegramStatusNotification(LoadConfig(), id, body.Status)
	}
	if body.AppointmentDate != "" || body.AppointmentTime != "" {
		_, _ = DB.Exec(`DELETE FROM telegram_sent WHERE request_id=$1`, id)
		go telegramRescheduleNotification(LoadConfig(), id)
	}
}
