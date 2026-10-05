package handler

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"smartg-odin/internal/ai"
	"smartg-odin/internal/common/errs"
	"smartg-odin/internal/common/logger"
	"smartg-odin/internal/ds/query"
	"smartg-odin/internal/middleware"
	"smartg-odin/internal/model"
)

// traceIDOf 从请求上下文读取中间件生成的 trace_id。
func traceIDOf(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return c.GetString(middleware.TraceIDKey)
}

// persistQueryTrace 写入一条 query_trace；失败仅告警，不影响主流程。
func persistQueryTrace(c *gin.Context, t *model.QueryTrace) {
	if t == nil {
		return
	}
	if strings.TrimSpace(t.TraceID) == "" {
		t.TraceID = traceIDOf(c)
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	if err := DB().Create(t).Error; err != nil {
		logger.Warnf("[query_trace] 写入查询链路失败: %v", err)
	}
}

// buildQueryTrace 由翻译结果组装 query_trace 记录（用于 ExecuteQuery / DryRunQuery）。
func buildQueryTrace(c *gin.Context, req *query.QueryRequest, tq *query.TranslatedQuery,
	status string, resultCount int, elapsedMs int64, errMsg string) *model.QueryTrace {
	t := &model.QueryTrace{
		TraceID:         traceIDOf(c),
		ParsePath:       "direct",
		Status:          status,
		ResultCount:     resultCount,
		ExecutionTimeMs: elapsedMs,
		ErrorMessage:    truncate(errMsg, 1000),
		CreatedAt:       time.Now(),
	}
	if req != nil {
		t.OntologyQueryJSON = marshalJSON(req)
	}
	if tq != nil {
		t.TranslatedSQL = tq.SQL
		t.ParamsJSON = marshalJSON(orEmptyParams(tq.Params))
		t.DataSourceID = tq.DataSourceID
		t.SourceTable = tq.SourceTable
		t.JoinsJSON = marshalJSON(tq.Joins)
		t.Explanation = tq.Explanation
		t.ReasonTraceJSON = marshalJSON(tq.ReasonTrace)
	}
	return t
}

// nlParsePayload nl-parse 请求体。
type nlParsePayload struct {
	OntologyID uint   `json:"ontology_id"`
	Text       string `json:"text"`
}

// NLParse POST /api/v1/query/nl-parse
// 意图解析入口：先规则（P1），未命中且 ai.enabled 时回退 Eino LLM 意图链；结果落 query_trace。
func NLParse(c *gin.Context) {
	var body nlParsePayload
	if appErr := bindNormalized(c, &body); appErr != nil {
		fail(c, appErr)
		return
	}
	if body.OntologyID == 0 {
		fail(c, errs.ParamError("ontology_id 不能为空"))
		return
	}
	body.Text = strings.TrimSpace(body.Text)
	if body.Text == "" {
		fail(c, errs.ParamError("text 不能为空"))
		return
	}
	if appErr := ensureOntology(body.OntologyID); appErr != nil {
		fail(c, appErr)
		return
	}

	result, perr := ai.ParseIntent(c.Request.Context(), DB(), body.OntologyID, body.Text)
	if perr != nil {
		fail(c, mapError(perr, "意图解析失败"))
		return
	}

	status := "parse_only"
	errMsg := ""
	if !result.Matched {
		errMsg = "意图解析未命中本体类"
	}
	llmJSON := ""
	if result.LLMMeta != nil {
		llmJSON = marshalJSON(result.LLMMeta)
	}
	persistQueryTrace(c, &model.QueryTrace{
		TraceID:           traceIDOf(c),
		NLText:            body.Text,
		ParsePath:         result.ParsePath,
		IntentJSON:        marshalJSON(result.Query),
		OntologyQueryJSON: marshalJSON(result.Query),
		LLMJSON:           llmJSON,
		Status:            status,
		ErrorMessage:      errMsg,
		CreatedAt:         time.Now(),
	})

	ok(c, gin.H{
		"query":         result.Query,
		"notes":         result.Notes,
		"hits":          result.Hits,
		"parse_path":    result.ParsePath,
		"matched":       result.Matched,
		"llm_meta":      result.LLMMeta,
		"canonical_key": result.CanonicalKey,
		"trace_id":      traceIDOf(c),
	})
}

// GetQueryTrace GET /api/v1/query/trace/:traceId
// 返回一次查询的全链路 JSON（自然语言 → 意图 → 推理 → SQL → 执行结果）。
func GetQueryTrace(c *gin.Context) {
	traceID := strings.TrimSpace(c.Param("traceId"))
	if traceID == "" {
		fail(c, errs.ParamError("缺少路径参数: traceId"))
		return
	}
	var t model.QueryTrace
	if err := DB().Where("trace_id = ?", traceID).Order("id desc").First(&t).Error; err != nil {
		fail(c, mapError(err, "查询链路不存在: "+traceID))
		return
	}
	ok(c, traceView(t))
}

// traceView 把 query_trace 的各 *_json 字段解析为对象，便于前端直接渲染。
func traceView(t model.QueryTrace) gin.H {
	return gin.H{
		"trace_id":          t.TraceID,
		"nl_text":           t.NLText,
		"parse_path":        t.ParsePath,
		"intent":            parseJSONField(t.IntentJSON),
		"ontology_query":    parseJSONField(t.OntologyQueryJSON),
		"llm":               parseJSONField(t.LLMJSON),
		"reason_trace":      parseJSONField(t.ReasonTraceJSON),
		"translated_sql":    t.TranslatedSQL,
		"params":            parseJSONField(t.ParamsJSON),
		"datasource_id":     t.DataSourceID,
		"source_table":      t.SourceTable,
		"joins":             parseJSONField(t.JoinsJSON),
		"explanation":       t.Explanation,
		"result_count":      t.ResultCount,
		"execution_time_ms": t.ExecutionTimeMs,
		"status":            t.Status,
		"error_message":     t.ErrorMessage,
		"nodes":             parseJSONField(t.NodesJSON),
		"created_at":        t.CreatedAt,
	}
}

// parseJSONField 将 JSON 字符串解析为对象；空串返回 nil，非法 JSON 返回原始字符串。
func parseJSONField(raw string) interface{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return raw
	}
	return v
}
