package main
type Dashboard struct {
	NewToday       int `json:"newToday"`
	ConfirmedToday int `json:"confirmedToday"`
	CancelledToday int `json:"cancelledToday"`
	LastRequests []Request `json:"lastRequests"`
}