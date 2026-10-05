// Package connector 定义数据源连接器的统一接口与相关数据结构，
// 屏蔽不同数据库（SQLite / PostgreSQL / MySQL）的差异。
package connector

import "context"

// ConnectionConfig 建立数据源连接所需的配置。
type ConnectionConfig struct {
	Type         string
	Host         string
	Port         int
	DatabaseName string
	Username     string
	Password     string
	SchemaName   string
	SSLMode      string
	Params       map[string]string
	// SQLite 专用：数据库文件路径
	FilePath string
	// 连接池配置（可选，0 表示使用驱动默认值）
	PoolMaxOpen     int
	PoolMaxIdle     int
	PoolMaxLifetime int // seconds
}

// TestResult 连通性测试结果。
type TestResult struct {
	Success   bool   `json:"success"`
	LatencyMs int64  `json:"latency_ms"`
	Message   string `json:"message"`
	Version   string `json:"version,omitempty"`
}

// SchemaInfo 内省得到的库表结构信息。
type SchemaInfo struct {
	Tables []TableInfo `json:"tables"`
}

// TableInfo 单张表的结构信息。
type TableInfo struct {
	Name        string       `json:"name"`
	Schema      string       `json:"schema"`
	Type        string       `json:"type"`
	Comment     string       `json:"comment"`
	Columns     []ColumnInfo `json:"columns"`
	ForeignKeys []FKInfo     `json:"foreign_keys"`
	Indexes     []IndexInfo  `json:"indexes"`
}

// ColumnInfo 字段信息。
type ColumnInfo struct {
	Name       string `json:"name"`
	DataType   string `json:"data_type"`
	IsNullable bool   `json:"is_nullable"`
	IsPrimary  bool   `json:"is_primary_key"`
	Default    string `json:"default_value"`
	Comment    string `json:"comment"`
	Position   int    `json:"ordinal_position"`
}

// FKInfo 外键信息。
type FKInfo struct {
	Name      string `json:"name"`
	Column    string `json:"column"`
	RefTable  string `json:"ref_table"`
	RefColumn string `json:"ref_column"`
}

// IndexInfo 索引信息。
type IndexInfo struct {
	Name      string `json:"name"`
	Columns   string `json:"columns"`
	IsUnique  bool   `json:"is_unique"`
	IsPrimary bool   `json:"is_primary"`
}

// QueryResult SQL 查询结果。
type QueryResult struct {
	Columns  []string        `json:"columns"`
	Rows     [][]interface{} `json:"rows"`
	RowCount int             `json:"row_count"`
	Duration int64           `json:"duration_ms"`
}

// Connector 数据源连接器统一接口。
type Connector interface {
	// Test 测试连通性。
	Test(ctx context.Context, config *ConnectionConfig) (*TestResult, error)
	// Introspect 内省库表结构。
	Introspect(ctx context.Context, config *ConnectionConfig) (*SchemaInfo, error)
	// Query 执行只读 SQL 查询并返回结果集。
	Query(ctx context.Context, config *ConnectionConfig, sql string, args []interface{}) (*QueryResult, error)
	// Type 返回连接器对应的数据源类型标识。
	Type() string
}
