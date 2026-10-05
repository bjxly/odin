package ai

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudwego/eino/schema"

	"smartg-odin/internal/ds/query"
)

// classByID 按主键查找类（词表规模小，线性查找即可）。
func (v *Vocabulary) classByID(id uint) *VocabClass {
	for i := range v.Classes {
		if v.Classes[i].ID == id {
			return &v.Classes[i]
		}
	}
	return nil
}

// propAliasMap 收集主类及其祖先类的属性别名（name/label/同义词，忽略大小写）到规范属性名的映射。
// 虚拟类的属性空间落在其基类上，故需沿 parent 链上溯合并。
func (v *Vocabulary) propAliasMap(cls *VocabClass) map[string]string {
	out := make(map[string]string)
	if cls == nil {
		return out
	}
	cur := cls
	seen := make(map[uint]bool)
	for cur != nil && !seen[cur.ID] {
		seen[cur.ID] = true
		for _, p := range cur.Properties {
			canonical := p.Name
			add := func(alias string) {
				alias = strings.ToLower(strings.TrimSpace(alias))
				if alias != "" {
					out[alias] = canonical
				}
			}
			add(p.Name)
			add(p.Label)
			for _, s := range p.Synonyms {
				add(s)
			}
		}
		if cur.ParentID == nil {
			break
		}
		cur = v.classByID(*cur.ParentID)
	}
	return out
}

// isSupportedOp 判断操作符是否在实现集内（忽略大小写与首尾空白）。
func isSupportedOp(op string) bool {
	op = strings.ToLower(strings.TrimSpace(op))
	for _, o := range supportedOperators {
		if o == op {
			return true
		}
	}
	return false
}

// isSupportedAggFunc 判断聚合函数是否在实现集内（忽略大小写与首尾空白）。
func isSupportedAggFunc(fn string) bool {
	fn = strings.ToLower(strings.TrimSpace(fn))
	for _, f := range supportedAggregateFuncs {
		if f == fn {
			return true
		}
	}
	return false
}

// aggPropResolvers 构建聚合上下文（aggregates/group_by）的属性解析器：
//   - unq：无限定属性名 → 规范属性名，覆盖「主类+祖先」并集，再并入各关系目标类
//     （主类优先，即主类已有的别名不被关系目标类覆盖，符合「主类→关系目标类」解析顺序）；
//   - qual："relationName.属性别名"（小写）→ "relationName.规范属性名"，用于跨类限定寻址。
//
// 关系按 v.Relations 的加载顺序（id 升序）遍历，保证解析结果确定。
func (v *Vocabulary) aggPropResolvers(cls *VocabClass) (map[string]string, map[string]string) {
	unq := v.propAliasMap(cls) // 主类 + 祖先
	qual := make(map[string]string)
	if cls == nil {
		return unq, qual
	}
	for _, rel := range v.Relations {
		if !strings.EqualFold(rel.From, cls.Name) {
			continue
		}
		target := v.Class(rel.To)
		if target == nil {
			continue
		}
		for alias, canonical := range v.propAliasMap(target) {
			qual[strings.ToLower(rel.Name)+"."+alias] = rel.Name + "." + canonical
			if _, exists := unq[alias]; !exists {
				unq[alias] = canonical
			}
		}
	}
	return unq, qual
}

// resolveAggProp 解析（可能带关系限定的）聚合属性名，返回规范形式与是否命中。
func resolveAggProp(prop string, unq, qual map[string]string) (string, bool) {
	p := strings.ToLower(strings.TrimSpace(prop))
	if p == "" {
		return "", false
	}
	if strings.Contains(p, ".") {
		c, ok := qual[p]
		return c, ok
	}
	c, ok := unq[p]
	return c, ok
}

// aggPropInfo 聚合上下文中一个规范属性的元信息（数据类型 + 中文 label）。
type aggPropInfo struct {
	dataType string
	label    string
}

