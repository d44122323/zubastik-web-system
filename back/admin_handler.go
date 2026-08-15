package main
import (
	"encoding/json"
	"net/http"
)
type LoginRequest struct {
	Login string `json:"login"`
	Password string `json:"password"`
}
func AdminLoginHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}
	var data LoginRequest
	err := json.NewDecoder(
		r.Body,
	).Decode(&data)
	if err != nil {
		http.Error(
			w,
			"bad request",
			http.StatusBadRequest,
		)
		return
	}
	_, err = CheckAdmin(
		data.Login,
		data.Password,
	)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusUnauthorized,
		)
		return
	}
	http.SetCookie(
		w,
		&http.Cookie{
			Name: "admin_session",
			Value: "true",
			Path: "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge: 3600,
		},
	)
	w.Header().Set(
		"Content-Type",
		"application/json",
	)
	json.NewEncoder(w).Encode(
		map[string]string{
			"status": "ok",
		},
	)
}