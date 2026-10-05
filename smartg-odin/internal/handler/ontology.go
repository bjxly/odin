package handler

import (
	"encoding/json"
	"fmt"
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
// 视图与请求体
// ---------------------------------------------------------------------------

// ontologyView 本体列表项视图。
type ontologyView struct {
	model.OntDefinition
	ClassCount    int64 `json:"class_count"`
	RelationCount int64 `json:"relation_count"`
	RuleCount     int64 `json:"rule_count"`
	MappingCount  int64 `json:"mapping_count"`
}

// ontologyDetail 本体详情视图：定义 + 类 / 关系 / 规则 / 映射概览。
type ontologyDetail struct {
	model.OntDefinition
	Classes   []classView              `json:"classes"`
	Relations []model.OntRelation      `json:"relations"`
	Rules     []ruleView               `json:"rules"`
	Mappings  []model.OntMappingConfig `json:"mappings"`
}

// classView 本体类视图，附带属性数量。
type classView struct {
	model.OntClass
	PropertyCount   int64  `json:"property_count"`
	ParentClassName string `json:"parent_class_name,omitempty"`
}

// ruleView 推理规则视图，附带其动作列表。
type ruleView struct {
	model.OntRule
	Actions []model.OntAction `json:"actions"`
}

// classPayload 创建类时可内联提交属性列表。
type classPayload struct {
	model.OntClass
	Properties []model.OntClassProperty `json:"properties"`
}

// rulePayload 创建规则时可内联提交动作列表。
type rulePayload struct {
	model.OntRule
	Actions []model.OntAction `json:"actions"`
}

// relationPayload 关系请求体：额外接受类名写法，便于前端直接提交名称。
type relationPayload struct {
	model.OntRelation
	FromClassName string `json:"from_class_name"`
	ToClassName   string `json:"to_class_name"`
}

var (
	ontologyUpdatableColumns = []string{"code", "name", "description", "version", "status", "layout_json"}
	classUpdatableColumns    = []string{"name", "label", "description", "class_type",
		"parent_class_id", "icon", "color", "position_x", "position_y"}
	propertyUpdatableColumns = []string{"name", "label", "data_type", "required",
		"description", "constraints_json", "sort_order"}
	relationUpdatableColumns = []string{"name", "label", "from_class_id", "to_class_id",
		"relation_type", "cardinality", "description", "join_condition_json"}
	ruleUpdatableColumns = []string{"name", "description", "rule_type",
		"condition_json", "action_json", "priority", "enabled"}
)

// ---------------------------------------------------------------------------
// 本体 CRUD
// ---------------------------------------------------------------------------

// ListOntologies GET /api/v1/ontologies?current=1&size=20&status=&keyword=
func ListOntologies(c *gin.Context) {
	current, size := parsePage(c)

	filter := func() *gorm.DB {
		tx := DB().Model(&model.OntDefinition{})
		if s := strings.TrimSpace(c.Query("status")); s != "" {
			tx = tx.Where("status = ?", s)
		}
		if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
			like := "%" + kw + "%"
			tx = tx.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", like, like, like)
		}
		return tx
	}

	var total int64
	if err := filter().Count(&total).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	var rows []model.OntDefinition
	if err := filter().Order("id ASC").Offset((current - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	ids := make([]uint, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	classCounts := countByOntology(&model.OntClass{}, ids)
	relationCounts := countByOntology(&model.OntRelation{}, ids)
	ruleCounts := countByOntology(&model.OntRule{}, ids)
	mappingCounts := countByOntology(&model.OntMappingConfig{}, ids)

	views := make([]ontologyView, 0, len(rows))
	for _, r := range rows {
		views = append(views, ontologyView{
			OntDefinition: r,
			ClassCount:    classCounts[r.ID],
			RelationCount: relationCounts[r.ID],
			RuleCount:     ruleCounts[r.ID],
			MappingCount:  mappingCounts[r.ID],
		})
	}

	middleware.SuccessPage(c, views, total, current, size)
}

// GetOntology GET /api/v1/ontologies/:id
func GetOntology(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var ont model.OntDefinition
	if err := DB().First(&ont, id).Error; err != nil {
		fail(c, mapError(err, fmt.Sprintf("本体不存在: %d", id)))
		return
	}

	detail := ontologyDetail{
		OntDefinition: ont,
		Classes:       loadClassViews(id),
		Relations:     []model.OntRelation{},
		Rules:         loadRuleViews(id),
		Mappings:      []model.OntMappingConfig{},
	}
	DB().Where("ontology_id = ?", id).Order("id ASC").Find(&detail.Relations)
	DB().Where("ontology_id = ?", id).Order("id ASC").Find(&detail.Mappings)

	ok(c, detail)
}

// CreateOntology POST /api/v1/ontologies
func CreateOntology(c *gin.Context) {
	var ont model.OntDefinition
	if appErr := bindNormalized(c, &ont); appErr != nil {
		fail(c, appErr)
		return
	}

	ont.Name = strings.TrimSpace(ont.Name)
	if ont.Name == "" {
		fail(c, errs.ParamError("本体名称 name 不能为空"))
		return
	}
	exists := func(candidate string) bool {
		var n int64
		DB().Model(&model.OntDefinition{}).Where("code = ?", candidate).Count(&n)
		return n > 0
	}
	code := strings.TrimSpace(ont.Code)
	if code == "" {
		code = uniqueCode("ont", ont.Name, exists)
	} else if exists(code) {
		fail(c, errs.ParamError("本体编码已存在: "+code))
		return
	}

	ont.ID = 0
	ont.Code = code
	ont.Version = firstNonEmpty(strings.TrimSpace(ont.Version), "1.0.0")
	ont.Status = firstNonEmpty(strings.TrimSpace(ont.Status), "draft")
	ont.CreatedAt = time.Now()
	ont.UpdatedAt = time.Now()

	if err := DB().Create(&ont).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	writeAudit(c, "create", "ontology", ont.ID, gin.H{"code": ont.Code, "name": ont.Name})
	logger.Infof("[ontology] created id=%d code=%s", ont.ID, ont.Code)
	okMsg(c, ontologyView{OntDefinition: ont}, "创建成功")
}

// UpdateOntology PUT /api/v1/ontologies/:id
func UpdateOntology(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var ont model.OntDefinition
	if err := DB().First(&ont, id).Error; err != nil {
		fail(c, mapError(err, fmt.Sprintf("本体不存在: %d", id)))
		return
	}

	body, appErr := bindMap(c)
	if appErr != nil {
		fail(c, appErr)
		return
	}
	updates := pickUpdates(body, ontologyUpdatableColumns...)
	if code, hasCode := updates["code"]; hasCode {
		if s, isStr := code.(string); isStr && s != "" && s != ont.Code {
			var n int64
			DB().Model(&model.OntDefinition{}).Where("code = ? AND id <> ?", s, id).Count(&n)
			if n > 0 {
				fail(c, errs.ParamError("本体编码已存在: "+s))
				return
			}
		} else if s, isStr := code.(string); isStr && s == "" {
			delete(updates, "code")
		}
	}
	if len(updates) == 0 {
		fail(c, errs.ParamError("没有可识别的更新字段"))
		return
	}
	updates["updated_at"] = time.Now()

	if err := DB().Model(&model.OntDefinition{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}
	DB().First(&ont, id)

	writeAudit(c, "update", "ontology", id, gin.H{"fields": keysOf(updates)})
	okMsg(c, ontologyView{OntDefinition: ont}, "更新成功")
}

// DeleteOntology DELETE /api/v1/ontologies/:id
// 级联删除类 / 属性 / 关系 / 规则 / 动作 / 映射配置。
func DeleteOntology(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var ont model.OntDefinition
	if err := DB().First(&ont, id).Error; err != nil {
		fail(c, mapError(err, fmt.Sprintf("本体不存在: %d", id)))
		return
	}

	tx := DB().Begin()
	if tx.Error != nil {
		fail(c, errs.DBError(tx.Error))
		return
	}
	rollback := func(err error) {
		tx.Rollback()
		fail(c, errs.DBError(err))
	}

	var classIDs, ruleIDs []uint
	if err := tx.Model(&model.OntClass{}).Where("ontology_id = ?", id).Pluck("id", &classIDs).Error; err != nil {
		rollback(err)
		return
	}
	if err := tx.Model(&model.OntRule{}).Where("ontology_id = ?", id).Pluck("id", &ruleIDs).Error; err != nil {
		rollback(err)
		return
	}

	if len(classIDs) > 0 {
		if err := tx.Where("class_id IN ?", classIDs).Delete(&model.OntClassProperty{}).Error; err != nil {
			rollback(err)
			return
		}
	}
	if err := tx.Where("ontology_id = ?", id).Delete(&model.OntClassProperty{}).Error; err != nil {
		rollback(err)
		return
	}
	if len(ruleIDs) > 0 {
		if err := tx.Where("rule_id IN ?", ruleIDs).Delete(&model.OntAction{}).Error; err != nil {
			rollback(err)
			return
		}
	}
	if err := tx.Where("ontology_id = ?", id).Delete(&model.OntRule{}).Error; err != nil {
		rollback(err)
		return
	}
	if err := tx.Where("ontology_id = ?", id).Delete(&model.OntRelation{}).Error; err != nil {
		rollback(err)
		return
	}
	if err := tx.Where("ontology_id = ?", id).Delete(&model.OntClass{}).Error; err != nil {
		rollback(err)
		return
	}
	if err := tx.Where("ontology_id = ?", id).Delete(&model.OntMappingConfig{}).Error; err != nil {
		rollback(err)
		return
	}
	if err := tx.Delete(&model.OntDefinition{}, id).Error; err != nil {
		rollback(err)
		return
	}
	if err := tx.Commit().Error; err != nil {
		rollback(err)
		return
	}

	writeAudit(c, "delete", "ontology", id, gin.H{"code": ont.Code, "name": ont.Name})
	okMsg(c, gin.H{
		"id":      id,
		"classes": len(classIDs),
		"rules":   len(ruleIDs),
		"cascade": true,
	}, "删除成功")
}

// ---------------------------------------------------------------------------
// Classes 子资源
// ---------------------------------------------------------------------------

// ListClasses GET /api/v1/ontologies/:id/classes
func ListClasses(c *gin.Context) {
	ontologyID, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}
	if appErr := ensureOntology(ontologyID); appErr != nil {
		fail(c, appErr)
		return
	}
	ok(c, loadClassViews(ontologyID))
}

// CreateClass POST /api/v1/ontologies/:id/classes
func CreateClass(c *gin.Context) {
	ontologyID, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}
	if appErr := ensureOntology(ontologyID); appErr != nil {
		fail(c, appErr)
		return
	}

	var req classPayload
	if appErr := bindNormalized(c, &req); appErr != nil {
		fail(c, appErr)
		return
	}
	req.OntClass.Name = strings.TrimSpace(req.OntClass.Name)
	if req.OntClass.Name == "" {
		fail(c, errs.ParamError("类名 name 不能为空"))
		return
	}
	if appErr := ensureClassUnique(ontologyID, req.OntClass.Name, 0); appErr != nil {
		fail(c, appErr)
		return
	}

	cls := req.OntClass
	cls.ID = 0
	cls.OntologyID = ontologyID
	cls.ClassType = firstNonEmpty(strings.TrimSpace(cls.ClassType), "normal")
	if cls.Label == "" {
		cls.Label = cls.Name
	}
	cls.CreatedAt = time.Now()
	cls.UpdatedAt = time.Now()

	tx := DB().Begin()
	if tx.Error != nil {
		fail(c, errs.DBError(tx.Error))
		return
	}
	if err := tx.Create(&cls).Error; err != nil {
		tx.Rollback()
		fail(c, errs.DBError(err))
		return
	}
	for i := range req.Properties {
		prop := req.Properties[i]
		prop.ID = 0
		prop.ClassID = cls.ID
		prop.OntologyID = ontologyID
		prop.DataType = firstNonEmpty(strings.TrimSpace(prop.DataType), "string")
		if prop.SortOrder == 0 {
			prop.SortOrder = i + 1
		}
		if err := tx.Create(&prop).Error; err != nil {
			tx.Rollback()
			fail(c, errs.DBError(err))
			return
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		fail(c, errs.DBError(err))
		return
	}

	touchOntology(ontologyID)
	writeAudit(c, "create", "ontology_class", cls.ID, gin.H{"ontology_id": ontologyID, "name": cls.Name})
	okMsg(c, cls, "创建成功")
}

// UpdateClass PUT /api/v1/ontologies/:id/classes/:classId
func UpdateClass(c *gin.Context) {
	ontologyID, classID, appErr := parseNestedIDs(c, "id", "classId")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var cls model.OntClass
	if err := DB().Where("id = ? AND ontology_id = ?", classID, ontologyID).First(&cls).Error; err != nil {
		fail(c, mapError(err, fmt.Sprintf("类不存在: %d", classID)))
		return
	}

	body, appErr := bindMap(c)
	if appErr != nil {
		fail(c, appErr)
		return
	}
	updates := pickUpdates(body, classUpdatableColumns...)
	if name, hasName := updates["name"]; hasName {
		if s, isStr := name.(string); isStr {
			if strings.TrimSpace(s) == "" {
				fail(c, errs.ParamError("类名 name 不能为空"))
				return
			}
			if appErr := ensureClassUnique(ontologyID, strings.TrimSpace(s), classID); appErr != nil {
				fail(c, appErr)
				return
			}
		}
	}
	if len(updates) == 0 {
		fail(c, errs.ParamError("没有可识别的更新字段"))
		return
	}
	updates["updated_at"] = time.Now()

	if err := DB().Model(&model.OntClass{}).Where("id = ?", classID).Updates(updates).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}
	DB().Where("id = ?", classID).First(&cls)

	touchOntology(ontologyID)
	writeAudit(c, "update", "ontology_class", classID, gin.H{"fields": keysOf(updates)})
	okMsg(c, cls, "更新成功")
}

// DeleteClass DELETE /api/v1/ontologies/:id/classes/:classId
// 级联删除该类属性、引用该类的关系与该类的映射配置。
func DeleteClass(c *gin.Context) {
	ontologyID, classID, appErr := parseNestedIDs(c, "id", "classId")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var cls model.OntClass
	if err := DB().Where("id = ? AND ontology_id = ?", classID, ontologyID).First(&cls).Error; err != nil {
		fail(c, mapError(err, fmt.Sprintf("类不存在: %d", classID)))
		return
	}

	tx := DB().Begin()
	if tx.Error != nil {
		fail(c, errs.DBError(tx.Error))
		return
	}
	propRes := tx.Where("class_id = ?", classID).Delete(&model.OntClassProperty{})
	if propRes.Error != nil {
		tx.Rollback()
		fail(c, errs.DBError(propRes.Error))
		return
	}
	relRes := tx.Where("from_class_id = ? OR to_class_id = ?", classID, classID).Delete(&model.OntRelation{})
	if relRes.Error != nil {
		tx.Rollback()
		fail(c, errs.DBError(relRes.Error))
		return
	}
	mapRes := tx.Where("class_id = ?", classID).Delete(&model.OntMappingConfig{})
	if mapRes.Error != nil {
		tx.Rollback()
		fail(c, errs.DBError(mapRes.Error))
		return
	}
	if err := tx.Delete(&model.OntClass{}, classID).Error; err != nil {
		tx.Rollback()
		fail(c, errs.DBError(err))
		return
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		fail(c, errs.DBError(err))
		return
	}

	touchOntology(ontologyID)
	writeAudit(c, "delete", "ontology_class", classID, gin.H{"name": cls.Name, "properties": propRes.RowsAffected})
	okMsg(c, gin.H{
		"id":         classID,
		"properties": propRes.RowsAffected,
		"relations":  relRes.RowsAffected,
		"mappings":   mapRes.RowsAffected,
	}, "删除成功")
}

// ---------------------------------------------------------------------------
// Properties 子资源
// ---------------------------------------------------------------------------

// ListProperties GET /api/v1/ontologies/:id/classes/:classId/properties
func ListProperties(c *gin.Context) {
	ontologyID, classID, appErr := parseNestedIDs(c, "id", "classId")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var props []model.OntClassProperty
	if err := DB().Where("class_id = ? AND ontology_id = ?", classID, ontologyID).
		Order("sort_order ASC, id ASC").Find(&props).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}
	ok(c, props)
}

// CreateProperty POST /api/v1/ontologies/:id/classes/:classId/properties
func CreateProperty(c *gin.Context) {
	ontologyID, classID, appErr := parseNestedIDs(c, "id", "classId")
	if appErr != nil {
		fail(c, appErr)
		return
	}
	if appErr := ensureClass(ontologyID, classID); appErr != nil {
		fail(c, appErr)
		return
	}

	var prop model.OntClassProperty
	if appErr := bindNormalized(c, &prop); appErr != nil {
		fail(c, appErr)
		return
	}
	prop.Name = strings.TrimSpace(prop.Name)
	if prop.Name == "" {
		fail(c, errs.ParamError("属性名 name 不能为空"))
		return
	}
	if appErr := ensurePropertyUnique(classID, prop.Name, 0); appErr != nil {
		fail(c, appErr)
		return
	}

	prop.ID = 0
	prop.ClassID = classID
	prop.OntologyID = ontologyID
	prop.DataType = firstNonEmpty(strings.TrimSpace(prop.DataType), "string")
	if prop.Label == "" {
		prop.Label = prop.Name
	}
	if prop.SortOrder == 0 {
		var n int64
		DB().Model(&model.OntClassProperty{}).Where("class_id = ?", classID).Count(&n)
		prop.SortOrder = int(n) + 1
	}

	if err := DB().Create(&prop).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	touchOntology(ontologyID)
	writeAudit(c, "create", "ontology_property", prop.ID, gin.H{"class_id": classID, "name": prop.Name})
	okMsg(c, prop, "创建成功")
}

// UpdateProperty PUT /api/v1/ontologies/:id/classes/:classId/properties/:propId
func UpdateProperty(c *gin.Context) {
	ontologyID, classID, propID, appErr := parseTripleIDs(c, "id", "classId", "propId")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var prop model.OntClassProperty
	if err := DB().Where("id = ? AND class_id = ? AND ontology_id = ?", propID, classID, ontologyID).
		First(&prop).Error; err != nil {
		fail(c, mapError(err, fmt.Sprintf("属性不存在: %d", propID)))
		return
	}

	body, appErr := bindMap(c)
	if appErr != nil {
		fail(c, appErr)
		return
	}
	updates := pickUpdates(body, propertyUpdatableColumns...)
	if name, hasName := updates["name"]; hasName {
		s, isStr := name.(string)
		if !isStr || strings.TrimSpace(s) == "" {
			fail(c, errs.ParamError("属性名 name 不能为空"))
			return
		}
		if appErr := ensurePropertyUnique(classID, strings.TrimSpace(s), propID); appErr != nil {
			fail(c, appErr)
			return
		}
	}
	if len(updates) == 0 {
		fail(c, errs.ParamError("没有可识别的更新字段"))
		return
	}

	if err := DB().Model(&model.OntClassProperty{}).Where("id = ?", propID).Updates(updates).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}
	DB().Where("id = ?", propID).First(&prop)

	touchOntology(ontologyID)
	writeAudit(c, "update", "ontology_property", propID, gin.H{"fields": keysOf(updates)})
	okMsg(c, prop, "更新成功")
}

