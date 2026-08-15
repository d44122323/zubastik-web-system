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
	COALESCE(MAX(r.created_at), p.created_at),
	COALESCE(
		(
			SELECT services
			FROM requests
			WHERE patient_id = p.id
			ORDER BY created_at DESC
			LIMIT 1
		),
		''
	)
FROM patients p
LEFT JOIN requests r
ON r.patient_id = p.id
GROUP BY
	p.id,
	p.name,
	p.phone,
	p.created_at
ORDER BY p.id DESC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var patients []Patient
	for rows.Next() {
		var p Patient
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Phone,
			&p.Visits,
			&p.TotalPrice,
			&p.LastVisit,
			&p.LastService,
		)
		if err != nil {
			return nil, err
		}
		if p.LastVisit.Valid {
			p.LastVisitStr = p.LastVisit.Time.Format(time.RFC3339)
		}
		patients = append(patients, p)
	}
	return patients, nil
}
func GetPatientByID(id int) (PatientDetails, error) {
	var patient PatientDetails
	err := DB.QueryRow(`
SELECT
	id,
	name,
	phone,
	comment
FROM patients
WHERE id=$1
`, id).Scan(
		&patient.ID,
		&patient.Name,
		&patient.Phone,
		&patient.Comment,
	)
	if err != nil {
		return patient, err
	}
	err = DB.QueryRow(`
SELECT
	COUNT(*),
	COALESCE(SUM(price),0),
	COALESCE(MAX(created_at), NOW())
FROM requests
WHERE patient_id=$1
`, id).Scan(
		&patient.Visits,
		&patient.TotalPrice,
		&patient.LastVisit,
	)
	if err != nil && err != sql.ErrNoRows {
		return patient, err
	}
	rows, err := DB.Query(`
SELECT
	created_at,
	services,
	price
FROM requests
WHERE patient_id=$1
ORDER BY created_at DESC
`, id)
	if err != nil {
		return patient, err
	}
	defer rows.Close()
	for rows.Next() {
		var visit Visit
		err := rows.Scan(
			&visit.Date,
			&visit.Service,
			&visit.Price,
		)
		if err != nil {
			return patient, err
		}
		patient.History = append(patient.History, visit)
	}
	return patient, nil
}
func CreatePatient(patient Patient) error {
	_, err := DB.Exec(`
INSERT INTO patients
(
	name,
	phone,
	comment
)
VALUES
(
	$1,
	$2,
	$3
)
`,
		patient.Name,
		patient.Phone,
		patient.Comment,
	)
	return err
}
func UpdatePatient(id int, patient Patient) error {
	_, err := DB.Exec(`
UPDATE patients
SET
	name=$1,
	phone=$2,
	comment=$3
WHERE id=$4
`,
		patient.Name,
		patient.Phone,
		patient.Comment,
		id,
	)
	return err
}
func GetOrCreatePatient(data FormData) (int, error) {
	var id int
	err := DB.QueryRow(`
SELECT id
FROM patients
WHERE phone=$1
`,
		data.Phone,
	).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	err = DB.QueryRow(`
INSERT INTO patients
(
	name,
	phone,
	comment
)
VALUES
(
	$1,
	$2,
	$3
)
RETURNING id
`,
		data.Name,
		data.Phone,
		data.Comment,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}