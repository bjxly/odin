package sqlgen

import (
	"fmt"
	"strconv"
	"strings"
)

// Merge 按指定策略合并多步执行结果。
//
// 策略：
//   - single：直接透传第一个（无错误的）结果；
//   - join：以第一个结果集为左表，按 keys 做内存 LEFT JOIN，合并后续结果集的列；
//   - side_by_side：保留各步独立结果（顶层列/行取第一步，完整数据在 Steps 中）；
//   - synthesize：同 side_by_side，由上层 handler 负责二次 LLM 综合；
//   - 其它/空：按 single 处理。
func Merge(results []StepResult, strategy string, keys []string) (*MergedResult, error) {
	strategy = strings.ToLower(strings.TrimSpace(strategy))
	if strategy == "" {
		strategy = "single"
	}
	switch strategy {
	case "join":
		return mergeJoin(results, keys), nil
	case "side_by_side", "synthesize":
		return mergeSideBySide(results, strategy), nil
	case "single":
		return mergeSingle(results, "single"), nil
	default:
		return mergeSingle(results, strategy), nil
	}
}

// validResults 过滤掉执行出错或无列的结果。
func validResults(results []StepResult) []StepResult {
	out := make([]StepResult, 0, len(results))
	for _, r := range results {
		if strings.TrimSpace(r.Error) != "" {
			continue
		}
		out = append(out, r)
	}
	return out
}

// mergeSingle 透传第一个可用结果；无可用结果时返回空壳。
func mergeSingle(results []StepResult, strategy string) *MergedResult {
	valid := validResults(results)
	m := &MergedResult{
		Columns:  []ColumnInfo{},
		Rows:     [][]interface{}{},
		Steps:    results,
		Strategy: strategy,
	}
	if len(valid) == 0 {
		return m
	}
	first := valid[0]
	m.Columns = first.Columns
	m.Rows = first.Rows
	m.RowCount = first.RowCount
	return m
}

// mergeSideBySide 保留各步独立结果，顶层列/行取第一步以便直接渲染。
func mergeSideBySide(results []StepResult, strategy string) *MergedResult {
	valid := validResults(results)
	m := &MergedResult{
		Columns:  []ColumnInfo{},
		Rows:     [][]interface{}{},
		Steps:    results,
		Strategy: strategy,
	}
	if len(valid) > 0 {
		m.Columns = valid[0].Columns
		m.Rows = valid[0].Rows
		m.RowCount = valid[0].RowCount
	}
	return m
}

// mergeJoin 以第一个结果集为左表，按 keys 依次对后续结果集做内存 LEFT JOIN。
//
// 对每个右表：若 keys 在左右两侧均可解析，则按 key 值匹配右表行（一对多会展开为多行，
// 无匹配则以 nil 填充右表列）；若 keys 无法解析则跳过该右表。右表与左表同名的 key 列不重复加入。
func mergeJoin(results []StepResult, keys []string) *MergedResult {
	valid := validResults(results)
	if len(valid) == 0 {
		return &MergedResult{Columns: []ColumnInfo{}, Rows: [][]interface{}{}, Steps: results, Strategy: "join"}
	}
	if len(valid) == 1 || len(keys) == 0 {
		return mergeSingle(results, "join")
	}

	left := valid[0]
	cols := make([]ColumnInfo, len(left.Columns))
	copy(cols, left.Columns)
	rows := make([][]interface{}, len(left.Rows))
	for i, r := range left.Rows {
		nr := make([]interface{}, len(r))
		copy(nr, r)
		rows[i] = nr
	}

	for _, right := range valid[1:] {
		leftKeyIdx := keyIndices(cols, keys)
		rightKeyIdx := keyIndices(right.Columns, keys)
		if len(leftKeyIdx) != len(keys) || len(rightKeyIdx) != len(keys) {
			// key 无法在两侧完全解析，跳过该右表。
			continue
		}
		// 建立右表 key → 行索引列表。
		index := map[string][]int{}
		for ri, rrow := range right.Rows {
			k := compositeKey(rrow, rightKeyIdx)
			index[k] = append(index[k], ri)
		}
		// 右表待追加的列（排除 key 列，避免重复）。
		appendIdx := make([]int, 0, len(right.Columns))
		isKey := map[int]bool{}
		for _, ki := range rightKeyIdx {
			isKey[ki] = true
		}
		for ci := range right.Columns {
			if !isKey[ci] {
				appendIdx = append(appendIdx, ci)
			}
		}
		newCols := make([]ColumnInfo, 0, len(cols)+len(appendIdx))
		newCols = append(newCols, cols...)
		for _, ci := range appendIdx {
			newCols = append(newCols, right.Columns[ci])
		}

		newRows := make([][]interface{}, 0, len(rows))
		for _, lrow := range rows {
			k := compositeKey(lrow, leftKeyIdx)
			matches := index[k]
			if len(matches) == 0 {
				combined := make([]interface{}, len(cols)+len(appendIdx))
				copy(combined, lrow)
				// 右表列填 nil。
				newRows = append(newRows, combined)
				continue
			}
			for _, ri := range matches {
				rrow := right.Rows[ri]
				combined := make([]interface{}, 0, len(cols)+len(appendIdx))
				combined = append(combined, lrow...)
				for _, ci := range appendIdx {
					if ci < len(rrow) {
						combined = append(combined, rrow[ci])
					} else {
						combined = append(combined, nil)
					}
				}
				newRows = append(newRows, combined)
			}
		}
		cols = newCols
		rows = newRows
	}

	return &MergedResult{
		Columns:  cols,
		Rows:     rows,
		RowCount: len(rows),
		Steps:    results,
		Strategy: "join",
	}
}

