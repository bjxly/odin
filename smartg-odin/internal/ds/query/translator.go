package query

import (
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"smartg-odin/internal/model"
	"smartg-odin/internal/ont/reason"
)

// QueryRequest 前端提交的本体查询请求。
type QueryRequest struct {
	OntologyID uint     `json:"ontology_id"`
	ClassName  string   `json:"class_name"`
	Properties []string `json:"properties"` // 要查询的属性名列表，空=全部
	Filters    []Filter `json:"filters"`    // 过滤条件
	Relations  []string `json:"relations"`  // 要展开的关系名
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
	OrderBy    string   `json:"order_by"`
	OrderDir   string   `json:"order_dir"` // asc, desc

	// 聚合三字段（见冻结契约）。键名 snake_case，与 LLM 工具参数逐字一致。
	Aggregates []Aggregate    `json:"aggregates"`
	GroupBy    []string       `json:"group_by"`
	Having     []HavingClause `json:"having"`
}

// Aggregate 单个聚合表达式（以本体属性名表达）。
type Aggregate struct {
	Func     string `json:"func"`     // count|sum|avg|min|max
	Property string `json:"property"` // 属性名；count 时可为空表示 COUNT(*)
	Alias    string `json:"alias"`    // 输出列名，选填；缺省自动生成并保证唯一
}

// HavingClause 聚合后的过滤条件，引用 aggregate 别名。
type HavingClause struct {
	Alias string      `json:"alias"`
	Op    string      `json:"op"`
	Value interface{} `json:"value"`
}

// Filter 单个过滤条件（以本体属性名表达）。
type Filter struct {
	Property string      `json:"property"`
	Op       string      `json:"op"` // eq, neq, gt, gte, lt, lte, like, in, not_in, is_null, is_not_null
	Value    interface{} `json:"value"`
}

// ColumnMeta 结果列的语义元数据，与 columns 顺序对齐（业务列前缀）。
type ColumnMeta struct {
	Name string `json:"name"`          // 列名（=SQL AS 别名，业务列为中文 label）
	Role string `json:"role"`          // dimension | measure
	Agg  string `json:"agg,omitempty"` // 聚合函数或空
}

// TranslatedQuery 本体查询翻译为物理 SQL 后的结果。
type TranslatedQuery struct {
	SQL          string              `json:"sql"`
	DataSourceID uint                `json:"datasource_id"`
	Params       []interface{}       `json:"params"`
	Explanation  string              `json:"explanation"`
	ClassName    string              `json:"class_name"`
	SourceTable  string              `json:"source_table"`
	Joins        []JoinInfo          `json:"joins,omitempty"`
	ReasonTrace  *reason.ReasonTrace `json:"reason_trace,omitempty"`
	Aggregated   bool                `json:"aggregated"`
	ColumnMeta   []ColumnMeta        `json:"column_meta,omitempty"`
	// Warning 确定性警告信息（如 1:N 扇出放大风险），供前端提示用户。
	Warning string `json:"warning,omitempty"`
	// DroppedJoins 被确定性丢弃的关系名列表（1:N 扇出且未被引用时自动丢弃）。
	DroppedJoins []string `json:"dropped_joins,omitempty"`
}

// JoinInfo 描述翻译过程中生成的一个 JOIN。
type JoinInfo struct {
	Relation  string `json:"relation"`
	TableName string `json:"table_name"`
	Condition string `json:"condition"`
	// joinType 内部使用，携带 JOIN 类型（LEFT/INNER/RIGHT），不对外序列化。
	joinType string
}

// PropertyMapping 对应 OntMappingConfig.PropertyMappingsJSON 中的单项。
type PropertyMapping struct {
	PropertyID          uint              `json:"property_id"`
	PropertyName        string            `json:"property_name"`
	ColumnName          string            `json:"column_name"`
	TransformExpression string            `json:"transform_expression"`
	ValueMap            map[string]string `json:"value_map"`
	Confidence          float64           `json:"confidence"`
	ValidationStatus    string            `json:"validation_status"`
}

// JoinCondition 对应 OntRelation.JoinConditionJSON。
type JoinCondition struct {
	FromColumn string `json:"from_column"`
	ToColumn   string `json:"to_column"`
	JoinType   string `json:"join_type"`
}

// Translate 将本体查询请求翻译为可在物理数据源上执行的 SQL。
func Translate(db *gorm.DB, req *QueryRequest) (*TranslatedQuery, error) {
	if req == nil {
		return nil, fmt.Errorf("nil query request")
	}
	if req.OntologyID == 0 || strings.TrimSpace(req.ClassName) == "" {
		return nil, fmt.Errorf("ontology_id and class_name are required")
	}

	// 1. 查找目标类。
	var cls model.OntClass
	if err := db.Where("ontology_id = ? AND name = ?", req.OntologyID, req.ClassName).First(&cls).Error; err != nil {
		return nil, fmt.Errorf("class %q not found in ontology %d", req.ClassName, req.OntologyID)
	}

	// 2. 构建查询上下文并应用推理规则（虚拟类会注入派生过滤条件）。
	trace := reason.NewReasonTrace()
	qctx := &reason.QueryContext{
		OntologyID: req.OntologyID,
		ClassName:  req.ClassName,
		ClassID:    cls.ID,
		Trace:      trace,
	}
	for _, f := range req.Filters {
		qctx.Filters = append(qctx.Filters, reason.ContextFilter{Column: f.Property, Op: f.Op, Value: f.Value})
	}
	qctx, err := reason.ApplyRules(db, qctx)
	if err != nil {
		return nil, err
	}

	// 3. 展开（虚拟）类为一组实际物理目标。
	expanded, err := reason.ExpandVirtualClassWithTrace(db, req.OntologyID, req.ClassName, trace)
	if err != nil {
		return nil, err
	}

	// 跨源聚合拒绝：有聚合且虚拟类展开为多个物理目标（可能跨数据源）时明确拒绝。
	if len(req.Aggregates) > 0 && len(expanded) > 1 {
		return nil, fmt.Errorf("聚合查询暂不支持跨数据源")
	}

	var tq *TranslatedQuery
	if len(expanded) == 1 {
		tq, err = buildSingleQuery(db, req, expanded[0], qctx)
	} else {
		tq, err = buildUnionQuery(db, req, expanded, qctx)
	}
	if err != nil {
		return nil, err
	}
	tq.ReasonTrace = trace
	return tq, nil
}

