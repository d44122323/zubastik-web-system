package main

import (
	"encoding/json"
	"net/http"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	if !isAdminRequest(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	data := Dashboard{}
	data.NewToday, _ = GetNewToday()
	data.ConfirmedToday, _ = GetConfirmedToday()
	data.CancelledToday, _ = GetCancelledToday()
	data.CompletedToday, _ = GetCompletedToday()
	data.AppointmentsToday, _ = GetAppointmentsTodayCount()
	data.ActiveDoctors, data.BusyDoctors, data.FreeDoctors, _ = GetDoctorOccupancyToday()
	data.LastRequests, _ = GetLastRequests()
	data.TodayAppointments, _ = GetTodayAppointments()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
