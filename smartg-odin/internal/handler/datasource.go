package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"smartg-odin/internal/common/errs"
	"smartg-odin/internal/common/logger"
	"smartg-odin/internal/ds/connector"
	"smartg-odin/internal/middleware"
	"smartg-odin/internal/model"
)

// datasourcePayload 数据源创建 / 更新的请求体。
// Password 与 SSHKey 在模型层被 json:"-" 屏蔽，因此需要独立的入参结构。
type datasourcePayload struct {
	Code            string          `json:"code"`
	Name            string          `json:"name"`
	Type            string          `json:"type"`
	Host            string          `json:"host"`
	Port            int             `json:"port"`
	DatabaseName    string          `json:"database_name"`
	Username        string          `json:"username"`
	Password        string          `json:"password"`
	SchemaName      string          `json:"schema_name"`
	ParamsJSON      json.RawMessage `json:"params_json"`
	PoolMaxOpen     int             `json:"pool_max_open"`
	PoolMaxIdle     int             `json:"pool_max_idle"`
	PoolMaxLifetime int             `json:"pool_max_lifetime"`
	SSLMode         string          `json:"ssl_mode"`
	SSHEnabled      bool            `json:"ssh_enabled"`
	SSHHost         string          `json:"ssh_host"`
	SSHPort         int             `json:"ssh_port"`
	SSHUser         string          `json:"ssh_user"`
	SSHKey          string          `json:"ssh_key"`
	Status          string          `json:"status"`
	Tags            json.RawMessage `json:"tags"`
	Description     string          `json:"description"`
}

// datasourceView 数据源响应视图，在模型之上补充统计与最近一次测试信息。
type datasourceView struct {
	model.DataSource
	TableCount int64               `json:"table_count"`
	TagsList   []string            `json:"tags_list"`
	Supported  bool                `json:"connector_supported"`
	LastTest   *datasourceTestView `json:"last_test,omitempty"`
}

// datasourceDetail 数据源详情视图：在列表视图基础上附带已缓存的表清单。
type datasourceDetail struct {
	datasourceView
	Tables []model.SchemaTable `json:"tables"`
}

// datasourceTestView 连通性测试记录视图。
type datasourceTestView struct {
	ID        uint      `json:"id"`
	Status    string    `json:"status"`
	LatencyMs int64     `json:"latency_ms"`
	Message   string    `json:"message"`
	TestedAt  time.Time `json:"tested_at"`
}

// datasourceUpdatableColumns 允许通过 PUT 更新的列。
var datasourceUpdatableColumns = []string{
	"code", "name", "type", "host", "port", "database_name", "username", "password",
	"schema_name", "params_json", "pool_max_open", "pool_max_idle", "pool_max_lifetime",
	"ssl_mode", "ssh_enabled", "ssh_host", "ssh_port", "ssh_user", "ssh_key",
	"status", "tags", "description",
}

