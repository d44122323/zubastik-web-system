package main

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func AdminMedicalFileHandler(w http.ResponseWriter, r *http.Request) {
	if !isAdminRequest(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, err := strconv.Atoi(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/admin/medical-files/"), "/"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var name, path, typ string
	err = DB.QueryRow(`SELECT f.file_name,f.file_path,f.file_type FROM medical_files f WHERE f.id=$1`, id).Scan(&name, &path, &typ)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	clean := filepath.Clean(path)
	if clean == "." || strings.Contains(clean, ".."+string(os.PathSeparator)) {
		http.Error(w, "invalid file path", http.StatusBadRequest)
		return
	}
	f, err := os.Open(clean)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if typ != "" {
		w.Header().Set("Content-Type", typ)
	}
	safeName := strings.ReplaceAll(filepath.Base(name), `"`, `'`)
	w.Header().Set("Content-Disposition", `inline; filename="`+safeName+`"`)
	http.ServeContent(w, r, safeName, st.ModTime(), f)
}
