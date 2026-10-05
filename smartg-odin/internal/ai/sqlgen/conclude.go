package sqlgen

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"

	"smartg-odin/internal/ai"
	"smartg-odin/internal/common/config"
)

// concludeMaxRows 喂给 LLM 的每步结果最多行数，防止 prompt 过长。
const concludeMaxRows = 20

// concludeSystemPrompt 二次综合结论的系统提示词：强约束只输出结论正文。
const concludeSystemPrompt = "你是数据分析助手。基于给定的用户问题与 SQL 查询结果，" +
	"用 2-4 句中文直接回答问题，必须引用结果中的具体数字/名称。" +
	"不要输出 JSON，不要解释过程，只输出结论正文。"

// Conclude 基于用户问题与实际查询结果，调用 LLM 生成自然语言结论。
//
// 将问题 + 各步 SQL + 结果数据（截断前 concludeMaxRows 行）喂给 LLM，
// 产出直接回答问题的 2-4 句结论。调用方式与 Generate 保持一致（复用 ai.NewChatModel）。
//
// 参数 aiCfg 为传给 Generate 的同一个 config 指针；调用失败或结果为空时返回 error，
// 由上层决定是否回退到通用兜底文案。
func Conclude(ctx context.Context, question string, steps []SQLStep, results []StepResult, cfg *SQLGenConfig, aiCfg *config.AIConfig) (string, error) {
	if aiCfg == nil {
		return "", fmt.Errorf("aiCfg 为空，无法构建 ChatModel")
	}
	if strings.TrimSpace(question) == "" {
		return "", fmt.Errorf("question 为空")
	}

	cm, err := ai.NewChatModel(ctx, *aiCfg)
	if err != nil {
		return "", fmt.Errorf("构建 ChatModel 失败: %w", err)
	}

	userPrompt := buildConcludeUserPrompt(question, steps, results)

	convo := []*schema.Message{
		schema.SystemMessage(concludeSystemPrompt),
		schema.UserMessage(userPrompt),
	}

	resp, gerr := cm.Generate(ctx, convo)
	if gerr != nil {
		return "", fmt.Errorf("LLM 结论生成调用失败: %w", gerr)
	}
	if resp == nil {
		return "", fmt.Errorf("LLM 结论生成返回空响应")
	}
	content := strings.TrimSpace(resp.Content)
	if content == "" {
		return "", fmt.Errorf("LLM 结论生成返回空内容")
	}
	return content, nil
}

// buildConcludeUserPrompt 构建喂给 LLM 的 user 提示词：用户问题 + 各步格式化结果。
//
// 优先使用 results（含执行后的列/行数据）；results 缺失时回退到 steps 描述。
func buildConcludeUserPrompt(question string, steps []SQLStep, results []StepResult) string {
	var sb strings.Builder
	sb.WriteString("用户问题：")
	sb.WriteString(strings.TrimSpace(question))
	sb.WriteString("\n\n查询结果：\n")

	if len(results) > 0 {
		for i := range results {
			sb.WriteString(formatStepResult(i, results[i]))
			sb.WriteString("\n")
		}
	} else {
		// 无执行结果时（异常情况）退回仅展示 SQL 步骤。
		for i := range steps {
			purpose := strings.TrimSpace(steps[i].Purpose)
			if purpose == "" {
				purpose = fmt.Sprintf("步骤 %d", i+1)
			}
			sb.WriteString(fmt.Sprintf("【%s】SQL: %s\n执行失败: 无查询结果\n", purpose, strings.TrimSpace(steps[i].SQL)))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// formatStepResult 将单步执行结果格式化为可读文本块。
//
// 形如：
//
//	【{purpose}】SQL: {sql}
//	列: {col1, col2, ...}
//	数据(前20行):
//	col1=val1, col2=val2
//	...
//
// 若该步执行失败，则以「执行失败: {error}」标注。
func formatStepResult(idx int, sr StepResult) string {
	purpose := strings.TrimSpace(sr.Purpose)
	if purpose == "" {
		purpose = fmt.Sprintf("步骤 %d", idx+1)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("【%s】SQL: %s\n", purpose, strings.TrimSpace(sr.SQL)))

	if errMsg := strings.TrimSpace(sr.Error); errMsg != "" {
		sb.WriteString(fmt.Sprintf("执行失败: %s\n", errMsg))
		return sb.String()
	}

	// 列名。
	cols := make([]string, 0, len(sr.Columns))
	for _, c := range sr.Columns {
		cols = append(cols, c.Name)
	}
	sb.WriteString("列: ")
	sb.WriteString(strings.Join(cols, ", "))
	sb.WriteString("\n")

	// 行数据（截断前 concludeMaxRows 行）。
	total := len(sr.Rows)
	limit := total
	if limit > concludeMaxRows {
		limit = concludeMaxRows
	}
	sb.WriteString(fmt.Sprintf("数据(前%d行):\n", concludeMaxRows))
	if total == 0 {
		sb.WriteString("(无数据行)\n")
		return sb.String()
	}
	for r := 0; r < limit; r++ {
		sb.WriteString(formatRow(cols, sr.Rows[r]))
		sb.WriteString("\n")
	}
	if total > limit {
		sb.WriteString(fmt.Sprintf("...(共 %d 行，已截断展示前 %d 行)\n", total, limit))
	}
	return sb.String()
}

// formatRow 将单行数据格式化为 "col1=val1, col2=val2" 形式。
func formatRow(cols []string, row []interface{}) string {
	parts := make([]string, 0, len(row))
	for i, v := range row {
		name := fmt.Sprintf("col%d", i+1)
		if i < len(cols) && strings.TrimSpace(cols[i]) != "" {
			name = cols[i]
		}
		parts = append(parts, fmt.Sprintf("%s=%v", name, v))
	}
	return strings.Join(parts, ", ")
}
