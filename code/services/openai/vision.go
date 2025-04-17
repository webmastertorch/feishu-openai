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
	Model     string           `json:"model"`
	Messages  []VisionMessages `json:"messages"`
	MaxTokens int              `json:"max_tokens"`
}

// VisionModel represents the model to use for vision requests
type VisionModel string

const (
	GPT4VisionPreview VisionModel = "o4-mini"
	GPT4o           VisionModel = "o4-mini"
	GPT4oMini       VisionModel = "o4-mini"
)

// GetVisionInfo processes vision requests with the specified model
func (gpt *ChatGPT) GetVisionInfo(msg []VisionMessages) (
	resp Messages, err error) {
	// Default to gpt-4-vision-preview if not using o4-mini
	visionModel := VisionModel("o4-mini")

	// If the model is set to o4-mini or gpt-4o, use that instead
	if gpt.Model == string(GPT4oMini) {
		visionModel = GPT4oMini
	} else if gpt.Model == string(GPT4o) {
		visionModel = GPT4o
	}

	requestBody := VisionRequestBody{
		Model:     string(visionModel),
		Messages:  msg,
		MaxTokens: gpt.MaxTokens,
	}
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
