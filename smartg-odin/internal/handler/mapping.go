package handler

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"smartg-odin/internal/common/errs"
	"smartg-odin/internal/common/logger"
	"smartg-odin/internal/middleware"
	"smartg-odin/internal/model"
)

// ---------------------------------------------------------------------------
// 属性映射结构
// ---------------------------------------------------------------------------

// PropertyMapping 本体属性到物理列的单条映射规则。
// 该结构同时定义了 mapping_config.property_mappings_json 的存储约定，
// 查询引擎（internal/ds/query）按相同字段解析。
type PropertyMapping struct {
	PropertyID          uint              `json:"property_id"`
	PropertyName        string            `json:"property_name"`
	ColumnName          string            `json:"column_name"`
	ColumnDataType      string            `json:"column_data_type,omitempty"`
	DataType            string            `json:"data_type,omitempty"`
	TransformExpression string            `json:"transform_expression,omitempty"`
	ValueMap            map[string]string `json:"value_map,omitempty"`
	IsKey               bool              `json:"is_key,omitempty"`
	Required            bool              `json:"required,omitempty"`
	Expression          string            `json:"expression,omitempty"`
	Confidence          float64           `json:"confidence,omitempty"`
	ValidationStatus    string            `json:"validation_status,omitempty"`
}

// UnmarshalJSON 兼容 snake_case 与 camelCase 两种键名写法。
func (p *PropertyMapping) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	pick := func(keys ...string) (json.RawMessage, bool) {
		for _, k := range keys {
			if v, ok := raw[k]; ok && len(v) > 0 && string(v) != "null" {
				return v, true
			}
		}
		return nil, false
	}
	if v, ok := pick("property_id", "propertyId", "id"); ok {
		p.PropertyID = unmarshalUint(v)
	}
	if v, ok := pick("property_name", "propertyName", "property", "name"); ok {
		p.PropertyName = unmarshalString(v)
	}
	if v, ok := pick("column_name", "columnName", "column", "source_column"); ok {
		p.ColumnName = unmarshalString(v)
	}
	if v, ok := pick("column_data_type", "columnDataType", "column_type"); ok {
		p.ColumnDataType = unmarshalString(v)
	}
	if v, ok := pick("data_type", "dataType", "property_type"); ok {
		p.DataType = unmarshalString(v)
	}
	if v, ok := pick("transform_expression", "transformExpression", "transform"); ok {
		p.TransformExpression = unmarshalString(v)
	}
	if v, ok := pick("expression", "expr"); ok {
		p.Expression = unmarshalString(v)
	}
	if v, ok := pick("value_map", "valueMap"); ok {
		p.ValueMap = unmarshalStringMap(v)
	}
	if v, ok := pick("is_key", "isKey", "key", "primary"); ok {
		p.IsKey = unmarshalBool(v)
	}
	if v, ok := pick("required", "not_null"); ok {
		p.Required = unmarshalBool(v)
	}
	if v, ok := pick("confidence", "score"); ok {
		var f float64
		if err := json.Unmarshal(v, &f); err == nil {
			p.Confidence = f
		}
	}
	if v, ok := pick("validation_status", "validationStatus", "status"); ok {
		p.ValidationStatus = unmarshalString(v)
	}
	return nil
}

// mappingPayload 映射配置的创建请求体。
type mappingPayload struct {
	OntologyID           uint              `json:"ontology_id"`
	ClassID              uint              `json:"class_id"`
	ClassName            string            `json:"class_name"`
	DataSourceID         uint              `json:"datasource_id"`
	DatasourceID         uint              `json:"datasourceId"`
	SourceTable          string            `json:"source_table"`
	TableName            string            `json:"table_name"`
	SourceSchema         string            `json:"source_schema"`
	SchemaName           string            `json:"schema_name"`
	Status               string            `json:"status"`
	PropertyMappings     []PropertyMapping `json:"property_mappings"`
	PropertyMappingsJSON string            `json:"property_mappings_json"`
}

// mappingView 映射配置响应视图。
type mappingView struct {
	model.OntMappingConfig
	ClassName        string            `json:"class_name"`
	OntologyName     string            `json:"ontology_name"`
	DataSourceName   string            `json:"datasource_name"`
	DataSourceType   string            `json:"datasource_type"`
	PropertyMappings []PropertyMapping `json:"property_mappings"`
	MappedCount      int               `json:"mapped_count"`
}

// mappingUpdatableColumns 允许通过 PUT 更新的列。
var mappingUpdatableColumns = []string{
	"ontology_id", "class_id", "datasource_id", "source_table", "source_schema", "status",
}

// ---------------------------------------------------------------------------
// 映射 CRUD
// ---------------------------------------------------------------------------

