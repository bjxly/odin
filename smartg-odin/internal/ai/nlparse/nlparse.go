// Package nlparse 实现服务端「规则版」自然语言解析器。
//
// 它把前端 odin-console parseNaturalLanguage 的正则规则移植为 Go，完全由本体
// 元数据（类名 / label / 同义词、属性名 / label / 同义词、关系、规则）驱动，
// 将一句自然语言解析为确定性的 query.QueryRequest。
//
// 设计原则：解析过程无任何随机性，同一句话 + 同一本体元数据必得同一 QueryRequest，
// 从而保证「同问同 SQL」。未命中本体类时 matched=false（P2 再引入 LLM 回退）。
package nlparse

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"smartg-odin/internal/ds/query"
	"smartg-odin/internal/model"
	"smartg-odin/internal/ont/reason"
)

// Result 规则解析结果。
type Result struct {
	Query     *query.QueryRequest `json:"query"`
	Notes     []string            `json:"notes"`
	Hits      []string            `json:"hits"`
	ParsePath string              `json:"parse_path"`
	Matched   bool                `json:"matched"`
}

// ---- 元数据视图 ----

type ontProp struct {
	Name     string
	Label    string
	DataType string
	Synonyms []string
}

type ontClass struct {
	ID            uint
	Name          string
	Label         string
	ClassType     string
	ParentClassID *uint
	Synonyms      []string
	Properties    []ontProp
}

type ontRelation struct {
	Name         string
	Label        string
	FromClassID  uint
	ToClassID    uint
	RelationType string
}

type propAlias struct {
	alias string
	prop  ontProp
}

// ---- 正则规则（移植自前端 NL_OPS / NL_NULL_RE 等） ----

type nlOp struct {
	re *regexp.Regexp
	op string
}

// nlOps 比较词表，顺序即优先级（长/强匹配在前），与前端 NL_OPS 一致。
var nlOps = []nlOp{
	{regexp.MustCompile(`(?i)^\s*(?:>=|≥|大于等于|不低于|不小于)`), "gte"},
	{regexp.MustCompile(`(?i)^\s*(?:<=|≤|小于等于|不超过|不大于)`), "lte"},
	{regexp.MustCompile(`(?i)^\s*(?:!=|<>|≠|不等于)`), "ne"},
	{regexp.MustCompile(`(?i)^\s*(?:>|大于|超过|高于)`), "gt"},
	{regexp.MustCompile(`(?i)^\s*(?:<|小于|低于)`), "lt"},
	{regexp.MustCompile(`(?i)^\s*(?:包含|含有|模糊匹配|like)`), "like"},
	{regexp.MustCompile(`(?i)^\s*(?:等于|为|是|=|:|：)`), "eq"},
}

var nlNullRe = regexp.MustCompile(`(?i)^\s*(?:为空|是空|不存在|没值|is\s+null)`)
var nlValueRe = regexp.MustCompile(`^\s*["'“]?([^"'“”，,。;；、\s]+)["'”]?`)
var nlOrderByRe = regexp.MustCompile(`(?:按|根据|依)\s*([^\s，,。;；、]{2,20})\s*(降序|从高到低|从大到小|倒序|升序|从低到高|从小到大|排序)?`)
var nlOrderByDescRe = regexp.MustCompile(`降序|从高到低|从大到小|倒序`)
var nlLimitTopRe = regexp.MustCompile(`(?i)(?:前|top)\s*(\d+)`)
var nlLimitWordRe = regexp.MustCompile(`(?i)(?:限制|limit)\s*(\d+)`)
var nlLimitCountRe = regexp.MustCompile(`(\d+)\s*[条项个行]`)

// nlBareIDRe 裸实体标识符正则（P0：治「实体键无比较词 → 漏过滤 → 全表查询」）。
// 匹配形如 ORD-20250115-007、SKU-12345、CUST-A1 这类「字母前缀 + 连字符 + 字母/数字串」标识符。
// 保守策略：必须含连字符，且在 injectBareEntityFilter/ContainsEntityIdentifier 中再要求「整段含数字」，
// 以免误伤 state-of-the-art、e-mail 这类普通连字符词。
var nlBareIDRe = regexp.MustCompile(`[A-Za-z]{2,8}-[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*`)

// ---- 聚合关键词正则（P0/P2）----
//
// 设计要点：当「可求和度量（来自关系目标类）」与「排名/TOP-N」共现时，优先归为
// SUM(度量) + 按主类维度分组（如「销售数量最多的产品TOP3」= SUM(quantity) GROUP BY product），
// 该分支优先于泛化的 count 关键词（统计/多少），以避免「统计…」被误判为 COUNT。
var nlAggCountRe = regexp.MustCompile(`多少|几笔|几多|总共|总数|共有|笔数|多少个|多少笔|多少条|统计|计数`)
var nlAggAvgRe = regexp.MustCompile(`平均|均值`)
var nlAggSumRe = regexp.MustCompile(`总和|求和|总量|汇总|合计`)
var nlAggMaxRe = regexp.MustCompile(`最多|最高|最大`)
var nlAggMinRe = regexp.MustCompile(`最少|最低|最小`)
var nlAggRankRe = regexp.MustCompile(`最多|最高|最大|最少|最低|最小|排名|排行`)
var nlAggGroupRe = regexp.MustCompile(`每个|分组|各类|各项|按.{1,12}分`)

