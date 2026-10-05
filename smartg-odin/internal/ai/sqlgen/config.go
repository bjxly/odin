// Package sqlgen 实现「LLM 直接生成 SQL」的确定性受控管线：
//
//	BuildContext(装配本体/映射/物理 Schema 上下文)
//	  → BuildSystemPrompt/BuildUserPrompt(构造提示词)
//	  → Generate(调用 LLM 产出结构化 SQLGenResult)
//	  → Validate(轻量安全校验 + 强制 LIMIT)
//	  → Execute(经 ds/connector 在各数据源执行)
//	  → Merge(按策略合并多步结果 + 图表建议)
//
// 设计约束：LLM 仅产出 SELECT 语句，绝不接触写操作；所有表名/列名必须来自
// 装配的物理 Schema；跨数据源查询被拆分为多条 SQL 分别执行后再内存合并。
package sqlgen

// SQLGenConfig SQL 生成管线配置。
//
// 各字段均带 yaml/json 标签，可直接从配置文件或 API 响应序列化；
// 通过 DefaultConfig() 获取带默认值的实例，再用 Normalize() 修正非法值。
type SQLGenConfig struct {
	MaxRepair      int `yaml:"max_repair" json:"max_repair"`           // 校验/解析失败时的重试次数，默认 2
	MaxLimit       int `yaml:"max_limit" json:"max_limit"`             // 强制 LIMIT 上限，默认 10000
	DefaultLimit   int `yaml:"default_limit" json:"default_limit"`     // SQL 无 LIMIT 时追加的默认值，默认 1000
	TimeoutSeconds int `yaml:"timeout_seconds" json:"timeout_seconds"` // 单条 SQL 执行超时秒数，默认 30
	MaxRows        int `yaml:"max_rows" json:"max_rows"`               // 结果行数截断上限，默认 10000
}

// 默认配置常量。
const (
	defaultMaxRepair      = 2
	defaultMaxLimit       = 10000
	defaultDefaultLimit   = 1000
	defaultTimeoutSeconds = 30
	defaultMaxRows        = 10000
)

// DefaultConfig 返回带推荐默认值的 SQLGenConfig。
func DefaultConfig() *SQLGenConfig {
	return &SQLGenConfig{
		MaxRepair:      defaultMaxRepair,
		MaxLimit:       defaultMaxLimit,
		DefaultLimit:   defaultDefaultLimit,
		TimeoutSeconds: defaultTimeoutSeconds,
		MaxRows:        defaultMaxRows,
	}
}

// Normalize 就地修正非法（<=0）字段为默认值，并保证 DefaultLimit 不超过 MaxLimit。
// 返回自身以便链式调用。
func (c *SQLGenConfig) Normalize() *SQLGenConfig {
	if c == nil {
		return DefaultConfig()
	}
	if c.MaxRepair < 0 {
		c.MaxRepair = 0
	}
	if c.MaxLimit <= 0 {
		c.MaxLimit = defaultMaxLimit
	}
	if c.DefaultLimit <= 0 {
		c.DefaultLimit = defaultDefaultLimit
	}
	if c.TimeoutSeconds <= 0 {
		c.TimeoutSeconds = defaultTimeoutSeconds
	}
	if c.MaxRows <= 0 {
		c.MaxRows = defaultMaxRows
	}
	if c.DefaultLimit > c.MaxLimit {
		c.DefaultLimit = c.MaxLimit
	}
	return c
}
