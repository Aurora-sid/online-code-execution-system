package ai

import (
	"code-exec/config"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler AI 分析 API Handler
type Handler struct {
	client *LLMClient
}

// NewHandler 创建 Handler 实例
func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		client: NewLLMClient(cfg),
	}
}

// AnalyzeRequest 分析请求
type AnalyzeRequest struct {
	Code     string `json:"code" binding:"required"`     // 代码内容
	Language string `json:"language" binding:"required"` // 编程语言
	Type     string `json:"type"`                        // 分析类型: bug, optimize, explain, security
}

// AnalyzeResponse 分析响应
type AnalyzeResponse struct {
	Result string `json:"result"` // 分析结果 (Markdown)
	Status string `json:"status"` // success / error
	Error  string `json:"error,omitempty"`
}

// Analyze 代码分析接口
func (h *Handler) Analyze(c *gin.Context) {
	log.Println("[AI] 收到分析请求")

	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AI] 请求参数解析失败: %v\n", err)
		c.JSON(http.StatusBadRequest, AnalyzeResponse{
			Status: "error",
			Error:  "请求参数无效: " + err.Error(),
		})
		return
	}

	log.Printf("[AI] 分析请求: 语言=%s, 类型=%s, 代码长度=%d\n", req.Language, req.Type, len(req.Code))

	// 默认分析类型为 bug 检测
	analysisType := AnalysisTypeBugDetect
	switch req.Type {
	case "optimize":
		analysisType = AnalysisTypeOptimize
	case "explain":
		analysisType = AnalysisTypeExplain
	case "security":
		analysisType = AnalysisTypeSecurityCheck
	case "error":
		analysisType = AnalysisTypeErrorAnalysis
	case "bug":
		analysisType = AnalysisTypeBugDetect
	}

	log.Printf("[AI] 开始调用 LLM API (类型: %s)...\n", analysisType)
	result, err := h.client.AnalyzeCode(req.Code, req.Language, analysisType)
	if err != nil {
		log.Printf("[AI] LLM 调用失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, AnalyzeResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	log.Printf("[AI] 分析完成, 结果长度: %d 字符\n", len(result))

	c.JSON(http.StatusOK, AnalyzeResponse{
		Result: result,
		Status: "success",
	})
}

// RegisterRoutes 注册 AI 路由
func RegisterRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	handler := NewHandler(cfg)

	ai := rg.Group("/ai")
	{
		ai.POST("/analyze", handler.Analyze)
	}
}
