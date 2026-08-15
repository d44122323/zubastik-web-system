package main

import "time"

type Analytics struct {
	NewPatients    int                 `json:"newPatients"`
	TotalRequests  int                 `json:"totalRequests"`
	AveragePrice   int                 `json:"averagePrice"`
	Conversion     int                 `json:"conversion"`
	Chart          []ChartPoint        `json:"chart"`
	RevenueChart   []RevenuePoint      `json:"revenueChart"`
	TopServices    []ServiceStat       `json:"topServices"`
	RecentRequests []RecentRequest     `json:"recentRequests"`
	StatusChart    []StatusStat        `json:"statusChart"`
	SourceChart    []SourceStat        `json:"sourceChart"`
	Funnel         Funnel              `json:"funnel"`
	Calculator     CalculatorAnalytics `json:"calculator"`
	AIRequests     int                 `json:"aiRequests"`

	AITopQuestions []AIQuestionStat `json:"aiTopQuestions"`

	AIRecent []AIRecentRequest `json:"aiRecent"`
}
type RevenuePoint struct {
	Month  string `json:"month"`
	Amount int    `json:"amount"`
}
type StatusStat struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}
type SourceStat struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}
type Funnel struct {
	New       int `json:"new"`
	Confirmed int `json:"confirmed"`
	Completed int `json:"completed"`
	Cancelled int `json:"cancelled"`
}
type CalculatorAnalytics struct {
	Open    int `json:"open"`
	Finish  int `json:"finish"`
	Request int `json:"request"`
}
type ChartPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
type ServiceStat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}
type RecentRequest struct {
	Name    string    `json:"name"`
	Service string    `json:"service"`
	Date    time.Time `json:"date"`
	Status  string    `json:"status"`
}

type AIQuestionStat struct {
	Question string `json:"question"`

	Count int `json:"count"`
}

type AIRecentRequest struct {
	Question string `json:"question"`

	Date time.Time `json:"date"`
}
