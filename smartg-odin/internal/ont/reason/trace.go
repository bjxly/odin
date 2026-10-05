package reason

// RuleFire 记录一条规则在推理过程中的评估与执行结果。
type RuleFire struct {
	RuleID          uint   `json:"rule_id"`
	Name            string `json:"name"`
	RuleType        string `json:"rule_type"`
	ConditionResult bool   `json:"condition_result"` // 条件是否命中
	Derived         bool   `json:"derived"`          // 是否以「派生当前类」方式注入过滤器
	Action          string `json:"action"`           // 动作摘要，如 expand_class:VipCustomer
	Detail          string `json:"detail"`           // 补充说明
}

// ExpansionStep 记录虚拟类展开的一个步骤。
type ExpansionStep struct {
	From   string `json:"from"`   // 源类名
	To     string `json:"to"`     // 目标类名
	Reason string `json:"reason"` // 展开原因，如 subclass / self-mapping / parent-fallback
}

// DerivedFilter 记录由规则派生注入的过滤条件。
type DerivedFilter struct {
	Property string      `json:"property"`
	Op       string      `json:"op"`
	Value    interface{} `json:"value"`
	Source   string      `json:"source"` // 来源规则名
}

// InferredRelation 记录一条被推导出的关系。
type InferredRelation struct {
	Name      string `json:"name"`
	Label     string `json:"label"`
	FromClass string `json:"from_class"`
	ToClass   string `json:"to_class"`
}

// ReasonTrace 汇总一次推理过程的全部可解释信息。
type ReasonTrace struct {
	RulesFired        []RuleFire         `json:"rules_fired"`
	Expansion         []ExpansionStep    `json:"expansion"`
	DerivedFilters    []DerivedFilter    `json:"derived_filters"`
	InferredRelations []InferredRelation `json:"inferred_relations"`
}

// NewReasonTrace 创建一个字段已初始化的空 ReasonTrace。
func NewReasonTrace() *ReasonTrace {
	return &ReasonTrace{
		RulesFired:        make([]RuleFire, 0),
		Expansion:         make([]ExpansionStep, 0),
		DerivedFilters:    make([]DerivedFilter, 0),
		InferredRelations: make([]InferredRelation, 0),
	}
}
