package query

// 本文件是仓库首批自动化测试，聚焦聚合查询翻译（Translate → SQL）的
// golden-SQL 断言：单表 count/sum/avg/min/max、group by、having、TOP-N、
// 跨类 JOIN group by、1:N 扇出下的 COUNT(DISTINCT)、跨源聚合拒绝，
// 以及「同请求重复调用逐字节一致」的确定性抽查。
//
// 测试使用内存 SQLite 承载元数据（本体/映射/数据源），业务数据源类型标记为
// mysql / postgres 以驱动方言引用符（反引号 / 双引号）分支，但不真正连库执行 SQL。

import (
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"smartg-odin/internal/model"
)

// seedDB 构建并填充一个内存 SQLite 元数据库，返回可复用的 *gorm.DB。
// 使用 MaxOpenConns(1) 保证 ":memory:" 下所有查询命中同一个内存库实例。
func seedDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(
		&model.DataSource{},
		&model.OntDefinition{},
		&model.OntClass{},
		&model.OntClassProperty{},
		&model.OntRelation{},
		&model.OntRule{},
		&model.OntMappingConfig{},
		&model.SchemaTable{},
		&model.SchemaColumn{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	// 数据源：1=mysql(业务 crm_mysql)，2=postgres(erp_postgres，用于跨源拒绝)。
	mustCreate(t, db, &model.DataSource{ID: 1, Code: "crm_mysql", Name: "crm_mysql", Type: "mysql", Status: "active"})
	mustCreate(t, db, &model.DataSource{ID: 2, Code: "erp_postgres", Name: "erp_postgres", Type: "postgres", Status: "active"})

	mustCreate(t, db, &model.OntDefinition{ID: 1, Code: "test", Name: "测试本体", Status: "published"})

	// 类。
	mustCreate(t, db, &model.OntClass{ID: 10, OntologyID: 1, Name: "Order", Label: "订单", ClassType: "normal"})
	mustCreate(t, db, &model.OntClass{ID: 11, OntologyID: 1, Name: "Product", Label: "产品", ClassType: "normal"})
	mustCreate(t, db, &model.OntClass{ID: 12, OntologyID: 1, Name: "WarehouseInventory", Label: "库存", ClassType: "normal"})
	// 虚拟类 MultiSource → 两个跨数据源子类，用于 Translate 层多分支聚合拒绝。
	mustCreate(t, db, &model.OntClass{ID: 13, OntologyID: 1, Name: "MultiSource", Label: "多源", ClassType: "virtual"})
	mustCreate(t, db, &model.OntClass{ID: 14, OntologyID: 1, Name: "SubA", Label: "子A", ClassType: "normal", ParentClassID: uPtr(13)})
	mustCreate(t, db, &model.OntClass{ID: 15, OntologyID: 1, Name: "SubB", Label: "子B", ClassType: "normal", ParentClassID: uPtr(13)})

	// 属性（带中文 label，驱动 AS 别名与缺省聚合别名）。
	mustCreate(t, db, &model.OntClassProperty{ID: 100, ClassID: 10, OntologyID: 1, Name: "orderNo", Label: "订单号", DataType: "string"})
	mustCreate(t, db, &model.OntClassProperty{ID: 101, ClassID: 10, OntologyID: 1, Name: "quantity", Label: "数量", DataType: "integer"})
	mustCreate(t, db, &model.OntClassProperty{ID: 102, ClassID: 10, OntologyID: 1, Name: "amount", Label: "金额", DataType: "float"})
	mustCreate(t, db, &model.OntClassProperty{ID: 110, ClassID: 11, OntologyID: 1, Name: "name", Label: "产品名称", DataType: "string"})
	mustCreate(t, db, &model.OntClassProperty{ID: 111, ClassID: 11, OntologyID: 1, Name: "price", Label: "单价", DataType: "float"})
	mustCreate(t, db, &model.OntClassProperty{ID: 120, ClassID: 12, OntologyID: 1, Name: "stock", Label: "库存", DataType: "integer"})
	mustCreate(t, db, &model.OntClassProperty{ID: 140, ClassID: 14, OntologyID: 1, Name: "aField", Label: "A字段", DataType: "string"})
	mustCreate(t, db, &model.OntClassProperty{ID: 150, ClassID: 15, OntologyID: 1, Name: "bField", Label: "B字段", DataType: "string"})

	// 关系。
	mustCreate(t, db, &model.OntRelation{ID: 1, OntologyID: 1, Name: "contains", Label: "包含", FromClassID: 10, ToClassID: 11, RelationType: "association", Cardinality: "N:1", JoinConditionJSON: `{"from_column":"product_id","to_column":"id","join_type":"LEFT"}`})
	mustCreate(t, db, &model.OntRelation{ID: 2, OntologyID: 1, Name: "orderedIn", Label: "对应订单", FromClassID: 11, ToClassID: 10, RelationType: "association", Cardinality: "1:N", JoinConditionJSON: `{"from_column":"id","to_column":"product_id","join_type":"LEFT"}`})
	mustCreate(t, db, &model.OntRelation{ID: 3, OntologyID: 1, Name: "stores", Label: "库存于", FromClassID: 11, ToClassID: 12, RelationType: "association", Cardinality: "1:N", JoinConditionJSON: `{"from_column":"id","to_column":"product_id","join_type":"LEFT"}`})

	// 映射配置。
	orderMappings := `[{"property_id":100,"property_name":"orderNo","column_name":"order_no"},{"property_id":101,"property_name":"quantity","column_name":"quantity"},{"property_id":102,"property_name":"amount","column_name":"amount"}]`
	productMappings := `[{"property_id":110,"property_name":"name","column_name":"name"},{"property_id":111,"property_name":"price","column_name":"price"}]`
	whMappings := `[{"property_id":120,"property_name":"stock","column_name":"stock"}]`
	mustCreate(t, db, &model.OntMappingConfig{ID: 1, OntologyID: 1, ClassID: 10, DataSourceID: 1, SourceTable: "orders", PropertyMappingsJSON: orderMappings, Status: "active"})
	mustCreate(t, db, &model.OntMappingConfig{ID: 2, OntologyID: 1, ClassID: 11, DataSourceID: 1, SourceTable: "products", PropertyMappingsJSON: productMappings, Status: "active"})
	mustCreate(t, db, &model.OntMappingConfig{ID: 3, OntologyID: 1, ClassID: 12, DataSourceID: 2, SourceTable: "warehouse_inventory", PropertyMappingsJSON: whMappings, Status: "active"})
	mustCreate(t, db, &model.OntMappingConfig{ID: 4, OntologyID: 1, ClassID: 14, DataSourceID: 1, SourceTable: "sub_a", PropertyMappingsJSON: `[{"property_id":140,"property_name":"aField","column_name":"a_field"}]`, Status: "active"})
	mustCreate(t, db, &model.OntMappingConfig{ID: 5, OntologyID: 1, ClassID: 15, DataSourceID: 2, SourceTable: "sub_b", PropertyMappingsJSON: `[{"property_id":150,"property_name":"bField","column_name":"b_field"}]`, Status: "active"})

	// products 表主键元数据（供 1:N 扇出 COUNT(DISTINCT t0.id) 解析）。
	mustCreate(t, db, &model.SchemaTable{ID: 1, DataSourceID: 1, TableName: "products"})
	mustCreate(t, db, &model.SchemaColumn{ID: 1, TableID: 1, DataSourceID: 1, ColumnName: "id", DataType: "bigint", IsPrimaryKey: true, OrdinalPos: 1})

	return db
}

func mustCreate(t *testing.T, db *gorm.DB, rec interface{}) {
	t.Helper()
	if err := db.Create(rec).Error; err != nil {
		t.Fatalf("seed record %#v: %v", rec, err)
	}
}

func uPtr(v uint) *uint { return &v }

// assertSQL 运行 Translate 并断言 SQL 逐字节等于 want，同时返回结果供进一步断言。
func assertSQL(t *testing.T, db *gorm.DB, req *QueryRequest, want string) *TranslatedQuery {
	t.Helper()
	tq, err := Translate(db, req)
	if err != nil {
		t.Fatalf("Translate returned error: %v", err)
	}
	if tq.SQL != want {
		t.Fatalf("SQL mismatch\n got: %s\nwant: %s", tq.SQL, want)
	}
	return tq
}

// 单表聚合：count(*)/sum/avg/min/max，中文缺省别名，MySQL 反引号，行级溯源常量列。
func TestAggregateSingleTable(t *testing.T) {
	db := seedDB(t)
	const lineage = ", 'crm_mysql' AS `_src_datasource`, 'orders' AS `_src_table` FROM orders t0"

	cases := []struct {
		name string
		agg  Aggregate
		want string
	}{
		{"count_star", Aggregate{Func: "count"}, "SELECT COUNT(*) AS `计数`" + lineage},
		{"sum", Aggregate{Func: "sum", Property: "quantity"}, "SELECT SUM(t0.`quantity`) AS `合计数量`" + lineage},
		{"avg", Aggregate{Func: "avg", Property: "amount"}, "SELECT AVG(t0.`amount`) AS `平均金额`" + lineage},
		{"min", Aggregate{Func: "min", Property: "amount"}, "SELECT MIN(t0.`amount`) AS `最小金额`" + lineage},
		{"max", Aggregate{Func: "max", Property: "amount"}, "SELECT MAX(t0.`amount`) AS `最大金额`" + lineage},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := &QueryRequest{OntologyID: 1, ClassName: "Order", Aggregates: []Aggregate{c.agg}}
			tq := assertSQL(t, db, req, c.want)
			if !tq.Aggregated {
				t.Fatalf("expected Aggregated=true")
			}
		})
	}
}

