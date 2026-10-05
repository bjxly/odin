package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"smartg-odin/internal/common/errs"
	"smartg-odin/internal/ds/query"
	"smartg-odin/internal/model"
	"smartg-odin/internal/ont/reason"
)

// reasonExplainPayload reason/explain 请求体。
type reasonExplainPayload struct {
	OntologyID uint           `json:"ontology_id"`
	ClassName  string         `json:"class_name"`
	Filters    []query.Filter `json:"filters"`
}

// ReasonExplain POST /api/v1/reason/explain
// 仅做推理（规则应用 + 虚拟类展开 + 关系推导），返回 ReasonTrace，不翻译也不执行 SQL。
func ReasonExplain(c *gin.Context) {
	var body reasonExplainPayload
	if appErr := bindNormalized(c, &body); appErr != nil {
		fail(c, appErr)
		return
	}
	body.ClassName = strings.TrimSpace(body.ClassName)
	if body.OntologyID == 0 {
		fail(c, errs.ParamError("ontology_id 不能为空"))
		return
	}
	if body.ClassName == "" {
		fail(c, errs.ParamError("class_name 不能为空"))
		return
	}
	if appErr := ensureOntology(body.OntologyID); appErr != nil {
		fail(c, appErr)
		return
	}

	db := DB()
	var cls model.OntClass
	if err := db.Where("ontology_id = ? AND name = ?", body.OntologyID, body.ClassName).First(&cls).Error; err != nil {
		fail(c, mapError(err, "本体类不存在: "+body.ClassName))
		return
	}

	trace := reason.NewReasonTrace()
	qctx := &reason.QueryContext{
		OntologyID: body.OntologyID,
		ClassName:  body.ClassName,
		ClassID:    cls.ID,
		Trace:      trace,
	}
	for _, f := range body.Filters {
		qctx.Filters = append(qctx.Filters, reason.ContextFilter{Column: f.Property, Op: f.Op, Value: f.Value})
	}
	if _, err := reason.ApplyRules(db, qctx); err != nil {
		fail(c, errs.Internal("规则应用失败", err))
		return
	}
	expanded, err := reason.ExpandVirtualClassWithTrace(db, body.OntologyID, body.ClassName, trace)
	if err != nil {
		fail(c, translateError(err))
		return
	}
	// 关系推导为附加解释信息，失败不阻断主流程。
	_, _ = reason.InferRelationsWithTrace(db, body.OntologyID, body.ClassName, trace)

	targets := make([]gin.H, 0, len(expanded))
	for _, ec := range expanded {
		dsName, dsType := datasourceBrief(ec.DataSourceID)
		targets = append(targets, gin.H{
			"class_name":      ec.ClassName,
			"source_table":    ec.SourceTable,
			"datasource_id":   ec.DataSourceID,
			"datasource_name": dsName,
			"datasource_type": dsType,
		})
	}

	// 落库 query_trace，使 GET /query/trace/:id 可查到本次推理（含 reason_trace）。
	explainReq := &query.QueryRequest{
		OntologyID: body.OntologyID,
		ClassName:  body.ClassName,
		Filters:    body.Filters,
	}
	qt := buildQueryTrace(c, explainReq, nil, "success", 0, 0, "")
	qt.ParsePath = "reason"
	qt.NLText = body.ClassName
	qt.ReasonTraceJSON = marshalJSON(trace)
	persistQueryTrace(c, qt)

	ok(c, gin.H{
		"ontology_id":      body.OntologyID,
		"class_name":       body.ClassName,
		"class_type":       cls.ClassType,
		"reason_trace":     trace,
		"expanded_targets": targets,
		"trace_id":         traceIDOf(c),
	})
}
