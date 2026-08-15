package main

import (
	"sort"
	"strings"
)

func SaveAIRequest(question string) {

	_, err := DB.Exec(
		`
		INSERT INTO ai_requests(question)
		VALUES($1)
		`,
		question,
	)

	if err != nil {

		println(err.Error())

	}

}

func GetAIAnalytics() (
	int,
	[]AIQuestionStat,
	[]AIRecentRequest,
	error,
) {

	var total int

	err := DB.QueryRow(
		`
		SELECT COUNT(*)
		FROM ai_requests
		`,
	).Scan(&total)

	if err != nil {

		return 0, nil, nil, err

	}

	counter := make(map[string]int)

	rows, err := DB.Query(
		`
		SELECT question
		FROM ai_requests
		`,
	)

	if err != nil {

		return 0, nil, nil, err

	}

	defer rows.Close()

	for rows.Next() {

		var q string

		rows.Scan(&q)

		q = strings.ToLower(q)

		counter[q]++

	}

	var top []AIQuestionStat

	for q, count := range counter {

		top = append(
			top,
			AIQuestionStat{
				Question: q,
				Count:    count,
			},
		)

	}

	sort.Slice(
		top,
		func(i, j int) bool {

			return top[i].Count >
				top[j].Count

		},
	)

	if len(top) > 5 {

		top = top[:5]

	}

	recentRows, err := DB.Query(
		`
		SELECT question,created_at
		FROM ai_requests
		ORDER BY created_at DESC
		LIMIT 10
		`,
	)

	if err != nil {

		return total, top, nil, err

	}

	defer recentRows.Close()

	var recent []AIRecentRequest

	for recentRows.Next() {

		var item AIRecentRequest

		recentRows.Scan(
			&item.Question,
			&item.Date,
		)

		recent = append(
			recent,
			item,
		)

	}

	return total, top, recent, nil

}