// ListMappings GET /api/v1/mappings?ontology_id=&class_id=&datasource_id=&current=1&size=20
func ListMappings(c *gin.Context) {
	current, size := parsePage(c)

	filter := func() *gorm.DB {
		tx := DB().Model(&model.OntMappingConfig{})
		if v := uintQuery(c, "ontology_id"); v > 0 {
			tx = tx.Where("ontology_id = ?", v)
		}
		if v := uintQuery(c, "class_id"); v > 0 {
			tx = tx.Where("class_id = ?", v)
		}
		if v := uintQuery(c, "datasource_id"); v > 0 {
			tx = tx.Where("datasource_id = ?", v)
		}
		if s := strings.TrimSpace(c.Query("status")); s != "" {
			tx = tx.Where("status = ?", s)
		}
		if t := strings.TrimSpace(c.Query("source_table")); t != "" {
			tx = tx.Where("source_table = ?", t)
		}
		return tx
	}

	var total int64
	if err := filter().Count(&total).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	var rows []model.OntMappingConfig
	if err := filter().Order("id ASC").Offset((current - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	views := make([]mappingView, 0, len(rows))
	for _, r := range rows {
		views = append(views, buildMappingView(r))
	}
	middleware.SuccessPage(c, views, total, current, size)
}

// GetMapping GET /api/v1/mappings/:id
func GetMapping(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var mapping model.OntMappingConfig
	if err := DB().First(&mapping, id).Error; err != nil {
		fail(c, mapError(err, fmt.Sprintf("映射配置不存在: %d", id)))
		return
	}

	view := buildMappingView(mapping)

	// 附带目标表已缓存的列信息，便于前端直接展示可选列
	var columns []model.SchemaColumn
	var table model.SchemaTable
	if err := DB().Where("datasource_id = ? AND table_name = ?", mapping.DataSourceID, mapping.SourceTable).
		First(&table).Error; err == nil {
		DB().Where("table_id = ?", table.ID).Order("ordinal_pos ASC").Find(&columns)
	}
	// 附带类的全部属性，便于识别未映射项
	var properties []model.OntClassProperty
	DB().Where("class_id = ?", mapping.ClassID).Order("sort_order ASC, id ASC").Find(&properties)

	ok(c, gin.H{
		"mapping":    view,
		"columns":    columns,
		"properties": properties,
		"unmapped":   unmappedProperties(properties, view.PropertyMappings),
	})
}

// CreateMapping POST /api/v1/mappings
func CreateMapping(c *gin.Context) {
	var req mappingPayload
	if appErr := bindNormalized(c, &req); appErr != nil {
		fail(c, appErr)
		return
	}

	if req.DataSourceID == 0 && req.DatasourceID > 0 {
		req.DataSourceID = req.DatasourceID
	}
	req.SourceTable = firstNonEmpty(strings.TrimSpace(req.SourceTable), strings.TrimSpace(req.TableName))
	req.SourceSchema = firstNonEmpty(strings.TrimSpace(req.SourceSchema), strings.TrimSpace(req.SchemaName))

	// 类 ID 与类名二者其一必填
	if req.ClassID == 0 && strings.TrimSpace(req.ClassName) != "" && req.OntologyID > 0 {
		var cls model.OntClass
		if err := DB().Where("ontology_id = ? AND name = ?", req.OntologyID, strings.TrimSpace(req.ClassName)).
			First(&cls).Error; err == nil {
			req.ClassID = cls.ID
		}
	}
	// 本体 ID 可由类反查
	if req.OntologyID == 0 && req.ClassID > 0 {
		var cls model.OntClass
		if err := DB().First(&cls, req.ClassID).Error; err == nil {
			req.OntologyID = cls.OntologyID
		}
	}

	if req.OntologyID == 0 {
		fail(c, errs.ParamError("ontology_id 不能为空"))
		return
	}
	if req.ClassID == 0 {
		fail(c, errs.ParamError("class_id（或可解析的 class_name）不能为空"))
		return
	}
	if req.DataSourceID == 0 {
		fail(c, errs.ParamError("datasource_id 不能为空"))
		return
	}
	if req.SourceTable == "" {
		fail(c, errs.ParamError("source_table 不能为空"))
		return
	}
	if appErr := ensureOntology(req.OntologyID); appErr != nil {
		fail(c, appErr)
		return
	}
	if appErr := ensureClass(req.OntologyID, req.ClassID); appErr != nil {
		fail(c, appErr)
		return
	}
	if _, appErr := loadDatasource(req.DataSourceID); appErr != nil {
		fail(c, appErr)
		return
	}
	// 同一本体 + 类 + 数据源 + 表 组合唯一
	var dup int64
	DB().Model(&model.OntMappingConfig{}).
		Where("ontology_id = ? AND class_id = ? AND datasource_id = ? AND source_table = ?",
			req.OntologyID, req.ClassID, req.DataSourceID, req.SourceTable).
		Count(&dup)
	if dup > 0 {
		fail(c, errs.ParamError("该类在此数据源表上的映射配置已存在"))
		return
	}

	mappings := req.PropertyMappings
	if len(mappings) == 0 && strings.TrimSpace(req.PropertyMappingsJSON) != "" {
		if err := json.Unmarshal([]byte(req.PropertyMappingsJSON), &mappings); err != nil {
			fail(c, errs.ParamError("property_mappings_json 解析失败: "+err.Error()))
			return
		}
	}
	mappings = normalizeMappings(req.ClassID, mappings)

	mapping := model.OntMappingConfig{
		OntologyID:           req.OntologyID,
		ClassID:              req.ClassID,
		DataSourceID:         req.DataSourceID,
		SourceTable:          req.SourceTable,
		SourceSchema:         req.SourceSchema,
		PropertyMappingsJSON: marshalJSON(mappings),
		Status:               firstNonEmpty(strings.TrimSpace(req.Status), "draft"),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	if err := DB().Create(&mapping).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	writeAudit(c, "create", "mapping", mapping.ID, gin.H{
		"ontology_id": req.OntologyID, "class_id": req.ClassID,
		"datasource_id": req.DataSourceID, "source_table": req.SourceTable,
		"mappings": len(mappings),
	})
	logger.Infof("[mapping] created id=%d class=%d table=%s props=%d",
		mapping.ID, mapping.ClassID, mapping.SourceTable, len(mappings))
	okMsg(c, buildMappingView(mapping), "创建成功")
}

// UpdateMapping PUT /api/v1/mappings/:id
func UpdateMapping(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var mapping model.OntMappingConfig
	if err := DB().First(&mapping, id).Error; err != nil {
		fail(c, mapError(err, fmt.Sprintf("映射配置不存在: %d", id)))
		return
	}

	body, appErr := bindMap(c)
	if appErr != nil {
		fail(c, appErr)
		return
	}
	updates := pickUpdates(body, mappingUpdatableColumns...)
	if table, hasTable := body["table_name"]; hasTable {
		if s, isStr := table.(string); isStr && strings.TrimSpace(s) != "" {
			updates["source_table"] = strings.TrimSpace(s)
		}
	}
	if schema, hasSchema := body["schema_name"]; hasSchema {
		if s, isStr := schema.(string); isStr && strings.TrimSpace(s) != "" {
			updates["source_schema"] = strings.TrimSpace(s)
		}
	}

	// property_mappings 与 property_mappings_json 二者任一提交即整体替换
	if raw, has := body["property_mappings"]; has {
		buf, err := json.Marshal(raw)
		if err != nil {
			fail(c, errs.ParamError("property_mappings 序列化失败: "+err.Error()))
			return
		}
		var mappings []PropertyMapping
		if err := json.Unmarshal(buf, &mappings); err != nil {
			fail(c, errs.ParamError("property_mappings 解析失败: "+err.Error()))
			return
		}
		classID := mapping.ClassID
		if v, hasClass := updates["class_id"]; hasClass {
			if n, isNum := toUint(v); isNum && n > 0 {
				classID = n
			}
		}
		updates["property_mappings_json"] = marshalJSON(normalizeMappings(classID, mappings))
	} else if raw, has := body["property_mappings_json"]; has {
		s := rawJSON(raw)
		if s == "" {
			s = "[]"
		}
		var probe []PropertyMapping
		if err := json.Unmarshal([]byte(s), &probe); err != nil {
			fail(c, errs.ParamError("property_mappings_json 必须为映射数组: "+err.Error()))
			return
		}
		updates["property_mappings_json"] = s
	}

	if len(updates) == 0 {
		fail(c, errs.ParamError("没有可识别的更新字段"))
		return
	}

	// 目标变更后重新校验引用完整性
	targetClass := mapping.ClassID
	if v, has := updates["class_id"]; has {
		if n, isNum := toUint(v); isNum {
			targetClass = n
		}
	}
	targetOntology := mapping.OntologyID
	if v, has := updates["ontology_id"]; has {
		if n, isNum := toUint(v); isNum {
			targetOntology = n
		}
	}
	if targetOntology > 0 && targetClass > 0 {
		if appErr := ensureClass(targetOntology, targetClass); appErr != nil {
			fail(c, appErr)
			return
		}
	}
	if v, has := updates["datasource_id"]; has {
		if n, isNum := toUint(v); isNum && n > 0 {
			if _, appErr := loadDatasource(n); appErr != nil {
				fail(c, appErr)
				return
			}
		}
	}
	updates["updated_at"] = time.Now()

	if err := DB().Model(&model.OntMappingConfig{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}
	DB().First(&mapping, id)

	writeAudit(c, "update", "mapping", id, gin.H{"fields": keysOf(updates)})
	okMsg(c, buildMappingView(mapping), "更新成功")
}

// DeleteMapping DELETE /api/v1/mappings/:id
func DeleteMapping(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var mapping model.OntMappingConfig
	if err := DB().First(&mapping, id).Error; err != nil {
		fail(c, mapError(err, fmt.Sprintf("映射配置不存在: %d", id)))
		return
	}
	if err := DB().Delete(&model.OntMappingConfig{}, id).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	writeAudit(c, "delete", "mapping", id, gin.H{"class_id": mapping.ClassID, "source_table": mapping.SourceTable})
	okMsg(c, gin.H{"id": id}, "删除成功")
}

// ---------------------------------------------------------------------------
// 映射校验
// ---------------------------------------------------------------------------

// validationIssue 单条校验问题。
type validationIssue struct {
	Code       string `json:"code"`
	Property   string `json:"property"`
	PropertyID uint   `json:"property_id,omitempty"`
	DataType   string `json:"data_type,omitempty"`
	Column     string `json:"column,omitempty"`
	ColumnType string `json:"column_type,omitempty"`
	Message    string `json:"message"`
}

// validationResult 映射校验结果。
type validationResult struct {
	Valid         bool              `json:"valid"`
	MappingID     uint              `json:"mapping_id"`
	OntologyID    uint              `json:"ontology_id"`
	ClassID       uint              `json:"class_id"`
	ClassName     string            `json:"class_name"`
	DataSourceID  uint              `json:"datasource_id"`
	SourceTable   string            `json:"source_table"`
	CheckedAt     time.Time         `json:"checked_at"`
	PropertyCount int               `json:"property_count"`
	MappedCount   int               `json:"mapped_count"`
	ColumnCount   int               `json:"column_count"`
	Errors        []validationIssue `json:"errors"`
	Warnings      []validationIssue `json:"warnings"`
	Status        string            `json:"status"`
}

// ValidateMapping POST /api/v1/mappings/:id/validate
// 校验映射列是否存在于目标表、类型是否兼容、必填属性是否遗漏。
func ValidateMapping(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var mapping model.OntMappingConfig
	if err := DB().First(&mapping, id).Error; err != nil {
		fail(c, mapError(err, fmt.Sprintf("映射配置不存在: %d", id)))
		return
	}

	result, appErr := validateMapping(&mapping, true)
	if appErr != nil {
		fail(c, appErr)
		return
	}

	// 校验通过则升级为 validated，失败则从 validated 回退到 draft
	newStatus := mapping.Status
	if result.Valid {
		newStatus = "validated"
	} else if mapping.Status == "validated" {
		newStatus = "draft"
	}
	if newStatus != mapping.Status {
		DB().Model(&model.OntMappingConfig{}).Where("id = ?", id).
			Updates(map[string]interface{}{"status": newStatus, "updated_at": time.Now()})
		result.Status = newStatus
	}

	writeAudit(c, "validate", "mapping", id, gin.H{"valid": result.Valid,
		"errors": len(result.Errors), "warnings": len(result.Warnings)})

	message := "校验通过"
	if !result.Valid {
		message = fmt.Sprintf("校验未通过：%d 个错误 / %d 个警告", len(result.Errors), len(result.Warnings))
	}
	okMsg(c, result, message)
}

// validateMapping 执行映射校验；allowIntrospect 为 true 时缓存缺失会触发实时内省。
func validateMapping(mapping *model.OntMappingConfig, allowIntrospect bool) (*validationResult, *errs.AppError) {
	result := &validationResult{
		Valid:        true,
		MappingID:    mapping.ID,
		OntologyID:   mapping.OntologyID,
		ClassID:      mapping.ClassID,
		DataSourceID: mapping.DataSourceID,
		SourceTable:  mapping.SourceTable,
		CheckedAt:    time.Now(),
		Errors:       []validationIssue{},
		Warnings:     []validationIssue{},
		Status:       mapping.Status,
	}
	addError := func(issue validationIssue) {
		result.Valid = false
		result.Errors = append(result.Errors, issue)
	}
	addWarning := func(issue validationIssue) {
		result.Warnings = append(result.Warnings, issue)
	}

	// 1. 关联对象存在性
	var cls model.OntClass
	if err := DB().First(&cls, mapping.ClassID).Error; err != nil {
		addError(validationIssue{Code: "class_not_found",
			Message: fmt.Sprintf("本体类不存在: %d", mapping.ClassID)})
	} else {
		result.ClassName = cls.Name
	}

	ds, appErr := loadDatasource(mapping.DataSourceID)
	if appErr != nil {
		addError(validationIssue{Code: "datasource_not_found",
			Message: fmt.Sprintf("数据源不存在: %d", mapping.DataSourceID)})
		return result, nil
	}

	// 2. 目标表列信息（优先缓存，必要时实时内省）
	columns, cached := loadTableColumns(mapping.DataSourceID, mapping.SourceTable)
	if !cached && allowIntrospect {
		if _, appErr := syncSchema(ds); appErr != nil {
			addWarning(validationIssue{Code: "introspect_failed",
				Message: "无法实时内省数据源结构：" + appErr.Message})
		} else {
			columns, cached = loadTableColumns(mapping.DataSourceID, mapping.SourceTable)
		}
	}
	if len(columns) == 0 {
		addError(validationIssue{Code: "table_not_found", Column: mapping.SourceTable,
			Message: fmt.Sprintf("数据源 %s 中未找到表 %s 的结构信息，请先执行内省", ds.Name, mapping.SourceTable)})
		return result, nil
	}
	columnIndex := make(map[string]model.SchemaColumn, len(columns))
	for _, col := range columns {
		columnIndex[strings.ToLower(col.ColumnName)] = col
	}
	result.ColumnCount = len(columns)

	// 3. 本体属性清单
	var properties []model.OntClassProperty
	DB().Where("class_id = ?", mapping.ClassID).Order("sort_order ASC, id ASC").Find(&properties)
	result.PropertyCount = len(properties)
	propByID := make(map[uint]model.OntClassProperty, len(properties))
	propByName := make(map[string]model.OntClassProperty, len(properties))
	for _, p := range properties {
		propByID[p.ID] = p
		propByName[strings.ToLower(p.Name)] = p
	}

	// 4. 逐条校验属性映射
	mappings := parseMappings(mapping.PropertyMappingsJSON)
	result.MappedCount = len(mappings)
	usedColumns := map[string]int{}

	for _, m := range mappings {
		prop, hasProp := resolveProperty(m, propByID, propByName)
		propName := m.PropertyName
		if propName == "" && hasProp {
			propName = prop.Name
		}
		if propName == "" {
			propName = fmt.Sprintf("#%d", m.PropertyID)
		}

		if !hasProp && len(properties) > 0 {
			addWarning(validationIssue{Code: "property_not_defined", Property: propName,
				PropertyID: m.PropertyID,
				Message:    "映射的属性未在本体类中定义，查询翻译时可能被忽略"})
		}

		dataType := m.DataType
		if hasProp && strings.TrimSpace(dataType) == "" {
			dataType = prop.DataType
		}

		if strings.TrimSpace(m.ColumnName) == "" {
			if strings.TrimSpace(m.Expression) != "" || strings.TrimSpace(m.TransformExpression) != "" {
				continue // 计算列，不依赖物理列
			}
			addError(validationIssue{Code: "column_required", Property: propName,
				PropertyID: m.PropertyID, DataType: dataType,
				Message: "未指定映射列 column_name"})
			continue
		}

		col, exists := columnIndex[strings.ToLower(strings.TrimSpace(m.ColumnName))]
		if !exists {
			addError(validationIssue{Code: "column_not_found", Property: propName,
				PropertyID: m.PropertyID, DataType: dataType, Column: m.ColumnName,
				Message: fmt.Sprintf("表 %s 中不存在列 %s", mapping.SourceTable, m.ColumnName)})
			continue
		}
		usedColumns[col.ColumnName]++

		if level, msg := checkTypeCompatibility(dataType, col.DataType); level != "" {
			issue := validationIssue{
				Code: "type_mismatch", Property: propName, PropertyID: m.PropertyID,
				DataType: dataType, Column: col.ColumnName, ColumnType: col.DataType, Message: msg,
			}
			if level == "error" {
				addError(issue)
			} else {
				addWarning(issue)
			}
		}

		if col.IsNullable && (m.Required || (hasProp && prop.Required)) && !m.IsKey {
			addWarning(validationIssue{Code: "nullable_required", Property: propName,
				PropertyID: m.PropertyID, Column: col.ColumnName, ColumnType: col.DataType,
				Message: fmt.Sprintf("属性 %s 要求非空，但列 %s 允许 NULL", propName, col.ColumnName)})
		}
		if m.IsKey && !col.IsPrimaryKey && len(columnIndex) > 0 {
			addWarning(validationIssue{Code: "key_not_primary", Property: propName,
				PropertyID: m.PropertyID, Column: col.ColumnName, ColumnType: col.DataType,
				Message: fmt.Sprintf("主键属性 %s 映射到非主键列 %s，跨源合并时可能产生重复", propName, col.ColumnName)})
		}
	}

	// 5. 重复映射同一列
	for columnName, n := range usedColumns {
		if n > 1 {
			addWarning(validationIssue{Code: "duplicate_column", Column: columnName,
				Message: fmt.Sprintf("列 %s 被 %d 个属性重复映射", columnName, n)})
		}
	}

	// 6. 必填属性遗漏
	mappedProps := map[uint]bool{}
	mappedNames := map[string]bool{}
	for _, m := range mappings {
		if m.PropertyID > 0 {
			mappedProps[m.PropertyID] = true
		}
		if strings.TrimSpace(m.PropertyName) != "" {
			mappedNames[strings.ToLower(strings.TrimSpace(m.PropertyName))] = true
		}
	}
	for _, p := range properties {
		if mappedProps[p.ID] || mappedNames[strings.ToLower(p.Name)] {
			continue
		}
		issue := validationIssue{Code: "property_unmapped", Property: p.Name,
			PropertyID: p.ID, DataType: p.DataType,
			Message: fmt.Sprintf("属性 %s 尚未映射到任何列", p.Name)}
		if p.Required {
			issue.Code = "required_property_unmapped"
			issue.Message = fmt.Sprintf("必填属性 %s 尚未映射到任何列", p.Name)
			addError(issue)
		} else {
			addWarning(issue)
		}
	}

	if len(mappings) == 0 {
		addError(validationIssue{Code: "empty_mapping",
			Message: "映射配置为空，至少需要配置一条属性映射"})
	}

	return result, nil
}

// resolveProperty 依据 property_id 或 property_name 定位本体属性。
func resolveProperty(m PropertyMapping, byID map[uint]model.OntClassProperty,
	byName map[string]model.OntClassProperty) (model.OntClassProperty, bool) {
	if m.PropertyID > 0 {
		if p, ok := byID[m.PropertyID]; ok {
			return p, true
		}
	}
	if name := strings.ToLower(strings.TrimSpace(m.PropertyName)); name != "" {
		if p, ok := byName[name]; ok {
			return p, true
		}
	}
	return model.OntClassProperty{}, false
}

// ---------------------------------------------------------------------------
// 智能映射建议
// ---------------------------------------------------------------------------

// suggestRequest 智能建议请求体。
type suggestRequest struct {
	OntologyID       uint    `json:"ontology_id"`
	ClassID          uint    `json:"class_id"`
	ClassName        string  `json:"class_name"`
	DataSourceID     uint    `json:"datasource_id"`
	DatasourceID     uint    `json:"datasourceId"`
	SourceTable      string  `json:"source_table"`
	TableName        string  `json:"table_name"`
	MinConfidence    float64 `json:"min_confidence"`
	TopN             int     `json:"top_n"`
	AutoIntrospect   *bool   `json:"auto_introspect"`
	IncludeUnmatched bool    `json:"include_unmatched"`
}

// mappingSuggestion 单条映射建议。
type mappingSuggestion struct {
	PropertyID     uint     `json:"property_id"`
	PropertyName   string   `json:"property_name"`
	PropertyLabel  string   `json:"property_label"`
	DataType       string   `json:"data_type"`
	Required       bool     `json:"required"`
	ColumnName     string   `json:"column_name"`
	ColumnDataType string   `json:"column_data_type"`
	ColumnComment  string   `json:"column_comment,omitempty"`
	Confidence     float64  `json:"confidence"`
	NameScore      float64  `json:"name_score"`
	TypeScore      float64  `json:"type_score"`
	CommentScore   float64  `json:"comment_score"`
	DataTypeMatch  bool     `json:"data_type_match"`
	Reason         string   `json:"reason"`
	Alternatives   []string `json:"alternatives,omitempty"`
}

// suggestResult 建议结果。
type suggestResult struct {
	OntologyID   uint                `json:"ontology_id"`
	ClassID      uint                `json:"class_id"`
	ClassName    string              `json:"class_name"`
	DataSourceID uint                `json:"datasource_id"`
	SourceTable  string              `json:"source_table"`
	ColumnCount  int                 `json:"column_count"`
	Introspected bool                `json:"introspected"`
	Suggestions  []mappingSuggestion `json:"suggestions"`
	UsedColumns  []string            `json:"used_columns"`
	Unmatched    []string            `json:"unmatched"`
}

// SuggestMappings POST /api/v1/mappings/suggest
// 基于名称相似度（camelCase / snake_case 拆分 + Levenshtein）、注释匹配与类型兼容性打分，
// 为本体类的每个属性推荐最合适的物理列。
func SuggestMappings(c *gin.Context) {
	var req suggestRequest
	if appErr := bindNormalized(c, &req); appErr != nil {
		fail(c, appErr)
		return
	}
	if req.DataSourceID == 0 && req.DatasourceID > 0 {
		req.DataSourceID = req.DatasourceID
	}
	req.SourceTable = firstNonEmpty(strings.TrimSpace(req.SourceTable), strings.TrimSpace(req.TableName))

	if req.DataSourceID == 0 {
		fail(c, errs.ParamError("datasource_id 不能为空"))
		return
	}
	if req.SourceTable == "" {
		fail(c, errs.ParamError("source_table（或 table_name）不能为空"))
		return
	}
	if req.ClassID == 0 && strings.TrimSpace(req.ClassName) == "" {
		fail(c, errs.ParamError("class_id 与 class_name 至少提供其一"))
		return
	}

	// 解析类
	var cls model.OntClass
	if req.ClassID > 0 {
		if err := DB().First(&cls, req.ClassID).Error; err != nil {
			fail(c, mapError(err, fmt.Sprintf("本体类不存在: %d", req.ClassID)))
			return
		}
	} else {
		name := strings.TrimSpace(req.ClassName)
		tx := DB().Where("name = ?", name)
		if req.OntologyID > 0 {
			tx = tx.Where("ontology_id = ?", req.OntologyID)
		}
		if err := tx.First(&cls).Error; err != nil {
			fail(c, mapError(err, "本体类不存在: "+name))
			return
		}
	}
	req.ClassID = cls.ID
	req.OntologyID = firstUint(req.OntologyID, cls.OntologyID)

	// 属性清单
	var properties []model.OntClassProperty
	DB().Where("class_id = ?", cls.ID).Order("sort_order ASC, id ASC").Find(&properties)
	if len(properties) == 0 {
		okMsg(c, []mappingSuggestion{}, "该类尚未定义属性，无建议可生成")
		return
	}

	// 列信息：优先缓存，必要时实时内省
	columns, cached := loadTableColumns(req.DataSourceID, req.SourceTable)
	autoIntrospect := req.AutoIntrospect == nil || *req.AutoIntrospect
	if !cached && autoIntrospect {
		ds, appErr := loadDatasource(req.DataSourceID)
		if appErr != nil {
			fail(c, appErr)
			return
		}
		if _, appErr := syncSchema(ds); appErr != nil {
			fail(c, appErr)
			return
		}
		columns, _ = loadTableColumns(req.DataSourceID, req.SourceTable)
	}
	if len(columns) == 0 {
		fail(c, errs.NotFound(fmt.Sprintf("数据源 %d 中未找到表 %s 的列信息", req.DataSourceID, req.SourceTable)))
		return
	}

	minConfidence := req.MinConfidence
	if minConfidence <= 0 {
		minConfidence = 0.35
	}
	if minConfidence > 1 {
		minConfidence = 1
	}
	topN := req.TopN
	if topN <= 0 {
		topN = 1
	}

	suggestions := make([]mappingSuggestion, 0, len(properties))
	usedColumns := map[string]bool{}
	unmatched := make([]string, 0)
	tableTokens := tokenize(req.SourceTable)

	for _, p := range properties {
		candidates := rankColumns(p, columns, tableTokens)
		if len(candidates) == 0 {
			unmatched = append(unmatched, p.Name)
			continue
		}
		best := candidates[0]
		if best.Confidence < minConfidence {
			unmatched = append(unmatched, p.Name)
			if !req.IncludeUnmatched {
				continue
			}
		}
		if usedColumns[best.ColumnName] {
			// 已被更高分属性占用，尝试次优候选
			picked := false
			for _, alt := range candidates[1:] {
				if alt.Confidence < minConfidence || usedColumns[alt.ColumnName] {
					continue
				}
				best = alt
				picked = true
				break
			}
			if !picked && !req.IncludeUnmatched {
				unmatched = append(unmatched, p.Name)
				continue
			}
		}
		usedColumns[best.ColumnName] = true

		alternatives := make([]string, 0, topN)
		for _, alt := range candidates[1:minInt(len(candidates), topN+1)] {
			alternatives = append(alternatives, fmt.Sprintf("%s(%.2f)", alt.ColumnName, alt.Confidence))
		}
		best.Alternatives = alternatives
		suggestions = append(suggestions, best)
	}

	writeAudit(c, "suggest", "mapping", 0, gin.H{
		"class_id": cls.ID, "datasource_id": req.DataSourceID,
		"source_table": req.SourceTable, "suggestions": len(suggestions),
		"unmatched": len(unmatched),
	})

	// 前端期望 data 为扁平的建议数组（Array.isArray(res) 判定）
	okMsg(c, suggestions, fmt.Sprintf("为 %d 个属性生成了 %d 条映射建议", len(properties), len(suggestions)))
}

// rankColumns 为单个属性对所有列打分并降序排序。
func rankColumns(p model.OntClassProperty, columns []model.SchemaColumn, tableTokens []string) []mappingSuggestion {
	scored := make([]mappingSuggestion, 0, len(columns))
	propTokens := tokenize(p.Name)
	labelTokens := tokenize(p.Label)

	for _, col := range columns {
		colTokens := tokenizeWithout(col.ColumnName, tableTokens)

		nameScore := nameSimilarity(propTokens, tokenize(col.ColumnName), p.Name, col.ColumnName)
		if len(labelTokens) > 0 {
			// 属性中文标签与列名 token 的匹配（少见但有用）
			if s := nameSimilarity(labelTokens, colTokens, p.Label, col.ColumnName); s > nameScore {
				nameScore = (nameScore + s) / 2
			}
		}

		typeLevel, typeMsg := checkTypeCompatibility(p.DataType, col.DataType)
		typeScore := 1.0
		dataTypeMatch := typeLevel == ""
		switch typeLevel {
		case "warning":
			typeScore = 0.55
		case "error":
			typeScore = 0.05
		}

		commentScore := commentSimilarity(p.Label, p.Name, col.Comment)

		// 主键列略微加权，便于优先推荐为 key 属性
		keyBonus := 0.0
		if col.IsPrimaryKey && strings.Contains(strings.ToLower(p.Name), "id") {
			keyBonus = 0.03
		}

		confidence := round3(0.60*nameScore + 0.22*typeScore + 0.18*commentScore + keyBonus)
		if confidence > 1 {
			confidence = 1
		}

		reason := buildReason(nameScore, typeScore, commentScore, dataTypeMatch, typeMsg, col)
		scored = append(scored, mappingSuggestion{
			PropertyID:     p.ID,
			PropertyName:   p.Name,
			PropertyLabel:  p.Label,
			DataType:       p.DataType,
			Required:       p.Required,
			ColumnName:     col.ColumnName,
			ColumnDataType: col.DataType,
			ColumnComment:  col.Comment,
			Confidence:     confidence,
			NameScore:      round3(nameScore),
			TypeScore:      round3(typeScore),
			CommentScore:   round3(commentScore),
			DataTypeMatch:  dataTypeMatch,
			Reason:         reason,
		})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].Confidence != scored[j].Confidence {
			return scored[i].Confidence > scored[j].Confidence
		}
		return scored[i].ColumnName < scored[j].ColumnName
	})
	return scored
}

// buildReason 生成建议理由说明。
func buildReason(nameScore, typeScore, commentScore float64, typeMatch bool, typeMsg string, col model.SchemaColumn) string {
	parts := make([]string, 0, 4)
	switch {
	case nameScore >= 0.99:
		parts = append(parts, "名称完全匹配")
	case nameScore >= 0.8:
		parts = append(parts, fmt.Sprintf("名称高度相似(%.2f)", nameScore))
	case nameScore >= 0.5:
		parts = append(parts, fmt.Sprintf("名称部分相似(%.2f)", nameScore))
	default:
		parts = append(parts, fmt.Sprintf("名称相似度较低(%.2f)", nameScore))
	}
	if typeMatch {
		parts = append(parts, "类型兼容")
	} else if typeMsg != "" {
		parts = append(parts, typeMsg)
	}
	if commentScore >= 0.8 {
		parts = append(parts, "列注释与属性标签一致")
	} else if commentScore >= 0.4 {
		parts = append(parts, "列注释与属性标签相近")
	}
	if col.IsPrimaryKey {
		parts = append(parts, "主键列")
	}
	return strings.Join(parts, " · ")
}

// ---------------------------------------------------------------------------
// 名称相似度与类型兼容
// ---------------------------------------------------------------------------

// nameSimilarity 综合 token 集合相似度、编辑距离相似度与包含关系，返回 [0,1]。
func nameSimilarity(aTokens, bTokens []string, aRaw, bRaw string) float64 {
	a := strings.ToLower(strings.TrimSpace(aRaw))
	b := strings.ToLower(strings.TrimSpace(bRaw))
	if a == "" || b == "" {
		return 0
	}
	if a == b {
		return 1
	}

	best := jaccard(aTokens, bTokens)

	// 去除分隔符后的编辑距离相似度（覆盖 customer_code vs customerCode 等情形）
	lev := levenshtein(a, b)
	maxLen := maxInt(len([]rune(a)), len([]rune(b)))
	if maxLen > 0 {
		sim := 1 - float64(lev)/float64(maxLen)
		if sim > best {
			best = sim
		}
	}

	// 归一化（去下划线）后完全一致
	if strings.ReplaceAll(a, "_", "") == strings.ReplaceAll(b, "_", "") {
		best = 1
	}

	// 包含关系（如列 customer_code 包含属性 code）给予中等分数
	short, long := a, b
	if len([]rune(short)) > len([]rune(long)) {
		short, long = long, short
	}
	if len([]rune(short)) >= 3 && strings.Contains(long, short) {
		containScore := 0.55 + 0.25*float64(len([]rune(short)))/float64(len([]rune(long)))
		if containScore > best {
			best = containScore
		}
	}

	// 单 token 完全命中（如属性 name 对应列 cust_name）
	if len(aTokens) == 1 && len(bTokens) > 1 {
		for _, t := range bTokens {
			if t == aTokens[0] && best < 0.85 {
				best = 0.85
			}
		}
	}
	if len(bTokens) == 1 && len(aTokens) > 1 {
		for _, t := range aTokens {
			if t == bTokens[0] && best < 0.85 {
				best = 0.85
			}
		}
	}

	if best < 0 {
		best = 0
	}
	if best > 1 {
		best = 1
	}
	return best
}

// commentSimilarity 比较属性标签 / 名称与列注释的相似度。
func commentSimilarity(label, name, comment string) float64 {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return 0
	}
	label = strings.TrimSpace(label)
	name = strings.TrimSpace(name)
	if label != "" && (label == comment || strings.Contains(comment, label) || strings.Contains(label, comment)) {
		return 1
	}
	if label != "" {
		sim := 1 - float64(levenshtein(label, comment))/float64(maxInt(len([]rune(label)), len([]rune(comment))))
		if sim >= 0.6 {
			return sim
		}
	}
	lcComment := strings.ToLower(comment)
	if name != "" && strings.ToLower(name) == lcComment {
		return 0.9
	}
	if name != "" && len([]rune(name)) >= 3 && strings.Contains(lcComment, strings.ToLower(name)) {
		return 0.6
	}
	return 0
}

// jaccard 计算 token 集合的 Jaccard 相似度。
func jaccard(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	setA := make(map[string]bool, len(a))
	for _, t := range a {
		setA[t] = true
	}
	setB := make(map[string]bool, len(b))
	intersection := 0
	for _, t := range b {
		setB[t] = true
		if setA[t] {
			intersection++
		}
	}
	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// levenshtein 计算两个字符串（按字符）的编辑距离。
func levenshtein(a, b string) int {
	ar := []rune(a)
	br := []rune(b)
	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}
	prev := make([]int, len(br)+1)
	curr := make([]int, len(br)+1)
	for j := 0; j <= len(br); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		curr[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			curr[j] = minInt(minInt(curr[j-1]+1, prev[j]+1), prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(br)]
}

// tokenize 将 camelCase / snake_case / 中划线名称拆分为小写 token 列表。
func tokenize(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var b strings.Builder
	runes := []rune(s)
	flush := func(out *[]string) {
		if b.Len() > 0 {
			*out = append(*out, strings.ToLower(b.String()))
			b.Reset()
		}
	}
	tokens := make([]string, 0, 4)
	for i, r := range runes {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			// 大写边界：前一个为小写/数字，或后一个为小写（缩写词结尾）
			prevLower := i > 0 && ((runes[i-1] >= 'a' && runes[i-1] <= 'z') || (runes[i-1] >= '0' && runes[i-1] <= '9'))
			nextLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'
			if prevLower || nextLower {
				flush(&tokens)
			}
			b.WriteRune(r + ('a' - 'A'))
		default:
			// 中文等非 ASCII 字符整体作为一个 token
			if r > 127 {
				flush(&tokens)
				tokens = append(tokens, string(r))
				continue
			}
			flush(&tokens)
		}
	}
	flush(&tokens)

	// 过滤无意义的噪声词
	filtered := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if t == "" || noiseTokens[t] {
			continue
		}
		filtered = append(filtered, t)
	}
	if len(filtered) == 0 {
		return tokens
	}
	return filtered
}

