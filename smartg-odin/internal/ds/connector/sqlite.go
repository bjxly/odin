package connector

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"smartg-odin/internal/ds/pool"
)

// SQLiteConnector SQLite 数据源连接器。
type SQLiteConnector struct{}

// Type 返回数据源类型标识。
func (c *SQLiteConnector) Type() string { return "sqlite" }

// sqlitePath 解析 SQLite 文件路径。
func sqlitePath(cfg *ConnectionConfig) string {
	if cfg.FilePath != "" {
		return cfg.FilePath
	}
	return cfg.DatabaseName
}

// open 打开 SQLite 连接并应用连接池配置。
func (c *SQLiteConnector) open(cfg *ConnectionConfig) (*sql.DB, error) {
	path := sqlitePath(cfg)
	if path == "" {
		return nil, fmt.Errorf("sqlite file path is empty")
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	maxOpen, maxIdle, maxLife := poolDefaults(cfg.PoolMaxOpen, cfg.PoolMaxIdle, cfg.PoolMaxLifetime)
	pool.Apply(db, pool.Config{MaxOpen: maxOpen, MaxIdle: maxIdle, MaxLifetime: maxLife})
	return db, nil
}

// Test 测试连通性并返回 SQLite 版本。
func (c *SQLiteConnector) Test(ctx context.Context, cfg *ConnectionConfig) (*TestResult, error) {
	start := time.Now()
	db, err := c.open(cfg)
	if err != nil {
		return &TestResult{Success: false, LatencyMs: time.Since(start).Milliseconds(), Message: err.Error()}, err
	}
	defer db.Close()

	var version string
	if err := db.QueryRowContext(ctx, "SELECT sqlite_version()").Scan(&version); err != nil {
		return &TestResult{Success: false, LatencyMs: time.Since(start).Milliseconds(), Message: err.Error()}, nil
	}
	return &TestResult{
		Success:   true,
		LatencyMs: time.Since(start).Milliseconds(),
		Message:   "connection successful",
		Version:   "SQLite " + version,
	}, nil
}

// Introspect 内省 SQLite 库表结构。
func (c *SQLiteConnector) Introspect(ctx context.Context, cfg *ConnectionConfig) (*SchemaInfo, error) {
	db, err := c.open(cfg)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	info := &SchemaInfo{Tables: make([]TableInfo, 0)}

	// 1. 表与视图
	tableRows, err := db.QueryContext(ctx,
		"SELECT name, type FROM sqlite_master WHERE type IN ('table','view') AND name NOT LIKE 'sqlite_%' ORDER BY name")
	if err != nil {
		return nil, err
	}
	type tbl struct {
		name    string
		typ     string
		comment string
	}
	var tables []tbl
	for tableRows.Next() {
		var name, typ string
		if err := tableRows.Scan(&name, &typ); err != nil {
			tableRows.Close()
			return nil, err
		}
		tables = append(tables, tbl{name: name, typ: strings.ToUpper(typ)})
	}
	tableRows.Close()
	if err := tableRows.Err(); err != nil {
		return nil, err
	}

	for _, t := range tables {
		ti := TableInfo{
			Name:        t.name,
			Type:        t.typ,
			Columns:     make([]ColumnInfo, 0),
			ForeignKeys: make([]FKInfo, 0),
			Indexes:     make([]IndexInfo, 0),
		}

		// 2. 字段：PRAGMA table_info
		colRows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", quoteIdent(t.name)))
		if err != nil {
			return nil, err
		}
		for colRows.Next() {
			var (
				cid      int
				name     string
				dataType sql.NullString
				notnull  int
				dflt     sql.NullString
				pk       int
			)
			if err := colRows.Scan(&cid, &name, &dataType, &notnull, &dflt, &pk); err != nil {
				colRows.Close()
				return nil, err
			}
			ti.Columns = append(ti.Columns, ColumnInfo{
				Name:       name,
				DataType:   dataType.String,
				IsNullable: notnull == 0,
				IsPrimary:  pk > 0,
				Default:    dflt.String,
				Position:   cid,
			})
		}
		colRows.Close()
		if err := colRows.Err(); err != nil {
			return nil, err
		}

		// 3. 外键：PRAGMA foreign_key_list
		fkRows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA foreign_key_list(%s)", quoteIdent(t.name)))
		if err != nil {
			return nil, err
		}
		for fkRows.Next() {
			var (
				id       int
				seq      int
				refTable string
				from     string
				to       sql.NullString
			)
			// 列: id, seq, table, from, to, on_update, on_delete, match
			var onUpdate, onDelete, match sql.NullString
			if err := fkRows.Scan(&id, &seq, &refTable, &from, &to, &onUpdate, &onDelete, &match); err != nil {
				fkRows.Close()
				return nil, err
			}
			ti.ForeignKeys = append(ti.ForeignKeys, FKInfo{
				Name:      fmt.Sprintf("fk_%s_%d", t.name, id),
				Column:    from,
				RefTable:  refTable,
				RefColumn: to.String,
			})
		}
		fkRows.Close()
		if err := fkRows.Err(); err != nil {
			return nil, err
		}

		// 4. 索引：PRAGMA index_list + PRAGMA index_info
		idxRows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA index_list(%s)", quoteIdent(t.name)))
		if err != nil {
			return nil, err
		}
		type idxMeta struct {
			name    string
			unique  bool
			primary bool
		}
		var idxMetas []idxMeta
		for idxRows.Next() {
			var (
				seq     int
				name    string
				unique  int
				origin  sql.NullString
				partial int
			)
			if err := idxRows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
				idxRows.Close()
				return nil, err
			}
			idxMetas = append(idxMetas, idxMeta{
				name:    name,
				unique:  unique != 0,
				primary: origin.String == "pk",
			})
		}
		idxRows.Close()
		if err := idxRows.Err(); err != nil {
			return nil, err
		}

		for _, im := range idxMetas {
			colRows2, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA index_info(%s)", quoteIdent(im.name)))
			if err != nil {
				return nil, err
			}
			var colNames []string
			for colRows2.Next() {
				var (
					seqno int
					cid   int
					name  sql.NullString
				)
				if err := colRows2.Scan(&seqno, &cid, &name); err != nil {
					colRows2.Close()
					return nil, err
				}
				if name.Valid {
					colNames = append(colNames, name.String)
				}
			}
			colRows2.Close()
			if err := colRows2.Err(); err != nil {
				return nil, err
			}
			ti.Indexes = append(ti.Indexes, IndexInfo{
				Name:      im.name,
				Columns:   strings.Join(colNames, ","),
				IsUnique:  im.unique,
				IsPrimary: im.primary,
			})
		}

		info.Tables = append(info.Tables, ti)
	}

	return info, nil
}

// Query 执行 SQL 查询并返回结果集。
func (c *SQLiteConnector) Query(ctx context.Context, cfg *ConnectionConfig, query string, args []interface{}) (*QueryResult, error) {
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

// quoteIdent 为 SQLite 标识符加上双引号，防止关键字冲突。
func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
