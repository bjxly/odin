// Package ai 实现 ODIN 的确定性意图解析层（P2：Eino 意图链）。
//
// 设计原则（见 设计文档/推理和AI助手设计.md §1~§3）：
//   - LLM 绝不写 SQL，只产出受词表约束的本体意图（query.QueryRequest AST）；
//   - 先走 P1 的确定性规则解析（nlparse），命中即 parse_path=rule（零幻觉零成本）；
//   - 规则未命中且 ai.enabled=true 时，才走 Eino 意图链（ChatModel + ForcedTool
//     OntologyIntent + IntentValidator 有界修复），parse_path=llm；
//   - ai 关闭且规则未命中 → matched=false，绝不调用 LLM。
package ai

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"

	"gorm.io/gorm"

	"smartg-odin/internal/model"
)

// 意图链的固定版本号，纳入 canonical_key 与 llm_meta，保证可复现审计。
// p3-agg-v1：引入聚合（aggregates/group_by/having）词表与工具 schema。
const promptVersion = "p3-agg-v1"

// supportedOperators 是 translator.filterExpr 的实现集，作为 LLM 意图的操作符词表与校验依据。
// 必须与 internal/ds/query/translator.go filterExpr、internal/handler/query.go supportedFilterOps 保持一致。
var supportedOperators = []string{
	"eq", "neq", "ne", "gt", "gte", "ge", "lt", "lte", "le",
	"like", "in", "not_in", "nin", "is_null", "is_not_null", "not_null",
}

// supportedAggregateFuncs 是聚合函数的实现集（单一口径），必须与
// translator.supportedAggFuncs、handler.supportedAggFuncs、intent_tool 的 Enum 一致。
var supportedAggregateFuncs = []string{"count", "sum", "avg", "min", "max"}

// limit 取值范围（与 handler.maxQueryLimit 对齐）。
const (
	limitMin = 1
	limitMax = 10000
)

// VocabProperty 词表中的属性视图。
type VocabProperty struct {
	Name     string   `json:"name"`
	Label    string   `json:"label"`
	DataType string   `json:"data_type"`
	Synonyms []string `json:"synonyms"`
	Required bool     `json:"required"`
}

// VocabClass 词表中的类视图。
type VocabClass struct {
	ID         uint            `json:"id"`
	Name       string          `json:"name"`
	Label      string          `json:"label"`
	ClassType  string          `json:"class_type"`
	ParentID   *uint           `json:"parent_id"`
	Synonyms   []string        `json:"synonyms"`
	Properties []VocabProperty `json:"properties"`
}

// VocabRelation 词表中的关系视图（端点以类名表达，便于 LLM 理解与校验）。
type VocabRelation struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	From  string `json:"from_class"`
	To    string `json:"to_class"`
}

// Vocabulary 一个本体的完整约束词表，既用于 prompt 注入，也用于意图校验与版本哈希。
type Vocabulary struct {
	OntologyID      uint            `json:"ontology_id"`
	OntologyVersion string          `json:"ontology_version"`
	Classes         []VocabClass    `json:"classes"`
	Relations       []VocabRelation `json:"relations"`
	Operators       []string        `json:"operators"`
	LimitMin        int             `json:"limit_min"`
	LimitMax        int             `json:"limit_max"`
	OrderDirs       []string        `json:"order_dirs"`
	AggFunctions    []string        `json:"agg_functions"`

	// 运行时索引（不序列化）。
	classesByName   map[string]*VocabClass
	relationsByName map[string]*VocabRelation
	version         string
}

// Version 返回词表内容哈希，作为 vocab_version 纳入 canonical_key 与 llm_meta。
func (v *Vocabulary) Version() string { return v.version }

// Class 按名称（忽略大小写）查找类。
func (v *Vocabulary) Class(name string) *VocabClass {
	if v == nil || v.classesByName == nil {
		return nil
	}
	return v.classesByName[strings.ToLower(strings.TrimSpace(name))]
}

// Relation 按名称（忽略大小写）查找关系。
func (v *Vocabulary) Relation(name string) *VocabRelation {
	if v == nil || v.relationsByName == nil {
		return nil
	}
	return v.relationsByName[strings.ToLower(strings.TrimSpace(name))]
}

// HasProperty 判断类是否拥有指定属性（按 name/label/同义词，忽略大小写）。
func (c *VocabClass) HasProperty(name string) bool {
	if c == nil {
		return false
	}
	target := strings.ToLower(strings.TrimSpace(name))
	if target == "" {
		return false
	}
	for _, p := range c.Properties {
		if strings.EqualFold(p.Name, target) || strings.EqualFold(p.Label, target) {
			return true
		}
		for _, s := range p.Synonyms {
			if strings.EqualFold(s, target) {
				return true
			}
		}
	}
	return false
}

// LoadVocabulary 从元数据库加载指定本体的约束词表（对外暴露，供技能 list_vocabulary 复用）。
func LoadVocabulary(db *gorm.DB, ontologyID uint) (*Vocabulary, error) {
	return loadVocabulary(db, ontologyID)
}

