package main

import (
	"database/sql"
	"fmt"
	"time"
)

type AppointmentSlot struct {
	Time        string `json:"time"`
	Available   bool   `json:"available"`
	RequestID   int    `json:"requestId,omitempty"`
	PatientName string `json:"patientName,omitempty"`
	Service     string `json:"service,omitempty"`
}

type DoctorAvailability struct {
	DoctorID          int               `json:"doctorId"`
	Date              string            `json:"date"`
	DayName           string            `json:"dayName"`
	IsWorking         bool              `json:"isWorking"`
	StartTime         string            `json:"startTime"`
	EndTime           string            `json:"endTime"`
	Unavailable       bool              `json:"unavailable"`
	UnavailableReason string            `json:"unavailableReason,omitempty"`
	Slots             []AppointmentSlot `json:"slots"`
}

var weekdayNames = map[time.Weekday]string{
	time.Monday: "Понедельник", time.Tuesday: "Вторник", time.Wednesday: "Среда",
	time.Thursday: "Четверг", time.Friday: "Пятница", time.Saturday: "Суббота", time.Sunday: "Воскресенье",
}

func GetDoctorAvailability(doctorID int, date time.Time) (DoctorAvailability, error) {
	result := DoctorAvailability{DoctorID: doctorID, Date: date.Format("2006-01-02"), DayName: weekdayNames[date.Weekday()], Slots: make([]AppointmentSlot, 0)}
	var working bool
	var start, end string
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	err := DB.QueryRow(`
        SELECT is_working, TO_CHAR(start_time,'HH24:MI'), TO_CHAR(end_time,'HH24:MI')
        FROM doctor_schedule WHERE doctor_id=$1 AND weekday=$2
    `, doctorID, weekday).Scan(&working, &start, &end)
	if err == sql.ErrNoRows {

		return result, nil
	}
	if err != nil {
		return result, err
	}
	result.IsWorking, result.StartTime, result.EndTime = working, start, end
	if !working {
		return result, nil
	}

	var unavailableReason string
	err = DB.QueryRow(`
        SELECT COALESCE(reason,'')
        FROM doctor_unavailability
        WHERE doctor_id=$1 AND start_date <= $2 AND end_date >= $2
        ORDER BY start_date DESC, id DESC
        LIMIT 1
    `, doctorID, date.Format("2006-01-02")).Scan(&unavailableReason)
	if err != nil && err != sql.ErrNoRows {
		return result, err
	}
	if err == nil {
		result.Unavailable = true
		result.UnavailableReason = unavailableReason
		result.IsWorking = false
		return result, nil
	}

	startT, _ := time.Parse("15:04", start)
	endT, _ := time.Parse("15:04", end)
	for cur := startT; cur.Before(endT); cur = cur.Add(30 * time.Minute) {
		slot := cur.Format("15:04")
		result.Slots = append(result.Slots, AppointmentSlot{Time: slot, Available: true})
	}
	rows, err := DB.Query(`
        SELECT TO_CHAR(r.appointment_time,'HH24:MI'), r.id, p.name, r.services
        FROM requests r
        JOIN patients p ON p.id=r.patient_id
        WHERE r.doctor_id=$1 AND r.appointment_date=$2
          AND r.appointment_time IS NOT NULL
          AND r.status NOT IN ('Отменена','Отменено','Отменено пациентом')
    `, doctorID, date.Format("2006-01-02"))
	if err != nil {
		return result, err
	}
	defer rows.Close()
	type bookedAppointment struct {
		id            int
		name, service string
	}
	booked := map[string]bookedAppointment{}
	for rows.Next() {
		var t string
		var b bookedAppointment
		if err := rows.Scan(&t, &b.id, &b.name, &b.service); err != nil {
			return result, err
		}
		booked[t] = b
	}
	for i := range result.Slots {
		if b, ok := booked[result.Slots[i].Time]; ok {
			result.Slots[i].Available = false
			result.Slots[i].RequestID = b.id
			result.Slots[i].PatientName = b.name
			result.Slots[i].Service = b.service
		}
	}
	return result, rows.Err()
}

func ValidateAppointment(doctorID int, dateStr, timeStr string) error {
	if doctorID <= 0 || dateStr == "" || timeStr == "" {
		return fmt.Errorf("выберите врача, дату и время")
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("некорректная дата")
	}
	if date.Before(time.Now().Truncate(24 * time.Hour)) {
		return fmt.Errorf("нельзя записаться на прошедшую дату")
	}
	availability, err := GetDoctorAvailability(doctorID, date)
	if err != nil {
		return err
	}
	if !availability.IsWorking {
		return fmt.Errorf("в этот день врач не принимает")
	}
	for _, slot := range availability.Slots {
		if slot.Time == timeStr && slot.Available {
			return nil
		}
	}
	return fmt.Errorf("выбранное время уже занято или недоступно")
}
