package model

import "time"

// OntMappingConfig 本体类到物理数据源表的映射配置。
type OntMappingConfig struct {
	ID                   uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	OntologyID           uint      `gorm:"index;not null" json:"ontology_id"`
	ClassID              uint      `gorm:"index;not null" json:"class_id"`
	DataSourceID         uint      `gorm:"column:datasource_id;index;not null" json:"datasource_id"`
	SourceTable          string    `gorm:"size:100;not null" json:"source_table"`
	SourceSchema         string    `gorm:"size:100" json:"source_schema"`
	PropertyMappingsJSON string    `gorm:"type:text" json:"property_mappings_json"`
	Status               string    `gorm:"size:20;default:draft" json:"status"` // draft, validated, active
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (OntMappingConfig) TableName() string { return "mapping_config" }
