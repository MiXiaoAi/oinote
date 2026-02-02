package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MiXiaoAi/oinote/backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AIHandler struct {
	DB *gorm.DB
}

func NewAIHandler(db *gorm.DB) *AIHandler {
	return &AIHandler{DB: db}
}

// GetAIConfig 获取 AI 配置（仅管理员）
func (h *AIHandler) GetAIConfig(c *fiber.Ctx) error {
	var config models.AIConfig
	// AI配置只有一条记录，ID 为 1
	if err := h.DB.First(&config, 1).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果不存在，返回空配置
			return c.JSON(models.AIConfig{
				ID:       1,
				OpenAIURL: "",
				APIKey:   "",
				Model:    "",
			})
		}
		return c.Status(500).JSON(fiber.Map{"error": "获取配置失败"})
	}
	return c.JSON(config)
}

// UpdateAIConfig 更新 AI 配置（仅管理员）
func (h *AIHandler) UpdateAIConfig(c *fiber.Ctx) error {
	userId := c.Locals("userId").(uint)

	type UpdateAIConfigInput struct {
		OpenAIURL *string `json:"openai_url"`
		APIKey    *string `json:"api_key"`
		Model     *string `json:"model"`
	}

	var input UpdateAIConfigInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "输入数据无效"})
	}

	var config models.AIConfig
	if err := h.DB.First(&config, 1).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果不存在，创建新配置
			config = models.AIConfig{ID: 1}
		} else {
			return c.Status(500).JSON(fiber.Map{"error": "更新配置失败"})
		}
	}

	if input.OpenAIURL != nil {
		config.OpenAIURL = *input.OpenAIURL
	}
	if input.APIKey != nil {
		config.APIKey = *input.APIKey
	}
	if input.Model != nil {
		config.Model = *input.Model
	}
	config.UpdatedBy = userId

	if err := h.DB.Save(&config).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "更新配置失败"})
	}

	return c.JSON(config)
}

// SummarizeNote AI 总结笔记
func (h *AIHandler) SummarizeNote(c *fiber.Ctx) error {
	userId := c.Locals("userId")
	if userId == nil {
		return c.Status(401).JSON(fiber.Map{"error": "未授权"})
	}

	type SummarizeInput struct {
		Content string `json:"content"`
		Title   string `json:"title"`
	}

	var input SummarizeInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "输入数据无效"})
	}

	if input.Content == "" {
		return c.Status(400).JSON(fiber.Map{"error": "内容不能为空"})
	}

	// 获取 AI 配置
	var aiConfig models.AIConfig
	if err := h.DB.First(&aiConfig, 1).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "AI 配置未设置，请联系管理员"})
	}

	if aiConfig.APIKey == "" || aiConfig.OpenAIURL == "" {
		return c.Status(400).JSON(fiber.Map{"error": "AI 配置不完整，请联系管理员"})
	}

	// 调用 OpenAI API
	summary, err := h.callOpenAI(aiConfig, input.Title, input.Content)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "AI 总结失败: " + err.Error()})
	}

	return c.JSON(fiber.Map{"summary": summary})
}

// callOpenAI 调用 OpenAI API
func (h *AIHandler) callOpenAI(config models.AIConfig, title, content string) (string, error) {
	// 构建请求体
	prompt := "请对以下笔记内容进行总结，提取关键要点：\n\n标题：" + title + "\n\n内容：" + content + "\n\n请用中文回答，条理清晰，格式如下：\n\n【AI 总结】\n\n关键要点：\n1. ...\n2. ...\n3. ...\n\n建议行动项：\n- ...\n- ..."

	requestBody := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("构建请求失败: %v", err)
	}

	// 验证和清理 URL
	url := config.OpenAIURL
	if url == "" {
		url = "https://api.openai.com/v1/chat/completions"
	}

	// 检查 URL 是否有协议前缀
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "", fmt.Errorf("OpenAI URL 格式错误：缺少协议前缀（http:// 或 https://），当前值为: %s", url)
	}

	// 验证 URL 格式
	_, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("OpenAI URL 格式错误: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 返回错误 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	// 提取生成的文本
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("响应格式错误: 缺少 choices")
	}

	choice := choices[0].(map[string]interface{})
	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("响应格式错误: 缺少 message")
	}

	resultContent, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("响应格式错误: 缺少 content")
	}

	return resultContent, nil
}

// PolishNote AI 润色笔记
func (h *AIHandler) PolishNote(c *fiber.Ctx) error {
	userId := c.Locals("userId")
	if userId == nil {
		return c.Status(401).JSON(fiber.Map{"error": "未授权"})
	}

	type PolishInput struct {
		Content string `json:"content"`
		Title   string `json:"title"`
	}

	var input PolishInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "输入数据无效"})
	}

	if input.Content == "" {
		return c.Status(400).JSON(fiber.Map{"error": "内容不能为空"})
	}

	// 获取 AI 配置
	var aiConfig models.AIConfig
	if err := h.DB.First(&aiConfig, 1).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "AI 配置未设置，请联系管理员"})
	}

	if aiConfig.APIKey == "" || aiConfig.OpenAIURL == "" {
		return c.Status(400).JSON(fiber.Map{"error": "AI 配置不完整，请联系管理员"})
	}

	// 调用 OpenAI API 进行润色
	polished, err := h.callOpenAIForPolish(aiConfig, input.Title, input.Content)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "AI 润色失败: " + err.Error()})
	}

	return c.JSON(fiber.Map{"polished": polished})
}

// callOpenAIForPolish 调用 OpenAI API 进行润色
func (h *AIHandler) callOpenAIForPolish(config models.AIConfig, title, content string) (string, error) {
	// 构建请求体 - 润色提示词
	prompt := `请对以下笔记内容进行润色和优化，使其更加清晰、准确、专业：

标题：` + title + `

内容：` + content + `

要求：
1. 保持原文的核心意思和结构
2. 改善语言的流畅性和专业性
3. 修正语法错误和拼写错误
4. 优化段落结构，使内容更有条理
5. 保持中文表达自然流畅
6. 不要添加原文中没有的新信息

请只返回润色后的内容，不要包含任何解释或说明。`

	requestBody := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3, // 润色使用更低的温度，使输出更稳定
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("构建请求失败: %v", err)
	}

	// 验证和清理 URL
	url := config.OpenAIURL
	if url == "" {
		url = "https://api.openai.com/v1/chat/completions"
	}

	// 检查 URL 是否有协议前缀
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "", fmt.Errorf("OpenAI URL 格式错误：缺少协议前缀（http:// 或 https://），当前值为: %s", url)
	}

	// 验证 URL 格式
	_, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("OpenAI URL 格式错误: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 返回错误 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	// 提取生成的文本
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("响应格式错误: 缺少 choices")
	}

	choice := choices[0].(map[string]interface{})
	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("响应格式错误: 缺少 message")
	}

	polishedContent, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("响应格式错误: 缺少 content")
	}

	return polishedContent, nil
}
