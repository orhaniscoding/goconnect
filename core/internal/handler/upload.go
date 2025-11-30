package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadHandler handles file uploads
type UploadHandler struct {
	uploadDir string
	baseURL   string
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(uploadDir, baseURL string) *UploadHandler {
	// Ensure upload directory exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		fmt.Printf("Failed to create upload directory: %v\n", err)
	}

	return &UploadHandler{
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}
}

// UploadFile handles POST /v1/uploads
func (h *UploadHandler) UploadFile(c *gin.Context) {
	// Limit file size (e.g., 10MB)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_INVALID_FILE",
			"message": "Invalid file upload",
		})
		return
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".pdf": true, ".txt": true, ".doc": true, ".docx": true,
		".zip": true,
	}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_INVALID_FILE_TYPE",
			"message": "File type not allowed",
		})
		return
	}

	// Generate unique filename
	newFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dst := filepath.Join(h.uploadDir, newFilename)

	// Save file
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL_SERVER",
			"message": "Failed to save file",
		})
		return
	}

	// Return URL
	url := fmt.Sprintf("%s/%s", h.baseURL, newFilename)
	c.JSON(http.StatusOK, gin.H{
		"url":      url,
		"filename": file.Filename,
		"size":     file.Size,
	})
}

// RegisterUploadRoutes registers upload routes
func RegisterUploadRoutes(r *gin.Engine, handler *UploadHandler, authMiddleware gin.HandlerFunc) {
	uploads := r.Group("/v1/uploads")
	uploads.Use(authMiddleware)
	uploads.POST("", handler.UploadFile)
}