// 单表 GROUP BY：SELECT = 维度列(中文 label 别名) ∪ 聚合列 ∪ 溯源列。
func TestAggregateGroupBy(t *testing.T) {
	db := seedDB(t)
	req := &QueryRequest{
		OntologyID: 1,
		ClassName:  "Order",
		GroupBy:    []string{"orderNo"},
		Aggregates: []Aggregate{{Func: "sum", Property: "amount"}},
	}
	want := "SELECT t0.`order_no` AS `订单号`, SUM(t0.`amount`) AS `合计金额`, 'crm_mysql' AS `_src_datasource`, 'orders' AS `_src_table` FROM orders t0 GROUP BY t0.`order_no`"
	tq := assertSQL(t, db, req, want)
	if len(tq.ColumnMeta) != 2 || tq.ColumnMeta[0].Role != "dimension" || tq.ColumnMeta[1].Role != "measure" || tq.ColumnMeta[1].Agg != "sum" {
		t.Fatalf("unexpected column_meta: %+v", tq.ColumnMeta)
	}
}

// HAVING：引用聚合表达式（非别名）以兼容 PG，参数在 WHERE 之后绑定。
func TestAggregateHaving(t *testing.T) {
	db := seedDB(t)
	req := &QueryRequest{
		OntologyID: 1,
		ClassName:  "Order",
		Aggregates: []Aggregate{{Func: "sum", Property: "amount", Alias: "总额"}},
		Having:     []HavingClause{{Alias: "总额", Op: "gt", Value: 1000}},
	}
	want := "SELECT SUM(t0.`amount`) AS `总额`, 'crm_mysql' AS `_src_datasource`, 'orders' AS `_src_table` FROM orders t0 HAVING SUM(t0.`amount`) > ?"
	tq := assertSQL(t, db, req, want)
	if len(tq.Params) != 1 || tq.Params[0] != 1000 {
		t.Fatalf("expected params [1000], got %v", tq.Params)
	}
}

