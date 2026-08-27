package main

import "fmt"

func GetAllRequests() ([]Request, error) {
	rows, err := DB.Query(`
SELECT
	r.id,
	p.name,
	p.phone,
	p.comment,
	r.services,
	r.price,
	r.source,
	COALESCE(r.doctor_id, 0),
	COALESCE(d.name, 'Не назначен'),
	COALESCE(TO_CHAR(r.appointment_date, 'YYYY-MM-DD'), ''),
	COALESCE(TO_CHAR(r.appointment_time, 'HH24:MI'), ''),
	r.status,
	r.created_at
FROM requests r
JOIN patients p
ON p.id = r.patient_id
LEFT JOIN doctors d ON d.id = r.doctor_id
ORDER BY r.created_at DESC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	requests := make([]Request, 0)
	for rows.Next() {
		var request Request
		err := rows.Scan(
			&request.ID,
			&request.Name,
			&request.Phone,
			&request.Comment,
			&request.Services,
			&request.Price,
			&request.Source,
			&request.DoctorID,
			&request.DoctorName,
			&request.AppointmentDate,
			&request.AppointmentTime,
			&request.Status,
			&request.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}
func GetRequestByID(id int) (Request, error) {
	var request Request
	err := DB.QueryRow(`
SELECT
	r.id,
	p.name,
	p.phone,
	p.comment,
	r.services,
	r.price,
	r.source,
	COALESCE(r.doctor_id, 0),
	COALESCE(d.name, 'Не назначен'),
	COALESCE(TO_CHAR(r.appointment_date, 'YYYY-MM-DD'), ''),
	COALESCE(TO_CHAR(r.appointment_time, 'HH24:MI'), ''),
	r.status,
	r.created_at
FROM requests r
JOIN patients p
ON p.id = r.patient_id
LEFT JOIN doctors d ON d.id = r.doctor_id
WHERE r.id = $1
`, id).Scan(
		&request.ID,
		&request.Name,
		&request.Phone,
		&request.Comment,
		&request.Services,
		&request.Price,
		&request.Source,
		&request.DoctorID,
		&request.DoctorName,
		&request.AppointmentDate,
		&request.AppointmentTime,
		&request.Status,
		&request.CreatedAt,
	)
	return request, err
}
func SaveRequest(patientID int, data FormData) error {
	// doctor_id, appointment_date and appointment_time are nullable.
	// A regular contact request may not have a selected doctor or time,
	// so we must send SQL NULL instead of 0 / empty strings.
	var doctorID any
	if data.DoctorID > 0 {
		doctorID = data.DoctorID
	} else {
		doctorID = nil
	}

	var appointmentDate any
	if data.AppointmentDate != "" {
		appointmentDate = data.AppointmentDate
	} else {
		appointmentDate = nil
	}

	var appointmentTime any
	if data.AppointmentTime != "" {
		appointmentTime = data.AppointmentTime
	} else {
		appointmentTime = nil
	}

	_, err := DB.Exec(`
		INSERT INTO requests
		(
			patient_id,
			services,
			price,
			source,
			doctor_id,
			appointment_date,
			appointment_time,
			status,
			created_at
		)
		VALUES
		(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			NOW()
		)
	`,
		patientID,
		data.Services,
		data.Price,
		data.Source,
		doctorID,
		appointmentDate,
		appointmentTime,
		"Новая",
	)
	return err
}
func UpdateRequestStatus(
	id int,
	status string,
) error {
	_, err :=
		DB.Exec(`
UPDATE requests
SET status=$1
WHERE id=$2
`,
			status,
			id,
		)
	return err
}

type AppointmentUpdate struct {
	Status          string `json:"status"`
	DoctorID        int    `json:"doctor_id"`
	AppointmentDate string `json:"appointment_date"`
	AppointmentTime string `json:"appointment_time"`
	ServiceIDs      []int  `json:"service_ids"`
	Services        string `json:"services"`
	Price           int    `json:"price"`
}

func UpdateAppointment(id int, update AppointmentUpdate) error {
	allowed := map[string]bool{"Новая": true, "Подтверждена": true, "Завершена": true, "Отменена": true}
	if !allowed[update.Status] {
		return fmt.Errorf("недопустимый статус")
	}

	current, err := GetRequestByID(id)
	if err != nil {
		return err
	}

	doctorID := update.DoctorID
	dateStr := update.AppointmentDate
	timeStr := update.AppointmentTime

	// If a field is not supplied, keep the current value.
	if doctorID <= 0 {
		doctorID = current.DoctorID
	}
	if dateStr == "" {
		dateStr = current.AppointmentDate
	}
	if timeStr == "" {
		timeStr = current.AppointmentTime
	}

	// A regular contact request may have no doctor/date/time.
	// Validate a slot only when a complete appointment is being saved.
	if update.Status != "Отменена" && doctorID > 0 && dateStr != "" && timeStr != "" {
		if err := ValidateAppointment(doctorID, dateStr, timeStr); err != nil {
			// Allow saving status-only changes for the already booked slot.
			if doctorID != current.DoctorID || dateStr != current.AppointmentDate || timeStr != current.AppointmentTime {
				return err
			}
		}
	}

	var doctorValue any
	if doctorID > 0 {
		doctorValue = doctorID
	}
	var dateValue any
	if dateStr != "" {
		dateValue = dateStr
	}
	var timeValue any
	if timeStr != "" {
		timeValue = timeStr
	}

	services := current.Services
	price := current.Price
	if update.ServiceIDs != nil {
		services, price, err = ResolveServices(update.ServiceIDs)
		if err != nil {
			return err
		}
	} else if update.Services != "" || update.Price != 0 {
		services = update.Services
		price = update.Price
	}
	if price < 0 {
		price = 0
	}
	_, err = DB.Exec(`
		UPDATE requests
		SET status=$1, doctor_id=$2, appointment_date=$3, appointment_time=$4, services=$5, price=$6
		WHERE id=$7
	`, update.Status, doctorValue, dateValue, timeValue, services, price, id)
	return err
}
