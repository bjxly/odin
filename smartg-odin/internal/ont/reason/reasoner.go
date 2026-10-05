package reason

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"smartg-odin/internal/model"
)

// ExpandedClass 虚拟类展开后的一个实际可查询目标。
type ExpandedClass struct {
	ClassID      uint
	ClassName    string
	SourceTable  string
	DataSourceID uint
	MappingJSON  string // property_mappings_json 原文
}

// ContextFilter 查询上下文中的一个过滤条件（以属性名表达，翻译阶段再映射为列名）。
type ContextFilter struct {
	Column string // 属性名
	Op     string
	Value  interface{}
}

// QueryContext 查询上下文，规则评估可对其进行修改。
type QueryContext struct {
	OntologyID      uint
	ClassName       string
	ClassID         uint
	Filters         []ContextFilter
	ExtraColumns    []string
	ExtraJoins      []string
	ExtraWheres     []string
	ExpandedClasses []string
	Params          []interface{}
	Trace           *ReasonTrace // 可选：非 nil 时收集推理过程用于可解释性
}

// ReferencesProperty 判断上下文是否引用了指定属性（出现在过滤条件或额外列中）。
func (c *QueryContext) ReferencesProperty(property string) bool {
	if property == "" {
		return false
	}
	for _, f := range c.Filters {
		if strings.EqualFold(f.Column, property) {
			return true
		}
	}
	for _, col := range c.ExtraColumns {
		if strings.EqualFold(col, property) {
			return true
		}
	}
	return false
}

// ReferencesClass 判断上下文是否已展开/引用了指定类或关系。
func (c *QueryContext) ReferencesClass(ref string) bool {
	if ref == "" {
		return false
	}
	if strings.EqualFold(c.ClassName, ref) {
		return true
	}
	for _, ec := range c.ExpandedClasses {
		if strings.EqualFold(ec, ref) {
			return true
		}
	}
	for _, j := range c.ExtraJoins {
		if strings.EqualFold(j, ref) {
			return true
		}
	}
	return false
}

// ExpandVirtualClass 将（可能是虚拟的）类展开为一组实际可查询的物理目标。
//
// 展开策略（按优先级）：
//  1. 非虚拟类：直接返回自身映射；
//  2. 虚拟类且存在子类（parent_class_id 指向它）：展开为所有子类；
//  3. 虚拟类自身带有映射配置：返回自身映射（如 VipCustomer→customers）；
//  4. 虚拟类无自身映射但有父类：回退到父类映射。
func ExpandVirtualClass(db *gorm.DB, ontologyID uint, className string) ([]ExpandedClass, error) {
	return ExpandVirtualClassWithTrace(db, ontologyID, className, nil)
}

