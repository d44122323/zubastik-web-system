package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type DoctorPayload struct {
	Name               string   `json:"name"`
	Photo              string   `json:"photo"`
	Position           string   `json:"position"`
	Experience         string   `json:"experience"`
	Specialization     string   `json:"specialization"`
	Education          string   `json:"education"`
	Description        []string `json:"description"`
	SpecializationFull string   `json:"specializationFull"`
}

func DoctorsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	switch r.Method {
	case http.MethodGet:
		activeOnly := true
		if r.URL.Query().Get("all") == "true" {
			if !isAdminRequest(r) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			activeOnly = false
		}
		doctors, err := GetDoctors(activeOnly)
		if err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(doctors)

	case http.MethodPost:
		if !isAdminRequest(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var payload DoctorPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(payload.Name) == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}

		doctor, err := CreateDoctor(payload)
		if err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(doctor)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func DoctorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	path := strings.TrimPrefix(r.URL.Path, "/api/doctors/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "invalid doctor id", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil || id <= 0 {
		http.Error(w, "invalid doctor id", http.StatusBadRequest)
		return
	}

	if len(parts) > 1 && parts[1] == "account" {
		if !isAdminRequest(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		switch r.Method {
		case http.MethodGet:
			account, err := GetDoctorAccount(id)
			if err != nil {
				http.Error(w, "database error", http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(account)
			return
		case http.MethodPost, http.MethodPut:
			var payload DoctorAccountPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}
			if err := SaveDoctorAccount(id, payload); err != nil {
				if strings.Contains(err.Error(), "duplicate key") {
					http.Error(w, "login already exists", http.StatusConflict)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			account, err := GetDoctorAccount(id)
			if err != nil {
				http.Error(w, "database error", http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(account)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}

	if len(parts) > 1 && parts[1] == "schedule" {
		if r.Method == http.MethodGet {
			schedule, err := GetDoctorSchedule(id)
			if err != nil {
				http.Error(w, "database error", http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(schedule)
			return
		}
		if r.Method != http.MethodPut || !isAdminRequest(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var schedule DoctorSchedule
		if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		schedule.DoctorID = id
		if len(schedule.Days) != 7 {
			http.Error(w, "schedule must contain 7 days", http.StatusBadRequest)
			return
		}
		if err := SaveDoctorSchedule(schedule); err != nil {
			http.Error(w, "invalid schedule", http.StatusBadRequest)
			return
		}
		saved, err := GetDoctorSchedule(id)
		if err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(saved)
		return
	}

	if len(parts) > 1 {
		if !isAdminRequest(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		switch parts[1] {
		case "activate":
			err = ActivateDoctor(id)
		case "deactivate":
			err = DeactivateDoctor(id)
		default:
			http.Error(w, "unknown action", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "doctor not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true})
		return
	}

	switch r.Method {
	case http.MethodGet:
		doctor, err := GetDoctor(id)
		if err != nil {
			http.Error(w, "doctor not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(doctor)

	case http.MethodPut:
		if !isAdminRequest(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var payload DoctorPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(payload.Name) == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}

		doctor, err := UpdateDoctor(id, payload)
		if err != nil {
			http.Error(w, "doctor not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(doctor)

	case http.MethodDelete:
		if !isAdminRequest(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		// DELETE is intentionally a soft delete: old appointments and medical
		// records keep their doctor, but the doctor disappears from new booking.
		if err := DeactivateDoctor(id); err != nil {
			http.Error(w, "doctor not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"message": "doctor deactivated",
		})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func isAdminRequest(r *http.Request) bool {
	cookie, err := r.Cookie("admin_session")
	return err == nil && cookie.Value == "true"
}
