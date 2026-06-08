package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestUploadSingleFile_DocxFile(t *testing.T) {
	tempDir := t.TempDir()
	uploader := NewUploader(tempDir)

	router := gin.New()
	router.POST("/upload", uploader.UploadSingleFile)

	tests := []struct {
		name           string
		filename       string
		content        string
		expectedStatus int
		expectedFiles  []string
		checkSavedPath bool
	}{
		{
			name:           "upload docx file",
			filename:       "test.docx",
			content:        "fake docx content",
			expectedStatus: http.StatusOK,
			expectedFiles:  []string{"test.docx"},
			checkSavedPath: true,
		},
		{
			name:           "upload file with chinese name",
			filename:       "测试文档.docx",
			content:        "fake docx content",
			expectedStatus: http.StatusOK,
			expectedFiles:  []string{"测试文档.docx"},
			checkSavedPath: true,
		},
		{
			name:           "upload uppercase extension",
			filename:       "TEST.DOCX",
			content:        "fake DOCX content",
			expectedStatus: http.StatusOK,
			expectedFiles:  []string{"TEST.DOCX"},
			checkSavedPath: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			part, err := writer.CreateFormFile("file", tt.filename)
			require.NoError(t, err)

			_, err = io.WriteString(part, tt.content)
			require.NoError(t, err)

			err = writer.Close()
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response UploadResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.True(t, response.Success)
			assert.Equal(t, "file uploaded successfully", response.Message)
			assert.Equal(t, tt.expectedFiles, response.Files)

			if tt.checkSavedPath {
				assert.NotEmpty(t, response.SavedPath)
				assert.FileExists(t, response.SavedPath)

				savedFileName := filepath.Base(response.SavedPath)
				ext := filepath.Ext(response.SavedPath)
				originalExt := filepath.Ext(tt.filename)
				assert.Equal(t, originalExt, ext)

				nameWithoutExt := strings.TrimSuffix(savedFileName, ext)
				assert.Len(t, nameWithoutExt, 36)

				fmt.Printf("[%s]: saved path: %s\n", tt.name, response.SavedPath)

				savedContent, err := os.ReadFile(response.SavedPath)
				require.NoError(t, err)
				assert.Equal(t, tt.content, string(savedContent))

				fileInfo, err := os.Stat(response.SavedPath)
				require.NoError(t, err)
				fmt.Printf("file size: %d bytes, modified: %s\n", fileInfo.Size(), fileInfo.ModTime().Format("2006-01-02 15:04:05"))
			}
		})
	}
}

func TestUploadSingleFile_DocxFile_ErrorCases(t *testing.T) {
	tempDir := t.TempDir()
	uploader := NewUploader(tempDir)

	router := gin.New()
	router.POST("/upload", uploader.UploadSingleFile)

	t.Run("no file field", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		err := writer.Close()
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response UploadResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "failed to read uploaded file")
	})

	t.Run("wrong content type", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/upload", strings.NewReader("not multipart"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response UploadResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "failed to read uploaded file")
	})

	t.Run("unsupported file type", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("file", "test.exe")
		require.NoError(t, err)
		_, err = io.WriteString(part, "malicious content")
		require.NoError(t, err)
		err = writer.Close()
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response UploadResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "unsupported file type")
	})
}

func TestUploadSingleFile_DocxFile_LargeFile(t *testing.T) {
	tempDir := t.TempDir()
	uploader := NewUploader(tempDir)

	router := gin.New()
	router.Use(gin.Recovery())
	router.POST("/upload", uploader.UploadSingleFile)

	t.Run("upload large docx file", func(t *testing.T) {
		largeContent := strings.Repeat("This is a large docx file content. ", 1000)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("file", "large_document.docx")
		require.NoError(t, err)

		_, err = io.WriteString(part, largeContent)
		require.NoError(t, err)

		err = writer.Close()
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response UploadResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.NotEmpty(t, response.SavedPath)
		assert.FileExists(t, response.SavedPath)

		fmt.Printf("large file saved: %s\n", response.SavedPath)

		fileInfo, err := os.Stat(response.SavedPath)
		require.NoError(t, err)
		fmt.Printf("large file size: %d bytes\n", fileInfo.Size())
		assert.Equal(t, int64(len(largeContent)), fileInfo.Size())
	})
}

func BenchmarkUploadSingleFile_DocxFile(b *testing.B) {
	tempDir := b.TempDir()
	uploader := NewUploader(tempDir)

	router := gin.New()
	router.POST("/upload", uploader.UploadSingleFile)

	content := "benchmark test docx content"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("file", fmt.Sprintf("benchmark_%d.docx", i))
		if err != nil {
			b.Fatal(err)
		}

		_, err = io.WriteString(part, content)
		if err != nil {
			b.Fatal(err)
		}

		err = writer.Close()
		if err != nil {
			b.Fatal(err)
		}

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("expected status 200, got %d", w.Code)
		}
	}
}
