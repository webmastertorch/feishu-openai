package handlers

import (
	"context"
	"fmt"
	"start-feishubot/services"
	"start-feishubot/services/openai"
)

// MultimodalAction 处理多模态消息（文本、图片或组合）
type MultimodalAction struct{}

func (ma *MultimodalAction) Execute(a *ActionInfo) bool {
	// 如果不是 o4-mini 或 gpt-4o 模型，则跳过此 Action
	if a.handler.gpt.Model != "o4-mini" && a.handler.gpt.Model != "gpt-4o" {
		return true
	}

	// 处理不同类型的消息
	switch a.info.msgType {
	case "text":
		return ma.handleTextMessage(a)
	case "image":
		return ma.handleImageMessage(a)
	case "post":
		return ma.handlePostMessage(a)
	default:
		return true
	}
}

func (ma *MultimodalAction) handleTextMessage(a *ActionInfo) bool {
	// 处理纯文本消息
	msg := a.handler.sessionCache.GetMsg(*a.info.sessionId)
	msg = setDefaultPrompt(msg)
	msg = append(msg, openai.Messages{
		Role: "user", Content: a.info.qParsed,
	})

	aiMode := a.handler.sessionCache.GetAIMode(*a.info.sessionId)
	completions, err := a.handler.gpt.Completions(msg, aiMode)
	if err != nil {
		replyMsg(*a.ctx, fmt.Sprintf("🤖️：消息处理失败，请稍后再试～\n错误信息: %v", err), a.info.msgId)
		return false
	}

	msg = append(msg, completions)
	a.handler.sessionCache.SetMsg(*a.info.sessionId, msg)

	if len(msg) == 3 {
		sendNewTopicCard(*a.ctx, a.info.sessionId, a.info.msgId, completions.Content)
	} else {
		sendOldTopicCard(*a.ctx, a.info.sessionId, a.info.msgId, completions.Content)
	}

	return false
}

func (ma *MultimodalAction) handleImageMessage(a *ActionInfo) bool {
	// 处理图片消息
	base64, err := downloadAndEncodeImage(a.info.imageKey, a.info.msgId)
	if err != nil {
		replyWithErrorMsg(*a.ctx, err, a.info.msgId)
		return false
	}

	// 创建多模态消息
	msg := createVisionMessages("解释这个图片", base64, "high")
	completions, err := a.handler.gpt.GetVisionInfo(msg)
	if err != nil {
		replyWithErrorMsg(*a.ctx, err, a.info.msgId)
		return false
	}

	sendVisionTopicCard(*a.ctx, a.info.sessionId, a.info.msgId, completions.Content)
	return false
}

func (ma *MultimodalAction) handlePostMessage(a *ActionInfo) bool {
	// 处理富文本消息（可能包含文本和图片）
	var base64s []string

	for _, imageKey := range a.info.imageKeys {
		if imageKey == "" {
			continue
		}
		base64, err := downloadAndEncodeImage(imageKey, a.info.msgId)
		if err != nil {
			replyWithErrorMsg(*a.ctx, err, a.info.msgId)
			return false
		}
		base64s = append(base64s, base64)
	}

	// 如果没有图片，则作为纯文本处理
	if len(base64s) == 0 {
		return ma.handleTextMessage(a)
	}

	// 创建多模态消息
	msg := createMultipleVisionMessages(a.info.qParsed, base64s, "high")
	completions, err := a.handler.gpt.GetVisionInfo(msg)
	if err != nil {
		replyWithErrorMsg(*a.ctx, err, a.info.msgId)
		return false
	}

	sendVisionTopicCard(*a.ctx, a.info.sessionId, a.info.msgId, completions.Content)
	return false
}

// 发送多模态消息处理中的提示卡片
func sendMultimodalProcessingCard(ctx context.Context, sessionId *string, msgId *string) {
	sendSystemCard(ctx, "🤖️：正在处理您的多模态消息，请稍候...", sessionId, msgId)
}

// 发送多模态消息回复卡片
func sendMultimodalResponseCard(ctx context.Context, sessionId *string, msgId *string, content string) {
	sendSystemCard(ctx, fmt.Sprintf("🤖️：%s", content), sessionId, msgId)
}