// aggPropInfoMaps 构建「规范属性名 → 元信息(DataType/Label)」解析器，与 aggPropResolvers 平行：
//   - unq：规范属性名(小写) → info，覆盖「主类+祖先」并集，再并入各关系目标类（主类优先）；
//   - qual："relationName.规范属性名"(小写) → info，用于跨类限定寻址。
//
// 关系按 v.Relations 加载顺序(id 升序)遍历，主类属性先于关系目标类登记，保证解析确定、主类优先。
func (v *Vocabulary) aggPropInfoMaps(cls *VocabClass) (map[string]aggPropInfo, map[string]aggPropInfo) {
	unq := make(map[string]aggPropInfo)
	qual := make(map[string]aggPropInfo)
	if cls == nil {
		return unq, qual
	}
	walk := func(c *VocabClass, fn func(p VocabProperty)) {
		seen := make(map[uint]bool)
		cur := c
		for cur != nil && !seen[cur.ID] {
			seen[cur.ID] = true
			for _, p := range cur.Properties {
				fn(p)
			}
			if cur.ParentID == nil {
				break
			}
			cur = v.classByID(*cur.ParentID)
		}
	}
	// 主类 + 祖先（主类优先，先登记者不被覆盖）。
	walk(cls, func(p VocabProperty) {
		k := strings.ToLower(p.Name)
		if _, ex := unq[k]; !ex {
			unq[k] = aggPropInfo{dataType: p.DataType, label: p.Label}
		}
	})
	// 各关系目标类（含其祖先）。
	for _, rel := range v.Relations {
		if !strings.EqualFold(rel.From, cls.Name) {
			continue
		}
		target := v.Class(rel.To)
		if target == nil {
			continue
		}
		relName := rel.Name
		walk(target, func(p VocabProperty) {
			info := aggPropInfo{dataType: p.DataType, label: p.Label}
			qk := strings.ToLower(relName + "." + p.Name)
			if _, ex := qual[qk]; !ex {
				qual[qk] = info
			}
			uk := strings.ToLower(p.Name)
			if _, ex := unq[uk]; !ex {
				unq[uk] = info
			}
		})
	}
	return unq, qual
}

// lookupAggPropInfo 依据规范属性名（可能带 relation. 限定）查其元信息。
func lookupAggPropInfo(canonicalProp string, unq, qual map[string]aggPropInfo) (aggPropInfo, bool) {
	k := strings.ToLower(strings.TrimSpace(canonicalProp))
	if k == "" {
		return aggPropInfo{}, false
	}
	if strings.Contains(k, ".") {
		info, ok := qual[k]
		return info, ok
	}
	info, ok := unq[k]
	return info, ok
}

// coerceNumericValue 按目标属性 DataType 把「数字字符串」归一为 int64/float64：
//   - 非数值型属性、非字符串值、空串或无法解析 → 原样返回（ok=false，表示未改动）；
//   - 整型(integer/int/long)优先解析为 int64，其余数值型(float/double/decimal/number)解析为 float64；
//     整型串带小数(如 "50000.0")回退为 float64。
//
// 规则固定、可复现（确定性），仅作用于 llm 路径；rule 路径既有产出不变。
func coerceNumericValue(value interface{}, dataType string) (interface{}, bool) {
	if !isNumericDataType(dataType) {
		return value, false
	}
	s, isStr := value.(string)
	if !isStr {
		return value, false // 已是数字或其它类型，无需处理
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return value, false
	}
	switch strings.ToLower(strings.TrimSpace(dataType)) {
	case "integer", "int", "long":
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i, true
		}
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, true
	}
	return value, false
}

// splitListValue 把 in/not_in 的过滤值拆为元素切片（字符串按英文逗号分隔并 trim；切片原样展开）。
func splitListValue(value interface{}) []interface{} {
	switch val := value.(type) {
	case []interface{}:
		return val
	case []string:
		out := make([]interface{}, len(val))
		for i := range val {
			out[i] = val[i]
		}
		return out
	case string:
		parts := strings.Split(val, ",")
		out := make([]interface{}, 0, len(parts))
		for _, p := range parts {
			if p = strings.TrimSpace(p); p != "" {
				out = append(out, p)
			}
		}
		return out
	}
	return nil
}

// normalizeFilterValue 按操作符与目标属性 DataType 归一过滤值：
//   - 标量比较(eq/neq/gt/gte/lt/lte 等)：数字字符串 → 数字；
//   - in/not_in：逐元素归一，全部成功才替换为 []interface{}（含数字），否则原样返回；
//   - like/is_null/is_not_null 等：原样返回。
func normalizeFilterValue(op string, value interface{}, dataType string) interface{} {
	if !isNumericDataType(dataType) {
		return value
	}
	switch strings.ToLower(strings.TrimSpace(op)) {
	case "is_null", "is_not_null", "not_null", "like":
		return value
	case "in", "not_in", "nin":
		elems := splitListValue(value)
		if len(elems) == 0 {
			return value
		}
		out := make([]interface{}, 0, len(elems))
		changed := false
		for _, e := range elems {
			if n, ok := coerceNumericValue(e, dataType); ok {
				out = append(out, n)
				changed = true
			} else {
				out = append(out, e)
			}
		}
		if changed {
			return out
		}
		return value
	default:
		if n, ok := coerceNumericValue(value, dataType); ok {
			return n
		}
		return value
	}
}