// 跨类 JOIN + GROUP BY + ORDER BY 聚合别名 + LIMIT = TOP-N（样例1 语义）。
func TestAggregateCrossClassTopN(t *testing.T) {
	db := seedDB(t)
	req := &QueryRequest{
		OntologyID: 1,
		ClassName:  "Product",
		Relations:  []string{"orderedIn"},
		GroupBy:    []string{"name"},
		Aggregates: []Aggregate{{Func: "sum", Property: "quantity"}},
		OrderBy:    "合计数量",
		OrderDir:   "desc",
		Limit:      3,
	}
	want := "SELECT t0.`name` AS `产品名称`, SUM(t1.`quantity`) AS `合计数量`, 'crm_mysql' AS `_src_datasource`, 'products' AS `_src_table` FROM products t0 LEFT JOIN orders t1 ON t0.`id` = t1.`product_id` GROUP BY t0.`name` ORDER BY `合计数量` DESC LIMIT 3"
	assertSQL(t, db, req, want)
}

// 1:N 扇出下的 COUNT(*) → COUNT(DISTINCT t0.<pk>) 防重复计数主实体（标量：group_by 为空）。
func TestAggregateFanoutCountDistinct(t *testing.T) {
	db := seedDB(t)
	req := &QueryRequest{
		OntologyID: 1,
		ClassName:  "Product",
		Relations:  []string{"orderedIn"},
		Aggregates: []Aggregate{{Func: "count"}},
	}
	want := "SELECT COUNT(DISTINCT t0.`id`) AS `计数`, 'crm_mysql' AS `_src_datasource`, 'products' AS `_src_table` FROM products t0 LEFT JOIN orders t1 ON t0.`id` = t1.`product_id`"
	assertSQL(t, db, req, want)
}

