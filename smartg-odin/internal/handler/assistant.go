package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"smartg-odin/internal/ai"
	"smartg-odin/internal/ai/sqlgen"
	"smartg-odin/internal/common/errs"
	"smartg-odin/internal/common/logger"
	"smartg-odin/internal/middleware"
	"smartg-odin/internal/model"
)

// sql_gen 单一路径的 parse_path / skill_id 标识（前端据此渲染徽章与消息头）。
const (
	parsePathSQLGen = "sql_gen"
	skillIDSQLGen   = "sql_gen"
)

// assistantChatPayload POST /api/v1/assistant/chat 请求体。
//
// 兼容两种问题字段：新契约 message 与旧前端 text（取先非空者）；
// ontology_id 兼容数字与字符串两种 JSON 形态（0/空则回退默认已发布本体）。
type assistantChatPayload struct {
	SessionID  string      `json:"session_id"`
	Text       string      `json:"text"`
	Message    string      `json:"message"`
	OntologyID interface{} `json:"ontology_id"`
}

// ---------------------------------------------------------------------------
// SSE 辅助
// ---------------------------------------------------------------------------

// setSSEHeaders 设置 Server-Sent Events 响应头并立即刷新。
func setSSEHeaders(c *gin.Context) {
	h := c.Writer.Header()
	h.Set("Content-Type", "text/event-stream; charset=utf-8")
	h.Set("Cache-Control", "no-cache, no-transform")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()
}

// sseSend 发送一个 SSE 事件（event: 名称 + data: JSON），并立即刷新到客户端。
func sseSend(c *gin.Context, event string, data interface{}) {
	b, err := marshalJSONBytes(data)
	if err != nil {
		b = []byte("{}")
	}
	fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, string(b))
	c.Writer.Flush()
}

// marshalJSONBytes 序列化为 JSON 字节；失败返回错误。
func marshalJSONBytes(v interface{}) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	return json.Marshal(v)
}

// ---------------------------------------------------------------------------
// POST /api/v1/assistant/chat —— SSE 流式对话（sqlgen 单一路径）
// ---------------------------------------------------------------------------

