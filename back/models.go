package main
import (
	"database/sql"
	"time"
)
type Patient struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Phone string `json:"phone"`
	Comment string `json:"comment"`
	Visits int `json:"visits"`
	TotalPrice int `json:"total"`
	LastService string `json:"lastService"`
	LastVisit sql.NullTime `json:"-"`
	LastVisitStr string `json:"lastVisit"`
}
type Visit struct {
	Date    time.Time `json:"date"`
	Service string    `json:"service"`
	Price   int       `json:"price"`
}
type PatientDetails struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Comment string `json:"comment"`
	Visits     int       `json:"visits"`
	TotalPrice int       `json:"total"`
	LastVisit  time.Time `json:"lastVisit"`
	History []Visit `json:"history"`
}