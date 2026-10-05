package sqlgen

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// selectOnlyRe 判定语句是否以 SELECT / WITH 开头。
	selectOnlyRe = regexp.MustCompile(`(?is)^\s*(SELECT|WITH)\b`)
	// blacklistRe 匹配写操作 / DDL / 危险关键字。
	blacklistRe = regexp.MustCompile(`(?i)\b(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|CREATE|REPLACE|GRANT|REVOKE|MERGE|CALL|EXEC|EXECUTE|ATTACH|DETACH|PRAGMA|VACUUM|RENAME|OUTFILE|DUMPFILE|HANDLER|LOCK|UNLOCK)\b`)
	// tableRefRe 提取 FROM/JOIN 后的表标识符（可选被反引号/双引号/方括号包裹）。
	tableRefRe = regexp.MustCompile("(?i)(?:FROM|JOIN)\\s+[`\"\\[]?(\\w+)[`\"\\]]?")
	// limitRe 匹配 LIMIT 子句：LIMIT n 或 LIMIT offset, count。
	limitRe = regexp.MustCompile(`(?i)\bLIMIT\s+(\d+)(\s*,\s*(\d+))?`)
	// hasLimitRe 判定是否已存在 LIMIT。
	hasLimitRe = regexp.MustCompile(`(?i)\bLIMIT\b`)
)

// Validate 对 LLM 生成的 SQL 步骤做轻量安全校验，并强制施加 LIMIT 上限。
//
// 返回两个值：
//   - []ValidationError：未通过硬校验（非 SELECT / 命中黑名单 / 多语句 / 引用未知表）的步骤；
//   - []SQLStep：通过校验并规整（强制 LIMIT）后的步骤，未通过的步骤会被剔除。
//
// 校验逐条独立进行；allowedTables 为空（nil 或长度 0）时跳过表名白名单校验。
func Validate(steps []SQLStep, allowedTables map[string]bool, cfg *SQLGenConfig) ([]ValidationError, []SQLStep) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	cfg = cfg.Normalize()

	errs := make([]ValidationError, 0)
	fixed := make([]SQLStep, 0, len(steps))

	for i, step := range steps {
		sql := strings.TrimSpace(step.SQL)
		if reason := hardCheck(sql, allowedTables); reason != "" {
			errs = append(errs, ValidationError{StepIndex: i, SQL: step.SQL, Reason: reason})
			continue
		}
		step.SQL = enforceLimit(sql, cfg)
		fixed = append(fixed, step)
	}
	return errs, fixed
}

// hardCheck 执行不可修复的硬性校验，返回空串表示通过，否则返回失败原因。
func hardCheck(sql string, allowedTables map[string]bool) string {
	if sql == "" {
		return "SQL 为空"
	}
	// 1. 多语句检测：去除结尾分号后不应再出现分号。
	body := strings.TrimRight(sql, "; \t\r\n")
	if strings.Contains(body, ";") {
		return "禁止一次执行多条语句（检测到分号分隔的多语句）"
	}
	// 2. SELECT-only 检测。
	if !selectOnlyRe.MatchString(body) {
		return "仅允许 SELECT / WITH 查询语句"
	}
	// 3. 关键字黑名单。
	if m := blacklistRe.FindString(body); m != "" {
		return "命中禁止的关键字: " + strings.ToUpper(m)
	}
	// 4. 表名白名单。
	if len(allowedTables) > 0 {
		for _, tbl := range extractTables(body) {
			if !allowedTables[strings.ToLower(tbl)] {
				return "引用了 Schema 之外的未知表: " + tbl
			}
		}
	}
	return ""
}

// extractTables 提取 SQL 中 FROM/JOIN 引用的表名（去重，保持出现顺序）。
func extractTables(sql string) []string {
	matches := tableRefRe.FindAllStringSubmatch(sql, -1)
	seen := map[string]bool{}
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		name := strings.TrimSpace(m[1])
		if name == "" {
			continue
		}
		lower := strings.ToLower(name)
		// 过滤 SQL 关键字误伤（如 FROM 后紧跟子查询别名等场景已被 \w+ 限制）。
		if seen[lower] {
			continue
		}
		seen[lower] = true
		out = append(out, name)
	}
	return out
}

// enforceLimit 保证 SQL 带 LIMIT 且不超过上限：无 LIMIT 则追加默认值，超限则收敛。
func enforceLimit(sql string, cfg *SQLGenConfig) string {
	body := strings.TrimRight(strings.TrimSpace(sql), "; \t\r\n")
	if !hasLimitRe.MatchString(body) {
		return body + " LIMIT " + strconv.Itoa(cfg.DefaultLimit)
	}
	return clampLimit(body, cfg)
}

// clampLimit 若已有 LIMIT 的行数超过上限，则将其收敛为 MaxLimit（保留 offset 部分与 OFFSET 子句）。
func clampLimit(sql string, cfg *SQLGenConfig) string {
	idx := limitRe.FindStringIndex(sql)
	if idx == nil {
		return sql
	}
	m := limitRe.FindStringSubmatch(sql)
	if m == nil {
		return sql
	}
	// m[1]=第一个数字；m[3] 非空表示 MySQL 的 "LIMIT offset, count" 形式，行数取 m[3]。
	countStr := m[1]
	offsetCountForm := m[3] != ""
	if offsetCountForm {
		countStr = m[3]
	}
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= cfg.MaxLimit {
		return sql
	}
	var replacement string
	if offsetCountForm {
		replacement = "LIMIT " + m[1] + ", " + strconv.Itoa(cfg.MaxLimit)
	} else {
		replacement = "LIMIT " + strconv.Itoa(cfg.MaxLimit)
	}
	return sql[:idx[0]] + replacement + sql[idx[1]:]
}