// AssistantChat 助手对话入口：LLM 直接生成 SQL 的确定性受控管线，以 SSE 流式返回渲染事件。
//
// 单一路径（不再有 concept/multihop/plan/agent/rule 分派）：
//
//	BuildContext(装配上下文) → Generate(LLM 生成结构化 SQL) → Validate(安全校验+强制 LIMIT)
//	  → Execute(各数据源执行) → Merge(合并多步结果) → ChartSuggestion(图表建议)
//
// 事件序列：
//
//	meta → block(sql) → block(table) → [block(chart)] → block(text 结论)
//	  → provenance → done；任一环节失败 → error → done(ok=false)。
func AssistantChat(c *gin.Context) {
	var body assistantChatPayload
	if appErr := bindNormalized(c, &body); appErr != nil {
		fail(c, appErr)
		return
	}
	question := strings.TrimSpace(body.Message)
	if question == "" {
		question = strings.TrimSpace(body.Text)
	}
	if question == "" {
		fail(c, errs.ParamError("message 不能为空"))
		return
	}

	db := DB()
	traceID := traceIDOf(c)

	sessionID := strings.TrimSpace(body.SessionID)
	if sessionID == "" {
		sessionID = uuid.NewString()
	}
	ontologyID := resolveAssistantOntologyID(db, body.OntologyID)
	ensureAssistantSession(db, sessionID, ontologyID, question)
	saveAssistantMessage(db, sessionID, "user", question, "", "", "", traceID, "")

	// 参数与会话就绪后再切换到 SSE 模式（此前失败走 JSON，前端据 Content-Type 分流）。
	setSSEHeaders(c)

	// 构建 sqlgen 管线配置（取自 ai.sql_gen 段），并规整非法值。
	aiCfg := ai.Config()
	sqlCfg := (&sqlgen.SQLGenConfig{
		MaxRepair:      aiCfg.SQLGen.MaxRepair,
		MaxLimit:       aiCfg.SQLGen.MaxLimit,
		DefaultLimit:   aiCfg.SQLGen.DefaultLimit,
		TimeoutSeconds: aiCfg.SQLGen.TimeoutSeconds,
		MaxRows:        aiCfg.SQLGen.MaxRows,
	}).Normalize()

	reqCtx := c.Request.Context()

	// 1) 装配 LLM 生成 SQL 所需的完整上下文（物理 Schema / JOIN / 同义词 / 规则）。
	sqlCtx, err := sqlgen.BuildContext(db, ontologyIDString(ontologyID))
	if err != nil {
		finishSQLError(c, db, sessionID, traceID, question, "上下文装配失败: "+err.Error())
		return
	}

	// 2) 调用 LLM 生成结构化 SQL（单条或多条）。
	result, err := sqlgen.Generate(reqCtx, sqlCtx, question, sqlCfg, &aiCfg)
	if err != nil {
		finishSQLError(c, db, sessionID, traceID, question, "SQL 生成失败: "+err.Error())
		return
	}

	// 3) 安全校验 + 强制 LIMIT；未通过的步骤被剔除，全部失败则整体报错。
	validationErrs, validatedSteps := sqlgen.Validate(result.Steps, sqlCtx.AllowedTables(), sqlCfg)
	if len(validatedSteps) == 0 {
		reason := "所有 SQL 均未通过安全校验"
		if len(validationErrs) > 0 {
			reason = validationErrs[0].Reason
		}
		finishSQLError(c, db, sessionID, traceID, question, "SQL 校验失败: "+reason)
		return
	}
	result.Steps = validatedSteps

	// 4) meta 事件（首个事件）。
	sseSend(c, "meta", gin.H{
		"trace_id":       traceID,
		"session_id":     sessionID,
		"parse_path":     parsePathSQLGen,
		"skill_id":       skillIDSQLGen,
		"matched":        true,
		"llm_calls":      2,
		"step_count":     len(result.Steps),
		"datasource_ids": sqlCtx.DatasourceIDs(),
	})

	// 5) sql 块（前端默认折叠展示生成的多条 SQL）。
	sseSend(c, "block", gin.H{
		"type": "sql",
		"payload": gin.H{
			"steps": buildSQLStepsPayload(result.Steps, sqlCtx),
		},
	})

	// 6) 执行 SQL（单步失败不终止整体，错误记录进对应 StepResult）。
	stepResults, err := sqlgen.Execute(reqCtx, db, result.Steps, sqlCfg)
	if err != nil {
		finishSQLError(c, db, sessionID, traceID, question, "SQL 执行失败: "+err.Error())
		return
	}
	if allStepsFailed(stepResults) {
		finishSQLError(c, db, sessionID, traceID, question, "SQL 执行失败: "+firstStepError(stepResults))
		return
	}

	// 7) 合并多步结果。
	merged, err := sqlgen.Merge(stepResults, result.MergeStrategy, result.MergeKeys)
	if err != nil {
		finishSQLError(c, db, sessionID, traceID, question, "结果合并失败: "+err.Error())
		return
	}

	// 8) table 块。
	sseSend(c, "block", buildTableBlock(merged))

	// 9) chart 块（如果有图表建议）。
	if chart := sqlgen.ChartSuggestion(merged); chart != nil {
		sseSend(c, "block", gin.H{"type": "chart", "payload": chart})
	}

	// 10) text 块（结论）：二次 LLM 综合，基于「用户问题 + 实际查询结果」生成真实结论；
	//     失败或空结论时回退到第一轮结论（若非占位符），最终兜底为通用文案。
	conclusion, concludeErr := sqlgen.Conclude(reqCtx, question, result.Steps, stepResults, sqlCfg, &aiCfg)
	if concludeErr != nil {
		logger.Warnf("[assistant] 结论二次生成失败，回退兜底文案: %v", concludeErr)
	}
	if concludeErr != nil || strings.TrimSpace(conclusion) == "" {
		conclusion = strings.TrimSpace(result.Conclusion)
		if conclusion == "" || conclusion == "待查询后补充" {
			conclusion = buildAutoConclusion(merged, result.Steps)
		}
	}
	sseSend(c, "block", gin.H{
		"type":    "text",
		"payload": gin.H{"content": conclusion, "text": conclusion},
	})

	// 11) 汇总耗时与行数。
	var totalDuration int64
	totalRows := 0
	for _, sr := range stepResults {
		totalDuration += sr.Duration
		totalRows += sr.RowCount
	}

	// 12) provenance 事件（溯源摘要）。
	sseSend(c, "provenance", gin.H{
		"trace_id":       traceID,
		"parse_path":     parsePathSQLGen,
		"row_count":      totalRows,
		"duration_ms":    totalDuration,
		"datasource_ids": sqlCtx.DatasourceIDs(),
		"sql_hash":       computeSQLHash(result.Steps),
	})

	// 13) done 事件。
	sseSend(c, "done", gin.H{
		"trace_id":   traceID,
		"session_id": sessionID,
		"skill_id":   skillIDSQLGen,
		"parse_path": parsePathSQLGen,
		"matched":    true,
		"ok":         true,
	})

	// 14) 落库：query_trace（溯源）+ assistant 消息（历史回溯）。
	persistSQLGenTrace(db, traceID, question, result, totalRows, totalDuration, "success", "")
	saveAssistantMessage(db, sessionID, "assistant", conclusion,
		marshalJSON(buildSQLGenBlocks(result, merged, sqlCtx, conclusion)),
		skillIDSQLGen, parsePathSQLGen, traceID,
		marshalJSON(sqlGenMessageMeta(result, sqlCtx, totalRows, totalDuration)))
}