// seedCustomer 补充 Customer 类、places(1:N) 关系、customers/orders 主键元数据，
// 用于任务 #54（按基类分组计数口径）相关用例。
func seedCustomer(t *testing.T, db *gorm.DB) {
	t.Helper()
	mustCreate(t, db, &model.OntClass{ID: 16, OntologyID: 1, Name: "Customer", Label: "客户", ClassType: "normal"})
	mustCreate(t, db, &model.OntClassProperty{ID: 160, ClassID: 16, OntologyID: 1, Name: "name", Label: "客户名称", DataType: "string"})
	mustCreate(t, db, &model.OntRelation{ID: 4, OntologyID: 1, Name: "places", Label: "下单", FromClassID: 16, ToClassID: 10, RelationType: "association", Cardinality: "1:N", JoinConditionJSON: `{"from_column":"id","to_column":"customer_id","join_type":"LEFT"}`})
	mustCreate(t, db, &model.OntMappingConfig{ID: 6, OntologyID: 1, ClassID: 16, DataSourceID: 1, SourceTable: "customers", PropertyMappingsJSON: `[{"property_id":160,"property_name":"name","column_name":"name"}]`, Status: "active"})
	mustCreate(t, db, &model.SchemaTable{ID: 2, DataSourceID: 1, TableName: "customers"})
	mustCreate(t, db, &model.SchemaColumn{ID: 2, TableID: 2, DataSourceID: 1, ColumnName: "id", DataType: "bigint", IsPrimaryKey: true, OrdinalPos: 1})
	mustCreate(t, db, &model.SchemaTable{ID: 3, DataSourceID: 1, TableName: "orders"})
	mustCreate(t, db, &model.SchemaColumn{ID: 3, TableID: 3, DataSourceID: 1, ColumnName: "id", DataType: "bigint", IsPrimaryKey: true, OrdinalPos: 1})
}

// 任务 #54(a)：按基类列分组 + count(无 property) + 1:N JOIN → COUNT(t1.<子表PK>)（子表行数）。
// 旧行为 COUNT(DISTINCT t0.id) 在每个客户分组内恒=1，TOP-N 排序无意义，属错误语义。
func TestAggregateFanoutCountGroupByBase(t *testing.T) {
	db := seedDB(t)
	seedCustomer(t, db)
	req := &QueryRequest{
		OntologyID: 1,
		ClassName:  "Customer",
		Relations:  []string{"places"},
		GroupBy:    []string{"name"},
		Aggregates: []Aggregate{{Func: "count"}},
		OrderBy:    "计数",
		OrderDir:   "desc",
		Limit:      3,
	}
	want := "SELECT t0.`name` AS `客户名称`, COUNT(t1.`id`) AS `计数`, 'crm_mysql' AS `_src_datasource`, 'customers' AS `_src_table` FROM customers t0 LEFT JOIN orders t1 ON t0.`id` = t1.`customer_id` GROUP BY t0.`name` ORDER BY `计数` DESC LIMIT 3"
	tq := assertSQL(t, db, req, want)
	if !strings.Contains(tq.Explanation, "计数口径") || !strings.Contains(tq.Explanation, "COUNT(t1.`id`)") {
		t.Fatalf("explanation 应说明计数口径，got: %s", tq.Explanation)
	}
}

