package query

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"smartg-odin/internal/ds/connector"
	"smartg-odin/internal/model"
)

// defaultQueryTimeout 未显式设置 ctx deadline 时使用的默认查询超时。
const defaultQueryTimeout = 30 * time.Second

// Execute 在目标数据源上执行翻译后的查询。
func Execute(ctx context.Context, db *gorm.DB, tq *TranslatedQuery) (*connector.QueryResult, error) {
	if tq == nil {
		return nil, fmt.Errorf("nil translated query")
	}

	// 1. 加载数据源配置。
	var ds model.DataSource
	if err := db.First(&ds, tq.DataSourceID).Error; err != nil {
		return nil, fmt.Errorf("datasource %d not found: %w", tq.DataSourceID, err)
	}

	// 2. 获取对应类型的连接器。
	conn, err := connector.Get(ds.Type)
	if err != nil {
		return nil, err
	}

	// 3. 构造连接配置。
	cfg := buildConnectionConfig(&ds)

	// 4. 设置超时（若调用方未设置 deadline）。
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultQueryTimeout)
		defer cancel()
	}

	// 5. 执行查询。
	return conn.Query(ctx, cfg, tq.SQL, tq.Params)
}

// ExecuteAndRecord 执行查询并把结果写入 query_history。
func ExecuteAndRecord(ctx context.Context, db *gorm.DB, req *QueryRequest, tq *TranslatedQuery) (*connector.QueryResult, error) {
	start := time.Now()
	result, err := Execute(ctx, db, tq)
	duration := time.Since(start).Milliseconds()

	history := model.QueryHistory{
		QueryText:       tq.SQL,
		QueryType:       "ontology",
		ExecutionTimeMs: duration,
	}
	if req != nil && req.OntologyID != 0 {
		oid := req.OntologyID
		history.OntologyID = &oid
	}
	if err != nil {
		history.Status = "failed"
		history.ErrorMessage = err.Error()
	} else {
		history.Status = "success"
		history.ResultCount = result.RowCount
	}
	if cerr := db.Create(&history).Error; cerr != nil {
		// 记录历史失败不应影响查询结果返回。
		_ = cerr
	}

	return result, err
}

// buildConnectionConfig 依据 DataSource 记录构造连接器所需的 ConnectionConfig。
func buildConnectionConfig(ds *model.DataSource) *connector.ConnectionConfig {
	cfg := &connector.ConnectionConfig{
		Type:            ds.Type,
		Host:            ds.Host,
		Port:            ds.Port,
		DatabaseName:    ds.DatabaseName,
		Username:        ds.Username,
		Password:        ds.Password,
		SchemaName:      ds.SchemaName,
		SSLMode:         ds.SSLMode,
		PoolMaxOpen:     ds.PoolMaxOpen,
		PoolMaxIdle:     ds.PoolMaxIdle,
		PoolMaxLifetime: ds.PoolMaxLifetime,
	}
	// SQLite 使用文件路径。
	if ds.Type == "sqlite" {
		cfg.FilePath = ds.DatabaseName
	}
	// 解析扩展参数。
	if ds.ParamsJSON != "" {
		var params map[string]string
		if err := json.Unmarshal([]byte(ds.ParamsJSON), &params); err == nil {
			cfg.Params = params
		}
	}
	return cfg
}
