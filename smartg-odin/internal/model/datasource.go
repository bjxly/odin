package model

import (
	"time"

	"gorm.io/gorm"
)

// DataSource 数据源连接配置。
type DataSource struct {
	ID              uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Code            string         `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Name            string         `gorm:"size:100;not null" json:"name"`
	Type            string         `gorm:"size:30;not null" json:"type"` // postgres, mysql, sqlite
	Host            string         `gorm:"size:200" json:"host"`
	Port            int            `json:"port"`
	DatabaseName    string         `gorm:"size:100" json:"database_name"`
	Username        string         `gorm:"size:100" json:"username"`
	Password        string         `gorm:"size:500" json:"-"` // 不返回给前端
	SchemaName      string         `gorm:"size:100" json:"schema_name"`
	ParamsJSON      string         `gorm:"type:text" json:"params_json"`
	PoolMaxOpen     int            `gorm:"default:10" json:"pool_max_open"`
	PoolMaxIdle     int            `gorm:"default:5" json:"pool_max_idle"`
	PoolMaxLifetime int            `gorm:"default:3600" json:"pool_max_lifetime"`
	SSLMode         string         `gorm:"size:20;default:disable" json:"ssl_mode"`
	SSHEnabled      bool           `gorm:"default:false" json:"ssh_enabled"`
	SSHHost         string         `gorm:"size:200" json:"ssh_host"`
	SSHPort         int            `json:"ssh_port"`
	SSHUser         string         `gorm:"size:100" json:"ssh_user"`
	SSHKey          string         `gorm:"type:text" json:"-"`
	Status          string         `gorm:"size:20;default:inactive" json:"status"` // active, inactive, error
	Tags            string         `gorm:"size:500" json:"tags"`
	Description     string         `gorm:"type:text" json:"description"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 显式指定表名。
func (DataSource) TableName() string { return "datasource" }

// DataSourceTest 数据源连通性测试记录。
type DataSourceTest struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	DataSourceID uint      `gorm:"column:datasource_id;index;not null" json:"datasource_id"`
	Status       string    `gorm:"size:20" json:"status"` // success, failed
	LatencyMs    int64     `json:"latency_ms"`
	Message      string    `gorm:"type:text" json:"message"`
	TestedAt     time.Time `json:"tested_at"`
}

func (DataSourceTest) TableName() string { return "datasource_test" }

// SchemaTable 数据源中同步过来的表元信息。
type SchemaTable struct {
	ID           uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	DataSourceID uint   `gorm:"column:datasource_id;index;not null" json:"datasource_id"`
	SchemaName   string `gorm:"size:100" json:"schema_name"`
	TableName    string `gorm:"size:100;not null" json:"table_name"`
	TableType    string `gorm:"size:20;default:TABLE" json:"table_type"`
	Comment      string `gorm:"type:text" json:"comment"`
	RowCount     int64  `json:"row_count"`
}

// NOTE: SchemaTable 含 TableName 字段，无法再定义 TableName() 方法，
// 因此使用 GORM 默认表名 schema_tables。

// SchemaColumn 表字段元信息。
type SchemaColumn struct {
	ID           uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	TableID      uint   `gorm:"index;not null" json:"table_id"`
	DataSourceID uint   `gorm:"column:datasource_id;index;not null" json:"datasource_id"`
	ColumnName   string `gorm:"size:100;not null" json:"column_name"`
	DataType     string `gorm:"size:50;not null" json:"data_type"`
	IsNullable   bool   `json:"is_nullable"`
	IsPrimaryKey bool   `json:"is_primary_key"`
	DefaultValue string `gorm:"size:200" json:"default_value"`
	Comment      string `gorm:"type:text" json:"comment"`
	OrdinalPos   int    `json:"ordinal_position"`
}

func (SchemaColumn) TableName() string { return "schema_column" }

// SchemaForeignKey 表外键约束信息。
type SchemaForeignKey struct {
	ID             uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	TableID        uint   `gorm:"index;not null" json:"table_id"`
	DataSourceID   uint   `gorm:"column:datasource_id;index;not null" json:"datasource_id"`
	ConstraintName string `gorm:"size:100" json:"constraint_name"`
	ColumnName     string `gorm:"size:100" json:"column_name"`
	RefTable       string `gorm:"size:100" json:"ref_table"`
	RefColumn      string `gorm:"size:100" json:"ref_column"`
}

func (SchemaForeignKey) TableName() string { return "schema_foreign_key" }

// SchemaIndex 表索引信息。
type SchemaIndex struct {
	ID           uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	TableID      uint   `gorm:"index;not null" json:"table_id"`
	DataSourceID uint   `gorm:"column:datasource_id;index;not null" json:"datasource_id"`
	IndexName    string `gorm:"size:100" json:"index_name"`
	Columns      string `gorm:"size:500" json:"columns"` // comma-separated
	IsUnique     bool   `json:"is_unique"`
	IsPrimary    bool   `json:"is_primary"`
}

func (SchemaIndex) TableName() string { return "schema_index" }
