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
	// 设置测试模式
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestUploadSingleFile_DocxFile(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	uploader := NewUploader(tempDir)

	// 设置路由
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
			name:           "上传docx文件成功",
			filename:       "test.docx",
			content:        "fake docx content",
			expectedStatus: http.StatusOK,
			expectedFiles:  []string{"test.docx"},
			checkSavedPath: true,
		},
		{
			name:           "上传带中文名的docx文件",
			filename:       "测试文档.docx",
			content:        "fake docx content with chinese name",
			expectedStatus: http.StatusOK,
			expectedFiles:  []string{"测试文档.docx"},
			checkSavedPath: true,
		},
		{
			name:           "上传大写扩展名的DOCX文件",
			filename:       "TEST.DOCX",
			content:        "fake DOCX content",
			expectedStatus: http.StatusOK,
			expectedFiles:  []string{"TEST.DOCX"},
			checkSavedPath: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建multipart form
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			// 添加文件
			part, err := writer.CreateFormFile("file", tt.filename)
			require.NoError(t, err)
			
			_, err = io.WriteString(part, tt.content)
			require.NoError(t, err)
			
			err = writer.Close()
			require.NoError(t, err)

			// 创建请求
			req := httptest.NewRequest("POST", "/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			
			// 记录响应
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证响应状态
			assert.Equal(t, tt.expectedStatus, w.Code)

			// 解析响应
			var response UploadResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// 验证响应内容
			assert.True(t, response.Success)
			assert.Equal(t, "文件上传成功", response.Message)
			assert.Equal(t, tt.expectedFiles, response.Files)

			if tt.checkSavedPath {
				// 验证保存路径不为空
				assert.NotEmpty(t, response.SavedPath)
				
				// 验证文件确实保存了
				assert.FileExists(t, response.SavedPath)
				
				// 验证保存路径的格式（应该包含UUID和原始扩展名）
				savedFileName := filepath.Base(response.SavedPath)
				ext := filepath.Ext(response.SavedPath)
				originalExt := filepath.Ext(tt.filename)
				assert.Equal(t, originalExt, ext, "保存的文件扩展名应该与原文件一致")
				
				// 验证文件名格式（UUID + 扩展名）
				nameWithoutExt := strings.TrimSuffix(savedFileName, ext)
				assert.Len(t, nameWithoutExt, 36, "文件名应该是36位的UUID")
				
				// 输出保存路径（按要求输出）
				fmt.Printf("测试用例 [%s]: 文件保存路径: %s\n", tt.name, response.SavedPath)
				
				// 验证文件内容
				savedContent, err := os.ReadFile(response.SavedPath)
				require.NoError(t, err)
				assert.Equal(t, tt.content, string(savedContent), "保存的文件内容应该与上传内容一致")
				
				// 输出文件信息
				fileInfo, err := os.Stat(response.SavedPath)
				require.NoError(t, err)
				fmt.Printf("文件大小: %d bytes, 修改时间: %s\n", fileInfo.Size(), fileInfo.ModTime().Format("2006-01-02 15:04:05"))
			}
		})
	}
}

func TestUploadSingleFile_DocxFile_ErrorCases(t *testing.T) {
	tempDir := t.TempDir()
	uploader := NewUploader(tempDir)

	router := gin.New()
	router.POST("/upload", uploader.UploadSingleFile)

	t.Run("没有文件字段", func(t *testing.T) {
		// 创建空的multipart form
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
		assert.Contains(t, response.Message, "获取文件失败")
		fmt.Printf("错误测试 - 没有文件字段: %s\n", response.Message)
	})

	t.Run("错误的Content-Type", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/upload", strings.NewReader("not multipart"))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response UploadResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "获取文件失败")
		fmt.Printf("错误测试 - 错误的Content-Type: %s\n", response.Message)
	})
}

func TestUploadSingleFile_DocxFile_LargeFile(t *testing.T) {
	tempDir := t.TempDir()
	uploader := NewUploader(tempDir)

	router := gin.New()
	router.Use(gin.Recovery())
	router.POST("/upload", uploader.UploadSingleFile)

	t.Run("上传大型docx文件", func(t *testing.T) {
		// 创建一个较大的测试内容（但不超过最大限制）
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
		
		fmt.Printf("大文件测试 - 文件保存路径: %s\n", response.SavedPath)
		
		// 验证文件大小
		fileInfo, err := os.Stat(response.SavedPath)
		require.NoError(t, err)
		fmt.Printf("大文件大小: %d bytes\n", fileInfo.Size())
		assert.Equal(t, int64(len(largeContent)), fileInfo.Size())
	})
}

// 基准测试
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
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
}