// buildSingleQuery 为单个物理目标构建 SQL（支持明细与聚合两种形态）。
func buildSingleQuery(db *gorm.DB, req *QueryRequest, ec reason.ExpandedClass, qctx *reason.QueryContext) (*TranslatedQuery, error) {
	mappings := parseMappings(ec.MappingJSON)
	if len(mappings) == 0 {
		return nil, fmt.Errorf("class %q has no property mappings", ec.ClassName)
	}

	const mainAlias = "t0"
	quote := quoterForDatasource(db, ec.DataSourceID)
	labels := loadPropertyLabels(db, mappings)
	colResolver := newColumnResolver(mainAlias, mappings, quote, labels)
	scope := &colScope{main: colResolver, quote: quote}

	aggregated := len(req.Aggregates) > 0

	// P2-3 扇出修复：在聚合查询中，当存在解析到主类(1侧)的可放大度量（SUM/AVG/MIN/MAX）时，
	// 未被任何 filter/group_by/aggregate(relation.prop)/order_by 引用的 1:N JOIN 对当前查询无贡献，
	// 却会把主类度量按子表行数重复放大 → 确定性丢弃该 JOIN（不放大，正确）。
	// count 度量已由 COUNT(DISTINCT) 扇出保护，不在此丢弃范围（保持既有口径）。
	// 若 1:N JOIN 确被引用（保留）且仍有主类度量 → 无法安全合并，改为在 Warning 字段声明放大风险。
	relationsToJoin := req.Relations
	droppedJoins := make([]string, 0)
	if aggregated && hasMainClassMeasure(scope, req) {
		relationsToJoin = make([]string, 0, len(req.Relations))
		for _, relName := range req.Relations {
			if isRelationReferencedByQuery(req, relName, colResolver) {
				relationsToJoin = append(relationsToJoin, relName)
				continue
			}
			// 未被引用：仅丢弃 1:N（会扇出放大）；N:1/1:1 收敛 JOIN 不放大，保留。
			if strings.EqualFold(strings.TrimSpace(relationCardinality(db, req.OntologyID, relName)), "1:N") {
				droppedJoins = append(droppedJoins, relName)
			} else {
				relationsToJoin = append(relationsToJoin, relName)
			}
		}
	}

	b := NewBuilder()

	// FROM
	b.From(ec.SourceTable, mainAlias)

	// JOIN（含同源/方向/基数校验；跨类列寻址通过 scope.joins 支持）
	joins := make([]JoinInfo, 0)
	fanout := false               // 存在 1:N 关系（主表为 1 侧）时，count 需防扇出
	var firstOneToMany *joinBuild // 第一个 1:N JOIN（按基类分组计数的口径来源）
	for i, relName := range relationsToJoin {
		jb, jerr := buildJoin(db, req.OntologyID, relName, mainAlias, i+1, quote, ec.ClassID, ec.DataSourceID)
		if jerr != nil {
			return nil, jerr
		}
		if jb == nil {
			continue
		}
		b.Join(joinTypeOf(jb.info), jb.info.TableName, jb.targetAlias, jb.info.Condition)
		joins = append(joins, jb.info)
		scope.joins = append(scope.joins, &joinScope{
			relName: relName,
			res:     newColumnResolver(jb.targetAlias, jb.mappings, quote, jb.labels),
		})
		if strings.EqualFold(strings.TrimSpace(jb.cardinality), "1:N") {
			fanout = true
			if firstOneToMany == nil {
				firstOneToMany = jb
			}
		}
	}

	// P2-3 扇出警告：当 1:N JOIN 确实被引用且存在主类度量可能被放大时，生成 warning。
	var fanoutWarning string
	if aggregated && fanout {
		if hasMainClassMeasure(scope, req) {
			fanoutWarning = "存在 1:N join，主类度量可能扇出放大"
		}
	}

	// 1:N 扇出下 count(无 property) 的口径判定（任务 #54）：
	//   - group_by 引用基类列（按基实体自身分组，如「按客户名分组统计订单数」）：
	//     计数语义 = 每个基实体对应的子表行数 → COUNT(<子表别名>.<子表PK>)。
	//     不能用 COUNT(DISTINCT t0.pk)：每个分组内基实体恒为 1 行，计数恒=1，TOP-N 排序失去意义；
	//     也不能用 COUNT(*)：LEFT JOIN 下无子行的组会错误得 1（COUNT(t1.pk) 对 NULL 不计数，正确得 0）。
	//   - group_by 为空（标量，如「有多少客户下过单」）或仅引用子/关系列（统计基实体数去重）：
	//     保持 COUNT(DISTINCT t0.<基表PK>) 防重复计数基实体。
	groupByBase := false
	if aggregated && fanout {
		for _, g := range req.GroupBy {
			if scope.resolvesToMain(g) {
				groupByBase = true
				break
			}
		}
	}
	// 主表主键列（供 1:N 扇出下的 COUNT(DISTINCT t0.<pk>)，按需解析）。
	mainPK := ""
	countBasisNote := ""
	if aggregated && fanout && !groupByBase {
		mainPK = primaryKeyColumn(db, ec.DataSourceID, ec.SourceTable)
	}

	// SELECT + GROUP BY + HAVING（聚合形态）
	var selectCols []string
	var columnMeta []ColumnMeta
	var groupByCols []string
	aggExprByAlias := make(map[string]string)
	if aggregated {
		used := make(map[string]bool)
		groupByCols = make([]string, 0, len(req.GroupBy))
		// 1) group_by 维度列（保留请求顺序，中文 label 别名）
		for _, g := range req.GroupBy {
			g = strings.TrimSpace(g)
			if g == "" {
				continue
			}
			col, ok := scope.resolve(g)
			if !ok {
				return nil, fmt.Errorf("group_by 属性 %q 无法解析到物理列", g)
			}
			label := uniqueAlias(scope.labelFor(g), used)
			selectCols = append(selectCols, fmt.Sprintf("%s AS %s", col, quote(label)))
			groupByCols = append(groupByCols, col)
			columnMeta = append(columnMeta, ColumnMeta{Name: label, Role: "dimension"})
		}
		// 2) 聚合度量列（带 AS 别名）
		for _, a := range req.Aggregates {
			fn := strings.ToLower(strings.TrimSpace(a.Func))
			if !isSupportedAggFunc(fn) {
				return nil, fmt.Errorf("unsupported aggregate function: %s", a.Func)
			}
			var expr string
			if fn == "count" && strings.TrimSpace(a.Property) == "" {
				switch {
				case fanout && groupByBase && firstOneToMany != nil:
					// 按基类分组 + 1:N JOIN：计数 = 每个基实体对应的子表行数。
					// 多个 1:N JOIN 时取第一个（与 req.Relations 顺序一致，确定性），口径写入 explanation。
					childPK := primaryKeyColumn(db, firstOneToMany.targetDSID, firstOneToMany.info.TableName)
					expr = fmt.Sprintf("COUNT(%s.%s)", firstOneToMany.targetAlias, quote(childPK))
					countBasisNote = fmt.Sprintf(" 计数口径：按基类分组统计关系 [%s] 目标表 [%s] 的行数（%s），无关联行的组计 0。",
						firstOneToMany.info.Relation, firstOneToMany.info.TableName, expr)
				case fanout && mainPK != "":
					// 标量或按子/关系列分组：COUNT(DISTINCT t0.<pk>) 防止 1:N 扇出重复计数主实体。
					expr = fmt.Sprintf("COUNT(DISTINCT %s.%s)", mainAlias, quote(mainPK))
					countBasisNote = fmt.Sprintf(" 计数口径：1:N 关系下按主表主键去重计数基实体（%s）。", expr)
				default:
					expr = "COUNT(*)"
				}
			} else {
				col, ok := scope.resolve(a.Property)
				if !ok {
					return nil, fmt.Errorf("聚合属性 %q 无法解析到物理列", a.Property)
				}
				if fn == "count" && fanout {
					expr = fmt.Sprintf("COUNT(DISTINCT %s)", col)
				} else {
					expr = fmt.Sprintf("%s(%s)", strings.ToUpper(fn), col)
				}
			}
			alias := strings.TrimSpace(a.Alias)
			if alias == "" {
				propLabel := ""
				if strings.TrimSpace(a.Property) != "" {
					propLabel = scope.labelFor(a.Property)
				}
				alias = defaultAggAlias(fn, propLabel, used)
			} else {
				alias = uniqueAlias(alias, used)
			}
			selectCols = append(selectCols, fmt.Sprintf("%s AS %s", expr, quote(alias)))
			aggExprByAlias[alias] = expr
			columnMeta = append(columnMeta, ColumnMeta{Name: alias, Role: "measure", Agg: fn})
		}
		// 3) 行级溯源常量列
		selectCols = append(selectCols, srcLineageCols(db, ec, quote)...)
		b.Select(selectCols...)
	} else {
		selectCols = buildSelectColumns(req.Properties, colResolver, mappings)
		selectCols = append(selectCols, srcLineageCols(db, ec, quote)...)
		b.Select(selectCols...)
	}

	// WHERE（来自请求过滤 + 规则注入的上下文过滤）
	for _, f := range mergeFilters(qctx) {
		expr, params, err := colResolver.filterExpr(f.Op, f.Column, f.Value)
		if err != nil {
			return nil, err
		}
		if expr != "" {
			b.Where(expr, params...)
		}
	}
	// 规则注入的裸 WHERE 片段。
	for _, w := range qctx.ExtraWheres {
		b.Where(w)
	}

	// GROUP BY + HAVING（仅聚合形态；在 WHERE 之后发射，保证参数顺序正确）
	if aggregated {
		if len(groupByCols) > 0 {
			b.GroupBy(groupByCols...)
		}
		for _, h := range req.Having {
			expr, ok := aggExprByAlias[strings.TrimSpace(h.Alias)]
			if !ok {
				return nil, fmt.Errorf("having 引用的别名 %q 不在聚合列内", h.Alias)
			}
			hexpr, params, herr := havingExpr(expr, h.Op, h.Value)
			if herr != nil {
				return nil, herr
			}
			if hexpr != "" {
				b.Having(hexpr, params...)
			}
		}
	}

	// ORDER BY（允许填 aggregate 别名）
	if strings.TrimSpace(req.OrderBy) != "" {
		ob := strings.TrimSpace(req.OrderBy)
		if _, isAlias := aggExprByAlias[ob]; isAlias {
			b.OrderBy(quote(ob), req.OrderDir)
		} else if col, ok := scope.resolve(ob); ok {
			b.OrderBy(col, req.OrderDir)
		}
	}

	// LIMIT / OFFSET
	if req.Limit > 0 {
		b.Limit(req.Limit)
	}
	if req.Offset > 0 {
		b.Offset(req.Offset)
	}

	sqlStr, params := b.Build()

	explanation := buildExplanation(req, ec, mappings, joins, qctx, countBasisNote)
	if len(droppedJoins) > 0 {
		explanation += fmt.Sprintf(" 确定性丢弃未被引用的 1:N 关系 [%s]（避免扇出放大）。", strings.Join(droppedJoins, ", "))
	}
	if fanoutWarning != "" {
		explanation += " " + fanoutWarning + "。"
	}
	return &TranslatedQuery{
		SQL:          sqlStr,
		DataSourceID: ec.DataSourceID,
		Params:       params,
		Explanation:  explanation,
		ClassName:    req.ClassName,
		SourceTable:  ec.SourceTable,
		Joins:        joins,
		Aggregated:   aggregated,
		ColumnMeta:   columnMeta,
		Warning:      fanoutWarning,
		DroppedJoins: droppedJoins,
	}, nil
}

