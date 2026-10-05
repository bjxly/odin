package model

import "time"

// OntDefinition 本体定义元数据。
type OntDefinition struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Code        string    `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Version     string    `gorm:"size:20;default:1.0.0" json:"version"`
	Status      string    `gorm:"size:20;default:draft" json:"status"` // draft, published, archived
	LayoutJSON  string    `gorm:"type:text" json:"layout_json"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (OntDefinition) TableName() string { return "ont_definition" }

// OntClass 本体类定义。
type OntClass struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	OntologyID    uint      `gorm:"index;not null" json:"ontology_id"`
	Name          string    `gorm:"size:100;not null" json:"name"`
	Label         string    `gorm:"size:100" json:"label"`
	Description   string    `gorm:"type:text" json:"description"`
	ClassType     string    `gorm:"size:20;default:normal" json:"class_type"` // normal, virtual
	ParentClassID *uint     `gorm:"index" json:"parent_class_id"`
	Synonyms      string    `gorm:"type:text" json:"synonyms"` // JSON 数组或逗号分隔的同义词，供 NL 解析与 vocabulary 使用
	Icon          string    `gorm:"size:50" json:"icon"`
	Color         string    `gorm:"size:20" json:"color"`
	PositionX     float64   `json:"position_x"`
	PositionY     float64   `json:"position_y"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (OntClass) TableName() string { return "ont_class" }

// OntClassProperty 本体类的属性定义。
type OntClassProperty struct {
	ID              uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	ClassID         uint   `gorm:"index;not null" json:"class_id"`
	OntologyID      uint   `gorm:"index;not null" json:"ontology_id"`
	Name            string `gorm:"size:100;not null" json:"name"`
	Label           string `gorm:"size:100" json:"label"`
	DataType        string `gorm:"size:30;not null" json:"data_type"` // string, integer, float, boolean, datetime, json
	Required        bool   `gorm:"default:false" json:"required"`
	Description     string `gorm:"type:text" json:"description"`
	ConstraintsJSON string `gorm:"type:text" json:"constraints_json"`
	Synonyms        string `gorm:"type:text" json:"synonyms"` // JSON 数组或逗号分隔的同义词，供 NL 解析与 vocabulary 使用
	SortOrder       int    `gorm:"default:0" json:"sort_order"`
}

func (OntClassProperty) TableName() string { return "ont_class_property" }

// OntRelation 本体类之间的关系定义。
type OntRelation struct {
	ID                uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	OntologyID        uint   `gorm:"index;not null" json:"ontology_id"`
	Name              string `gorm:"size:100;not null" json:"name"`
	Label             string `gorm:"size:100" json:"label"`
	FromClassID       uint   `gorm:"index;not null" json:"from_class_id"`
	ToClassID         uint   `gorm:"index;not null" json:"to_class_id"`
	RelationType      string `gorm:"size:30;default:association" json:"relation_type"` // association, composition, inheritance
	Cardinality       string `gorm:"size:10;default:1:N" json:"cardinality"`           // 1:1, 1:N, N:M
	Description       string `gorm:"type:text" json:"description"`
	JoinConditionJSON string `gorm:"type:text" json:"join_condition_json"`
}

func (OntRelation) TableName() string { return "ont_relation" }

// OntRule 本体推理规则。
type OntRule struct {
	ID            uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	OntologyID    uint   `gorm:"index;not null" json:"ontology_id"`
	Name          string `gorm:"size:100;not null" json:"name"`
	Description   string `gorm:"type:text" json:"description"`
	RuleType      string `gorm:"size:30;not null" json:"rule_type"` // derivation, constraint, inference
	ConditionJSON string `gorm:"type:text;not null" json:"condition_json"`
	ActionJSON    string `gorm:"type:text;not null" json:"action_json"`
	Priority      int    `gorm:"default:0" json:"priority"`
	Enabled       bool   `gorm:"default:true" json:"enabled"`
}

func (OntRule) TableName() string { return "ont_rule" }

// OntAction 规则对应的动作定义。
type OntAction struct {
	ID              uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	RuleID          uint   `gorm:"index;not null" json:"rule_id"`
	ActionType      string `gorm:"size:30;not null" json:"action_type"` // add_filter, add_property, expand_class, set_value
	TargetClassID   *uint  `json:"target_class_id"`
	TargetProperty  string `gorm:"size:100" json:"target_property"`
	ValueExpression string `gorm:"type:text" json:"value_expression"`
	ParamsJSON      string `gorm:"type:text" json:"params_json"`
}

func (OntAction) TableName() string { return "ont_action" }