// noiseTokens 参与相似度计算时应忽略的通用词。
var noiseTokens = map[string]bool{
	"the": true, "a": true, "an": true, "of": true, "is": true,
	"col": true, "column": true, "field": true, "val": true, "value": true,
}

// tokenizeWithout 拆分名称并剔除指定 token（如表名前缀）。
func tokenizeWithout(s string, exclude []string) []string {
	tokens := tokenize(s)
	if len(exclude) == 0 {
		return tokens
	}
	skip := make(map[string]bool, len(exclude))
	for _, t := range exclude {
		skip[t] = true
	}
	out := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if skip[t] {
			continue
		}
		out = append(out, t)
	}
	if len(out) == 0 {
		return tokens
	}
	return out
}

// ---------------------------------------------------------------------------
// 类型兼容判定
// ---------------------------------------------------------------------------

// 类型分类常量
const (
	catString   = "string"
	catInteger  = "integer"
	catFloat    = "float"
	catBoolean  = "boolean"
	catDatetime = "datetime"
	catJSON     = "json"
	catBinary   = "binary"
	catUnknown  = "unknown"
)

// sqlTypeCategories 数据库列类型 → 分类。
var sqlTypeCategories = map[string]string{
	"varchar": catString, "char": catString, "nvarchar": catString, "nchar": catString,
	"character": catString, "bpchar": catString, "text": catString, "tinytext": catString,
	"mediumtext": catString, "longtext": catString, "clob": catString, "nclob": catString,
	"uuid": catString, "name": catString, "enum": catString, "set": catString,
	"character varying": catString, "varchar2": catString, "nvarchar2": catString,

	"int": catInteger, "integer": catInteger, "bigint": catInteger, "smallint": catInteger,
	"tinyint": catInteger, "mediumint": catInteger, "int2": catInteger, "int4": catInteger,
	"int8": catInteger, "serial": catInteger, "bigserial": catInteger, "smallserial": catInteger,
	"long": catInteger, "int64": catInteger, "int32": catInteger,

	"float": catFloat, "double": catFloat, "real": catFloat, "decimal": catFloat,
	"numeric": catFloat, "float4": catFloat, "float8": catFloat, "money": catFloat,
	"double precision": catFloat, "number": catFloat,

	"bool": catBoolean, "boolean": catBoolean, "bit": catBoolean,

	"date": catDatetime, "datetime": catDatetime, "datetime2": catDatetime,
	"smalldatetime": catDatetime, "timestamp": catDatetime, "timestamptz": catDatetime,
	"timestamp with time zone": catDatetime, "timestamp without time zone": catDatetime,
	"time": catDatetime, "timetz": catDatetime, "year": catDatetime,

	"json": catJSON, "jsonb": catJSON,

	"blob": catBinary, "bytea": catBinary, "binary": catBinary, "varbinary": catBinary,
	"image": catBinary, "raw": catBinary, "longblob": catBinary, "tinyblob": catBinary,
}

