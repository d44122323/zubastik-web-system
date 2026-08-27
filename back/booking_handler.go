package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func DoctorAvailabilityHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	path := strings.TrimPrefix(r.URL.Path, "/api/doctors/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[1] != "availability" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil || id <= 0 {
		http.Error(w, "invalid doctor id", http.StatusBadRequest)
		return
	}
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		http.Error(w, "date is required", http.StatusBadRequest)
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "invalid date", http.StatusBadRequest)
		return
	}
	if _, err := GetDoctor(id); err != nil {
		http.Error(w, "doctor not found", http.StatusNotFound)
		return
	}
	availability, err := GetDoctorAvailability(id, date)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(availability)
}
