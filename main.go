
package main

import (
	"log"
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
	"doc2script/pkg" 
)


type GenerateRequest struct {
	Model     string `json:"model" binding:"required"`     // "hc" or "lc"
	GenType   string `json:"gentype" binding:"required"`   // "d" or "r"
	Pattern   string `json:"pattern" binding:"required"`   // 正则表达式
	FilePath  string `json:"file_path" binding:"required"` // 上传文件的路径
	NumScript int    `json:"num_script" binding:"required"` // 生成剧本数量
}

type GenerateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	SaveDir string `json:"save_dir,omitempty"`
}

type Service struct {
	uploader *pkg.Uploader
	tempDir  string
	saveDir  string
}

func NewService(tempDir, saveDir string) *Service {
	os.MkdirAll(tempDir, 0755)
	os.MkdirAll(saveDir, 0755)
	
	return &Service{
		uploader: pkg.NewUploader(tempDir),
		tempDir:  tempDir,
		saveDir:  saveDir,
	}
}

func (s *Service) generateScript(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, GenerateResponse{
			Success: false,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}
	
	config, err := pkg.NewEngineConfig(
		req.Model,
		req.GenType,
		req.Pattern,
		req.FilePath,
		s.saveDir,
		req.NumScript,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, GenerateResponse{
			Success: false,
			Message: "配置错误: " + err.Error(),
		})
		return
	}
	
	engine, err := pkg.NewEngine(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenerateResponse{
			Success: false,
			Message: "创建引擎失败: " + err.Error(),
		})
		return
	}
	
	err = engine.GenerateScript()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenerateResponse{
			Success: false,
			Message: "生成脚本失败: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, GenerateResponse{
		Success: true,
		Message: "脚本生成成功",
		SaveDir: s.saveDir,
	})
}

func (s *Service) setupRoutes() *gin.Engine {
	r := gin.Default()
	r.MaxMultipartMemory = pkg.MaxUploadSize
	api := r.Group("/api/v1")
	{
		api.POST("/upload", s.uploader.UploadSingleFile)
		api.POST("/generate", s.generateScript)
	}
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return r
}

func main() {
	tempDir := "./temp"
	saveDir := "./scripts"
	service := NewService(tempDir, saveDir)
	r := service.setupRoutes()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("服务器启动在端口: %s", port)
	log.Printf("上传目录: %s", tempDir)
	log.Printf("脚本保存目录: %s", saveDir)
	log.Fatal(r.Run(":" + port))
}