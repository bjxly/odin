package sqlgen

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"smartg-odin/internal/ds/connector"
	"smartg-odin/internal/model"
)

// defaultStepTimeout 未在配置中指定时使用的单条 SQL 执行超时。
const defaultStepTimeout = 30 * time.Second

// Execute 按顺序在各目标数据源上执行 SQL 步骤，并收集每步结果。
//
// 步骤顺序执行（后一步可能依赖前一步的语义）；单步失败不会终止整体流程，
// 而是把错误记录进对应 StepResult.Error 后继续下一步。返回的 error 仅用于
// 表达「整体无法开展」的致命问题（如 db 为空），单步错误不通过 error 返回。
//
// 每步依据 step.DatasourceID 加载 model.DataSource，复用 ds/connector 建立连接并执行；
// 结果行数按 cfg.MaxRows 截断。
func Execute(ctx context.Context, db *gorm.DB, steps []SQLStep, cfg *SQLGenConfig) ([]StepResult, error) {
	if db == nil {
		return nil, fmt.Errorf("db 为空，无法执行 SQL")
	}
	if cfg == nil {
		cfg = DefaultConfig()
	}
	cfg = cfg.Normalize()

	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = defaultStepTimeout
	}

	dsCache := map[uint]*model.DataSource{}
	connCache := map[string]connector.Connector{}
	cfgCache := map[uint]*connector.ConnectionConfig{}

	results := make([]StepResult, 0, len(steps))
	for _, step := range steps {
		sr := StepResult{
			Purpose: step.Purpose,
			SQL:     step.SQL,
			Columns: []ColumnInfo{},
			Rows:    [][]interface{}{},
		}
		start := time.Now()

		conn, ccfg, err := resolveConnection(db, step.DatasourceID, dsCache, connCache, cfgCache)
		if err != nil {
			sr.Error = err.Error()
			sr.Duration = time.Since(start).Milliseconds()
			results = append(results, sr)
			continue
		}

		stepCtx, cancel := context.WithTimeout(ctx, timeout)
		qr, qerr := conn.Query(stepCtx, ccfg, step.SQL, nil)
		cancel()

		sr.Duration = time.Since(start).Milliseconds()
		if qerr != nil {
			sr.Error = qerr.Error()
			results = append(results, sr)
			continue
		}
		if qr != nil {
			for _, name := range qr.Columns {
				sr.Columns = append(sr.Columns, ColumnInfo{Name: name})
			}
			rows := qr.Rows
			if cfg.MaxRows > 0 && len(rows) > cfg.MaxRows {
				rows = rows[:cfg.MaxRows]
			}
			if rows != nil {
				sr.Rows = rows
			}
			sr.RowCount = len(sr.Rows)
		}
		results = append(results, sr)
	}
	return results, nil
}

// resolveConnection 加载（并缓存）指定数据源的连接器与连接配置。
func resolveConnection(db *gorm.DB, datasourceID uint,
	dsCache map[uint]*model.DataSource,
	connCache map[string]connector.Connector,
	cfgCache map[uint]*connector.ConnectionConfig) (connector.Connector, *connector.ConnectionConfig, error) {
	if datasourceID == 0 {
		return nil, nil, fmt.Errorf("step 未指定有效的 datasource_id")
	}
	ds, ok := dsCache[datasourceID]
	if !ok {
		var loaded model.DataSource
		if err := db.First(&loaded, datasourceID).Error; err != nil {
			return nil, nil, fmt.Errorf("数据源 %d 不存在: %w", datasourceID, err)
		}
		ds = &loaded
		dsCache[datasourceID] = ds
	}

	conn, ok := connCache[ds.Type]
	if !ok {
		c, err := connector.Get(ds.Type)
		if err != nil {
			return nil, nil, fmt.Errorf("数据源 %d 类型 %q 无可用连接器: %w", datasourceID, ds.Type, err)
		}
		conn = c
		connCache[ds.Type] = c
	}

	ccfg, ok := cfgCache[datasourceID]
	if !ok {
		ccfg = buildConnectionConfig(ds)
		cfgCache[datasourceID] = ccfg
	}
	return conn, ccfg, nil
}

// buildConnectionConfig 依据 DataSource 记录构造连接器所需的 ConnectionConfig
//（与 internal/ds/query 的同名逻辑保持一致）。
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
	if ds.Type == "sqlite" {
		cfg.FilePath = ds.DatabaseName
	}
	if strings.TrimSpace(ds.ParamsJSON) != "" {
		var params map[string]string
		if err := json.Unmarshal([]byte(ds.ParamsJSON), &params); err == nil {
			cfg.Params = params
		}
	}
	return cfg
}