// ontologyTypeCategories 本体属性类型 → 分类。
var ontologyTypeCategories = map[string]string{
	"string": catString, "str": catString, "text": catString, "varchar": catString,
	"char": catString, "enum": catString, "uuid": catString, "code": catString,
	"integer": catInteger, "int": catInteger, "long": catInteger, "bigint": catInteger,
	"short": catInteger, "number": catInteger, "count": catInteger,
	"float": catFloat, "double": catFloat, "decimal": catFloat, "numeric": catFloat,
	"money": catFloat, "amount": catFloat,
	"boolean": catBoolean, "bool": catBoolean, "flag": catBoolean,
	"datetime": catDatetime, "date": catDatetime, "time": catDatetime,
	"timestamp": catDatetime, "timestamptz": catDatetime,
	"json": catJSON, "object": catJSON, "array": catJSON, "map": catJSON,
	"binary": catBinary, "blob": catBinary, "bytes": catBinary,
}

// normalizeTypeName 去除长度、精度与 unsigned 等修饰，返回基础类型名。
func normalizeTypeName(t string) string {
	s := strings.ToLower(strings.TrimSpace(t))
	if s == "" {
		return ""
	}
	if idx := strings.IndexAny(s, "("); idx >= 0 {
		s = s[:idx]
	}
	s = strings.TrimSpace(s)
	for _, suffix := range []string{" unsigned", " zerofill", " signed"} {
		s = strings.TrimSuffix(s, suffix)
	}
	s = strings.TrimSpace(s)
	// 形如 "timestamp with time zone" 保留原样，其余按空格首段处理
	if strings.Contains(s, " ") && !strings.HasPrefix(s, "double") &&
		!strings.HasPrefix(s, "timestamp") && !strings.HasPrefix(s, "character") {
		s = strings.Fields(s)[0]
	}
	return s
}

