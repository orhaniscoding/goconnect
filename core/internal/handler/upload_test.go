package handler

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupUploadTest creates a test environment for upload handler
func setupUploadTest(t *testing.T) (*gin.Engine, *UploadHandler, string) {
	gin.SetMode(gin.TestMode)

	// Create temp upload directory
	tempDir, err := os.MkdirTemp("", "upload_test_*")
	require.NoError(t, err)

	handler := NewUploadHandler(tempDir, "http://localhost:8080/uploads")
	r := gin.New()

	return r, handler, tempDir
}

// uploadAuthMiddleware returns a test auth middleware
func uploadAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		if token == "valid-token" {
			c.Set("user_id", "user1")
			c.Set("tenant_id", "t1")
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	}
}

// createMultipartForm creates a multipart form with a file
func createMultipartForm(t *testing.T, fieldName, filename, contentType string, content []byte) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, filename)
	require.NoError(t, err)

	_, err = io.Copy(part, bytes.NewReader(content))
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	return body, writer.FormDataContentType()
}

// ==================== UPLOAD FILE TESTS ====================

func TestUploadHandler_UploadFile(t *testing.T) {
	t.Run("Success - JPG Image", func(t *testing.T) {
		r, handler, tempDir := setupUploadTest(t)
		defer os.RemoveAll(tempDir)

		r.POST("/v1/uploads", uploadAuthMiddleware(), handler.UploadFile)

		// Create a fake JPG file content
		content := []byte("fake image content")
		body, contentType := createMultipartForm(t, "file", "test.jpg", "image/jpeg", content)

		req := httptest.NewRequest("POST", "/v1/uploads", body)
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify response contains URL
		assert.Contains(t, w.Body.String(), "url")
		assert.Contains(t, w.Body.String(), "filename")
	})

	t.Run("Success - PNG Image", func(t *testing.T) {
		r, handler, tempDir := setupUploadTest(t)
		defer os.RemoveAll(tempDir)

		r.POST("/v1/uploads", uploadAuthMiddleware(), handler.UploadFile)

		content := []byte("fake png content")
		body, contentType := createMultipartForm(t, "file", "image.png", "image/png", content)

		req := httptest.NewRequest("POST", "/v1/uploads", body)
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Success - PDF Document", func(t *testing.T) {
		r, handler, tempDir := setupUploadTest(t)
		defer os.RemoveAll(tempDir)

		r.POST("/v1/uploads", uploadAuthMiddleware(), handler.UploadFile)

		content := []byte("fake pdf content")
		body, contentType := createMultipartForm(t, "file", "document.pdf", "application/pdf", content)

		req := httptest.NewRequest("POST", "/v1/uploads", body)
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid File Type", func(t *testing.T) {
		r, handler, tempDir := setupUploadTest(t)
		defer os.RemoveAll(tempDir)

		r.POST("/v1/uploads", uploadAuthMiddleware(), handler.UploadFile)

		content := []byte("fake exe content")
		body, contentType := createMultipartForm(t, "file", "malware.exe", "application/octet-stream", content)

		req := httptest.NewRequest("POST", "/v1/uploads", body)
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "ERR_INVALID_FILE_TYPE")
	})

	t.Run("Invalid File Type - Script", func(t *testing.T) {
		r, handler, tempDir := setupUploadTest(t)
		defer os.RemoveAll(tempDir)

		r.POST("/v1/uploads", uploadAuthMiddleware(), handler.UploadFile)

		content := []byte("#!/bin/bash\nrm -rf /")
		body, contentType := createMultipartForm(t, "file", "script.sh", "text/x-shellscript", content)

		req := httptest.NewRequest("POST", "/v1/uploads", body)
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("No File Provided", func(t *testing.T) {
		r, handler, tempDir := setupUploadTest(t)
		defer os.RemoveAll(tempDir)

		r.POST("/v1/uploads", uploadAuthMiddleware(), handler.UploadFile)

		req := httptest.NewRequest("POST", "/v1/uploads", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("Content-Type", "multipart/form-data")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "ERR_INVALID_FILE")
	})

	t.Run("Unauthorized - No Token", func(t *testing.T) {
		r, handler, tempDir := setupUploadTest(t)
		defer os.RemoveAll(tempDir)

		r.POST("/v1/uploads", uploadAuthMiddleware(), handler.UploadFile)

		content := []byte("test")
		body, contentType := createMultipartForm(t, "file", "test.jpg", "image/jpeg", content)

		req := httptest.NewRequest("POST", "/v1/uploads", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Unauthorized - Invalid Token", func(t *testing.T) {
		r, handler, tempDir := setupUploadTest(t)
		defer os.RemoveAll(tempDir)

		r.POST("/v1/uploads", uploadAuthMiddleware(), handler.UploadFile)

		content := []byte("test")
		body, contentType := createMultipartForm(t, "file", "test.jpg", "image/jpeg", content)

		req := httptest.NewRequest("POST", "/v1/uploads", body)
		req.Header.Set("Authorization", "Bearer invalid-token")
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("File Saved To Disk", func(t *testing.T) {
		r, handler, tempDir := setupUploadTest(t)
		defer os.RemoveAll(tempDir)

		r.POST("/v1/uploads", uploadAuthMiddleware(), handler.UploadFile)

		content := []byte("test file content for disk check")
		body, contentType := createMultipartForm(t, "file", "disktest.txt", "text/plain", content)

		req := httptest.NewRequest("POST", "/v1/uploads", body)
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify file exists in temp directory
		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		assert.Equal(t, 1, len(files), "Expected one file to be saved")

		// Verify file extension
		if len(files) > 0 {
			assert.Equal(t, ".txt", filepath.Ext(files[0].Name()))
		}
	})

	t.Run("ZIP File Upload", func(t *testing.T) {
		r, handler, tempDir := setupUploadTest(t)
		defer os.RemoveAll(tempDir)

		r.POST("/v1/uploads", uploadAuthMiddleware(), handler.UploadFile)

		content := []byte("PK fake zip content")
		body, contentType := createMultipartForm(t, "file", "archive.zip", "application/zip", content)

		req := httptest.NewRequest("POST", "/v1/uploads", body)
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// ==================== REGISTER ROUTES TESTS ====================

func TestRegisterUploadRoutes(t *testing.T) {
	t.Run("Routes Registered", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()

		tempDir, err := os.MkdirTemp("", "upload_route_test_*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		handler := NewUploadHandler(tempDir, "http://localhost:8080/uploads")
		RegisterUploadRoutes(r, handler, uploadAuthMiddleware())

		// Test that POST /v1/uploads is registered
		content := []byte("test")
		body, contentType := createMultipartForm(t, "file", "test.jpg", "image/jpeg", content)

		req := httptest.NewRequest("POST", "/v1/uploads", body)
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should not be 404 (route exists)
		assert.NotEqual(t, http.StatusNotFound, w.Code)
	})
}
