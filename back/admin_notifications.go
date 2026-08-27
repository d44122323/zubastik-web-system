package main

import (
	"encoding/json"
	"net/http"
)

type AdminNotification struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
	Link      string `json:"link"`
}

func AdminNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	if !isAdminRequest(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rows, err := DB.Query(`
        SELECT r.id, COALESCE(p.name,''), COALESCE(r.services,''), COALESCE(d.name,''),
               COALESCE(TO_CHAR(r.appointment_date,'YYYY-MM-DD'),''),
               COALESCE(TO_CHAR(r.appointment_time,'HH24:MI'),''),
               COALESCE(TO_CHAR(r.created_at,'YYYY-MM-DD HH24:MI:SS'),'')
        FROM requests r
        JOIN patients p ON p.id = r.patient_id
        LEFT JOIN doctors d ON d.id = r.doctor_id
        WHERE r.status = 'Новая'
        ORDER BY r.created_at DESC
        LIMIT 12`)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	result := []AdminNotification{}
	for rows.Next() {
		var n AdminNotification
		var name, services, doctor, date, tm, created string
		if err := rows.Scan(&n.ID, &name, &services, &doctor, &date, &tm, &created); err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		n.Type = "request"
		n.Title = "Новая заявка"
		n.Text = name + " · " + services
		if doctor != "" {
			n.Text += " · " + doctor
		}
		if date != "" {
			n.Text += " · " + date
		}
		if tm != "" {
			n.Text += " " + tm
		}
		n.Link = "requests.html"
		n.CreatedAt = created
		result = append(result, n)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(result)
}

func AdminSystemCheckHandler(w http.ResponseWriter, r *http.Request) {
	if !isAdminRequest(r) {
		http.Error(w, "unauthorized", 401)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 405)
		return
	}
	checks := []map[string]any{}
	tables := []string{"doctors", "requests", "patients", "services", "doctor_schedule", "medical_records", "medical_files"}
	for _, table := range tables {
		var exists bool
		err := DB.QueryRow(`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name=$1)`, table).Scan(&exists)
		checks = append(checks, map[string]any{"name": table, "ok": err == nil && exists})
	}
	var activeDoctors int
	err := DB.QueryRow(`SELECT COUNT(*) FROM doctors WHERE is_active=true`).Scan(&activeDoctors)
	checks = append(checks, map[string]any{"name": "active_doctors", "ok": err == nil && activeDoctors > 0})
	var serviceCount int
	err = DB.QueryRow(`SELECT COUNT(*) FROM services WHERE is_active=true`).Scan(&serviceCount)
	checks = append(checks, map[string]any{"name": "active_services", "ok": err == nil && serviceCount > 0})
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(checks)
}
