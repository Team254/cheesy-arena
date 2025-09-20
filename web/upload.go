package web

import (
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
	"github.com/Team254/cheesy-arena/model"
)

// Serves the upload image page.
func (web *Web) uploadImagePageHandler(w http.ResponseWriter, r *http.Request) {
    //http.ServeFile(w, r, filepath.Join(model.BaseDir, "templates", "upload_image.html"))
    template, err := web.parseFiles("templates/upload_image.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	data := struct {
		*model.EventSettings
	}{web.arena.EventSettings} 
	err = template.ExecuteTemplate(w, "base", data)
    
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Upload imag.png via multipart/form-data field "file". Saved to static/images/imag.png.
func (web *Web) uploadImagePostHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    if err := r.ParseMultipartForm(20 << 20); err != nil { // 20 MB
        http.Error(w, "Invalid multipart form", http.StatusBadRequest)
        return
    }

    f, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Missing file field 'file'", http.StatusBadRequest)
        return
    }
    defer f.Close()

    filename := filepath.Base(header.Filename)
    if !strings.HasSuffix(strings.ToLower(filename), ".png") {
        http.Error(w, "Only PNG files allowed", http.StatusBadRequest)
        return
    }

    // Enforce filename pattern and require non-empty suffix of at least 4 characters.
    nameNoExt := strings.TrimSuffix(filename, filepath.Ext(filename))
    allowedBases := []string{"alliance-station-logo", "blinds-logo", "game-logo"}
    matched := false
    for _, base := range allowedBases {
        if strings.HasPrefix(nameNoExt, base) {
            matched = true
            suffix := strings.TrimPrefix(nameNoExt, base)
            if len(suffix) < 4 {
                http.Error(w, "Filename suffix is required and must be at least 4 characters", http.StatusBadRequest)
                return
            }
            break
        }
    }
    if !matched {
        http.Error(w, "Filename must start with one of: alliance-station-logo, blinds-logo, game-logo", http.StatusBadRequest)
        return
    }

    dstDir := filepath.Join("static", "img")
    if err := os.MkdirAll(dstDir, 0755); err != nil {
        http.Error(w, "Unable to create images directory", http.StatusInternalServerError)
        return
    }

    dstPath := filepath.Join(dstDir, filename)
    out, err := os.Create(dstPath)
    if err != nil {
        http.Error(w, "Unable to create destination file", http.StatusInternalServerError)
        return
    }
    defer out.Close()

    if _, err := io.Copy(out, f); err != nil {
        http.Error(w, "Failed to save file", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, `{"ok": true, "path": "/static/img/%s"}`, filename)
}