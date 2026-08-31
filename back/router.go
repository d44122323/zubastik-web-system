package main

import (
	"net/http"
	"strings"
)

func SetupRouter(cfg Config) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/submit", FormHandler(cfg))
	mux.HandleFunc("/api/captcha", issueCaptcha)
	mux.HandleFunc(
		"/admin/login",
		AdminLoginHandler,
	)
	mux.HandleFunc("/patient/login", PatientLoginHandler)
	mux.HandleFunc("/patient/register", PatientRegisterHandler)
	mux.HandleFunc("/patient/logout", PatientLogoutHandler)
	mux.HandleFunc("/api/patient/me", PatientMeHandler)
	mux.HandleFunc("/api/patient/profile", PatientProfileHandler)
	mux.HandleFunc("/api/patient/appointments", PatientAppointmentsHandler)
	mux.HandleFunc("/api/patient/appointments/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && strings.HasSuffix(strings.TrimRight(r.URL.Path, "/"), "/reschedule") {
			PatientRescheduleAppointmentHandler(w, r)
			return
		}
		PatientAppointmentsHandler(w, r)
	})
	mux.HandleFunc("/api/patient/medical", PatientMedicalHandler)
	mux.HandleFunc("/api/patient/documents", PatientDocumentsHandler)
	mux.HandleFunc("/api/patient/files/", PatientMedicalFileHandler)
	mux.HandleFunc("/api/patient/notifications", PatientNotificationsHandler)
	mux.HandleFunc("/api/patient/telegram/status", TelegramStatusHandler)
	mux.HandleFunc("/api/patient/telegram/connect", TelegramConnectHandler)
	mux.HandleFunc("/api/patient/telegram/disconnect", TelegramDisconnectHandler)
	mux.HandleFunc("/api/telegram/request/", func(w http.ResponseWriter, r *http.Request) { TelegramSendRequestHandler(w, r, cfg, false) })
	mux.HandleFunc("/api/patient/questions", PatientDoctorQuestionsHandler)
	mux.HandleFunc("/api/doctor/questions", DoctorQuestionsHandler)
	mux.HandleFunc("/api/doctor/questions/", DoctorQuestionsHandler)
	mux.HandleFunc("/api/doctor/telegram/request/", func(w http.ResponseWriter, r *http.Request) { TelegramSendRequestHandler(w, r, cfg, true) })
	mux.HandleFunc("/api/patient/payments", PatientPaymentsHandler)
	mux.HandleFunc("/doctor/login", DoctorLoginHandler)
	mux.HandleFunc("/doctor/logout", DoctorLogoutHandler)
	mux.HandleFunc("/api/doctor/me", DoctorMeHandler)
	mux.HandleFunc("/api/doctor/requests", DoctorRequestsHandler)
	mux.HandleFunc("/api/doctor/requests/", DoctorRequestByIDHandler)
	mux.HandleFunc("/api/doctor/patients", DoctorPatientsHandler)
	mux.HandleFunc("/api/doctor/appointments", DoctorAppointmentsHandler)
	mux.HandleFunc("/api/doctor/appointments/create", DoctorCreateAppointmentHandler)
	mux.HandleFunc("/api/doctor/medical-records/", DoctorMedicalRecordHandler)
	mux.HandleFunc("/api/doctor/medical-files/", DoctorMedicalFileHandler)
	mux.HandleFunc("/api/doctor/schedule", DoctorScheduleHandler)
	mux.HandleFunc("/api/doctor/schedule/", DoctorScheduleHandler)
	mux.HandleFunc("/api/doctor/patients/", DoctorPatientByIDHandler)
	mux.HandleFunc("/api/requests", GetRequestsHandler)
	mux.HandleFunc("/api/requests/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetRequestHandler(w, r)
		case http.MethodPut:
			UpdateRequestHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/doctors", DoctorsHandler)
	mux.HandleFunc("/api/doctors/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/unavailability") || strings.Contains(r.URL.Path, "/unavailability/") {
			DoctorUnavailabilityHandler(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/photo") {
			DoctorPhotoHandler(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/availability") {
			DoctorAvailabilityHandler(w, r)
			return
		}
		DoctorHandler(w, r)
	})
	mux.HandleFunc("/api/dashboard", DashboardHandler)
	mux.HandleFunc("/api/admin/notifications", AdminNotificationsHandler)
	mux.HandleFunc("/api/admin/system-check", AdminSystemCheckHandler)
	mux.HandleFunc("/api/services", ServicesPublicHandler)
	mux.HandleFunc("/api/admin/services", ServicesAdminHandler)
	mux.HandleFunc("/api/admin/services/", ServicesAdminHandler)
	mux.HandleFunc("/api/analytics", AnalyticsHandler)
	mux.HandleFunc(
		"/api/calculator-event",
		CalculatorEventHandler,
	)
	mux.HandleFunc("/api/patients", PatientsHandler)
	mux.HandleFunc("/api/patients/", PatientHandler)
	mux.HandleFunc("/api/admin/medical-files/", AdminMedicalFileHandler)
	frontFS := http.FileServer(
		http.Dir("../front"),
	)
	mux.Handle("/politika.html", frontFS)
	mux.Handle("/politika.css", frontFS)
	mux.HandleFunc("/politika/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../front/politika.html")
	})
	mux.Handle("/patient-prefill.js", frontFS)
	mux.Handle("/services.js", frontFS)
	mux.Handle("/account/", PatientAuthMiddleware(http.StripPrefix("/account/", http.FileServer(http.Dir("../front/account")))))
	mux.Handle(
		"/img/",
		frontFS,
	)
	uploadsFS := http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsRoot())))
	mux.Handle("/uploads/", uploadsFS)

	doctorFS := http.StripPrefix("/doctor/", http.FileServer(http.Dir("../front/doctor")))
	mux.Handle("/doctor/", doctorFS)
	mux.Handle("/doctor/dashboard.html", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/dashboard.css", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/dashboard.js", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/requests.html", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/requests.css", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/requests.js", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/patients.html", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/patients.css", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/patients.js", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/book.html", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/book.css", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/book.js", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/appointments.html", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/appointments.css", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/appointments.js", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/schedule.html", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/schedule.css", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/schedule.js", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/medical-record.html", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/medical-record.css", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/medical-record.js", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/questions.html", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/questions.css", DoctorAuthMiddleware(doctorFS))
	mux.Handle("/doctor/questions.js", DoctorAuthMiddleware(doctorFS))

	mux.HandleFunc("/doctor-login.html", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/?account=doctor", http.StatusSeeOther)
	})
	mux.HandleFunc("/doctor-dashboard.html", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/doctor/dashboard.html", http.StatusFound)
	})

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
		"/doctors-admin.html",
		AuthMiddleware(frontFS),
	)
	mux.Handle(
		"/doctors-admin.js",
		AuthMiddleware(frontFS),
	)
	mux.Handle(
		"/doctors-admin.css",
		AuthMiddleware(frontFS),
	)
	mux.Handle(
		"/prices-admin.html",
		AuthMiddleware(frontFS),
	)
	mux.Handle(
		"/prices-admin.js",
		AuthMiddleware(frontFS),
	)
	mux.Handle(
		"/prices-admin.css",
		AuthMiddleware(frontFS),
	)
	mux.Handle(
		"/analytics.html",
		AuthMiddleware(frontFS),
	)
	mux.Handle(
		"/schedule-admin.html",
		AuthMiddleware(frontFS),
	)
	mux.Handle(
		"/schedule-admin.js",
		AuthMiddleware(frontFS),
	)
	mux.Handle(
		"/schedule-admin.css",
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
