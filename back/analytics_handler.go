package main
import (
	"encoding/json"
	"net/http"
)
func AnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	analytics, err := GetAnalytics()
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}
	w.Header().Set(
		"Content-Type",
		"application/json",
	)
	json.NewEncoder(w).Encode(analytics)
}