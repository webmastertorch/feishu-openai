package openai

import (
	"context"
	"errors"
	"fmt"
	go_openai "github.com/sashabaranov/go-openai"
	"io"
	"time"
)

func (c *ChatGPT) StreamChat(ctx context.Context,
	msg []Messages, mode AIMode,
	responseStream chan string) error {
	//change msg type from Messages to openai.ChatCompletionMessage
	chatMsgs := make([]go_openai.ChatCompletionMessage, len(msg))
	for i, m := range msg {
		chatMsgs[i] = go_openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}
	return c.StreamChatWithHistory(ctx, chatMsgs, 2000, mode,
		responseStream)
}

func (c *ChatGPT) StreamChatWithHistory(ctx context.Context,
	msg []go_openai.ChatCompletionMessage, maxTokens int,
	aiMode AIMode,
	responseStream chan string,
) error {
	// 如果是 o4-mini 或 gpt-4o 模型，使用非流式 API 并模拟流式响应
	if c.Model == "o4-mini" || c.Model == "gpt-4o" {
		return c.nonStreamChatWithHistory(ctx, msg, maxTokens, aiMode, responseStream)
	}

	// 对于其他模型，使用原来的流式 API
	config := go_openai.DefaultConfig(c.ApiKey[0])
	config.BaseURL = c.ApiUrl + "/v1"
	if c.Platform != OpenAI {
		baseUrl := fmt.Sprintf("https://%s.%s",
			c.AzureConfig.ResourceName, "openai.azure.com")
		config = go_openai.DefaultAzureConfig(c.AzureConfig.
			ApiToken, baseUrl)
		config.AzureModelMapperFunc = func(model string) string {
			return c.AzureConfig.DeploymentName
		}
	}

	proxyClient, parseProxyError := GetProxyClient(c.HttpProxy)
	if parseProxyError != nil {
		return parseProxyError
	}
	config.HTTPClient = proxyClient

	client := go_openai.NewClientWithConfig(config)
	var temperature float32
	temperature = float32(aiMode)
	req := go_openai.ChatCompletionRequest{
		Model:       c.Model,
		Messages:    msg,
		N:           1,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return fmt.Errorf("CreateCompletionStream returned error: %v", err)
	}

	defer stream.Close()
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("Stream error: %v", err)
		}
		responseStream <- response.Choices[0].Delta.Content
	}
}

// nonStreamChatWithHistory 使用非流式 API 并模拟流式响应
func (c *ChatGPT) nonStreamChatWithHistory(ctx context.Context,
	msg []go_openai.ChatCompletionMessage, maxTokens int,
	aiMode AIMode,
	responseStream chan string,
) error {
	// 将 go_openai.ChatCompletionMessage 转换为 Messages
	chatMsgs := make([]Messages, len(msg))
	for i, m := range msg {
		chatMsgs[i] = Messages{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	// 使用我们自己的 API 调用
	completions, err := c.Completions(chatMsgs, aiMode)
	if err != nil {
		return err
	}

	// 模拟流式响应
	// 将完整响应分成小块发送
	content := completions.Content
	chunkSize := 10 // 每次发送 10 个字符

	for i := 0; i < len(content); i += chunkSize {
		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}
		chunk := content[i:end]
		responseStream <- chunk

		// 添加一点延迟，模拟流式响应
		select {
		case <-time.After(50 * time.Millisecond):
			// 继续
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}
