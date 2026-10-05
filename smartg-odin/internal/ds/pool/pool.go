// Package pool 提供数据源连接池的配置与应用能力。
// 实际的连接池由 database/sql 内置管理，这里仅负责将配置应用到底层 *sql.DB。
package pool

import (
	"database/sql"
	"time"
)

// Config 连接池配置。
type Config struct {
	MaxOpen     int // 最大打开连接数，<=0 表示不限制
	MaxIdle     int // 最大空闲连接数，<=0 使用默认值
	MaxLifetime int // 连接最大存活时间（秒），<=0 表示不限制
}

// Default 返回默认连接池配置。
func Default() Config {
	return Config{
		MaxOpen:     10,
		MaxIdle:     5,
		MaxLifetime: 3600,
	}
}

// Apply 将连接池配置应用到 *sql.DB。
// 对非法值做兜底处理，避免因配置错误导致 panic。
func Apply(db *sql.DB, cfg Config) {
	if db == nil {
		return
	}
	if cfg.MaxOpen > 0 {
		db.SetMaxOpenConns(cfg.MaxOpen)
	}
	if cfg.MaxIdle > 0 {
		db.SetMaxIdleConns(cfg.MaxIdle)
	} else {
		db.SetMaxIdleConns(2)
	}
	if cfg.MaxLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)
	}
}
