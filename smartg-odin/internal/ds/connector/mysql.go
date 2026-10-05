package connector

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"smartg-odin/internal/ds/pool"
)

// MySQLConnector MySQL 数据源连接器。
type MySQLConnector struct{}

// Type 返回数据源类型标识。
func (c *MySQLConnector) Type() string { return "mysql" }

// mysqlDSN 构造 MySQL 连接串。
func mysqlDSN(cfg *ConnectionConfig) string {
	host := cfg.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Port
	if port == 0 {
		port = 3306
	}
	// user:pass@tcp(host:port)/dbname?params
	var sb strings.Builder
	sb.WriteString(cfg.Username)
	sb.WriteString(":")
	sb.WriteString(cfg.Password)
	sb.WriteString(fmt.Sprintf("@tcp(%s:%d)/", host, port))
	sb.WriteString(cfg.DatabaseName)

	params := []string{"charset=utf8mb4", "parseTime=true", "loc=Local"}
	if cfg.Params != nil {
		for k, v := range cfg.Params {
			params = append(params, fmt.Sprintf("%s=%s", k, v))
		}
	}
	sb.WriteString("?")
	sb.WriteString(strings.Join(params, "&"))
	return sb.String()
}

// open 打开 MySQL 连接并应用连接池配置。
func (c *MySQLConnector) open(cfg *ConnectionConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", mysqlDSN(cfg))
	if err != nil {
		return nil, err
	}
	maxOpen, maxIdle, maxLife := poolDefaults(cfg.PoolMaxOpen, cfg.PoolMaxIdle, cfg.PoolMaxLifetime)
	pool.Apply(db, pool.Config{MaxOpen: maxOpen, MaxIdle: maxIdle, MaxLifetime: maxLife})
	return db, nil
}

// Test 测试连通性并返回服务器版本。
func (c *MySQLConnector) Test(ctx context.Context, cfg *ConnectionConfig) (*TestResult, error) {
	start := time.Now()
	db, err := c.open(cfg)
	if err != nil {
		return &TestResult{Success: false, LatencyMs: time.Since(start).Milliseconds(), Message: err.Error()}, err
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return &TestResult{Success: false, LatencyMs: time.Since(start).Milliseconds(), Message: err.Error()}, nil
	}
	var version string
	if err := db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version); err != nil {
		return &TestResult{Success: false, LatencyMs: time.Since(start).Milliseconds(), Message: err.Error()}, nil
	}
	return &TestResult{
		Success:   true,
		LatencyMs: time.Since(start).Milliseconds(),
		Message:   "connection successful",
		Version:   "MySQL " + version,
	}, nil
}

