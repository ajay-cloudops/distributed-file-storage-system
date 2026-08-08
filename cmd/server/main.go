package main

import (
	"fmt"
	"log"
	"net/http"

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

	// Start server
	fmt.Println("Server running on http://localhost:8080")

	log.Fatal(
		http.ListenAndServe(
			":8080",
			enableCORS(http.DefaultServeMux),
		),
	)
}