// DeleteProperty DELETE /api/v1/ontologies/:id/classes/:classId/properties/:propId
func DeleteProperty(c *gin.Context) {
	ontologyID, classID, propID, appErr := parseTripleIDs(c, "id", "classId", "propId")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var prop model.OntClassProperty
	if err := DB().Where("id = ? AND class_id = ? AND ontology_id = ?", propID, classID, ontologyID).
		First(&prop).Error; err != nil {
		fail(c, mapError(err, fmt.Sprintf("属性不存在: %d", propID)))
		return
	}
	if err := DB().Delete(&model.OntClassProperty{}, propID).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	touchOntology(ontologyID)
	writeAudit(c, "delete", "ontology_property", propID, gin.H{"name": prop.Name})
	okMsg(c, gin.H{"id": propID}, "删除成功")
}

// ---------------------------------------------------------------------------
// Relations 子资源
// ---------------------------------------------------------------------------

// ListRelations GET /api/v1/ontologies/:id/relations
func ListRelations(c *gin.Context) {
	ontologyID, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var relations []model.OntRelation
	if err := DB().Where("ontology_id = ?", ontologyID).Order("id ASC").Find(&relations).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	// 附带两端类名，避免前端再次查询
	names := loadClassNames(ontologyID)
	views := make([]gin.H, 0, len(relations))
	for _, r := range relations {
		views = append(views, gin.H{
			"relation":        r,
			"from_class_name": names[r.FromClassID],
			"to_class_name":   names[r.ToClassID],
		})
	}
	ok(c, gin.H{"relations": relations, "items": views})
}