// nlSumMeasureRe 显式「求和类度量词」（#38 缺口补齐）：即便缺少「统计/求和/平均」等聚合动词，
// 出现这些词也按 SUM 处理——销量/销售数量/销售量→quantity 求和，总额/总金额/销售额/销售金额/总价→amount 求和。
// 与 #34 已补充的属性同义词口径一致，避免重复造轮子。
var nlSumMeasureRe = regexp.MustCompile(`销量|销售数量|销售量|销售额|销售金额|总金额|总额|总价|总量|合计`)

// aggFuncLabelCN 聚合函数的中文前缀，必须与 translator.aggFuncLabels 保持一致（缺省别名口径）。
var aggFuncLabelCN = map[string]string{"count": "计数", "sum": "合计", "avg": "平均", "min": "最小", "max": "最大"}

// opLabels 操作符的中文可读标签，用于 notes/hits。
var opLabels = map[string]string{
	"eq": "等于", "ne": "不等于", "neq": "不等于",
	"gt": "大于", "gte": "大于等于", "lt": "小于", "lte": "小于等于",
	"like": "包含", "in": "属于", "not_in": "不属于",
	"is_null": "为空", "is_not_null": "非空",
}

// Parse 依据本体元数据将自然语言解析为查询请求。
func Parse(db *gorm.DB, ontologyID uint, text string) *Result {
	res := &Result{
		Query:     &query.QueryRequest{OntologyID: ontologyID},
		Notes:     make([]string, 0),
		Hits:      make([]string, 0),
		ParsePath: "rule",
		Matched:   false,
	}

	raw := strings.TrimSpace(text)
	if raw == "" {
		res.Notes = append(res.Notes, "请输入查询意图")
		return res
	}
	lower := strings.ToLower(raw)

	classes := loadClasses(db, ontologyID)
	relations := loadRelations(db, ontologyID)
	if len(classes) == 0 {
		res.Notes = append(res.Notes, "本体下没有可用的类")
		return res
	}
	byName := make(map[string]*ontClass, len(classes))
	byID := make(map[uint]*ontClass, len(classes))
	for i := range classes {
		byName[classes[i].Name] = &classes[i]
		byID[classes[i].ID] = &classes[i]
	}

	// 1) 主类识别：按 label / name / 同义词包含匹配，label 长者优先。
	hits := make([]*ontClass, 0)
	for i := range classes {
		c := &classes[i]
		keys := make([]string, 0, 3)
		if c.Label != "" {
			keys = append(keys, c.Label)
		}
		if c.Name != "" {
			keys = append(keys, c.Name)
		}
		keys = append(keys, c.Synonyms...)
		for _, k := range keys {
			k = strings.ToLower(strings.TrimSpace(k))
			if k != "" && strings.Contains(lower, k) {
				hits = append(hits, c)
				break
			}
		}
	}
	sort.SliceStable(hits, func(i, j int) bool {
		return len([]rune(hits[i].Label)) > len([]rune(hits[j].Label))
	})

	if len(hits) == 0 {
		res.Notes = append(res.Notes, "未识别到本体类，请手动选择")
		return res
	}

	primary := hits[0]
	res.Matched = true
	res.Hits = append(res.Hits, primary.Label)
	res.Notes = append(res.Notes, "主类："+displayLabel(primary))

	// 2) 解析基类（虚拟类落到父类属性空间，并作为规范化 class_name）。
	base := resolveBase(primary, byID)
	if base != nil && base.ID != primary.ID {
		res.Notes = append(res.Notes, "规范化到基类："+displayLabel(base))
	}

	// 3) 确定最终 class_name：虚拟类归一到基类。
	target := base
	if target == nil {
		target = primary
	}
	res.Query.ClassName = target.Name

	// 4) 虚拟类语义化：命中以该类为 target_class 的启用规则，把条件转为过滤。
	if strings.EqualFold(primary.ClassType, "virtual") {
		res.Notes = append(res.Notes, displayLabel(primary)+" 为虚拟类，套用派生规则")
		applyVirtualRules(db, ontologyID, primary.Name, res)
	}

	// 5) 属性过滤：<属性别名><比较词><值>。
	parseFilters(raw, lower, target, res)

	// 6) 排序：按 <属性> 降序/升序。
	parseOrderBy(raw, target, res)

	// 7) limit。
	parseLimit(raw, res)

	// 8) join：其余命中类若与主类/基类存在关系则挂载。
	parseRelations(relations, hits, primary, target, byID, res)

	// 9) 聚合识别（count/sum/avg/min/max + TOP-N 复合模式 + 按度量属性反推跨类关系）。
	parseAggregation(raw, lower, classes, relations, primary, target, res)

	if len(res.Hits) == 0 {
		res.Notes = append(res.Notes, "未识别到明确条件，可手动调整")
	}
	return res
}

