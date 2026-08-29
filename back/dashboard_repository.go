package main

func GetNewToday() (int, error)       { return countTodayStatus("Новая") }
func GetConfirmedToday() (int, error) { return countTodayStatus("Подтверждена") }
func GetCancelledToday() (int, error) { return countTodayStatus("Отменена") }
func GetCompletedToday() (int, error) { return countTodayStatus("Завершена") }

func countTodayStatus(status string) (int, error) {
	var count int
	err := DB.QueryRow(`SELECT COUNT(*) FROM requests WHERE status=$1 AND appointment_date=CURRENT_DATE`, status).Scan(&count)
	if err != nil {

		err = DB.QueryRow(`SELECT COUNT(*) FROM requests WHERE status=$1 AND DATE(created_at)=CURRENT_DATE`, status).Scan(&count)
	}
	return count, err
}

func GetAppointmentsTodayCount() (int, error) {
	var count int
	err := DB.QueryRow(`SELECT COUNT(*) FROM requests WHERE appointment_date=CURRENT_DATE AND status NOT IN ('Отменена','Отменено','Отменено пациентом') AND appointment_time IS NOT NULL`).Scan(&count)
	return count, err
}

func GetDoctorOccupancyToday() (active, busy, free int, err error) {
	if err = DB.QueryRow(`SELECT COUNT(*) FROM doctors WHERE is_active=true`).Scan(&active); err != nil {
		return
	}
	if err = DB.QueryRow(`SELECT COUNT(DISTINCT doctor_id) FROM requests WHERE appointment_date=CURRENT_DATE AND doctor_id IS NOT NULL AND status NOT IN ('Отменена','Отменено','Отменено пациентом')`).Scan(&busy); err != nil {
		return
	}
	free = active - busy
	if free < 0 {
		free = 0
	}
	return
}

func GetTodayAppointments() ([]DashboardAppointment, error) {
	rows, err := DB.Query(`
SELECT r.id,
       COALESCE(TO_CHAR(r.appointment_time,'HH24:MI'),''),
       p.name,
       COALESCE(r.services,'Без услуг'),
       COALESCE(d.name,'Не назначен'),
       r.status
FROM requests r
JOIN patients p ON p.id=r.patient_id
LEFT JOIN doctors d ON d.id=r.doctor_id
WHERE r.appointment_date=CURRENT_DATE
ORDER BY r.appointment_time NULLS LAST, r.id
LIMIT 12`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]DashboardAppointment, 0)
	for rows.Next() {
		var a DashboardAppointment
		if err := rows.Scan(&a.ID, &a.Time, &a.Patient, &a.Service, &a.Doctor, &a.Status); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

func GetLastRequests() ([]Request, error) {
	rows, err := DB.Query(`
        SELECT r.id,p.name,p.phone,COALESCE(p.comment,''),r.services,COALESCE(r.price,0),COALESCE(r.source,''),r.status,r.created_at
        FROM requests r JOIN patients p ON p.id=r.patient_id
        ORDER BY r.created_at DESC LIMIT 5`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	requests := make([]Request, 0)
	for rows.Next() {
		var request Request
		if err := rows.Scan(&request.ID, &request.Name, &request.Phone, &request.Comment, &request.Services, &request.Price, &request.Source, &request.Status, &request.CreatedAt); err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, rows.Err()
}
