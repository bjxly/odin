package connector

import (
	"database/sql"
	"strconv"
	"strings"
	"time"
)

// rebindPlaceholders 把 translator 统一生成的 "?" 占位符按出现顺序重绑为
// PostgreSQL 风格的 "$1, $2, ..."。translator/SQLBuilder 一律使用 "?"（见
// query/builder.go），MySQL、SQLite 原生支持，但 PostgreSQL 使用 $N 且把裸 "?"
// 当作 jsonb「存在键」操作符，直接执行会报语法错误。
//
// 为避免误伤字符串字面量中的 "?"（如规则注入的裸 WHERE 片段可能内联带 "?" 的文本），
// 扫描时跟踪单引号状态：仅重绑引号外的 "?"；单引号内以 ” 表示转义的单引号。
func rebindPlaceholders(query string) string {
	if !strings.ContainsRune(query, '?') {
		return query
	}
	var b strings.Builder
	b.Grow(len(query) + 16)
	inString := false
	n := 0
	for i := 0; i < len(query); i++ {
		ch := query[i]
		switch {
		case ch == '\'':
			// 处理 '' 转义：字符串内相邻两个单引号表示一个字面单引号。
			if inString && i+1 < len(query) && query[i+1] == '\'' {
				b.WriteByte(ch)
				b.WriteByte(query[i+1])
				i++
				continue
			}
			inString = !inString
			b.WriteByte(ch)
		case ch == '?' && !inString:
			n++
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(n))
		default:
			b.WriteByte(ch)
		}
	}
	return b.String()
}

// scanRows 将 *sql.Rows 扫描为 QueryResult，字段值统一转换为 JSON 友好类型。
// start 用于计算查询耗时。
func scanRows(rows *sql.Rows, start time.Time) (*QueryResult, error) {
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := &QueryResult{
		Columns: cols,
		Rows:    make([][]interface{}, 0),
	}

	for rows.Next() {
		// 使用 interface{} 承接任意类型
		values := make([]interface{}, len(cols))
		pointers := make([]interface{}, len(cols))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}
		row := make([]interface{}, len(cols))
		for i, v := range values {
			row[i] = normalizeValue(v)
		}
		result.Rows = append(result.Rows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result.RowCount = len(result.Rows)
	result.Duration = time.Since(start).Milliseconds()
	return result, nil
}

// normalizeValue 将驱动返回的原始值转换为便于 JSON 序列化的类型。
func normalizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case []byte:
		return string(val)
	case time.Time:
		return val.Format(time.RFC3339Nano)
	default:
		return val
	}
}

// poolConfigFrom 从连接配置构造连接池参数（供各连接器复用）。
func poolDefaults(maxOpen, maxIdle, maxLifetime int) (int, int, int) {
	if maxOpen <= 0 {
		maxOpen = 10
	}
	if maxIdle <= 0 {
		maxIdle = 5
	}
	if maxLifetime <= 0 {
		maxLifetime = 3600
	}
	return maxOpen, maxIdle, maxLifetime
}