// applyVirtualRules 查找 target_class==虚拟类 的启用规则，把条件转换为过滤注入。
func applyVirtualRules(db *gorm.DB, ontologyID uint, virtualName string, res *Result) {
	var rules []model.OntRule
	db.Where("ontology_id = ? AND enabled = ?", ontologyID, true).
		Order("priority asc, id asc").
		Find(&rules)
	for _, r := range rules {
		act, aerr := reason.ParseAction(r.ActionJSON)
		cond, cerr := reason.ParseCondition(r.ConditionJSON)
		if aerr != nil || cerr != nil {
			continue
		}
		if !strings.EqualFold(act.Type, "expand_class") ||
			!strings.EqualFold(act.TargetClass, virtualName) ||
			cond.Property == "" {
			continue
		}
		op := normalizeOp(cond.Operator)
		res.Query.Filters = append(res.Query.Filters, query.Filter{
			Property: cond.Property,
			Op:       op,
			Value:    cond.Value,
		})
		res.Hits = append(res.Hits, cond.Property+" "+opLabel(op)+" "+valueString(cond.Value))
		res.Notes = append(res.Notes, "规则["+r.Name+"]派生过滤："+cond.Property+" "+opLabel(op)+" "+valueString(cond.Value))
	}
}

// parseFilters 移植前端「属性别名 + 比较词 + 值」的过滤解析。
func parseFilters(raw, lower string, scope *ontClass, res *Result) {
	if scope == nil {
		return
	}
	aliases := buildAliases(scope)
	consumed := make([][2]int, 0)
	for _, a := range aliases {
		la := strings.ToLower(a.alias)
		idx := strings.Index(lower, la)
		if idx < 0 {
			continue
		}
		end := idx + len(la)
		if overlaps(consumed, idx, end) {
			continue
		}
		restText := raw[end:]

		if m := nlNullRe.FindString(restText); m != "" {
			res.Query.Filters = append(res.Query.Filters, query.Filter{Property: a.prop.Name, Op: "is_null"})
			res.Hits = append(res.Hits, a.alias+" 为空")
			res.Notes = append(res.Notes, "过滤："+propLabel(a.prop)+" 为空")
			consumed = append(consumed, [2]int{idx, end + len(m)})
			continue
		}

		op, opText := matchOp(restText)
		if op == "" {
			continue
		}
		after := restText[len(opText):]
		vm := nlValueRe.FindStringSubmatch(after)
		if vm == nil || len(vm) < 2 || vm[1] == "" {
			continue
		}
		value := vm[1]
		res.Query.Filters = append(res.Query.Filters, query.Filter{Property: a.prop.Name, Op: op, Value: value})
		res.Hits = append(res.Hits, a.alias+" "+opLabel(op)+" "+value)
		res.Notes = append(res.Notes, "过滤："+propLabel(a.prop)+" "+opLabel(op)+" "+value)
		consumed = append(consumed, [2]int{idx, end + len(opText) + len(vm[0])})
	}

	// 裸实体标识符兜底（P0）：将 ORD-20250115-007 / SKU-xxx 这类无比较词的标识符
	// 映射到 scope 类的键属性，产出 eq 过滤；仅当该键属性尚未被上面的三段式抽到时才注入。
	injectBareEntityFilter(raw, scope, consumed, res)
}

// injectBareEntityFilter 识别文本中的裸实体标识符（如 ORD-20250115-007），映射到 scope 类的键属性，
// 产出 Filter{Property: 键属性, Op: eq, Value: 标识符}。仅当该键属性尚未被比较词三段式抽到
// （hasFilterOnProp 为 false）、且标识符片段未被其他分支消费、且标识符含数字时才注入，避免重复/误伤。
func injectBareEntityFilter(raw string, scope *ontClass, consumed [][2]int, res *Result) {
	if scope == nil || res == nil || res.Query == nil {
		return
	}
	keyProp := findKeyProp(scope)
	if keyProp == nil {
		return
	}
	// 已有针对键属性的过滤（比较词三段式已抽到）→ 不重复注入。
	if hasFilterOnProp(res, keyProp.Name) {
		return
	}
	for _, m := range nlBareIDRe.FindAllStringIndex(raw, -1) {
		if overlaps(consumed, m[0], m[1]) {
			continue
		}
		tok := raw[m[0]:m[1]]
		if !containsDigit(tok) {
			continue
		}
		res.Query.Filters = append(res.Query.Filters, query.Filter{Property: keyProp.Name, Op: "eq", Value: tok})
		res.Hits = append(res.Hits, keyProp.Name+" "+opLabel("eq")+" "+tok)
		res.Notes = append(res.Notes, "过滤："+propLabel(*keyProp)+" "+opLabel("eq")+" "+tok+"（裸实体标识符）")
		return
	}
}