// buildUnionQuery 为多个物理目标构建 UNION ALL 查询（虚拟类展开为多个子类时使用）。
func buildUnionQuery(db *gorm.DB, req *QueryRequest, expanded []reason.ExpandedClass, qctx *reason.QueryContext) (*TranslatedQuery, error) {
	// 首期：多分支 UNION + 聚合直接拒绝（与跨源聚合拒绝口径一致）。
	if len(req.Aggregates) > 0 {
		return nil, fmt.Errorf("聚合查询暂不支持多分支（UNION）虚拟类展开")
	}
	subSQLs := make([]string, 0, len(expanded))
	allParams := make([]interface{}, 0)
	var datasourceID uint
	var mainQuote func(string) string

	for _, ec := range expanded {
		mappings := parseMappings(ec.MappingJSON)
		if len(mappings) == 0 {
			continue
		}
		const alias = "t0"
		quote := quoterForDatasource(db, ec.DataSourceID)
		labels := loadPropertyLabels(db, mappings)
		resolver := newColumnResolver(alias, mappings, quote, labels)

		b := NewBuilder()
		selCols := buildSelectColumns(req.Properties, resolver, mappings)
		selCols = append(selCols, srcLineageCols(db, ec, quote)...)
		b.Select(selCols...)
		b.From(ec.SourceTable, alias)
		for _, f := range mergeFilters(qctx) {
			expr, params, err := resolver.filterExpr(f.Op, f.Column, f.Value)
			if err != nil {
				return nil, err
			}
			if expr != "" {
				b.Where(expr, params...)
			}
		}
		sqlStr, params := b.Build()
		subSQLs = append(subSQLs, sqlStr)
		allParams = append(allParams, params...)
		if datasourceID == 0 {
			datasourceID = ec.DataSourceID
			mainQuote = quote
		}
	}

	if len(subSQLs) == 0 {
		return nil, fmt.Errorf("no queryable subclass found for virtual class %q", req.ClassName)
	}

	union := strings.Join(subSQLs, " UNION ALL ")
	// 外层统一排序与分页。
	var tail strings.Builder
	if mainQuote == nil {
		mainQuote = quoteIdent
	}
	if strings.TrimSpace(req.OrderBy) != "" {
		tail.WriteString(fmt.Sprintf(" ORDER BY %s", mainQuote(req.OrderBy)))
		if strings.EqualFold(strings.TrimSpace(req.OrderDir), "desc") {
			tail.WriteString(" DESC")
		} else {
			tail.WriteString(" ASC")
		}
	}
	if req.Limit > 0 {
		tail.WriteString(fmt.Sprintf(" LIMIT %d", req.Limit))
	}
	if req.Offset > 0 {
		tail.WriteString(fmt.Sprintf(" OFFSET %d", req.Offset))
	}

	explanation := fmt.Sprintf("虚拟类 %s 展开为 %d 个物理目标，使用 UNION ALL 合并查询。", req.ClassName, len(subSQLs))
	return &TranslatedQuery{
		SQL:          union + tail.String(),
		DataSourceID: datasourceID,
		Params:       allParams,
		Explanation:  explanation,
		ClassName:    req.ClassName,
		SourceTable:  strings.Join(collectTables(expanded), ", "),
	}, nil
}

