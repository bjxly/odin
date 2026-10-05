// Package handler 汇集 ODIN 全部 REST API 处理函数。
//
// 约定：
//   - 所有处理函数签名统一为 func Xxx(c *gin.Context)，通过 database.GetDB() 获取 GORM 实例；
//   - 成功响应统一走 middleware.Success / SuccessPage，HTTP 状态码 200；
//   - 业务错误统一走 middleware.Error（HTTP 200 + 非 0 业务码），仅健康探针等基础设施
//     场景使用 ErrorWithHTTP 返回非 200；
//   - URL 中的 :id 一律解析为 uint 自增主键。
package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"smartg-odin/internal/common/errs"
	"smartg-odin/internal/common/logger"
	"smartg-odin/internal/database"
	"smartg-odin/internal/ds/connector"
	"smartg-odin/internal/middleware"
	"smartg-odin/internal/model"
)

// Version 服务版本号，健康检查与启动信息使用。
const Version = "1.0.0"

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 500
	// connectorTimeout 连接测试 / 内省的默认超时时间。
	connectorTimeout = 20 * time.Second
	// queryTimeout 联邦查询的默认超时时间。
	queryTimeout = 60 * time.Second
)

// ---------------------------------------------------------------------------
// 响应与错误
// ---------------------------------------------------------------------------

// DB 返回全局 GORM 实例。
func DB() *gorm.DB { return database.GetDB() }

// ok 返回成功响应。
func ok(c *gin.Context, data interface{}) { middleware.Success(c, data) }

// okMsg 返回带自定义提示语的成功响应。
func okMsg(c *gin.Context, data interface{}, msg string) {
	middleware.SuccessWithMessage(c, data, msg)
}

// fail 返回业务错误响应（HTTP 200 + 业务错误码）。
func fail(c *gin.Context, appErr *errs.AppError) { middleware.Error(c, appErr) }

// mapError 将 GORM 查询错误映射为业务错误。
func mapError(err error, notFoundMsg string) *errs.AppError {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.NotFound(notFoundMsg)
	}
	return errs.DBError(err)
}

// ---------------------------------------------------------------------------
// 参数解析
// ---------------------------------------------------------------------------

// parseIDParam 解析路径中的自增主键参数。
func parseIDParam(c *gin.Context, key string) (uint, *errs.AppError) {
	raw := strings.TrimSpace(c.Param(key))
	if raw == "" {
		return 0, errs.ParamError("缺少路径参数: " + key)
	}
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || v == 0 {
		return 0, errs.ParamError("非法的 ID: " + raw)
	}
	if v > math.MaxUint32 {
		return 0, errs.ParamError("ID 超出取值范围: " + raw)
	}
	return uint(v), nil
}

// parsePage 解析分页参数 current / size（兼容 page / pageSize）。
func parsePage(c *gin.Context) (current, size int) {
	current = parseIntQuery(c, "current", 0)
	if current <= 0 {
		current = parseIntQuery(c, "page", defaultPage)
	}
	size = parseIntQuery(c, "size", 0)
	if size <= 0 {
		size = parseIntQuery(c, "pageSize", defaultPageSize)
	}
	if current < 1 {
		current = defaultPage
	}
	if size < 1 {
		size = defaultPageSize
	}
	if size > maxPageSize {
		size = maxPageSize
	}
	return current, size
}

// parseIntQuery 读取整型查询参数，非法或缺省时返回 def。
func parseIntQuery(c *gin.Context, key string, def int) int {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return v
}

// uintQuery 读取可选的 uint 查询参数，缺省或非法返回 0。
func uintQuery(c *gin.Context, key string) uint {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return 0
	}
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0
	}
	return uint(v)
}

// bindJSON 绑定 JSON 请求体，失败返回参数错误。
func bindJSON(c *gin.Context, v interface{}) *errs.AppError {
	if err := c.ShouldBindJSON(v); err != nil {
		return errs.ParamError("请求体解析失败: " + err.Error())
	}
	return nil
}

