package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"smartg-odin/internal/common/errs"
	"smartg-odin/internal/common/logger"
	"smartg-odin/internal/ds/connector"
	"smartg-odin/internal/ds/query"
	"smartg-odin/internal/middleware"
	"smartg-odin/internal/model"
)

// 查询相关默认值
const (
	defaultQueryLimit = 100
	maxQueryLimit     = 10000
	defaultExplain    = false
)

// supportedFilterOps 是过滤操作符的实现集，与 translator.filterExpr 支持的分支严格一致。
// 校验词表、vocabulary 端点均复用此集合，避免出现「校验通过但翻译未实现」的操作符。
var supportedFilterOps = []string{
	"eq", "neq", "ne", "gt", "gte", "ge", "lt", "lte", "le",
	"like", "in", "not_in", "nin", "is_null", "is_not_null", "not_null",
}

// orderDirs vocabulary 支持的排序方向。
var orderDirs = []string{"asc", "desc"}

// supportedAggFuncs 是聚合函数的实现集，与 translator.supportedAggFuncs、ai.supportedAggregateFuncs、
// intent_tool 的 Enum 严格一致（单一口径），避免出现「校验通过但翻译未实现」的聚合函数。
var supportedAggFuncs = []string{"count", "sum", "avg", "min", "max"}

// isSupportedAggFunc 判断聚合函数是否受支持。
func isSupportedAggFunc(fn string) bool {
	fn = strings.ToLower(strings.TrimSpace(fn))
	for _, f := range supportedAggFuncs {
		if f == fn {
			return true
		}
	}
	return false
}

// orEmptyColumnMeta 保证 column_meta 数组不为 null。
func orEmptyColumnMeta(meta []query.ColumnMeta) []query.ColumnMeta {
	if meta == nil {
		return []query.ColumnMeta{}
	}
	return meta
}

// ---------------------------------------------------------------------------
// 查询执行
// ---------------------------------------------------------------------------

// ExecuteQuery POST /api/v1/query
// 本体查询：翻译为物理 SQL → 在目标数据源执行 → 记录历史 → 返回结果集。
func ExecuteQuery(c *gin.Context) {
	req, appErr := bindQueryRequest(c)
	if appErr != nil {
		fail(c, appErr)
		return
	}

	tq, err := query.Translate(DB(), req)
	if err != nil {
		recordQueryHistory(req, "", "failed", 0, 0, err.Error())
		persistQueryTrace(c, buildQueryTrace(c, req, nil, "failed", 0, 0, err.Error()))
		fail(c, translateError(err))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), requestTimeout(c))
	defer cancel()

	started := time.Now()
	result, err := query.Execute(ctx, DB(), tq)
	elapsed := time.Since(started).Milliseconds()

	if err != nil {
		recordQueryHistory(req, tq.SQL, "failed", 0, elapsed, err.Error())
		persistQueryTrace(c, buildQueryTrace(c, req, tq, "failed", 0, elapsed, err.Error()))
		writeAudit(c, "query", "query", tq.DataSourceID, gin.H{
			"sql": tq.SQL, "status": "failed", "class_name": tq.ClassName,
		})
		logger.Warnf("[query] 执行失败 ontology=%d class=%s err=%v", req.OntologyID, req.ClassName, err)
		fail(c, classifyQueryError(err))
		return
	}
	if result == nil {
		result = emptyQueryResult()
	}

	recordQueryHistory(req, tq.SQL, "success", result.RowCount, elapsed, "")
	persistQueryTrace(c, buildQueryTrace(c, req, tq, "success", result.RowCount, elapsed, ""))
	writeAudit(c, "query", "query", tq.DataSourceID, gin.H{
		"sql": tq.SQL, "status": "success", "rows": result.RowCount, "class_name": tq.ClassName,
	})
	logger.Infof("[query] 执行成功 ontology=%d class=%s rows=%d cost=%dms",
		req.OntologyID, req.ClassName, result.RowCount, elapsed)

	ok(c, queryResponse(c, tq, result, req, elapsed))
}