// joinBuild buildJoin 的完整产物：除 JoinInfo 外，携带目标类映射/别名/基数/数据源，供跨类列寻址与计数口径解析。
type joinBuild struct {
	info        JoinInfo
	targetAlias string
	mappings    []PropertyMapping
	labels      map[string]string
	cardinality string
	targetDSID  uint
}

// buildJoin 依据关系名构建 JOIN 信息，并做同源/方向/基数校验。
//   - 关系/目标映射/连接条件缺失 → 返回 (nil, nil)，调用方跳过；
//   - 方向不符（rel.FromClassID != 主类ID）或跨数据源 → 返回明确错误。
func buildJoin(db *gorm.DB, ontologyID uint, relName, fromAlias string, aliasIndex int, quote func(string) string, mainClassID, mainDSID uint) (*joinBuild, error) {
	if quote == nil {
		quote = quoteIdent
	}
	var rel model.OntRelation
	if err := db.Where("ontology_id = ? AND name = ?", ontologyID, relName).First(&rel).Error; err != nil {
		return nil, nil
	}

	// 方向校验：关系必须从主类出发（FromClassID == 查询类ID）。
	if rel.FromClassID != mainClassID {
		return nil, fmt.Errorf("关系 %q 的方向与查询类不匹配（仅支持从主类出发的正向关系）", relName)
	}

	// 目标类映射 → 目标物理表（显式按 id 升序，保证多映射时取哪条确定）。
	var mc model.OntMappingConfig
	if err := db.Where("class_id = ?", rel.ToClassID).Order("id asc").First(&mc).Error; err != nil {
		return nil, nil
	}

	// 同数据源校验：JOIN 目标必须与主表在同一数据源（单条 SQL 无法跨库 JOIN）。
	if mc.DataSourceID != mainDSID {
		return nil, fmt.Errorf("关系 %q 跨数据源（主表 datasource=%d，目标表 datasource=%d），无法 JOIN", relName, mainDSID, mc.DataSourceID)
	}

	var jc JoinCondition
	if strings.TrimSpace(rel.JoinConditionJSON) != "" {
		_ = json.Unmarshal([]byte(rel.JoinConditionJSON), &jc)
	}
	if jc.FromColumn == "" || jc.ToColumn == "" {
		return nil, nil
	}

	targetAlias := fmt.Sprintf("t%d", aliasIndex)
	condition := fmt.Sprintf("%s.%s = %s.%s", fromAlias, quote(jc.FromColumn), targetAlias, quote(jc.ToColumn))
	ji := JoinInfo{
		Relation:  relName,
		TableName: mc.SourceTable,
		Condition: condition,
	}
	ji.joinType = strings.ToUpper(defaultStr(jc.JoinType, "LEFT"))

	targetMappings := parseMappings(mc.PropertyMappingsJSON)
	return &joinBuild{
		info:        ji,
		targetAlias: targetAlias,
		mappings:    targetMappings,
		labels:      loadPropertyLabels(db, targetMappings),
		cardinality: rel.Cardinality,
		targetDSID:  mc.DataSourceID,
	}, nil
}

// primaryKeyColumn 尽力解析物理表的主键列名（用于 1:N 扇出下的 COUNT DISTINCT）；
// 无法从 schema 元数据确定时回退为 "id"。
func primaryKeyColumn(db *gorm.DB, datasourceID uint, table string) string {
	if db == nil || datasourceID == 0 || strings.TrimSpace(table) == "" {
		return "id"
	}
	var tbl model.SchemaTable
	if err := db.Select("id").Where("datasource_id = ? AND table_name = ?", datasourceID, table).Order("id asc").First(&tbl).Error; err != nil {
		return "id"
	}
	var col model.SchemaColumn
	if err := db.Select("column_name").Where("table_id = ? AND is_primary_key = ?", tbl.ID, true).Order("ordinal_pos asc, id asc").First(&col).Error; err == nil && strings.TrimSpace(col.ColumnName) != "" {
		return col.ColumnName
	}
	return "id"
}

