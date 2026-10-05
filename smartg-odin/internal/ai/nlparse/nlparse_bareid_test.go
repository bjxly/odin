package nlparse

import (
	"strings"
	"testing"

	"smartg-odin/internal/ds/query"
)

// orderScope 构造一个仅含元数据的 Order 类视图（无需 DB），用于直接测试 parseFilters 的裸标识符分支。
// 属性口径与 seed_ontology.go 一致：orderNo（Label=订单号）应被识别为键属性。
func orderScope() *ontClass {
	return &ontClass{
		ID:    1,
		Name:  "Order",
		Label: "订单",
		Properties: []ontProp{
			{Name: "orderNo", Label: "订单号", DataType: "string", Synonyms: []string{"订单编号", "单号"}},
			{Name: "quantity", Label: "数量", DataType: "integer", Synonyms: []string{"销量", "销售数量"}},
			{Name: "amount", Label: "金额", DataType: "float", Synonyms: []string{"销售额", "总金额"}},
			{Name: "status", Label: "状态", DataType: "string"},
		},
	}
}

func newResult() *Result {
	return &Result{Query: &query.QueryRequest{}, Notes: []string{}, Hits: []string{}}
}

// findFilter 返回 res 中首个匹配 (property, op) 的过滤，未找到返回 nil。
func findFilter(res *Result, prop, op string) *query.Filter {
	for i := range res.Query.Filters {
		f := &res.Query.Filters[i]
		if f.Property == prop && f.Op == op {
			return f
		}
	}
	return nil
}

// TestParseFiltersBareEntityIdentifier 断言「订单 ORD-20250115-007 为何延误」在缺少比较词时，
// 仍能通过裸实体标识符分支产出 orderNo eq ORD-20250115-007 过滤（治「15 行不相关」）。
func TestParseFiltersBareEntityIdentifier(t *testing.T) {
	scope := orderScope()
	raw := "订单 ORD-20250115-007 为何延误"
	res := newResult()

	parseFilters(raw, strings.ToLower(raw), scope, res)

	f := findFilter(res, "orderNo", "eq")
	if f == nil {
		t.Fatalf("期望产出 orderNo eq 过滤，实际 filters=%+v", res.Query.Filters)
	}
	if got, _ := f.Value.(string); got != "ORD-20250115-007" {
		t.Fatalf("过滤值不符：期望 ORD-20250115-007，实际 %v", f.Value)
	}
}

// TestParseFiltersBareEntityNoDuplicate 断言当三段式已抽到 orderNo 过滤时，裸标识符分支不重复注入。
func TestParseFiltersBareEntityNoDuplicate(t *testing.T) {
	scope := orderScope()
	raw := "订单号 等于 ORD-20250115-007 的状态"
	res := newResult()

	parseFilters(raw, strings.ToLower(raw), scope, res)

	count := 0
	for _, fl := range res.Query.Filters {
		if fl.Property == "orderNo" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("orderNo 过滤应恰好 1 条（避免重复注入），实际 %d 条：%+v", count, res.Query.Filters)
	}
}

// TestParseFiltersBareEntityConservative 断言不含数字的普通连字符词（state-of-the-art）不会被误判为实体标识符。
func TestParseFiltersBareEntityConservative(t *testing.T) {
	scope := orderScope()
	raw := "查询 state-of-the-art 的订单"
	res := newResult()

	parseFilters(raw, strings.ToLower(raw), scope, res)

	if f := findFilter(res, "orderNo", "eq"); f != nil {
		t.Fatalf("普通连字符词不应触发实体键过滤，实际注入 %+v", *f)
	}
}

// TestContainsEntityIdentifier 校验实体标识符探测（供 intent.go 降级判定复用）。
func TestContainsEntityIdentifier(t *testing.T) {
	cases := []struct {
		text string
		want bool
	}{
		{"订单 ORD-20250115-007 为何延误", true},
		{"SKU-12345 的库存", true},
		{"查询 VIP 客户", false},
		{"state-of-the-art 技术", false}, // 无数字，保守不识别
		{"", false},
	}
	for _, c := range cases {
		if got := ContainsEntityIdentifier(c.text); got != c.want {
			t.Errorf("ContainsEntityIdentifier(%q)=%v，期望 %v", c.text, got, c.want)
		}
	}
}