// CreateRelation POST /api/v1/ontologies/:id/relations
func CreateRelation(c *gin.Context) {
	ontologyID, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}
	if appErr := ensureOntology(ontologyID); appErr != nil {
		fail(c, appErr)
		return
	}

	var req relationPayload
	if appErr := bindNormalized(c, &req); appErr != nil {
		fail(c, appErr)
		return
	}
	rel := req.OntRelation
	rel.Name = strings.TrimSpace(rel.Name)
	if rel.Name == "" {
		fail(c, errs.ParamError("关系名 name 不能为空"))
		return
	}
	// 支持通过类名解析两端类 ID
	names := loadClassNames(ontologyID)
	idByName := make(map[string]uint, len(names))
	for cid, nm := range names {
		idByName[nm] = cid
	}
	if rel.FromClassID == 0 && req.FromClassName != "" {
		rel.FromClassID = idByName[strings.TrimSpace(req.FromClassName)]
	}
	if rel.ToClassID == 0 && req.ToClassName != "" {
		rel.ToClassID = idByName[strings.TrimSpace(req.ToClassName)]
	}
	if rel.FromClassID == 0 || rel.ToClassID == 0 {
		fail(c, errs.ParamError("关系两端 from_class_id / to_class_id 必须指向已存在的类"))
		return
	}
	if _, exists := names[rel.FromClassID]; !exists {
		fail(c, errs.ParamError(fmt.Sprintf("起始类不存在: %d", rel.FromClassID)))
		return
	}
	if _, exists := names[rel.ToClassID]; !exists {
		fail(c, errs.ParamError(fmt.Sprintf("目标类不存在: %d", rel.ToClassID)))
		return
	}

	rel.ID = 0
	rel.OntologyID = ontologyID
	rel.RelationType = firstNonEmpty(strings.TrimSpace(rel.RelationType), "association")
	rel.Cardinality = firstNonEmpty(strings.TrimSpace(rel.Cardinality), "1:N")
	if rel.Label == "" {
		rel.Label = rel.Name
	}

	if err := DB().Create(&rel).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	touchOntology(ontologyID)
	writeAudit(c, "create", "ontology_relation", rel.ID, gin.H{"ontology_id": ontologyID, "name": rel.Name})
	okMsg(c, rel, "创建成功")
}

