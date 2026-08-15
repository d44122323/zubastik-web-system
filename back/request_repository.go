package main
func GetAllRequests() ([]Request, error) {
	rows, err := DB.Query(`
SELECT
	r.id,
	p.name,
	p.phone,
	p.comment,
	r.services,
	r.price,
	r.source,
	r.status,
	r.created_at
FROM requests r
JOIN patients p
ON p.id = r.patient_id
ORDER BY r.created_at DESC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var requests []Request
	for rows.Next() {
		var request Request
		err := rows.Scan(
			&request.ID,
			&request.Name,
			&request.Phone,
			&request.Comment,
			&request.Services,
			&request.Price,
			&request.Source,
			&request.Status,
			&request.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}
func GetRequestByID(id int) (Request, error) {
	var request Request
	err := DB.QueryRow(`
SELECT
	r.id,
	p.name,
	p.phone,
	p.comment,
	r.services,
	r.price,
	r.source,
	r.status,
	r.created_at
FROM requests r
JOIN patients p
ON p.id = r.patient_id
WHERE r.id = $1
`, id).Scan(
		&request.ID,
		&request.Name,
		&request.Phone,
		&request.Comment,
		&request.Services,
		&request.Price,
		&request.Source,
		&request.Status,
		&request.CreatedAt,
	)
	return request, err
}
func SaveRequest(patientID int, data FormData) error {
	_, err := DB.Exec(`
		INSERT INTO requests
		(
			patient_id,
			services,
			price,
			source,
			status,
			created_at
		)
		VALUES
		(
			$1,
			$2,
			$3,
			$4,
			$5,
			NOW()
		)
	`,
		patientID,
		data.Services,
		data.Price,
		data.Source,
		"Новая",
	)
	return err
}
func UpdateRequestStatus(
	id int,
	status string,
) error {
	_, err :=
		DB.Exec(`
UPDATE requests
SET status=$1
WHERE id=$2
`,
			status,
			id,
		)
	return err
}