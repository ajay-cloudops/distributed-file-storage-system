const ADMIN_API =
    "https://distributed-file-storage-gs0v.onrender.com";

function adminHeaders() {
    return {
        "Authorization":
            `Bearer ${sessionStorage.getItem("dfsAdminAccessToken")}`
    };
}

async function loadAdminFiles() {
    const response = await fetch(
        `${ADMIN_API}/api/admin/files`,
        { headers: adminHeaders() }
    );

    if (!response.ok) {
        throw new Error(await response.text());
    }

    return response.json();
}

async function loadDeletedFiles() {
    const response = await fetch(
        `${ADMIN_API}/api/admin/deleted`,
        { headers: adminHeaders() }
    );

    if (!response.ok) {
        throw new Error(await response.text());
    }

    return response.json();
}

async function restoreDeletedFile(key) {
    if (!confirm("Restore this deleted file?")) return;

    const response = await fetch(
        `${ADMIN_API}/api/admin/restore?key=${encodeURIComponent(key)}`,
        {
            method: "POST",
            headers: adminHeaders()
        }
    );

    if (!response.ok) {
        alert("Restore failed: " + await response.text());
        return;
    }

    alert("File restored successfully ✓");
    await loadAdminDashboard();
}

function renderFiles(files) {
    const panel = document.querySelector("#users .empty");

    document.getElementById("totalFiles").textContent = files.length;

    const users = new Set(files.map(file => file.ownerSub));
    document.getElementById("totalUsers").textContent = users.size;

    if (!files.length) {
        panel.innerHTML = "No user files found.";
        return;
    }

    panel.innerHTML = files.map(file => `
        <div class="admin-file-row">
            <div>
                <strong>${escapeAdmin(file.fileName)}</strong>
                <small>👤 ${escapeAdmin(file.ownerName || file.ownerEmail || "User")}</small>
                <small>${escapeAdmin(file.ownerEmail || "")}</small>
            </div>
            <div class="bucket-actions">
                <span>${formatBytes(file.size)}</span>
                <button
                    class="admin-delete-button"
                    onclick='deleteBucketFile(${JSON.stringify(file.key)})'
                >
                    🗑 Delete
                </button>
            </div>
        </div>
    `).join("");
}

function renderDeleted(files) {
    const panel = document.querySelector("#deleted .empty");

    document.getElementById("deletedFiles").textContent = files.length;

    if (!files.length) {
        panel.innerHTML = "No deleted files found.";
        return;
    }

    panel.innerHTML = files.map(file => `
        <div class="deleted-file-card">
            <div class="deleted-main">
                <strong>📄 ${escapeAdmin(file.fileName)}</strong>

                <span class="user-badge">
                    👤 ${escapeAdmin(file.ownerName || file.ownerEmail || "Unknown")}
                </span>

                <small>Owner: ${escapeAdmin(file.ownerEmail || "")}</small>
                <small>Deleted by: ${escapeAdmin(file.deletedBy || "Unknown")}</small>
                <small>Deleted: ${formatDate(file.deletedAt)}</small>
                <small>Size: ${formatBytes(file.size)}</small>
            </div>

            <button
                class="admin-restore-button"
                onclick='restoreDeletedFile(${JSON.stringify(file.key)})'
            >
                ↻ Restore
            </button>
        </div>
    `).join("");
}


async function loadBucketFiles() {
    const response = await fetch(
        `${ADMIN_API}/api/admin/bucket-files`,
        { headers: adminHeaders() }
    );

    if (!response.ok) {
        throw new Error(await response.text());
    }

    return response.json();
}


async function deleteBucketFile(key) {
    if (!confirm("Delete this S3 file? It will be moved to Recovery.")) {
        return;
    }

    const response = await fetch(
        `${ADMIN_API}/api/admin/delete-bucket-file?key=${encodeURIComponent(key)}`,
        {
            method: "DELETE",
            headers: adminHeaders()
        }
    );

    if (!response.ok) {
        alert("Delete failed: " + await response.text());
        return;
    }

    alert("File moved to Recovery ✓");
    await loadAdminDashboard();
}

function renderBucketFiles(files) {
    const panel = document.getElementById("s3Files");

    if (!files.length) {
        panel.innerHTML = "No S3 files found.";
        return;
    }

    panel.innerHTML = files.map(file => `
        <div class="admin-file-row">
            <div>
                <strong>${escapeAdmin(file.fileName)}</strong>
                <small>${escapeAdmin(file.key)}</small>
            </div>
            <span>${formatBytes(file.size)}</span>
        </div>
    `).join("");
}

async function loadAdminDashboard() {
    try {
        const [files, deleted, bucketFiles] = await Promise.all([
            loadAdminFiles(),
            loadDeletedFiles(),
            loadBucketFiles()
        ]);

        renderFiles(files);
        renderDeleted(deleted);
        renderBucketFiles(bucketFiles);

    } catch (error) {
        console.error(error);
        alert("Unable to load admin data: " + error.message);
    }
}

function escapeAdmin(value) {
    return String(value || "")
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;")
        .replaceAll("'", "&#039;");
}

function formatBytes(bytes) {
    if (!bytes) return "0 B";
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;

    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function formatDate(value) {
    if (!value) return "-";
    return new Date(value).toLocaleString();
}

loadAdminDashboard();