// joinTypeOf 从 JoinInfo 中取出 join 类型。
func joinTypeOf(ji JoinInfo) string {
	if ji.joinType == "" {
		return "LEFT"
	}
	return ji.joinType
}

// columnResolver 负责把本体属性名解析为带别名的物理列表达式。
type columnResolver struct {
	alias  string
	byName map[string]PropertyMapping
	labels map[string]string   // lower(属性名) → 中文 label（缺失回退属性名）
	quote  func(string) string // 方言相关的标识符引用函数
}

func newColumnResolver(alias string, mappings []PropertyMapping, quote func(string) string, labels map[string]string) *columnResolver {
	if quote == nil {
		quote = quoteIdent
	}
	r := &columnResolver{alias: alias, byName: make(map[string]PropertyMapping, len(mappings)), labels: labels, quote: quote}
	for _, m := range mappings {
		if m.PropertyName != "" {
			r.byName[strings.ToLower(m.PropertyName)] = m
		}
	}
	return r
}

// labelFor 返回属性的中文 label，缺失时回退为属性名本身。
func (r *columnResolver) labelFor(property string) string {
	if r.labels != nil {
		if l, ok := r.labels[strings.ToLower(property)]; ok && strings.TrimSpace(l) != "" {
			return l
		}
	}
	return property
}

// joinScope 描述一个 JOIN 目标类的列解析器（供跨类聚合寻址，P2 使用）。
type joinScope struct {
	relName string
	res     *columnResolver
}

// colScope 多类列解析作用域：主类 + 若干 JOIN 目标类。
// resolve 支持 "relationName.property" 限定；无限定且不歧义时按 主类→关系目标类 顺序解析。
type colScope struct {
	main  *columnResolver
	joins []*joinScope
	quote func(string) string
}

// resolve 把（可能带关系限定的）属性名解析为带别名的物理列表达式。
func (s *colScope) resolve(property string) (string, bool) {
	property = strings.TrimSpace(property)
	if property == "" {
		return "", false
	}
	if i := strings.Index(property, "."); i >= 0 {
		relName := strings.TrimSpace(property[:i])
		rest := strings.TrimSpace(property[i+1:])
		for _, j := range s.joins {
			if strings.EqualFold(j.relName, relName) {
				return j.res.column(rest)
			}
		}
		return "", false
	}
	if col, ok := s.main.column(property); ok {
		return col, true
	}
	for _, j := range s.joins {
		if col, ok := j.res.column(property); ok {
			return col, true
		}
	}
	return "", false
}

// resolvesToMain 判断（可能带关系限定的）属性名是否解析到主类（t0）列。
// 带 "关系名.属性" 限定的引用一律视为非主类列；无限定名按 colScope.resolve 的
// 同一口径以主类解析器优先判定，二者结论保持一致。
func (s *colScope) resolvesToMain(property string) bool {
	property = strings.TrimSpace(property)
	if property == "" || strings.Contains(property, ".") {
		return false
	}
	_, ok := s.main.column(property)
	return ok
}

// labelFor 返回属性（可带关系限定）的中文 label，缺失回退属性名。
func (s *colScope) labelFor(property string) string {
	name := strings.TrimSpace(property)
	if i := strings.Index(name, "."); i >= 0 {
		name = strings.TrimSpace(name[i+1:])
	}
	if l := s.main.labelFor(name); l != name {
		return l
	}
	for _, j := range s.joins {
		if l := j.res.labelFor(name); l != name {
			return l
		}
	}
	return name
}

// loadPropertyLabels 依据 mappings 的 PropertyID 批量查询 OntClassProperty.Label，
// 构建 lower(属性名)→中文 label 映射。用 PropertyID 精确定位，对虚拟类复用父类属性亦有效；
// label 缺失时回退为属性名。
func loadPropertyLabels(db *gorm.DB, mappings []PropertyMapping) map[string]string {
	out := make(map[string]string, len(mappings))
	if db == nil || len(mappings) == 0 {
		return out
	}
	ids := make([]uint, 0, len(mappings))
	for _, m := range mappings {
		if m.PropertyID > 0 {
			ids = append(ids, m.PropertyID)
		}
	}
	labelByID := make(map[uint]string, len(ids))
	if len(ids) > 0 {
		var props []model.OntClassProperty
		db.Select("id", "label").Where("id IN ?", ids).Find(&props)
		for _, p := range props {
			labelByID[p.ID] = p.Label
		}
	}
	for _, m := range mappings {
		if m.PropertyName == "" {
			continue
		}
		label := m.PropertyName
		if l, ok := labelByID[m.PropertyID]; ok && strings.TrimSpace(l) != "" {
			label = l
		}
		out[strings.ToLower(m.PropertyName)] = label
	}
	return out
}

// aggFuncLabels 聚合函数的中文可读前缀，用于缺省别名生成。
var aggFuncLabels = map[string]string{
	"count": "计数", "sum": "合计", "avg": "平均", "min": "最小", "max": "最大",
}

// supportedAggFuncs translator 支持的聚合函数集（单一来源）。
// 必须与 handler.supportedAggFuncs、ai.supportedAggregateFuncs、intent_tool 的 Enum 保持一致。
var supportedAggFuncs = []string{"count", "sum", "avg", "min", "max"}

// isSupportedAggFunc 判断聚合函数是否在实现集内。
func isSupportedAggFunc(fn string) bool {
	fn = strings.ToLower(strings.TrimSpace(fn))
	for _, f := range supportedAggFuncs {
		if f == fn {
			return true
		}
	}
	return false
}

// AggFuncLabelCN 返回聚合函数的中文缺省别名前缀（COUNT→计数、SUM→合计、AVG→平均、MIN→最小、MAX→最大）。
// 对外导出，供 ai 意图归一层（validator）复用，保证 rule / llm 两条 parse_path 的缺省别名口径完全一致。
func AggFuncLabelCN(fn string) string {
	if l, ok := aggFuncLabels[strings.ToLower(strings.TrimSpace(fn))]; ok {
		return l
	}
	return strings.ToLower(strings.TrimSpace(fn))
}