// UpdateRelation PUT /api/v1/ontologies/:id/relations/:relId
func UpdateRelation(c *gin.Context) {
	ontologyID, relID, appErr := parseNestedIDs(c, "id", "relId")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var rel model.OntRelation
	if err := DB().Where("id = ? AND ontology_id = ?", relID, ontologyID).First(&rel).Error; err != nil {
		fail(c, mapError(err, fmt.Sprintf("关系不存在: %d", relID)))
		return
	}

	body, appErr := bindMap(c)
	if appErr != nil {
		fail(c, appErr)
		return
	}
	updates := pickUpdates(body, relationUpdatableColumns...)
	for _, key := range []string{"from_class_id", "to_class_id"} {
		if v, has := updates[key]; has {
			classID, isNum := toUint(v)
			if !isNum || classID == 0 {
				fail(c, errs.ParamError(key+" 必须为有效的类 ID"))
				return
			}
			var n int64
			DB().Model(&model.OntClass{}).Where("id = ? AND ontology_id = ?", classID, ontologyID).Count(&n)
			if n == 0 {
				fail(c, errs.ParamError(fmt.Sprintf("类不存在: %d", classID)))
				return
			}
		}
	}
	if len(updates) == 0 {
		fail(c, errs.ParamError("没有可识别的更新字段"))
		return
	}

	if err := DB().Model(&model.OntRelation{}).Where("id = ?", relID).Updates(updates).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}
	DB().Where("id = ?", relID).First(&rel)

	touchOntology(ontologyID)
	writeAudit(c, "update", "ontology_relation", relID, gin.H{"fields": keysOf(updates)})
	okMsg(c, rel, "更新成功")
}

