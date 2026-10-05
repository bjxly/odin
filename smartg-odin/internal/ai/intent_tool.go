package ai

import (
	"strings"

	"github.com/cloudwego/eino/schema"
)

// intentToolName 是强制工具（ForcedTool）的名称，LLM 只能以该工具调用形式产出意图。
const intentToolName = "OntologyIntent"

// ontologyIntentTool 依据词表构建 ForcedTool `OntologyIntent` 的参数 JSON Schema。
//
// 参数结构直接对应后端 query.QueryRequest（见 设计文档 §2.1），并通过：
//   - class_name / relations / filters[].op 的 Enum 施加硬词表约束；
//   - 各字段 Desc 内联可选类名、属性名与同义词，作为软约束提示。
//
// ontology_id 由服务端填充，不暴露给 LLM，避免其臆造。
func ontologyIntentTool(vocab *Vocabulary) *schema.ToolInfo {
	classEnum := vocab.classNames()
	relEnum := relationNames(vocab)

	tool := &schema.ToolInfo{
		Name: intentToolName,
		Desc: "把用户的自然语言数据查询意图，转换为受本体词表严格约束的结构化查询意图。" +
			"只能使用给定的类名、属性名、关系名与操作符；无法确定时留空对应字段，绝不臆造词表之外的值。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"class_name": {
				Type:     schema.String,
				Required: true,
				Enum:     classEnum,
				Desc:     "查询的主类名，必须取自本体类词表：" + strings.Join(classEnum, ", ") + "。" + classHint(vocab),
			},
			"properties": {
				Type:     schema.Array,
				ElemInfo: &schema.ParameterInfo{Type: schema.String},
				Desc:     "要返回的属性名列表（取自该类属性词表）；留空表示返回全部属性。",
			},
			"relations": {
				Type:     schema.Array,
				ElemInfo: &schema.ParameterInfo{Type: schema.String, Enum: relEnum},
				Desc:     "需要展开的关系名，必须取自关系词表：" + strings.Join(relEnum, ", ") + "；无则留空。",
			},
			"filters": {
				Type: schema.Array,
				ElemInfo: &schema.ParameterInfo{
					Type: schema.Object,
					SubParams: map[string]*schema.ParameterInfo{
						"property": {
							Type:     schema.String,
							Required: true,
							Desc:     "过滤属性名，必须是主类（或其基类）的属性名。",
						},
						"op": {
							Type:     schema.String,
							Required: true,
							Enum:     append([]string{}, supportedOperators...),
							Desc:     "比较操作符，必须取自实现集：" + strings.Join(supportedOperators, ", ") + "。",
						},
						"value": {
							Type: schema.String,
							Desc: "过滤值；is_null/is_not_null 无需值可留空，in/not_in 用英文逗号分隔多值。",
						},
					},
				},
				Desc: "过滤条件列表；无则留空。",
			},
			"order_by": {
				Type: schema.String,
				Desc: "排序字段：可为属性名，也可为 aggregates 的别名（TOP-N = 填聚合别名 + order_dir=desc + limit=N）；无则留空。",
			},
			"order_dir": {
				Type: schema.String,
				Enum: []string{"asc", "desc"},
				Desc: "排序方向，仅可为 asc 或 desc。",
			},
			"aggregates": {
				Type: schema.Array,
				ElemInfo: &schema.ParameterInfo{
					Type: schema.Object,
					SubParams: map[string]*schema.ParameterInfo{
						"func": {
							Type:     schema.String,
							Required: true,
							Enum:     append([]string{}, supportedAggregateFuncs...),
							Desc:     "聚合函数，必须取自：" + strings.Join(supportedAggregateFuncs, ", ") + "。count 可省略 property 表示 COUNT(*)。",
						},
						"property": {
							Type: schema.String,
							Desc: "被聚合的属性名（sum/avg/min/max 需数值列，即 classHint 中标记为可聚合的列）；count 时可留空表示 COUNT(*)。",
						},
						"alias": {
							Type: schema.String,
							Desc: "输出列名，选填。除非用户明确指定了列名，否则请留空——系统会自动生成统一的中文别名（如 COUNT(*)→计数、SUM(quantity)→合计数量、AVG(price)→平均单价），以保证与规则路径一致；不要自行填写英文别名。having/order_by 可引用该（自动生成的中文）别名。",
						},
					},
				},
				Desc: "聚合表达式列表；普通明细查询无聚合则留空。",
			},
			"group_by": {
				Type:     schema.Array,
				ElemInfo: &schema.ParameterInfo{Type: schema.String},
				Desc:     "分组属性名列表；跨类属性用关系名限定如 contains.name；无分组则留空。",
			},
			"having": {
				Type: schema.Array,
				ElemInfo: &schema.ParameterInfo{
					Type: schema.Object,
					SubParams: map[string]*schema.ParameterInfo{
						"alias": {
							Type:     schema.String,
							Required: true,
							Desc:     "聚合列别名（对应 aggregates 的 alias）。",
						},
						"op": {
							Type:     schema.String,
							Required: true,
							Enum:     append([]string{}, supportedOperators...),
							Desc:     "比较操作符（数值比较，常用 gt/gte/lt/lte/eq）。",
						},
						"value": {
							Type: schema.String,
							Desc: "聚合阈值（数值）。",
						},
					},
				},
				Desc: "聚合后过滤（HAVING）；无则留空。",
			},
			"limit": {
				Type: schema.Integer,
				Desc: "返回行数上限，取值范围 [1,10000]；不确定则留空由服务端取默认值。",
			},
			"offset": {
				Type: schema.Integer,
				Desc: "跳过的行数，>=0；无则留空。",
			},
		}),
	}
	return tool
}

