package main
import (
	"fmt"
	"net/http"
)
func FormHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		data := FormData{
			Name:     r.FormValue("name"),
			Phone:    r.FormValue("phone"),
			Comment:  r.FormValue("comment"),
			Services: r.FormValue("services"),
			Source:   r.FormValue("source"),
		}
		if data.Source == "" {
			data.Source = "Главная"
		}
		price := r.FormValue("price")
		if price != "" {
			fmt.Sscanf(price, "%d", &data.Price)
		}
		if err := ValidateForm(data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		patientID, err := GetOrCreatePatient(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = SaveRequest(patientID, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		SendBitrix(cfg, data)
		SendTelegram(cfg, data)
		SendEmail(cfg, data)
		http.Redirect(
			w,
			r,
			"/success.html",
			http.StatusSeeOther,
		)
	}
}