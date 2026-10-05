package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用全局配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	CORS     CORSConfig     `mapstructure:"cors"`
	Log      LogConfig      `mapstructure:"log"`
	Pool     PoolConfig     `mapstructure:"pool"`
	Security SecurityConfig `mapstructure:"security"`
	AI       AIConfig       `mapstructure:"ai"`
}

// AIConfig AI / LLM 相关配置（Eino 意图链）。
// enabled 默认 false：关闭时仅走确定性规则解析，规则未命中即 matched=false，绝不调用 LLM。
type AIConfig struct {
	Enabled        bool                `mapstructure:"enabled"`
	Provider       string              `mapstructure:"provider"` // openai_compatible
	BaseURL        string              `mapstructure:"base_url"`
	APIKey         string              `mapstructure:"api_key"` // 支持 ${ENV} 展开
	Model          string              `mapstructure:"model"`
	Temperature    float32             `mapstructure:"temperature"`
	Seed           int                 `mapstructure:"seed"`
	ResponseFormat string              `mapstructure:"response_format"`
	MaxRepair      int                 `mapstructure:"max_repair"`
	Stream         bool                `mapstructure:"stream"`
	IntentCache    IntentCacheConfig   `mapstructure:"intent_cache"`
	Clarification  ClarificationConfig `mapstructure:"clarification"`
	ConceptRAG     ConceptRAGConfig    `mapstructure:"concept_rag"`
	MultiHop       MultiHopConfig      `mapstructure:"multihop"`
	SQLGen         SQLGenYAML          `mapstructure:"sql_gen"`
}

// SQLGenYAML SQL 生成管线（internal/ai/sqlgen）的配置，对应 ai.sql_gen 段。
//
// 与 sqlgen.SQLGenConfig 字段一一对应，位于 config 包以避免 config↔ai 的循环依赖；
// 上层可据此构造 sqlgen.SQLGenConfig。
type SQLGenYAML struct {
	MaxRepair      int `mapstructure:"max_repair"`      // 解析/校验失败重试次数
	MaxLimit       int `mapstructure:"max_limit"`       // 强制 LIMIT 上限
	DefaultLimit   int `mapstructure:"default_limit"`   // 无 LIMIT 时追加的默认值
	TimeoutSeconds int `mapstructure:"timeout_seconds"` // 单条 SQL 执行超时秒数
	MaxRows        int `mapstructure:"max_rows"`        // 结果行数截断上限
}

// MultiHopConfig P2 确定性本体遍历规划器的「受限 LLM 槽位」开关。
//
// 两者默认 false：入口实体锚定与关系路径剪枝均走纯 Go 词法/正则解析（无 LLM、无随机、
// 不依赖 intent_cache 是否开启），从根本上保证「同问同锚 → 同 selected_edges → 同 SQL → 同 hop_count」；
// 全链路仅保留「最终综合」一次 LLM 调用（自由文本，不参与 SQL 生成，故不破坏取数确定性）。
// 置 true 可回退到 LLM 槽位（调试/对比用），但 LLM 输出抖动会破坏上述不变量。
type MultiHopConfig struct {
	LLMEntryAnchor bool `mapstructure:"llm_entry_anchor"` // true=入口锚走 LLM（词法解析失败时始终 LLM 兜底）
	LLMPathSelect  bool `mapstructure:"llm_path_select"`  // true=关系路径剪枝走 LLM（PathSelect ForcedTool）
}

// IntentCacheConfig 意图计划缓存开关（本期默认关闭）。
type IntentCacheConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// ClarificationConfig 未命中时的问题澄清配置。
type ClarificationConfig struct {
	Enabled    bool `mapstructure:"enabled"`
	MaxOptions int  `mapstructure:"max_options"`
}

// ConceptRAGConfig 概念问答 RAG 摘要开关（默认关闭）。
type ConceptRAGConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// ResolvedAPIKey 返回展开环境变量后的 api_key（形如 ${ODIN_AI_API_KEY} 会被替换）。
func (a AIConfig) ResolvedAPIKey() string { return expandEnv(a.APIKey) }

