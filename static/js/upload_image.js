document.addEventListener("DOMContentLoaded", () => {
  const boxes = Array.from(document.querySelectorAll(".drop-box"));
  const uploadBtn = document.getElementById("uploadBtn");
  const statusEl = document.getElementById("status");
  const nameSuffixInput = document.getElementById("nameSuffix");

  function showStatus(msg, isError) {
    if (statusEl) {
      statusEl.textContent = msg;
      statusEl.style.color = isError ? "#f66" : "#9cf";
    }
    console.log("upload_image:", msg);
  }

  // disable upload if suffix < 4 chars
  function updateUploadButtonState() {
    if (!uploadBtn) return;
    const val = (nameSuffixInput && nameSuffixInput.value || "").trim();
    const ok = true; // val.length >= 4; (left as before)
    uploadBtn.disabled = !ok;
    uploadBtn.title = ok ? "" : "Enter at least 4 characters in Filename suffix";
  }

  // initial state + hook input
  updateUploadButtonState();
  if (nameSuffixInput) nameSuffixInput.addEventListener("input", updateUploadButtonState);

  // map of key -> File
  const selectedFiles = {};

  boxes.forEach((box) => {
    const key = box.dataset.key;
    const dropZone = box.querySelector(".drop-zone");
    const fileInput = box.querySelector(".file-input");
    const preview = box.querySelector(".preview");

    if (!key) return;

    function clearPreview() {
      if (preview && preview.src && preview.src.startsWith("blob:")) URL.revokeObjectURL(preview.src);
      if (preview) { preview.removeAttribute("src"); preview.style.display = "none"; }
    }

    function handleFile(file) {
      if (!file) return;
      if (!file.name.toLowerCase().endsWith(".png")) {
        showStatus("Only PNG files allowed", true);
        return;
      }
      selectedFiles[key] = file;
      if (preview) {
        clearPreview();
        preview.src = URL.createObjectURL(file);
        preview.style.display = "block";
      }
      showStatus(`${file.name} selected for ${key}`, false);
    }

    if (fileInput) {
      fileInput.addEventListener("change", (e) => {
        const f = e.target.files && e.target.files[0];
        if (f) handleFile(f);
      });
    }

    if (dropZone) {
      ["dragenter", "dragover"].forEach(evt =>
        dropZone.addEventListener(evt, (e) => { e.preventDefault(); e.stopPropagation(); dropZone.classList.add("dragover"); })
      );
      ["dragleave", "drop"].forEach(evt =>
        dropZone.addEventListener(evt, (e) => { e.preventDefault(); e.stopPropagation(); dropZone.classList.remove("dragover"); })
      );
      dropZone.addEventListener("drop", (e) => {
        const f = e.dataTransfer && e.dataTransfer.files && e.dataTransfer.files[0];
        if (f) handleFile(f);
      });
      dropZone.addEventListener("click", () => fileInput && fileInput.click());
    }
  });

  if (uploadBtn) {
    uploadBtn.addEventListener("click", async (ev) => {
      ev.preventDefault();
      const suffixRaw = (nameSuffixInput && nameSuffixInput.value) || "";
      const suffix = suffixRaw.trim();
      if (!suffix || suffix.length < 4) {
        showStatus("Filename suffix is required and must be at least 4 characters.", true);
        if (nameSuffixInput) nameSuffixInput.focus();
        updateUploadButtonState();
        return;
      }

      const entries = Object.entries(selectedFiles);
      if (entries.length === 0) {
        showStatus("No files selected to upload", true);
        return;
      }

      showStatus("Uploading...", false);
      for (const [key, file] of entries) {
        const safeSuffix = suffix.replace(/[^a-zA-Z0-9_\-\.]/g, "_");
        const filename = `${key}${safeSuffix}.png`;
        const remotePath = `/static/img/${encodeURIComponent(filename)}`;

        // Check if file already exists on server
        let exists = false;
        try {
          const headResp = await fetch(remotePath, { method: "HEAD", cache: "no-cache" });
          if (headResp.status === 200) exists = true;
          else if (headResp.status === 404) exists = false;
          else exists = headResp.ok;
        } catch (err) {
          // If HEAD fails, log and proceed to upload
          console.warn("HEAD check failed for", remotePath, err);
          exists = false;
        }

        if (exists) {
          const proceed = window.confirm(`File "${filename}" already exists on the server. Overwrite?`);
          if (!proceed) {
            showStatus(`Skipped ${filename}`, false);
            continue; // skip this file
          }
        }

        const form = new FormData();
        form.append("file", file, filename);

        try {
          const resp = await fetch("/freezy/upload/image", { method: "POST", body: form });
          if (!resp.ok) {
            const text = await resp.text().catch(() => resp.statusText);
            showStatus(`Upload failed for ${filename}: ${resp.status} ${text}`, true);
            return;
          } else {
            showStatus(`Uploaded ${filename}`, false);
          }
        } catch (err) {
          console.error("upload error", err);
          showStatus(`Network error uploading ${filename}`, true);
          return;
        }
      }

      // cleanup
      boxes.forEach((box) => {
        const preview = box.querySelector(".preview");
        if (preview && preview.src && preview.src.startsWith("blob:")) URL.revokeObjectURL(preview.src);
        if (preview) { preview.removeAttribute("src"); preview.style.display = "none"; }
      });
      for (const k of Object.keys(selectedFiles)) delete selectedFiles[k];

      showStatus("All uploads complete", false);
    });
  }
});