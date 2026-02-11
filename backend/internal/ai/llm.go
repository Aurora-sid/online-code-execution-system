package ai

import (
	"bytes"
	"code-exec/config"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// LLMClient 大模型客户端
type LLMClient struct {
	apiKey string
	apiURL string
	model  string
	client *http.Client
}

// NewLLMClient 创建 LLM 客户端
func NewLLMClient(cfg *config.Config) *LLMClient {
	log.Printf("[LLM] 初始化客户端: URL=%s, Model=%s, APIKey已配置=%v\n",
		cfg.LLMAPIURL, cfg.LLMModel, cfg.LLMAPIKey != "")
	return &LLMClient{
		apiKey: cfg.LLMAPIKey,
		apiURL: cfg.LLMAPIURL,
		model:  cfg.LLMModel,
		client: &http.Client{
			Timeout: 60 * time.Second, // 大模型可能需要较长时间
		},
	}
}

// ChatMessage 对话消息
type ChatMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"` // 消息内容
}

// ChatRequest 请求体
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

// ChatResponse 响应体 (兼容 OpenAI 格式)
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// AnalysisType 分析类型
type AnalysisType string

const (
	AnalysisTypeBugDetect     AnalysisType = "bug"      // Bug 检测
	AnalysisTypeOptimize      AnalysisType = "optimize" // 代码优化
	AnalysisTypeExplain       AnalysisType = "explain"  // 代码解释
	AnalysisTypeSecurityCheck AnalysisType = "security" // 安全检查
	AnalysisTypeErrorAnalysis AnalysisType = "error"    // 报错分析
)

// AnalyzeCode 分析代码
func (c *LLMClient) AnalyzeCode(code, language string, analysisType AnalysisType) (string, error) {
	log.Printf("[LLM] 开始分析: 语言=%s, 类型=%s, 代码长度=%d\n", language, analysisType, len(code))

	if c.apiKey == "" {
		log.Println("[LLM] 错误: API Key 未配置")
		return "", fmt.Errorf("LLM API Key 未配置")
	}

	// 根据分析类型构建 Prompt
	systemPrompt := getSystemPrompt(analysisType, language)

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("请分析以下 %s 代码:\n\n```%s\n%s\n```", language, language, code)},
	}

	reqBody := ChatRequest{
		Model:    c.model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", c.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	log.Printf("[LLM] 发送请求到 %s (Model: %s)\n", c.apiURL, c.model)
	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("[LLM] HTTP 请求失败: %v\n", err)
		return "", fmt.Errorf("请求 LLM API 失败: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[LLM] 收到响应: 状态码=%d\n", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[LLM] API 返回错误: 状态码=%d, 响应=%s\n", resp.StatusCode, string(body))
		return "", fmt.Errorf("LLM API 返回错误 (%d): %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("LLM 返回错误: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		log.Println("[LLM] 错误: API 未返回结果")
		return "", fmt.Errorf("LLM 未返回结果")
	}

	log.Printf("[LLM] 分析成功, 结果长度: %d 字符\n", len(chatResp.Choices[0].Message.Content))
	return chatResp.Choices[0].Message.Content, nil
}

// getSystemPrompt 根据分析类型生成系统提示词
func getSystemPrompt(analysisType AnalysisType, language string) string {
	base := fmt.Sprintf("你是一个专业的代码分析助手，精通多种编程语言。请针对 %s 语言进行分析，并使用中文回答。", language)

	switch analysisType {
	case AnalysisTypeBugDetect:
		return base + `
你的任务是检测代码中的潜在 Bug 和逻辑错误。请：
1. 找出代码中可能存在的 Bug
2. 解释每个问题的原因
3. 提供修复建议
4. 使用 Markdown 格式输出`

	case AnalysisTypeOptimize:
		return base + `
你的任务是分析代码并提供优化建议。请：
1. 识别性能瓶颈
2. 提出代码结构优化建议
3. 建议更好的算法或数据结构
4. 提供优化后的代码示例
5. 使用 Markdown 格式输出`

	case AnalysisTypeExplain:
		return base + `
你的任务是解释代码的功能和逻辑。请：
1. 概述代码的主要功能
2. 逐段解释代码逻辑
3. 说明关键变量和函数的作用
4. 使用 Markdown 格式输出`

	case AnalysisTypeSecurityCheck:
		return base + `
你的任务是进行安全审计。请：
1. 检查常见安全漏洞（SQL注入、XSS、缓冲区溢出等）
2. 识别敏感信息泄露风险
3. 评估输入验证和边界检查
4. 提供安全加固建议
5. 使用 Markdown 格式输出`

	case AnalysisTypeErrorAnalysis:
		return base + `
你的任务是分析代码运行时产生的错误。用户会提供代码和终端的报错信息，请：
1. 准确定位错误发生的位置和原因
2. 用通俗易懂的语言解释为什么会出现这个错误
3. 提供具体的修复代码示例
4. 如有必要，给出避免类似错误的建议
5. 使用 Markdown 格式输出`

	default:
		return base + "请全面分析这段代码，包括其功能、潜在问题和优化建议。使用 Markdown 格式输出。"
	}
}
