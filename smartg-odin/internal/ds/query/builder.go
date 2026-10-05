// Package query 实现本体概念到物理 SQL 的翻译与执行。
//
// 本文件提供一个链式 SQL 构建器 SQLBuilder，负责把 SELECT / FROM / JOIN /
// WHERE / GROUP BY / HAVING / ORDER BY / LIMIT / OFFSET 各子句拼装成完整的
// 参数化 SQL 语句。占位符统一使用 "?"，由底层 database/sql 驱动做参数绑定，
// 从而避免 SQL 注入。
package query

import (
	"fmt"
	"strings"
)

// joinClause 描述一个 JOIN 子句。
type joinClause struct {
	joinType  string // LEFT, INNER, RIGHT
	table     string
	alias     string
	condition string
}

// whereClause 描述一个 WHERE 条件及其绑定参数。
type whereClause struct {
	expr   string
	params []interface{}
}

// SQLBuilder 链式 SQL 构建器。
type SQLBuilder struct {
	selectCols []string
	fromTable  string
	fromAlias  string
	joins      []joinClause
	wheres     []whereClause
	groupBys   []string
	havings    []whereClause
	orderBys   []string
	limit      int
	offset     int
	params     []interface{}
	paramIndex int
}

// NewBuilder 创建一个新的 SQLBuilder，limit 默认 -1（不限制）。
func NewBuilder() *SQLBuilder {
	return &SQLBuilder{limit: -1, offset: -1}
}

// Select 追加要查询的列（可以是 "t0.col AS alias" 形式）。
func (b *SQLBuilder) Select(cols ...string) *SQLBuilder {
	b.selectCols = append(b.selectCols, cols...)
	return b
}

// From 设置主表及别名。
func (b *SQLBuilder) From(table, alias string) *SQLBuilder {
	b.fromTable = table
	b.fromAlias = alias
	return b
}

// Join 追加一个 JOIN 子句。joinType 为空时默认 LEFT。
func (b *SQLBuilder) Join(joinType, table, alias, condition string) *SQLBuilder {
	if strings.TrimSpace(joinType) == "" {
		joinType = "LEFT"
	}
	b.joins = append(b.joins, joinClause{
		joinType:  strings.ToUpper(strings.TrimSpace(joinType)),
		table:     table,
		alias:     alias,
		condition: condition,
	})
	return b
}

// Where 追加一个 WHERE 条件（多个条件之间以 AND 连接），params 为绑定参数。
func (b *SQLBuilder) Where(expr string, params ...interface{}) *SQLBuilder {
	if strings.TrimSpace(expr) == "" {
		return b
	}
	b.wheres = append(b.wheres, whereClause{expr: expr, params: params})
	return b
}

// GroupBy 追加 GROUP BY 列。
func (b *SQLBuilder) GroupBy(cols ...string) *SQLBuilder {
	b.groupBys = append(b.groupBys, cols...)
	return b
}

// Having 追加 HAVING 条件，params 为其绑定参数。
// HAVING 参数在 Build 时于 WHERE 参数之后追加，严格匹配子句发射顺序，
// 避免 PostgreSQL 的 $N 占位符重绑定发生错位（见 connector.rebindPlaceholders）。
func (b *SQLBuilder) Having(expr string, params ...interface{}) *SQLBuilder {
	if strings.TrimSpace(expr) != "" {
		b.havings = append(b.havings, whereClause{expr: expr, params: params})
	}
	return b
}

// OrderBy 追加排序列，dir 为空时默认 ASC。
func (b *SQLBuilder) OrderBy(col, dir string) *SQLBuilder {
	if strings.TrimSpace(col) == "" {
		return b
	}
	if strings.TrimSpace(dir) == "" {
		dir = "ASC"
	}
	b.orderBys = append(b.orderBys, fmt.Sprintf("%s %s", col, strings.ToUpper(strings.TrimSpace(dir))))
	return b
}

// Limit 设置返回行数上限，n < 0 表示不限制。
func (b *SQLBuilder) Limit(n int) *SQLBuilder {
	b.limit = n
	return b
}

// Offset 设置偏移量，n < 0 表示不偏移。
func (b *SQLBuilder) Offset(n int) *SQLBuilder {
	b.offset = n
	return b
}

// Build 拼装完整 SQL 并返回语句与按序绑定的参数列表。
func (b *SQLBuilder) Build() (string, []interface{}) {
	var sb strings.Builder
	params := make([]interface{}, 0)

	// SELECT
	sb.WriteString("SELECT ")
	if len(b.selectCols) == 0 {
		sb.WriteString("*")
	} else {
		sb.WriteString(strings.Join(b.selectCols, ", "))
	}

	// FROM
	if b.fromTable != "" {
		sb.WriteString(" FROM ")
		sb.WriteString(b.fromTable)
		if b.fromAlias != "" {
			sb.WriteString(" " + b.fromAlias)
		}
	}

	// JOIN
	for _, j := range b.joins {
		sb.WriteString(fmt.Sprintf(" %s JOIN %s", j.joinType, j.table))
		if j.alias != "" {
			sb.WriteString(" " + j.alias)
		}
		if j.condition != "" {
			sb.WriteString(" ON " + j.condition)
		}
	}

	// WHERE
	if len(b.wheres) > 0 {
		exprs := make([]string, 0, len(b.wheres))
		for _, w := range b.wheres {
			exprs = append(exprs, w.expr)
			params = append(params, w.params...)
		}
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(exprs, " AND "))
	}

	// GROUP BY
	if len(b.groupBys) > 0 {
		sb.WriteString(" GROUP BY " + strings.Join(b.groupBys, ", "))
	}

	// HAVING（参数在 WHERE 参数之后追加，保持与子句发射顺序一致）
	if len(b.havings) > 0 {
		exprs := make([]string, 0, len(b.havings))
		for _, h := range b.havings {
			exprs = append(exprs, h.expr)
			params = append(params, h.params...)
		}
		sb.WriteString(" HAVING " + strings.Join(exprs, " AND "))
	}

	// ORDER BY
	if len(b.orderBys) > 0 {
		sb.WriteString(" ORDER BY " + strings.Join(b.orderBys, ", "))
	}

	// LIMIT / OFFSET
	if b.limit >= 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", b.limit))
	}
	if b.offset > 0 {
		sb.WriteString(fmt.Sprintf(" OFFSET %d", b.offset))
	}

	b.params = params
	b.paramIndex = len(params)
	return sb.String(), params
}

// Params 返回最近一次 Build 生成的参数列表。
func (b *SQLBuilder) Params() []interface{} { return b.params }

// ParamCount 返回最近一次 Build 绑定的参数个数。
func (b *SQLBuilder) ParamCount() int { return b.paramIndex }
