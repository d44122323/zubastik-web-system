package main

import (
	"net/http"
	"strings"
	"time"
)

func DoctorScheduleHandler(w http.ResponseWriter, r *http.Request) {
	u, err := DoctorRequestGuard(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/doctor/schedule")
	path = strings.Trim(path, "/")
	if path == "" {
		schedule, err := GetDoctorSchedule(u.DoctorID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, schedule)
		return
	}
	if path != "availability" {
		http.Error(w, "not found", http.StatusNotFound)
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
	availability, err := GetDoctorAvailability(u.DoctorID, date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, availability)
}
