package connector

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"smartg-odin/internal/ds/pool"
)

// PostgresConnector PostgreSQL 数据源连接器（基于 pgx/v5 stdlib）。
type PostgresConnector struct{}

// Type 返回数据源类型标识。
func (c *PostgresConnector) Type() string { return "postgres" }

// pgDSN 构造 PostgreSQL 连接串。
func pgDSN(cfg *ConnectionConfig) string {
	// 支持直接传入完整 DSN（以 postgres:// 或 postgresql:// 开头）
	if strings.HasPrefix(cfg.DatabaseName, "postgres://") || strings.HasPrefix(cfg.DatabaseName, "postgresql://") {
		return cfg.DatabaseName
	}
	host := cfg.Host
	if host == "" {
		host = "localhost"
	}
	port := cfg.Port
	if port == 0 {
		port = 5432
	}
	ssl := cfg.SSLMode
	if ssl == "" {
		ssl = "disable"
	}
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.Username, cfg.Password),
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   cfg.DatabaseName,
	}
	q := u.Query()
	q.Set("sslmode", ssl)
	if cfg.SchemaName != "" {
		q.Set("search_path", cfg.SchemaName)
	}
	for k, v := range cfg.Params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// open 打开 PostgreSQL 连接并应用连接池配置。
func (c *PostgresConnector) open(cfg *ConnectionConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", pgDSN(cfg))
	if err != nil {
		return nil, err
	}
	maxOpen, maxIdle, maxLife := poolDefaults(cfg.PoolMaxOpen, cfg.PoolMaxIdle, cfg.PoolMaxLifetime)
	pool.Apply(db, pool.Config{MaxOpen: maxOpen, MaxIdle: maxIdle, MaxLifetime: maxLife})
	return db, nil
}

// Test 测试连通性并返回服务器版本。
func (c *PostgresConnector) Test(ctx context.Context, cfg *ConnectionConfig) (*TestResult, error) {
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
	if err := db.QueryRowContext(ctx, "SELECT version()").Scan(&version); err != nil {
		return &TestResult{Success: false, LatencyMs: time.Since(start).Milliseconds(), Message: err.Error()}, nil
	}
	return &TestResult{
		Success:   true,
		LatencyMs: time.Since(start).Milliseconds(),
		Message:   "connection successful",
		Version:   version,
	}, nil
}

// Introspect 内省 PostgreSQL 库表结构。
func (c *PostgresConnector) Introspect(ctx context.Context, cfg *ConnectionConfig) (*SchemaInfo, error) {
	db, err := c.open(cfg)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	schema := cfg.SchemaName
	if schema == "" {
		schema = "public"
	}

	info := &SchemaInfo{Tables: make([]TableInfo, 0)}

	// 1. 表
	tableQuery := `
		SELECT t.table_name,
		       CASE t.table_type WHEN 'BASE TABLE' THEN 'TABLE' ELSE 'VIEW' END AS table_type,
		       COALESCE(obj_description((quote_ident(t.table_schema)||'.'||quote_ident(t.table_name))::regclass), '') AS comment
		FROM information_schema.tables t
		WHERE t.table_schema = $1
		ORDER BY t.table_name`
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
			SELECT c.column_name,
			       c.data_type,
			       CASE c.is_nullable WHEN 'YES' THEN true ELSE false END,
			       COALESCE(c.column_default, ''),
			       COALESCE(col_description((quote_ident(c.table_schema)||'.'||quote_ident(c.table_name))::regclass, c.ordinal_position), ''),
			       c.ordinal_position,
			       EXISTS (
			         SELECT 1 FROM information_schema.table_constraints tc
			         JOIN information_schema.key_column_usage kcu
			           ON tc.constraint_name = kcu.constraint_name AND tc.table_schema = kcu.table_schema
			         WHERE tc.constraint_type = 'PRIMARY KEY'
			           AND tc.table_schema = c.table_schema
			           AND tc.table_name = c.table_name
			           AND kcu.column_name = c.column_name
			       ) AS is_primary
			FROM information_schema.columns c
			WHERE c.table_schema = $1 AND c.table_name = $2
			ORDER BY c.ordinal_position`
		colRows, err := db.QueryContext(ctx, colQuery, schema, t.name)
		if err != nil {
			return nil, err
		}
		for colRows.Next() {
			var ci ColumnInfo
			if err := colRows.Scan(&ci.Name, &ci.DataType, &ci.IsNullable, &ci.Default, &ci.Comment, &ci.Position, &ci.IsPrimary); err != nil {
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
			SELECT tc.constraint_name,
			       kcu.column_name,
			       ccu.table_name AS ref_table,
			       ccu.column_name AS ref_column
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
			  ON tc.constraint_name = kcu.constraint_name AND tc.table_schema = kcu.table_schema
			JOIN information_schema.constraint_column_usage ccu
			  ON ccu.constraint_name = tc.constraint_name AND ccu.table_schema = tc.table_schema
			WHERE tc.constraint_type = 'FOREIGN KEY'
			  AND tc.table_schema = $1 AND tc.table_name = $2`
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

		// 4. 索引
		idxQuery := `
			SELECT i.relname AS index_name,
			       ix.indisunique AS is_unique,
			       ix.indisprimary AS is_primary,
			       array_to_string(
			         ARRAY(SELECT pg_get_indexdef(ix.indexrelid, k + 1, false)
			               FROM generate_subscripts(ix.indkey, 1) AS k), ',') AS columns
			FROM pg_index ix
			JOIN pg_class t ON t.oid = ix.indrelid
			JOIN pg_class i ON i.oid = ix.indexrelid
			JOIN pg_namespace n ON n.oid = t.relnamespace
			WHERE n.nspname = $1 AND t.relname = $2
			ORDER BY i.relname`
		idxRows, err := db.QueryContext(ctx, idxQuery, schema, t.name)
		if err != nil {
			return nil, err
		}
		for idxRows.Next() {
			var ii IndexInfo
			if err := idxRows.Scan(&ii.Name, &ii.IsUnique, &ii.IsPrimary, &ii.Columns); err != nil {
				idxRows.Close()
				return nil, err
			}
			ti.Indexes = append(ti.Indexes, ii)
		}
		idxRows.Close()
		if err := idxRows.Err(); err != nil {
			return nil, err
		}

		info.Tables = append(info.Tables, ti)
	}

	return info, nil
}

// Query 执行 SQL 查询并返回结果集。
func (c *PostgresConnector) Query(ctx context.Context, cfg *ConnectionConfig, query string, args []interface{}) (*QueryResult, error) {
	db, err := c.open(cfg)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	start := time.Now()
	// translator 统一生成 "?" 占位符，PostgreSQL 需重绑为 $N（见 rebindPlaceholders）。
	rows, err := db.QueryContext(ctx, rebindPlaceholders(query), args...)
	if err != nil {
		return nil, err
	}
	return scanRows(rows, start)
}