// finishSQLError 统一处理 sqlgen 管线中途失败：发 error + done(ok=false)，并落库失败痕迹。
func finishSQLError(c *gin.Context, db *gorm.DB, sessionID, traceID, nlText, msg string) {
	logger.Warnf("[assistant] sql_gen 处理失败: %s", msg)
	sseSend(c, "error", gin.H{"message": msg, "trace_id": traceID})
	sseSend(c, "done", gin.H{
		"trace_id":   traceID,
		"session_id": sessionID,
		"skill_id":   skillIDSQLGen,
		"parse_path": parsePathSQLGen,
		"matched":    false,
		"ok":         false,
	})
	persistSQLGenTrace(db, traceID, nlText, nil, 0, 0, "failed", msg)
	saveAssistantMessage(db, sessionID, "assistant", "抱歉，处理该问题时出现错误："+msg,
		"", skillIDSQLGen, parsePathSQLGen, traceID, "")
}

// ---------------------------------------------------------------------------
// sqlgen 载荷构建辅助
// ---------------------------------------------------------------------------

// buildSQLStepsPayload 构建 sql 块的 steps 数据：{purpose, sql, datasource_id, datasource_label}。
func buildSQLStepsPayload(steps []sqlgen.SQLStep, ctx *sqlgen.SQLGenContext) []map[string]interface{} {
	dsName := map[uint]string{}
	if ctx != nil {
		for _, ds := range ctx.Datasources {
			dsName[ds.ID] = ds.Name
		}
	}
	out := make([]map[string]interface{}, 0, len(steps))
	for _, s := range steps {
		label := dsName[s.DatasourceID]
		if strings.TrimSpace(label) == "" {
			label = "数据源 #" + strconv.FormatUint(uint64(s.DatasourceID), 10)
		}
		out = append(out, map[string]interface{}{
			"purpose":          s.Purpose,
			"sql":              s.SQL,
			"datasource_id":    s.DatasourceID,
			"datasource_label": label,
		})
	}
	return out
}

// buildTableBlock 构建 table 块：payload={columns[], rows[][], records[], row_count}。
//
// columns 为列名数组、rows 为对齐的二维数组、records 为对象数组（列名→值），
// 与前端 TablePayload 契约一致。
func buildTableBlock(merged *sqlgen.MergedResult) map[string]interface{} {
	columns := make([]string, 0)
	rows := make([][]interface{}, 0)
	if merged != nil {
		for _, col := range merged.Columns {
			columns = append(columns, col.Name)
		}
		if merged.Rows != nil {
			rows = merged.Rows
		}
	}
	records := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		rec := make(map[string]interface{}, len(columns))
		for i, name := range columns {
			if i < len(row) {
				rec[name] = row[i]
			} else {
				rec[name] = nil
			}
		}
		records = append(records, rec)
	}
	return map[string]interface{}{
		"type": "table",
		"payload": map[string]interface{}{
			"columns":   columns,
			"rows":      rows,
			"records":   records,
			"row_count": len(rows),
		},
	}
}

