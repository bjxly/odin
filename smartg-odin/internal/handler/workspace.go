package handler

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"

	"smartg-odin/internal/common/errs"
	"smartg-odin/internal/common/logger"
	"smartg-odin/internal/model"
)

// ---------------------------------------------------------------------------
// Workspace 兼容端点（deprecated）
//
// 旧版前端通过 GET/PUT /api/v1/workspace 一次性加载/保存整个工作区。
// 新架构已将数据拆分到 datasources / ontologies / mappings 等规范化表中。
// 以下实现从规范化表聚合数据返回，PUT 做 best-effort upsert。
// ---------------------------------------------------------------------------

// workspaceDataSource GET 响应中的数据源摘要。
type workspaceDataSource struct {
	ID           uint     `json:"id"`
	Code         string   `json:"code"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Host         string   `json:"host,omitempty"`
	Port         int      `json:"port,omitempty"`
	DatabaseName string   `json:"database_name,omitempty"`
	Username     string   `json:"username,omitempty"`
	SchemaName   string   `json:"schema_name,omitempty"`
	Status       string   `json:"status"`
	Tags         []string `json:"tags,omitempty"`
	Description  string   `json:"description,omitempty"`
	TableCount   int64    `json:"table_count"`
	CreatedAt    int64    `json:"created_at"`
	UpdatedAt    int64    `json:"updated_at"`
}

// workspaceClass 类摘要。
type workspaceClass struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Label         string `json:"label,omitempty"`
	Description   string `json:"description,omitempty"`
	ClassType     string `json:"class_type,omitempty"`
	ParentClassID uint   `json:"parent_class_id,omitempty"`
	PropertyCount int64  `json:"property_count"`
}

// workspaceRelation 关系摘要。
type workspaceRelation struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Label        string `json:"label,omitempty"`
	FromClassID  uint   `json:"from_class_id"`
	FromClass    string `json:"from_class,omitempty"`
	ToClassID    uint   `json:"to_class_id"`
	ToClass      string `json:"to_class,omitempty"`
	RelationType string `json:"relation_type,omitempty"`
	Cardinality  string `json:"cardinality,omitempty"`
	Description  string `json:"description,omitempty"`
}

// workspaceOntology 本体摘要（含 classes 和 relations）。
type workspaceOntology struct {
	ID          uint                `json:"id"`
	Code        string              `json:"code"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Version     string              `json:"version,omitempty"`
	Status      string              `json:"status"`
	ClassCount  int64               `json:"class_count"`
	Classes     []workspaceClass    `json:"classes"`
	Relations   []workspaceRelation `json:"relations"`
	CreatedAt   int64               `json:"created_at"`
	UpdatedAt   int64               `json:"updated_at"`
}

// workspaceMapping 映射摘要。
type workspaceMapping struct {
	ID             uint   `json:"id"`
	OntologyID     uint   `json:"ontology_id"`
	ClassID        uint   `json:"class_id"`
	ClassName      string `json:"class_name,omitempty"`
	DataSourceID   uint   `json:"datasource_id"`
	DataSourceName string `json:"datasource_name,omitempty"`
	SourceTable    string `json:"source_table"`
	Status         string `json:"status"`
}

// workspaceResponse GET /api/v1/workspace 的完整聚合响应。
type workspaceResponse struct {
	Datasources []workspaceDataSource `json:"datasources"`
	Ontologies  []workspaceOntology   `json:"ontologies"`
	Mappings    []workspaceMapping    `json:"mappings"`
	GeneratedAt int64                 `json:"generated_at"`
}