// loadVocabulary 从元数据库加载指定本体的约束词表。
func loadVocabulary(db *gorm.DB, ontologyID uint) (*Vocabulary, error) {
	v := &Vocabulary{
		OntologyID:   ontologyID,
		Classes:      make([]VocabClass, 0),
		Relations:    make([]VocabRelation, 0),
		Operators:    append([]string{}, supportedOperators...),
		LimitMin:     limitMin,
		LimitMax:     limitMax,
		OrderDirs:    []string{"asc", "desc"},
		AggFunctions: append([]string{}, supportedAggregateFuncs...),
	}

	var def model.OntDefinition
	if err := db.Select("id", "version").First(&def, ontologyID).Error; err == nil {
		v.OntologyVersion = def.Version
	}
	if strings.TrimSpace(v.OntologyVersion) == "" {
		v.OntologyVersion = "1.0.0"
	}

	// 类。
	var classes []model.OntClass
	db.Where("ontology_id = ?", ontologyID).Order("id asc").Find(&classes)
	for _, c := range classes {
		v.Classes = append(v.Classes, VocabClass{
			ID:         c.ID,
			Name:       c.Name,
			Label:      c.Label,
			ClassType:  c.ClassType,
			ParentID:   c.ParentClassID,
			Synonyms:   parseSynonyms(c.Synonyms),
			Properties: make([]VocabProperty, 0),
		})
	}

	// 属性。
	var props []model.OntClassProperty
	db.Where("ontology_id = ?", ontologyID).Order("class_id asc, sort_order asc, id asc").Find(&props)
	idx := make(map[uint]int, len(v.Classes))
	for i := range v.Classes {
		idx[v.Classes[i].ID] = i
	}
	for _, p := range props {
		i, okk := idx[p.ClassID]
		if !okk {
			continue
		}
		v.Classes[i].Properties = append(v.Classes[i].Properties, VocabProperty{
			Name:     p.Name,
			Label:    p.Label,
			DataType: p.DataType,
			Synonyms: parseSynonyms(p.Synonyms),
			Required: p.Required,
		})
	}

	// 关系（端点转类名）。
	nameByID := make(map[uint]string, len(v.Classes))
	for i := range v.Classes {
		nameByID[v.Classes[i].ID] = v.Classes[i].Name
	}
	var rels []model.OntRelation
	db.Where("ontology_id = ?", ontologyID).Order("id asc").Find(&rels)
	for _, r := range rels {
		v.Relations = append(v.Relations, VocabRelation{
			Name:  r.Name,
			Label: r.Label,
			From:  nameByID[r.FromClassID],
			To:    nameByID[r.ToClassID],
		})
	}

	// 构建索引与版本哈希。
	v.classesByName = make(map[string]*VocabClass, len(v.Classes))
	for i := range v.Classes {
		v.classesByName[strings.ToLower(v.Classes[i].Name)] = &v.Classes[i]
	}
	v.relationsByName = make(map[string]*VocabRelation, len(v.Relations))
	for i := range v.Relations {
		v.relationsByName[strings.ToLower(v.Relations[i].Name)] = &v.Relations[i]
	}
	v.version = computeVocabVersion(v)
	return v, nil
}

// computeVocabVersion 对词表结构（类/属性/关系/操作符）做稳定哈希，忽略运行时索引。
func computeVocabVersion(v *Vocabulary) string {
	type propSig struct {
		Name     string   `json:"n"`
		DataType string   `json:"t"`
		Synonyms []string `json:"s"`
		Required bool     `json:"r"`
	}
	type classSig struct {
		Name       string    `json:"n"`
		ClassType  string    `json:"c"`
		Properties []propSig `json:"p"`
	}
	type relSig struct {
		Name string `json:"n"`
		From string `json:"f"`
		To   string `json:"t"`
	}
	sig := struct {
		Classes   []classSig `json:"classes"`
		Relations []relSig   `json:"relations"`
		Operators []string   `json:"operators"`
		AggFuncs  []string   `json:"agg_functions"`
	}{
		Classes:   make([]classSig, 0, len(v.Classes)),
		Relations: make([]relSig, 0, len(v.Relations)),
		Operators: v.Operators,
		AggFuncs:  v.AggFunctions,
	}
	for _, c := range v.Classes {
		cs := classSig{Name: c.Name, ClassType: c.ClassType, Properties: make([]propSig, 0, len(c.Properties))}
		for _, p := range c.Properties {
			cs.Properties = append(cs.Properties, propSig{Name: p.Name, DataType: p.DataType, Synonyms: p.Synonyms, Required: p.Required})
		}
		sig.Classes = append(sig.Classes, cs)
	}
	for _, r := range v.Relations {
		sig.Relations = append(sig.Relations, relSig{Name: r.Name, From: r.From, To: r.To})
	}
	buf, _ := json.Marshal(sig)
	return shortHash(buf)
}

// shortHash 返回内容的 sha256 前 16 位十六进制串。
func shortHash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])[:16]
}

// parseSynonyms 解析同义词字段：支持 JSON 数组或逗号分隔字符串。
func parseSynonyms(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if strings.HasPrefix(raw, "[") {
		var arr []string
		if err := json.Unmarshal([]byte(raw), &arr); err == nil {
			out := make([]string, 0, len(arr))
			for _, s := range arr {
				if s = strings.TrimSpace(s); s != "" {
					out = append(out, s)
				}
			}
			return out
		}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// classNames 返回按字母序排列的类名列表，用于稳定的 prompt 输出。
func (v *Vocabulary) classNames() []string {
	names := make([]string, 0, len(v.Classes))
	for _, c := range v.Classes {
		names = append(names, c.Name)
	}
	sort.Strings(names)
	return names
}