// isAggAlias 判断 order_by 值是否引用了某个 aggregate 的显式别名。
func isAggAlias(ob string, aggs []query.Aggregate) bool {
	ob = strings.TrimSpace(ob)
	if ob == "" {
		return false
	}
	for _, a := range aggs {
		if alias := strings.TrimSpace(a.Alias); alias != "" && strings.EqualFold(alias, ob) {
			return true
		}
	}
	return false
}

// extractToolArgs 从模型输出中取出 OntologyIntent 工具调用的参数 JSON 串。
func extractToolArgs(msg *schema.Message) string {
	if msg == nil {
		return ""
	}
	// 优先匹配指定工具名，否则取第一个带参数的工具调用。
	fallback := ""
	for _, tc := range msg.ToolCalls {
		args := strings.TrimSpace(tc.Function.Arguments)
		if args == "" {
			continue
		}
		if strings.EqualFold(tc.Function.Name, intentToolName) {
			return args
		}
		if fallback == "" {
			fallback = args
		}
	}
	return fallback
}

// extractAndValidate 解析工具参数为 QueryRequest 并做词表校验，返回请求与硬性问题列表。
func extractAndValidate(vocab *Vocabulary, ontologyID uint, msg *schema.Message) (*query.QueryRequest, []string) {
	args := extractToolArgs(msg)
	if args == "" {
		return nil, []string{"模型未以 OntologyIntent 工具调用形式输出意图"}
	}
	var req query.QueryRequest
	if err := json.Unmarshal([]byte(args), &req); err != nil {
		return nil, []string{"意图参数 JSON 解析失败: " + err.Error()}
	}
	req.OntologyID = ontologyID
	issues := validateRequest(vocab, &req)
	return &req, issues
}

// ValidateRequest 是 validateRequest 的对外包装，供技能结构化入参通道（skills 包）复用：
// 当 Agent 直接注入确定性意图（class_name/filters/... 绕过 nlparse 自由文本抽取）时，
// 用同一套词表归一与校验逻辑把关，绝不把非法 AST 交给 translator。
// 返回无法自动修复的硬性问题列表；issues 为空表示可安全进入 L1 确定性翻译。
func ValidateRequest(vocab *Vocabulary, req *query.QueryRequest) []string {
	return validateRequest(vocab, req)
}