// keyIndices 按列名（大小写不敏感）解析 keys 在给定列集合中的下标；
// 仅当全部 key 均命中时返回完整下标切片，否则返回不完整切片供调用方判定。
func keyIndices(cols []ColumnInfo, keys []string) []int {
	idx := make([]int, 0, len(keys))
	for _, k := range keys {
		found := -1
		for ci, c := range cols {
			if strings.EqualFold(strings.TrimSpace(c.Name), strings.TrimSpace(k)) {
				found = ci
				break
			}
		}
		if found < 0 {
			return idx // 不完整
		}
		idx = append(idx, found)
	}
	return idx
}

// compositeKey 依据给定列下标拼接行的复合键字符串。
func compositeKey(row []interface{}, idx []int) string {
	parts := make([]string, 0, len(idx))
	for _, i := range idx {
		if i < len(row) {
			parts = append(parts, fmt.Sprintf("%v", row[i]))
		} else {
			parts = append(parts, "")
		}
	}
	return strings.Join(parts, "\x00")
}

// ChartSuggestion 根据合并结果自动建议图表类型；无法判断时返回 nil。
//
// 规则（迁移自 skills.chartSuggestion 的思路）：
//   - 单行结果：把所有数值列作为度量 → chart_type=metric（标量/多指标卡）；
//   - 多行结果：首列为维度、其余数值列为度量；无数值度量则返回 nil；
//     单维度单度量时按维度语义选择 line（时间）/ pie（类别且行数≤6）/ bar，其余为 bar。
func ChartSuggestion(result *MergedResult) *ChartPayload {
	if result == nil || len(result.Columns) == 0 || len(result.Rows) == 0 {
		return nil
	}
	names := make([]string, 0, len(result.Columns))
	for _, c := range result.Columns {
		names = append(names, c.Name)
	}

	// 单行：全部数值列作为度量 → metric。
	if len(result.Rows) == 1 {
		metrics := make([]ChartField, 0)
		for i, c := range result.Columns {
			if isNumericColumn(result.Rows, i) {
				metrics = append(metrics, ChartField{Field: c.Name, Label: labelOf(c.Name)})
			}
		}
		if len(metrics) == 0 {
			return nil
		}
		return &ChartPayload{
			ChartType:  "metric",
			Dimensions: []ChartField{},
			Metrics:    metrics,
			Rows:       result.Rows,
			Columns:    names,
		}
	}

	// 多行：首列维度，其余数值列度量。
	dims := []ChartField{{Field: result.Columns[0].Name, Label: labelOf(result.Columns[0].Name)}}
	metrics := make([]ChartField, 0)
	for i := 1; i < len(result.Columns); i++ {
		if isNumericColumn(result.Rows, i) {
			metrics = append(metrics, ChartField{Field: result.Columns[i].Name, Label: labelOf(result.Columns[i].Name)})
		}
	}
	if len(metrics) == 0 {
		return nil
	}

	chartType := "bar"
	if len(dims) == 1 && len(metrics) == 1 {
		switch {
		case isTimeDimension(result.Columns[0].Name, result.Rows):
			chartType = "line"
		case len(result.Rows) <= 6:
			chartType = "pie"
		default:
			chartType = "bar"
		}
	}

	return &ChartPayload{
		ChartType:  chartType,
		Dimensions: dims,
		Metrics:    metrics,
		Rows:       result.Rows,
		Columns:    names,
	}
}

// isNumericColumn 采样前若干行判断指定列是否为数值列。
//
// 数值判定需兼容驱动返回的字符串/字节形式：MySQL 的 DECIMAL（如 SUM(amount)）
// 经 connector.normalizeValue 会把 []byte 转为 string（如 "391500.00"），
// 若仅按原生数值类型判定会漏掉金额类度量，导致多度量结果只保留 COUNT 一列。
// 因此这里对 string/[]byte 也尝试解析为 float64，解析成功即视为数值。
func isNumericColumn(rows [][]interface{}, colIdx int) bool {
	sampled := 0
	for _, row := range rows {
		if sampled >= 5 {
			break
		}
		if colIdx >= len(row) {
			continue
		}
		v := row[colIdx]
		if v == nil {
			continue
		}
		sampled++
		if !isNumericValue(v) {
			return false
		}
	}
	return sampled > 0
}

// isNumericValue 判断单个值是否为数值：原生数值类型直接通过，
// string/[]byte 形式去除千分位逗号后尝试解析为 float64。
func isNumericValue(v interface{}) bool {
	switch val := v.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	case string:
		return parseNumericString(val)
	case []byte:
		return parseNumericString(string(val))
	default:
		return false
	}
}

// parseNumericString 去除首尾空白与千分位逗号后尝试解析为 float64。
func parseNumericString(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	s = strings.ReplaceAll(s, ",", "")
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// isTimeDimension 依据列名关键字或首行值类型判断维度是否为时间序列。
func isTimeDimension(name string, rows [][]interface{}) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	for _, kw := range []string{"date", "time", "day", "month", "year", "日期", "时间", "年", "月", "季度"} {
		if strings.Contains(n, kw) || strings.Contains(name, kw) {
			return true
		}
	}
	return false
}

// labelOf 返回字段的展示标签（当前直接复用列名，列名通常已是中文别名）。
func labelOf(name string) string { return name }