// ContainsEntityIdentifier 报告文本是否含「前缀-数字/字母串」型实体标识符（如 ORD-20250115-007、SKU-12345）。
// 供 intent.go 在「规则命中主类但无过滤」时判定是否需降级到 LLM 意图链补救；
// 判定保守（须含连字符且带数字），避免误伤普通词。
func ContainsEntityIdentifier(text string) bool {
	for _, tok := range nlBareIDRe.FindAllString(text, -1) {
		if containsDigit(tok) {
			return true
		}
	}
	return false
}

// EntityIdentifier 文本中抽到的一个裸实体标识符（含其在原文中的字节区间）。
type EntityIdentifier struct {
	Value string // 标识符原文，如 ORD-20250115-007
	Start int    // 字节起始下标（含）
	End   int    // 字节结束下标（不含）
}

// ExtractEntityIdentifiers 按出现顺序抽取文本中全部「含数字」的裸实体标识符。
//
// 与 ContainsEntityIdentifier 同口径（nlBareIDRe + containsDigit），供 ai 包的确定性词法
// 入口锚（ResolveEntryAnchorLexical）复用作实体键值来源，避免两处正则漂移。
// 返回按 Start 升序，无命中时返回 nil。
func ExtractEntityIdentifiers(text string) []EntityIdentifier {
	if text == "" {
		return nil
	}
	var out []EntityIdentifier
	for _, m := range nlBareIDRe.FindAllStringIndex(text, -1) {
		tok := text[m[0]:m[1]]
		if !containsDigit(tok) {
			continue
		}
		out = append(out, EntityIdentifier{Value: tok, Start: m[0], End: m[1]})
	}
	return out
}

// findKeyProp 返回 scope 类的「键属性」（订单号/SKU/编码等实体标识符），按属性顺序取首个命中者，保证确定性。
// 说明：当前 ont_class_property 无 is_key 列，故以命名/标签约定识别键属性（见 isKeyProperty）；
// seed 中 Order.orderNo（Label=订单号）即由此命中。
func findKeyProp(scope *ontClass) *ontProp {
	if scope == nil {
		return nil
	}
	for i := range scope.Properties {
		if isKeyProperty(scope.Properties[i]) {
			return &scope.Properties[i]
		}
	}
	return nil
}

// isKeyProperty 以命名/标签约定判定一个属性是否为实体键（标识符）：
//   - Name 以 no/id/code/key 结尾（orderNo→orderno、shipmentNo、skuCode、customerId）；或
//   - Label/同义词含「号/编号/编码/代码/标识」。
func isKeyProperty(p ontProp) bool {
	name := strings.ToLower(strings.TrimSpace(p.Name))
	for _, suf := range []string{"no", "id", "code", "key"} {
		if strings.HasSuffix(name, suf) {
			return true
		}
	}
	labels := append([]string{p.Label}, p.Synonyms...)
	for _, s := range labels {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		for _, kw := range []string{"号", "编号", "编码", "代码", "标识"} {
			if strings.Contains(s, kw) {
				return true
			}
		}
	}
	return false
}

