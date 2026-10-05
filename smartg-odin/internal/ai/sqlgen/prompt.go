package sqlgen

import (
	"fmt"
	"sort"
	"strings"
)

// BuildSystemPrompt 构建 SQL 生成的 system prompt。
//
// 内容顺序：角色定义 → 数据库 Schema（CREATE TABLE 风格，列后注释本体中文名）→
// 可用 JOIN 路径 → 业务同义词 → 虚拟类规则 → 输出格式要求（JSON）→ 铁律约束。
// 相同的上下文多次调用产出完全一致的字符串（确定性），便于复现与缓存。
func BuildSystemPrompt(ctx *SQLGenContext) string {
	var b strings.Builder

	// 1. 角色定义
	b.WriteString("你是 ODIN 本体数据洞察系统的 SQL 生成器。你的唯一任务：把用户的中文自然语言问题，" +
		"翻译为可在给定物理数据库上直接执行的 SELECT 查询，并严格以下方规定的 JSON 格式输出。\n\n")

	// 2. 数据库 Schema
	b.WriteString("【数据库 Schema】\n")
	if ctx == nil || len(ctx.Datasources) == 0 {
		b.WriteString("（无可用数据源）\n")
	} else {
		for _, ds := range ctx.Datasources {
			b.WriteString(fmt.Sprintf("-- 数据源 id=%d name=%s type=%s\n", ds.ID, ds.Name, ds.Type))
			if len(ds.Tables) == 0 {
				b.WriteString("（该数据源无表）\n")
			}
			for _, t := range ds.Tables {
				b.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", t.Name))
				for i, c := range t.Columns {
					line := "  " + c.Name + " " + c.Type
					if c.IsPrimary {
						line += " PRIMARY KEY"
					}
					comment := columnComment(c)
					if comment != "" {
						line += " COMMENT '" + escapeSingleQuote(comment) + "'"
					}
					if i < len(t.Columns)-1 {
						line += ","
					}
					b.WriteString(line + "\n")
				}
				b.WriteString(")")
				if strings.TrimSpace(t.Comment) != "" {
					b.WriteString(" COMMENT '" + escapeSingleQuote(t.Comment) + "'")
				}
				b.WriteString(";\n")
			}
			b.WriteString("\n")
		}
	}

	// 3. 可用 JOIN 路径
	b.WriteString("【可用 JOIN 路径】\n")
	if ctx == nil || len(ctx.Relations) == 0 {
		b.WriteString("（无预定义 JOIN 路径，可依据主外键列自行 JOIN）\n")
	} else {
		for _, r := range ctx.Relations {
			label := r.Label
			if strings.TrimSpace(label) == "" {
				label = r.Name
			}
			b.WriteString(fmt.Sprintf("- %s (%s→%s→%s, %s)\n", r.JoinSQL, r.FromClass, label, r.ToClass, r.Cardinality))
		}
	}
	b.WriteString("\n")

	// 4. 业务同义词
	b.WriteString("【业务同义词】\n")
	if ctx == nil || len(ctx.Synonyms) == 0 {
		b.WriteString("（无）\n")
	} else {
		keys := make([]string, 0, len(ctx.Synonyms))
		for k := range ctx.Synonyms {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			syn := ctx.Synonyms[k]
			if len(syn) == 0 {
				continue
			}
			b.WriteString(fmt.Sprintf("- %s ← %s\n", k, strings.Join(syn, " / ")))
		}
	}
	b.WriteString("\n")

	// 5. 虚拟类规则
	b.WriteString("【虚拟类/派生规则】\n")
	if ctx == nil || len(ctx.Rules) == 0 {
		b.WriteString("（无）\n")
	} else {
		for _, r := range ctx.Rules {
			desc := strings.TrimSpace(r.Description)
			if desc == "" {
				desc = r.Name
			}
			line := fmt.Sprintf("- %s(%s)", r.Name, desc)
			if strings.TrimSpace(r.BaseTable) != "" {
				line += " = " + r.BaseTable
			}
			if strings.TrimSpace(r.Condition) != "" {
				line += " WHERE " + r.Condition
			}
			b.WriteString(line + "\n")
		}
	}
	b.WriteString("\n")

	// 6. 输出格式要求
	b.WriteString("【输出格式】\n")
	b.WriteString("你必须以如下 JSON 格式回复（不要输出任何其他内容，不要用 Markdown 代码块包裹）：\n")
	b.WriteString(`{
  "steps": [
    {"purpose": "步骤说明", "sql": "SELECT ...", "datasource_id": 1}
  ],
  "merge_strategy": "single|join|side_by_side|synthesize",
  "merge_keys": ["用于合并的列名"],
  "conclusion": "基于查询结果的自然语言结论"
}` + "\n\n")

	// 7. 铁律约束
	b.WriteString("【铁律约束】\n")
	b.WriteString("- merge_strategy 说明：single=只有一条 SQL 直接返回; join=多条 SQL 按 merge_keys 做内存 JOIN 合并为一张表; " +
		"side_by_side=多条 SQL 结果并列展示; synthesize=需要综合分析多条结果给出结论。\n")
	b.WriteString("- 只生成 SELECT 语句（可含 WITH/CTE/子查询/窗口函数/JOIN/GROUP BY/HAVING/ORDER BY/LIMIT）；" +
		"严禁 INSERT/UPDATE/DELETE/DROP/ALTER/TRUNCATE/CREATE 等任何写操作或 DDL。\n")
	b.WriteString("- 表名和列名必须严格来自上面给出的 Schema，不得编造。\n")
	b.WriteString("- SELECT 输出列使用 AS 起中文别名（取自本体 label），如 SELECT amount AS '金额'。\n")
	b.WriteString("- 跨数据源的查询必须拆为多条 SQL（每条指定对应的 datasource_id）；单数据源能一条解决就只生成一条。\n")
	b.WriteString("- 每条 SQL 必须显式包含 LIMIT（不超过系统上限），避免返回过多行。\n")
	b.WriteString("- conclusion 字段在生成时先写“待查询后补充”，系统会用实际结果更新。\n")

	return b.String()
}

// BuildUserPrompt 构建用户消息。当前直接透传问题原文，
// 独立成函数以便后续注入澄清上下文或多轮历史。
func BuildUserPrompt(question string) string {
	return strings.TrimSpace(question)
}

// columnComment 组合列的物理注释与本体中文语义，作为 CREATE TABLE 中的 COMMENT。
func columnComment(c ColumnSchema) string {
	parts := make([]string, 0, 2)
	if lbl := strings.TrimSpace(c.OntLabel); lbl != "" {
		parts = append(parts, lbl)
	}
	if cm := strings.TrimSpace(c.Comment); cm != "" && (len(parts) == 0 || !strings.EqualFold(cm, parts[0])) {
		parts = append(parts, cm)
	}
	return strings.Join(parts, " / ")
}

// escapeSingleQuote 转义单引号，避免破坏 prompt 中的 SQL 字面量。
func escapeSingleQuote(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
