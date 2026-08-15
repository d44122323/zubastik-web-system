package main

import (
	"net/http"
	"os"
)

func SetupRouter(cfg Config) http.Handler {

	mux := http.NewServeMux()

	mux.Handle(
		"/submit",
		FormHandler(cfg),
	)

	mux.HandleFunc(
		"/admin/login",
		AdminLoginHandler,
	)

	mux.HandleFunc(
		"/api/requests",
		GetRequestsHandler,
	)

	mux.HandleFunc(
		"/api/requests/",
		func(w http.ResponseWriter, r *http.Request) {

			switch r.Method {

			case http.MethodGet:
				GetRequestHandler(w, r)

			case http.MethodPut:
				UpdateRequestHandler(w, r)

			default:
				http.Error(
					w,
					"Method not allowed",
					http.StatusMethodNotAllowed,
				)
			}
		},
	)

	mux.HandleFunc(
		"/api/dashboard",
		DashboardHandler,
	)

	mux.HandleFunc(
		"/api/analytics",
		AnalyticsHandler,
	)

	mux.HandleFunc(
		"/api/calculator-event",
		CalculatorEventHandler,
	)

	mux.HandleFunc(
		"/api/patients",
		PatientsHandler,
	)

	mux.HandleFunc(
		"/api/patients/",
		PatientHandler,
	)

	frontPath := "front"

	if _, err := os.Stat(frontPath); os.IsNotExist(err) {

		if _, err := os.Stat("../front"); err == nil {
			frontPath = "../front"
		}
	}

	frontFS := http.FileServer(
		http.Dir(frontPath),
	)

	mux.Handle(
		"/login.html",
		frontFS,
	)

	mux.Handle(
		"/login.css",
		frontFS,
	)

	mux.Handle(
		"/login.js",
		frontFS,
	)

	mux.Handle(
		"/img/",
		frontFS,
	)

	mux.Handle(
		"/dashboard.html",
		AuthMiddleware(frontFS),
	)

	mux.Handle(
		"/requests.html",
		AuthMiddleware(frontFS),
	)

	mux.Handle(
		"/patients.html",
		AuthMiddleware(frontFS),
	)

	mux.Handle(
		"/analytics.html",
		AuthMiddleware(frontFS),
	)

	mux.Handle(
		"/",
		frontFS,
	)

	mux.HandleFunc(
		"/chat",
		AIChatHandler,
	)

	return mux
}
