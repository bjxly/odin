package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"smartg-odin/internal/ai/nlparse"
	"smartg-odin/internal/common/config"
	"smartg-odin/internal/common/logger"
	"smartg-odin/internal/ds/query"
	"smartg-odin/internal/model"
)

// aiCfg 包级 AI 配置，由 main 启动时通过 Configure 注入；零值即「AI 关闭」，保证默认确定性。
var aiCfg config.AIConfig

// Configure 注入 AI 配置（进程启动时调用一次）。
func Configure(cfg config.AIConfig) { aiCfg = cfg }

// Config 返回当前 AI 配置副本。
func Config() config.AIConfig { return aiCfg }

// Enabled 报告 LLM 回退是否开启。
func Enabled() bool { return aiCfg.Enabled }

// IntentCacheEnabled 报告意图缓存是否开启。
func IntentCacheEnabled() bool { return aiCfg.IntentCache.Enabled }

// IntentResult 意图解析结果（规则 / LLM / 缓存 三条路径统一结构）。
type IntentResult struct {
	Query        *query.QueryRequest `json:"query"`
	Notes        []string            `json:"notes"`
	Hits         []string            `json:"hits"`
	ParsePath    string              `json:"parse_path"` // rule / llm / cache
	Matched      bool                `json:"matched"`
	LLMMeta      *LLMMeta            `json:"llm_meta,omitempty"`
	CanonicalKey string              `json:"canonical_key,omitempty"`
}

