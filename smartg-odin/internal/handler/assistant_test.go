package handler

import (
	"encoding/json"
	"testing"
)

// ---------------------------------------------------------------------------
// D2：meta.hop_count 契约防御单测 —— 恒为单一整数，绝不为数组
// ---------------------------------------------------------------------------

func TestAsHopCount(t *testing.T) {
	cases := []struct {
		name string
		in   interface{}
		want int
	}{
		{"int", 4, 4},
		{"int64", int64(3), 3},
		{"float64", float64(5), 5},
		{"json.Number", json.Number("7"), 7},
		{"字符串整数", "6", 6},
		{"字符串浮点", "2.9", 2},
		{"空字符串", "", 0},
		{"nil", nil, 0},
		{"负数钳制", -3, 0},
		{"数组取末位（D2 病灶）", []interface{}{float64(3), float64(4)}, 4},
		{"空数组", []interface{}{}, 0},
		{"[]int 取末位", []int{3, 4}, 4},
		{"嵌套数组", []interface{}{[]interface{}{float64(1), float64(2)}}, 2},
		{"不可解析对象", struct{}{}, 0},
		{"map", map[string]interface{}{"a": 1}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := asHopCount(c.in); got != c.want {
				t.Fatalf("asHopCount(%#v) = %d，期望 %d", c.in, got, c.want)
			}
		})
	}
}

// TestNormalizeMessageMeta 历史消息 meta 中的数组型 hop_count 必须被归一为整数（旧数据防御）。
func TestNormalizeMessageMeta(t *testing.T) {
	in := map[string]interface{}{
		"hop_count":        []interface{}{float64(3), float64(4)},
		"llm_calls":        float64(1),
		"sub_intent_count": "0",
		"parse_path":       "p2_traverse",
	}
	got := normalizeMessageMeta(in).(map[string]interface{})
	if got["hop_count"] != 4 {
		t.Errorf("hop_count 应归一为 4，实得 %#v", got["hop_count"])
	}
	if got["llm_calls"] != 1 {
		t.Errorf("llm_calls 应归一为 1，实得 %#v", got["llm_calls"])
	}
	if got["sub_intent_count"] != 0 {
		t.Errorf("sub_intent_count 应归一为 0，实得 %#v", got["sub_intent_count"])
	}
	if got["parse_path"] != "p2_traverse" {
		t.Errorf("非数值字段不得被改写，实得 %#v", got["parse_path"])
	}
	// 非 map 结构原样返回
	if normalizeMessageMeta(nil) != nil {
		t.Errorf("nil meta 应原样返回")
	}
	if s := normalizeMessageMeta("x"); s != "x" {
		t.Errorf("非 map meta 应原样返回，实得 %#v", s)
	}
}