// bindNormalized 读取请求体并将顶层键名统一为 snake_case 后绑定到目标结构体，
// 以便同时兼容 camelCase（旧前端）与 snake_case（新模型 json tag）两种写法。
// 嵌套结构不做递归转换，避免破坏用户自定义字典（如 value_map）的键。
func bindNormalized(c *gin.Context, v interface{}) *errs.AppError {
	if c.Request.Body == nil {
		return errs.ParamError("请求体为空")
	}
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return errs.ParamError("读取请求体失败: " + err.Error())
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return errs.ParamError("请求体为空")
	}

	var generic interface{}
	if err := json.Unmarshal(raw, &generic); err != nil {
		// 非对象结构（如数组）直接按原样绑定
		if err2 := json.Unmarshal(raw, v); err2 != nil {
			return errs.ParamError("请求体解析失败: " + err2.Error())
		}
		return nil
	}

	normalized := normalizeKeys(generic)
	buf, err := json.Marshal(normalized)
	if err != nil {
		return errs.ParamError("请求体规范化失败: " + err.Error())
	}
	if err := json.Unmarshal(buf, v); err != nil {
		return errs.ParamError("请求体解析失败: " + err.Error())
	}
	return nil
}

// normalizeKeys 将顶层 map 的键名转换为 snake_case。
func normalizeKeys(v interface{}) interface{} {
	m, isMap := v.(map[string]interface{})
	if !isMap {
		return v
	}
	out := make(map[string]interface{}, len(m))
	for k, val := range m {
		out[camelToSnake(k)] = val
	}
	return out
}

