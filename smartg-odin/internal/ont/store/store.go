// Package store 提供本体数据的访问层，封装常用的 GORM 查询，
// 供翻译引擎、推理引擎与上层 handler 复用。
package store

import (
	"errors"

	"gorm.io/gorm"

	"smartg-odin/internal/model"
)

// OntStore 本体数据访问器。
type OntStore struct {
	db *gorm.DB
}

// New 创建一个 OntStore。
func New(db *gorm.DB) *OntStore {
	return &OntStore{db: db}
}

// DB 暴露底层 gorm.DB，便于特殊场景直接查询。
func (s *OntStore) DB() *gorm.DB { return s.db }

// ---- 本体定义 ----

// GetOntologyByCode 按 code 查询本体定义。
func (s *OntStore) GetOntologyByCode(code string) (*model.OntDefinition, error) {
	var o model.OntDefinition
	if err := s.db.Where("code = ?", code).First(&o).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

// GetOntologyByID 按 ID 查询本体定义。
func (s *OntStore) GetOntologyByID(id uint) (*model.OntDefinition, error) {
	var o model.OntDefinition
	if err := s.db.First(&o, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

// ---- 类相关 ----

// GetClassByName 按本体 ID 与类名查询类定义。
func (s *OntStore) GetClassByName(ontologyID uint, name string) (*model.OntClass, error) {
	var c model.OntClass
	if err := s.db.Where("ontology_id = ? AND name = ?", ontologyID, name).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// GetClassByID 按主键查询类定义。
func (s *OntStore) GetClassByID(id uint) (*model.OntClass, error) {
	var c model.OntClass
	if err := s.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// ListClasses 列出本体下所有类。
func (s *OntStore) ListClasses(ontologyID uint) ([]model.OntClass, error) {
	var list []model.OntClass
	err := s.db.Where("ontology_id = ?", ontologyID).Order("id asc").Find(&list).Error
	return list, err
}

// GetSubClasses 查询某类的直接子类。
func (s *OntStore) GetSubClasses(parentID uint) ([]model.OntClass, error) {
	var list []model.OntClass
	err := s.db.Where("parent_class_id = ?", parentID).Order("id asc").Find(&list).Error
	return list, err
}

// GetClassProperties 查询某类的所有属性（按 sort_order 排序）。
func (s *OntStore) GetClassProperties(classID uint) ([]model.OntClassProperty, error) {
	var list []model.OntClassProperty
	err := s.db.Where("class_id = ?", classID).Order("sort_order asc, id asc").Find(&list).Error
	return list, err
}

// ---- 映射相关 ----

// GetMappingByClassID 查询某类的第一条映射配置。
func (s *OntStore) GetMappingByClassID(classID uint) (*model.OntMappingConfig, error) {
	var m model.OntMappingConfig
	if err := s.db.Where("class_id = ?", classID).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// GetMappingByClassAndSource 查询某类在指定数据源上的映射配置。
func (s *OntStore) GetMappingByClassAndSource(classID, datasourceID uint) (*model.OntMappingConfig, error) {
	var m model.OntMappingConfig
	if err := s.db.Where("class_id = ? AND datasource_id = ?", classID, datasourceID).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// ListMappingsByOntology 列出本体下所有映射配置。
func (s *OntStore) ListMappingsByOntology(ontologyID uint) ([]model.OntMappingConfig, error) {
	var list []model.OntMappingConfig
	err := s.db.Where("ontology_id = ?", ontologyID).Order("id asc").Find(&list).Error
	return list, err
}

// ---- 关系相关 ----

// GetRelationByName 按本体 ID 与关系名查询关系定义。
func (s *OntStore) GetRelationByName(ontologyID uint, name string) (*model.OntRelation, error) {
	var r model.OntRelation
	if err := s.db.Where("ontology_id = ? AND name = ?", ontologyID, name).First(&r).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

// GetRelationsByClass 查询与某类相关（作为起点或终点）的所有关系。
func (s *OntStore) GetRelationsByClass(ontologyID uint, classID uint) ([]model.OntRelation, error) {
	var list []model.OntRelation
	err := s.db.Where("ontology_id = ? AND (from_class_id = ? OR to_class_id = ?)", ontologyID, classID, classID).
		Order("id asc").Find(&list).Error
	return list, err
}

// ---- 规则相关 ----

// GetEnabledRules 查询本体下所有启用的规则（按 priority 升序）。
func (s *OntStore) GetEnabledRules(ontologyID uint) ([]model.OntRule, error) {
	var list []model.OntRule
	err := s.db.Where("ontology_id = ? AND enabled = ?", ontologyID, true).
		Order("priority asc, id asc").Find(&list).Error
	return list, err
}

// GetRuleActions 查询某条规则关联的动作定义。
func (s *OntStore) GetRuleActions(ruleID uint) ([]model.OntAction, error) {
	var list []model.OntAction
	err := s.db.Where("rule_id = ?", ruleID).Order("id asc").Find(&list).Error
	return list, err
}

// ---- Schema 相关 ----

// GetSchemaTables 查询某数据源下的所有表元信息。
func (s *OntStore) GetSchemaTables(datasourceID uint) ([]model.SchemaTable, error) {
	var list []model.SchemaTable
	err := s.db.Where("datasource_id = ?", datasourceID).Order("table_name asc").Find(&list).Error
	return list, err
}

// GetSchemaColumns 查询某数据源某张表的字段元信息（按序）。
func (s *OntStore) GetSchemaColumns(datasourceID uint, tableName string) ([]model.SchemaColumn, error) {
	var table model.SchemaTable
	if err := s.db.Where("datasource_id = ? AND table_name = ?", datasourceID, tableName).First(&table).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var list []model.SchemaColumn
	err := s.db.Where("table_id = ?", table.ID).Order("ordinal_pos asc, id asc").Find(&list).Error
	return list, err
}