// 任务 #54(b)：标量（group_by 为空）「有多少客户下过单」→ 保持 COUNT(DISTINCT t0.<基表PK>)。
func TestAggregateFanoutCountScalarKeepsDistinct(t *testing.T) {
	db := seedDB(t)
	seedCustomer(t, db)
	req := &QueryRequest{
		OntologyID: 1,
		ClassName:  "Customer",
		Relations:  []string{"places"},
		Aggregates: []Aggregate{{Func: "count"}},
	}
	want := "SELECT COUNT(DISTINCT t0.`id`) AS `计数`, 'crm_mysql' AS `_src_datasource`, 'customers' AS `_src_table` FROM customers t0 LEFT JOIN orders t1 ON t0.`id` = t1.`customer_id`"
	tq := assertSQL(t, db, req, want)
	if !strings.Contains(tq.Explanation, "去重计数") {
		t.Fatalf("explanation 应说明去重计数口径，got: %s", tq.Explanation)
	}
}

// 任务 #54 补充：group_by 仅引用子/关系列（统计基实体数去重）→ 仍为 COUNT(DISTINCT t0.<pk>)。
func TestAggregateFanoutCountGroupByChildKeepsDistinct(t *testing.T) {
	db := seedDB(t)
	req := &QueryRequest{
		OntologyID: 1,
		ClassName:  "Product",
		Relations:  []string{"orderedIn"},
		GroupBy:    []string{"orderedIn.orderNo"},
		Aggregates: []Aggregate{{Func: "count"}},
	}
	want := "SELECT t1.`order_no` AS `订单号`, COUNT(DISTINCT t0.`id`) AS `计数`, 'crm_mysql' AS `_src_datasource`, 'products' AS `_src_table` FROM products t0 LEFT JOIN orders t1 ON t0.`id` = t1.`product_id` GROUP BY t1.`order_no`"
	assertSQL(t, db, req, want)
}

// 任务 #54 补充：count 带 property 的语义保持不变（扇出下 COUNT(DISTINCT col)，即使按基类分组）。
func TestAggregateFanoutCountWithPropertyUnchanged(t *testing.T) {
	db := seedDB(t)
	seedCustomer(t, db)
	req := &QueryRequest{
		OntologyID: 1,
		ClassName:  "Customer",
		Relations:  []string{"places"},
		GroupBy:    []string{"name"},
		Aggregates: []Aggregate{{Func: "count", Property: "places.orderNo"}},
	}
	want := "SELECT t0.`name` AS `客户名称`, COUNT(DISTINCT t1.`order_no`) AS `计数订单号`, 'crm_mysql' AS `_src_datasource`, 'customers' AS `_src_table` FROM customers t0 LEFT JOIN orders t1 ON t0.`id` = t1.`customer_id` GROUP BY t0.`name`"
	assertSQL(t, db, req, want)
}

// 任务 #54(c) 无 JOIN 回归：N:1 收敛 JOIN 下的 count 不受扇出保护影响，仍为 COUNT(*)。
func TestAggregateConvergentJoinCountStar(t *testing.T) {
	db := seedDB(t)
	req := &QueryRequest{
		OntologyID: 1,
		ClassName:  "Order",
		Relations:  []string{"contains"},
		GroupBy:    []string{"contains.name"},
		Aggregates: []Aggregate{{Func: "count"}},
	}
	want := "SELECT t1.`name` AS `产品名称`, COUNT(*) AS `计数`, 'crm_mysql' AS `_src_datasource`, 'orders' AS `_src_table` FROM orders t0 LEFT JOIN products t1 ON t0.`product_id` = t1.`id` GROUP BY t1.`name`"
	assertSQL(t, db, req, want)
}