// ParseIntent 先规则、后（可选）LLM 的确定性意图解析入口。
//
// 流程（见 设计文档 §2.3 三重防幻觉、§3 确定性保证）：
//  1. intent_cache 开启且命中 → parse_path=cache；
//  2. P1 规则解析命中 → parse_path=rule（零幻觉零成本快路径）；
//  3. 规则未命中且 ai.enabled → Eino 意图链 → parse_path=llm（带 LLMMeta）；
//  4. 规则未命中且 ai 关闭 / 无 key → matched=false，绝不调用 LLM。
func ParseIntent(ctx context.Context, db *gorm.DB, ontologyID uint, nlText string) (*IntentResult, error) {
	nlText = strings.TrimSpace(nlText)

	vocab, err := loadVocabulary(db, ontologyID)
	if err != nil {
		return nil, err
	}
	normalized := normalizeNL(nlText)
	key := canonicalKey(normalized, vocab.OntologyVersion, vocab.Version())

	// 1) 缓存命中（仅当开关开启）。
	if aiCfg.IntentCache.Enabled {
		if entry, hit := lookupIntentCache(db, key); hit {
			var req query.QueryRequest
			uerr := json.Unmarshal([]byte(entry.QueryRequestJSON), &req)
			if uerr == nil {
				return &IntentResult{
					Query:        &req,
					Notes:        []string{"命中意图缓存（parse_path=cache）"},
					Hits:         []string{req.ClassName},
					ParsePath:    "cache",
					Matched:      true,
					CanonicalKey: key,
				}, nil
			}
			logger.Warnf("[ai] 意图缓存反序列化失败，忽略缓存: %v", uerr)
		}
	}

	// 2) 规则解析（P1 快路径）。
	ruleRes := nlparse.Parse(db, ontologyID, nlText)
	if ruleRes.Matched {
		// P0 降级（治「LLM 无法补救漏过滤」）：规则命中主类但未抽到任何过滤，且原文含
		// 实体标识符（如 ORD-20250115-007）时，规则结果极可能漏掉等值过滤（退化为全表查询）。
		// 此时不立即返回，改走 LLM 意图链尝试补救；仅在「filters 空 + 含标识符」条件下触发，保证确定性。
		needRemedy := len(ruleRes.Query.Filters) == 0 && nlparse.ContainsEntityIdentifier(nlText)
		if !needRemedy {
			cacheIntent(db, key, vocab, ontologyID, nlText, normalized, "rule", ruleRes.Query, nil)
			return &IntentResult{
				Query:        ruleRes.Query,
				Notes:        ruleRes.Notes,
				Hits:         ruleRes.Hits,
				ParsePath:    "rule",
				Matched:      true,
				CanonicalKey: key,
			}, nil
		}
		// 尝试 LLM 补救（仅当 AI 可用）；温度/seed 已固定（temperature=0+seed），补救结果稳定。
		if aiCfg.Enabled && strings.TrimSpace(aiCfg.ResolvedAPIKey()) != "" {
			chainRes, cerr := runIntentChain(ctx, aiCfg, vocab, ontologyID, nlText)
			if cerr == nil && chainRes != nil && chainRes.OK && chainRes.Request != nil {
				req := chainRes.Request
				req.OntologyID = ontologyID
				notes := []string{"规则命中主类但缺过滤且含实体标识符，已由 LLM 意图链补救（parse_path=llm）"}
				if chainRes.Meta != nil && chainRes.Meta.Repairs > 0 {
					notes = append(notes, fmt.Sprintf("IntentValidator 有界修复 %d 次", chainRes.Meta.Repairs))
				}
				cacheIntent(db, key, vocab, ontologyID, nlText, normalized, "llm", req, chainRes.Meta)
				return &IntentResult{
					Query:        req,
					Notes:        notes,
					Hits:         []string{req.ClassName},
					ParsePath:    "llm",
					Matched:      true,
					LLMMeta:      chainRes.Meta,
					CanonicalKey: key,
				}, nil
			}
			if cerr != nil {
				logger.Warnf("[ai] 缺过滤补救的意图链失败，回退规则结果: %v", cerr)
			}
		}
		// LLM 不可用/失败/未命中 → 回退使用规则结果（保持现状行为）。
		cacheIntent(db, key, vocab, ontologyID, nlText, normalized, "rule", ruleRes.Query, nil)
		return &IntentResult{
			Query:        ruleRes.Query,
			Notes:        appendNotes(ruleRes.Notes, "规则命中主类但无过滤且含实体标识符，LLM 补救不可用，回退规则结果"),
			Hits:         ruleRes.Hits,
			ParsePath:    "rule",
			Matched:      true,
			CanonicalKey: key,
		}, nil
	}

	// 3) 规则未命中：AI 关闭 → matched=false（不调用 LLM）。
	if !aiCfg.Enabled {
		return &IntentResult{
			Query:        ruleRes.Query,
			Notes:        appendNotes(ruleRes.Notes, "规则未命中且 AI 未启用（ai.enabled=false），未调用 LLM"),
			Hits:         ruleRes.Hits,
			ParsePath:    "rule",
			Matched:      false,
			CanonicalKey: key,
		}, nil
	}

	// AI 开启但缺少 api_key → 明确告警并返回未命中，不发起网络调用。
	if strings.TrimSpace(aiCfg.ResolvedAPIKey()) == "" {
		return &IntentResult{
			Query:        ruleRes.Query,
			Notes:        appendNotes(ruleRes.Notes, "AI 已启用但未配置 api_key，无法调用 LLM"),
			Hits:         ruleRes.Hits,
			ParsePath:    "rule",
			Matched:      false,
			CanonicalKey: key,
		}, nil
	}

	// 4) AI 开启 → Eino 意图链。
	chainRes, cerr := runIntentChain(ctx, aiCfg, vocab, ontologyID, nlText)
	if cerr != nil {
		logger.Warnf("[ai] 意图链失败，回退为未命中: %v", cerr)
		return &IntentResult{
			Query:        ruleRes.Query,
			Notes:        appendNotes(ruleRes.Notes, "LLM 意图链调用失败："+cerr.Error()),
			Hits:         ruleRes.Hits,
			ParsePath:    "rule",
			Matched:      false,
			CanonicalKey: key,
		}, nil
	}
	if !chainRes.OK || chainRes.Request == nil {
		return &IntentResult{
			Query:        chainRes.Request,
			Notes:        appendNotes(ruleRes.Notes, "LLM 意图未通过词表校验："+strings.Join(chainRes.Issues, "；")),
			Hits:         ruleRes.Hits,
			ParsePath:    "llm",
			Matched:      false,
			LLMMeta:      chainRes.Meta,
			CanonicalKey: key,
		}, nil
	}

	req := chainRes.Request
	req.OntologyID = ontologyID
	notes := []string{"规则未命中，已由 LLM 意图链解析（parse_path=llm）"}
	if chainRes.Meta != nil && chainRes.Meta.Repairs > 0 {
		notes = append(notes, fmt.Sprintf("IntentValidator 有界修复 %d 次", chainRes.Meta.Repairs))
	}
	cacheIntent(db, key, vocab, ontologyID, nlText, normalized, "llm", req, chainRes.Meta)
	return &IntentResult{
		Query:        req,
		Notes:        notes,
		Hits:         []string{req.ClassName},
		ParsePath:    "llm",
		Matched:      true,
		LLMMeta:      chainRes.Meta,
		CanonicalKey: key,
	}, nil
}

// cacheIntent 在缓存开关开启时写入意图缓存；关闭时不做任何写操作。
func cacheIntent(db *gorm.DB, key string, vocab *Vocabulary, ontologyID uint,
	nlText, normalized, path string, req *query.QueryRequest, meta *LLMMeta) {
	if !aiCfg.IntentCache.Enabled || req == nil || key == "" || db == nil {
		return
	}
	entry := &model.IntentCache{
		CanonicalKey:     key,
		OntologyID:       ontologyID,
		NLText:           nlText,
		NormalizedNL:     normalized,
		ParsePath:        path,
		QueryRequestJSON: marshalJSONString(req),
		OntologyVersion:  vocab.OntologyVersion,
		VocabVersion:     vocab.Version(),
		CreatedAt:        time.Now(),
	}
	if meta != nil {
		entry.LLMJSON = marshalJSONString(meta)
	}
	storeIntentCache(db, entry)
}

// appendNotes 在已有 notes 后追加一条，返回新切片（不修改入参）。
func appendNotes(notes []string, extra string) []string {
	out := make([]string, 0, len(notes)+1)
	out = append(out, notes...)
	return append(out, extra)
}

// marshalJSONString 序列化为 JSON 字符串，失败返回空串。
func marshalJSONString(v interface{}) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