// expandEnv 展开字符串中的 ${VAR} 与 $VAR 环境变量引用；未设置的变量替换为空串。
func expandEnv(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || !strings.Contains(s, "$") {
		return s
	}
	return os.Expand(s, func(key string) string {
		key = strings.TrimSpace(key)
		if key == "" {
			return ""
		}
		return os.Getenv(key)
	})
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

// Addr 返回监听地址，例如 0.0.0.0:8080
func (s ServerConfig) Addr() string {
	host := s.Host
	if host == "" {
		host = "0.0.0.0"
	}
	port := s.Port
	if port == 0 {
		port = 8080
	}
	return fmt.Sprintf("%s:%d", host, port)
}

// DatabaseConfig 元数据存储配置
type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"` // sqlite, mysql, postgres
	DSN          string `mapstructure:"dsn"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

// CORSConfig 跨域配置
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // json, console
}

// PoolConfig 数据源连接池默认配置
type PoolConfig struct {
	MaxOpen     int `mapstructure:"max_open"`
	MaxIdle     int `mapstructure:"max_idle"`
	MaxLifetime int `mapstructure:"max_lifetime"` // seconds
}

// SecurityConfig 安全相关配置
type SecurityConfig struct {
	EncryptKey string `mapstructure:"encrypt_key"`
}

// Load 从指定路径加载配置文件，并支持环境变量覆盖。
// 环境变量前缀为 ODIN_，使用 _ 作为层级分隔符，例如：
// ODIN_SERVER_PORT=9090 覆盖 server.port。
func Load(path string) (*Config, error) {
	v := viper.New()

	// 设置默认值
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.dsn", "./data/odin.db")
	v.SetDefault("database.max_open_conns", 10)
	v.SetDefault("log.level", "debug")
	v.SetDefault("log.format", "console")
	v.SetDefault("pool.max_open", 10)
	v.SetDefault("pool.max_idle", 5)
	v.SetDefault("pool.max_lifetime", 3600)
	v.SetDefault("security.encrypt_key", "odin-default-key")
	// AI 段默认值：默认关闭，确定性优先。
	v.SetDefault("ai.enabled", false)
	v.SetDefault("ai.provider", "openai_compatible")
	v.SetDefault("ai.base_url", "https://api.deepseek.com/v1")
	v.SetDefault("ai.model", "deepseek-chat")
	v.SetDefault("ai.temperature", 0)
	v.SetDefault("ai.seed", 42)
	v.SetDefault("ai.response_format", "json_schema_strict")
	v.SetDefault("ai.max_repair", 2)
	v.SetDefault("ai.stream", true)
	v.SetDefault("ai.intent_cache.enabled", false)
	v.SetDefault("ai.clarification.enabled", true)
	v.SetDefault("ai.clarification.max_options", 4)
	v.SetDefault("ai.concept_rag.enabled", false)
	// P2 多跳：默认全走确定性词法解析，LLM 槽位关闭（见 MultiHopConfig 注释）。
	v.SetDefault("ai.multihop.llm_entry_anchor", false)
	v.SetDefault("ai.multihop.llm_path_select", false)
	// SQL 生成管线（internal/ai/sqlgen）默认值。
	v.SetDefault("ai.sql_gen.max_repair", 2)
	v.SetDefault("ai.sql_gen.max_limit", 10000)
	v.SetDefault("ai.sql_gen.default_limit", 1000)
	v.SetDefault("ai.sql_gen.timeout_seconds", 30)
	v.SetDefault("ai.sql_gen.max_rows", 10000)

	if path != "" {
		// 拆分目录与文件名，兼容传入完整文件路径
		dir, file := splitPath(path)
		v.AddConfigPath(dir)
		v.SetConfigName(strings.TrimSuffix(file, ".yaml"))
		v.SetConfigType("yaml")
		if err := v.ReadInConfig(); err != nil {
			// 配置文件不存在时不视为致命错误，使用默认值 + 环境变量
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("read config: %w", err)
			}
		}
	}

	// 环境变量覆盖：ODIN_SERVER_PORT -> server.port
	v.SetEnvPrefix("ODIN")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}

// splitPath 将形如 "configs/config.yaml" 的路径拆分为目录与文件名。
func splitPath(path string) (dir, file string) {
	path = strings.ReplaceAll(path, "\\", "/")
	idx := strings.LastIndex(path, "/")
	if idx < 0 {
		return ".", path
	}
	dir = path[:idx]
	file = path[idx+1:]
	if dir == "" {
		dir = "."
	}
	if file == "" {
		file = "config.yaml"
	}
	return dir, file
}