// 跨源拒绝之一：JOIN 目标跨数据源（Product@mysql --stores--> WarehouseInventory@pg）。
func TestRejectCrossSourceJoin(t *testing.T) {
	db := seedDB(t)
	req := &QueryRequest{
		OntologyID: 1,
		ClassName:  "Product",
		Relations:  []string{"stores"},
		Aggregates: []Aggregate{{Func: "sum", Property: "stores.stock"}},
	}
	_, err := Translate(db, req)
	if err == nil {
		t.Fatalf("expected cross-source rejection error, got nil")
	}
	if !strings.Contains(err.Error(), "跨数据源") {
		t.Fatalf("error should mention 跨数据源, got: %v", err)
	}
}

// 跨源拒绝之二：虚拟类展开为多个跨源物理目标 + 聚合 → Translate 层拒绝。
func TestRejectCrossSourceUnion(t *testing.T) {
	db := seedDB(t)
	req := &QueryRequest{
		OntologyID: 1,
		ClassName:  "MultiSource",
		Aggregates: []Aggregate{{Func: "count"}},
	}
	_, err := Translate(db, req)
	if err == nil {
		t.Fatalf("expected union aggregate rejection error, got nil")
	}
	if !strings.Contains(err.Error(), "跨数据源") {
		t.Fatalf("error should mention 跨数据源, got: %v", err)
	}
}

// 确定性抽查：同请求重复调用，SQL 与参数逐字节一致。
func TestDeterministicRepeat(t *testing.T) {
	db := seedDB(t)
	newReq := func() *QueryRequest {
		return &QueryRequest{
			OntologyID: 1,
			ClassName:  "Product",
			Relations:  []string{"orderedIn"},
			GroupBy:    []string{"name"},
			Aggregates: []Aggregate{{Func: "sum", Property: "quantity"}},
			OrderBy:    "合计数量",
			OrderDir:   "desc",
			Limit:      3,
		}
	}
	first, err := Translate(db, newReq())
	if err != nil {
		t.Fatalf("first Translate: %v", err)
	}
	for i := 0; i < 5; i++ {
		next, err := Translate(db, newReq())
		if err != nil {
			t.Fatalf("repeat #%d Translate: %v", i, err)
		}
		if next.SQL != first.SQL {
			t.Fatalf("non-deterministic SQL\n first: %s\n next:  %s", first.SQL, next.SQL)
		}
		if len(next.Params) != len(first.Params) {
			t.Fatalf("param count drift: %d vs %d", len(next.Params), len(first.Params))
		}
	}
}

// ---------------------------------------------------------------------------
// 任务 #55 P2-3：1:N 扇出放大修复（确定性丢弃 / 警告）golden 测试。
// ---------------------------------------------------------------------------

// P2-3 丢弃：主类度量 SUM(price) + 未被引用的 1:N 关系 orderedIn → 确定性丢弃该 JOIN，
// SQL 不再 JOIN 子表（消除按子表行数重复放大），无 warning，DroppedJoins 记录被丢弃关系。
func TestFanoutDropUnreferencedOneToMany(t *testing.T) {
	db := seedDB(t)
	req := &QueryRequest{
		OntologyID: 1,
		ClassName:  "Product",
		Relations:  []string{"orderedIn"},
		GroupBy:    []string{"name"},
		Aggregates: []Aggregate{{Func: "sum", Property: "price"}},
	}
	// 丢弃 orderedIn 后：无 JOIN，SUM(t0.price) 不再被 orders 行数放大。
	want := "SELECT t0.`name` AS `产品名称`, SUM(t0.`price`) AS `合计单价`, 'crm_mysql' AS `_src_datasource`, 'products' AS `_src_table` FROM products t0 GROUP BY t0.`name`"
	tq := assertSQL(t, db, req, want)
	if strings.Contains(tq.SQL, "JOIN") {
		t.Fatalf("expected 1:N join dropped, got SQL with JOIN: %s", tq.SQL)
	}
	if len(tq.DroppedJoins) != 1 || tq.DroppedJoins[0] != "orderedIn" {
		t.Fatalf("expected DroppedJoins=[orderedIn], got %v", tq.DroppedJoins)
	}
	if tq.Warning != "" {
		t.Fatalf("expected no warning after deterministic drop, got %q", tq.Warning)
	}
	if !strings.Contains(tq.Explanation, "丢弃") {
		t.Fatalf("explanation 应说明丢弃未被引用的 1:N 关系，got: %s", tq.Explanation)
	}
}

