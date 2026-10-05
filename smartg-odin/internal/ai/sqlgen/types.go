package sqlgen

// SQLGenContext LLM 生成 SQL 所需的完整上下文。
//
// 由 BuildContext 从元数据库装配：包含按数据源分组的物理 Schema、可用的
// JOIN 路径、业务同义词表以及虚拟类/派生规则。它会被序列化进 system prompt，
// 是「LLM 只能引用真实存在的表列」这一铁律的信息来源。
type SQLGenContext struct {
	Datasources []DatasourceSchema  `json:"datasources"` // 按数据源分组的物理 Schema
	Relations   []RelationInfo      `json:"relations"`   // JOIN 路径信息
	Synonyms    map[string][]string `json:"synonyms"`    // 业务同义词表
	Rules       []RuleInfo          `json:"rules"`       // 虚拟类/派生规则
}

// DatasourceSchema 单个数据源的物理 Schema。
type DatasourceSchema struct {
	ID     uint          `json:"id"`
	Name   string        `json:"name"`
	Type   string        `json:"type"` // mysql/postgres/sqlite
	Tables []TableSchema `json:"tables"`
}

// TableSchema 单张物理表的结构。
type TableSchema struct {
	Name    string         `json:"name"`
	Comment string         `json:"comment"`
	Columns []ColumnSchema `json:"columns"`
}

// ColumnSchema 单个物理列的结构，并携带从本体映射关联到的中文语义。
type ColumnSchema struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	IsPrimary   bool   `json:"is_primary"`
	Comment     string `json:"comment"`
	OntLabel    string `json:"ont_label"`    // 本体中文 label（从映射关联）
	OntProperty string `json:"ont_property"` // 对应的本体属性名
}

// RelationInfo 本体关系对应的物理 JOIN 路径。
type RelationInfo struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	FromClass   string `json:"from_class"`
	ToClass     string `json:"to_class"`
	Cardinality string `json:"cardinality"` // 1:1, 1:N, N:M
	JoinSQL     string `json:"join_sql"`    // 可直接使用的 ON 条件，如 "orders.product_id = products.id"
	FromTable   string `json:"from_table"`
	ToTable     string `json:"to_table"`
}

// RuleInfo 虚拟类/派生规则信息（如「大额订单 = orders WHERE amount > 50000」）。
type RuleInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Condition   string `json:"condition"` // 如 "amount > 50000"
	BaseClass   string `json:"base_class"`
	BaseTable   string `json:"base_table"`
}

// SQLGenResult LLM 生成的结构化结果。
type SQLGenResult struct {
	Steps         []SQLStep `json:"steps"`
	MergeStrategy string    `json:"merge_strategy"` // single|join|side_by_side|synthesize
	MergeKeys     []string  `json:"merge_keys"`
	Conclusion    string    `json:"conclusion"`
}

// SQLStep 单条 SQL 及其归属数据源。
type SQLStep struct {
	Purpose      string `json:"purpose"`
	SQL          string `json:"sql"`
	DatasourceID uint   `json:"datasource_id"`
}

// StepResult 单条 SQL 的执行结果。
type StepResult struct {
	Purpose  string          `json:"purpose"`
	SQL      string          `json:"sql"`
	Columns  []ColumnInfo    `json:"columns"`
	Rows     [][]interface{} `json:"rows"`
	RowCount int             `json:"row_count"`
	Duration int64           `json:"duration"` // milliseconds
	Error    string          `json:"error,omitempty"`
}

// ColumnInfo 结果列的名称与类型。
type ColumnInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// MergedResult 合并后的最终结果。
type MergedResult struct {
	Columns  []ColumnInfo    `json:"columns"`
	Rows     [][]interface{} `json:"rows"`
	RowCount int             `json:"row_count"`
	Steps    []StepResult    `json:"steps"` // 各步原始结果（side_by_side/synthesize 模式用）
	Strategy string          `json:"strategy"`
}

// ChartPayload 图表建议载荷，供前端可视化渲染。
type ChartPayload struct {
	ChartType  string          `json:"chart_type"` // bar|line|pie|metric
	Dimensions []ChartField    `json:"dimensions"`
	Metrics    []ChartField    `json:"metrics"`
	Rows       [][]interface{} `json:"rows"`
	Columns    []string        `json:"columns"`
}

// ChartField 图表中的一个维度/度量字段。
type ChartField struct {
	Field string `json:"field"`
	Label string `json:"label"`
	Agg   string `json:"agg,omitempty"`
}

// ValidationError SQL 校验错误，定位到具体的 step。
type ValidationError struct {
	StepIndex int    `json:"step_index"`
	SQL       string `json:"sql"`
	Reason    string `json:"reason"`
}