// validateRequest 就地规范化 req（trim/别名归一/操作符小写/limit 夹取），并返回无法自动修复的硬性问题。
// 绝不把非法 AST 交给 translator：只要 issues 非空，调用方必须修复或放弃。
func validateRequest(vocab *Vocabulary, req *query.QueryRequest) []string {
	issues := make([]string, 0)
	if req == nil {
		return []string{"意图为空"}
	}

	req.ClassName = strings.TrimSpace(req.ClassName)
	if req.ClassName == "" {
		issues = append(issues, "class_name 为空，必须指定一个本体类")
		return issues
	}
	cls := vocab.Class(req.ClassName)
	if cls == nil {
		issues = append(issues, fmt.Sprintf("class_name %q 不在本体类词表内", req.ClassName))
		return issues
	}
	req.ClassName = cls.Name // 归一为规范大小写

	alias := vocab.propAliasMap(cls)

	// properties 归一 + 校验。
	if len(req.Properties) > 0 {
		normalized := make([]string, 0, len(req.Properties))
		for _, p := range req.Properties {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if canonical, ok := alias[strings.ToLower(p)]; ok {
				normalized = append(normalized, canonical)
			} else {
				issues = append(issues, fmt.Sprintf("属性 %q 不属于类 %s 的词表", p, cls.Name))
			}
		}
		req.Properties = normalized
	}

	// relations 归一 + 校验。
	if len(req.Relations) > 0 {
		normalized := make([]string, 0, len(req.Relations))
		for _, r := range req.Relations {
			r = strings.TrimSpace(r)
			if r == "" {
				continue
			}
			if rel := vocab.Relation(r); rel != nil {
				normalized = append(normalized, rel.Name)
			} else {
				issues = append(issues, fmt.Sprintf("关系 %q 不在关系词表内", r))
			}
		}
		req.Relations = normalized
	}

	// 属性元信息（DataType/Label）解析器：主类+祖先 ∪ 关系目标类（支持 relationName.property 限定）。
	// 供 filter/having 值按 DataType 数值归一、以及聚合缺省别名取中文 label 复用。
	aggInfoUnq, aggInfoQual := vocab.aggPropInfoMaps(cls)

	// filters 归一 + 校验（属性别名归一 + 操作符白名单 + 值按 DataType 数值归一）。
	if len(req.Filters) > 0 {
		normalized := make([]query.Filter, 0, len(req.Filters))
		for _, f := range req.Filters {
			prop := strings.TrimSpace(f.Property)
			op := strings.ToLower(strings.TrimSpace(f.Op))
			canonical, ok := alias[strings.ToLower(prop)]
			if !ok {
				issues = append(issues, fmt.Sprintf("过滤属性 %q 不属于类 %s 的词表", prop, cls.Name))
				continue
			}
			if !isSupportedOp(op) {
				issues = append(issues, fmt.Sprintf("过滤操作符 %q 不在实现集内", f.Op))
				continue
			}
			// 值类型归一：数值型属性把数字字符串转为 int64/float64，规避强类型方言(PG)类型不匹配。
			val := f.Value
			if info, iok := lookupAggPropInfo(canonical, aggInfoUnq, aggInfoQual); iok {
				val = normalizeFilterValue(op, f.Value, info.dataType)
			}
			normalized = append(normalized, query.Filter{Property: canonical, Op: op, Value: val})
		}
		req.Filters = normalized
	}

	// 聚合上下文的属性解析器：主类+祖先 ∪ 关系目标类（支持 relationName.property 限定）。
	aggUnq, aggQual := vocab.aggPropResolvers(cls)
	// aliasDataType：聚合别名(小写) → 结果值 DataType，供 having.value 数值归一（count → integer）。
	aliasDataType := make(map[string]string)

	// aggregates 归一 + 校验（func 白名单、property 经扩展别名归一、count 外必须带 property）。
	// 缺省/空 alias 归一到与 rule 路径同一中文口径（query.DefaultAggAlias），保证两路径表头/图表 label 一致；
	// LLM 显式提供的非空 alias 保留（仍做重复校验），唯一化去重交由 translator.uniqueAlias 在翻译期统一处理。
	if len(req.Aggregates) > 0 {
		normalized := make([]query.Aggregate, 0, len(req.Aggregates))
		seenAlias := make(map[string]bool)
		for _, a := range req.Aggregates {
			fn := strings.ToLower(strings.TrimSpace(a.Func))
			if !isSupportedAggFunc(fn) {
				issues = append(issues, fmt.Sprintf("聚合函数 %q 不在实现集内", a.Func))
				continue
			}
			prop := strings.TrimSpace(a.Property)
			canonicalProp := ""
			propLabel := ""
			resultDT := "integer" // count(*) 结果为整型
			if prop != "" {
				c, ok := resolveAggProp(prop, aggUnq, aggQual)
				if !ok {
					issues = append(issues, fmt.Sprintf("聚合属性 %q 不属于类 %s 及其关系目标类的词表", prop, cls.Name))
					continue
				}
				canonicalProp = c
				if info, iok := lookupAggPropInfo(canonicalProp, aggInfoUnq, aggInfoQual); iok {
					propLabel = info.label
					resultDT = info.dataType
				}
			} else if fn != "count" {
				issues = append(issues, fmt.Sprintf("聚合函数 %q 必须指定 property（仅 count 可空表示 COUNT(*)）", fn))
				continue
			}
			al := strings.TrimSpace(a.Alias)
			if al == "" {
				// 缺省别名归一为中文（聚合函数中文 + 属性中文 label），与 rule 路径 / translator 同口径。
				al = query.DefaultAggAlias(fn, propLabel)
			} else {
				lk := strings.ToLower(al)
				if seenAlias[lk] {
					issues = append(issues, fmt.Sprintf("聚合别名 %q 重复", al))
					continue
				}
				seenAlias[lk] = true
			}
			aliasDataType[strings.ToLower(al)] = resultDT
			normalized = append(normalized, query.Aggregate{Func: fn, Property: canonicalProp, Alias: al})
		}
		req.Aggregates = normalized
	}

	// group_by 归一 + 校验（支持 relationName.property 限定）。
	if len(req.GroupBy) > 0 {
		normalized := make([]string, 0, len(req.GroupBy))
		for _, g := range req.GroupBy {
			g = strings.TrimSpace(g)
			if g == "" {
				continue
			}
			if c, ok := resolveAggProp(g, aggUnq, aggQual); ok {
				normalized = append(normalized, c)
			} else {
				issues = append(issues, fmt.Sprintf("分组属性 %q 不属于类 %s 及其关系目标类的词表", g, cls.Name))
			}
		}
		req.GroupBy = normalized
	}

	// having 归一 + 校验（alias 必填、op 走 isSupportedOp）。
	if len(req.Having) > 0 {
		normalized := make([]query.HavingClause, 0, len(req.Having))
		for _, h := range req.Having {
			al := strings.TrimSpace(h.Alias)
			op := strings.ToLower(strings.TrimSpace(h.Op))
			if al == "" {
				issues = append(issues, "having.alias 不能为空")
				continue
			}
			if !isSupportedOp(op) {
				issues = append(issues, fmt.Sprintf("having 操作符 %q 不在实现集内", h.Op))
				continue
			}
			// having.value 同样按引用聚合的结果 DataType 做数值归一（count → integer，其余 → 属性类型）。
			val := h.Value
			if dt, dok := aliasDataType[strings.ToLower(al)]; dok {
				val = normalizeFilterValue(op, h.Value, dt)
			}
			normalized = append(normalized, query.HavingClause{Alias: al, Op: op, Value: val})
		}
		req.Having = normalized
	}

	// order_by / order_dir 归一 + 校验。
	// 聚合上下文：order_by 允许填 aggregate 别名（TOP-N）；否则按属性（含关系限定）解析。
	req.OrderBy = strings.TrimSpace(req.OrderBy)
	if req.OrderBy != "" && !isAggAlias(req.OrderBy, req.Aggregates) {
		if canonical, ok := resolveAggProp(req.OrderBy, aggUnq, aggQual); ok {
			req.OrderBy = canonical
		} else if canonical, ok := alias[strings.ToLower(req.OrderBy)]; ok {
			req.OrderBy = canonical
		} else {
			issues = append(issues, fmt.Sprintf("排序字段 %q 不属于类 %s 的词表", req.OrderBy, cls.Name))
			req.OrderBy = ""
		}
	}
	req.OrderDir = strings.ToLower(strings.TrimSpace(req.OrderDir))
	if req.OrderDir != "" && req.OrderDir != "asc" && req.OrderDir != "desc" {
		issues = append(issues, fmt.Sprintf("排序方向 %q 非法，仅可为 asc/desc", req.OrderDir))
		req.OrderDir = ""
	}

	// limit / offset 夹取。
	if req.Limit < 0 {
		issues = append(issues, "limit 不能为负")
		req.Limit = 0
	}
	if req.Limit > limitMax {
		issues = append(issues, fmt.Sprintf("limit %d 超出上限 %d", req.Limit, limitMax))
		req.Limit = limitMax
	}
	if req.Offset < 0 {
		issues = append(issues, "offset 不能为负")
		req.Offset = 0
	}

	// 空切片规范化：LLM 对「无值」字段有时省略(null)、有时给空数组([])，二者语义相同但 JSON 形态不同。
	// 统一归一为 nil，保证同一 NL 连跑两次 intent JSON 逐字节一致，且与 rule 路径(空字段输出 null)同形。
	// 不影响翻译结果（Translate 对 nil 与空切片处理一致），故 SQL 仍逐字节相同。
	if len(req.Properties) == 0 {
		req.Properties = nil
	}
	if len(req.Filters) == 0 {
		req.Filters = nil
	}
	if len(req.Relations) == 0 {
		req.Relations = nil
	}
	if len(req.Aggregates) == 0 {
		req.Aggregates = nil
	}
	if len(req.GroupBy) == 0 {
		req.GroupBy = nil
	}
	if len(req.Having) == 0 {
		req.Having = nil
	}

	return issues
}

// addUsage 累加一次模型响应的 token 用量到 LLMMeta。
func addUsage(meta *LLMMeta, msg *schema.Message) {
	if meta == nil || msg == nil || msg.ResponseMeta == nil || msg.ResponseMeta.Usage == nil {
		return
	}
	u := msg.ResponseMeta.Usage
	meta.PromptTokens += u.PromptTokens
	meta.CompletionTokens += u.CompletionTokens
	meta.TotalTokens += u.TotalTokens
}