// ListDatasources GET /api/v1/datasources?current=1&size=20&type=&status=&keyword=
// 分页返回数据源列表，支持按类型、状态、关键字筛选。
func ListDatasources(c *gin.Context) {
	current, size := parsePage(c)

	filter := func() *gorm.DB {
		tx := DB().Model(&model.DataSource{})
		if t := strings.TrimSpace(c.Query("type")); t != "" {
			tx = tx.Where("type = ?", t)
		}
		if s := strings.TrimSpace(c.Query("status")); s != "" {
			tx = tx.Where("status = ?", s)
		}
		if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
			like := "%" + kw + "%"
			tx = tx.Where("name LIKE ? OR code LIKE ? OR host LIKE ? OR database_name LIKE ?", like, like, like, like)
		}
		return tx
	}

	var total int64
	if err := filter().Count(&total).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	var rows []model.DataSource
	if err := filter().Order("id ASC").Offset((current - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	ids := make([]uint, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	tableCounts := countTablesByDatasource(ids)
	lastTests := loadLastTests(ids)
	supported := connector.SupportedTypes()

	views := make([]datasourceView, 0, len(rows))
	for _, r := range rows {
		views = append(views, buildDatasourceView(r, tableCounts[r.ID], lastTests[r.ID], supported))
	}

	middleware.SuccessPage(c, views, total, current, size)
}

// GetDatasource GET /api/v1/datasources/:id
// 返回单个数据源详情（密码字段不会出现在响应中）。
func GetDatasource(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var ds model.DataSource
	if err := DB().First(&ds, id).Error; err != nil {
		fail(c, mapError(err, fmt.Sprintf("数据源不存在: %d", id)))
		return
	}

	tableCounts := countTablesByDatasource([]uint{ds.ID})
	lastTests := loadLastTests([]uint{ds.ID})
	view := buildDatasourceView(ds, tableCounts[ds.ID], lastTests[ds.ID], connector.SupportedTypes())

	// 附加该数据源下已缓存的表清单，方便前端一次拉取
	var tables []model.SchemaTable
	DB().Where("datasource_id = ?", ds.ID).Order("table_name ASC").Find(&tables)
	ok(c, datasourceDetail{
		datasourceView: view,
		Tables:         tables,
	})
}

// CreateDatasource POST /api/v1/datasources
// 创建数据源，返回创建后的完整对象。
func CreateDatasource(c *gin.Context) {
	var req datasourcePayload
	if appErr := bindNormalized(c, &req); appErr != nil {
		fail(c, appErr)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Type = strings.ToLower(strings.TrimSpace(req.Type))
	if req.Name == "" {
		fail(c, errs.ParamError("数据源名称 name 不能为空"))
		return
	}
	if req.Type == "" {
		fail(c, errs.ParamError("数据源类型 type 不能为空"))
		return
	}

	code := strings.TrimSpace(req.Code)
	exists := func(candidate string) bool {
		var n int64
		DB().Model(&model.DataSource{}).Where("code = ?", candidate).Count(&n)
		return n > 0
	}
	if code == "" {
		code = uniqueCode("ds", req.Name, exists)
	} else if exists(code) {
		fail(c, errs.ParamError("数据源编码已存在: "+code))
		return
	}

	ds := model.DataSource{
		Code:            code,
		Name:            req.Name,
		Type:            req.Type,
		Host:            strings.TrimSpace(req.Host),
		Port:            req.Port,
		DatabaseName:    strings.TrimSpace(req.DatabaseName),
		Username:        strings.TrimSpace(req.Username),
		Password:        req.Password,
		SchemaName:      strings.TrimSpace(req.SchemaName),
		ParamsJSON:      string(req.ParamsJSON),
		PoolMaxOpen:     req.PoolMaxOpen,
		PoolMaxIdle:     req.PoolMaxIdle,
		PoolMaxLifetime: req.PoolMaxLifetime,
		SSLMode:         firstNonEmpty(strings.TrimSpace(req.SSLMode), "disable"),
		SSHEnabled:      req.SSHEnabled,
		SSHHost:         req.SSHHost,
		SSHPort:         req.SSHPort,
		SSHUser:         req.SSHUser,
		SSHKey:          req.SSHKey,
		Status:          firstNonEmpty(strings.TrimSpace(req.Status), "inactive"),
		Tags:            tagsToString(rawToInterface(req.Tags)),
		Description:     req.Description,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	applyDatasourceDefaults(&ds)

	if err := DB().Create(&ds).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	writeAudit(c, "create", "datasource", ds.ID, gin.H{"code": ds.Code, "name": ds.Name, "type": ds.Type})
	logger.Infof("[datasource] created id=%d code=%s type=%s", ds.ID, ds.Code, ds.Type)

	view := buildDatasourceView(ds, 0, nil, connector.SupportedTypes())
	okMsg(c, view, "创建成功")
}

// UpdateDatasource PUT /api/v1/datasources/:id
// 部分更新数据源；未提交的字段保持原值，密码为空时不覆盖。
func UpdateDatasource(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var ds model.DataSource
	if err := DB().First(&ds, id).Error; err != nil {
		fail(c, mapError(err, fmt.Sprintf("数据源不存在: %d", id)))
		return
	}

	body, appErr := bindMap(c)
	if appErr != nil {
		fail(c, appErr)
		return
	}
	if len(body) == 0 {
		fail(c, errs.ParamError("没有需要更新的字段"))
		return
	}

	updates := pickUpdates(body, datasourceUpdatableColumns...)
	if tags, hasTags := body["tags"]; hasTags {
		updates["tags"] = tagsToString(tags)
	}
	if pwd, hasPwd := body["password"]; hasPwd {
		if s, isStr := pwd.(string); isStr && s == "" {
			delete(updates, "password") // 空密码视为不修改
		}
	}
	if sshKey, hasKey := body["ssh_key"]; hasKey {
		if s, isStr := sshKey.(string); isStr && s == "" {
			delete(updates, "ssh_key")
		}
	}
	if code, hasCode := updates["code"]; hasCode {
		if s, isStr := code.(string); isStr && s != "" && s != ds.Code {
			var n int64
			DB().Model(&model.DataSource{}).Where("code = ? AND id <> ?", s, id).Count(&n)
			if n > 0 {
				fail(c, errs.ParamError("数据源编码已存在: "+s))
				return
			}
		}
	}
	if len(updates) == 0 {
		fail(c, errs.ParamError("没有可识别的更新字段"))
		return
	}
	updates["updated_at"] = time.Now()

	if err := DB().Model(&model.DataSource{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	if err := DB().First(&ds, id).Error; err != nil {
		fail(c, mapError(err, "数据源不存在"))
		return
	}
	writeAudit(c, "update", "datasource", ds.ID, gin.H{"fields": keysOf(updates)})
	okMsg(c, buildDatasourceView(ds, countTablesByDatasource([]uint{id})[id],
		loadLastTests([]uint{id})[id], connector.SupportedTypes()), "更新成功")
}

// DeleteDatasource DELETE /api/v1/datasources/:id
// 软删除数据源（GORM DeletedAt），同时清理其 schema 缓存与映射配置。
func DeleteDatasource(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var ds model.DataSource
	if err := DB().First(&ds, id).Error; err != nil {
		fail(c, mapError(err, fmt.Sprintf("数据源不存在: %d", id)))
		return
	}

	tx := DB().Begin()
	if tx.Error != nil {
		fail(c, errs.DBError(tx.Error))
		return
	}
	if err := tx.Delete(&model.DataSource{}, id).Error; err != nil {
		tx.Rollback()
		fail(c, errs.DBError(err))
		return
	}
	// schema 缓存为派生数据，随数据源一并清理（硬删除）
	tx.Where("datasource_id = ?", id).Delete(&model.SchemaColumn{})
	tx.Where("datasource_id = ?", id).Delete(&model.SchemaForeignKey{})
	tx.Where("datasource_id = ?", id).Delete(&model.SchemaIndex{})
	tx.Where("datasource_id = ?", id).Delete(&model.SchemaTable{})
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		fail(c, errs.DBError(err))
		return
	}

	writeAudit(c, "delete", "datasource", id, gin.H{"code": ds.Code, "name": ds.Name})
	okMsg(c, gin.H{"id": id}, "删除成功")
}

// TestDatasource POST /api/v1/datasources/:id/test
// 使用真实连接器测试连通性，写入测试记录并回写数据源状态。
func TestDatasource(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	ds, appErr := loadDatasource(id)
	if appErr != nil {
		fail(c, appErr)
		return
	}

	result, appErr := runConnectionTest(ds)
	// 无论成功失败都返回测试结果，由前端展示详情
	if appErr != nil {
		fail(c, appErr)
		return
	}

	writeAudit(c, "test", "datasource", id, gin.H{"success": result.Success, "latency_ms": result.LatencyMs})
	okMsg(c, result, result.Message)
}

// runConnectionTest 执行连通性测试并持久化结果。
func runConnectionTest(ds *model.DataSource) (*connector.TestResult, *errs.AppError) {
	started := time.Now()

	record := func(success bool, latency int64, message string) {
		testLog := model.DataSourceTest{
			DataSourceID: ds.ID,
			Status:       boolToStatus(success),
			LatencyMs:    latency,
			Message:      truncate(message, 2000),
			TestedAt:     time.Now(),
		}
		if err := DB().Create(&testLog).Error; err != nil {
			logger.Warnf("[datasource] 写入测试记录失败 id=%d err=%v", ds.ID, err)
		}
		newStatus := "error"
		if success {
			newStatus = "active"
		}
		if err := DB().Model(&model.DataSource{}).Where("id = ?", ds.ID).
			Updates(map[string]interface{}{"status": newStatus, "updated_at": time.Now()}).Error; err != nil {
			logger.Warnf("[datasource] 回写状态失败 id=%d err=%v", ds.ID, err)
		}
	}

	conn, appErr := getConnector(ds.Type)
	if appErr != nil {
		latency := time.Since(started).Milliseconds()
		record(false, latency, appErr.Message)
		return &connector.TestResult{Success: false, LatencyMs: latency, Message: appErr.Message}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectorTimeout)
	defer cancel()

	res, err := conn.Test(ctx, connectionConfig(ds))
	if res == nil {
		res = &connector.TestResult{}
	}
	if err != nil {
		res.Success = false
		if res.Message == "" {
			res.Message = err.Error()
		}
	}
	if res.LatencyMs == 0 {
		res.LatencyMs = time.Since(started).Milliseconds()
	}
	if res.Message == "" {
		if res.Success {
			res.Message = "连接正常"
		} else {
			res.Message = "连接失败"
		}
	}

	record(res.Success, res.LatencyMs, res.Message)
	logger.Infof("[datasource] test id=%d type=%s success=%v latency=%dms",
		ds.ID, ds.Type, res.Success, res.LatencyMs)
	return res, nil
}

// GetDatasourceSchema GET /api/v1/datasources/:id/schema
// 缓存优先返回数据源结构：
//   - 缓存非空且未强制刷新时，直接返回缓存（快速路径，cached=true）；
//   - 缓存为空（如启动时 docker 未就绪导致 RefreshAllSchemas 内省失败）时，
//     自动调用 syncSchema 内省并持久化后返回，使 Schema 对话框首次打开即有数据；
//   - 自动内省失败（如数据源不可达）时优雅降级，返回空 tables 但不报错（code=0），
//     避免前端首次打开即弹错误。
//
// 需要强制重新内省时可传 refresh=true；差量刷新请使用 POST /schema/refresh。
func GetDatasourceSchema(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	ds, appErr := loadDatasource(id)
	if appErr != nil {
		fail(c, appErr)
		return
	}

	started := time.Now()

	// refresh=true 时强制重新内省；默认（false）走缓存优先策略。
	forceRefresh := strings.EqualFold(strings.TrimSpace(c.DefaultQuery("refresh", "false")), "true")

	// 缓存中的表数量，用于判断缓存是否为空。
	cachedCount := countTablesByDatasource([]uint{ds.ID})[ds.ID]

	// 快速路径：缓存非空且未强制刷新，直接返回缓存。
	if cachedCount > 0 && !forceRefresh {
		resp, _ := buildSchemaResponse(ds.ID, true, time.Since(started).Milliseconds())
		ok(c, resp)
		return
	}

	// 缓存为空或强制刷新：尝试内省填充缓存。
	if _, syncErr := syncSchema(ds); syncErr != nil {
		// 内省失败（如 docker 未运行 / 数据源不可达）时优雅降级：
		// 记录告警并返回当前缓存（可能为空），不向前端抛错。
		logger.Warnf("[datasource] schema 自动内省失败 id=%d err=%s，返回现有缓存", ds.ID, syncErr.Message)
		resp, _ := buildSchemaResponse(ds.ID, true, time.Since(started).Milliseconds())
		ok(c, resp)
		return
	}

	resp, stats := buildSchemaResponse(ds.ID, false, time.Since(started).Milliseconds())
	writeAudit(c, "introspect", "datasource", ds.ID, gin.H{"tables": stats.Tables, "columns": stats.Columns})
	okMsg(c, resp, fmt.Sprintf("内省完成：%d 张表 / %d 个字段", stats.Tables, stats.Columns))
}

// RefreshDatasourceSchema POST /api/v1/datasources/:id/schema/refresh
// 增量刷新：内省真实结构并与缓存做 diff（新增 / 删除 / 变更的表与列），
// 覆盖写入缓存后返回刷新后的完整 schema 与差异摘要；
// 对删除 / 变更的列，会同步更新引用它们的映射项校验状态。
func RefreshDatasourceSchema(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	ds, appErr := loadDatasource(id)
	if appErr != nil {
		fail(c, appErr)
		return
	}

	started := time.Now()

	// 1. 刷新前的旧快照（table_name -> column_name -> data_type）
	old := snapshotSchema(ds.ID)

	// 2. 内省并覆盖写入缓存
	info, appErr := syncSchema(ds)
	if appErr != nil {
		fail(c, appErr)
		return
	}

	// 3. 计算 diff
	diff := computeSchemaDiff(old, info)

	// 4. 将删除 / 变更列同步到映射校验状态
	applyDiffToMappings(ds.ID, diff)

	resp, stats := buildSchemaResponse(ds.ID, false, time.Since(started).Milliseconds())
	writeAudit(c, "refresh_schema", "datasource", ds.ID, gin.H{
		"tables": stats.Tables, "columns": stats.Columns,
		"added_tables": len(diff.AddedTables), "removed_tables": len(diff.RemovedTables),
		"added_columns": len(diff.AddedColumns), "removed_columns": len(diff.RemovedColumns),
		"changed_columns": len(diff.ChangedColumns),
	})
	logger.Infof("[datasource] refresh schema id=%d +%d表 -%d表 +%d列 -%d列 ~%d列",
		ds.ID, len(diff.AddedTables), len(diff.RemovedTables),
		len(diff.AddedColumns), len(diff.RemovedColumns), len(diff.ChangedColumns))

	okMsg(c, gin.H{"schema": resp, "diff": diff},
		fmt.Sprintf("刷新完成：%d 张表 / %d 个字段", stats.Tables, stats.Columns))
}

// schemaColumnView 字段响应视图（含主键 id，供前端精确关联）。
type schemaColumnView struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	DataType        string `json:"data_type"`
	IsNullable      bool   `json:"is_nullable"`
	IsPrimaryKey    bool   `json:"is_primary_key"`
	DefaultValue    string `json:"default_value"`
	Comment         string `json:"comment"`
	OrdinalPosition int    `json:"ordinal_position"`
}

// schemaForeignKeyView 外键响应视图。
type schemaForeignKeyView struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Column    string `json:"column"`
	RefTable  string `json:"ref_table"`
	RefColumn string `json:"ref_column"`
}

// schemaIndexView 索引响应视图。
type schemaIndexView struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Columns   string `json:"columns"`
	IsUnique  bool   `json:"is_unique"`
	IsPrimary bool   `json:"is_primary"`
}

// schemaTableView 表响应视图，内嵌该表的全部字段 / 外键 / 索引。
type schemaTableView struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Schema      string                 `json:"schema"`
	Type        string                 `json:"type"`
	Comment     string                 `json:"comment"`
	Columns     []schemaColumnView     `json:"columns"`
	ForeignKeys []schemaForeignKeyView `json:"foreign_keys"`
	Indexes     []schemaIndexView      `json:"indexes"`
}

// schemaStats schema 统计信息。
type schemaStats struct {
	Tables      int `json:"tables"`
	Columns     int `json:"columns"`
	PrimaryKeys int `json:"primary_keys"`
}

// buildSchemaResponse 从缓存表组装完整的 schema 响应体与统计信息。
// tables 位于顶层，每张表内嵌其全部字段（按 ordinal_position 排序）。
func buildSchemaResponse(datasourceID uint, cached bool, durationMs int64) (gin.H, schemaStats) {
	tables, stats := buildSchemaViews(datasourceID)
	return gin.H{
		"datasource_id": datasourceID,
		"cached":        cached,
		"tables":        tables,
		"stats":         stats,
		"table_count":   stats.Tables,
		"column_count":  stats.Columns,
		"duration_ms":   durationMs,
	}, stats
}

// buildSchemaViews 从 schema_table / schema_column / schema_foreign_key / schema_index
// 缓存表读取并组装带 id 的表视图，确保每张表内嵌其全部字段（而非仅主键列）。
func buildSchemaViews(datasourceID uint) ([]schemaTableView, schemaStats) {
	views := []schemaTableView{}
	stats := schemaStats{}

	var tables []model.SchemaTable
	DB().Where("datasource_id = ?", datasourceID).Order("table_name ASC").Find(&tables)
	if len(tables) == 0 {
		return views, stats
	}

	var columns []model.SchemaColumn
	DB().Where("datasource_id = ?", datasourceID).Order("table_id ASC, ordinal_pos ASC").Find(&columns)
	var fks []model.SchemaForeignKey
	DB().Where("datasource_id = ?", datasourceID).Find(&fks)
	var idxList []model.SchemaIndex
	DB().Where("datasource_id = ?", datasourceID).Find(&idxList)

	colsByTable := map[uint][]schemaColumnView{}
	for _, col := range columns {
		colsByTable[col.TableID] = append(colsByTable[col.TableID], schemaColumnView{
			ID:              col.ID,
			Name:            col.ColumnName,
			DataType:        col.DataType,
			IsNullable:      col.IsNullable,
			IsPrimaryKey:    col.IsPrimaryKey,
			DefaultValue:    col.DefaultValue,
			Comment:         col.Comment,
			OrdinalPosition: col.OrdinalPos,
		})
		if col.IsPrimaryKey {
			stats.PrimaryKeys++
		}
	}
	fksByTable := map[uint][]schemaForeignKeyView{}
	for _, fk := range fks {
		fksByTable[fk.TableID] = append(fksByTable[fk.TableID], schemaForeignKeyView{
			ID: fk.ID, Name: fk.ConstraintName, Column: fk.ColumnName,
			RefTable: fk.RefTable, RefColumn: fk.RefColumn,
		})
	}
	idxByTable := map[uint][]schemaIndexView{}
	for _, idx := range idxList {
		idxByTable[idx.TableID] = append(idxByTable[idx.TableID], schemaIndexView{
			ID: idx.ID, Name: idx.IndexName, Columns: idx.Columns,
			IsUnique: idx.IsUnique, IsPrimary: idx.IsPrimary,
		})
	}

	for _, t := range tables {
		cols := colsByTable[t.ID]
		if cols == nil {
			cols = []schemaColumnView{}
		}
		fkViews := fksByTable[t.ID]
		if fkViews == nil {
			fkViews = []schemaForeignKeyView{}
		}
		idxViews := idxByTable[t.ID]
		if idxViews == nil {
			idxViews = []schemaIndexView{}
		}
		stats.Tables++
		stats.Columns += len(cols)
		views = append(views, schemaTableView{
			ID:          t.ID,
			Name:        t.TableName,
			Schema:      t.SchemaName,
			Type:        firstNonEmpty(t.TableType, "TABLE"),
			Comment:     t.Comment,
			Columns:     cols,
			ForeignKeys: fkViews,
			Indexes:     idxViews,
		})
	}
	return views, stats
}

// RefreshAllSchemas 启动时遍历所有 active 数据源并尝试内省 schema，
// 成功则刷新缓存，失败仅记录告警不中断启动（docker 可能未运行）。
func RefreshAllSchemas() {
	var sources []model.DataSource
	if err := DB().Where("status = ?", "active").Find(&sources).Error; err != nil {
		logger.Warnf("[startup] load active datasources failed: %v", err)
		return
	}
	for i := range sources {
		ds := &sources[i]
		_, appErr := syncSchema(ds)
		if appErr != nil {
			logger.Warnf("[startup] introspect datasource %s (id=%d) failed: %s", ds.Code, ds.ID, appErr.Message)
		} else {
			logger.Infof("[startup] introspect datasource %s (id=%d) OK", ds.Code, ds.ID)
		}
	}
}

// ---------------------------------------------------------------------------
// 内部辅助
// ---------------------------------------------------------------------------

// syncSchema 内省数据源结构并刷新元数据缓存，供数据源与映射模块复用。
func syncSchema(ds *model.DataSource) (*connector.SchemaInfo, *errs.AppError) {
	conn, appErr := getConnector(ds.Type)
	if appErr != nil {
		return nil, appErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectorTimeout)
	defer cancel()

	info, err := conn.Introspect(ctx, connectionConfig(ds))
	if err != nil {
		return nil, errs.ConnError(err)
	}
	if info == nil {
		info = &connector.SchemaInfo{Tables: []connector.TableInfo{}}
	}
	if info.Tables == nil {
		info.Tables = []connector.TableInfo{}
	}

	if appErr := persistSchema(ds.ID, info); appErr != nil {
		return nil, appErr
	}
	// 内省成功即认为连接可用
	DB().Model(&model.DataSource{}).Where("id = ? AND status <> ?", ds.ID, "active").
		Update("status", "active")
	return info, nil
}

// persistSchema 覆盖写入某个数据源的 schema 缓存（表 / 字段 / 外键 / 索引）。
func persistSchema(datasourceID uint, info *connector.SchemaInfo) *errs.AppError {
	tx := DB().Begin()
	if tx.Error != nil {
		return errs.DBError(tx.Error)
	}
	rollback := func(err error) *errs.AppError {
		tx.Rollback()
		return errs.DBError(err)
	}

	// 先清理旧缓存，保证与真实结构一致
	if err := tx.Where("datasource_id = ?", datasourceID).Delete(&model.SchemaColumn{}).Error; err != nil {
		return rollback(err)
	}
	if err := tx.Where("datasource_id = ?", datasourceID).Delete(&model.SchemaForeignKey{}).Error; err != nil {
		return rollback(err)
	}
	if err := tx.Where("datasource_id = ?", datasourceID).Delete(&model.SchemaIndex{}).Error; err != nil {
		return rollback(err)
	}
	if err := tx.Where("datasource_id = ?", datasourceID).Delete(&model.SchemaTable{}).Error; err != nil {
		return rollback(err)
	}

	for _, t := range info.Tables {
		table := model.SchemaTable{
			DataSourceID: datasourceID,
			SchemaName:   t.Schema,
			TableName:    t.Name,
			TableType:    firstNonEmpty(strings.ToUpper(t.Type), "TABLE"),
			Comment:      t.Comment,
		}
		if err := tx.Create(&table).Error; err != nil {
			return rollback(err)
		}

		if len(t.Columns) > 0 {
			columns := make([]model.SchemaColumn, 0, len(t.Columns))
			for _, col := range t.Columns {
				columns = append(columns, model.SchemaColumn{
					TableID:      table.ID,
					DataSourceID: datasourceID,
					ColumnName:   col.Name,
					DataType:     col.DataType,
					IsNullable:   col.IsNullable,
					IsPrimaryKey: col.IsPrimary,
					DefaultValue: col.Default,
					Comment:      col.Comment,
					OrdinalPos:   col.Position,
				})
			}
			if err := tx.Create(&columns).Error; err != nil {
				return rollback(err)
			}
		}

		if len(t.ForeignKeys) > 0 {
			fks := make([]model.SchemaForeignKey, 0, len(t.ForeignKeys))
			for _, fk := range t.ForeignKeys {
				fks = append(fks, model.SchemaForeignKey{
					TableID:        table.ID,
					DataSourceID:   datasourceID,
					ConstraintName: fk.Name,
					ColumnName:     fk.Column,
					RefTable:       fk.RefTable,
					RefColumn:      fk.RefColumn,
				})
			}
			if err := tx.Create(&fks).Error; err != nil {
				return rollback(err)
			}
		}

		if len(t.Indexes) > 0 {
			idxList := make([]model.SchemaIndex, 0, len(t.Indexes))
			for _, idx := range t.Indexes {
				idxList = append(idxList, model.SchemaIndex{
					TableID:      table.ID,
					DataSourceID: datasourceID,
					IndexName:    idx.Name,
					Columns:      idx.Columns,
					IsUnique:     idx.IsUnique,
					IsPrimary:    idx.IsPrimary,
				})
			}
			if err := tx.Create(&idxList).Error; err != nil {
				return rollback(err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return errs.DBError(err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Schema diff（增量刷新）
// ---------------------------------------------------------------------------

// diffColumn 描述某张表下的一个新增 / 删除列。
type diffColumn struct {
	Table  string `json:"table"`
	Column string `json:"column"`
}

// diffColumnChange 描述列的数据类型变更。
type diffColumnChange struct {
	Table   string `json:"table"`
	Column  string `json:"column"`
	OldType string `json:"old_type"`
	NewType string `json:"new_type"`
}

// schemaDiff 增量刷新的差异摘要。
type schemaDiff struct {
	AddedTables    []string           `json:"added_tables"`
	RemovedTables  []string           `json:"removed_tables"`
	AddedColumns   []diffColumn       `json:"added_columns"`
	RemovedColumns []diffColumn       `json:"removed_columns"`
	ChangedColumns []diffColumnChange `json:"changed_columns"`
}

func newSchemaDiff() *schemaDiff {
	return &schemaDiff{
		AddedTables:    []string{},
		RemovedTables:  []string{},
		AddedColumns:   []diffColumn{},
		RemovedColumns: []diffColumn{},
		ChangedColumns: []diffColumnChange{},
	}
}

// snapshotSchema 读取当前缓存，构造 table_name -> (column_name -> data_type) 快照。
func snapshotSchema(datasourceID uint) map[string]map[string]string {
	result := map[string]map[string]string{}
	var tables []model.SchemaTable
	DB().Where("datasource_id = ?", datasourceID).Find(&tables)
	if len(tables) == 0 {
		return result
	}
	idToName := map[uint]string{}
	for _, t := range tables {
		idToName[t.ID] = t.TableName
		result[t.TableName] = map[string]string{}
	}
	var columns []model.SchemaColumn
	DB().Where("datasource_id = ?", datasourceID).Find(&columns)
	for _, col := range columns {
		if name, ok := idToName[col.TableID]; ok {
			result[name][col.ColumnName] = col.DataType
		}
	}
	return result
}

// computeSchemaDiff 对比旧快照与实时内省结果，生成差异摘要（按 table_name + column_name 匹配）。
func computeSchemaDiff(old map[string]map[string]string, live *connector.SchemaInfo) *schemaDiff {
	diff := newSchemaDiff()

	liveTables := map[string]map[string]string{}
	if live != nil {
		for _, t := range live.Tables {
			cols := map[string]string{}
			for _, col := range t.Columns {
				cols[col.Name] = col.DataType
			}
			liveTables[t.Name] = cols
		}
	}

	// 新增 / 删除表
	for name := range liveTables {
		if _, ok := old[name]; !ok {
			diff.AddedTables = append(diff.AddedTables, name)
		}
	}
	for name := range old {
		if _, ok := liveTables[name]; !ok {
			diff.RemovedTables = append(diff.RemovedTables, name)
		}
	}
	sort.Strings(diff.AddedTables)
	sort.Strings(diff.RemovedTables)

	// 共有表的列级差异
	for name, liveCols := range liveTables {
		oldCols, ok := old[name]
		if !ok {
			continue
		}
		for col, typ := range liveCols {
			oldType, exists := oldCols[col]
			if !exists {
				diff.AddedColumns = append(diff.AddedColumns, diffColumn{Table: name, Column: col})
				continue
			}
			if !strings.EqualFold(strings.TrimSpace(oldType), strings.TrimSpace(typ)) {
				diff.ChangedColumns = append(diff.ChangedColumns, diffColumnChange{
					Table: name, Column: col, OldType: oldType, NewType: typ,
				})
			}
		}
		for col := range oldCols {
			if _, exists := liveCols[col]; !exists {
				diff.RemovedColumns = append(diff.RemovedColumns, diffColumn{Table: name, Column: col})
			}
		}
	}
	sortDiffColumns(diff.AddedColumns)
	sortDiffColumns(diff.RemovedColumns)
	sort.SliceStable(diff.ChangedColumns, func(i, j int) bool {
		if diff.ChangedColumns[i].Table != diff.ChangedColumns[j].Table {
			return diff.ChangedColumns[i].Table < diff.ChangedColumns[j].Table
		}
		return diff.ChangedColumns[i].Column < diff.ChangedColumns[j].Column
	})
	return diff
}

func sortDiffColumns(list []diffColumn) {
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].Table != list[j].Table {
			return list[i].Table < list[j].Table
		}
		return list[i].Column < list[j].Column
	})
}

// applyDiffToMappings 根据 diff 更新 mapping_config 中受影响列的校验状态：
// 删除列 -> column_deleted，变更列 -> column_changed，整表删除 -> 其全部映射项 column_deleted。
func applyDiffToMappings(datasourceID uint, diff *schemaDiff) {
	if diff == nil {
		return
	}
	deletedByTable := map[string]map[string]bool{}
	changedByTable := map[string]map[string]bool{}
	for _, rc := range diff.RemovedColumns {
		if deletedByTable[rc.Table] == nil {
			deletedByTable[rc.Table] = map[string]bool{}
		}
		deletedByTable[rc.Table][rc.Column] = true
	}
	for _, cc := range diff.ChangedColumns {
		if changedByTable[cc.Table] == nil {
			changedByTable[cc.Table] = map[string]bool{}
		}
		changedByTable[cc.Table][cc.Column] = true
	}

	// 整表删除：标记该表所有映射项
	for _, table := range diff.RemovedTables {
		markTableMappingsDeleted(datasourceID, table)
	}

	tables := map[string]bool{}
	for t := range deletedByTable {
		tables[t] = true
	}
	for t := range changedByTable {
		tables[t] = true
	}
	for table := range tables {
		updateMappingValidation(datasourceID, table, deletedByTable[table], changedByTable[table])
	}
}

// markTableMappingsDeleted 将某张表下所有映射项标记为 column_deleted。
func markTableMappingsDeleted(datasourceID uint, table string) {
	var configs []model.OntMappingConfig
	DB().Where("datasource_id = ? AND source_table = ?", datasourceID, table).Find(&configs)
	for _, cfg := range configs {
		mappings := parseMappings(cfg.PropertyMappingsJSON)
		changed := false
		for i := range mappings {
			if mappings[i].ValidationStatus != "column_deleted" {
				mappings[i].ValidationStatus = "column_deleted"
				changed = true
			}
		}
		if changed {
			saveMappingsJSON(cfg.ID, mappings)
		}
	}
}

// updateMappingValidation 针对指定表，将删除列标记为 column_deleted、变更列标记为 column_changed。
func updateMappingValidation(datasourceID uint, table string, deleted, changed map[string]bool) {
	if len(deleted) == 0 && len(changed) == 0 {
		return
	}
	var configs []model.OntMappingConfig
	DB().Where("datasource_id = ? AND source_table = ?", datasourceID, table).Find(&configs)
	for _, cfg := range configs {
		mappings := parseMappings(cfg.PropertyMappingsJSON)
		dirty := false
		for i := range mappings {
			col := mappings[i].ColumnName
			if col == "" {
				continue
			}
			if deleted[col] {
				if mappings[i].ValidationStatus != "column_deleted" {
					mappings[i].ValidationStatus = "column_deleted"
					dirty = true
				}
			} else if changed[col] {
				if mappings[i].ValidationStatus != "column_changed" {
					mappings[i].ValidationStatus = "column_changed"
					dirty = true
				}
			}
		}
		if dirty {
			saveMappingsJSON(cfg.ID, mappings)
		}
	}
}

// saveMappingsJSON 回写映射项 JSON 与更新时间。
func saveMappingsJSON(configID uint, mappings []PropertyMapping) {
	if err := DB().Model(&model.OntMappingConfig{}).Where("id = ?", configID).
		Updates(map[string]interface{}{
			"property_mappings_json": marshalJSON(mappings),
			"updated_at":             time.Now(),
		}).Error; err != nil {
		logger.Warnf("[datasource] 回写映射校验状态失败 config_id=%d err=%v", configID, err)
	}
}

// loadCachedSchema 从缓存表组装 SchemaInfo。
func loadCachedSchema(datasourceID uint) *connector.SchemaInfo {
	info := &connector.SchemaInfo{Tables: []connector.TableInfo{}}

	var tables []model.SchemaTable
	DB().Where("datasource_id = ?", datasourceID).Order("table_name ASC").Find(&tables)
	if len(tables) == 0 {
		return info
	}

	var columns []model.SchemaColumn
	DB().Where("datasource_id = ?", datasourceID).Order("table_id ASC, ordinal_pos ASC").Find(&columns)
	var fks []model.SchemaForeignKey
	DB().Where("datasource_id = ?", datasourceID).Find(&fks)
	var idxList []model.SchemaIndex
	DB().Where("datasource_id = ?", datasourceID).Find(&idxList)

	colsByTable := map[uint][]connector.ColumnInfo{}
	for _, col := range columns {
		colsByTable[col.TableID] = append(colsByTable[col.TableID], connector.ColumnInfo{
			Name:       col.ColumnName,
			DataType:   col.DataType,
			IsNullable: col.IsNullable,
			IsPrimary:  col.IsPrimaryKey,
			Default:    col.DefaultValue,
			Comment:    col.Comment,
			Position:   col.OrdinalPos,
		})
	}
	fksByTable := map[uint][]connector.FKInfo{}
	for _, fk := range fks {
		fksByTable[fk.TableID] = append(fksByTable[fk.TableID], connector.FKInfo{
			Name: fk.ConstraintName, Column: fk.ColumnName,
			RefTable: fk.RefTable, RefColumn: fk.RefColumn,
		})
	}
	idxByTable := map[uint][]connector.IndexInfo{}
	for _, idx := range idxList {
		idxByTable[idx.TableID] = append(idxByTable[idx.TableID], connector.IndexInfo{
			Name: idx.IndexName, Columns: idx.Columns, IsUnique: idx.IsUnique, IsPrimary: idx.IsPrimary,
		})
	}

	for _, t := range tables {
		info.Tables = append(info.Tables, connector.TableInfo{
			Name:        t.TableName,
			Schema:      t.SchemaName,
			Type:        t.TableType,
			Comment:     t.Comment,
			Columns:     orEmptyColumns(colsByTable[t.ID]),
			ForeignKeys: orEmptyFKs(fksByTable[t.ID]),
			Indexes:     orEmptyIndexes(idxByTable[t.ID]),
		})
	}
	return info
}

// loadTableColumns 读取指定数据源表的字段缓存；缓存缺失时返回空切片。
func loadTableColumns(datasourceID uint, tableName string) ([]model.SchemaColumn, bool) {
	var table model.SchemaTable
	err := DB().Where("datasource_id = ? AND table_name = ?", datasourceID, tableName).
		First(&table).Error
	if err != nil {
		return nil, false
	}
	var columns []model.SchemaColumn
	DB().Where("table_id = ?", table.ID).Order("ordinal_pos ASC").Find(&columns)
	return columns, len(columns) > 0
}

// countTablesByDatasource 统计数据源下的表数量。
func countTablesByDatasource(ids []uint) map[uint]int64 {
	result := map[uint]int64{}
	if len(ids) == 0 {
		return result
	}
	type row struct {
		DataSourceID uint
		N            int64
	}
	var rows []row
	DB().Model(&model.SchemaTable{}).
		Select("datasource_id AS data_source_id, COUNT(*) AS n").
		Where("datasource_id IN ?", ids).
		Group("datasource_id").
		Scan(&rows)
	for _, r := range rows {
		result[r.DataSourceID] = r.N
	}
	return result
}

// loadLastTests 批量读取每个数据源最近一次连通性测试记录。
func loadLastTests(ids []uint) map[uint]*datasourceTestView {
	result := map[uint]*datasourceTestView{}
	if len(ids) == 0 {
		return result
	}
	var logs []model.DataSourceTest
	subQuery := DB().Model(&model.DataSourceTest{}).
		Select("MAX(id)").
		Where("datasource_id IN ?", ids).
		Group("datasource_id")
	if err := DB().Where("id IN (?)", subQuery).Order("id ASC").Find(&logs).Error; err != nil {
		logger.Warnf("[datasource] 读取测试记录失败: %v", err)
		return result
	}
	for _, l := range logs {
		result[l.DataSourceID] = &datasourceTestView{
			ID:        l.ID,
			Status:    l.Status,
			LatencyMs: l.LatencyMs,
			Message:   l.Message,
			TestedAt:  l.TestedAt,
		}
	}
	return result
}

// buildDatasourceView 组装数据源响应视图。
func buildDatasourceView(ds model.DataSource, tableCount int64, lastTest *datasourceTestView, supported []string) datasourceView {
	return datasourceView{
		DataSource: ds,
		TableCount: tableCount,
		TagsList:   splitTags(ds.Tags),
		Supported:  containsString(supported, ds.Type),
		LastTest:   lastTest,
	}
}

// applyDatasourceDefaults 为连接池等字段填充默认值。
func applyDatasourceDefaults(ds *model.DataSource) {
	if ds.PoolMaxOpen <= 0 {
		ds.PoolMaxOpen = 10
	}
	if ds.PoolMaxIdle <= 0 {
		ds.PoolMaxIdle = 5
	}
	if ds.PoolMaxLifetime <= 0 {
		ds.PoolMaxLifetime = 3600
	}
	if strings.TrimSpace(ds.SSLMode) == "" {
		ds.SSLMode = "disable"
	}
	if strings.TrimSpace(ds.Status) == "" {
		ds.Status = "inactive"
	}
	if ds.Port <= 0 {
		ds.Port = defaultPort(ds.Type)
	}
}

// defaultPort 返回各数据源类型的默认端口。
func defaultPort(dsType string) int {
	switch strings.ToLower(strings.TrimSpace(dsType)) {
	case "postgres", "postgresql":
		return 5432
	case "mysql", "mariadb":
		return 3306
	case "dameng", "dm":
		return 5236
	case "kingbase":
		return 54321
	case "oracle":
		return 1521
	case "sqlserver", "mssql":
		return 1433
	default:
		return 0
	}
}

// splitTags 将逗号分隔的标签串拆分为切片。
func splitTags(tags string) []string {
	tags = strings.TrimSpace(tags)
	if tags == "" {
		return []string{}
	}
	// 兼容历史 JSON 数组格式：["a","b"]
	if strings.HasPrefix(tags, "[") {
		var arr []string
		if err := json.Unmarshal([]byte(tags), &arr); err == nil {
			return arr
		}
	}
	parts := strings.Split(tags, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			result = append(result, s)
		}
	}
	return result
}

// rawToInterface 将 json.RawMessage 转为通用 interface{}。
func rawToInterface(raw json.RawMessage) interface{} {
	if len(raw) == 0 {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	return v
}

// boolToStatus 将布尔结果映射为测试状态字符串。
func boolToStatus(success bool) string {
	if success {
		return "success"
	}
	return "failed"
}

// orEmptyColumns 保证 JSON 输出为空数组而非 null。
func orEmptyColumns(v []connector.ColumnInfo) []connector.ColumnInfo {
	if v == nil {
		return []connector.ColumnInfo{}
	}
	return v
}

func orEmptyFKs(v []connector.FKInfo) []connector.FKInfo {
	if v == nil {
		return []connector.FKInfo{}
	}
	return v
}

func orEmptyIndexes(v []connector.IndexInfo) []connector.IndexInfo {
	if v == nil {
		return []connector.IndexInfo{}
	}
	return v
}

// keysOf 返回 map 的键集合（用于审计详情）。
func keysOf(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// truncate 按字符（而非字节）截断过长文本，避免破坏 UTF-8 编码。
func truncate(s string, max int) string {
	if max <= 0 {
		return s
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}