// buildSQLGenBlocks 返回用于 blocks_json 落库的块集合（sql + table + 可选 chart + text），
// 与 SSE 下发的块保持一致，便于历史消息原样回放。
func buildSQLGenBlocks(result *sqlgen.SQLGenResult, merged *sqlgen.MergedResult,
	ctx *sqlgen.SQLGenContext, conclusion string) []map[string]interface{} {
	blocks := make([]map[string]interface{}, 0, 4)
	blocks = append(blocks, map[string]interface{}{
		"type":    "sql",
		"payload": map[string]interface{}{"steps": buildSQLStepsPayload(result.Steps, ctx)},
	})
	blocks = append(blocks, buildTableBlock(merged))
	if chart := sqlgen.ChartSuggestion(merged); chart != nil {
		blocks = append(blocks, map[string]interface{}{"type": "chart", "payload": chart})
	}
	blocks = append(blocks, map[string]interface{}{
		"type":    "text",
		"payload": map[string]interface{}{"content": conclusion, "text": conclusion},
	})
	return blocks
}

// sqlGenMessageMeta 提炼 assistant_message.meta_json 的轻量元数据（不含结果行）。
func sqlGenMessageMeta(result *sqlgen.SQLGenResult, ctx *sqlgen.SQLGenContext,
	totalRows int, totalDuration int64) map[string]interface{} {
	return map[string]interface{}{
		"skill_id":       skillIDSQLGen,
		"parse_path":     parsePathSQLGen,
		"matched":        true,
		"step_count":     len(result.Steps),
		"merge_strategy": result.MergeStrategy,
		"llm_calls":      2,
		"row_count":      totalRows,
		"duration_ms":    totalDuration,
		"datasource_ids": ctx.DatasourceIDs(),
		"sql_hash":       computeSQLHash(result.Steps),
	}
}

// buildAutoConclusion 当 LLM 未提供有效结论时，自动生成简要结论。
func buildAutoConclusion(merged *sqlgen.MergedResult, steps []sqlgen.SQLStep) string {
	rows := 0
	if merged != nil {
		rows = merged.RowCount
	}
	return fmt.Sprintf("查询返回 %d 行数据，涉及 %d 条 SQL 查询。", rows, len(steps))
}

// computeSQLHash 计算所有 SQL 的 sha256 哈希（截断为 16 字符，用于溯源比对）。
func computeSQLHash(steps []sqlgen.SQLStep) string {
	h := sha256.New()
	for _, s := range steps {
		h.Write([]byte(s.SQL))
		h.Write([]byte{'\n'})
	}
	sum := hex.EncodeToString(h.Sum(nil))
	if len(sum) > 16 {
		sum = sum[:16]
	}
	return sum
}

// allStepsFailed 报告是否存在步骤且全部执行失败（单步错误记录在 StepResult.Error）。
func allStepsFailed(results []sqlgen.StepResult) bool {
	if len(results) == 0 {
		return false
	}
	for _, sr := range results {
		if strings.TrimSpace(sr.Error) == "" {
			return false
		}
	}
	return true
}

// firstStepError 返回首个非空的步骤错误信息。
func firstStepError(results []sqlgen.StepResult) string {
	for _, sr := range results {
		if msg := strings.TrimSpace(sr.Error); msg != "" {
			return msg
		}
	}
	return "未知执行错误"
}

// ontologyIDString 将本体 ID 转为 BuildContext 所需的十进制字符串（0 返回空串）。
func ontologyIDString(id uint) string {
	if id == 0 {
		return ""
	}
	return strconv.FormatUint(uint64(id), 10)
}

// resolveAssistantOntologyID 归一请求中的 ontology_id：非零直接用；
// 为零/缺失时回退到默认已发布本体（无则取最小 ID 本体），复刻原 planner.ResolveOntologyID 语义。
func resolveAssistantOntologyID(db *gorm.DB, raw interface{}) uint {
	if u, ok := toUint(raw); ok && u > 0 {
		return u
	}
	if db == nil {
		return 0
	}
	var def model.OntDefinition
	if err := db.Where("status = ?", "published").Order("id asc").First(&def).Error; err == nil {
		return def.ID
	}
	if err := db.Order("id asc").First(&def).Error; err == nil {
		return def.ID
	}
	return 0
}

