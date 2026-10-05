package ai

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode"

	"gorm.io/gorm"

	"smartg-odin/internal/common/logger"
	"smartg-odin/internal/model"
)

// normalizeNL 对自然语言做确定性规范化：全角→半角、大小写归一、空白折叠、首尾及尾随标点清理。
// 规范化结果参与 canonical_key 计算，保证「同一问题的不同书写」落到同一缓存键。
func normalizeNL(text string) string {
	var b strings.Builder
	prevSpace := false
	for _, r := range text {
		r = foldWidth(r)
		switch {
		case unicode.IsSpace(r):
			prevSpace = true
		default:
			if prevSpace && b.Len() > 0 {
				b.WriteRune(' ')
			}
			prevSpace = false
			b.WriteRune(unicode.ToLower(r))
		}
	}
	s := strings.TrimSpace(b.String())
	// 去除尾随的句读标点，避免「查询VIP客户。」与「查询VIP客户」被判为不同问题。
	s = strings.TrimRight(s, "。，、；：！？.,;:!？? \t\n")
	return s
}

// FoldWidthRune 是 foldWidth 的导出包装：将全角 ASCII 与全角空格折叠为半角。
// 供 planner 的确定性词法选边/入口锚复用，保证两处归一口径完全一致（不会因全角“库存”而漏匹配）。
func FoldWidthRune(r rune) rune { return foldWidth(r) }

// foldWidth 将全角 ASCII 与全角空格折叠为半角。
func foldWidth(r rune) rune {
	if r == 0x3000 { // 全角空格
		return ' '
	}
	if r >= 0xFF01 && r <= 0xFF5E { // 全角 ! ~ ~
		return r - 0xFEE0
	}
	return r
}

// canonicalKey 计算意图缓存键：hash(normalized_nl + ontology_version + vocab_version)。
func canonicalKey(normalizedNL, ontologyVersion, vocabVersion string) string {
	raw := normalizedNL + "\x00" + ontologyVersion + "\x00" + vocabVersion
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// lookupIntentCache 按 canonical_key 查缓存；仅在 ai.intent_cache.enabled=true 时调用。
func lookupIntentCache(db *gorm.DB, key string) (*model.IntentCache, bool) {
	if db == nil || key == "" {
		return nil, false
	}
	var entry model.IntentCache
	if err := db.Where("canonical_key = ?", key).First(&entry).Error; err != nil {
		return nil, false
	}
	return &entry, true
}

// storeIntentCache 写入/更新意图缓存；仅在 ai.intent_cache.enabled=true 时调用。失败仅告警。
func storeIntentCache(db *gorm.DB, entry *model.IntentCache) {
	if db == nil || entry == nil || entry.CanonicalKey == "" {
		return
	}
	// canonical_key 为主键，使用 Save 实现 upsert 语义。
	if err := db.Save(entry).Error; err != nil {
		logger.Warnf("[ai] 写入意图缓存失败: %v", err)
	}
}
