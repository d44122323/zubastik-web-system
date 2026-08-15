package main

import (
	"sort"
	"strings"
)

func GetAnalytics() (Analytics, error) {
	var analytics Analytics
	err := DB.QueryRow(`
		SELECT COUNT(*)
		FROM requests
	`).Scan(&analytics.TotalRequests)
	if err != nil {
		return analytics, err
	}
	err = DB.QueryRow(`
		SELECT COUNT(*)
		FROM patients
	`).Scan(&analytics.NewPatients)
	if err != nil {
		return analytics, err
	}
	err = DB.QueryRow(`
		SELECT COALESCE(AVG(price),0)::INT
		FROM requests
		WHERE price > 0
	`).Scan(&analytics.AveragePrice)
	if err != nil {
		return analytics, err
	}
	var success int
	err = DB.QueryRow(`
		SELECT COUNT(*)
		FROM requests
		WHERE status='Подтверждена'
		   OR status='Завершена'
	`).Scan(&success)
	if err != nil {
		return analytics, err
	}
	if analytics.TotalRequests > 0 {
		analytics.Conversion =
			success * 100 / analytics.TotalRequests
	}
	rows, err := DB.Query(`
		SELECT
			DATE(created_at),
			COUNT(*)
		FROM requests
		GROUP BY DATE(created_at)
		ORDER BY DATE(created_at)
	`)
	if err != nil {
		return analytics, err
	}
	defer rows.Close()
	for rows.Next() {
		var point ChartPoint
		err := rows.Scan(
			&point.Date,
			&point.Count,
		)
		if err != nil {
			return analytics, err
		}
		analytics.Chart = append(
			analytics.Chart,
			point,
		)
	}
	serviceCounter := make(map[string]int)
	rows2, err := DB.Query(`
		SELECT services
		FROM requests
		WHERE services IS NOT NULL
		  AND services <> ''
	`)
	if err != nil {
		return analytics, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var services string
		rows2.Scan(&services)
		list := strings.Split(services, ",")
		for _, s := range list {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			serviceCounter[s]++
		}
	}
	for name, count := range serviceCounter {
		analytics.TopServices = append(
			analytics.TopServices,
			ServiceStat{
				Name:  name,
				Count: count,
			},
		)
	}
	sort.Slice(
		analytics.TopServices,
		func(i, j int) bool {
			return analytics.TopServices[i].Count >
				analytics.TopServices[j].Count
		},
	)
	if len(analytics.TopServices) > 5 {
		analytics.TopServices =
			analytics.TopServices[:5]
	}
	rowsSource, err := DB.Query(`
		SELECT
			source,
			COUNT(*)
		FROM requests
		GROUP BY source
		ORDER BY COUNT(*) DESC
	`)
	if err != nil {
		return analytics, err
	}
	defer rowsSource.Close()
	for rowsSource.Next() {
		var item SourceStat
		err := rowsSource.Scan(
			&item.Source,
			&item.Count,
		)
		if err != nil {
			return analytics, err
		}
		switch item.Source {
		case "calculator":
			item.Source = "Калькулятор"
		case "form":
			item.Source = "Форма"
		case "phone":
			item.Source = "Телефон"
		case "admin":
			item.Source = "Администратор"
		case "":
			item.Source = "Неизвестно"
		}
		analytics.SourceChart = append(
			analytics.SourceChart,
			item,
		)
	}
	rowsStatus, err := DB.Query(`
		SELECT
			status,
			COUNT(*)
		FROM requests
		GROUP BY status
	`)
	if err != nil {
		return analytics, err
	}
	defer rowsStatus.Close()
	for rowsStatus.Next() {
		var item StatusStat
		err := rowsStatus.Scan(
			&item.Status,
			&item.Count,
		)
		if err != nil {
			return analytics, err
		}
		analytics.StatusChart = append(
			analytics.StatusChart,
			item,
		)
		switch item.Status {
		case "Новая":
			analytics.Funnel.New = item.Count
		case "Подтверждена":
			analytics.Funnel.Confirmed = item.Count
		case "Завершена":
			analytics.Funnel.Completed = item.Count
		case "Отменена":
			analytics.Funnel.Cancelled = item.Count
		}
	}
	rowsRevenue, err := DB.Query(`
		SELECT
			TO_CHAR(DATE_TRUNC('month', created_at), 'MM.YYYY') AS month,
			COALESCE(SUM(price), 0) AS amount
		FROM requests
		WHERE status = 'Завершена'
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY DATE_TRUNC('month', created_at)
	`)
	if err != nil {
		return analytics, err
	}
	defer rowsRevenue.Close()
	for rowsRevenue.Next() {
		var point RevenuePoint
		err := rowsRevenue.Scan(
			&point.Month,
			&point.Amount,
		)
		if err != nil {
			return analytics, err
		}
		analytics.RevenueChart = append(
			analytics.RevenueChart,
			point,
		)
	}
	rows3, err := DB.Query(`
		SELECT
			p.name,
			r.services,
			r.created_at,
			r.status
		FROM requests r
		JOIN patients p
			ON p.id = r.patient_id
		ORDER BY r.created_at DESC
		LIMIT 5
	`)
	if err != nil {
		return analytics, err
	}
	defer rows3.Close()
	for rows3.Next() {
		var item RecentRequest
		err := rows3.Scan(
			&item.Name,
			&item.Service,
			&item.Date,
			&item.Status,
		)
		if err != nil {
			return analytics, err
		}
		analytics.RecentRequests = append(
			analytics.RecentRequests,
			item,
		)
	}
	calculator, err := GetCalculatorAnalytics()
	if err != nil {
		return analytics, err
	}
	analytics.Calculator = calculator

	aiTotal,
		aiTop,
		aiRecent,
		err := GetAIAnalytics()

	if err != nil {

		return analytics, err

	}

	analytics.AIRequests = aiTotal

	analytics.AITopQuestions = aiTop

	analytics.AIRecent = aiRecent

	return analytics, nil
}