// persistSQLGenTrace 写入一次 sqlgen 问答的 query_trace（parse_path=sql_gen，便于全链路溯源）。
func persistSQLGenTrace(db *gorm.DB, traceID, nlText string, result *sqlgen.SQLGenResult,
	totalRows int, totalDuration int64, status, errMsg string) {
	if db == nil {
		return
	}
	t := &model.QueryTrace{
		TraceID:         traceID,
		NLText:          nlText,
		ParsePath:       parsePathSQLGen,
		Status:          status,
		ResultCount:     totalRows,
		ExecutionTimeMs: totalDuration,
		CreatedAt:       time.Now(),
	}
	if result != nil && len(result.Steps) > 0 {
		t.TranslatedSQL = result.Steps[0].SQL
		t.DataSourceID = result.Steps[0].DatasourceID
		t.IntentJSON = marshalJSON(result.Steps)
		t.OntologyQueryJSON = marshalJSON(map[string]interface{}{
			"merge_strategy": result.MergeStrategy,
			"merge_keys":     result.MergeKeys,
		})
		t.Explanation = truncate(result.Conclusion, 4000)
	}
	t.ErrorMessage = truncate(errMsg, 1000)
	if err := db.Create(t).Error; err != nil {
		logger.Warnf("[assistant] 写入 sql_gen query_trace 失败: %v", err)
	}
}

// ---------------------------------------------------------------------------
// 会话与消息持久化
// ---------------------------------------------------------------------------

// ensureAssistantSession 加载或创建会话；新建时以首条用户消息截取标题。
func ensureAssistantSession(db *gorm.DB, sessionID string, ontologyID uint, firstText string) {
	if db == nil || strings.TrimSpace(sessionID) == "" {
		return
	}
	var n int64
	if err := db.Model(&model.AssistantSession{}).Where("session_id = ?", sessionID).Count(&n).Error; err != nil {
		logger.Warnf("[assistant] 查询会话失败: %v", err)
		return
	}
	if n > 0 {
		return
	}
	title := strings.TrimSpace(firstText)
	if r := []rune(title); len(r) > 30 {
		title = string(r[:30]) + "…"
	}
	now := time.Now()
	s := model.AssistantSession{
		SessionID:     sessionID,
		Title:         title,
		OntologyID:    ontologyID,
		MessageCount:  0,
		LastMessageAt: now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := db.Create(&s).Error; err != nil {
		logger.Warnf("[assistant] 创建会话失败: %v", err)
	}
}

// saveAssistantMessage 落库一条消息并更新会话计数/最后活跃时间。
func saveAssistantMessage(db *gorm.DB, sessionID, role, content, blocksJSON, skillID, parsePath, traceID, metaJSON string) {
	if db == nil {
		return
	}
	now := time.Now()
	m := model.AssistantMessage{
		SessionID:  sessionID,
		Role:       role,
		Content:    content,
		BlocksJSON: blocksJSON,
		SkillID:    skillID,
		ParsePath:  parsePath,
		TraceID:    traceID,
		MetaJSON:   metaJSON,
		CreatedAt:  now,
	}
	if err := db.Create(&m).Error; err != nil {
		logger.Warnf("[assistant] 写入消息失败: %v", err)
		return
	}
	db.Model(&model.AssistantSession{}).Where("session_id = ?", sessionID).Updates(map[string]interface{}{
		"message_count":   gorm.Expr("message_count + 1"),
		"last_message_at": now,
		"updated_at":      now,
	})
}

// ---------------------------------------------------------------------------
// GET /api/v1/assistant/skills —— 技能清单
// ---------------------------------------------------------------------------

// AssistantSkills 返回助手能力清单。
//
// sqlgen 单一路径下不再有「多技能」概念，此端点保留路由与响应形态以兼容前端，
// 返回空清单（技能体系已随旧分派路径退役）。
func AssistantSkills(c *gin.Context) {
	ok(c, gin.H{
		"skills":   []interface{}{},
		"count":    0,
		"trace_id": traceIDOf(c),
	})
}

// ---------------------------------------------------------------------------
// GET /api/v1/assistant/sessions —— 会话列表
// ---------------------------------------------------------------------------

// AssistantSessions 分页返回会话列表（按最后活跃时间倒序）。
func AssistantSessions(c *gin.Context) {
	current, size := parsePage(c)
	var total int64
	var list []model.AssistantSession
	db := DB()
	if err := db.Model(&model.AssistantSession{}).Count(&total).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}
	if err := db.Order("last_message_at desc, id desc").
		Offset((current - 1) * size).Limit(size).Find(&list).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}
	middleware.SuccessPage(c, list, total, current, size)
}

