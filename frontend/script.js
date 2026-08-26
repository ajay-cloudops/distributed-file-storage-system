const LIVE_API_URL = "https://distributed-file-storage-gs0v.onrender.com";
const LOCAL_API_URL = "http://localhost:8080";

const params = new URLSearchParams(window.location.search);
const selectedStorage =
    params.get("storage") ||
    localStorage.getItem("selectedStorage") ||
    "s3";

const API_URL =
    selectedStorage === "local"
        ? LOCAL_API_URL
        : LIVE_API_URL;

const FILES_ENDPOINT =
    selectedStorage === "local"
        ? "/api/me/files/local"
        : "/api/me/files/s3";

const DELETE_ENDPOINT =
    selectedStorage === "local"
        ? "/api/me/delete/local"
        : "/api/me/delete/s3";

const USER_UPLOAD_ENDPOINT = "/api/me/upload";

function authHeaders() {
    const token = sessionStorage.getItem("dfsAccessToken");

    return {
        "Authorization": `Bearer ${token}`
    };
}

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
        const response = await fetch(`${API_URL}${USER_UPLOAD_ENDPOINT}`, {
            method: "POST",
            headers: authHeaders(),
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
        const response = await fetch(
            `${API_URL}${FILES_ENDPOINT}`,
            {
                headers: authHeaders()
            }
        );

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
                    class="download-button"
                    onclick="showVersions('${encodeURIComponent(fileName)}')"
                >
                    🕘 Versions
                </button>

                <button
                    class="delete-button"
                    onclick="deleteFile('${encodeURIComponent(fileName)}')"
                >
                    🗑 Delete
                </button>
            </div>

            <div
                class="version-panel"
                id="versions-${encodeURIComponent(fileName)}"
            ></div>
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
            `${API_URL}${DELETE_ENDPOINT}?name=${encodeURIComponent(fileName)}`,
            {
                method: "DELETE",
                headers: authHeaders()
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


async function showVersions(encodedFileName) {
    const fileName = decodeURIComponent(encodedFileName);

    const panel = document.getElementById(
        `versions-${encodeURIComponent(fileName)}`
    );

    panel.innerHTML = "Loading versions...";

    try {
        const response = await fetch(
            `${API_URL}/versions?name=${encodeURIComponent(fileName)}`
        );

        if (!response.ok) {
            throw new Error("Unable to load versions");
        }

        const versions = await response.json();

        if (versions.length === 0) {
            panel.innerHTML = "<p>No previous versions found.</p>";
            return;
        }

        panel.innerHTML = versions.map((version, index) => `
            <div class="version-item">
                <div>
                    <strong>
                        ${version.isLatest ? "Current Version" : `Version ${index + 1}`}
                    </strong>
                    <span>${escapeHTML(version.lastModified)}</span>
                </div>

                ${
                    version.isLatest
                    ? `<span class="current-version">Current</span>`
                    : `
                    <button
                        class="restore-button"
                        onclick="restoreVersion(
                            '${encodeURIComponent(fileName)}',
                            '${encodeURIComponent(version.versionId)}'
                        )"
                    >
                        Restore
                    </button>
                    `
                }
            </div>
        `).join("");

    } catch (error) {
        console.error("Version history error:", error);
        panel.innerHTML = "<p>Unable to load version history.</p>";
    }
}

async function restoreVersion(encodedFileName, encodedVersionID) {
    const fileName = decodeURIComponent(encodedFileName);
    const versionID = decodeURIComponent(encodedVersionID);

    const confirmed = confirm(
        `Restore previous version of "${fileName}"?`
    );

    if (!confirmed) {
        return;
    }

    try {
        const response = await fetch(
            `${API_URL}/restore?name=${encodeURIComponent(fileName)}&versionId=${encodeURIComponent(versionID)}`,
            {
                method: "POST"
            }
        );

        const result = await response.text();

        if (!response.ok) {
            throw new Error(result);
        }

        alert("✓ Previous version restored successfully");

        await loadFiles();
        await showVersions(encodedFileName);

    } catch (error) {
        console.error("Restore error:", error);
        alert("Unable to restore previous version.");
    }
}

// Update dashboard labels according to selected storage
function updateStorageUI() {
    const title = document.getElementById("dashboardTitle");

    if (selectedStorage === "local") {
        if (title) {
            title.textContent = "Local Storage Dashboard";
        }

        document.title = "Local Storage";

    } else {
        if (title) {
            title.textContent = "AWS S3 Storage Dashboard";
        }

        document.title = "AWS S3 Storage";
    }
}

updateStorageUI();

// Set dashboard text according to selected storage
function applySelectedStorageMode() {
    const sidebarStorage =
        document.getElementById("sidebarStorageType");

    const primaryLabel =
        document.getElementById("primaryStorageLabel");

    const primaryStatus =
        document.getElementById("primaryStorageStatus");

    const secondaryLabel =
        document.getElementById("secondaryStorageLabel");

    const secondaryStatus =
        document.getElementById("secondaryStorageStatus");

    const replicationLabel =
        document.getElementById("replicationLabel");

    const replicationStatus =
        document.getElementById("replicationStatus");

    if (selectedStorage === "local") {
        if (sidebarStorage) {
            sidebarStorage.textContent = "Local Storage";
        }

        if (primaryLabel) {
            primaryLabel.textContent = "Storage Type";
        }

        if (primaryStatus) {
            primaryStatus.textContent = "Local Device";
        }

        if (secondaryLabel) {
            secondaryLabel.textContent = "Cloud Backup";
        }

        if (secondaryStatus) {
            secondaryStatus.textContent = "AWS S3 Available";
        }

        if (replicationLabel) {
            replicationLabel.textContent = "Current View";
        }

        if (replicationStatus) {
            replicationStatus.textContent = "Local Files";
        }

    } else {
        if (sidebarStorage) {
            sidebarStorage.textContent = "AWS S3 Storage";
        }

        if (primaryLabel) {
            primaryLabel.textContent = "Storage Type";
        }

        if (primaryStatus) {
            primaryStatus.textContent = "AWS S3 Cloud";
        }

        if (secondaryLabel) {
            secondaryLabel.textContent = "Region";
        }

        if (secondaryStatus) {
            secondaryStatus.textContent = "Mumbai";
        }

        if (replicationLabel) {
            replicationLabel.textContent = "Current View";
        }

        if (replicationStatus) {
            replicationStatus.textContent = "Cloud Files";
        }
    }
}

applySelectedStorageMode();

// Sidebar storage label
const sidebarStorageType =
    document.getElementById("sidebarStorageType");

if (sidebarStorageType) {
    sidebarStorageType.textContent =
        selectedStorage === "local"
            ? "Local Storage"
            : "AWS S3 Storage";
}

// ======================================
// AUTHENTICATION GUARD
// ======================================

if (sessionStorage.getItem("dfsAuthenticated") !== "true") {
    window.location.href = "login.html";
}

// ======================================
// LOGOUT
// ======================================

function logoutUser() {
    sessionStorage.removeItem("dfsAuthenticated");
    sessionStorage.removeItem("dfsIdToken");
    sessionStorage.removeItem("dfsAccessToken");
    sessionStorage.removeItem("dfsUserEmail");

    localStorage.removeItem("selectedStorage");

    window.location.href = "login.html";
}

// ===============================
// USER PROFILE
// ===============================

async function loadUserProfile() {
    try {
        const response = await fetch(
            `${API_URL}/api/me`,
            {
                headers: authHeaders()
            }
        );

        if (!response.ok) {
            return;
        }

        const user = await response.json();

        const welcome =
            document.querySelector(".welcome");

        if (welcome && user.name) {
            welcome.textContent =
                `Welcome, ${user.name} 👋`;
        }

        sessionStorage.setItem(
            "dfsUserName",
            user.name || ""
        );

        sessionStorage.setItem(
            "dfsUserSub",
            user.sub || ""
        );

    } catch (error) {
        console.error(
            "Unable to load user profile:",
            error
        );
    }
}

loadUserProfile();
