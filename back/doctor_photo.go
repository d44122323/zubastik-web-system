package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const maxDoctorPhotoSize = 5 << 20

func DoctorPhotoHandler(w http.ResponseWriter, r *http.Request) {
	if !isAdminRequest(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/doctors/"))
	if err != nil || id <= 0 || !strings.HasSuffix(r.URL.Path, "/photo") {
		http.Error(w, "invalid doctor id", http.StatusBadRequest)
		return
	}
	if _, err := GetDoctor(id); err != nil {
		http.Error(w, "doctor not found", http.StatusNotFound)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxDoctorPhotoSize+1024*1024)
	if err := r.ParseMultipartForm(maxDoctorPhotoSize); err != nil {
		http.Error(w, "photo is too large", http.StatusRequestEntityTooLarge)
		return
	}
	file, header, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "photo is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size > maxDoctorPhotoSize {
		http.Error(w, "photo is too large", http.StatusRequestEntityTooLarge)
		return
	}

	contentType, err := detectDoctorPhotoType(file)
	if err != nil {
		http.Error(w, "invalid image", http.StatusBadRequest)
		return
	}
	ext := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}[contentType]
	if ext == "" {
		http.Error(w, "only JPG, PNG and WEBP are allowed", http.StatusBadRequest)
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "invalid image", http.StatusBadRequest)
		return
	}

	token := make([]byte, 12)
	if _, err := rand.Read(token); err != nil {
		http.Error(w, "could not create filename", http.StatusInternalServerError)
		return
	}
	name := hex.EncodeToString(token) + ext
	dir := filepath.Join(uploadsRoot(), "doctors", strconv.Itoa(id))
	if err := os.MkdirAll(dir, 0755); err != nil {
		http.Error(w, "could not create upload directory", http.StatusInternalServerError)
		return
	}
	path := filepath.Join(dir, name)
	dst, err := os.Create(path)
	if err != nil {
		http.Error(w, "could not save photo", http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(dst, io.LimitReader(file, maxDoctorPhotoSize+1)); err != nil {
		dst.Close()
		_ = os.Remove(path)
		http.Error(w, "could not save photo", http.StatusInternalServerError)
		return
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(path)
		http.Error(w, "could not save photo", http.StatusInternalServerError)
		return
	}

	photoURL := fmt.Sprintf("/uploads/doctors/%d/%s", id, name)
	if err := UpdateDoctorPhoto(id, photoURL); err != nil {
		_ = os.Remove(path)
		http.Error(w, "could not update doctor photo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]string{"photo": photoURL})
}

func detectDoctorPhotoType(file multipart.File) (string, error) {
	var buf [512]byte
	n, err := file.Read(buf[:])
	if err != nil && err != io.EOF {
		return "", err
	}
	return http.DetectContentType(buf[:n]), nil
}
