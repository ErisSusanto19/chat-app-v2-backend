package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/service"
)

type UploadHandler struct {
	uploadService service.UploadService
}

func NewUploadHandler(uploadService service.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

func (h *UploadHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "The uploaded file is too big. Please choose an image that is less than 10MB in size.", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file key in request. Please use 'file' key.", http.StatusBadRequest)
		return
	}
	defer file.Close()

	url, err := h.uploadService.UploadImage(r.Context(), file, fileHeader)
	if err != nil {
		http.Error(w, "Failed to upload image.", http.StatusInternalServerError)
		return
	}

	response := map[string]string{"url": url}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
