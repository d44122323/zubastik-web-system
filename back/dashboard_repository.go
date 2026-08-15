package main
func GetNewToday() (int, error) {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*)
		FROM requests
		WHERE status = 'Новая'
		AND DATE(created_at) = CURRENT_DATE
	`).Scan(&count)
	return count, err
}
func GetConfirmedToday() (int, error) {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*)
		FROM requests
		WHERE status = 'Подтверждена'
		AND DATE(created_at) = CURRENT_DATE
	`).Scan(&count)
	return count, err
}
func GetCancelledToday() (int, error) {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*)
		FROM requests
		WHERE status = 'Отменена'
		AND DATE(created_at) = CURRENT_DATE
	`).Scan(&count)
	return count, err
}
func GetLastRequests() ([]Request, error) {
	rows, err := DB.Query(`
		SELECT
			r.id,
			p.name,
			p.phone,
			COALESCE(p.comment, ''),
			r.services,
			COALESCE(r.price,0),
			COALESCE(r.source,''),
			r.status,
			r.created_at
		FROM requests r
		JOIN patients p
			ON p.id = r.patient_id
		ORDER BY r.created_at DESC
		LIMIT 5
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