// Introspect 内省 MySQL 库表结构。
func (c *MySQLConnector) Introspect(ctx context.Context, cfg *ConnectionConfig) (*SchemaInfo, error) {
	db, err := c.open(cfg)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	schema := cfg.DatabaseName

	info := &SchemaInfo{Tables: make([]TableInfo, 0)}

	// 1. 表
	tableQuery := `
		SELECT TABLE_NAME,
		       CASE TABLE_TYPE WHEN 'BASE TABLE' THEN 'TABLE' ELSE TABLE_TYPE END AS table_type,
		       COALESCE(TABLE_COMMENT, '')
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME`
	tableRows, err := db.QueryContext(ctx, tableQuery, schema)
	if err != nil {
		return nil, err
	}
	type tbl struct{ name, typ, comment string }
	var tables []tbl
	for tableRows.Next() {
		var t tbl
		if err := tableRows.Scan(&t.name, &t.typ, &t.comment); err != nil {
			tableRows.Close()
			return nil, err
		}
		tables = append(tables, t)
	}
	tableRows.Close()
	if err := tableRows.Err(); err != nil {
		return nil, err
	}

	for _, t := range tables {
		ti := TableInfo{
			Name:        t.name,
			Schema:      schema,
			Type:        t.typ,
			Comment:     t.comment,
			Columns:     make([]ColumnInfo, 0),
			ForeignKeys: make([]FKInfo, 0),
			Indexes:     make([]IndexInfo, 0),
		}

		// 2. 字段
		colQuery := `
			SELECT COLUMN_NAME,
			       DATA_TYPE,
			       CASE IS_NULLABLE WHEN 'YES' THEN true ELSE false END,
			       CASE COLUMN_KEY WHEN 'PRI' THEN true ELSE false END,
			       COALESCE(COLUMN_DEFAULT, ''),
			       COALESCE(COLUMN_COMMENT, ''),
			       ORDINAL_POSITION
			FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
			ORDER BY ORDINAL_POSITION`
		colRows, err := db.QueryContext(ctx, colQuery, schema, t.name)
		if err != nil {
			return nil, err
		}
		for colRows.Next() {
			var ci ColumnInfo
			if err := colRows.Scan(&ci.Name, &ci.DataType, &ci.IsNullable, &ci.IsPrimary, &ci.Default, &ci.Comment, &ci.Position); err != nil {
				colRows.Close()
				return nil, err
			}
			ti.Columns = append(ti.Columns, ci)
		}
		colRows.Close()
		if err := colRows.Err(); err != nil {
			return nil, err
		}

		// 3. 外键
		fkQuery := `
			SELECT CONSTRAINT_NAME, COLUMN_NAME, REFERENCED_TABLE_NAME, REFERENCED_COLUMN_NAME
			FROM information_schema.KEY_COLUMN_USAGE
			WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND REFERENCED_TABLE_NAME IS NOT NULL`
		fkRows, err := db.QueryContext(ctx, fkQuery, schema, t.name)
		if err != nil {
			return nil, err
		}
		for fkRows.Next() {
			var fk FKInfo
			if err := fkRows.Scan(&fk.Name, &fk.Column, &fk.RefTable, &fk.RefColumn); err != nil {
				fkRows.Close()
				return nil, err
			}
			ti.ForeignKeys = append(ti.ForeignKeys, fk)
		}
		fkRows.Close()
		if err := fkRows.Err(); err != nil {
			return nil, err
		}

		// 4. 索引（按索引名聚合列）
		idxQuery := `
			SELECT INDEX_NAME, NON_UNIQUE, INDEX_TYPE, SEQ_IN_INDEX, COLUMN_NAME
			FROM information_schema.STATISTICS
			WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
			ORDER BY INDEX_NAME, SEQ_IN_INDEX`
		idxRows, err := db.QueryContext(ctx, idxQuery, schema, t.name)
		if err != nil {
			return nil, err
		}
		idxOrder := make([]string, 0)
		idxMap := make(map[string]*IndexInfo)
		for idxRows.Next() {
			var (
				indexName string
				nonUnique int
				indexType sql.NullString
				seqInIdx  int
				colName   sql.NullString
			)
			if err := idxRows.Scan(&indexName, &nonUnique, &indexType, &seqInIdx, &colName); err != nil {
				idxRows.Close()
				return nil, err
			}
			ii, ok := idxMap[indexName]
			if !ok {
				ii = &IndexInfo{
					Name:      indexName,
					IsUnique:  nonUnique == 0,
					IsPrimary: indexName == "PRIMARY",
				}
				idxMap[indexName] = ii
				idxOrder = append(idxOrder, indexName)
			}
			if colName.Valid {
				if ii.Columns == "" {
					ii.Columns = colName.String
				} else {
					ii.Columns += "," + colName.String
				}
			}
		}
		idxRows.Close()
		if err := idxRows.Err(); err != nil {
			return nil, err
		}
		for _, name := range idxOrder {
			ti.Indexes = append(ti.Indexes, *idxMap[name])
		}

		info.Tables = append(info.Tables, ti)
	}

	return info, nil
}

// Query 执行 SQL 查询并返回结果集。
func (c *MySQLConnector) Query(ctx context.Context, cfg *ConnectionConfig, query string, args []interface{}) (*QueryResult, error) {
	db, err := c.open(cfg)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	start := time.Now()
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return scanRows(rows, start)
}
