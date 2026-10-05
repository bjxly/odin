package database

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"smartg-odin/internal/common/logger"
	"smartg-odin/internal/model"
)

// DB 全局 GORM 实例，Init 成功后可用。
var DB *gorm.DB

// Init 使用 SQLite 打开数据库并执行自动迁移。
// dsn 为数据库文件路径，例如 ./data/odin.db。
func Init(dsn string) error {
	return InitWithDriver("sqlite", dsn)
}

// InitWithDriver 根据 driver 打开数据库并执行自动迁移。
// 目前仅支持 sqlite，预留 mysql / postgres 扩展点。
func InitWithDriver(driver, dsn string) error {
	if driver == "" {
		driver = "sqlite"
	}

	// 确保数据文件所在目录存在
	if driver == "sqlite" {
		if dir := filepath.Dir(dsn); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("create db dir: %w", err)
			}
		}
	}

	var (
		db  *gorm.DB
		err error
	)

	cfg := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	}

	switch driver {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(dsn), cfg)
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	DB = db
	if err := AutoMigrate(); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	logger.Infof("database connected: %s (%s)", dsn, driver)
	return nil
}

// AutoMigrate 对所有模型执行自动建表/迁移。
func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	return DB.AutoMigrate(
		&model.DataSource{},
		&model.DataSourceTest{},
		&model.SchemaTable{},
		&model.SchemaColumn{},
		&model.SchemaForeignKey{},
		&model.SchemaIndex{},
		&model.OntDefinition{},
		&model.OntClass{},
		&model.OntClassProperty{},
		&model.OntRelation{},
		&model.OntRule{},
		&model.OntAction{},
		&model.OntMappingConfig{},
		&model.QueryTemplate{},
		&model.QueryHistory{},
		&model.QueryTrace{},
		&model.IntentCache{},
		&model.AssistantSession{},
		&model.AssistantMessage{},
		&model.AuditLog{},
	)
}

// GetDB 返回全局 GORM 实例。
func GetDB() *gorm.DB { return DB }

// Close 关闭底层 SQL 连接。
func Close() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
