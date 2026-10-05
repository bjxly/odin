package query

// IN 操作符 golden 测试：验证派生后置分类器依赖的 `in` 过滤器翻译正确，
// 且多个过滤器的参数按 WHERE 出现顺序绑定（PG $N / MySQL ? 占位符顺序正确）。
// 复用 translator_agg_test.go 的 seedDB 内存元数据库。

import (
	"testing"
)

// TestFilterInOperator 单 IN 过滤：占位符数量与参数一致，列名解析为物理列。
func TestFilterInOperator(t *testing.T) {
	db := seedDB(t)
	req := &QueryRequest{
		OntologyID: 1,
		ClassName:  "Order",
		Filters:    []Filter{{Property: "orderNo", Op: "in", Value: []interface{}{"A", "B", "C"}}},
		Aggregates: []Aggregate{{Func: "count"}},
	}
	want := "SELECT COUNT(*) AS `计数`, 'crm_mysql' AS `_src_datasource`, 'orders' AS `_src_table` FROM orders t0 WHERE t0.`order_no` IN (?,?,?)"
	tq := assertSQL(t, db, req, want)
	if len(tq.Params) != 3 || tq.Params[0] != "A" || tq.Params[1] != "B" || tq.Params[2] != "C" {
		t.Fatalf("IN 参数应为 [A B C]，得到 %v", tq.Params)
	}
}

// TestFilterInParamOrder 多过滤器参数绑定顺序：gt(?) 在前、in(?,?,?) 在后，顺序不可错位。
func TestFilterInParamOrder(t *testing.T) {
	db := seedDB(t)
	req := &QueryRequest{
		OntologyID: 1,
		ClassName:  "Order",
		Filters: []Filter{
			{Property: "amount", Op: "gt", Value: 100},
			{Property: "orderNo", Op: "in", Value: []interface{}{"X", "Y"}},
		},
		Aggregates: []Aggregate{{Func: "count"}},
	}
	want := "SELECT COUNT(*) AS `计数`, 'crm_mysql' AS `_src_datasource`, 'orders' AS `_src_table` FROM orders t0 WHERE t0.`amount` > ? AND t0.`order_no` IN (?,?)"
	tq := assertSQL(t, db, req, want)
	// 绑定顺序必须是 [100, X, Y]——占位符与参数一一对应，错位将导致查询语义错误。
	if len(tq.Params) != 3 || tq.Params[0] != 100 || tq.Params[1] != "X" || tq.Params[2] != "Y" {
		t.Fatalf("参数绑定顺序应为 [100 X Y]，得到 %v", tq.Params)
	}
}

// TestFilterInEmptyIgnored 空 IN 值列表应被忽略（不生成非法 `IN ()`）。
func TestFilterInEmptyIgnored(t *testing.T) {
	db := seedDB(t)
	req := &QueryRequest{
		OntologyID: 1,
		ClassName:  "Order",
		Filters:    []Filter{{Property: "orderNo", Op: "in", Value: []interface{}{}}},
		Aggregates: []Aggregate{{Func: "count"}},
	}
	want := "SELECT COUNT(*) AS `计数`, 'crm_mysql' AS `_src_datasource`, 'orders' AS `_src_table` FROM orders t0"
	tq := assertSQL(t, db, req, want)
	if len(tq.Params) != 0 {
		t.Fatalf("空 IN 不应产生参数，得到 %v", tq.Params)
	}
}
