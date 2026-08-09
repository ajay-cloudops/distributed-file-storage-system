const API_URL = "https://distributed-file-storage-gs0v.onrender.com";

const fileInput = document.getElementById("fileInput");
const uploadButton = document.getElementById("uploadButton");
const selectedFile = document.getElementById("selectedFile");
const fileList = document.getElementById("fileList");
const totalFiles = document.getElementById("totalFiles");
const refreshButton = document.getElementById("refreshButton");

// Choose File button
uploadButton.addEventListener("click", () => {
    fileInput.click();
});

// File selected
fileInput.addEventListener("change", () => {
    if (fileInput.files.length === 0) {
        selectedFile.textContent = "No file selected";
        return;
    }

    selectedFile.textContent = fileInput.files[0].name;

    uploadFile(fileInput.files[0]);
});

// Upload file
async function uploadFile(file) {
    const formData = new FormData();
    formData.append("file", file);

    selectedFile.textContent = `Uploading ${file.name}...`;

    try {
        const response = await fetch(`${API_URL}/upload`, {
            method: "POST",
            body: formData
        });

        const result = await response.text();

        if (!response.ok) {
            throw new Error(result);
        }

        selectedFile.textContent = "✓ File uploaded successfully";

        fileInput.value = "";

        await loadFiles();

    } catch (error) {
        console.error("Upload error:", error);

        selectedFile.textContent = "✕ Upload failed";
        alert("File upload failed. Please try again.");
    }
}

// Load files
async function loadFiles() {
    try {
        const response = await fetch(`${API_URL}/files`);

        if (!response.ok) {
            throw new Error("Failed to load files");
        }

        const files = await response.json();

        displayFiles(files);

    } catch (error) {
        console.error("File list error:", error);

        fileList.innerHTML = `
            <div class="empty-state">
                <div class="empty-icon">⚠️</div>
                <h3>Unable to load files</h3>
                <p>Make sure the Go backend is running.</p>
            </div>
        `;

        totalFiles.textContent = "0";
    }
}

// Display files
function displayFiles(files) {
    totalFiles.textContent = files.length;

    if (files.length === 0) {
        fileList.innerHTML = `
            <div class="empty-state">
                <div class="empty-icon">📂</div>
                <h3>No files found</h3>
                <p>Upload your first file to get started.</p>
            </div>
        `;

        return;
    }

    fileList.innerHTML = "";

    files.forEach((fileName) => {
        const fileItem = document.createElement("div");

        fileItem.className = "file-item";

        fileItem.innerHTML = `
            <div class="file-info">
                <div class="file-icon">📄</div>

                <div>
                    <div class="file-name">
                        ${escapeHTML(fileName)}
                    </div>
                </div>
            </div>

            <div class="file-actions">
                <button
                    class="download-button"
                    onclick="downloadFile('${encodeURIComponent(fileName)}')"
                >
                    ↓ Download
                </button>

                <button
                    class="delete-button"
                    onclick="deleteFile('${encodeURIComponent(fileName)}')"
                >
                    🗑 Delete
                </button>
            </div>
        `;

        fileList.appendChild(fileItem);
    });
}

// Download file
function downloadFile(encodedFileName) {
    const fileName = decodeURIComponent(encodedFileName);

    const downloadURL =
        `${API_URL}/download?name=${encodeURIComponent(fileName)}`;

    window.open(downloadURL, "_blank");
}

// Delete file
async function deleteFile(encodedFileName) {
    const fileName = decodeURIComponent(encodedFileName);

    const confirmed = confirm(
        `Are you sure you want to delete "${fileName}"?`
    );

    if (!confirmed) {
        return;
    }

    try {
        const response = await fetch(
            `${API_URL}/delete?name=${encodeURIComponent(fileName)}`,
            {
                method: "DELETE"
            }
        );

        const result = await response.text();

        if (!response.ok) {
            throw new Error(result);
        }

        await loadFiles();

    } catch (error) {
        console.error("Delete error:", error);

        alert("Unable to delete the file.");
    }
}

// Refresh files
refreshButton.addEventListener("click", () => {
    loadFiles();
});

// Basic HTML escaping
function escapeHTML(value) {
    return value
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;")
        .replaceAll("'", "&#039;");
}

// Load files when page opens
loadFiles();
