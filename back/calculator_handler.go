package main
import (
	"net/http"
)
func CalculatorEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	event := r.FormValue("event")
	if event == "" {
		http.Error(w, "event required", http.StatusBadRequest)
		return
	}
	err := SaveCalculatorEvent(event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}