// camelToSnake 将 camelCase / PascalCase 转为 snake_case，正确识别连续大写缩写词。
func camelToSnake(s string) string {
	if !strings.ContainsFunc(s, func(r rune) bool { return r >= 'A' && r <= 'Z' }) {
		return strings.ReplaceAll(strings.ReplaceAll(s, "-", "_"), " ", "_")
	}
	runes := []rune(s)
	var b strings.Builder
	for i, r := range runes {
		switch {
		case r >= 'A' && r <= 'Z':
			prevLower := i > 0 && ((runes[i-1] >= 'a' && runes[i-1] <= 'z') || (runes[i-1] >= '0' && runes[i-1] <= '9'))
			nextLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'
			if i > 0 && (prevLower || nextLower) {
				b.WriteRune('_')
			}
			b.WriteRune(r + ('a' - 'A'))
		case r == '-' || r == ' ':
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// bindMap 将请求体解析为通用 map；空请求体返回空 map（便于实现部分更新）。
func bindMap(c *gin.Context) (map[string]interface{}, *errs.AppError) {
	body := map[string]interface{}{}
	if c.Request.Body == nil || c.Request.ContentLength == 0 {
		return body, nil
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		// 允许前端提交空 JSON 或无 body 的 PUT
		if strings.Contains(err.Error(), "EOF") {
			return map[string]interface{}{}, nil
		}
		return nil, errs.ParamError("请求体解析失败: " + err.Error())
	}
	if body == nil {
		body = map[string]interface{}{}
	}
	// 顶层键名统一为 snake_case，兼容 camelCase 请求
	if normalized, isMap := normalizeKeys(body).(map[string]interface{}); isMap {
		body = normalized
	}
	return body, nil
}

// pickUpdates 从请求体中挑选允许更新的列（json tag 与数据库列名一致），
// 同时兼容 camelCase 写法；未提供的字段保持原值，实现 PUT 的部分更新语义。
func pickUpdates(body map[string]interface{}, allowed ...string) map[string]interface{} {
	updates := map[string]interface{}{}
	for _, key := range allowed {
		val, found := body[key]
		if !found {
			if alt := toCamel(key); alt != key {
				val, found = body[alt]
			}
		}
		if !found {
			continue
		}
		updates[key] = normalizeUpdateValue(key, val)
	}
	return updates
}

// normalizeUpdateValue 归一化待写入的字段值：
//   - 无小数部分的 float64 转为 int64，避免整型列被写成实数；
//   - *_json 列若收到对象/数组，自动序列化为字符串。
func normalizeUpdateValue(key string, val interface{}) interface{} {
	switch v := val.(type) {
	case float64:
		if v == math.Trunc(v) && !math.IsInf(v, 0) {
			return int64(v)
		}
		return v
	case nil:
		return nil
	case map[string]interface{}, []interface{}:
		if strings.HasSuffix(key, "_json") {
			if b, err := json.Marshal(v); err == nil {
				return string(b)
			}
		}
		return v
	default:
		return val
	}
}

// toUint 将 JSON 解析得到的值（float64 / 字符串 / 整型）转为 uint。
func toUint(v interface{}) (uint, bool) {
	switch tv := v.(type) {
	case nil:
		return 0, false
	case float64:
		if tv < 0 || tv != math.Trunc(tv) {
			return 0, false
		}
		return uint(tv), true
	case float32:
		return toUint(float64(tv))
	case int:
		if tv < 0 {
			return 0, false
		}
		return uint(tv), true
	case int64:
		if tv < 0 {
			return 0, false
		}
		return uint(tv), true
	case uint:
		return tv, true
	case uint64:
		return uint(tv), true
	case json.Number:
		n, err := tv.Int64()
		if err != nil || n < 0 {
			return 0, false
		}
		return uint(n), true
	default:
		n, err := strconv.ParseUint(strings.TrimSpace(fmt.Sprint(tv)), 10, 64)
		if err != nil {
			return 0, false
		}
		return uint(n), true
	}
}

// toCamel 将 snake_case 转为 camelCase。
func toCamel(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) < 2 {
		return s
	}
	var b strings.Builder
	for i, p := range parts {
		if p == "" {
			continue
		}
		if i == 0 {
			b.WriteString(p)
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]))
		b.WriteString(p[1:])
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// 数据源连接配置
// ---------------------------------------------------------------------------

// connectionConfig 由数据源模型构造连接器所需的连接配置。
func connectionConfig(ds *model.DataSource) *connector.ConnectionConfig {
	cfg := &connector.ConnectionConfig{
		Type:            strings.ToLower(strings.TrimSpace(ds.Type)),
		Host:            ds.Host,
		Port:            ds.Port,
		DatabaseName:    ds.DatabaseName,
		Username:        ds.Username,
		Password:        ds.Password,
		SchemaName:      ds.SchemaName,
		SSLMode:         ds.SSLMode,
		Params:          parseParams(ds.ParamsJSON),
		PoolMaxOpen:     ds.PoolMaxOpen,
		PoolMaxIdle:     ds.PoolMaxIdle,
		PoolMaxLifetime: ds.PoolMaxLifetime,
	}

	// SQLite 以文件路径连接：优先取 params 中的 file/path，其次 database_name。
	if cfg.Type == "sqlite" {
		switch {
		case cfg.Params["file"] != "":
			cfg.FilePath = cfg.Params["file"]
		case cfg.Params["path"] != "":
			cfg.FilePath = cfg.Params["path"]
		case cfg.Params["dsn"] != "":
			cfg.FilePath = cfg.Params["dsn"]
		default:
			cfg.FilePath = ds.DatabaseName
		}
	}
	// 完整 DSN 直连（postgres:// 或 user:pass@tcp(...) 形式）
	if dsn := cfg.Params["dsn"]; dsn != "" && cfg.Type != "sqlite" {
		cfg.DatabaseName = dsn
	}
	return cfg
}

// parseParams 解析 params_json 为字符串字典，兼容 {"k":"v"} 与 URL query 两种写法。
func parseParams(raw string) map[string]string {
	params := map[string]string{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return params
	}
	if strings.HasPrefix(raw, "{") {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &m); err == nil {
			for k, v := range m {
				switch tv := v.(type) {
				case string:
					params[k] = tv
				case float64:
					if tv == math.Trunc(tv) {
						params[k] = strconv.FormatInt(int64(tv), 10)
					} else {
						params[k] = strconv.FormatFloat(tv, 'f', -1, 64)
					}
				case bool:
					params[k] = strconv.FormatBool(tv)
				case nil:
				default:
					params[k] = fmt.Sprint(tv)
				}
			}
			return params
		}
	}
	// 退化为 URL query 形式：a=1&b=2
	for _, pair := range strings.Split(strings.TrimPrefix(raw, "?"), "&") {
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			params[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return params
}

// getConnector 按数据源类型获取连接器实例。
func getConnector(dsType string) (connector.Connector, *errs.AppError) {
	conn, err := connector.Get(strings.ToLower(strings.TrimSpace(dsType)))
	if err != nil {
		return nil, errs.Wrap(errs.CodeConnError,
			"暂不支持的数据源类型: "+dsType+"（已支持: "+strings.Join(connector.SupportedTypes(), ", ")+"）", err)
	}
	return conn, nil
}

// loadDatasource 按主键加载数据源，未找到返回 NotFound。
func loadDatasource(id uint) (*model.DataSource, *errs.AppError) {
	var ds model.DataSource
	if err := DB().First(&ds, id).Error; err != nil {
		return nil, mapError(err, fmt.Sprintf("数据源不存在: %d", id))
	}
	return &ds, nil
}

// ---------------------------------------------------------------------------
// 审计与工具
// ---------------------------------------------------------------------------

// writeAudit 写入操作审计日志；失败仅记录告警，不影响主流程。
func writeAudit(c *gin.Context, action, resourceType string, resourceID interface{}, detail interface{}) {
	db := DB()
	if db == nil {
		return
	}
	entry := model.AuditLog{
		UserID:       firstNonEmpty(c.GetHeader("X-User-Id"), c.GetHeader("X-User"), "anonymous"),
		Action:       action,
		ResourceType: resourceType,
		IPAddress:    c.ClientIP(),
		CreatedAt:    time.Now(),
	}
	if resourceID != nil {
		entry.ResourceID = fmt.Sprint(resourceID)
	}
	if detail != nil {
		if b, err := json.Marshal(detail); err == nil {
			entry.DetailJSON = string(b)
		}
	}
	if err := db.Create(&entry).Error; err != nil {
		logger.Warnf("[audit] 写入审计日志失败: %v", err)
	}
}

// firstNonEmpty 返回第一个非空字符串。
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// uniqueCode 基于名称生成唯一编码；名称不可用时退化为 prefix，exists 用于判重。
func uniqueCode(prefix, name string, exists func(code string) bool) string {
	base := slugify(name)
	if base == "" {
		base = slugify(prefix)
	}
	if base == "" {
		base = "item"
	}
	candidate := base
	if exists == nil || !exists(candidate) {
		return candidate
	}
	for i := 2; i < 1000; i++ {
		candidate = fmt.Sprintf("%s_%d", base, i)
		if !exists(candidate) {
			return candidate
		}
	}
	return fmt.Sprintf("%s_%d", base, time.Now().UnixNano())
}

// slugify 将任意名称转为小写下划线编码（非 ASCII 字符被忽略）。
func slugify(s string) string {
	var b strings.Builder
	lastUnderscore := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastUnderscore = false
		case r == '_' || r == '-' || r == ' ' || r == '.':
			if !lastUnderscore && b.Len() > 0 {
				b.WriteRune('_')
				lastUnderscore = true
			}
		}
	}
	return strings.Trim(b.String(), "_")
}

// marshalJSON 序列化对象为字符串，失败返回空串。
func marshalJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		logger.Warnf("[handler] JSON 序列化失败: %v", err)
		return ""
	}
	return string(b)
}

