package main

type FormData struct {
	Name            string
	Phone           string
	Comment         string
	Services        string
	ServiceIDs      []int
	Price           int
	Source          string
	DoctorID        int
	AppointmentDate string
	AppointmentTime string
	CaptchaToken    string
	CaptchaAnswer   string
}
