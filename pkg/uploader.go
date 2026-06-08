package pkg

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"mime/multipart"
)

const MaxUploadSize = 100 << 20

var allowedExtensions = map[string]bool{
	".docx": true,
	".txt":  true,
	".doc":  true,
}

type UploadResponse struct {
	Success   bool     `json:"success"`
	Message   string   `json:"message"`
	Files     []string `json:"files,omitempty"`
	SavedPath string   `json:"saved_path,omitempty"`
}

type Uploader struct {
	tempDir string
}

func NewUploader(tempDir string) *Uploader {
	return &Uploader{tempDir: tempDir}
}

func (u *Uploader) UploadSingleFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Success: false,
			Message: "failed to read uploaded file: " + err.Error(),
		})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExtensions[ext] {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Success: false,
			Message: fmt.Sprintf("unsupported file type: %s", ext),
		})
		return
	}

	savedPath, err := u.saveFile(c, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "failed to save file: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UploadResponse{
		Success:   true,
		Message:   "file uploaded successfully",
		Files:     []string{file.Filename},
		SavedPath: savedPath,
	})
}

func (u *Uploader) saveFile(c *gin.Context, file *multipart.FileHeader) (string, error) {
	id := uuid.New().String()
	ext := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%s%s", id, ext)
	savedPath := filepath.Join(u.tempDir, newFileName)
	if err := c.SaveUploadedFile(file, savedPath); err != nil {
		return "", err
	}
	return savedPath, nil
}