// DryRunQuery POST /api/v1/query/dry-run
// 只做本体 → SQL 翻译，不实际执行，用于预览与调试。
func DryRunQuery(c *gin.Context) {
	req, appErr := bindQueryRequest(c)
	if appErr != nil {
		fail(c, appErr)
		return
	}

	started := time.Now()
	tq, err := query.Translate(DB(), req)
	if err != nil {
		persistQueryTrace(c, buildQueryTrace(c, req, nil, "failed", 0, 0, err.Error()))
		fail(c, translateError(err))
		return
	}
	translateMs := time.Since(started).Milliseconds()
	persistQueryTrace(c, buildQueryTrace(c, req, tq, "dry_run", 0, translateMs, ""))

	dsName, dsType := datasourceBrief(tq.DataSourceID)
	okMsg(c, gin.H{
		"executed":        false,
		"sql":             tq.SQL,
		"params":          orEmptyParams(tq.Params),
		"param_count":     len(tq.Params),
		"explanation":     tq.Explanation,
		"joins":           tq.Joins,
		"class_name":      tq.ClassName,
		"source_table":    tq.SourceTable,
		"datasource_id":   tq.DataSourceID,
		"datasource_name": dsName,
		"datasource_type": dsType,
		"reason_trace":    tq.ReasonTrace,
		"aggregated":      tq.Aggregated,
		"column_meta":     orEmptyColumnMeta(tq.ColumnMeta),
		"trace_id":        traceIDOf(c),
		"request":         req,
		"translate_ms":    translateMs,
		// P2-3：1:N 扇出警告与被确定性丢弃的 join（供前端提示）。
		"warning":       tq.Warning,
		"dropped_joins": orEmptyStrSlice(tq.DroppedJoins),
	}, "翻译完成（未执行）")
}

// ExplainQuery POST /api/v1/query/explain
// 翻译后以 EXPLAIN 方式执行，返回数据源的查询计划。
func ExplainQuery(c *gin.Context) {
	req, appErr := bindQueryRequest(c)
	if appErr != nil {
		fail(c, appErr)
		return
	}

	tq, err := query.Translate(DB(), req)
	if err != nil {
		fail(c, translateError(err))
		return
	}

	ds, appErr := loadDatasource(tq.DataSourceID)
	if appErr != nil {
		fail(c, appErr)
		return
	}

	explainSQL := explainStatement(ds.Type, tq.SQL)
	planQuery := *tq
	planQuery.SQL = explainSQL

	ctx, cancel := context.WithTimeout(c.Request.Context(), requestTimeout(c))
	defer cancel()

	started := time.Now()
	result, err := query.Execute(ctx, DB(), &planQuery)
	if err != nil {
		writeAudit(c, "explain", "query", tq.DataSourceID, gin.H{"sql": explainSQL, "status": "failed"})
		fail(c, classifyQueryError(err))
		return
	}
	if result == nil {
		result = emptyQueryResult()
	}

	okMsg(c, gin.H{
		"sql":             tq.SQL,
		"explain_sql":     explainSQL,
		"explanation":     tq.Explanation,
		"class_name":      tq.ClassName,
		"source_table":    tq.SourceTable,
		"datasource_id":   tq.DataSourceID,
		"datasource_name": ds.Name,
		"datasource_type": ds.Type,
		"plan": gin.H{
			"columns": result.Columns,
			"rows":    result.Rows,
			"records": rowsToRecords(result),
		},
		"duration_ms": time.Since(started).Milliseconds(),
	}, "查询计划生成完成")
}

// ---------------------------------------------------------------------------
// 查询历史
// ---------------------------------------------------------------------------

// GetQueryHistory GET /api/v1/query/history?current=1&size=20&ontology_id=&status=&query_type=
func GetQueryHistory(c *gin.Context) {
	current, size := parsePage(c)

	filter := func() *gorm.DB {
		tx := DB().Model(&model.QueryHistory{})
		if v := uintQuery(c, "ontology_id"); v > 0 {
			tx = tx.Where("ontology_id = ?", v)
		}
		if s := strings.TrimSpace(c.Query("status")); s != "" {
			tx = tx.Where("status = ?", s)
		}
		if t := strings.TrimSpace(c.Query("query_type")); t != "" {
			tx = tx.Where("query_type = ?", t)
		}
		if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
			tx = tx.Where("query_text LIKE ?", "%"+kw+"%")
		}
		return tx
	}

	var total int64
	if err := filter().Count(&total).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	var rows []model.QueryHistory
	if err := filter().Order("id DESC").Offset((current - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}
	if rows == nil {
		rows = []model.QueryHistory{}
	}

	// 附带本体名称，便于前端展示
	names := ontologyNames(collectOntologyIDs(rows))
	views := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		item := gin.H{
			"id":                r.ID,
			"query_text":        r.QueryText,
			"query_type":        r.QueryType,
			"ontology_id":       r.OntologyID,
			"ontology_name":     "",
			"result_count":      r.ResultCount,
			"execution_time_ms": r.ExecutionTimeMs,
			"status":            r.Status,
			"error_message":     r.ErrorMessage,
			"created_at":        r.CreatedAt,
		}
		if r.OntologyID != nil {
			item["ontology_name"] = names[*r.OntologyID]
		}
		views = append(views, item)
	}

	middleware.SuccessPage(c, views, total, current, size)
}

