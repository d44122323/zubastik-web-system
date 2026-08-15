package main
import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)
func PatientsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		list, err := GetAllPatients()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	case http.MethodPost:
		var patient Patient
		err := json.NewDecoder(r.Body).Decode(&patient)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		err = CreatePatient(patient)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusCreated)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
func PatientHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(
		r.URL.Path,
		"/api/patients/",
	)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", 400)
		return
	}
	switch r.Method {
	case http.MethodGet:
		patient, err := GetPatientByID(id)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(patient)
	case http.MethodPut:
		var patient Patient
		err := json.NewDecoder(r.Body).Decode(&patient)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		err = UpdatePatient(id, patient)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", 405)
	}
}