package model

import "time"

// AuditLog 操作审计日志。
type AuditLog struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       string    `gorm:"size:100" json:"user_id"`
	Action       string    `gorm:"size:50;not null" json:"action"` // create, update, delete, query, test
	ResourceType string    `gorm:"size:50" json:"resource_type"`   // datasource, ontology, mapping, query
	ResourceID   string    `gorm:"size:50" json:"resource_id"`
	DetailJSON   string    `gorm:"type:text" json:"detail_json"`
	IPAddress    string    `gorm:"size:50" json:"ip_address"`
	CreatedAt    time.Time `gorm:"index" json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_log" }
