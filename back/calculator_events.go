package main
func SaveCalculatorEvent(event string) error {
	_, err := DB.Exec(`
		INSERT INTO calculator_events(
			event_type
		)
		VALUES($1)
	`, event)
	return err
}
func GetCalculatorAnalytics() (CalculatorAnalytics, error) {
	var stats CalculatorAnalytics
	rows, err := DB.Query(`
		SELECT
			event_type,
			COUNT(*)
		FROM calculator_events
		GROUP BY event_type
	`)
	if err != nil {
		return stats, err
	}
	defer rows.Close()
	for rows.Next() {
		var event string
		var count int
		rows.Scan(&event, &count)
		switch event {
		case "open":
			stats.Open = count
		case "finish":
			stats.Finish = count
		case "request":
			stats.Request = count
		}
	}
	return stats, nil
}