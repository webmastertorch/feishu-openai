package openai

import (
	"errors"
	"start-feishubot/logger"
)

type ImageURL struct {
	URL    string `json:"url,omitempty"`
	Detail string `json:"detail,omitempty"`
}

type ContentType struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}
type VisionMessages struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type VisionRequestBody struct {
	Model               string           `json:"model"`
	Messages            []VisionMessages `json:"messages"`
	MaxTokens           int              `json:"max_tokens,omitempty"`
	MaxCompletionTokens int              `json:"max_completion_tokens,omitempty"`
}

// VisionModel represents the model to use for vision requests
type VisionModel string

const (
	GPT4VisionPreview VisionModel = "gpt-4-vision-preview"
	GPT4o           VisionModel = "gpt-4o"
	GPT4oMini       VisionModel = "o4-mini"
)

// GetVisionInfo processes vision requests with the specified model
func (gpt *ChatGPT) GetVisionInfo(msg []VisionMessages) (
	resp Messages, err error) {
	// Default to gpt-4-vision-preview if not using o4-mini
	visionModel := GPT4VisionPreview

	// If the model is set to o4-mini or gpt-4o, use that instead
	if gpt.Model == string(GPT4oMini) {
		visionModel = GPT4oMini
	} else if gpt.Model == string(GPT4o) {
		visionModel = GPT4o
	}

	// Create request body based on model type
	requestBody := VisionRequestBody{
		Model:    string(visionModel),
		Messages: msg,
	}

	// 注意：我们不在这里设置 MaxTokens 或 MaxCompletionTokens
	// 这些参数会在 doAPIRequestWithRetry 方法中根据模型类型进行处理
	// 这样可以确保 o4-mini 和 gpt-4o 模型使用 max_completion_tokens 参数
	// 而其他模型使用 max_tokens 参数

	gptResponseBody := &ChatGPTResponseBody{}
	url := gpt.FullUrl("chat/completions")
	logger.Debug("request body ", requestBody)
	if url == "" {
		return resp, errors.New("无法获取openai请求地址")
	}

	err = gpt.sendRequestWithBodyType(url, "POST", jsonBody, requestBody, gptResponseBody)
	if err == nil && len(gptResponseBody.Choices) > 0 {
		resp = gptResponseBody.Choices[0].Message
	} else {
		logger.Errorf("ERROR %v", err)
		resp = Messages{}
		err = errors.New("openai 请求失败")
	}
	return resp, err
}
