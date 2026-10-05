package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"

	"smartg-odin/internal/common/config"
)

// NewChatModel 是 newChatModel 的导出入口，供 ADK Agent 编排（internal/ai/agent）复用，
// 保证意图链与 Agent 使用同一套 OpenAI 兼容 ChatModel 构建逻辑（BaseURL/Model/Temperature/Seed）。
func NewChatModel(ctx context.Context, cfg config.AIConfig) (*openai.ChatModel, error) {
	return newChatModel(ctx, cfg)
}

// newChatModel 用 eino-ext 的 OpenAI 兼容组件构建 ChatModel。
//
// BaseURL 可指向任意 OpenAI 兼容端点（DeepSeek/Qwen/vLLM/Ollama/Azure）。
// Temperature 与 Seed 用于确定性解码（同问同意图）。结构化输出由 ForcedTool
// (OntologyIntent) 承载，故此处不再设置 response_format，避免与 tool calling 冲突。
func newChatModel(ctx context.Context, cfg config.AIConfig) (*openai.ChatModel, error) {
	apiKey := cfg.ResolvedAPIKey()
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("ai.api_key 未配置（可通过 ${ODIN_AI_API_KEY} 环境变量注入）")
	}
	modelName := strings.TrimSpace(cfg.Model)
	if modelName == "" {
		modelName = "deepseek-chat"
	}
	temp := cfg.Temperature
	seed := cfg.Seed
	ocfg := &openai.ChatModelConfig{
		APIKey:      apiKey,
		BaseURL:     strings.TrimSpace(cfg.BaseURL),
		Model:       modelName,
		Temperature: &temp,
		Seed:        &seed,
	}
	cm, err := openai.NewChatModel(ctx, ocfg)
	if err != nil {
		return nil, fmt.Errorf("构建 openai ChatModel 失败: %w", err)
	}
	return cm, nil
}