// ---------------------------------------------------------------------------
// 查询模板
// ---------------------------------------------------------------------------

// queryTemplateView 查询模板视图，附带解析后的查询体。
type queryTemplateView struct {
	model.QueryTemplate
	Query interface{} `json:"query,omitempty"`
}

// templatePayload 保存模板的请求体。
type templatePayload struct {
	ID          uint            `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	OntologyID  uint            `json:"ontology_id"`
	QueryJSON   string          `json:"query_json"`
	Query       json.RawMessage `json:"query"`
	ParamsJSON  string          `json:"params_json"`
	Params      json.RawMessage `json:"params"`
	CreatedBy   string          `json:"created_by"`
}

// ListQueryTemplates GET /api/v1/query/templates?ontology_id=&keyword=&current=1&size=20
func ListQueryTemplates(c *gin.Context) {
	current, size := parsePage(c)

	filter := func() *gorm.DB {
		tx := DB().Model(&model.QueryTemplate{})
		if v := uintQuery(c, "ontology_id"); v > 0 {
			tx = tx.Where("ontology_id = ?", v)
		}
		if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
			like := "%" + kw + "%"
			tx = tx.Where("name LIKE ? OR description LIKE ?", like, like)
		}
		return tx
	}

	var total int64
	if err := filter().Count(&total).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	var rows []model.QueryTemplate
	if err := filter().Order("id DESC").Offset((current - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	views := make([]queryTemplateView, 0, len(rows))
	for _, r := range rows {
		views = append(views, buildTemplateView(r))
	}
	middleware.SuccessPage(c, views, total, current, size)
}

// SaveQueryTemplate POST /api/v1/query/templates
// 携带 id 时执行更新，否则创建新模板。
func SaveQueryTemplate(c *gin.Context) {
	var req templatePayload
	if appErr := bindNormalized(c, &req); appErr != nil {
		fail(c, appErr)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		fail(c, errs.ParamError("模板名称 name 不能为空"))
		return
	}

	queryJSON := strings.TrimSpace(req.QueryJSON)
	if queryJSON == "" && len(req.Query) > 0 {
		queryJSON = string(req.Query)
	}
	if queryJSON == "" {
		fail(c, errs.ParamError("查询内容 query（或 query_json）不能为空"))
		return
	}
	var probe interface{}
	if err := json.Unmarshal([]byte(queryJSON), &probe); err != nil {
		fail(c, errs.ParamError("query 必须为合法 JSON: "+err.Error()))
		return
	}

	paramsJSON := strings.TrimSpace(req.ParamsJSON)
	if paramsJSON == "" && len(req.Params) > 0 {
		paramsJSON = string(req.Params)
	}

	if req.OntologyID > 0 {
		if appErr := ensureOntology(req.OntologyID); appErr != nil {
			fail(c, appErr)
			return
		}
	}

	// 更新分支
	if req.ID > 0 {
		var existing model.QueryTemplate
		if err := DB().First(&existing, req.ID).Error; err != nil {
			fail(c, mapError(err, fmt.Sprintf("查询模板不存在: %d", req.ID)))
			return
		}
		updates := map[string]interface{}{
			"name":        req.Name,
			"description": req.Description,
			"query_json":  queryJSON,
			"updated_at":  time.Now(),
		}
		if req.OntologyID > 0 {
			updates["ontology_id"] = req.OntologyID
		}
		if paramsJSON != "" {
			updates["params_json"] = paramsJSON
		}
		if strings.TrimSpace(req.CreatedBy) != "" {
			updates["created_by"] = strings.TrimSpace(req.CreatedBy)
		}
		if err := DB().Model(&model.QueryTemplate{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
			fail(c, errs.DBError(err))
			return
		}
		DB().First(&existing, req.ID)
		writeAudit(c, "update", "query_template", existing.ID, gin.H{"name": existing.Name})
		okMsg(c, buildTemplateView(existing), "模板已更新")
		return
	}

	// 同名模板视为覆盖更新，避免重复堆积
	var existing model.QueryTemplate
	err := DB().Where("name = ?", req.Name).First(&existing).Error
	if err == nil {
		updates := map[string]interface{}{
			"description": req.Description,
			"query_json":  queryJSON,
			"updated_at":  time.Now(),
		}
		if req.OntologyID > 0 {
			updates["ontology_id"] = req.OntologyID
		}
		if paramsJSON != "" {
			updates["params_json"] = paramsJSON
		}
		if err := DB().Model(&model.QueryTemplate{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
			fail(c, errs.DBError(err))
			return
		}
		DB().First(&existing, existing.ID)
		writeAudit(c, "update", "query_template", existing.ID, gin.H{"name": existing.Name, "by": "same-name"})
		okMsg(c, buildTemplateView(existing), "同名模板已更新")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		fail(c, errs.DBError(err))
		return
	}

	tpl := model.QueryTemplate{
		Name:        req.Name,
		Description: req.Description,
		OntologyID:  req.OntologyID,
		QueryJSON:   queryJSON,
		ParamsJSON:  paramsJSON,
		CreatedBy:   firstNonEmpty(strings.TrimSpace(req.CreatedBy), c.GetHeader("X-User-Id"), "anonymous"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := DB().Create(&tpl).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	writeAudit(c, "create", "query_template", tpl.ID, gin.H{"name": tpl.Name})
	okMsg(c, buildTemplateView(tpl), "模板已保存")
}

// ---------------------------------------------------------------------------
// 内部辅助
// ---------------------------------------------------------------------------

// emptyQueryResult 返回兜底的空结果集，避免响应中出现 null。
func emptyQueryResult() *connector.QueryResult {
	return &connector.QueryResult{
		Columns: []string{},
		Rows:    [][]interface{}{},
	}
}

// queryResponse 组装查询执行响应：同时返回二维数组与对象数组两种结果形式，
// 前端表格可直接使用 records，导出场景可使用 columns + rows。
func queryResponse(c *gin.Context, tq *query.TranslatedQuery, result *connector.QueryResult,
	req *query.QueryRequest, elapsed int64) gin.H {
	dsName, dsType := datasourceBrief(tq.DataSourceID)
	return gin.H{
		"sql":             tq.SQL,
		"params":          orEmptyParams(tq.Params),
		"explanation":     tq.Explanation,
		"joins":           tq.Joins,
		"class_name":      tq.ClassName,
		"source_table":    tq.SourceTable,
		"datasource_id":   tq.DataSourceID,
		"datasource_name": dsName,
		"datasource_type": dsType,
		"columns":         result.Columns,
		"rows":            result.Rows,
		"records":         rowsToRecords(result),
		"row_count":       result.RowCount,
		"duration_ms":     elapsed,
		"executed":        true,
		"aggregated":      tq.Aggregated,
		"column_meta":     orEmptyColumnMeta(tq.ColumnMeta),
		"reason_trace":    tq.ReasonTrace,
		"trace_id":        traceIDOf(c),
		"request":         req,
		// P2-3：1:N 扇出警告与被确定性丢弃的 join（供前端提示）。
		"warning":       tq.Warning,
		"dropped_joins": orEmptyStrSlice(tq.DroppedJoins),
	}
}

// rowsToRecords 将 columns + rows 转换为对象数组，便于前端直接渲染。
func rowsToRecords(result *connector.QueryResult) []map[string]interface{} {
	if result == nil || len(result.Rows) == 0 || len(result.Columns) == 0 {
		return []map[string]interface{}{}
	}
	records := make([]map[string]interface{}, 0, len(result.Rows))
	for _, row := range result.Rows {
		record := make(map[string]interface{}, len(result.Columns))
		for i, col := range result.Columns {
			if i < len(row) {
				record[col] = row[i]
			} else {
				record[col] = nil
			}
		}
		records = append(records, record)
	}
	return records
}

// bindQueryRequest 解析并校验本体查询请求。
func bindQueryRequest(c *gin.Context) (*query.QueryRequest, *errs.AppError) {
	var req query.QueryRequest
	if appErr := bindNormalized(c, &req); appErr != nil {
		return nil, appErr
	}

	req.ClassName = strings.TrimSpace(req.ClassName)
	if req.OntologyID == 0 {
		return nil, errs.ParamError("ontology_id 不能为空")
	}
	if req.ClassName == "" {
		return nil, errs.ParamError("class_name 不能为空")
	}
	// limit/offset 兜底统一交由 query.NormalizeLimit（聚合查询不做 100 截断、明细查询缺省 100），
	// 与 assistant SSE 链路（skills.parseAndTranslate）共用同一策略，保证同一 intent 双入口 SQL 逐字节一致。
	query.NormalizeLimit(&req)
	switch strings.ToLower(strings.TrimSpace(req.OrderDir)) {
	case "desc", "descending", "down":
		req.OrderDir = "desc"
	default:
		if strings.TrimSpace(req.OrderDir) != "" {
			req.OrderDir = "asc"
		}
	}
	for i := range req.Filters {
		op := strings.ToLower(strings.TrimSpace(req.Filters[i].Op))
		if op == "" {
			op = "eq"
		}
		if !isSupportedFilterOp(op) {
			return nil, errs.ParamError("不支持的过滤操作符: " + req.Filters[i].Op)
		}
		req.Filters[i].Op = op
		req.Filters[i].Property = strings.TrimSpace(req.Filters[i].Property)
		if req.Filters[i].Property == "" {
			return nil, errs.ParamError("过滤条件缺少 property 字段")
		}
	}

	// aggregates 校验归一（func 白名单；count 外必须指定 property）。
	for i := range req.Aggregates {
		fn := strings.ToLower(strings.TrimSpace(req.Aggregates[i].Func))
		if !isSupportedAggFunc(fn) {
			return nil, errs.ParamError("不支持的聚合函数: " + req.Aggregates[i].Func)
		}
		req.Aggregates[i].Func = fn
		req.Aggregates[i].Property = strings.TrimSpace(req.Aggregates[i].Property)
		req.Aggregates[i].Alias = strings.TrimSpace(req.Aggregates[i].Alias)
		if fn != "count" && req.Aggregates[i].Property == "" {
			return nil, errs.ParamError("聚合函数 " + fn + " 必须指定 property")
		}
	}
	// group_by 归一（trim；保留请求顺序，不做排序以保证确定性）。
	for i := range req.GroupBy {
		req.GroupBy[i] = strings.TrimSpace(req.GroupBy[i])
	}
	// having 校验（op 复用现有过滤操作符集；alias 必填）。
	for i := range req.Having {
		op := strings.ToLower(strings.TrimSpace(req.Having[i].Op))
		if op == "" {
			op = "eq"
		}
		if !isSupportedFilterOp(op) {
			return nil, errs.ParamError("不支持的 having 操作符: " + req.Having[i].Op)
		}
		req.Having[i].Op = op
		req.Having[i].Alias = strings.TrimSpace(req.Having[i].Alias)
		if req.Having[i].Alias == "" {
			return nil, errs.ParamError("having 条件缺少 alias 字段")
		}
	}
	return &req, nil
}

// isSupportedFilterOp 判断过滤操作符是否受支持（与实现集严格一致）。
func isSupportedFilterOp(op string) bool {
	for _, s := range supportedFilterOps {
		if s == op {
			return true
		}
	}
	return false
}

// requestTimeout 读取请求指定的超时时间（秒），缺省使用全局默认值。
func requestTimeout(c *gin.Context) time.Duration {
	seconds := parseIntQuery(c, "timeout", 0)
	if seconds <= 0 {
		if v := parseIntQuery(c, "timeout_ms", 0); v > 0 {
			return time.Duration(v) * time.Millisecond
		}
		return queryTimeout
	}
	if seconds > 600 {
		seconds = 600
	}
	return time.Duration(seconds) * time.Second
}

// recordQueryHistory 写入查询执行历史。
func recordQueryHistory(req *query.QueryRequest, sqlText, status string, resultCount int, elapsedMs int64, errMsg string) {
	history := model.QueryHistory{
		QueryType:       "ontology",
		ResultCount:     resultCount,
		ExecutionTimeMs: elapsedMs,
		Status:          status,
		ErrorMessage:    truncate(errMsg, 1000),
		CreatedAt:       time.Now(),
	}
	if strings.TrimSpace(sqlText) != "" {
		history.QueryText = sqlText
	} else if req != nil {
		history.QueryText = marshalJSON(req)
	}
	if req != nil && req.OntologyID > 0 {
		ontologyID := req.OntologyID
		history.OntologyID = &ontologyID
	}
	if err := DB().Create(&history).Error; err != nil {
		logger.Warnf("[query] 写入查询历史失败: %v", err)
	}
}

// translateError 将翻译阶段的错误映射为业务错误码。
func translateError(err error) *errs.AppError {
	if err == nil {
		return errs.New(errs.CodeInternal, "查询翻译失败")
	}
	msg := err.Error()
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "not found"), strings.Contains(lower, "no mapping"), strings.Contains(lower, "不存在"):
		return errs.Wrap(errs.CodeNotFound, "查询翻译失败: "+msg, err)
	case strings.Contains(lower, "required"), strings.Contains(lower, "invalid"),
		strings.Contains(lower, "unsupported"), strings.Contains(lower, "empty"):
		return errs.Wrap(errs.CodeParamError, "查询翻译失败: "+msg, err)
	default:
		return errs.Wrap(errs.CodeInternal, "查询翻译失败: "+msg, err)
	}
}

// classifyQueryError 将执行阶段的错误映射为业务错误码。
func classifyQueryError(err error) *errs.AppError {
	if err == nil {
		return errs.New(errs.CodeInternal, "查询执行失败")
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return errs.QueryTimeout(err)
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline"), strings.Contains(msg, "i/o"):
		return errs.QueryTimeout(err)
	case strings.Contains(msg, "connection refused"), strings.Contains(msg, "dial tcp"),
		strings.Contains(msg, "no such host"), strings.Contains(msg, "connect:"):
		return errs.ConnRefused(err)
	case strings.Contains(msg, "password"), strings.Contains(msg, "authentication"), strings.Contains(msg, "denied"):
		return errs.ConnError(err)
	case strings.Contains(msg, "syntax error"), strings.Contains(msg, "no such table"),
		strings.Contains(msg, "no such column"), strings.Contains(msg, "unknown column"):
		return errs.Wrap(errs.CodeDBError, "SQL 执行失败: "+err.Error(), err)
	default:
		return errs.Wrap(errs.CodeConnError, "查询执行失败: "+err.Error(), err)
	}
}

// explainStatement 依据数据源类型生成 EXPLAIN 语句。
func explainStatement(dsType, sqlText string) string {
	prefix := "EXPLAIN "
	switch strings.ToLower(strings.TrimSpace(dsType)) {
	case "sqlite":
		prefix = "EXPLAIN QUERY PLAN "
	case "postgres", "postgresql":
		prefix = "EXPLAIN (FORMAT TEXT) "
	case "mysql", "mariadb":
		prefix = "EXPLAIN FORMAT=TRADITIONAL "
	}
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sqlText)), "EXPLAIN") {
		return sqlText
	}
	return prefix + sqlText
}

// datasourceBrief 返回数据源名称与类型。
func datasourceBrief(id uint) (string, string) {
	if id == 0 {
		return "", ""
	}
	var ds model.DataSource
	if err := DB().Select("id, name, type").First(&ds, id).Error; err != nil {
		return "", ""
	}
	return ds.Name, ds.Type
}

// buildTemplateView 组装模板视图。
func buildTemplateView(tpl model.QueryTemplate) queryTemplateView {
	view := queryTemplateView{QueryTemplate: tpl}
	if strings.TrimSpace(tpl.QueryJSON) != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(tpl.QueryJSON), &parsed); err == nil {
			view.Query = parsed
		}
	}
	if tpl.ParamsJSON == "" {
		view.ParamsJSON = "{}"
	}
	return view
}

// collectOntologyIDs 从查询历史中收集本体 ID。
func collectOntologyIDs(rows []model.QueryHistory) []uint {
	ids := make([]uint, 0, len(rows))
	seen := map[uint]bool{}
	for _, r := range rows {
		if r.OntologyID == nil || seen[*r.OntologyID] {
			continue
		}
		seen[*r.OntologyID] = true
		ids = append(ids, *r.OntologyID)
	}
	return ids
}

// ontologyNames 批量加载本体 ID → 名称映射。
func ontologyNames(ids []uint) map[uint]string {
	result := map[uint]string{}
	if len(ids) == 0 {
		return result
	}
	var rows []model.OntDefinition
	DB().Select("id, name").Where("id IN ?", ids).Find(&rows)
	for _, r := range rows {
		result[r.ID] = r.Name
	}
	return result
}

// orEmptyParams 保证参数数组不为 null。
func orEmptyParams(params []interface{}) []interface{} {
	if params == nil {
		return []interface{}{}
	}
	return params
}