// DefaultAggAlias 依据聚合函数与属性中文 label 生成缺省别名（聚合函数中文 + 属性中文 label），
// 不做唯一化去重（去重由 uniqueAlias 在翻译期统一处理）。这是缺省别名的单一口径，
// rule 路径（nlparse）、llm 路径（ai.validateRequest）与 translator 内部均复用之。
func DefaultAggAlias(fn, propLabel string) string {
	base := AggFuncLabelCN(fn)
	if strings.TrimSpace(propLabel) != "" {
		return base + propLabel
	}
	return base
}

// defaultAggAlias 生成缺省聚合别名（聚合函数中文 + 属性中文 label），并保证在 used 内唯一。
func defaultAggAlias(fn, propLabel string, used map[string]bool) string {
	return uniqueAlias(DefaultAggAlias(fn, propLabel), used)
}

// 缺省 / 上限行数（与 handler.defaultQueryLimit、handler.maxQueryLimit、ai.limitMax 对齐）。
const (
	DefaultQueryLimit = 100
	MaxQueryLimit     = 10000
)

// NormalizeLimit 统一 limit/offset 兜底策略，是 POST /query 与 assistant SSE 两条链路的
// 唯一 limit 归一入口（ translator.Translate 保持纯粹、不再自行兜底，以保证同一 intent
// 在两链路产出逐字节一致的 SQL）。
//
// 规则（确定性、可复现）：
//   - 聚合查询（len(Aggregates)>0）：缺省不做 100 截断——分组聚合的分组数天然有界、
//     非分组聚合仅 1 行，强加 LIMIT 100 会在分组数>100 时静默截断，故保持 Limit=0（不加 LIMIT 子句）；
//   - 非聚合明细查询：缺省 Limit=DefaultQueryLimit(100)，维持既有兜底行为；
//   - 任何情况下 Limit 上限 MaxQueryLimit、Offset 非负。
func NormalizeLimit(req *QueryRequest) {
	if req == nil {
		return
	}
	if req.Limit < 0 {
		req.Limit = 0
	}
	aggregated := len(req.Aggregates) > 0
	if req.Limit == 0 && !aggregated {
		req.Limit = DefaultQueryLimit
	}
	if req.Limit > MaxQueryLimit {
		req.Limit = MaxQueryLimit
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
}

// uniqueAlias 若 alias 已被占用则追加数字后缀保证唯一，并把结果登记进 used。
func uniqueAlias(alias string, used map[string]bool) string {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		alias = "col"
	}
	final := alias
	for i := 2; used[final]; i++ {
		final = fmt.Sprintf("%s%d", alias, i)
	}
	used[final] = true
	return final
}

// havingExpr 依据操作符为聚合表达式构建参数化 HAVING 片段。
// 直接引用聚合表达式（而非别名），以兼容 PostgreSQL（其 HAVING 不支持输出列别名）。
func havingExpr(expr, op string, value interface{}) (string, []interface{}, error) {
	switch strings.ToLower(strings.TrimSpace(op)) {
	case "eq", "":
		return fmt.Sprintf("%s = ?", expr), []interface{}{value}, nil
	case "neq", "ne":
		return fmt.Sprintf("%s <> ?", expr), []interface{}{value}, nil
	case "gt":
		return fmt.Sprintf("%s > ?", expr), []interface{}{value}, nil
	case "gte", "ge":
		return fmt.Sprintf("%s >= ?", expr), []interface{}{value}, nil
	case "lt":
		return fmt.Sprintf("%s < ?", expr), []interface{}{value}, nil
	case "lte", "le":
		return fmt.Sprintf("%s <= ?", expr), []interface{}{value}, nil
	default:
		return "", nil, fmt.Errorf("unsupported having operator: %s", op)
	}
}

// column 返回属性对应的列表达式（含别名/transform），第二个返回值表示是否找到映射。
func (r *columnResolver) column(property string) (string, bool) {
	m, ok := r.byName[strings.ToLower(property)]
	if !ok {
		return "", false
	}
	if strings.TrimSpace(m.TransformExpression) != "" {
		return m.TransformExpression, true
	}
	return fmt.Sprintf("%s.%s", r.alias, r.quote(m.ColumnName)), true
}

// selectExpr 返回用于 SELECT 的表达式，附带中文 label 别名（缺失回退属性名）。
func (r *columnResolver) selectExpr(m PropertyMapping) string {
	base := fmt.Sprintf("%s.%s", r.alias, r.quote(m.ColumnName))
	if strings.TrimSpace(m.TransformExpression) != "" {
		base = m.TransformExpression
	}
	return fmt.Sprintf("%s AS %s", base, r.quote(r.labelFor(m.PropertyName)))
}

// filterExpr 依据操作符构建参数化 WHERE 片段。
func (r *columnResolver) filterExpr(op, property string, value interface{}) (string, []interface{}, error) {
	col, ok := r.column(property)
	if !ok {
		// 未映射的属性直接忽略，避免生成非法列名。
		return "", nil, nil
	}
	op = strings.ToLower(strings.TrimSpace(op))
	switch op {
	case "eq", "":
		return fmt.Sprintf("%s = ?", col), []interface{}{value}, nil
	case "neq", "ne":
		return fmt.Sprintf("%s <> ?", col), []interface{}{value}, nil
	case "gt":
		return fmt.Sprintf("%s > ?", col), []interface{}{value}, nil
	case "gte", "ge":
		return fmt.Sprintf("%s >= ?", col), []interface{}{value}, nil
	case "lt":
		return fmt.Sprintf("%s < ?", col), []interface{}{value}, nil
	case "lte", "le":
		return fmt.Sprintf("%s <= ?", col), []interface{}{value}, nil
	case "like":
		return fmt.Sprintf("%s LIKE ?", col), []interface{}{fmt.Sprintf("%%%v%%", value)}, nil
	case "in":
		vals := toSlice(value)
		if len(vals) == 0 {
			return "", nil, nil
		}
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(vals)), ",")
		return fmt.Sprintf("%s IN (%s)", col, placeholders), vals, nil
	case "not_in", "nin":
		vals := toSlice(value)
		if len(vals) == 0 {
			return "", nil, nil
		}
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(vals)), ",")
		return fmt.Sprintf("%s NOT IN (%s)", col, placeholders), vals, nil
	case "is_null":
		return fmt.Sprintf("%s IS NULL", col), nil, nil
	case "is_not_null", "not_null":
		return fmt.Sprintf("%s IS NOT NULL", col), nil, nil
	default:
		return "", nil, fmt.Errorf("unsupported filter operator: %s", op)
	}
}

