package main

type Dashboard struct {
	NewToday          int                    `json:"newToday"`
	ConfirmedToday    int                    `json:"confirmedToday"`
	CancelledToday    int                    `json:"cancelledToday"`
	CompletedToday    int                    `json:"completedToday"`
	AppointmentsToday int                    `json:"appointmentsToday"`
	ActiveDoctors     int                    `json:"activeDoctors"`
	BusyDoctors       int                    `json:"busyDoctors"`
	FreeDoctors       int                    `json:"freeDoctors"`
	LastRequests      []Request              `json:"lastRequests"`
	TodayAppointments []DashboardAppointment `json:"todayAppointments"`
}

type DashboardAppointment struct {
	ID      int    `json:"id"`
	Time    string `json:"time"`
	Patient string `json:"patient"`
	Service string `json:"service"`
	Doctor  string `json:"doctor"`
	Status  string `json:"status"`
}
