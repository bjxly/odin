package model

import "time"

// AssistantSession 助手会话。一次会话包含多轮 AssistantMessage，可绑定一个默认本体。
type AssistantSession struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID     string    `gorm:"size:64;uniqueIndex" json:"session_id"` // 对外暴露的会话标识（UUID）
	Title         string    `gorm:"size:200" json:"title"`
	OntologyID    uint      `gorm:"index" json:"ontology_id"`
	MessageCount  int       `json:"message_count"`
	LastMessageAt time.Time `gorm:"index" json:"last_message_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (AssistantSession) TableName() string { return "assistant_session" }

// AssistantMessage 会话中的一条消息（用户或助手）。
//
// 助手消息以「渲染块」列表（BlocksJSON，见 设计文档 §8.6）承载结构化产物，
// 并挂 trace_id 以便一键溯源；无 trace 的数据答复视为异常。
type AssistantMessage struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID  string    `gorm:"size:64;index" json:"session_id"`
	Role       string    `gorm:"size:20" json:"role"` // user / assistant
	Content    string    `gorm:"type:text" json:"content"`
	BlocksJSON string    `gorm:"type:text" json:"blocks_json"`
	SkillID    string    `gorm:"size:60" json:"skill_id"`
	ParsePath  string    `gorm:"size:20" json:"parse_path"` // rule / llm / cache / concept / clarification
	TraceID    string    `gorm:"size:64;index" json:"trace_id"`
	MetaJSON   string    `gorm:"type:text" json:"meta_json"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

// TableName 指定表名。
func (AssistantMessage) TableName() string { return "assistant_message" }
