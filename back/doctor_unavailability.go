package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type DoctorUnavailability struct {
	ID        int    `json:"id"`
	DoctorID  int    `json:"doctorId"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Reason    string `json:"reason"`
}

func DoctorUnavailabilityHandler(w http.ResponseWriter, r *http.Request) {
	if !isAdminRequest(r) {
		http.Error(w, "unauthorized", 401)
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/doctors/"), "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] != "unavailability" {
		http.Error(w, "not found", 404)
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil || id <= 0 {
		http.Error(w, "invalid doctor id", 400)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch r.Method {
	case http.MethodGet:
		rows, err := DB.Query(`SELECT id,doctor_id,start_date::text,end_date::text,reason FROM doctor_unavailability WHERE doctor_id=$1 ORDER BY start_date`, id)
		if err != nil {
			http.Error(w, "database error", 500)
			return
		}
		defer rows.Close()
		out := []DoctorUnavailability{}
		for rows.Next() {
			var x DoctorUnavailability
			if err := rows.Scan(&x.ID, &x.DoctorID, &x.StartDate, &x.EndDate, &x.Reason); err != nil {
				http.Error(w, "database error", 500)
				return
			}
			out = append(out, x)
		}
		json.NewEncoder(w).Encode(out)
	case http.MethodPost:
		var x DoctorUnavailability
		if json.NewDecoder(r.Body).Decode(&x) != nil {
			http.Error(w, "invalid json", 400)
			return
		}
		if x.StartDate == "" || x.EndDate == "" {
			http.Error(w, "date is required", 400)
			return
		}
		var out DoctorUnavailability
		err := DB.QueryRow(`INSERT INTO doctor_unavailability(doctor_id,start_date,end_date,reason) VALUES($1,$2::date,$3::date,$4) RETURNING id,doctor_id,start_date::text,end_date::text,reason`, id, x.StartDate, x.EndDate, strings.TrimSpace(x.Reason)).Scan(&out.ID, &out.DoctorID, &out.StartDate, &out.EndDate, &out.Reason)
		if err != nil {
			http.Error(w, "invalid date range", 400)
			return
		}
		json.NewEncoder(w).Encode(out)
	case http.MethodDelete:
		if len(parts) < 3 {
			http.Error(w, "invalid id", 400)
			return
		}
		uid, e := strconv.Atoi(parts[2])
		if e != nil {
			http.Error(w, "invalid id", 400)
			return
		}
		res, e := DB.Exec(`DELETE FROM doctor_unavailability WHERE id=$1 AND doctor_id=$2`, uid, id)
		if e != nil {
			http.Error(w, "database error", 500)
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true})
	default:
		http.Error(w, "method not allowed", 405)
	}
}