// GetWorkspace GET /api/v1/workspace
// 从规范化表聚合 datasources + ontologies(含 classes/relations) + mappings 返回。
func GetWorkspace(c *gin.Context) {
	db := DB()
	resp := workspaceResponse{
		Datasources: []workspaceDataSource{},
		Ontologies:  []workspaceOntology{},
		Mappings:    []workspaceMapping{},
		GeneratedAt: time.Now().UnixMilli(),
	}

	// 1. 数据源
	var dataSources []model.DataSource
	db.Order("id ASC").Find(&dataSources)
	for _, ds := range dataSources {
		var tableCount int64
		db.Model(&model.SchemaTable{}).Where("datasource_id = ?", ds.ID).Count(&tableCount)
		resp.Datasources = append(resp.Datasources, workspaceDataSource{
			ID:           ds.ID,
			Code:         ds.Code,
			Name:         ds.Name,
			Type:         ds.Type,
			Host:         ds.Host,
			Port:         ds.Port,
			DatabaseName: ds.DatabaseName,
			Username:     ds.Username,
			SchemaName:   ds.SchemaName,
			Status:       ds.Status,
			Tags:         splitTags(ds.Tags),
			Description:  ds.Description,
			TableCount:   tableCount,
			CreatedAt:    ds.CreatedAt.UnixMilli(),
			UpdatedAt:    ds.UpdatedAt.UnixMilli(),
		})
	}

	// 2. 本体（含 classes + relations）
	var ontologies []model.OntDefinition
	db.Order("id ASC").Find(&ontologies)
	for _, ont := range ontologies {
		wOnt := workspaceOntology{
			ID:          ont.ID,
			Code:        ont.Code,
			Name:        ont.Name,
			Description: ont.Description,
			Version:     ont.Version,
			Status:      ont.Status,
			Classes:     []workspaceClass{},
			Relations:   []workspaceRelation{},
			CreatedAt:   ont.CreatedAt.UnixMilli(),
			UpdatedAt:   ont.UpdatedAt.UnixMilli(),
		}

		var classes []model.OntClass
		db.Where("ontology_id = ?", ont.ID).Order("id ASC").Find(&classes)
		for _, cls := range classes {
			var propCount int64
			db.Model(&model.OntClassProperty{}).Where("class_id = ?", cls.ID).Count(&propCount)
			var parentID uint
			if cls.ParentClassID != nil {
				parentID = *cls.ParentClassID
			}
			wOnt.Classes = append(wOnt.Classes, workspaceClass{
				ID:            cls.ID,
				Name:          cls.Name,
				Label:         cls.Label,
				Description:   cls.Description,
				ClassType:     cls.ClassType,
				ParentClassID: parentID,
				PropertyCount: propCount,
			})
		}
		wOnt.ClassCount = int64(len(classes))

		var relations []model.OntRelation
		db.Where("ontology_id = ?", ont.ID).Order("id ASC").Find(&relations)
		// 构建类 ID→名称映射
		classNameMap := make(map[uint]string, len(classes))
		for _, cls := range classes {
			classNameMap[cls.ID] = cls.Name
		}
		for _, rel := range relations {
			wOnt.Relations = append(wOnt.Relations, workspaceRelation{
				ID:           rel.ID,
				Name:         rel.Name,
				Label:        rel.Label,
				FromClassID:  rel.FromClassID,
				FromClass:    classNameMap[rel.FromClassID],
				ToClassID:    rel.ToClassID,
				ToClass:      classNameMap[rel.ToClassID],
				RelationType: rel.RelationType,
				Cardinality:  rel.Cardinality,
				Description:  rel.Description,
			})
		}

		resp.Ontologies = append(resp.Ontologies, wOnt)
	}

	// 3. 映射配置
	var mappings []model.OntMappingConfig
	db.Order("id ASC").Find(&mappings)
	// 预加载数据源名称 & 类名称
	dsNames := make(map[uint]string, len(dataSources))
	for _, ds := range dataSources {
		dsNames[ds.ID] = ds.Name
	}
	var allClasses []model.OntClass
	db.Find(&allClasses)
	classNames := make(map[uint]string, len(allClasses))
	for _, cls := range allClasses {
		classNames[cls.ID] = cls.Name
	}
	for _, m := range mappings {
		resp.Mappings = append(resp.Mappings, workspaceMapping{
			ID:             m.ID,
			OntologyID:     m.OntologyID,
			ClassID:        m.ClassID,
			ClassName:      classNames[m.ClassID],
			DataSourceID:   m.DataSourceID,
			DataSourceName: dsNames[m.DataSourceID],
			SourceTable:    m.SourceTable,
			Status:         m.Status,
		})
	}

	ok(c, resp)
}

