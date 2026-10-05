package model

import "time"

// IntentCache 意图计划缓存：把规范化后的自然语言（canonical_key）映射到确定的 QueryRequest。
//
// 本期默认关闭（ai.intent_cache.enabled=false），仅在开关打开时读写；
// 启用后命中即复用存储的 QueryRequest，从字节级保证「同问同 SQL」。
type IntentCache struct {
	// CanonicalKey = hash(normalized_nl + ontology_version + vocab_version)，作为主键。
	CanonicalKey     string    `gorm:"primaryKey;size:128" json:"canonical_key"`
	OntologyID       uint      `gorm:"index" json:"ontology_id"`
	NLText           string    `gorm:"type:text" json:"nl_text"`
	NormalizedNL     string    `gorm:"type:text" json:"normalized_nl"`
	ParsePath        string    `gorm:"size:20" json:"parse_path"` // rule / llm
	QueryRequestJSON string    `gorm:"type:text" json:"query_request_json"`
	LLMJSON          string    `gorm:"type:text" json:"llm_json"`
	OntologyVersion  string    `gorm:"size:20" json:"ontology_version"`
	VocabVersion     string    `gorm:"size:64" json:"vocab_version"`
	CreatedAt        time.Time `gorm:"index" json:"created_at"`
}

func (IntentCache) TableName() string { return "intent_cache" }
