package main

import "time"

type Request struct {
	ID                 int       `json:"id"`
	PatientID          int       `json:"patient_id,omitempty"`
	Name               string    `json:"name"`
	Phone              string    `json:"phone"`
	Comment            string    `json:"comment"`
	Services           string    `json:"services"`
	Price              int       `json:"price"`
	Source             string    `json:"source"`
	DoctorID           int       `json:"doctor_id"`
	DoctorName         string    `json:"doctor_name,omitempty"`
	AppointmentDate    string    `json:"appointment_date,omitempty"`
	AppointmentTime    string    `json:"appointment_time,omitempty"`
	AppointmentComment string    `json:"appointment_comment,omitempty"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
}