// sqlTypeCategory 返回数据库列类型的分类。
func sqlTypeCategory(colType string) string {
	base := normalizeTypeName(colType)
	if base == "" {
		return catUnknown
	}
	if cat, ok := sqlTypeCategories[base]; ok {
		return cat
	}
	// 前缀兜底匹配（如 int4、varchar2 的变体）
	for key, cat := range sqlTypeCategories {
		if strings.HasPrefix(base, key) {
			return cat
		}
	}
	return catUnknown
}

// ontologyTypeCategory 返回本体属性类型的分类。
func ontologyTypeCategory(propType string) string {
	base := strings.ToLower(strings.TrimSpace(propType))
	if base == "" {
		return catUnknown
	}
	if cat, ok := ontologyTypeCategories[base]; ok {
		return cat
	}
	if cat, ok := sqlTypeCategories[base]; ok {
		return cat
	}
	return catUnknown
}

// checkTypeCompatibility 判定本体属性类型与列类型是否兼容。
// 返回级别（"" 表示兼容、warning 表示需注意、error 表示不兼容）与说明文案。
func checkTypeCompatibility(propType, colType string) (level, message string) {
	propCat := ontologyTypeCategory(propType)
	colCat := sqlTypeCategory(colType)

	if propCat == catUnknown {
		if strings.TrimSpace(propType) == "" {
			return "", ""
		}
		return "warning", fmt.Sprintf("无法识别的属性类型 %s，跳过类型校验", propType)
	}
	if colCat == catUnknown {
		if strings.TrimSpace(colType) == "" {
			return "warning", "目标列缺少类型信息，无法校验类型兼容性"
		}
		return "warning", fmt.Sprintf("无法识别的列类型 %s，跳过类型校验", colType)
	}
	if propCat == colCat {
		return "", ""
	}

	numeric := map[string]bool{catInteger: true, catFloat: true}
	switch {
	case numeric[propCat] && numeric[colCat]:
		if propCat == catInteger && colCat == catFloat {
			return "warning", fmt.Sprintf("整型属性 %s 映射到小数列 %s，聚合结果可能丢失精度", propType, colType)
		}
		return "", ""

	case propCat == catString:
		switch colCat {
		case catBinary:
			return "warning", fmt.Sprintf("字符串属性映射到二进制列 %s，展示可能为乱码", colType)
		case catJSON, catDatetime, catBoolean, catInteger, catFloat:
			return "warning", fmt.Sprintf("字符串属性映射到 %s 列，查询时依赖隐式类型转换", colType)
		}
		return "", ""

	case colCat == catString:
		switch propCat {
		case catJSON:
			return "", ""
		case catDatetime:
			return "warning", fmt.Sprintf("日期时间属性映射到字符串列 %s，需保证存储格式统一", colType)
		case catBoolean:
			return "warning", fmt.Sprintf("布尔属性映射到字符串列 %s，需通过 value_map 归一化取值", colType)
		case catInteger, catFloat:
			return "warning", fmt.Sprintf("数值属性映射到字符串列 %s，范围比较与聚合可能失效", colType)
		}
		return "", ""

	case propCat == catBoolean && colCat == catInteger:
		return "", ""
	case propCat == catInteger && colCat == catBoolean:
		return "warning", fmt.Sprintf("整型属性映射到布尔列 %s，取值仅为 0/1", colType)
	case propCat == catDatetime && colCat == catInteger:
		return "warning", fmt.Sprintf("日期时间属性映射到整数列 %s，按时间戳解析", colType)
	case propCat == catJSON && (colCat == catJSON || colCat == catString):
		return "", ""
	case colCat == catBinary:
		return "warning", fmt.Sprintf("属性 %s 映射到二进制列 %s，无法直接展示", propType, colType)
	case propCat == catDatetime || colCat == catDatetime:
		return "error", fmt.Sprintf("属性类型 %s 与列类型 %s 不兼容（时间类型与 %s 无法互转）",
			propType, colType, otherCategoryName(propCat, colCat))
	case propCat == catBoolean || colCat == catBoolean:
		return "warning", fmt.Sprintf("布尔属性与列类型 %s 需通过 value_map 显式转换", colType)
	}

	return "error", fmt.Sprintf("属性类型 %s 与列类型 %s 不兼容", propType, colType)
}

