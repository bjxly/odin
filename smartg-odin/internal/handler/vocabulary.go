package handler

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"

	"smartg-odin/internal/common/errs"
	"smartg-odin/internal/model"
)

// vocabularyProperty 词表中的属性项。
type vocabularyProperty struct {
	Name     string   `json:"name"`
	Label    string   `json:"label"`
	DataType string   `json:"data_type"`
	Synonyms []string `json:"synonyms"`
	Required bool     `json:"required"`
}

// vocabularyClass 词表中的类项。
type vocabularyClass struct {
	Name       string               `json:"name"`
	Label      string               `json:"label"`
	ClassType  string               `json:"class_type"`
	Synonyms   []string             `json:"synonyms"`
	Properties []vocabularyProperty `json:"properties"`
}

// vocabularyRelation 词表中的关系项。
type vocabularyRelation struct {
	Name      string `json:"name"`
	Label     string `json:"label"`
	FromClass string `json:"from_class"`
	ToClass   string `json:"to_class"`
}

// GetVocabulary GET /api/v1/ontologies/:id/vocabulary
// 返回本体词表：类（含属性、同义词）、关系、支持的过滤操作符、limit 边界与排序方向。
// 供 LLM 约束词表与前端使用。
func GetVocabulary(c *gin.Context) {
	id, appErr := parseIDParam(c, "id")
	if appErr != nil {
		fail(c, appErr)
		return
	}
	if appErr := ensureOntology(id); appErr != nil {
		fail(c, appErr)
		return
	}

	var classes []model.OntClass
	if err := DB().Where("ontology_id = ?", id).Order("id asc").Find(&classes).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}
	var props []model.OntClassProperty
	if err := DB().Where("ontology_id = ?", id).Order("class_id asc, sort_order asc, id asc").Find(&props).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}
	var relations []model.OntRelation
	if err := DB().Where("ontology_id = ?", id).Order("id asc").Find(&relations).Error; err != nil {
		fail(c, errs.DBError(err))
		return
	}

	// 类 id → name，供关系端点解析。
	nameByID := make(map[uint]string, len(classes))
	propsByClass := make(map[uint][]vocabularyProperty, len(classes))
	for _, p := range props {
		propsByClass[p.ClassID] = append(propsByClass[p.ClassID], vocabularyProperty{
			Name:     p.Name,
			Label:    p.Label,
			DataType: p.DataType,
			Synonyms: parseSynonymField(p.Synonyms),
			Required: p.Required,
		})
	}

	classViews := make([]vocabularyClass, 0, len(classes))
	for _, cl := range classes {
		nameByID[cl.ID] = cl.Name
		pv := propsByClass[cl.ID]
		if pv == nil {
			pv = []vocabularyProperty{}
		}
		classViews = append(classViews, vocabularyClass{
			Name:       cl.Name,
			Label:      cl.Label,
			ClassType:  cl.ClassType,
			Synonyms:   parseSynonymField(cl.Synonyms),
			Properties: pv,
		})
	}

	relViews := make([]vocabularyRelation, 0, len(relations))
	for _, r := range relations {
		relViews = append(relViews, vocabularyRelation{
			Name:      r.Name,
			Label:     r.Label,
			FromClass: nameByID[r.FromClassID],
			ToClass:   nameByID[r.ToClassID],
		})
	}

	ok(c, gin.H{
		"ontology_id":   id,
		"classes":       classViews,
		"relations":     relViews,
		"operators":     supportedFilterOps,
		"agg_functions": supportedAggFuncs,
		"limit":         gin.H{"min": 1, "max": maxQueryLimit},
		"order_dirs":    orderDirs,
	})
}

// parseSynonymField 解析同义词字段：支持 JSON 数组或逗号分隔字符串，始终返回非 nil 切片。
func parseSynonymField(raw string) []string {
	out := make([]string, 0)
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return out
	}
	if strings.HasPrefix(raw, "[") {
		var arr []string
		if err := json.Unmarshal([]byte(raw), &arr); err == nil {
			for _, s := range arr {
				if s = strings.TrimSpace(s); s != "" {
					out = append(out, s)
				}
			}
			return out
		}
	}
	for _, p := range strings.Split(raw, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
