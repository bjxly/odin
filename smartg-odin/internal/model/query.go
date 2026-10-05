package model

import "time"

// QueryTemplate 保存的查询模板。
type QueryTemplate struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	OntologyID  uint      `gorm:"index" json:"ontology_id"`
	QueryJSON   string    `gorm:"type:text;not null" json:"query_json"`
	ParamsJSON  string    `gorm:"type:text" json:"params_json"`
	CreatedBy   string    `gorm:"size:100" json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (QueryTemplate) TableName() string { return "query_template" }

// QueryHistory 查询执行历史。
type QueryHistory struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	QueryText       string    `gorm:"type:text" json:"query_text"`
	QueryType       string    `gorm:"size:20" json:"query_type"` // ontology, sql, nl
	OntologyID      *uint     `gorm:"index" json:"ontology_id"`
	ResultCount     int       `json:"result_count"`
	ExecutionTimeMs int64     `json:"execution_time_ms"`
	Status          string    `gorm:"size:20" json:"status"` // success, failed, timeout
	ErrorMessage    string    `gorm:"type:text" json:"error_message"`
	CreatedAt       time.Time `gorm:"index" json:"created_at"`
}

func (QueryHistory) TableName() string { return "query_history" }

// QueryTrace 查询全链路跟踪，记录从自然语言到 SQL 执行结果的每一步，用于调试与可解释性。
type QueryTrace struct {
	ID                uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TraceID           string    `gorm:"size:64;index" json:"trace_id"`
	NLText            string    `gorm:"type:text" json:"nl_text"`
	ParsePath         string    `gorm:"size:20" json:"parse_path"` // rule, llm, cache
	IntentJSON        string    `gorm:"type:text" json:"intent_json"`
	OntologyQueryJSON string    `gorm:"type:text" json:"ontology_query_json"`
	LLMJSON           string    `gorm:"type:text" json:"llm_json"`
	ReasonTraceJSON   string    `gorm:"type:text" json:"reason_trace_json"`
	TranslatedSQL     string    `gorm:"type:text" json:"translated_sql"`
	ParamsJSON        string    `gorm:"type:text" json:"params_json"`
	DataSourceID      uint      `gorm:"column:datasource_id" json:"datasource_id"`
	SourceTable       string    `gorm:"size:200" json:"source_table"`
	JoinsJSON         string    `gorm:"type:text" json:"joins_json"`
	Explanation       string    `gorm:"type:text" json:"explanation"`
	ResultCount       int       `json:"result_count"`
	ExecutionTimeMs   int64     `json:"execution_time_ms"`
	Status            string    `gorm:"size:20" json:"status"` // success, failed, dry_run, parse_only
	ErrorMessage      string    `gorm:"type:text" json:"error_message"`
	NodesJSON         string    `gorm:"type:text" json:"nodes_json"`
	CreatedAt         time.Time `gorm:"index" json:"created_at"`
}

func (QueryTrace) TableName() string { return "query_trace" }