// otherCategoryName 返回两个分类中非 datetime 的那个，用于错误文案。
func otherCategoryName(a, b string) string {
	if a != catDatetime {
		return a
	}
	return b
}

// ---------------------------------------------------------------------------
// 内部辅助
// ---------------------------------------------------------------------------

// buildMappingView 组装映射配置响应视图。
func buildMappingView(mapping model.OntMappingConfig) mappingView {
	view := mappingView{
		OntMappingConfig: mapping,
		PropertyMappings: parseMappings(mapping.PropertyMappingsJSON),
	}
	if mapping.PropertyMappingsJSON == "" {
		view.PropertyMappingsJSON = "[]"
	}
	view.MappedCount = len(view.PropertyMappings)

	var cls model.OntClass
	if err := DB().Select("id, name").First(&cls, mapping.ClassID).Error; err == nil {
		view.ClassName = cls.Name
	}
	var ont model.OntDefinition
	if err := DB().Select("id, name").First(&ont, mapping.OntologyID).Error; err == nil {
		view.OntologyName = ont.Name
	}
	var ds model.DataSource
	if err := DB().Select("id, name, type").First(&ds, mapping.DataSourceID).Error; err == nil {
		view.DataSourceName = ds.Name
		view.DataSourceType = ds.Type
	}
	return view
}