// ---------------------------------------------------------------------------
// GET /api/v1/assistant/sessions/:id/messages —— 会话历史
// ---------------------------------------------------------------------------

// assistantMessageView 会话消息对外视图（blocks_json 解析为对象）。
type assistantMessageView struct {
	ID        uint        `json:"id"`
	Role      string      `json:"role"`
	Content   string      `json:"content"`
	Blocks    interface{} `json:"blocks"`
	SkillID   string      `json:"skill_id"`
	ParsePath string      `json:"parse_path"`
	TraceID   string      `json:"trace_id"`
	Meta      interface{} `json:"meta"`
	CreatedAt time.Time   `json:"created_at"`
}

// AssistantSessionMessages 返回指定会话的全部消息（:id = session_id 字符串）。
func AssistantSessionMessages(c *gin.Context) {
	sid := strings.TrimSpace(c.Param("id"))
	if sid == "" {
		fail(c, errs.ParamError("缺少路径参数: id"))
		return
	}
	var msgs []model.AssistantMessage
	if err := DB().Where("session_id = ?", sid).Order("id asc").Find(&msgs).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}
	views := make([]assistantMessageView, 0, len(msgs))
	for _, m := range msgs {
		views = append(views, assistantMessageView{
			ID:        m.ID,
			Role:      m.Role,
			Content:   m.Content,
			Blocks:    parseJSONField(m.BlocksJSON),
			SkillID:   m.SkillID,
			ParsePath: m.ParsePath,
			TraceID:   m.TraceID,
			Meta:      normalizeMessageMeta(parseJSONField(m.MetaJSON)),
			CreatedAt: m.CreatedAt,
		})
	}
	ok(c, gin.H{
		"session_id": sid,
		"messages":   views,
		"count":      len(views),
		"trace_id":   traceIDOf(c),
	})
}

// ---------------------------------------------------------------------------
// meta 契约防御：hop_count 恒为单一整数
// ---------------------------------------------------------------------------

// asHopCount 将任意来源的 hop_count 归一为单一非负整数（D2 防御）。
//
// 契约：前端 MessageMetaBar 以 Number(m.hop_count) 渲染，types/assistant.ts 声明为 number；
// 若历史上曾写入数组（如 [3,4]）则 Number() 得 NaN。本函数对数组取末位（最终跳数）、
// 对字符串/浮点做整数归一、无法解析时返回 0，保证对外永远是单一整数。
func asHopCount(v interface{}) int {
	switch x := v.(type) {
	case nil:
		return 0
	case int:
		return nonNegInt(x)
	case int32:
		return nonNegInt(int(x))
	case int64:
		return nonNegInt(int(x))
	case float32:
		return nonNegInt(int(x))
	case float64:
		return nonNegInt(int(x))
	case json.Number:
		if n, err := x.Int64(); err == nil {
			return nonNegInt(int(n))
		}
		return 0
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return 0
		}
		if n, err := strconv.Atoi(s); err == nil {
			return nonNegInt(n)
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return nonNegInt(int(f))
		}
		return 0
	case []interface{}:
		if len(x) == 0 {
			return 0
		}
		return asHopCount(x[len(x)-1]) // 数组取末位 = 最终成功跳数
	case []int:
		if len(x) == 0 {
			return 0
		}
		return nonNegInt(x[len(x)-1])
	default:
		return 0
	}
}

// nonNegInt 钳制负数为 0（跳数不可能为负）。
func nonNegInt(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

// normalizeMessageMeta 将已落库的消息 meta 中的 hop_count/llm_calls/sub_intent_count 归一为整数。
//
// 仅对历史数据做防御（旧版本可能因重试场景写入数组）；非 map 结构原样返回，不新增也不删除字段。
func normalizeMessageMeta(meta interface{}) interface{} {
	m, isMap := meta.(map[string]interface{})
	if !isMap {
		return meta
	}
	for _, k := range []string{"hop_count", "llm_calls", "sub_intent_count"} {
		if v, exists := m[k]; exists {
			m[k] = asHopCount(v)
		}
	}
	return m
}

// orEmptyStrSlice 保证字符串切片序列化为空数组而非 null（契约字段稳定）。
func orEmptyStrSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
