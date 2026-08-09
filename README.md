# Distributed File Storage System

A lightweight file storage backend built with Go, with a simple web-based frontend for uploading, downloading, deleting, and listing files.

## 🚀 Features

- Upload files
- Download files
- Delete files
- List stored files
- Local file-system storage
- HTTP API
- Simple web frontend

## 🛠️ Tech Stack

- Backend: Go
- Frontend: HTML, CSS, JavaScript
- Storage: Local file system
- API: HTTP
- Version Control: Git & GitHub

## 📁 Project Structure

```text
distributed-file-storage/
├── cmd/
│   └── server/
│       └── main.go
├── frontend/
│   ├── index.html
│   ├── script.js
│   └── style.css
├── internal/
│   ├── handlers/
│   │   └── handler.go
│   ├── models/
│   │   └── file.go
│   ├── services/
│   │   └── file_service.go
│   └── storage/
│       └── storage.go
├── storage/
├── go.mod
├── README.md
└── .gitignore
```