// parseMappings 解析 property_mappings_json；容错处理空值与非法 JSON。
func parseMappings(raw string) []PropertyMapping {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []PropertyMapping{}
	}
	var mappings []PropertyMapping
	if err := json.Unmarshal([]byte(raw), &mappings); err != nil {
		// 兼容对象形式：{"properties": [...]}
		var wrapper struct {
			Properties []PropertyMapping `json:"properties"`
			Mappings   []PropertyMapping `json:"mappings"`
		}
		if err2 := json.Unmarshal([]byte(raw), &wrapper); err2 == nil {
			if len(wrapper.Properties) > 0 {
				return wrapper.Properties
			}
			if len(wrapper.Mappings) > 0 {
				return wrapper.Mappings
			}
		}
		logger.Warnf("[mapping] property_mappings_json 解析失败: %v", err)
		return []PropertyMapping{}
	}
	if mappings == nil {
		return []PropertyMapping{}
	}
	return mappings
}

// normalizeMappings 规整映射列表：去重、补全名称、剔除空项。
func normalizeMappings(classID uint, mappings []PropertyMapping) []PropertyMapping {
	result := make([]PropertyMapping, 0, len(mappings))
	seen := map[string]bool{}

	// 预加载属性名，便于按 property_id 补全 property_name
	names := map[uint]string{}
	types := map[uint]string{}
	if classID > 0 {
		var props []model.OntClassProperty
		DB().Select("id, name, data_type, required").Where("class_id = ?", classID).Find(&props)
		for _, p := range props {
			names[p.ID] = p.Name
			types[p.ID] = p.DataType
		}
	}

	for _, m := range mappings {
		if strings.TrimSpace(m.ColumnName) == "" && strings.TrimSpace(m.Expression) == "" &&
			strings.TrimSpace(m.TransformExpression) == "" {
			continue
		}
		if m.PropertyID > 0 {
			if strings.TrimSpace(m.PropertyName) == "" {
				m.PropertyName = names[m.PropertyID]
			}
			if strings.TrimSpace(m.DataType) == "" {
				m.DataType = types[m.PropertyID]
			}
		}
		key := strings.ToLower(m.PropertyName) + "|" + strconv.FormatUint(uint64(m.PropertyID), 10) +
			"|" + strings.ToLower(m.ColumnName)
		if seen[key] {
			continue
		}
		seen[key] = true
		m.ColumnName = strings.TrimSpace(m.ColumnName)
		result = append(result, m)
	}
	return result
}

