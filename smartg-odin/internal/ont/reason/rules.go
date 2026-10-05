// Package reason 实现本体推理引擎：虚拟类展开、规则评估与执行、隐含关系推导。
package reason

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// RuleCondition 规则条件，对应 OntRule.ConditionJSON。
type RuleCondition struct {
	Type     string      `json:"type"`      // property_exists, value_match, relation_exists, threshold
	Property string      `json:"property"`  // 目标属性名
	Operator string      `json:"operator"`  // eq, neq, gt, lt, contains, exists
	Value    interface{} `json:"value"`     // 比较值
	ClassRef string      `json:"class_ref"` // 用于 relation_exists
}

// RuleAction 规则动作，对应 OntRule.ActionJSON。
type RuleAction struct {
	Type           string      `json:"type"` // add_filter, add_property, expand_class, set_value
	TargetProperty string      `json:"target_property"`
	Operator       string      `json:"operator"`
	Value          interface{} `json:"value"`
	TargetClass    string      `json:"target_class"`
}

// ParseCondition 解析 condition_json 字符串为 RuleCondition。
func ParseCondition(jsonStr string) (*RuleCondition, error) {
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return nil, fmt.Errorf("empty condition json")
	}
	var c RuleCondition
	if err := json.Unmarshal([]byte(jsonStr), &c); err != nil {
		return nil, fmt.Errorf("parse condition: %w", err)
	}
	return &c, nil
}

// ParseAction 解析 action_json 字符串为 RuleAction。
func ParseAction(jsonStr string) (*RuleAction, error) {
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return nil, fmt.Errorf("empty action json")
	}
	var a RuleAction
	if err := json.Unmarshal([]byte(jsonStr), &a); err != nil {
		return nil, fmt.Errorf("parse action: %w", err)
	}
	return &a, nil
}

// EvaluateCondition 在查询上下文中评估单条规则条件是否满足。
//
// 查询期规则评估基于「上下文引用」而非实际数据行，语义如下：
//   - property_exists: 上下文（过滤条件 / 额外列）中引用了该属性
//   - value_match:     上下文过滤条件中存在同名属性且值/操作符匹配
//   - threshold:       上下文过滤条件中存在同名属性的数值比较
//   - relation_exists: 上下文已展开 ClassRef 指定的关系/类
func EvaluateCondition(condition *RuleCondition, ctx *QueryContext) bool {
	if condition == nil || ctx == nil {
		return false
	}
	switch strings.ToLower(condition.Type) {
	case "property_exists":
		return ctx.ReferencesProperty(condition.Property)
	case "value_match":
		for _, f := range ctx.Filters {
			if strings.EqualFold(f.Column, condition.Property) &&
				strings.EqualFold(f.Op, condition.Operator) &&
				looseEqual(f.Value, condition.Value) {
				return true
			}
		}
		return false
	case "threshold":
		for _, f := range ctx.Filters {
			if strings.EqualFold(f.Column, condition.Property) {
				return true
			}
		}
		return false
	case "relation_exists":
		return ctx.ReferencesClass(condition.ClassRef)
	default:
		// 未知条件类型：只要引用了相关属性即认为满足，保证向前兼容。
		if condition.Property != "" {
			return ctx.ReferencesProperty(condition.Property)
		}
		return false
	}
}

// ExecuteAction 在查询上下文中执行规则动作，返回修改后的上下文。
//   - add_filter:    向 ctx.Filters 追加过滤条件
//   - add_property:  向 ctx.ExtraColumns 追加列
//   - expand_class:  记录需要展开的目标类
//   - set_value:     以固定值覆盖/追加过滤条件
func ExecuteAction(action *RuleAction, ctx *QueryContext) *QueryContext {
	if action == nil || ctx == nil {
		return ctx
	}
	switch strings.ToLower(action.Type) {
	case "add_filter":
		if action.TargetProperty != "" {
			ctx.Filters = append(ctx.Filters, ContextFilter{
				Column: action.TargetProperty,
				Op:     defaultOp(action.Operator),
				Value:  action.Value,
			})
		}
	case "add_property":
		if action.TargetProperty != "" && !containsStr(ctx.ExtraColumns, action.TargetProperty) {
			ctx.ExtraColumns = append(ctx.ExtraColumns, action.TargetProperty)
		}
	case "expand_class":
		if action.TargetClass != "" && !containsStr(ctx.ExpandedClasses, action.TargetClass) {
			ctx.ExpandedClasses = append(ctx.ExpandedClasses, action.TargetClass)
		}
	case "set_value":
		if action.TargetProperty != "" {
			ctx.Filters = append(ctx.Filters, ContextFilter{
				Column: action.TargetProperty,
				Op:     "eq",
				Value:  action.Value,
			})
		}
	}
	return ctx
}

// defaultOp 为空操作符提供默认值 eq。
func defaultOp(op string) string {
	if strings.TrimSpace(op) == "" {
		return "eq"
	}
	return op
}

// containsStr 判断字符串切片是否包含目标（忽略大小写）。
func containsStr(list []string, target string) bool {
	for _, s := range list {
		if strings.EqualFold(s, target) {
			return true
		}
	}
	return false
}

// looseEqual 宽松比较两个 interface{} 是否相等（支持数值与字符串互转）。
func looseEqual(a, b interface{}) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	if reflect.DeepEqual(a, b) {
		return true
	}
	// 尝试数值比较
	if fa, oka := toFloat(a); oka {
		if fb, okb := toFloat(b); okb {
			return fa == fb
		}
	}
	// 退化为字符串比较
	return strings.EqualFold(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
}

// toFloat 尽力将 interface{} 转为 float64。
func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		return f, err == nil
	default:
		return 0, false
	}
}