// ExpandVirtualClassWithTrace 与 ExpandVirtualClass 相同，但在 trace 非 nil 时收集展开步骤。
func ExpandVirtualClassWithTrace(db *gorm.DB, ontologyID uint, className string, trace *ReasonTrace) ([]ExpandedClass, error) {
	var cls model.OntClass
	if err := db.Where("ontology_id = ? AND name = ?", ontologyID, className).First(&cls).Error; err != nil {
		return nil, fmt.Errorf("class %q not found in ontology %d: %w", className, ontologyID, err)
	}

	// 1. 非虚拟类：直接返回自身。
	if !strings.EqualFold(cls.ClassType, "virtual") {
		ec, err := buildExpandedClass(db, &cls)
		if err != nil {
			return nil, err
		}
		recordExpansion(trace, cls.Name, ec.ClassName, "non-virtual")
		return []ExpandedClass{*ec}, nil
	}

	result := make([]ExpandedClass, 0)

	// 2. 展开子类（显式按 id 升序，保证 UNION 分支顺序确定）。
	var subs []model.OntClass
	db.Where("ontology_id = ? AND parent_class_id = ?", ontologyID, cls.ID).
		Order("id asc").
		Find(&subs)
	for i := range subs {
		if ec, err := buildExpandedClass(db, &subs[i]); err == nil {
			result = append(result, *ec)
			recordExpansion(trace, cls.Name, ec.ClassName, "subclass")
		}
	}

	// 3. 无子类则尝试自身映射。
	if len(result) == 0 {
		if ec, err := buildExpandedClass(db, &cls); err == nil {
			result = append(result, *ec)
			recordExpansion(trace, cls.Name, ec.ClassName, "self-mapping")
		}
	}

	// 4. 仍无结果则回退父类映射。
	if len(result) == 0 && cls.ParentClassID != nil {
		var parent model.OntClass
		if err := db.First(&parent, *cls.ParentClassID).Error; err == nil {
			if ec, err := buildExpandedClass(db, &parent); err == nil {
				result = append(result, *ec)
				recordExpansion(trace, cls.Name, ec.ClassName, "parent-fallback")
			}
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("cannot expand virtual class %q: no mapping found", className)
	}
	return result, nil
}

// recordExpansion 在 trace 非 nil 时追加一条展开记录。
func recordExpansion(trace *ReasonTrace, from, to, reason string) {
	if trace == nil {
		return
	}
	trace.Expansion = append(trace.Expansion, ExpansionStep{From: from, To: to, Reason: reason})
}

// buildExpandedClass 依据类的映射配置构造 ExpandedClass。
func buildExpandedClass(db *gorm.DB, cls *model.OntClass) (*ExpandedClass, error) {
	var mc model.OntMappingConfig
	if err := db.Where("class_id = ?", cls.ID).Order("id asc").First(&mc).Error; err != nil {
		return nil, fmt.Errorf("mapping for class %q not found: %w", cls.Name, err)
	}
	return &ExpandedClass{
		ClassID:      cls.ID,
		ClassName:    cls.Name,
		SourceTable:  mc.SourceTable,
		DataSourceID: mc.DataSourceID,
		MappingJSON:  mc.PropertyMappingsJSON,
	}, nil
}

// ApplyRules 评估并应用指定本体下的所有启用规则，返回修改后的查询上下文。
//
// 规则按 priority 升序执行。对于「派生当前查询类」的规则
// （action=expand_class 且 target_class 命中 ctx.ClassName），
// 直接把规则条件转换为过滤条件注入上下文——这正是虚拟类的语义，
// 例如 VipCustomer = Customer WHERE level = 'VIP'。
func ApplyRules(db *gorm.DB, ctx *QueryContext) (*QueryContext, error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil query context")
	}

	var rules []model.OntRule
	db.Where("ontology_id = ? AND enabled = ?", ctx.OntologyID, true).
		Order("priority asc, id asc").
		Find(&rules)

	for _, r := range rules {
		cond, cerr := ParseCondition(r.ConditionJSON)
		act, aerr := ParseAction(r.ActionJSON)
		if cerr != nil || aerr != nil {
			// 规则解析失败不影响查询主流程，跳过即可。
			continue
		}

		actionSummary := summarizeAction(act)

		// 派生当前查询类：把条件作为过滤器注入。
		if strings.EqualFold(act.Type, "expand_class") &&
			strings.EqualFold(act.TargetClass, ctx.ClassName) &&
			cond.Property != "" {
			op := defaultOp(cond.Operator)
			ctx.Filters = append(ctx.Filters, ContextFilter{
				Column: cond.Property,
				Op:     op,
				Value:  cond.Value,
			})
			if ctx.Trace != nil {
				ctx.Trace.DerivedFilters = append(ctx.Trace.DerivedFilters, DerivedFilter{
					Property: cond.Property,
					Op:       op,
					Value:    cond.Value,
					Source:   r.Name,
				})
				ctx.Trace.RulesFired = append(ctx.Trace.RulesFired, RuleFire{
					RuleID:          r.ID,
					Name:            r.Name,
					RuleType:        r.RuleType,
					ConditionResult: true,
					Derived:         true,
					Action:          actionSummary,
					Detail:          fmt.Sprintf("虚拟类 %s 语义化为过滤条件 %s %s %v", ctx.ClassName, cond.Property, op, cond.Value),
				})
			}
			continue
		}

		// 常规：条件满足则执行动作。
		matched := EvaluateCondition(cond, ctx)
		if matched {
			ctx = ExecuteAction(act, ctx)
		}
		if ctx.Trace != nil {
			ctx.Trace.RulesFired = append(ctx.Trace.RulesFired, RuleFire{
				RuleID:          r.ID,
				Name:            r.Name,
				RuleType:        r.RuleType,
				ConditionResult: matched,
				Derived:         false,
				Action:          actionSummary,
			})
		}
	}
	return ctx, nil
}