// DeleteRelation DELETE /api/v1/ontologies/:id/relations/:relId
func DeleteRelation(c *gin.Context) {
	ontologyID, relID, appErr := parseNestedIDs(c, "id", "relId")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var rel model.OntRelation
	if err := DB().Where("id = ? AND ontology_id = ?", relID, ontologyID).First(&rel).Error; err != nil {
		fail(c, mapError(err, fmt.Sprintf("关系不存在: %d", relID)))
		return
	}
	if err := DB().Delete(&model.OntRelation{}, relID).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	touchOntology(ontologyID)
	writeAudit(c, "delete", "ontology_relation", relID, gin.H{"name": rel.Name})
	okMsg(c, gin.H{"id": relID}, "删除成功")
}

// ---------------------------------------------------------------------------
// Rules 子资源
// ---------------------------------------------------------------------------

// ListRules GET /api/v1/ontologies/:id/rules
func ListRules(c *gin.Context) {
	ontologyID, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}
	ok(c, loadRuleViews(ontologyID))
}

// CreateRule POST /api/v1/ontologies/:id/rules
func CreateRule(c *gin.Context) {
	ontologyID, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}
	if appErr := ensureOntology(ontologyID); appErr != nil {
		fail(c, appErr)
		return
	}

	var req rulePayload
	if appErr := bindNormalized(c, &req); appErr != nil {
		fail(c, appErr)
		return
	}
	rule := req.OntRule
	rule.Name = strings.TrimSpace(rule.Name)
	if rule.Name == "" {
		fail(c, errs.ParamError("规则名 name 不能为空"))
		return
	}
	if strings.TrimSpace(rule.RuleType) == "" {
		fail(c, errs.ParamError("规则类型 rule_type 不能为空（derivation / constraint / inference）"))
		return
	}

	rule.ID = 0
	rule.OntologyID = ontologyID
	if rule.ConditionJSON == "" {
		rule.ConditionJSON = "{}"
	}
	if rule.ActionJSON == "" {
		rule.ActionJSON = "{}"
	}

	tx := DB().Begin()
	if tx.Error != nil {
		fail(c, errs.DBError(tx.Error))
		return
	}
	if err := tx.Create(&rule).Error; err != nil {
		tx.Rollback()
		fail(c, errs.DBError(err))
		return
	}
	for i := range req.Actions {
		action := req.Actions[i]
		action.ID = 0
		action.RuleID = rule.ID
		if strings.TrimSpace(action.ActionType) == "" {
			continue
		}
		if err := tx.Create(&action).Error; err != nil {
			tx.Rollback()
			fail(c, errs.DBError(err))
			return
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		fail(c, errs.DBError(err))
		return
	}

	touchOntology(ontologyID)
	writeAudit(c, "create", "ontology_rule", rule.ID, gin.H{"ontology_id": ontologyID, "name": rule.Name})
	okMsg(c, ruleView{OntRule: rule, Actions: req.Actions}, "创建成功")
}