// unmappedProperties 返回尚未配置映射的属性列表。
func unmappedProperties(properties []model.OntClassProperty, mappings []PropertyMapping) []model.OntClassProperty {
	mappedIDs := map[uint]bool{}
	mappedNames := map[string]bool{}
	for _, m := range mappings {
		if m.PropertyID > 0 {
			mappedIDs[m.PropertyID] = true
		}
		if strings.TrimSpace(m.PropertyName) != "" {
			mappedNames[strings.ToLower(strings.TrimSpace(m.PropertyName))] = true
		}
	}
	result := make([]model.OntClassProperty, 0)
	for _, p := range properties {
		if mappedIDs[p.ID] || mappedNames[strings.ToLower(p.Name)] {
			continue
		}
		result = append(result, p)
	}
	return result
}

// unmarshalUint 从原始 JSON 值解析 uint。
func unmarshalUint(v json.RawMessage) uint {
	var raw interface{}
	if err := json.Unmarshal(v, &raw); err != nil {
		return 0
	}
	n, _ := toUint(raw)
	return n
}

// unmarshalString 从原始 JSON 值解析字符串（数字也会被转换）。
func unmarshalString(v json.RawMessage) string {
	var s string
	if err := json.Unmarshal(v, &s); err == nil {
		return strings.TrimSpace(s)
	}
	var raw interface{}
	if err := json.Unmarshal(v, &raw); err == nil && raw != nil {
		return strings.TrimSpace(stringifyValue(raw))
	}
	return ""
}

// unmarshalBool 从原始 JSON 值解析布尔。
func unmarshalBool(v json.RawMessage) bool {
	var b bool
	if err := json.Unmarshal(v, &b); err == nil {
		return b
	}
	var s string
	if err := json.Unmarshal(v, &s); err == nil {
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "true", "1", "yes", "y":
			return true
		}
	}
	var n float64
	if err := json.Unmarshal(v, &n); err == nil {
		return n != 0
	}
	return false
}

// unmarshalStringMap 从原始 JSON 值解析字符串字典。
func unmarshalStringMap(v json.RawMessage) map[string]string {
	var raw map[string]interface{}
	if err := json.Unmarshal(v, &raw); err != nil {
		return nil
	}
	out := make(map[string]string, len(raw))
	for k, val := range raw {
		if val == nil {
			continue
		}
		out[k] = stringifyValue(val)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// stringifyValue 将任意 JSON 值转为字符串，整数不出现小数点。
func stringifyValue(v interface{}) string {
	switch tv := v.(type) {
	case nil:
		return ""
	case string:
		return tv
	case bool:
		return strconv.FormatBool(tv)
	case float64:
		if tv == math.Trunc(tv) && !math.IsInf(tv, 0) {
			return strconv.FormatInt(int64(tv), 10)
		}
		return strconv.FormatFloat(tv, 'f', -1, 64)
	case json.Number:
		return tv.String()
	default:
		return fmt.Sprint(tv)
	}
}

// firstUint 返回第一个大于 0 的值。
func firstUint(values ...uint) uint {
	for _, v := range values {
		if v > 0 {
			return v
		}
	}
	return 0
}

// round3 保留三位小数。
func round3(v float64) float64 {
	return math.Round(v*1000) / 1000
}

// minInt 返回较小的整数。
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// maxInt 返回较大的整数。
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
