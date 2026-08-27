package main

type Doctor struct {
	ID                 int      `json:"id"`
	Name               string   `json:"name"`
	Photo              string   `json:"photo"`
	Position           string   `json:"position"`
	Experience         string   `json:"experience"`
	Specialization     string   `json:"specialization"`
	Education          string   `json:"education"`
	Description        []string `json:"description"`
	SpecializationFull string   `json:"specializationFull"`
	IsActive           bool     `json:"isActive"`
}

type DoctorScheduleDay struct {
	Weekday   int    `json:"weekday"`
	DayName   string `json:"dayName"`
	IsWorking bool   `json:"isWorking"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type DoctorSchedule struct {
	DoctorID int                 `json:"doctorId"`
	Days     []DoctorScheduleDay `json:"days"`
}
