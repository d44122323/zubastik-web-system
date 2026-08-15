package main
import (
	"encoding/json"
	"net/http"
)
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	data := Dashboard{}
	data.NewToday, _ = GetNewToday()
	data.ConfirmedToday, _ = GetConfirmedToday()
	data.CancelledToday, _ = GetCancelledToday()
	data.LastRequests, _ = GetLastRequests()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}