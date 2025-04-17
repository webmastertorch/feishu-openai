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
	// 对于所有模型，使用我们自己的非流式 API 并模拟流式响应
	// 这样可以避免 go-openai 库中的 max_tokens 参数问题
	return c.nonStreamChatWithHistory(ctx, msg, maxTokens, aiMode, responseStream)
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
	// 注意：这里不需要传递 maxTokens 参数，因为 c.Completions 方法会使用 c.MaxTokens
	// 而 c.MaxTokens 会在 doAPIRequestWithRetry 方法中根据模型类型转换为 max_completion_tokens
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
