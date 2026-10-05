// Package seed 负责在数据库为空时写入开发与演示用的种子数据，
// 包括：数据源、Schema 元信息、本体（类/属性/关系/规则）、映射配置，
// 以及可供实际查询执行的 SQLite 物理数据文件。
package seed

import (
	"encoding/json"

	"gorm.io/gorm"

	"smartg-odin/internal/common/logger"
	"smartg-odin/internal/model"
)

// propertyMapping 对应 OntMappingConfig.PropertyMappingsJSON 中的单项。
type propertyMapping struct {
	PropertyID          uint              `json:"property_id"`
	PropertyName        string            `json:"property_name"`
	ColumnName          string            `json:"column_name"`
	TransformExpression string            `json:"transform_expression"`
	ValueMap            map[string]string `json:"value_map"`
	Confidence          float64           `json:"confidence"`
	ValidationStatus    string            `json:"validation_status"`
}

// Run 执行种子数据插入（仅当 datasource 表为空时）。
func Run(db *gorm.DB) {
	var count int64
	db.Model(&model.DataSource{}).Count(&count)
	if count > 0 {
		return
	}

	logger.L.Info("Seeding database...")

	sourceIDs := seedDatasources(db)
	seedSchemas(db, sourceIDs)
	createSQLiteFiles()
	seedOntologies(db, sourceIDs)

	logger.L.Info("Database seeded successfully")
}

// mustJSON 将任意值序列化为 JSON 字符串，失败时返回空串。
func mustJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

// ---- 数据源 ----

// seedDatasources 写入 3 个数据源，返回 code→ID 映射。
func seedDatasources(db *gorm.DB) map[string]uint {
	sources := []model.DataSource{
		{
			// 真实 PostgreSQL 容器 odin-pgsql（docker-compose 映射 15432:5432）。
			Code: "erp_postgres", Name: "ERP PostgreSQL", Type: "postgres",
			Host: "localhost", Port: 15432, DatabaseName: "demo",
			Username: "odin", Password: "Cqmygnsdss@123", SchemaName: "public", SSLMode: "disable",
			Status: "active", Description: "仓储/物流库（PostgreSQL 17，odin-pgsql:15432/demo）",
		},
		{
			// 真实 MySQL 容器 odin-mysql（docker-compose 映射 13306:3306）。
			Code: "crm_mysql", Name: "CRM MySQL", Type: "mysql",
			Host: "localhost", Port: 13306, DatabaseName: "smartj",
			Username: "root", Password: "odin123456", SSLMode: "disable",
			Status: "active", Description: "ERP 客户/订单/产品库（MySQL 8.4，odin-mysql:13306/smartj）",
		},
		{
			Code: "warehouse_sqlite", Name: "Warehouse SQLite", Type: "sqlite",
			DatabaseName: "./data/warehouse.db",
			Status:       "active", Description: "本地仓储/销售演示库（SQLite）",
		},
	}

	ids := make(map[string]uint, len(sources))
	for i := range sources {
		if err := db.Create(&sources[i]).Error; err != nil {
			logger.L.Errorf("seed datasource %s failed: %v", sources[i].Code, err)
			continue
		}
		ids[sources[i].Code] = sources[i].ID
	}
	return ids
}

// ---- Schema 元信息 ----

// columnSpec 描述一个字段元信息。
type columnSpec struct {
	Name       string
	DataType   string
	Nullable   bool
	PrimaryKey bool
	Comment    string
}

// tableSpec 描述一张表及其字段。
type tableSpec struct {
	Name    string
	Comment string
	Rows    int64
	Columns []columnSpec
}

// seedSchemas 为 SQLite 数据源写入表/字段元信息（启动时会被内省结果刷新）。
func seedSchemas(db *gorm.DB, sourceIDs map[string]uint) {
	specs := map[string][]tableSpec{
		"warehouse_sqlite": warehouseTables(),
	}
	for code, tables := range specs {
		dsID, ok := sourceIDs[code]
		if !ok {
			continue
		}
		for _, t := range tables {
			st := model.SchemaTable{
				DataSourceID: dsID,
				TableName:    t.Name,
				TableType:    "TABLE",
				Comment:      t.Comment,
				RowCount:     t.Rows,
			}
			if err := db.Create(&st).Error; err != nil {
				logger.L.Errorf("seed schema table %s failed: %v", t.Name, err)
				continue
			}
			for pos, c := range t.Columns {
				sc := model.SchemaColumn{
					TableID:      st.ID,
					DataSourceID: dsID,
					ColumnName:   c.Name,
					DataType:     c.DataType,
					IsNullable:   c.Nullable,
					IsPrimaryKey: c.PrimaryKey,
					Comment:      c.Comment,
					OrdinalPos:   pos + 1,
				}
				db.Create(&sc)
			}
		}
	}
}

func warehouseTables() []tableSpec {
	return []tableSpec{
		{
			Name: "customers", Comment: "客户主数据", Rows: 5,
			Columns: []columnSpec{
				{Name: "id", DataType: "INTEGER", PrimaryKey: true},
				{Name: "name", DataType: "TEXT", Nullable: true, Comment: "客户名称"},
				{Name: "email", DataType: "TEXT", Nullable: true, Comment: "邮箱"},
				{Name: "level", DataType: "TEXT", Nullable: true, Comment: "等级 VIP/Normal"},
				{Name: "total_orders", DataType: "INTEGER", Nullable: true, Comment: "累计订单数"},
				{Name: "created_at", DataType: "TEXT", Nullable: true, Comment: "创建时间"},
			},
		},
		{
			Name: "orders", Comment: "销售订单", Rows: 8,
			Columns: []columnSpec{
				{Name: "id", DataType: "INTEGER", PrimaryKey: true},
				{Name: "customer_id", DataType: "INTEGER", Nullable: true, Comment: "客户ID"},
				{Name: "product", DataType: "TEXT", Nullable: true, Comment: "产品名称"},
				{Name: "quantity", DataType: "INTEGER", Nullable: true, Comment: "数量"},
				{Name: "amount", DataType: "REAL", Nullable: true, Comment: "金额"},
				{Name: "status", DataType: "TEXT", Nullable: true, Comment: "状态"},
				{Name: "created_at", DataType: "TEXT", Nullable: true, Comment: "下单时间"},
			},
		},
		{
			Name: "products", Comment: "产品目录", Rows: 5,
			Columns: []columnSpec{
				{Name: "id", DataType: "INTEGER", PrimaryKey: true},
				{Name: "name", DataType: "TEXT", Nullable: true, Comment: "产品名称"},
				{Name: "category", DataType: "TEXT", Nullable: true, Comment: "品类"},
				{Name: "price", DataType: "REAL", Nullable: true, Comment: "单价"},
				{Name: "stock", DataType: "INTEGER", Nullable: true, Comment: "库存"},
			},
		},
	}
}
