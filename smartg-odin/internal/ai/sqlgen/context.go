package sqlgen

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"smartg-odin/internal/model"
)

// propertyMappingJSON 对应 OntMappingConfig.PropertyMappingsJSON 数组中的单项。
type propertyMappingJSON struct {
	PropertyName        string `json:"property_name"`
	ColumnName          string `json:"column_name"`
	TransformExpression string `json:"transform_expression"`
}

// joinCondJSON 对应 OntRelation.JoinConditionJSON。
type joinCondJSON struct {
	FromColumn string `json:"from_column"`
	ToColumn   string `json:"to_column"`
	JoinType   string `json:"join_type"`
}

// ruleCondJSON 对应 OntRule.ConditionJSON（仅取 SQL 生成所需字段）。
type ruleCondJSON struct {
	Type     string      `json:"type"`
	Property string      `json:"property"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// columnSemantic 物理列关联到的本体语义（中文名 + 属性名）。
type columnSemantic struct {
	label    string
	property string
}

// BuildContext 从元数据库装配 LLM 生成 SQL 所需的完整上下文。
//
// 装配流程：
//  1. mapping_config → 得到本体涉及的数据源、物理表与「类属性↔物理列」映射；
//  2. schema_tables + schema_column → 按数据源分组的物理 Schema；
//  3. ont_class + ont_class_property → 中文 label 与同义词（语义层）；
//  4. ont_relation + JoinConditionJSON → 物理表列的 JOIN SQL 片段；
//  5. ont_rule(derivation) + ont_action(expand_class/add_filter) → 虚拟类/派生规则；
//  6. 为每个物理列回填本体中文 label。
//
// ontologyID 为十进制字符串（与 HTTP 层路径参数一致）。
func BuildContext(db *gorm.DB, ontologyID string) (*SQLGenContext, error) {
	if db == nil {
		return nil, fmt.Errorf("db 为空，无法装配 SQL 生成上下文")
	}
	oid64, err := strconv.ParseUint(strings.TrimSpace(ontologyID), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("非法的 ontology_id %q: %w", ontologyID, err)
	}
	oid := uint(oid64)

	out := &SQLGenContext{
		Datasources: []DatasourceSchema{},
		Relations:   []RelationInfo{},
		Synonyms:    map[string][]string{},
		Rules:       []RuleInfo{},
	}

	// ---- 语义层：类与属性 ----
	var classes []model.OntClass
	if err := db.Where("ontology_id = ?", oid).Order("id asc").Find(&classes).Error; err != nil {
		return nil, fmt.Errorf("查询本体类失败: %w", err)
	}
	className := make(map[uint]string, len(classes))
	classLabel := make(map[uint]string, len(classes))
	for _, c := range classes {
		className[c.ID] = c.Name
		lbl := strings.TrimSpace(c.Label)
		if lbl == "" {
			lbl = c.Name
		}
		classLabel[c.ID] = lbl
		// 类级同义词：key=类 label，value=同义词 + 类名。
		addSynonyms(out.Synonyms, lbl, c.Name, c.Synonyms)
	}

	var props []model.OntClassProperty
	if err := db.Where("ontology_id = ?", oid).Order("id asc").Find(&props).Error; err != nil {
		return nil, fmt.Errorf("查询本体属性失败: %w", err)
	}
	// propLabel[classID][propNameLower] = 中文 label
	propLabel := map[uint]map[string]string{}
	for _, p := range props {
		if _, ok := propLabel[p.ClassID]; !ok {
			propLabel[p.ClassID] = map[string]string{}
		}
		lbl := strings.TrimSpace(p.Label)
		if lbl == "" {
			lbl = p.Name
		}
		propLabel[p.ClassID][strings.ToLower(p.Name)] = lbl
		// 属性级同义词：key=属性 label，value=同义词 + 属性名。
		addSynonyms(out.Synonyms, lbl, p.Name, p.Synonyms)
	}

	// ---- 映射层：类 ↔ 物理表列 ----
	var mappings []model.OntMappingConfig
	if err := db.Where("ontology_id = ?", oid).Order("id asc").Find(&mappings).Error; err != nil {
		return nil, fmt.Errorf("查询映射配置失败: %w", err)
	}
	classDS := map[uint]uint{}                    // classID → datasourceID
	classTable := map[uint]string{}                // classID → 物理表名
	classPropToCol := map[uint]map[string]string{} // classID → propNameLower → 物理列名
	colSem := map[string]columnSemantic{}          // "dsID|table|col"(小写) → 语义
	involvedDS := map[uint]bool{}
	for _, m := range mappings {
		classDS[m.ClassID] = m.DataSourceID
		classTable[m.ClassID] = m.SourceTable
		involvedDS[m.DataSourceID] = true
		if _, ok := classPropToCol[m.ClassID]; !ok {
			classPropToCol[m.ClassID] = map[string]string{}
		}
		for _, pm := range parsePropertyMappings(m.PropertyMappingsJSON) {
			col := strings.TrimSpace(pm.ColumnName)
			pn := strings.TrimSpace(pm.PropertyName)
			if col == "" || pn == "" {
				continue
			}
			classPropToCol[m.ClassID][strings.ToLower(pn)] = col
			lbl := ""
			if lm, ok := propLabel[m.ClassID]; ok {
				lbl = lm[strings.ToLower(pn)]
			}
			key := colSemKey(m.DataSourceID, m.SourceTable, col)
			colSem[key] = columnSemantic{label: lbl, property: pn}
		}
	}

	// ---- 物理层：数据源 + Schema ----
	if len(involvedDS) > 0 {
		dsIDs := make([]uint, 0, len(involvedDS))
		for id := range involvedDS {
			dsIDs = append(dsIDs, id)
		}
		sort.Slice(dsIDs, func(i, j int) bool { return dsIDs[i] < dsIDs[j] })

		var dataSources []model.DataSource
		if err := db.Where("id IN ?", dsIDs).Find(&dataSources).Error; err != nil {
			return nil, fmt.Errorf("查询数据源失败: %w", err)
		}
		dsByID := make(map[uint]model.DataSource, len(dataSources))
		for _, ds := range dataSources {
			dsByID[ds.ID] = ds
		}

		var tables []model.SchemaTable
		if err := db.Where("datasource_id IN ?", dsIDs).Order("id asc").Find(&tables).Error; err != nil {
			return nil, fmt.Errorf("查询 schema 表失败: %w", err)
		}
		var columns []model.SchemaColumn
		if err := db.Where("datasource_id IN ?", dsIDs).Order("ordinal_pos asc, id asc").Find(&columns).Error; err != nil {
			return nil, fmt.Errorf("查询 schema 列失败: %w", err)
		}
		colsByTable := map[uint][]model.SchemaColumn{}
		for _, c := range columns {
			colsByTable[c.TableID] = append(colsByTable[c.TableID], c)
		}

		// 按数据源分组装配 DatasourceSchema。
		dsGroups := map[uint][]TableSchema{}
		for _, t := range tables {
			ts := TableSchema{Name: t.TableName, Comment: t.Comment, Columns: []ColumnSchema{}}
			for _, c := range colsByTable[t.ID] {
				cs := ColumnSchema{
					Name:      c.ColumnName,
					Type:      c.DataType,
					IsPrimary: c.IsPrimaryKey,
					Comment:   c.Comment,
				}
				if sem, ok := colSem[colSemKey(t.DataSourceID, t.TableName, c.ColumnName)]; ok {
					cs.OntLabel = sem.label
					cs.OntProperty = sem.property
				}
				ts.Columns = append(ts.Columns, cs)
			}
			dsGroups[t.DataSourceID] = append(dsGroups[t.DataSourceID], ts)
		}
		for _, id := range dsIDs {
			ds := dsByID[id]
			out.Datasources = append(out.Datasources, DatasourceSchema{
				ID:     id,
				Name:   ds.Name,
				Type:   ds.Type,
				Tables: dsGroups[id],
			})
		}
	}

	// ---- 关系层：JOIN 路径 ----
	var relations []model.OntRelation
	if err := db.Where("ontology_id = ?", oid).Order("id asc").Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("查询本体关系失败: %w", err)
	}
	for _, rel := range relations {
		fromTable := classTable[rel.FromClassID]
		toTable := classTable[rel.ToClassID]
		jc := parseJoinCond(rel.JoinConditionJSON)
		if fromTable == "" || toTable == "" || jc.FromColumn == "" || jc.ToColumn == "" {
			continue
		}
		card := strings.TrimSpace(rel.Cardinality)
		if card == "" {
			card = "1:N"
		}
		out.Relations = append(out.Relations, RelationInfo{
			Name:        rel.Name,
			Label:       rel.Label,
			FromClass:   className[rel.FromClassID],
			ToClass:     className[rel.ToClassID],
			Cardinality: card,
			JoinSQL:     fmt.Sprintf("%s.%s = %s.%s", fromTable, jc.FromColumn, toTable, jc.ToColumn),
			FromTable:   fromTable,
			ToTable:     toTable,
		})
	}

	// ---- 规则层：虚拟类/派生 ----
	var rules []model.OntRule
	if err := db.Where("ontology_id = ? AND rule_type = ? AND enabled = ?", oid, "derivation", true).
		Order("priority asc, id asc").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("查询本体规则失败: %w", err)
	}
	if len(rules) > 0 {
		ruleIDs := make([]uint, 0, len(rules))
		for _, r := range rules {
			ruleIDs = append(ruleIDs, r.ID)
		}
		var actions []model.OntAction
		if err := db.Where("rule_id IN ? AND action_type IN ?", ruleIDs, []string{"expand_class", "add_filter"}).
			Order("id asc").Find(&actions).Error; err != nil {
			return nil, fmt.Errorf("查询本体动作失败: %w", err)
		}
		actionsByRule := map[uint][]model.OntAction{}
		for _, a := range actions {
			actionsByRule[a.RuleID] = append(actionsByRule[a.RuleID], a)
		}
		for _, r := range rules {
			cond := parseRuleCond(r.ConditionJSON)
			var baseClassID uint
			for _, a := range actionsByRule[r.ID] {
				if a.TargetClassID != nil && *a.TargetClassID > 0 {
					baseClassID = *a.TargetClassID
					break
				}
			}
			out.Rules = append(out.Rules, RuleInfo{
				Name:        r.Name,
				Description: r.Description,
				Condition:   buildConditionString(cond, classPropToCol[baseClassID]),
				BaseClass:   className[baseClassID],
				BaseTable:   classTable[baseClassID],
			})
		}
	}

	return out, nil
}

// AllowedTables 返回上下文中所有物理表名的小写集合，供 Validate 做表名白名单校验。
func (c *SQLGenContext) AllowedTables() map[string]bool {
	allowed := map[string]bool{}
	if c == nil {
		return allowed
	}
	for _, ds := range c.Datasources {
		for _, t := range ds.Tables {
			if name := strings.ToLower(strings.TrimSpace(t.Name)); name != "" {
				allowed[name] = true
			}
		}
	}
	return allowed
}

// DatasourceIDs 返回上下文中所有数据源 ID（升序去重）。
func (c *SQLGenContext) DatasourceIDs() []uint {
	seen := map[uint]bool{}
	ids := make([]uint, 0)
	if c == nil {
		return ids
	}
	for _, ds := range c.Datasources {
		if !seen[ds.ID] {
			seen[ds.ID] = true
			ids = append(ids, ds.ID)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// colSemKey 构造列语义查表的键（数据源 + 表 + 列，统一小写）。
func colSemKey(dsID uint, table, column string) string {
	return fmt.Sprintf("%d|%s|%s", dsID, strings.ToLower(table), strings.ToLower(column))
}

// addSynonyms 将同义词写入表：key 为规范业务词，value 追加别名（去重、去空）。
func addSynonyms(m map[string][]string, canonical, name, rawSynonyms string) {
	canonical = strings.TrimSpace(canonical)
	if canonical == "" {
		return
	}
	cands := make([]string, 0)
	if n := strings.TrimSpace(name); n != "" && !strings.EqualFold(n, canonical) {
		cands = append(cands, n)
	}
	cands = append(cands, parseSynonyms(rawSynonyms)...)
	if len(cands) == 0 {
		return
	}
	existing := map[string]bool{}
	for _, s := range m[canonical] {
		existing[strings.ToLower(s)] = true
	}
	for _, s := range cands {
		s = strings.TrimSpace(s)
		if s == "" || strings.EqualFold(s, canonical) || existing[strings.ToLower(s)] {
			continue
		}
		existing[strings.ToLower(s)] = true
		m[canonical] = append(m[canonical], s)
	}
}

// parsePropertyMappings 解析属性映射 JSON（容错：非法输入返回 nil）。
func parsePropertyMappings(jsonStr string) []propertyMappingJSON {
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return nil
	}
	var list []propertyMappingJSON
	if err := json.Unmarshal([]byte(jsonStr), &list); err != nil {
		return nil
	}
	return list
}

// parseJoinCond 解析关系 JOIN 条件 JSON。
func parseJoinCond(jsonStr string) joinCondJSON {
	var jc joinCondJSON
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return jc
	}
	_ = json.Unmarshal([]byte(jsonStr), &jc)
	return jc
}

// parseRuleCond 解析规则条件 JSON。
func parseRuleCond(jsonStr string) ruleCondJSON {
	var rc ruleCondJSON
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return rc
	}
	_ = json.Unmarshal([]byte(jsonStr), &rc)
	return rc
}

// parseSynonyms 解析同义词字段：支持 JSON 数组或逗号分隔字符串。
func parseSynonyms(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if strings.HasPrefix(raw, "[") {
		var arr []string
		if err := json.Unmarshal([]byte(raw), &arr); err == nil {
			out := make([]string, 0, len(arr))
			for _, s := range arr {
				if s = strings.TrimSpace(s); s != "" {
					out = append(out, s)
				}
			}
			return out
		}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// buildConditionString 将规则条件转为可读的 SQL 过滤片段（属性名映射为物理列名）。
func buildConditionString(cond ruleCondJSON, propToCol map[string]string) string {
	prop := strings.TrimSpace(cond.Property)
	if prop == "" {
		return ""
	}
	col := prop
	if propToCol != nil {
		if c, ok := propToCol[strings.ToLower(prop)]; ok && strings.TrimSpace(c) != "" {
			col = c
		}
	}
	val := formatCondValue(cond.Value)
	switch strings.ToLower(strings.TrimSpace(cond.Operator)) {
	case "contains":
		return fmt.Sprintf("%s LIKE '%%%s%%'", col, trimQuote(val))
	case "neq", "ne":
		return fmt.Sprintf("%s <> %s", col, val)
	case "gt":
		return fmt.Sprintf("%s > %s", col, val)
	case "gte", "ge":
		return fmt.Sprintf("%s >= %s", col, val)
	case "lt":
		return fmt.Sprintf("%s < %s", col, val)
	case "lte", "le":
		return fmt.Sprintf("%s <= %s", col, val)
	default: // eq 及未指定
		return fmt.Sprintf("%s = %s", col, val)
	}
}

// formatCondValue 将条件值格式化为 SQL 字面量（字符串加单引号，数值原样）。
func formatCondValue(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return "NULL"
	case string:
		return "'" + strings.ReplaceAll(t, "'", "''") + "'"
	case bool:
		if t {
			return "1"
		}
		return "0"
	default:
		return fmt.Sprintf("%v", t)
	}
}

// trimQuote 去除字符串字面量两端的单引号（用于 LIKE 拼接）。
func trimQuote(s string) string {
	if len(s) >= 2 && strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'") {
		return s[1 : len(s)-1]
	}
	return s
}
