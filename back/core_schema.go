package main

import "os"

// InitCoreSchema creates the base tables required before the feature-specific
// schema initializers run. A fresh Railway PostgreSQL database starts empty,
// so feature migrations such as ALTER TABLE requests must never run before
// these base tables exist.
func InitCoreSchema() error {
	_, err := DB.Exec(`
CREATE TABLE IF NOT EXISTS patients (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL DEFAULT '',
    phone VARCHAR(100) NOT NULL DEFAULT '',
    comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_patients_phone ON patients(phone);

CREATE TABLE IF NOT EXISTS requests (
    id SERIAL PRIMARY KEY,
    patient_id INTEGER NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    services TEXT NOT NULL DEFAULT '',
    price INTEGER NOT NULL DEFAULT 0 CHECK (price >= 0),
    source VARCHAR(255) NOT NULL DEFAULT '',
    appointment_comment TEXT NOT NULL DEFAULT '',
    appointment_date DATE,
    appointment_time TIME,
    status VARCHAR(100) NOT NULL DEFAULT 'Новая',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_requests_patient_id ON requests(patient_id);
CREATE INDEX IF NOT EXISTS idx_requests_created_at ON requests(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_requests_status ON requests(status);
CREATE INDEX IF NOT EXISTS idx_requests_appointment_date ON requests(appointment_date);

CREATE TABLE IF NOT EXISTS ai_requests (
    id BIGSERIAL PRIMARY KEY,
    question TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ai_requests_created_at ON ai_requests(created_at DESC);

CREATE TABLE IF NOT EXISTS calculator_events (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_calculator_events_created_at ON calculator_events(created_at DESC);

CREATE TABLE IF NOT EXISTS admins (
    id SERIAL PRIMARY KEY,
    login VARCHAR(100) NOT NULL UNIQUE,
    password TEXT NOT NULL
);
`)
	if err != nil {
		return err
	}

	// A fresh production database has no administrator row. If the operator
	// provides ADMIN_LOGIN and ADMIN_PASSWORD, create the first admin once.
	// Existing administrator rows are never overwritten.
	login := os.Getenv("ADMIN_LOGIN")
	password := os.Getenv("ADMIN_PASSWORD")
	if login != "" && password != "" {
		var exists bool
		if err := DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM admins)`).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			if _, err := DB.Exec(`INSERT INTO admins(login,password) VALUES($1,$2)`, login, password); err != nil {
				return err
			}
		}
	}
	return nil
}