// buildSelectColumns 依据请求属性列表构建 SELECT 列（非聚合明细路径）。
// 传入已构建的 resolver（含中文 label 映射），避免重复构造。
func buildSelectColumns(props []string, resolver *columnResolver, mappings []PropertyMapping) []string {
	if len(props) == 0 {
		cols := make([]string, 0, len(mappings))
		for _, m := range mappings {
			cols = append(cols, resolver.selectExpr(m))
		}
		return cols
	}
	cols := make([]string, 0, len(props))
	for _, p := range props {
		if m, ok := resolver.byName[strings.ToLower(p)]; ok {
			cols = append(cols, resolver.selectExpr(m))
		}
	}
	if len(cols) == 0 {
		// 指定的属性都没有映射时兜底选择全部。
		for _, m := range mappings {
			cols = append(cols, resolver.selectExpr(m))
		}
	}
	return cols
}

// mergeFilters 汇总上下文过滤条件（含请求过滤与规则注入的派生过滤），
// 并按 (property, op, value) 三元组去重。
//
// 去重动机：虚拟类查询（如 VipCustomer 规则派生 level eq VIP）当请求体也显式
// 带同一过滤时，若不去重会生成 WHERE t0.`level` = ? AND t0.`level` = ? 冗余条件。
// 结果虽正确但 SQL 冗余，故按三元组保留首次出现的过滤，跳过后续重复项。
//
// 确定性保证：遍历 qctx.Filters 切片（顺序稳定：请求过滤在前、规则派生在后，
// 规则本身按 priority/id 升序注入），map 仅用于 O(1) 判重、不参与顺序输出，
// 因此同请求产出的过滤序列逐次一致，最终 SQL 逐字节稳定。
func mergeFilters(qctx *reason.QueryContext) []reason.ContextFilter {
	if qctx == nil {
		return nil
	}
	if len(qctx.Filters) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(qctx.Filters))
	result := make([]reason.ContextFilter, 0, len(qctx.Filters))
	for _, f := range qctx.Filters {
		key := filterDedupeKey(f)
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, f)
	}
	return result
}

// filterDedupeKey 生成 (property, op, value) 三元组的规范化去重键。
//
// property 与 op 做小写+去空白归一，与下游 columnResolver.column（按
// strings.ToLower 匹配属性名）、filterExpr（op = strings.ToLower(strings.TrimSpace(op))）
// 的解析语义保持一致，确保「显式过滤」与「派生过滤」按同一口径判等，
// 不改变实际匹配行为；value 用 %v 统一格式化以兼容 string/数值等类型。
// 各段之间用 \x00 分隔，避免属性名或值内含分隔符造成的键碰撞。
func filterDedupeKey(f reason.ContextFilter) string {
	return strings.ToLower(strings.TrimSpace(f.Column)) + "\x00" +
		strings.ToLower(strings.TrimSpace(f.Op)) + "\x00" +
		fmt.Sprintf("%v", f.Value)
}

// parseMappings 解析 property_mappings_json。
func parseMappings(jsonStr string) []PropertyMapping {
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return nil
	}
	var list []PropertyMapping
	if err := json.Unmarshal([]byte(jsonStr), &list); err != nil {
		return nil
	}
	return list
}

// toSlice 尽力把 interface{} 转为 []interface{}，支持 in / not_in。
func toSlice(v interface{}) []interface{} {
	switch val := v.(type) {
	case []interface{}:
		return val
	case []string:
		out := make([]interface{}, len(val))
		for i := range val {
			out[i] = val[i]
		}
		return out
	case []int:
		out := make([]interface{}, len(val))
		for i := range val {
			out[i] = val[i]
		}
		return out
	case string:
		parts := strings.Split(val, ",")
		out := make([]interface{}, 0, len(parts))
		for _, p := range parts {
			out = append(out, strings.TrimSpace(p))
		}
		return out
	default:
		return []interface{}{v}
	}
}

// collectTables 收集展开目标的物理表名。
func collectTables(expanded []reason.ExpandedClass) []string {
	tables := make([]string, 0, len(expanded))
	for _, ec := range expanded {
		tables = append(tables, ec.SourceTable)
	}
	return tables
}

// buildExplanation 生成可读的翻译说明文本。countBasisNote 为 1:N 扇出下 count 口径说明（可为空）。
func buildExplanation(req *QueryRequest, ec reason.ExpandedClass, mappings []PropertyMapping, joins []JoinInfo, qctx *reason.QueryContext, countBasisNote string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("本体类 [%s] 映射到数据源表 [%s]（datasource_id=%d）。", req.ClassName, ec.SourceTable, ec.DataSourceID))
	sb.WriteString(fmt.Sprintf(" 共 %d 个属性映射。", len(mappings)))
	if len(joins) > 0 {
		sb.WriteString(fmt.Sprintf(" 展开 %d 个关系：", len(joins)))
		names := make([]string, 0, len(joins))
		for _, j := range joins {
			names = append(names, j.Relation+"→"+j.TableName)
		}
		sb.WriteString(strings.Join(names, ", ") + "。")
	}
	if len(req.Aggregates) > 0 {
		aggDescs := make([]string, 0, len(req.Aggregates))
		for _, a := range req.Aggregates {
			fn := strings.ToUpper(strings.TrimSpace(a.Func))
			if fn == "COUNT" && strings.TrimSpace(a.Property) == "" {
				aggDescs = append(aggDescs, "COUNT(*)")
			} else {
				aggDescs = append(aggDescs, fn+"("+a.Property+")")
			}
		}
		sb.WriteString(" 聚合查询：" + strings.Join(aggDescs, ", ") + "。")
		if len(req.GroupBy) > 0 {
			sb.WriteString(" 按 [" + strings.Join(req.GroupBy, ", ") + "] 分组。")
		}
		if len(req.Having) > 0 {
			sb.WriteString(fmt.Sprintf(" 含 %d 个 HAVING 过滤。", len(req.Having)))
		}
		sb.WriteString(countBasisNote)
	}
	derived := 0
	if qctx != nil {
		derived = len(qctx.Filters) - len(req.Filters)
	}
	if derived > 0 {
		sb.WriteString(fmt.Sprintf(" 推理规则注入了 %d 个派生过滤条件。", derived))
	}
	if req.Limit > 0 {
		sb.WriteString(fmt.Sprintf(" 限制返回 %d 行。", req.Limit))
	}
	return sb.String()
}

