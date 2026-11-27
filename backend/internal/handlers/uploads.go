package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"backend/internal/middleware"
)

const maxUploadSize = 5 << 20 // 5 MB

// Allowed image types, detected from the file's actual bytes (not its name).
var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	middleware.EnableCors(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Hard cap on the whole request body (file + small multipart overhead).
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+(512<<10))
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			http.Error(w, "File too large (max 5 MB)", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Invalid upload", http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size > maxUploadSize {
		http.Error(w, "File too large (max 5 MB)", http.StatusRequestEntityTooLarge)
		return
	}

	// Check the real content type from the first bytes of the file.
	head := make([]byte, 512)
	n, err := io.ReadFull(file, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	ext, ok := allowedImageTypes[http.DetectContentType(head[:n])]
	if !ok {
		http.Error(w, "Invalid file type (only JPG, PNG, GIF, WEBP)", http.StatusBadRequest)
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Random, unguessable file name; the extension comes from the detected type.
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	filename := hex.EncodeToString(randBytes) + ext
	filePath := filepath.Join("uploads", filename)

	dst, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		os.Remove(filePath)
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	if err := dst.Close(); err != nil {
		os.Remove(filePath)
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	fileURL := fmt.Sprintf("/uploads/%s", filename)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"url": fileURL})
}