// summarizeAction 生成规则动作的可读摘要。
func summarizeAction(act *RuleAction) string {
	if act == nil {
		return ""
	}
	switch strings.ToLower(act.Type) {
	case "expand_class":
		return "expand_class:" + act.TargetClass
	case "add_filter", "set_value":
		return act.Type + ":" + act.TargetProperty + " " + defaultOp(act.Operator) + " " + fmt.Sprintf("%v", act.Value)
	case "add_property":
		return "add_property:" + act.TargetProperty
	default:
		return act.Type
	}
}

// InferRelations 推导与指定类相关的所有关系（直接关系 + 继承链上的关系）。
func InferRelations(db *gorm.DB, ontologyID uint, className string) ([]model.OntRelation, error) {
	return InferRelationsWithTrace(db, ontologyID, className, nil)
}

// InferRelationsWithTrace 与 InferRelations 相同，但在 trace 非 nil 时收集推导关系。
func InferRelationsWithTrace(db *gorm.DB, ontologyID uint, className string, trace *ReasonTrace) ([]model.OntRelation, error) {
	var cls model.OntClass
	if err := db.Where("ontology_id = ? AND name = ?", ontologyID, className).First(&cls).Error; err != nil {
		return nil, fmt.Errorf("class %q not found: %w", className, err)
	}

	// 收集继承链上的所有类 ID（自身 + 所有祖先 + 所有后代）。
	classIDs := collectInheritanceChain(db, ontologyID, &cls)

	relations := make([]model.OntRelation, 0)
	if len(classIDs) == 0 {
		return relations, nil
	}
	db.Where("ontology_id = ? AND (from_class_id IN ? OR to_class_id IN ?)", ontologyID, classIDs, classIDs).
		Order("id asc").
		Find(&relations)
	if trace != nil {
		names := map[uint]string{}
		var all []model.OntClass
		db.Where("ontology_id = ?", ontologyID).Find(&all)
		for _, c := range all {
			names[c.ID] = c.Name
		}
		for _, rel := range relations {
			trace.InferredRelations = append(trace.InferredRelations, InferredRelation{
				Name:      rel.Name,
				Label:     rel.Label,
				FromClass: names[rel.FromClassID],
				ToClass:   names[rel.ToClassID],
			})
		}
	}
	return relations, nil
}

// collectInheritanceChain 收集一个类的完整继承链 ID（祖先 + 自身 + 后代）。
func collectInheritanceChain(db *gorm.DB, ontologyID uint, cls *model.OntClass) []uint {
	seen := map[uint]bool{}
	var walkUp func(c *model.OntClass)
	var walkDown func(parentID uint)

	walkUp = func(c *model.OntClass) {
		if c == nil || seen[c.ID] {
			return
		}
		seen[c.ID] = true
		if c.ParentClassID != nil {
			var parent model.OntClass
			if err := db.First(&parent, *c.ParentClassID).Error; err == nil {
				walkUp(&parent)
			}
		}
	}
	walkDown = func(parentID uint) {
		var subs []model.OntClass
		db.Where("ontology_id = ? AND parent_class_id = ?", ontologyID, parentID).Find(&subs)
		for i := range subs {
			if !seen[subs[i].ID] {
				seen[subs[i].ID] = true
				walkDown(subs[i].ID)
			}
		}
	}

	walkUp(cls)
	walkDown(cls.ID)

	ids := make([]uint, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	return ids
}
