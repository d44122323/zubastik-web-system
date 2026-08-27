package main

import (
	"database/sql"
	"time"
)

type Patient struct {
	ID             int          `json:"id"`
	Name           string       `json:"name"`
	Phone          string       `json:"phone"`
	Comment        string       `json:"comment"`
	Visits         int          `json:"visits"`
	TotalPrice     int          `json:"total"`
	LastService    string       `json:"lastService"`
	LastDoctorID   int          `json:"lastDoctorId,omitempty"`
	LastDoctorName string       `json:"lastDoctorName,omitempty"`
	LastVisit      sql.NullTime `json:"-"`
	LastVisitStr   string       `json:"lastVisit"`
}
type Visit struct {
	Date            time.Time `json:"date"`
	Service         string    `json:"service"`
	Price           int       `json:"price"`
	Status          string    `json:"status,omitempty"`
	DoctorName      string    `json:"doctorName,omitempty"`
	AppointmentDate string    `json:"appointmentDate,omitempty"`
	AppointmentTime string    `json:"appointmentTime,omitempty"`
}
type PatientMedicalRecord struct {
	ID              int           `json:"id"`
	AppointmentID   int           `json:"appointment_id"`
	AppointmentDate string        `json:"appointment_date,omitempty"`
	AppointmentTime string        `json:"appointment_time,omitempty"`
	Complaints      string        `json:"complaints"`
	Diagnosis       string        `json:"diagnosis"`
	Treatment       string        `json:"treatment"`
	Recommendations string        `json:"recommendations"`
	CreatedAt       time.Time     `json:"created_at"`
	Files           []MedicalFile `json:"files,omitempty"`
}
type PatientDetails struct {
	ID             int                    `json:"id"`
	Name           string                 `json:"name"`
	Phone          string                 `json:"phone"`
	Comment        string                 `json:"comment"`
	Visits         int                    `json:"visits"`
	TotalPrice     int                    `json:"total"`
	LastVisit      time.Time              `json:"lastVisit"`
	LastDoctorID   int                    `json:"lastDoctorId,omitempty"`
	LastDoctorName string                 `json:"lastDoctorName,omitempty"`
	History        []Visit                `json:"history"`
	MedicalRecords []PatientMedicalRecord `json:"medical_records,omitempty"`
}
