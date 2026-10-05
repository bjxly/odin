package sqlgen

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"

	"smartg-odin/internal/ai"
	"smartg-odin/internal/common/config"
	"smartg-odin/internal/common/logger"
)

// Generate 调用 LLM 基于装配好的上下文生成结构化 SQL（单条或多条）。
//
// 流程：构建 system/user prompt → 调用 OpenAI 兼容 ChatModel（temperature=0、固定 seed，
// 由 chatModelConfig 承载）→ 容错解析 JSON → 解析失败或 steps 非法时带纠错信息重试，
// 重试次数上限为 cfg.MaxRepair。
//
// 说明：DeepSeek 等模型在思考模式下拒绝强制 tool_choice（已知限制），故此处不使用
// ForcedTool，而是以 prompt 强约束「只输出 JSON」，并在解析层做代码块/裸 JSON 容错。
func Generate(ctx context.Context, sqlCtx *SQLGenContext, question string, cfg *SQLGenConfig, chatModelConfig *config.AIConfig) (*SQLGenResult, error) {
	if chatModelConfig == nil {
		return nil, fmt.Errorf("chatModelConfig 为空，无法构建 ChatModel")
	}
	if strings.TrimSpace(question) == "" {
		return nil, fmt.Errorf("question 为空")
	}
	if cfg == nil {
		cfg = DefaultConfig()
	}
	cfg = cfg.Normalize()

	cm, err := ai.NewChatModel(ctx, *chatModelConfig)
	if err != nil {
		return nil, fmt.Errorf("构建 ChatModel 失败: %w", err)
	}

	sysPrompt := BuildSystemPrompt(sqlCtx)
	userPrompt := BuildUserPrompt(question)

	convo := []*schema.Message{
		schema.SystemMessage(sysPrompt),
		schema.UserMessage(userPrompt),
	}

	var lastErr error
	for attempt := 0; attempt <= cfg.MaxRepair; attempt++ {
		resp, gerr := cm.Generate(ctx, convo)
		if gerr != nil {
			lastErr = fmt.Errorf("LLM 调用失败: %w", gerr)
			// 调用层错误直接重试（追加一条继续指令）。
			convo = append(convo, schema.UserMessage("上一次调用未完成，请重新只输出符合格式要求的 JSON。"))
			continue
		}
		content := ""
		if resp != nil {
			content = resp.Content
		}
		convo = append(convo, resp)

		result, perr := parseSQLGenResult(content)
		if perr == nil {
			return result, nil
		}
		lastErr = perr
		logger.Warnf("[sqlgen] 第 %d 次解析失败: %v", attempt+1, perr)
		correction := "你上一次的输出无法解析为要求的 JSON，原因：" + perr.Error() +
			"\n请严格只输出符合格式的 JSON（不要输出解释文本，也不要用 Markdown 代码块包裹）。"
		convo = append(convo, schema.UserMessage(correction))
	}

	return nil, fmt.Errorf("SQL 生成失败（已重试 %d 次）: %w", cfg.MaxRepair, lastErr)
}

// parseSQLGenResult 从 LLM 原始文本中容错解析出 SQLGenResult。
//
// 支持三种形态：纯 JSON、```json ... ``` 代码块包裹、以及夹杂少量前后缀文本的 JSON。
// 解析成功后做最小合法性校验：steps 非空且每条 SQL 非空。
func parseSQLGenResult(content string) (*SQLGenResult, error) {
	jsonStr := extractJSONBlock(content)
	if strings.TrimSpace(jsonStr) == "" {
		return nil, fmt.Errorf("响应中未找到 JSON 内容")
	}
	var result SQLGenResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %w", err)
	}
	if len(result.Steps) == 0 {
		return nil, fmt.Errorf("steps 为空")
	}
	for i := range result.Steps {
		if strings.TrimSpace(result.Steps[i].SQL) == "" {
			return nil, fmt.Errorf("第 %d 条 step 的 sql 为空", i+1)
		}
	}
	if strings.TrimSpace(result.MergeStrategy) == "" {
		if len(result.Steps) == 1 {
			result.MergeStrategy = "single"
		} else {
			result.MergeStrategy = "side_by_side"
		}
	}
	if result.MergeKeys == nil {
		result.MergeKeys = []string{}
	}
	return &result, nil
}

// extractJSONBlock 从文本中抽取 JSON 主体：优先剥离 Markdown 代码块围栏，
// 否则截取首个 '{' 到末个 '}' 之间的子串。
func extractJSONBlock(content string) string {
	s := strings.TrimSpace(content)
	if s == "" {
		return ""
	}
	// 剥离 ```json ... ``` 或 ``` ... ``` 围栏。
	if idx := strings.Index(s, "```"); idx >= 0 {
		rest := s[idx+3:]
		// 去掉可能的语言标识（如 json）。
		if nl := strings.Index(rest, "\n"); nl >= 0 {
			lang := strings.TrimSpace(rest[:nl])
			if strings.EqualFold(lang, "json") || lang == "" {
				rest = rest[nl+1:]
			}
		}
		if end := strings.Index(rest, "```"); end >= 0 {
			candidate := strings.TrimSpace(rest[:end])
			if candidate != "" {
				return candidate
			}
		}
	}
	// 回退：截取首个 '{' 到末个 '}'。
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}