// P2-3 警告：1:N 关系 orderedIn 被 filter 显式引用（无法安全丢弃）+ 主类度量 SUM(price) →
// 保留 JOIN 但在 Warning 声明「主类度量可能扇出放大」，供前端提示。
func TestFanoutWarnReferencedOneToMany(t *testing.T) {
	db := seedDB(t)
	req := &QueryRequest{
		OntologyID: 1,
		ClassName:  "Product",
		Relations:  []string{"orderedIn"},
		GroupBy:    []string{"name"},
		Aggregates: []Aggregate{{Func: "sum", Property: "price"}},
		Filters:    []Filter{{Property: "orderedIn.orderNo", Op: "eq", Value: "NO-1"}},
	}
	tq, err := Translate(db, req)
	if err != nil {
		t.Fatalf("Translate returned error: %v", err)
	}
	if !strings.Contains(tq.SQL, "JOIN") {
		t.Fatalf("referenced 1:N join must be kept, got SQL: %s", tq.SQL)
	}
	if len(tq.DroppedJoins) != 0 {
		t.Fatalf("expected no dropped joins, got %v", tq.DroppedJoins)
	}
	if !strings.Contains(tq.Warning, "1:N") || !strings.Contains(tq.Warning, "扇出放大") {
		t.Fatalf("expected fanout warning, got %q", tq.Warning)
	}
	if !strings.Contains(tq.Explanation, "扇出放大") {
		t.Fatalf("explanation 应含扇出放大警告，got: %s", tq.Explanation)
	}
}

// P2-3 无放大不丢弃：主类无 SUM/AVG/MIN/MAX 度量（仅 count，已由 DISTINCT 保护）时，
// 未被引用的 1:N 关系保持既有口径（不丢弃），与 TestAggregateFanoutCountDistinct 一致。
func TestFanoutCountNotDropped(t *testing.T) {
	db := seedDB(t)
	req := &QueryRequest{
		OntologyID: 1,
		ClassName:  "Product",
		Relations:  []string{"orderedIn"},
		Aggregates: []Aggregate{{Func: "count"}},
	}
	want := "SELECT COUNT(DISTINCT t0.`id`) AS `计数`, 'crm_mysql' AS `_src_datasource`, 'products' AS `_src_table` FROM products t0 LEFT JOIN orders t1 ON t0.`id` = t1.`product_id`"
	tq := assertSQL(t, db, req, want)
	if len(tq.DroppedJoins) != 0 {
		t.Fatalf("count 口径不应丢弃 join，got DroppedJoins=%v", tq.DroppedJoins)
	}
}

// P2-3 子表度量不丢弃：SUM 度量解析到子表（orderedIn.quantity，非主类）时不存在主类放大，
// 关系被 aggregate 显式引用 → 保留 JOIN，无丢弃、无警告。
func TestFanoutChildMeasureKept(t *testing.T) {
	db := seedDB(t)
	req := &QueryRequest{
		OntologyID: 1,
		ClassName:  "Product",
		Relations:  []string{"orderedIn"},
		GroupBy:    []string{"name"},
		Aggregates: []Aggregate{{Func: "sum", Property: "orderedIn.quantity"}},
	}
	tq, err := Translate(db, req)
	if err != nil {
		t.Fatalf("Translate returned error: %v", err)
	}
	if !strings.Contains(tq.SQL, "JOIN") {
		t.Fatalf("子表度量需保留 JOIN，got SQL: %s", tq.SQL)
	}
	if len(tq.DroppedJoins) != 0 || tq.Warning != "" {
		t.Fatalf("子表度量不应丢弃/警告，got dropped=%v warning=%q", tq.DroppedJoins, tq.Warning)
	}
}
