package pkg

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"github.com/google/uuid"
)

const (
	MaxUploadSize = 100 << 20
)

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
	log.Printf("文件将保存到临时目录: %s", tempDir)
	
	return &Uploader{
		tempDir: tempDir,
	}
}

func (u *Uploader) UploadSingleFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Success: false,
			Message: "获取文件失败: " + err.Error(),
		})
		return
	}

	savedPath, err := u.saveFile(c, file)
	if err != nil {
		log.Printf("保存文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "保存文件失败: " + err.Error(),
		})
		return
	}

	log.Printf("文件 %s 上传成功: %s", file.Filename, savedPath)
	c.JSON(http.StatusOK, UploadResponse{
		Success:   true,
		Message:   "文件上传成功",
		Files:     []string{file.Filename},
		SavedPath: savedPath,
	})
}

func (u *Uploader) saveFile(c *gin.Context, file *multipart.FileHeader) (string, error) {
	id := uuid.New().String()
	ext := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%s%s", id, ext)
	savedPath := filepath.Join(u.tempDir, newFileName)
	err := c.SaveUploadedFile(file, savedPath)
	if err != nil {
		return "", err
	}
	
	return savedPath, nil
}