// UpdateRule PUT /api/v1/ontologies/:id/rules/:ruleId
func UpdateRule(c *gin.Context) {
	ontologyID, ruleID, appErr := parseNestedIDs(c, "id", "ruleId")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var rule model.OntRule
	if err := DB().Where("id = ? AND ontology_id = ?", ruleID, ontologyID).First(&rule).Error; err != nil {
		fail(c, mapError(err, fmt.Sprintf("规则不存在: %d", ruleID)))
		return
	}

	body, appErr := bindMap(c)
	if appErr != nil {
		fail(c, appErr)
		return
	}
	updates := pickUpdates(body, ruleUpdatableColumns...)
	if name, hasName := updates["name"]; hasName {
		if s, isStr := name.(string); !isStr || strings.TrimSpace(s) == "" {
			fail(c, errs.ParamError("规则名 name 不能为空"))
			return
		}
	}
	if len(updates) == 0 && body["actions"] == nil {
		fail(c, errs.ParamError("没有可识别的更新字段"))
		return
	}

	tx := DB().Begin()
	if tx.Error != nil {
		fail(c, errs.DBError(tx.Error))
		return
	}
	if len(updates) > 0 {
		if err := tx.Model(&model.OntRule{}).Where("id = ?", ruleID).Updates(updates).Error; err != nil {
			tx.Rollback()
			fail(c, errs.DBError(err))
			return
		}
	}
	// 提交 actions 时整体替换该规则的动作列表
	if raw, hasActions := body["actions"]; hasActions {
		if err := tx.Where("rule_id = ?", ruleID).Delete(&model.OntAction{}).Error; err != nil {
			tx.Rollback()
			fail(c, errs.DBError(err))
			return
		}
		if list, isList := raw.([]interface{}); isList {
			buf := marshalJSON(list)
			var actions []model.OntAction
			if err := json.Unmarshal([]byte(buf), &actions); err != nil {
				tx.Rollback()
				fail(c, errs.ParamError("actions 结构解析失败: "+err.Error()))
				return
			}
			for i := range actions {
				action := actions[i]
				action.ID = 0
				action.RuleID = ruleID
				if strings.TrimSpace(action.ActionType) == "" {
					continue
				}
				if err := tx.Create(&action).Error; err != nil {
					tx.Rollback()
					fail(c, errs.DBError(err))
					return
				}
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		fail(c, errs.DBError(err))
		return
	}

	DB().Where("id = ?", ruleID).First(&rule)
	touchOntology(ontologyID)
	writeAudit(c, "update", "ontology_rule", ruleID, gin.H{"fields": keysOf(updates)})
	okMsg(c, ruleView{OntRule: rule, Actions: loadActions(ruleID)}, "更新成功")
}

// DeleteRule DELETE /api/v1/ontologies/:id/rules/:ruleId
// 级联删除规则下的动作。
func DeleteRule(c *gin.Context) {
	ontologyID, ruleID, appErr := parseNestedIDs(c, "id", "ruleId")
	if appErr != nil {
		fail(c, appErr)
		return
	}

	var rule model.OntRule
	if err := DB().Where("id = ? AND ontology_id = ?", ruleID, ontologyID).First(&rule).Error; err != nil {
		fail(c, mapError(err, fmt.Sprintf("规则不存在: %d", ruleID)))
		return
	}

	tx := DB().Begin()
	if tx.Error != nil {
		fail(c, errs.DBError(tx.Error))
		return
	}
	actionRes := tx.Where("rule_id = ?", ruleID).Delete(&model.OntAction{})
	if actionRes.Error != nil {
		tx.Rollback()
		fail(c, errs.DBError(actionRes.Error))
		return
	}
	if err := tx.Delete(&model.OntRule{}, ruleID).Error; err != nil {
		tx.Rollback()
		fail(c, errs.DBError(err))
		return
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		fail(c, errs.DBError(err))
		return
	}

	touchOntology(ontologyID)
	writeAudit(c, "delete", "ontology_rule", ruleID, gin.H{"name": rule.Name, "actions": actionRes.RowsAffected})
	okMsg(c, gin.H{"id": ruleID, "actions": actionRes.RowsAffected}, "删除成功")
}

// ---------------------------------------------------------------------------
// 内部辅助
// ---------------------------------------------------------------------------

// ensureOntology 校验本体存在。
func ensureOntology(ontologyID uint) *errs.AppError {
	var n int64
	if err := DB().Model(&model.OntDefinition{}).Where("id = ?", ontologyID).Count(&n).Error; err != nil {
		return errs.DBError(err)
	}
	if n == 0 {
		return errs.NotFound(fmt.Sprintf("本体不存在: %d", ontologyID))
	}
	return nil
}

// ensureClass 校验类归属于指定本体。
func ensureClass(ontologyID, classID uint) *errs.AppError {
	var n int64
	if err := DB().Model(&model.OntClass{}).
		Where("id = ? AND ontology_id = ?", classID, ontologyID).Count(&n).Error; err != nil {
		return errs.DBError(err)
	}
	if n == 0 {
		return errs.NotFound(fmt.Sprintf("类不存在: %d", classID))
	}
	return nil
}

// ensureClassUnique 校验同一本体下类名唯一。
func ensureClassUnique(ontologyID uint, name string, excludeID uint) *errs.AppError {
	tx := DB().Model(&model.OntClass{}).Where("ontology_id = ? AND name = ?", ontologyID, name)
	if excludeID > 0 {
		tx = tx.Where("id <> ?", excludeID)
	}
	var n int64
	if err := tx.Count(&n).Error; err != nil {
		return errs.DBError(err)
	}
	if n > 0 {
		return errs.ParamError("同一本体下类名已存在: " + name)
	}
	return nil
}

// ensurePropertyUnique 校验同一类下属性名唯一。
func ensurePropertyUnique(classID uint, name string, excludeID uint) *errs.AppError {
	tx := DB().Model(&model.OntClassProperty{}).Where("class_id = ? AND name = ?", classID, name)
	if excludeID > 0 {
		tx = tx.Where("id <> ?", excludeID)
	}
	var n int64
	if err := tx.Count(&n).Error; err != nil {
		return errs.DBError(err)
	}
	if n > 0 {
		return errs.ParamError("同一类下属性名已存在: " + name)
	}
	return nil
}

// loadClassViews 加载本体下全部类（含属性数量与父类名）。
func loadClassViews(ontologyID uint) []classView {
	var classes []model.OntClass
	DB().Where("ontology_id = ?", ontologyID).Order("id ASC").Find(&classes)
	if len(classes) == 0 {
		return []classView{}
	}

	ids := make([]uint, 0, len(classes))
	names := make(map[uint]string, len(classes))
	for _, cls := range classes {
		ids = append(ids, cls.ID)
		names[cls.ID] = cls.Name
	}

	type row struct {
		ClassID uint
		N       int64
	}
	var rows []row
	DB().Model(&model.OntClassProperty{}).
		Select("class_id, COUNT(*) AS n").
		Where("class_id IN ?", ids).
		Group("class_id").
		Scan(&rows)
	counts := make(map[uint]int64, len(rows))
	for _, r := range rows {
		counts[r.ClassID] = r.N
	}

	views := make([]classView, 0, len(classes))
	for _, cls := range classes {
		view := classView{OntClass: cls, PropertyCount: counts[cls.ID]}
		if cls.ParentClassID != nil && *cls.ParentClassID > 0 {
			view.ParentClassName = names[*cls.ParentClassID]
		}
		views = append(views, view)
	}
	return views
}

// loadClassNames 加载本体下类 ID → 类名映射。
func loadClassNames(ontologyID uint) map[uint]string {
	var classes []model.OntClass
	DB().Select("id, name").Where("ontology_id = ?", ontologyID).Find(&classes)
	names := make(map[uint]string, len(classes))
	for _, cls := range classes {
		names[cls.ID] = cls.Name
	}
	return names
}

// loadRuleViews 加载本体下全部规则（含动作列表）。
func loadRuleViews(ontologyID uint) []ruleView {
	var rules []model.OntRule
	DB().Where("ontology_id = ?", ontologyID).Order("priority DESC, id ASC").Find(&rules)
	if len(rules) == 0 {
		return []ruleView{}
	}

	ids := make([]uint, 0, len(rules))
	for _, r := range rules {
		ids = append(ids, r.ID)
	}
	var actions []model.OntAction
	DB().Where("rule_id IN ?", ids).Order("id ASC").Find(&actions)
	grouped := make(map[uint][]model.OntAction, len(rules))
	for _, a := range actions {
		grouped[a.RuleID] = append(grouped[a.RuleID], a)
	}

	views := make([]ruleView, 0, len(rules))
	for _, r := range rules {
		list := grouped[r.ID]
		if list == nil {
			list = []model.OntAction{}
		}
		views = append(views, ruleView{OntRule: r, Actions: list})
	}
	return views
}

// loadActions 加载单个规则的动作列表。
func loadActions(ruleID uint) []model.OntAction {
	var actions []model.OntAction
	DB().Where("rule_id = ?", ruleID).Order("id ASC").Find(&actions)
	if actions == nil {
		return []model.OntAction{}
	}
	return actions
}

// touchOntology 更新本体的 updated_at 时间戳。
func touchOntology(ontologyID uint) {
	if err := DB().Model(&model.OntDefinition{}).Where("id = ?", ontologyID).
		Update("updated_at", time.Now()).Error; err != nil {
		logger.Warnf("[ontology] 更新时间戳失败 id=%d err=%v", ontologyID, err)
	}
}

// parseNestedIDs 解析两级嵌套资源 ID。
func parseNestedIDs(c *gin.Context, first, second string) (uint, uint, *errs.AppError) {
	a, appErr := parseIDParam(c, first)
	if appErr != nil {
		return 0, 0, appErr
	}
	b, appErr := parseIDParam(c, second)
	if appErr != nil {
		return 0, 0, appErr
	}
	return a, b, nil
}

// parseTripleIDs 解析三级嵌套资源 ID。
func parseTripleIDs(c *gin.Context, first, second, third string) (uint, uint, uint, *errs.AppError) {
	a, b, appErr := parseNestedIDs(c, first, second)
	if appErr != nil {
		return 0, 0, 0, appErr
	}
	d, appErr := parseIDParam(c, third)
	if appErr != nil {
		return 0, 0, 0, appErr
	}
	return a, b, d, nil
}

// countByOntology 按 ontology_id 分组统计数量。
func countByOntology(dest interface{}, ids []uint) map[uint]int64 {
	result := map[uint]int64{}
	if len(ids) == 0 {
		return result
	}
	type row struct {
		OntologyID uint
		N          int64
	}
	var rows []row
	if err := DB().Model(dest).
		Select("ontology_id, COUNT(*) AS n").
		Where("ontology_id IN ?", ids).
		Group("ontology_id").
		Scan(&rows).Error; err != nil {
		logger.Warnf("[ontology] 统计数量失败: %v", err)
		return result
	}
	for _, r := range rows {
		result[r.OntologyID] = r.N
	}
	return result
}