// rawJSON 将可能为字符串或对象的原始输入统一转为紧凑 JSON 字符串。
func rawJSON(v interface{}) string {
	switch tv := v.(type) {
	case nil:
		return ""
	case string:
		s := strings.TrimSpace(tv)
		if s == "" {
			return ""
		}
		var probe interface{}
		if err := json.Unmarshal([]byte(s), &probe); err != nil {
			// 非合法 JSON，按普通字符串封装
			return marshalJSON(s)
		}
		return s
	default:
		return marshalJSON(v)
	}
}

// tagsToString 将标签输入（数组或字符串）统一为逗号分隔字符串。
func tagsToString(v interface{}) string {
	switch tv := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(tv)
	case []interface{}:
		items := make([]string, 0, len(tv))
		for _, it := range tv {
			if s := strings.TrimSpace(fmt.Sprint(it)); s != "" {
				items = append(items, s)
			}
		}
		return strings.Join(items, ",")
	case []string:
		return strings.Join(tv, ",")
	default:
		return strings.TrimSpace(fmt.Sprint(tv))
	}
}

// containsString 判断字符串切片是否包含指定值（忽略大小写与首尾空白）。
func containsString(list []string, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	for _, item := range list {
		if strings.ToLower(strings.TrimSpace(item)) == target {
			return true
		}
	}
	return false
}
