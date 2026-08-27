package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type MedicalRecord struct {
	ID              int           `json:"id"`
	AppointmentID   int           `json:"appointment_id"`
	PatientID       int           `json:"patient_id"`
	DoctorID        int           `json:"doctor_id"`
	AppointmentDate string        `json:"appointment_date,omitempty"`
	AppointmentTime string        `json:"appointment_time,omitempty"`
	PatientName     string        `json:"patient_name,omitempty"`
	Complaints      string        `json:"complaints"`
	Diagnosis       string        `json:"diagnosis"`
	Treatment       string        `json:"treatment"`
	Recommendations string        `json:"recommendations"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	Files           []MedicalFile `json:"files,omitempty"`
}

type MedicalFile struct {
	ID              int       `json:"id"`
	MedicalRecordID int       `json:"medical_record_id"`
	FileName        string    `json:"file_name"`
	FileType        string    `json:"file_type"`
	FileSize        int64     `json:"file_size"`
	CreatedAt       time.Time `json:"created_at"`
}

func InitMedicalRecordsSchema() error {
	_, err := DB.Exec(`
        CREATE TABLE IF NOT EXISTS medical_records (
            id SERIAL PRIMARY KEY,
            appointment_id INTEGER NOT NULL UNIQUE REFERENCES requests(id) ON DELETE CASCADE,
            patient_id INTEGER NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
            doctor_id INTEGER NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
            complaints TEXT NOT NULL DEFAULT '',
            diagnosis TEXT NOT NULL DEFAULT '',
            treatment TEXT NOT NULL DEFAULT '',
            recommendations TEXT NOT NULL DEFAULT '',
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW()
        );
        CREATE INDEX IF NOT EXISTS idx_medical_records_patient_doctor
            ON medical_records(patient_id, doctor_id, created_at DESC);
        CREATE TABLE IF NOT EXISTS medical_files (
            id SERIAL PRIMARY KEY,
            medical_record_id INTEGER NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
            file_name TEXT NOT NULL,
            file_path TEXT NOT NULL,
            file_type TEXT NOT NULL DEFAULT '',
            file_size BIGINT NOT NULL DEFAULT 0,
            uploaded_by INTEGER NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW()
        );
        CREATE INDEX IF NOT EXISTS idx_medical_files_record
            ON medical_files(medical_record_id, created_at DESC);
        ALTER TABLE medical_records ADD COLUMN IF NOT EXISTS complaints TEXT NOT NULL DEFAULT '';
        ALTER TABLE medical_records ADD COLUMN IF NOT EXISTS diagnosis TEXT NOT NULL DEFAULT '';
        ALTER TABLE medical_records ADD COLUMN IF NOT EXISTS treatment TEXT NOT NULL DEFAULT '';
        ALTER TABLE medical_records ADD COLUMN IF NOT EXISTS recommendations TEXT NOT NULL DEFAULT '';
        ALTER TABLE medical_records ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT NOW();
        ALTER TABLE medical_records ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT NOW();
        ALTER TABLE medical_files ADD COLUMN IF NOT EXISTS file_name TEXT NOT NULL DEFAULT '';
        ALTER TABLE medical_files ADD COLUMN IF NOT EXISTS file_path TEXT NOT NULL DEFAULT '';
        ALTER TABLE medical_files ADD COLUMN IF NOT EXISTS file_type TEXT NOT NULL DEFAULT '';
        ALTER TABLE medical_files ADD COLUMN IF NOT EXISTS file_size BIGINT NOT NULL DEFAULT 0;
        ALTER TABLE medical_files ADD COLUMN IF NOT EXISTS uploaded_by INTEGER NOT NULL DEFAULT 0;
        ALTER TABLE medical_files ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT NOW();
    `)
	if err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(uploadsRoot(), "medical"), 0755)
}

func getDoctorAppointmentForRecord(doctorID, appointmentID int) (Request, error) {
	return GetDoctorRequest(doctorID, appointmentID)
}

func GetMedicalRecordByAppointment(doctorID, appointmentID int) (MedicalRecord, error) {
	var m MedicalRecord
	err := DB.QueryRow(`
        SELECT mr.id, mr.appointment_id, mr.patient_id, mr.doctor_id,
               COALESCE(TO_CHAR(r.appointment_date,'YYYY-MM-DD'),''),
               COALESCE(TO_CHAR(r.appointment_time,'HH24:MI'),''),
               COALESCE(p.name,''), mr.complaints, mr.diagnosis, mr.treatment,
               mr.recommendations, mr.created_at, mr.updated_at
        FROM medical_records mr
        JOIN requests r ON r.id=mr.appointment_id
        JOIN patients p ON p.id=mr.patient_id
        WHERE mr.doctor_id=$1 AND mr.appointment_id=$2`, doctorID, appointmentID).
		Scan(&m.ID, &m.AppointmentID, &m.PatientID, &m.DoctorID, &m.AppointmentDate,
			&m.AppointmentTime, &m.PatientName, &m.Complaints, &m.Diagnosis,
			&m.Treatment, &m.Recommendations, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return m, err
	}
	m.Files, err = getMedicalFiles(doctorID, m.ID)
	return m, err
}

func getMedicalFiles(doctorID, recordID int) ([]MedicalFile, error) {
	rows, err := DB.Query(`
        SELECT f.id, f.medical_record_id, f.file_name, f.file_type, f.file_size, f.created_at
        FROM medical_files f
        JOIN medical_records mr ON mr.id=f.medical_record_id
        WHERE f.medical_record_id=$1 AND mr.doctor_id=$2
        ORDER BY f.created_at DESC`, recordID, doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]MedicalFile, 0)
	for rows.Next() {
		var f MedicalFile
		if err := rows.Scan(&f.ID, &f.MedicalRecordID, &f.FileName, &f.FileType, &f.FileSize, &f.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func UpsertMedicalRecord(doctorID, appointmentID int, complaints, diagnosis, treatment, recommendations string) (MedicalRecord, error) {
	appt, err := getDoctorAppointmentForRecord(doctorID, appointmentID)
	if err == sql.ErrNoRows {
		return MedicalRecord{}, errors.New("запись не найдена")
	}
	if err != nil {
		return MedicalRecord{}, err
	}
	if appt.Status != "Подтверждена" && appt.Status != "Завершена" {
		return MedicalRecord{}, errors.New("медицинскую карту можно заполнить только для подтверждённого или завершённого приёма")
	}
	tx, err := DB.Begin()
	if err != nil {
		return MedicalRecord{}, err
	}
	defer tx.Rollback()
	var id int
	err = tx.QueryRow(`
        INSERT INTO medical_records
            (appointment_id, patient_id, doctor_id, complaints, diagnosis, treatment, recommendations)
        VALUES ($1,$2,$3,$4,$5,$6,$7)
        ON CONFLICT (appointment_id) DO UPDATE SET
            complaints=EXCLUDED.complaints,
            diagnosis=EXCLUDED.diagnosis,
            treatment=EXCLUDED.treatment,
            recommendations=EXCLUDED.recommendations,
            updated_at=NOW()
        RETURNING id`, appointmentID, appt.PatientID, doctorID, strings.TrimSpace(complaints),
		strings.TrimSpace(diagnosis), strings.TrimSpace(treatment), strings.TrimSpace(recommendations)).Scan(&id)
	if err != nil {
		return MedicalRecord{}, err
	}
	if appt.Status == "Подтверждена" {
		if _, err = tx.Exec(`UPDATE requests SET status='Завершена' WHERE id=$1 AND doctor_id=$2`, appointmentID, doctorID); err != nil {
			return MedicalRecord{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return MedicalRecord{}, err
	}
	// Запись переводится в "Завершена" именно после сохранения медицинской карты.
	// Уведомляем пациента один раз в момент завершения приёма, без передачи
	// медицинских подробностей (диагноза/лечения) в Telegram.
	if appt.Status == "Подтверждена" {
		go telegramStatusNotification(LoadConfig(), appointmentID, "Завершена")
	}
	return GetMedicalRecordByAppointment(doctorID, appointmentID)
}

func saveMedicalUpload(doctorID, recordID int, fh *multipart.FileHeader) (MedicalFile, error) {
	if fh == nil || fh.Size <= 0 {
		return MedicalFile{}, errors.New("пустой файл")
	}
	const maxSize = 15 << 20
	if fh.Size > maxSize {
		return MedicalFile{}, errors.New("файл слишком большой: максимум 15 МБ")
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	allowed := map[string]bool{".pdf": true, ".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".doc": true, ".docx": true}
	if !allowed[ext] {
		return MedicalFile{}, errors.New("тип файла не поддерживается")
	}
	safe := filepath.Base(fh.Filename)
	safe = strings.ReplaceAll(safe, "\\", "_")
	safe = strings.ReplaceAll(safe, "/", "_")
	if safe == "." || safe == ".." || safe == "" {
		safe = "file" + ext
	}
	dir := filepath.Join(uploadsRoot(), "medical", strconv.Itoa(doctorID), strconv.Itoa(recordID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return MedicalFile{}, err
	}
	finalName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), safe)
	path := filepath.Join(dir, finalName)
	src, err := fh.Open()
	if err != nil {
		return MedicalFile{}, err
	}
	defer src.Close()
	dst, err := os.Create(path)
	if err != nil {
		return MedicalFile{}, err
	}
	_, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()
	if copyErr != nil {
		os.Remove(path)
		return MedicalFile{}, copyErr
	}
	if closeErr != nil {
		os.Remove(path)
		return MedicalFile{}, closeErr
	}
	var f MedicalFile
	err = DB.QueryRow(`
        INSERT INTO medical_files(medical_record_id,file_name,file_path,file_type,file_size,uploaded_by)
        VALUES($1,$2,$3,$4,$5,$6)
        RETURNING id,medical_record_id,file_name,file_type,file_size,created_at`,
		recordID, safe, path, fh.Header.Get("Content-Type"), fh.Size, doctorID).
		Scan(&f.ID, &f.MedicalRecordID, &f.FileName, &f.FileType, &f.FileSize, &f.CreatedAt)
	if err != nil {
		os.Remove(path)
		return MedicalFile{}, err
	}
	return f, nil
}

func DoctorMedicalRecordHandler(w http.ResponseWriter, r *http.Request) {
	u, err := DoctorRequestGuard(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/doctor/medical-records/")
	appointmentID, err := strconv.Atoi(strings.Trim(idStr, "/"))
	if err != nil || appointmentID <= 0 {
		http.Error(w, "invalid appointment id", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		rec, err := GetMedicalRecordByAppointment(u.DoctorID, appointmentID)
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, rec)
	case http.MethodPost:
		if err := r.ParseMultipartForm(20 << 20); err != nil {
			http.Error(w, "invalid multipart form", 400)
			return
		}
		rec, err := UpsertMedicalRecord(u.DoctorID, appointmentID, r.FormValue("complaints"), r.FormValue("diagnosis"), r.FormValue("treatment"), r.FormValue("recommendations"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		files := r.MultipartForm.File["files"]
		for _, fh := range files {
			if _, err := saveMedicalUpload(u.DoctorID, rec.ID, fh); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
		}
		rec, err = GetMedicalRecordByAppointment(u.DoctorID, appointmentID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, rec)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func DoctorMedicalFileHandler(w http.ResponseWriter, r *http.Request) {
	u, err := DoctorRequestGuard(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 405)
		return
	}
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/doctor/medical-files/"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", 400)
		return
	}
	var name, path, typ string
	err = DB.QueryRow(`SELECT f.file_name,f.file_path,f.file_type FROM medical_files f JOIN medical_records mr ON mr.id=f.medical_record_id WHERE f.id=$1 AND mr.doctor_id=$2`, id, u.DoctorID).Scan(&name, &path, &typ)
	if err == sql.ErrNoRows {
		http.Error(w, "not found", 404)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		http.Error(w, "file unavailable", 404)
		return
	}
	defer f.Close()
	if typ != "" {
		w.Header().Set("Content-Type", typ)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Header().Set("Content-Disposition", `inline; filename="`+strings.ReplaceAll(name, `"`, "")+`"`)
	http.ServeContent(w, r, name, time.Time{}, f)
}