// relationNames 返回按名称排序的关系名列表，用于 Enum 与描述。
func relationNames(vocab *Vocabulary) []string {
	names := make([]string, 0, len(vocab.Relations))
	for _, r := range vocab.Relations {
		names = append(names, r.Name)
	}
	return names
}

// classHint 生成「类名(label/同义词): 属性名...」的紧凑词表提示，内联进工具描述。
// 数值型属性（integer/float）后缀†标记，提示可作 SUM/AVG/MIN/MAX 聚合。
func classHint(vocab *Vocabulary) string {
	var b strings.Builder
	b.WriteString(" 各类可用属性（†=数值列可 SUM/AVG/MIN/MAX）：")
	for _, c := range vocab.Classes {
		b.WriteString("[")
		b.WriteString(c.Name)
		alias := classAlias(c)
		if alias != "" {
			b.WriteString("(")
			b.WriteString(alias)
			b.WriteString(")")
		}
		b.WriteString(": ")
		props := make([]string, 0, len(c.Properties))
		for _, p := range c.Properties {
			if isNumericDataType(p.DataType) {
				props = append(props, p.Name+"†")
			} else {
				props = append(props, p.Name)
			}
		}
		b.WriteString(strings.Join(props, ","))
		b.WriteString("] ")
	}
	return strings.TrimSpace(b.String())
}

// isNumericDataType 判断数据类型是否为可聚合的数值型。
func isNumericDataType(dt string) bool {
	switch strings.ToLower(strings.TrimSpace(dt)) {
	case "integer", "int", "float", "double", "decimal", "number", "long":
		return true
	default:
		return false
	}
}

// classAlias 返回类的 label 与同义词拼接（用于提示 LLM 对齐中文表达）。
func classAlias(c VocabClass) string {
	parts := make([]string, 0, len(c.Synonyms)+1)
	if strings.TrimSpace(c.Label) != "" {
		parts = append(parts, c.Label)
	}
	parts = append(parts, c.Synonyms...)
	return strings.Join(parts, "/")
}
