package main

import (
	"database/sql"
	"time"
)

func GetAllPatients() ([]Patient, error) {
	rows, err := DB.Query(`
SELECT
    p.id,
    p.name,
    p.phone,
    COUNT(r.id),
    COALESCE(SUM(r.price),0),
    COALESCE(MAX(r.appointment_date::timestamp + r.appointment_time), MAX(r.created_at), p.created_at),
    COALESCE((
        SELECT services FROM requests WHERE patient_id=p.id
        ORDER BY appointment_date DESC NULLS LAST, appointment_time DESC NULLS LAST, created_at DESC LIMIT 1
    ), ''),
    COALESCE((
        SELECT doctor_id FROM requests WHERE patient_id=p.id AND doctor_id IS NOT NULL
        ORDER BY appointment_date DESC NULLS LAST, appointment_time DESC NULLS LAST, created_at DESC LIMIT 1
    ), 0),
    COALESCE((
        SELECT d.name FROM requests rr JOIN doctors d ON d.id=rr.doctor_id
        WHERE rr.patient_id=p.id AND rr.doctor_id IS NOT NULL
        ORDER BY rr.appointment_date DESC NULLS LAST, rr.appointment_time DESC NULLS LAST, rr.created_at DESC LIMIT 1
    ), '')
FROM patients p
LEFT JOIN requests r ON r.patient_id = p.id
GROUP BY p.id,p.name,p.phone,p.created_at
ORDER BY p.id DESC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	patients := make([]Patient, 0)
	for rows.Next() {
		var p Patient
		if err := rows.Scan(&p.ID, &p.Name, &p.Phone, &p.Visits, &p.TotalPrice, &p.LastVisit, &p.LastService, &p.LastDoctorID, &p.LastDoctorName); err != nil {
			return nil, err
		}
		if p.LastVisit.Valid {
			p.LastVisitStr = p.LastVisit.Time.Format(time.RFC3339)
		}
		patients = append(patients, p)
	}
	return patients, rows.Err()
}

func GetPatientByID(id int) (PatientDetails, error) {
	var patient PatientDetails
	err := DB.QueryRow(`SELECT id,name,phone,comment FROM patients WHERE id=$1`, id).Scan(&patient.ID, &patient.Name, &patient.Phone, &patient.Comment)
	if err != nil {
		return patient, err
	}

	err = DB.QueryRow(`
SELECT COUNT(*), COALESCE(SUM(price),0),
       COALESCE(MAX(appointment_date::timestamp + appointment_time), MAX(created_at), NOW())
FROM requests WHERE patient_id=$1`, id).Scan(&patient.Visits, &patient.TotalPrice, &patient.LastVisit)
	if err != nil && err != sql.ErrNoRows {
		return patient, err
	}

	doctorErr := DB.QueryRow(`
SELECT COALESCE(r.doctor_id,0), COALESCE(d.name,'')
FROM requests r
LEFT JOIN doctors d ON d.id=r.doctor_id
WHERE r.patient_id=$1 AND r.doctor_id IS NOT NULL
ORDER BY r.appointment_date DESC NULLS LAST,
         r.appointment_time DESC NULLS LAST,
         r.created_at DESC
LIMIT 1`, id).Scan(&patient.LastDoctorID, &patient.LastDoctorName)
	if doctorErr != nil && doctorErr != sql.ErrNoRows {
		return patient, doctorErr
	}

	rows, err := DB.Query(`
SELECT COALESCE(r.appointment_date::timestamp + r.appointment_time, r.created_at),
       r.services,r.price,r.status,COALESCE(d.name,'Не назначен'),
       COALESCE(TO_CHAR(r.appointment_date,'YYYY-MM-DD'),''),
       COALESCE(TO_CHAR(r.appointment_time,'HH24:MI'),'')
FROM requests r LEFT JOIN doctors d ON d.id=r.doctor_id
WHERE r.patient_id=$1
ORDER BY r.appointment_date DESC NULLS LAST,r.appointment_time DESC NULLS LAST,r.created_at DESC`, id)
	if err != nil {
		return patient, err
	}
	defer rows.Close()
	for rows.Next() {
		var visit Visit
		if err := rows.Scan(&visit.Date, &visit.Service, &visit.Price, &visit.Status, &visit.DoctorName, &visit.AppointmentDate, &visit.AppointmentTime); err != nil {
			return patient, err
		}
		patient.History = append(patient.History, visit)
	}
	if err := rows.Err(); err != nil {
		return patient, err
	}

	medicalRows, err := DB.Query(`
SELECT mr.id,mr.appointment_id,COALESCE(TO_CHAR(r.appointment_date,'YYYY-MM-DD'),''),
       COALESCE(TO_CHAR(r.appointment_time,'HH24:MI'),''),mr.complaints,mr.diagnosis,
       mr.treatment,mr.recommendations,mr.created_at
FROM medical_records mr JOIN requests r ON r.id=mr.appointment_id
WHERE mr.patient_id=$1
ORDER BY r.appointment_date DESC NULLS LAST,r.appointment_time DESC NULLS LAST,mr.created_at DESC`, id)
	if err != nil {
		return patient, err
	}
	defer medicalRows.Close()
	for medicalRows.Next() {
		var m PatientMedicalRecord
		if err := medicalRows.Scan(&m.ID, &m.AppointmentID, &m.AppointmentDate, &m.AppointmentTime, &m.Complaints, &m.Diagnosis, &m.Treatment, &m.Recommendations, &m.CreatedAt); err != nil {
			return patient, err
		}
		m.Files, err = GetMedicalFilesForRecord(m.ID)
		if err != nil {
			return patient, err
		}
		patient.MedicalRecords = append(patient.MedicalRecords, m)
	}
	return patient, medicalRows.Err()
}

func GetMedicalFilesForRecord(recordID int) ([]MedicalFile, error) {
	rows, err := DB.Query(`SELECT id,medical_record_id,file_name,file_type,file_size,created_at FROM medical_files WHERE medical_record_id=$1 ORDER BY created_at DESC`, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	files := make([]MedicalFile, 0)
	for rows.Next() {
		var f MedicalFile
		if err := rows.Scan(&f.ID, &f.MedicalRecordID, &f.FileName, &f.FileType, &f.FileSize, &f.CreatedAt); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

func CreatePatient(patient Patient) error {
	_, err := DB.Exec(`INSERT INTO patients(name,phone,comment) VALUES($1,$2,$3)`, patient.Name, patient.Phone, patient.Comment)
	return err
}
func UpdatePatient(id int, patient Patient) error {
	_, err := DB.Exec(`UPDATE patients SET name=$1,phone=$2,comment=$3 WHERE id=$4`, patient.Name, patient.Phone, patient.Comment, id)
	return err
}
func GetOrCreatePatient(data FormData) (int, error) {
	var id int
	err := DB.QueryRow(`SELECT id FROM patients WHERE phone=$1`, data.Phone).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	err = DB.QueryRow(`INSERT INTO patients(name,phone,comment) VALUES($1,$2,$3) RETURNING id`, data.Name, data.Phone, data.Comment).Scan(&id)
	return id, err
}
