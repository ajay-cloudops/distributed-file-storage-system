package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"distributed-file-storage/internal/handlers"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Allow frontend to communicate with backend
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle browser preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {

	// Home / status route
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Distributed File Storage System Backend is running!")
	})

	// File upload
	http.HandleFunc("/upload", handlers.UploadHandler)

	// File download
	http.HandleFunc("/download", handlers.DownloadHandler)

	// File delete
	http.HandleFunc("/delete", handlers.DeleteHandler)

	// File list
	http.HandleFunc("/files", handlers.ListFilesHandler)

	// Separate Local and AWS S3 storage views
	http.HandleFunc("/files/local", handlers.ListLocalFilesHandler)
	http.HandleFunc("/files/s3", handlers.ListS3FilesHandler)

	// Independent delete operations
	http.HandleFunc("/delete/local", handlers.DeleteLocalFileHandler)
	http.HandleFunc("/delete/s3", handlers.DeleteS3FileHandler)

	// File version history
	http.HandleFunc("/versions", handlers.VersionsHandler)

	// Restore previous file version
	http.HandleFunc("/restore", handlers.RestoreVersionHandler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server running on port %s\n", port)

	log.Fatal(
		http.ListenAndServe(
			":"+port,
			enableCORS(http.DefaultServeMux),
		),
	)
}
