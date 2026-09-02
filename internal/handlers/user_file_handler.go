package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	appauth "distributed-file-storage/internal/auth"
	"distributed-file-storage/internal/services"
)

func UserProfileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := appauth.UserFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func UserUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	user, err := appauth.UserFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	location, err := services.SaveUserFile(
		user,
		header.Filename,
		file,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, location)
}

func UserLocalFilesHandler(w http.ResponseWriter, r *http.Request) {
	user, err := appauth.UserFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	files, err := services.ListUserLocalFiles(user.Sub)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func UserS3FilesHandler(w http.ResponseWriter, r *http.Request) {
	user, err := appauth.UserFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	files, err := services.ListUserS3Files(user.Sub)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func UserDeleteLocalHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "DELETE required", http.StatusMethodNotAllowed)
		return
	}

	user, err := appauth.UserFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fileName := filepath.Base(r.URL.Query().Get("name"))

	if fileName == "." || fileName == "" {
		http.Error(w, "File name required", http.StatusBadRequest)
		return
	}

	if err := services.DeleteUserLocalFile(user, fileName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Local copy deleted")
}

func UserDeleteS3Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "DELETE required", http.StatusMethodNotAllowed)
		return
	}

	user, err := appauth.UserFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fileName := filepath.Base(r.URL.Query().Get("name"))

	if fileName == "." || fileName == "" {
		http.Error(w, "File name required", http.StatusBadRequest)
		return
	}

	if err := services.DeleteUserS3File(user, fileName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "S3 file moved to recycle bin")
}

func AdminFilesHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := appauth.AdminFromRequest(r); err != nil {
		http.Error(w, "Admin access required", http.StatusUnauthorized)
		return
	}

	files, err := services.AdminListAllFiles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func AdminDeletedFilesHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := appauth.AdminFromRequest(r); err != nil {
		http.Error(w, "Admin access required", http.StatusUnauthorized)
		return
	}

	files, err := services.AdminListDeletedFiles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func AdminRestoreDeletedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	if _, err := appauth.AdminFromRequest(r); err != nil {
		http.Error(w, "Admin access required", http.StatusUnauthorized)
		return
	}

	key := r.URL.Query().Get("key")

	if key == "" {
		http.Error(w, "Deleted file key required", http.StatusBadRequest)
		return
	}

	if err := services.AdminRestoreDeletedFile(key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "File restored successfully")
}

func AdminBucketFilesHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := appauth.AdminFromRequest(r); err != nil {
		http.Error(w, "Admin access required", http.StatusUnauthorized)
		return
	}

	files, err := services.AdminListBucketObjects()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func AdminDeleteBucketObjectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "DELETE required", http.StatusMethodNotAllowed)
		return
	}

	admin, err := appauth.AdminFromRequest(r)
	if err != nil {
		http.Error(w, "Admin access required", http.StatusUnauthorized)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "File key required", http.StatusBadRequest)
		return
	}

	if err := services.AdminDeleteBucketObject(admin, key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "File moved to recovery bin")
}
