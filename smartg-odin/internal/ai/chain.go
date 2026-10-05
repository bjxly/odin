package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"smartg-odin/internal/common/config"
	"smartg-odin/internal/common/logger"
	"smartg-odin/internal/ds/query"
)

// LLMMeta LLM 路径的可复现审计元数据，写入 query_trace.llm_json（见 设计文档 §3.1、§6.1）。
type LLMMeta struct {
	Model            string  `json:"model"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	Seed             int     `json:"seed"`
	Temperature      float32 `json:"temperature"`
	PromptVersion    string  `json:"prompt_version"`
	VocabVersion     string  `json:"vocab_version"`
	LatencyMs        int64   `json:"latency_ms"`
	Repairs          int     `json:"repairs"`
}

// intentChainResult 意图链的最终输出。
type intentChainResult struct {
	Request *query.QueryRequest
	Meta    *LLMMeta
	Issues  []string
	OK      bool
}

// runIntentChain 构建并执行 Eino 意图链：
//
//	ChatModel(BindForcedTools: OntologyIntent) → IntentValidator(Lambda: 词表校验 + 有界修复)
//
// 系统 prompt 内联 vocabulary 词表；ChatModel 以 ForcedTool 强约束只产出 OntologyIntent。
// 校验 Lambda 在词表越界时以修正提示重跑模型（≤ max_repair），绝不把非法 AST 交给 translator。
func runIntentChain(ctx context.Context, cfg config.AIConfig, vocab *Vocabulary, ontologyID uint, nlText string) (*intentChainResult, error) {
	cm, err := newChatModel(ctx, cfg)
	if err != nil {
		return nil, err
	}
	tool := ontologyIntentTool(vocab)
	// 优先 ForcedTool（tool_choice 指定为 OntologyIntent，结构约束最强）。
	if err := cm.BindForcedTools([]*schema.ToolInfo{tool}); err != nil {
		return nil, fmt.Errorf("绑定 ForcedTool 失败: %w", err)
	}
	sysPrompt := buildSystemPrompt(vocab)

	// IntentValidator（Lambda）：解析工具调用 → 词表校验 → 有界修复。
	validator := compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) (*intentChainResult, error) {
		return validateAndRepair(ctx, cm, cfg, vocab, ontologyID, sysPrompt, nlText, msg)
	})

	chain := compose.NewChain[[]*schema.Message, *intentChainResult]().
		AppendChatModel(cm).
		AppendLambda(validator)

	runnable, err := chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("编译 Eino 意图链失败: %w", err)
	}

	input := []*schema.Message{
		schema.SystemMessage(sysPrompt),
		schema.UserMessage(nlText),
	}
	start := time.Now()
	out, err := runnable.Invoke(ctx, input)
	// 兼容性回退：部分模型（如 DeepSeek 思考模式）拒绝「指定函数」的强制 tool_choice，
	// 返回 400 "Thinking mode does not support this tool_choice"。此时降级为 tool_choice=auto
	// 重试一次——结构约束仍由工具参数 JSON Schema 的 enum 与 IntentValidator 词表校验兜底，
	// 不会产生幻觉 AST；系统提示已强制要求以 OntologyIntent 工具形式输出。
	if err != nil && isToolChoiceIncompatible(err) {
		logger.Warnf("[ai] ForcedTool 强制 tool_choice 不被模型支持，回退 tool_choice=auto 重试: %v", err)
		if berr := cm.BindTools([]*schema.ToolInfo{tool}); berr != nil {
			return nil, fmt.Errorf("回退绑定工具(auto)失败: %w", berr)
		}
		out, err = runnable.Invoke(ctx, input)
	}
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return nil, fmt.Errorf("意图链执行失败: %w", err)
	}
	if out != nil && out.Meta != nil {
		out.Meta.LatencyMs = latency
	}
	return out, nil
}

// isToolChoiceIncompatible 判断错误是否源于模型不支持当前的强制 tool_choice。
// 覆盖 OpenAI 兼容端点（DeepSeek/Qwen 等）在思考模式或旧版本下拒绝指定函数
// tool_choice 的常见报文，命中后由调用方降级为 auto 重试。
func isToolChoiceIncompatible(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "tool_choice") {
		return true
	}
	return strings.Contains(msg, "thinking") && strings.Contains(msg, "tool")
}

// validateAndRepair 是 IntentValidator 的实现：校验模型输出，越界则有界修复重试。
func validateAndRepair(ctx context.Context, cm model.BaseChatModel, cfg config.AIConfig,
	vocab *Vocabulary, ontologyID uint, sysPrompt, nlText string, msg *schema.Message) (*intentChainResult, error) {
	meta := &LLMMeta{
		Model:         cfg.Model,
		Seed:          cfg.Seed,
		Temperature:   cfg.Temperature,
		PromptVersion: promptVersion,
		VocabVersion:  vocab.Version(),
	}
	addUsage(meta, msg)
	req, issues := extractAndValidate(vocab, ontologyID, msg)

	maxRepair := cfg.MaxRepair
	if maxRepair < 0 {
		maxRepair = 0
	}

	convo := []*schema.Message{schema.SystemMessage(sysPrompt), schema.UserMessage(nlText), msg}
	repairs := 0
	for len(issues) > 0 && repairs < maxRepair {
		repairs++
		correction := "你上一次产出的意图存在如下词表违规，请严格依据系统词表修正后，重新调用 OntologyIntent 工具输出（不要输出任何解释文本）：\n- " +
			strings.Join(issues, "\n- ")
		convo = append(convo, schema.UserMessage(correction))
		fixMsg, err := cm.Generate(ctx, convo)
		if err != nil {
			meta.Repairs = repairs
			return &intentChainResult{Request: req, Meta: meta, Issues: issues, OK: false}, nil
		}
		addUsage(meta, fixMsg)
		convo = append(convo, fixMsg)
		msg = fixMsg
		req, issues = extractAndValidate(vocab, ontologyID, msg)
	}
	meta.Repairs = repairs
	ok := len(issues) == 0 && req != nil && strings.TrimSpace(req.ClassName) != ""
	return &intentChainResult{Request: req, Meta: meta, Issues: issues, OK: ok}, nil
}

// buildSystemPrompt 生成内联 vocabulary 词表的系统提示词（prompt_version 固定，保证可复现）。
func buildSystemPrompt(vocab *Vocabulary) string {
	var b strings.Builder
	b.WriteString("你是 ODIN 本体数据查询意图解析器。唯一任务：把用户的中文自然语言查询，转换为对本体词表严格约束的结构化意图，并通过调用 OntologyIntent 工具输出。\n")
	b.WriteString("硬性规则：\n")
	b.WriteString("1. 只能使用下方词表中出现的类名、属性名、关系名与操作符，绝不臆造词表之外的值。\n")
	b.WriteString("2. 必须以 OntologyIntent 工具调用形式输出，不要输出任何自然语言解释。\n")
	b.WriteString("3. 无法确定的字段留空（空数组 / 空串），不要猜测。\n")
	b.WriteString("4. 操作符仅可为：" + strings.Join(supportedOperators, ", ") + "。\n")
	b.WriteString(fmt.Sprintf("5. limit 取值范围 [%d,%d]；order_dir 仅可为 asc 或 desc。\n", limitMin, limitMax))

	b.WriteString("\n【本体类与属性词表】\n")
	for _, c := range vocab.Classes {
		b.WriteString("- 类 " + c.Name)
		if alias := classAlias(c); alias != "" {
			b.WriteString("（别名：" + alias + "）")
		}
		if strings.EqualFold(c.ClassType, "virtual") {
			b.WriteString("［虚拟类］")
		}
		b.WriteString("：属性 ")
		props := make([]string, 0, len(c.Properties))
		for _, p := range c.Properties {
			s := p.Name
			details := make([]string, 0, 3)
			if strings.TrimSpace(p.Label) != "" {
				details = append(details, "label="+p.Label)
			}
			if strings.TrimSpace(p.DataType) != "" {
				details = append(details, "type="+p.DataType)
			}
			if len(p.Synonyms) > 0 {
				details = append(details, "同义词="+strings.Join(p.Synonyms, "/"))
			}
			if len(details) > 0 {
				s += "(" + strings.Join(details, ",") + ")"
			}
			props = append(props, s)
		}
		b.WriteString(strings.Join(props, "; "))
		b.WriteString("\n")
	}

	if len(vocab.Relations) > 0 {
		b.WriteString("\n【关系词表】\n")
		for _, r := range vocab.Relations {
			b.WriteString("- " + r.Name)
			if strings.TrimSpace(r.Label) != "" {
				b.WriteString("（" + r.Label + "）")
			}
			b.WriteString(fmt.Sprintf("：%s → %s\n", r.From, r.To))
		}
	}
	return b.String()
}