// quoteIdent 为标识符加双引号（ANSI/PostgreSQL/SQLite 方言），防止关键字冲突与注入。
func quoteIdent(name string) string {
	name = strings.TrimSpace(name)
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// quoteMySQL 为标识符加反引号（MySQL/MariaDB 方言），内部反引号翻倍转义。
func quoteMySQL(name string) string {
	name = strings.TrimSpace(name)
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

// quoterForDatasource 依据数据源类型返回标识符引用函数：
// MySQL/MariaDB 用反引号，其余（PostgreSQL/SQLite 等）用双引号。
// 查询失败或数据源未知时回退为双引号，保证向后兼容。
func quoterForDatasource(db *gorm.DB, id uint) func(string) string {
	if id == 0 {
		return quoteIdent
	}
	var ds model.DataSource
	if err := db.Select("type").First(&ds, id).Error; err != nil {
		return quoteIdent
	}
	switch strings.ToLower(strings.TrimSpace(ds.Type)) {
	case "mysql", "mariadb":
		return quoteMySQL
	default:
		return quoteIdent
	}
}

// srcLineageCols 为一条查询分支生成行级溯源常量列：_src_datasource 与 _src_table。
// 保证结果集每行都携带数据来源信息，无论单源还是 UNION 多源。
func srcLineageCols(db *gorm.DB, ec reason.ExpandedClass, quote func(string) string) []string {
	if quote == nil {
		quote = quoteIdent
	}
	dsLabel := datasourceLabel(db, ec.DataSourceID)
	return []string{
		fmt.Sprintf("%s AS %s", sqlLiteral(dsLabel), quote("_src_datasource")),
		fmt.Sprintf("%s AS %s", sqlLiteral(ec.SourceTable), quote("_src_table")),
	}
}

// datasourceLabel 返回数据源的可读名称，查询失败时回退为 #<id>。
func datasourceLabel(db *gorm.DB, id uint) string {
	if id == 0 {
		return ""
	}
	var ds model.DataSource
	if err := db.Select("name").First(&ds, id).Error; err == nil && strings.TrimSpace(ds.Name) != "" {
		return ds.Name
	}
	return fmt.Sprintf("#%d", id)
}

// sqlLiteral 将字符串包装为 SQL 单引号字面量，内部单引号翻倍转义。
func sqlLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// defaultStr 返回非空的 s，否则返回 def。
func defaultStr(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}

// ---------------------------------------------------------------------------
// P2-3 1:N 扇出确定性丢弃/警告辅助函数
// ---------------------------------------------------------------------------

// isRelationReferencedByQuery 判断关系名是否被当前查询显式引用（不依赖已构建的 scope.joins）。
//
// 引用形式：
//  1. "relationName.property" 显式限定（group_by / aggregate.property / filter.property / order_by）。
//  2. 裸属性名无法解析到主类（即需要 JOIN 目标类才能解析），此时保留该关系。
//
// mainResolver 为主类列解析器，用于判断裸属性名是否属于主类。
func isRelationReferencedByQuery(req *QueryRequest, relName string, mainResolver *columnResolver) bool {
	// 1) group_by 中显式限定。
	for _, g := range req.GroupBy {
		g = strings.TrimSpace(g)
		if g == "" {
			continue
		}
		if i := strings.Index(g, "."); i >= 0 {
			if strings.EqualFold(strings.TrimSpace(g[:i]), relName) {
				return true
			}
		} else {
			// 裸名不在主类 → 可能需要该 JOIN 才能解析。
			if _, ok := mainResolver.column(g); !ok {
				return true
			}
		}
	}

	// 2) aggregates 中显式限定。
	for _, a := range req.Aggregates {
		p := strings.TrimSpace(a.Property)
		if p == "" {
			continue
		}
		if i := strings.Index(p, "."); i >= 0 {
			if strings.EqualFold(strings.TrimSpace(p[:i]), relName) {
				return true
			}
		} else {
			// 裸名不在主类 → 可能需要该 JOIN 才能解析。
			if _, ok := mainResolver.column(p); !ok {
				return true
			}
		}
	}

	// 3) filters 中显式限定。
	for _, f := range req.Filters {
		p := strings.TrimSpace(f.Property)
		if p == "" {
			continue
		}
		if i := strings.Index(p, "."); i >= 0 {
			if strings.EqualFold(strings.TrimSpace(p[:i]), relName) {
				return true
			}
		} else {
			if _, ok := mainResolver.column(p); !ok {
				return true
			}
		}
	}

	// 4) order_by 中显式限定。
	ob := strings.TrimSpace(req.OrderBy)
	if ob != "" {
		if i := strings.Index(ob, "."); i >= 0 {
			if strings.EqualFold(strings.TrimSpace(ob[:i]), relName) {
				return true
			}
		} else {
			if _, ok := mainResolver.column(ob); !ok {
				return true
			}
		}
	}

	return false
}

// relationCardinality 查询关系的基数（"1:N"/"N:1"/"1:1"/"N:M"）。
func relationCardinality(db *gorm.DB, ontologyID uint, relName string) string {
	if db == nil || strings.TrimSpace(relName) == "" {
		return ""
	}
	var rel model.OntRelation
	if err := db.Select("cardinality").Where("ontology_id = ? AND name = ?", ontologyID, relName).First(&rel).Error; err != nil {
		return ""
	}
	return rel.Cardinality
}

// hasMainClassMeasure 判断聚合查询中是否存在解析到主类(t0)的度量列（SUM/AVG/MIN/MAX）。
// 当 1:N JOIN 存在时，这些主类度量会被子表行数重复放大。
func hasMainClassMeasure(scope *colScope, req *QueryRequest) bool {
	for _, a := range req.Aggregates {
		fn := strings.ToLower(strings.TrimSpace(a.Func))
		if fn == "count" {
			continue // count 已有单独口径处理
		}
		p := strings.TrimSpace(a.Property)
		if p == "" {
			continue
		}
		if scope.resolvesToMain(p) {
			return true
		}
	}
	return false
}
