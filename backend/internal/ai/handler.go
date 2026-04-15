package ai

import (
	"code-exec/config" // code-exec意为项目名，在go.mod中定义 module code-exec
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 处理 AI 相关请求
type Handler struct {
	// 定义一个名为 Client 的字段，指向 LLMClient 类型的指针
	Client *LLMClient // 公开，便于传递给 AdminHandler
}

//	创建 Handler 实例
//
// 传入全局配置，初始化 LLM 客户端
func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		// cfg 包含了 LLM 配置（API Key、URL、模型等），直接传递给 LLMClient 构造函数
		Client: NewLLMClient(cfg),
	}
}

// 分析请求
type AnalyzeRequest struct {
	Code     string `json:"code" binding:"required"`     // 代码内容
	Language string `json:"language" binding:"required"` // 编程语言
	Type     string `json:"type"`                        // 分析类型: bug, optimize, explain, security
}

// 分析响应
type AnalyzeResponse struct {
	Result string `json:"result"` // 分析结果 (Markdown)
	Status string `json:"status"` // success / error
	Error  string `json:"error,omitempty"`
}

// 代码分析接口
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
	//log.Println(req)

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
	result, err := h.Client.AnalyzeCode(req.Code, req.Language, analysisType)
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

/*================接口创建================*/
// RegisterRoutes 注册 AI 路由，返回 Handler 以便获取 LLM 客户端实例
// 开启POST /ai/analyze接口，供前端调用进行代码分析
func RegisterRoutes(rg *gin.RouterGroup, cfg *config.Config) *Handler {
	handler := NewHandler(cfg)

	ai := rg.Group("/ai")
	{
		ai.POST("/analyze", handler.Analyze)
	}
	// 当有用户输入‘/ai/analyze’时，Gin 会将请求路由到 handler.Analyze 方法进行处理，

	return handler
}