// workspaceSaveRequest PUT 请求体（best-effort）。
type workspaceSaveRequest struct {
	Datasources []json.RawMessage `json:"datasources"`
	Ontologies  []json.RawMessage `json:"ontologies"`
	Mappings    []json.RawMessage `json:"mappings"`
}

// SaveWorkspace PUT /api/v1/workspace
// Best-effort 保存：尝试从大 JSON 中解析并 upsert 数据源。
// 完整的数据写入建议使用各子资源的专用端点。
func SaveWorkspace(c *gin.Context) {
	var req workspaceSaveRequest
	if appErr := bindJSON(c, &req); appErr != nil {
		fail(c, appErr)
		return
	}

	db := DB()
	tx := db.Begin()
	var stats struct {
		DatasourcesUpserted int `json:"datasources_upserted"`
		OntologiesUpserted  int `json:"ontologies_upserted"`
	}

	// Upsert datasources by code
	for _, raw := range req.Datasources {
		var payload struct {
			ID           uint   `json:"id"`
			Code         string `json:"code"`
			Name         string `json:"name"`
			Type         string `json:"type"`
			Host         string `json:"host"`
			Port         int    `json:"port"`
			DatabaseName string `json:"database_name"`
			Username     string `json:"username"`
			Status       string `json:"status"`
			Description  string `json:"description"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			continue
		}
		if payload.Code == "" {
			continue
		}

		var existing model.DataSource
		err := tx.Where("code = ?", payload.Code).First(&existing).Error
		if err == nil {
			// Update
			updates := map[string]interface{}{
				"name":          firstNonEmpty(payload.Name, existing.Name),
				"type":          firstNonEmpty(payload.Type, existing.Type),
				"host":          payload.Host,
				"port":          payload.Port,
				"database_name": payload.DatabaseName,
				"username":      payload.Username,
				"status":        firstNonEmpty(payload.Status, existing.Status),
				"description":   payload.Description,
				"updated_at":    time.Now(),
			}
			tx.Model(&model.DataSource{}).Where("id = ?", existing.ID).Updates(updates)
		} else {
			// Create
			ds := model.DataSource{
				Code:         payload.Code,
				Name:         firstNonEmpty(payload.Name, payload.Code),
				Type:         payload.Type,
				Host:         payload.Host,
				Port:         payload.Port,
				DatabaseName: payload.DatabaseName,
				Username:     payload.Username,
				Status:       firstNonEmpty(payload.Status, "inactive"),
				Description:  payload.Description,
			}
			tx.Create(&ds)
		}
		stats.DatasourcesUpserted++
	}

	// Upsert ontologies by code
	for _, raw := range req.Ontologies {
		var payload struct {
			ID          uint   `json:"id"`
			Code        string `json:"code"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Version     string `json:"version"`
			Status      string `json:"status"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			continue
		}
		if payload.Code == "" {
			continue
		}

		var existing model.OntDefinition
		err := tx.Where("code = ?", payload.Code).First(&existing).Error
		if err == nil {
			updates := map[string]interface{}{
				"name":        firstNonEmpty(payload.Name, existing.Name),
				"description": payload.Description,
				"version":     firstNonEmpty(payload.Version, existing.Version),
				"status":      firstNonEmpty(payload.Status, existing.Status),
				"updated_at":  time.Now(),
			}
			tx.Model(&model.OntDefinition{}).Where("id = ?", existing.ID).Updates(updates)
		} else {
			ont := model.OntDefinition{
				Code:        payload.Code,
				Name:        firstNonEmpty(payload.Name, payload.Code),
				Description: payload.Description,
				Version:     firstNonEmpty(payload.Version, "1.0.0"),
				Status:      firstNonEmpty(payload.Status, "draft"),
			}
			tx.Create(&ont)
		}
		stats.OntologiesUpserted++
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		logger.Errorf("[workspace] 保存失败: %v", err)
		fail(c, errs.DBError(err))
		return
	}

	okMsg(c, stats, "工作区已保存（best-effort）")
}