// containsDigit 报告字符串是否含至少一个 ASCII 数字。
func containsDigit(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

// matchOp 按优先级返回命中的操作符及其匹配到的原文片段。
func matchOp(s string) (string, string) {
	for _, o := range nlOps {
		if m := o.re.FindString(s); m != "" {
			return o.op, m
		}
	}
	return "", ""
}

// parseOrderBy 解析「按 X 降序/升序」。
func parseOrderBy(raw string, scope *ontClass, res *Result) {
	m := nlOrderByRe.FindStringSubmatch(raw)
	if m == nil || len(m) < 2 {
		return
	}
	name := strings.TrimSpace(m[1])
	p := findProp(scope, name)
	if p == nil {
		return
	}
	desc := len(m) >= 3 && m[2] != "" && nlOrderByDescRe.MatchString(m[2])
	res.Query.OrderBy = p.Name
	if desc {
		res.Query.OrderDir = "desc"
	} else {
		res.Query.OrderDir = "asc"
	}
	dir := "升序"
	if desc {
		dir = "降序"
	}
	res.Hits = append(res.Hits, "按 "+propLabel(*p)+" "+dir)
	res.Notes = append(res.Notes, "排序："+propLabel(*p)+" "+dir)
}

// parseLimit 解析 limit：前 N / top N / 限制 N / N 条。
func parseLimit(raw string, res *Result) {
	var m []string
	if m = nlLimitTopRe.FindStringSubmatch(raw); m == nil {
		if m = nlLimitWordRe.FindStringSubmatch(raw); m == nil {
			m = nlLimitCountRe.FindStringSubmatch(raw)
		}
	}
	if m == nil || len(m) < 2 {
		return
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n <= 0 {
		return
	}
	res.Query.Limit = n
	res.Hits = append(res.Hits, "limit="+strconv.Itoa(n))
	res.Notes = append(res.Notes, "Limit = "+strconv.Itoa(n))
}

// parseRelations 挂载其余命中类之间的关系（最多 2 个）。
func parseRelations(relations []ontRelation, hits []*ontClass, primary, base *ontClass, byID map[uint]*ontClass, res *Result) {
	anchorID := primary.ID
	if base != nil {
		anchorID = base.ID
	}
	used := map[string]bool{}
	for _, oc := range hits[1:] {
		if len(res.Query.Relations) >= 2 {
			break
		}
		// 跳过与锚点/主类相同的类，避免自指连接。
		if oc.ID == anchorID || oc.ID == primary.ID {
			continue
		}
		for _, rel := range relations {
			if used[rel.Name] {
				continue
			}
			// 继承关系不是可 JOIN 的关联，跳过。
			if strings.EqualFold(rel.RelationType, "inheritance") {
				continue
			}
			from := byID[rel.FromClassID]
			to := byID[rel.ToClassID]
			if from == nil || to == nil {
				continue
			}
			connect := (rel.FromClassID == anchorID && rel.ToClassID == oc.ID) ||
				(rel.ToClassID == anchorID && rel.FromClassID == oc.ID) ||
				(rel.FromClassID == primary.ID && rel.ToClassID == oc.ID) ||
				(rel.ToClassID == primary.ID && rel.FromClassID == oc.ID)
			if !connect {
				continue
			}
			res.Query.Relations = append(res.Query.Relations, rel.Name)
			used[rel.Name] = true
			res.Hits = append(res.Hits, oc.Label)
			res.Notes = append(res.Notes, "Join："+rel.Label+"（"+from.Name+" → "+to.Name+"）")
			break
		}
	}
}

// measureHit 命中的可求和度量属性及其所属类。
type measureHit struct {
	prop  ontProp
	cls   *ontClass
	alias string
}

// parseAggregation 识别聚合意图并写入 aggregates/group_by/having/order_by。
//
// 判定优先级（确定性，无随机）：
//  1. 度量来自关系目标类 + 排名/TOP-N → SUM(度量) 并按主类维度分组（跨类聚合，如产品销量 TOP-N）；
//  2. 平均/求和/最大/最小 关键词 + 度量 → 对应聚合；
//  3. count 关键词 → COUNT(度量) 或 COUNT(*)（无度量时）；
//  4. 以上均不命中 → 非聚合查询，直接返回。
func parseAggregation(raw, lower string, classes []ontClass, relations []ontRelation, primary, target *ontClass, res *Result) {
	mainClass := target
	if mainClass == nil {
		mainClass = primary
	}
	if mainClass == nil {
		return
	}

	isTopN := nlLimitTopRe.MatchString(raw)
	hasRank := nlAggRankRe.MatchString(raw)
	hasCount := nlAggCountRe.MatchString(raw)
	hasAvg := nlAggAvgRe.MatchString(raw)
	hasSum := nlAggSumRe.MatchString(raw)
	hasMax := nlAggMaxRe.MatchString(raw)
	hasMin := nlAggMinRe.MatchString(raw)
	hasGroupKw := nlAggGroupRe.MatchString(raw)
	// #38 缺口补齐：
	//  - hasSumWord：显式「求和类度量词」（销量/总额…），无聚合动词时也按 sum 处理；
	//  - hasCountMeasure：「<主类>数/数量/数目」计数度量词（如「订单数」），归入计数语义。
	hasSumWord := nlSumMeasureRe.MatchString(raw)
	hasCountMeasure := hasCountMeasureWord(raw, mainClass)
	hasCountWord := hasCount || hasCountMeasure

	measure := findMeasure(lower, classes, mainClass.ID)

	// 「<主类>数量」这类计数度量词（如「订单数量」）语义为「主类实体的计数」，
	// 而非对 quantity 字段求和/计数；在无 平均/求和度量词/排名 时抑制度量，避免误判。
	if hasCountMeasure && !hasAvg && !hasSumWord && !hasRank && !isTopN {
		measure = nil
	}

	var fn, prop, propLabel string
	measureFromRelated := false
	if measure != nil {
		prop = measure.prop.Name
		propLabel = measure.prop.Label
		if measure.cls.ID != mainClass.ID {
			measureFromRelated = true
		}
	}

	// 求和类度量词若与该属性上的过滤共现（如「总金额大于1000」），保守地不触发聚合，
	// 以免把过滤查询误判为聚合查询。
	sumWordEffective := hasSumWord && !(measure != nil && !hasSum && hasFilterOnProp(res, prop))

	switch {
	case measure != nil && measureFromRelated && (hasRank || isTopN):
		fn = "sum"
	case measure != nil && hasAvg:
		fn = "avg"
	case measure != nil && (hasSum || sumWordEffective):
		fn = "sum"
	case measure != nil && hasMax:
		fn = "max"
	case measure != nil && hasMin:
		fn = "min"
	case measure != nil && measureFromRelated:
		// 跨类数值度量（如「每个产品的销量」）在无其它动词时，默认按主类维度求和。
		fn = "sum"
	case hasCountWord:
		fn = "count"
	default:
		return // 非聚合查询
	}

	// count 且（无度量 或 计数度量词）→ COUNT(*)（属性置空）。
	if fn == "count" && (measure == nil || hasCountMeasure) {
		prop = ""
		propLabel = ""
	}

	alias := aggFuncLabelCN[fn]
	if strings.TrimSpace(propLabel) != "" {
		alias += propLabel
	}
	res.Query.Aggregates = append(res.Query.Aggregates, query.Aggregate{Func: fn, Property: prop, Alias: alias})
	res.Hits = append(res.Hits, "聚合:"+fn)
	res.Notes = append(res.Notes, "聚合："+fn+"("+prop+") 别名="+alias)

	// 跨类度量：反推并挂载「主类→度量类」关系（方向与 translator.buildJoin 一致）。
	if measureFromRelated {
		if relName := findRelationBetween(relations, mainClass.ID, measure.cls.ID); relName != "" && !containsStr(res.Query.Relations, relName) {
			res.Query.Relations = append(res.Query.Relations, relName)
			res.Notes = append(res.Notes, "聚合反推关系："+relName+"（"+mainClass.Name+" → "+measure.cls.Name+"）")
		}
	}

	// 分组（#38 缺口补齐）：先从文本「各X/每个X/按X统计/X的总数」显式解析主类维度属性，
	// 保留文本出现顺序、可复现；若无显式维度但存在跨类度量或分组关键词，则回退到主类主维度。
	parseGroupBy(raw, mainClass, res)
	if len(res.Query.GroupBy) == 0 && (measureFromRelated || (hasGroupKw && measure != nil)) {
		if dim := findDimensionProp(mainClass); dim != nil && !containsStr(res.Query.GroupBy, dim.Name) {
			res.Query.GroupBy = append(res.Query.GroupBy, dim.Name)
			res.Notes = append(res.Notes, "分组维度："+propLabelOf(*dim))
		}
	}

	// TOP-N / 排名 → order_by 聚合别名 + 方向（最少/最低 → asc，其余 → desc）。
	if isTopN || hasRank {
		res.Query.OrderBy = alias
		if hasMin && !hasMax {
			res.Query.OrderDir = "asc"
		} else {
			res.Query.OrderDir = "desc"
		}
	}
}

// ---- 加载与工具 ----

func loadClasses(db *gorm.DB, ontologyID uint) []ontClass {
	var rows []model.OntClass
	db.Where("ontology_id = ?", ontologyID).Order("id asc").Find(&rows)
	classes := make([]ontClass, 0, len(rows))
	for _, r := range rows {
		classes = append(classes, ontClass{
			ID:            r.ID,
			Name:          r.Name,
			Label:         r.Label,
			ClassType:     r.ClassType,
			ParentClassID: r.ParentClassID,
			Synonyms:      parseSynonyms(r.Synonyms),
		})
	}
	// 属性。
	var props []model.OntClassProperty
	db.Where("ontology_id = ?", ontologyID).Order("class_id asc, sort_order asc, id asc").Find(&props)
	idx := map[uint]int{}
	for i := range classes {
		idx[classes[i].ID] = i
	}
	for _, p := range props {
		i, okk := idx[p.ClassID]
		if !okk {
			continue
		}
		classes[i].Properties = append(classes[i].Properties, ontProp{
			Name:     p.Name,
			Label:    p.Label,
			DataType: p.DataType,
			Synonyms: parseSynonyms(p.Synonyms),
		})
	}
	return classes
}

func loadRelations(db *gorm.DB, ontologyID uint) []ontRelation {
	var rows []model.OntRelation
	db.Where("ontology_id = ?", ontologyID).Order("id asc").Find(&rows)
	out := make([]ontRelation, 0, len(rows))
	for _, r := range rows {
		out = append(out, ontRelation{
			Name:         r.Name,
			Label:        r.Label,
			FromClassID:  r.FromClassID,
			ToClassID:    r.ToClassID,
			RelationType: r.RelationType,
		})
	}
	return out
}

// resolveBase 沿 parent_class_id 上溯到最顶层基类。
func resolveBase(c *ontClass, byID map[uint]*ontClass) *ontClass {
	if c == nil {
		return nil
	}
	base := c
	seen := map[uint]bool{}
	for base.ParentClassID != nil && !seen[base.ID] {
		seen[base.ID] = true
		parent, ok := byID[*base.ParentClassID]
		if !ok {
			break
		}
		base = parent
	}
	return base
}

// buildAliases 收集类属性别名（label / name / 同义词），长名优先。
func buildAliases(c *ontClass) []propAlias {
	out := make([]propAlias, 0)
	for _, p := range c.Properties {
		names := make([]string, 0, 3)
		if p.Label != "" {
			names = append(names, p.Label)
		}
		if p.Name != "" {
			names = append(names, p.Name)
		}
		names = append(names, p.Synonyms...)
		for _, n := range names {
			n = strings.TrimSpace(n)
			if len([]rune(n)) >= 2 {
				out = append(out, propAlias{alias: n, prop: p})
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return len([]rune(out[i].alias)) > len([]rune(out[j].alias))
	})
	return out
}

func findProp(c *ontClass, name string) *ontProp {
	if c == nil {
		return nil
	}
	name = strings.TrimSpace(name)
	for i := range c.Properties {
		p := &c.Properties[i]
		if p.Name == name || p.Label == name {
			return p
		}
		for _, s := range p.Synonyms {
			if s == name {
				return p
			}
		}
	}
	return nil
}

func overlaps(consumed [][2]int, s, e int) bool {
	for _, r := range consumed {
		if s < r[1] && e > r[0] {
			return true
		}
	}
	return false
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

// normalizeOp 把规则/前端口径的操作符归一到后端实现集。
func normalizeOp(op string) string {
	switch strings.ToLower(strings.TrimSpace(op)) {
	case "", "eq", "=":
		return "eq"
	case "isnull", "is_null":
		return "is_null"
	case "notnull", "not_null", "is_not_null":
		return "is_not_null"
	case "contains":
		return "like"
	case "neq":
		return "neq"
	default:
		return strings.ToLower(strings.TrimSpace(op))
	}
}

func opLabel(op string) string {
	if l, ok := opLabels[op]; ok {
		return l
	}
	return op
}

func propLabel(p ontProp) string {
	if p.Label != "" {
		return p.Label
	}
	return p.Name
}

func displayLabel(c *ontClass) string {
	if c.Label != "" {
		return c.Label + "（" + c.Name + "）"
	}
	return c.Name
}

func valueString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch tv := v.(type) {
	case string:
		return tv
	case float64:
		if tv == float64(int64(tv)) {
			return strconv.FormatInt(int64(tv), 10)
		}
		return strconv.FormatFloat(tv, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", tv)
	}
}

// findMeasure 在文本中查找命中的可求和数值度量属性（最长别名优先；同长时主类属性优先，
// 如「产品的平均单价」应取 Product.price 而非 Order.unitPrice；再按类/属性原始顺序稳定）。
func findMeasure(lower string, classes []ontClass, mainClassID uint) *measureHit {
	type cand struct {
		alias string
		ci    int
		pi    int
	}
	cands := make([]cand, 0)
	for ci := range classes {
		for pi := range classes[ci].Properties {
			p := classes[ci].Properties[pi]
			if !isNumericType(p.DataType) {
				continue
			}
			names := make([]string, 0, 3)
			if p.Label != "" {
				names = append(names, p.Label)
			}
			if p.Name != "" {
				names = append(names, p.Name)
			}
			names = append(names, p.Synonyms...)
			for _, n := range names {
				n = strings.TrimSpace(n)
				if len([]rune(n)) >= 2 && strings.Contains(lower, strings.ToLower(n)) {
					cands = append(cands, cand{alias: n, ci: ci, pi: pi})
				}
			}
		}
	}
	if len(cands) == 0 {
		return nil
	}
	sort.SliceStable(cands, func(i, j int) bool {
		li, lj := len([]rune(cands[i].alias)), len([]rune(cands[j].alias))
		if li != lj {
			return li > lj
		}
		// 同长时主类属性优先（确定性地消除跨类同名属性歧义，如 单价）。
		mi := classes[cands[i].ci].ID == mainClassID
		mj := classes[cands[j].ci].ID == mainClassID
		if mi != mj {
			return mi
		}
		return false // 保持原始（类 id、属性序）顺序
	})
	best := cands[0]
	return &measureHit{prop: classes[best.ci].Properties[best.pi], cls: &classes[best.ci], alias: best.alias}
}

// hasCountMeasureWord 判断文本是否含「<主类>数/数量/数目」这类计数度量词（如 订单数、客户数）。
// 以主类 label/name/同义词 + “数” 为锚点，避免泛化的 “数” 误伤属性名（如 “数量”）。
func hasCountMeasureWord(raw string, mainClass *ontClass) bool {
	if mainClass == nil {
		return false
	}
	keys := make([]string, 0, 3)
	if mainClass.Label != "" {
		keys = append(keys, mainClass.Label)
	}
	if mainClass.Name != "" {
		keys = append(keys, mainClass.Name)
	}
	keys = append(keys, mainClass.Synonyms...)
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if strings.Contains(raw, k+"数") {
			return true
		}
	}
	return false
}

// hasFilterOnProp 判断 res 中是否已存在针对指定属性的过滤条件。
func hasFilterOnProp(res *Result, prop string) bool {
	if prop == "" {
		return false
	}
	for _, f := range res.Query.Filters {
		if f.Property == prop {
			return true
		}
	}
	return false
}

// parseGroupBy 从「各X/每个X/每X/逐X/按X统计(分组/分类/汇总)/X的总数」等模式解析分组维度，
// X 归一到主类属性（label/name/同义词）。按文本出现位置排序追加，保证确定性与可复现。
func parseGroupBy(raw string, mainClass *ontClass, res *Result) {
	if mainClass == nil {
		return
	}
	type ghit struct {
		pos   int
		name  string
		label string
	}
	hits := make([]ghit, 0)
	for i := range mainClass.Properties {
		p := mainClass.Properties[i]
		if containsStr(res.Query.GroupBy, p.Name) {
			continue
		}
		pos, ok := groupPatternPos(raw, p)
		if !ok {
			continue
		}
		hits = append(hits, ghit{pos: pos, name: p.Name, label: propLabel(p)})
	}
	sort.SliceStable(hits, func(a, b int) bool { return hits[a].pos < hits[b].pos })
	for _, h := range hits {
		if containsStr(res.Query.GroupBy, h.name) {
			continue
		}
		res.Query.GroupBy = append(res.Query.GroupBy, h.name)
		res.Hits = append(res.Hits, "分组:"+h.label)
		res.Notes = append(res.Notes, "分组维度："+h.label)
	}
}

// groupPatternPos 若属性 p 的某别名出现在分组语境中，返回最靠前的文本位置与 true。
func groupPatternPos(raw string, p ontProp) (int, bool) {
	aliases := make([]string, 0, 3)
	if p.Label != "" {
		aliases = append(aliases, p.Label)
	}
	if p.Name != "" {
		aliases = append(aliases, p.Name)
	}
	aliases = append(aliases, p.Synonyms...)
	best, found := -1, false
	for _, a := range aliases {
		a = strings.TrimSpace(a)
		if len([]rune(a)) < 2 {
			continue
		}
		if pos, ok := hasGroupPattern(raw, a); ok && (!found || pos < best) {
			best, found = pos, true
		}
	}
	return best, found
}

// hasGroupPattern 判断别名 alias 是否出现在分组语境（各X/每个X/X的总数/按X统计…），返回位置。
// 注意排除「按X排序」这类排序语境（仅当「按X」后接分组动词/「的」时才视为分组）。
func hasGroupPattern(raw, alias string) (int, bool) {
	for _, pre := range []string{"每个", "各", "每", "逐"} {
		if idx := strings.Index(raw, pre+alias); idx >= 0 {
			return idx, true
		}
	}
	for _, suf := range []string{"的总数", "总数"} {
		if idx := strings.Index(raw, alias+suf); idx >= 0 {
			return idx, true
		}
	}
	if idx := strings.Index(raw, "按"+alias); idx >= 0 {
		rest := raw[idx+len("按"+alias):]
		for _, v := range []string{"统计", "分组", "分类", "汇总", "分别", "的"} {
			if strings.HasPrefix(rest, v) {
				return idx, true
			}
		}
	}
	return -1, false
}

// findDimensionProp 选取主类的分组维度属性：优先 name，其次 label 含「名称」，再次首个字符串属性。
func findDimensionProp(c *ontClass) *ontProp {
	if c == nil {
		return nil
	}
	for i := range c.Properties {
		if strings.EqualFold(c.Properties[i].Name, "name") {
			return &c.Properties[i]
		}
	}
	for i := range c.Properties {
		if strings.Contains(c.Properties[i].Label, "名称") {
			return &c.Properties[i]
		}
	}
	for i := range c.Properties {
		if strings.EqualFold(c.Properties[i].DataType, "string") {
			return &c.Properties[i]
		}
	}
	if len(c.Properties) > 0 {
		return &c.Properties[0]
	}
	return nil
}

// findRelationBetween 查找 from→to 的可 JOIN 关系名（跳过继承关系，按 id 升序取首个）。
func findRelationBetween(relations []ontRelation, fromID, toID uint) string {
	for _, rel := range relations {
		if strings.EqualFold(rel.RelationType, "inheritance") {
			continue
		}
		if rel.FromClassID == fromID && rel.ToClassID == toID {
			return rel.Name
		}
	}
	return ""
}

// isNumericType 判断数据类型是否为可求和的数值类型。
func isNumericType(dt string) bool {
	switch strings.ToLower(strings.TrimSpace(dt)) {
	case "integer", "int", "bigint", "long", "float", "double", "decimal", "numeric", "number":
		return true
	}
	return false
}

// containsStr 判断字符串切片是否已含某值。
func containsStr(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// propLabelOf 返回属性的可读标签（缺失回退属性名）。
func propLabelOf(p ontProp) string {
	if p.Label != "" {
		return p.Label
	}
	return p.Name
}
