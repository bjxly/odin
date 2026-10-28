/*
 * SmartJ 企业级多租户基础框架
 * Copyright (C) 2024-2026 SmartJ Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package com.smartj.odin.assistant.ai.tool;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.smartj.odin.common.udap.Connector;
import com.smartj.odin.common.udap.model.ConnectorConfig;
import com.smartj.odin.common.udap.model.QueryResult;
import com.smartj.odin.mapping.dto.MappingConfigQuery;
import com.smartj.odin.mapping.service.MappingService;
import com.smartj.odin.mapping.vo.MappingConfigVO;
import com.smartj.odin.ontology.service.OntologyService;
import com.smartj.odin.ontology.vo.OntClassVO;
import com.smartj.odin.ontology.vo.VocabularyVO;
import com.smartj.odin.datasource.service.DatasourceService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.LinkedHashMap;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.function.Function;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * 本体逻辑 SQL 执行器：词表校验 → 逻辑→物理编译 → 单源直查。
 *
 * <p>遵循「LLM 只写本体逻辑 SQL（仅类/属性编码）→ 封闭词表静态校验 → 映射编译为物理 SQL →
 * 连接器单源方言直查」路径。跨源（多数据源）查询不在本执行器范围（联邦执行待 P6），返回明确错误。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Slf4j
@Component
@RequiredArgsConstructor
public class OntologySqlExecutor {

    private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();

    /** SQL 关键字（校验标识符时忽略） */
    private static final Set<String> SQL_KEYWORDS = Set.of(
            "select", "from", "where", "group", "by", "order", "having", "limit", "offset",
            "join", "inner", "left", "right", "outer", "full", "cross", "on", "as", "and", "or",
            "not", "in", "is", "null", "desc", "asc", "count", "sum", "avg", "min", "max",
            "distinct", "between", "like", "case", "when", "then", "else", "end", "union", "all",
            "exists", "true", "false");

    /** 标识符 token */
    private static final Pattern IDENT = Pattern.compile("[A-Za-z_][A-Za-z0-9_]*");

    private final OntologyService ontologyService;
    private final MappingService mappingService;
    private final DatasourceService datasourceService;
    private final Connector connector;

    /**
     * 执行结果。
     */
    public static class ExecOutcome {
        private boolean success;
        private String error;
        private String physicalSql;
        private List<String> columns = new ArrayList<>();
        private List<Map<String, Object>> rows = new ArrayList<>();

        public static ExecOutcome fail(String error) {
            ExecOutcome o = new ExecOutcome();
            o.success = false;
            o.error = error;
            return o;
        }

        public static ExecOutcome ok(String physicalSql, QueryResult r) {
            ExecOutcome o = new ExecOutcome();
            o.success = true;
            o.physicalSql = physicalSql;
            o.columns = r.getColumns();
            o.rows = r.getRows();
            return o;
        }

        public boolean isSuccess() {
            return success;
        }

        public String getError() {
            return error;
        }

        public String getPhysicalSql() {
            return physicalSql;
        }

        public List<String> getColumns() {
            return columns;
        }

        public List<Map<String, Object>> getRows() {
            return rows;
        }
    }

    /** 单个类的物理映射信息 */
    private static class ClassMapping {
        Long datasourceId;
        String schemaName;
        String tableName;
        /** propertyCode → columnName */
        Map<String, String> propToCol = new HashMap<>();
    }

    /**
     * 校验并执行本体逻辑 SQL。
     *
     * @param ontologyId 本体ID
     * @param logicalSql 本体逻辑 SQL（仅类/属性编码）
     * @param maxRows    最大返回行数
     * @return 执行结果
     */
    public ExecOutcome execute(Long ontologyId, String logicalSql, int maxRows) {
        if (ontologyId == null) {
            return ExecOutcome.fail("未指定本体，无法确定语义上下文");
        }
        if (!StringUtils.hasText(logicalSql)) {
            return ExecOutcome.fail("逻辑SQL为空");
        }
        try {
            VocabularyVO vocab = ontologyService.getVocabulary(ontologyId);
            if (vocab == null || vocab.getClasses() == null || vocab.getClasses().isEmpty()) {
                return ExecOutcome.fail("本体词表为空，无法校验逻辑SQL");
            }
            // 类编码 → 类ID（用于关联映射）
            Map<String, Long> codeToClassId = new HashMap<>();
            flattenClasses(ontologyService.listClasses(ontologyId), codeToClassId);
            // 类编码 → 属性编码集合（词表）
            Map<String, Set<String>> classProps = new HashMap<>();
            for (VocabularyVO.ClassVocab c : vocab.getClasses()) {
                Set<String> props = new LinkedHashSet<>();
                if (c.getProperties() != null) {
                    for (VocabularyVO.PropertyVocab p : c.getProperties()) {
                        props.add(p.getCode());
                    }
                }
                classProps.put(c.getCode(), props);
            }
            // 类ID → 物理映射
            Map<Long, ClassMapping> mappings = loadMappings(ontologyId);

            // 1) 识别引用的类
            Set<String> referenced = new LinkedHashSet<>();
            for (String code : classProps.keySet()) {
                if (containsIdent(logicalSql, code)) {
                    referenced.add(code);
                }
            }
            if (referenced.isEmpty()) {
                return ExecOutcome.fail("未能从逻辑SQL中识别任何本体类，可用类：" + String.join(", ", classProps.keySet()));
            }
            // 2) 校验标识符（防幻觉）：非关键字、非类编码的标识符必须是引用类的属性
            Set<String> knownProps = new HashSet<>();
            for (String c : referenced) {
                knownProps.addAll(classProps.getOrDefault(c, Set.of()));
            }
            List<String> unknown = new ArrayList<>();
            for (String ident : extractIdents(logicalSql)) {
                String low = ident.toLowerCase();
                if (SQL_KEYWORDS.contains(low)) {
                    continue;
                }
                if (referenced.contains(ident) || classProps.containsKey(ident)) {
                    continue;
                }
                if (knownProps.contains(ident)) {
                    continue;
                }
                unknown.add(ident);
            }
            if (!unknown.isEmpty()) {
                return ExecOutcome.fail("逻辑SQL包含本体词表之外的标识符（疑似幻觉）：" + String.join(", ", unknown)
                        + "；可用属性：" + String.join(", ", knownProps));
            }
            // 3) 解析物理映射并校验单源
            Set<Long> dsIds = new LinkedHashSet<>();
            Map<String, ClassMapping> refMapping = new HashMap<>();
            for (String code : referenced) {
                Long classId = codeToClassId.get(code);
                ClassMapping cm = classId == null ? null : mappings.get(classId);
                if (cm == null) {
                    return ExecOutcome.fail("本体类 " + code + " 尚未建立物理表映射，无法执行查询");
                }
                refMapping.put(code, cm);
                dsIds.add(cm.datasourceId);
            }
            if (dsIds.size() > 1) {
                return ExecOutcome.fail("当前查询涉及 " + dsIds.size() + " 个数据源，跨源联邦执行待接入（P6），请缩小到单一数据源的类");
            }
            // 4) 编译逻辑→物理
            String physical = compile(logicalSql, referenced, refMapping, classProps);
            // 5) 单源直查
            Long dsId = dsIds.iterator().next();
            ConnectorConfig config = datasourceService.getConnectorConfig(dsId);
            QueryResult result = connector.query(config, physical, maxRows);
            return ExecOutcome.ok(physical, result);
        } catch (Exception e) {
            log.warn("本体逻辑SQL执行失败: ontologyId={}, sql={}, err={}", ontologyId, logicalSql, e.getMessage());
            return ExecOutcome.fail("执行失败：" + e.getMessage());
        }
    }

    /**
     * 编译：类编码→物理表、属性编码→物理列（仅改写引号外片段）。
     */
    private String compile(String sql, Set<String> referenced, Map<String, ClassMapping> refMapping,
                           Map<String, Set<String>> classProps) {
        // 属性编码 → 列（仅当在引用类中唯一时允许裸引用）
        Map<String, String> barePropToCol = new HashMap<>();
        Map<String, Integer> propOccur = new HashMap<>();
        for (String c : referenced) {
            ClassMapping cm = refMapping.get(c);
            for (String p : classProps.getOrDefault(c, Set.of())) {
                propOccur.merge(p, 1, Integer::sum);
                barePropToCol.put(p, cm.propToCol.get(p));
            }
        }
        String out = transformOutsideLiterals(sql, segment -> {
            String s = segment;
            // 4.1 点引用 Class.prop → table.column
            for (String c : referenced) {
                ClassMapping cm = refMapping.get(c);
                for (String p : classProps.getOrDefault(c, Set.of())) {
                    String col = cm.propToCol.get(p);
                    if (col == null) {
                        continue;
                    }
                    s = s.replaceAll("(?i)\\b" + Pattern.quote(c) + "\\." + Pattern.quote(p) + "\\b",
                            Matcher.quoteReplacement(qualify(cm) + "." + col));
                }
            }
            // 4.2 FROM/JOIN 类编码 → 物理表
            // 注意：$1 是对分组 (from|join 前缀) 的反向引用，必须保留为「活引用」；
            // 只对物理表名做 quoteReplacement，避免表名中的 $/\ 被误当作正则替换元字符。
            for (String c : referenced) {
                ClassMapping cm = refMapping.get(c);
                s = s.replaceAll("(?i)(\\b(?:from|join)\\s+)" + Pattern.quote(c) + "\\b",
                        "$1" + Matcher.quoteReplacement(qualify(cm)));
            }
            // 4.3 裸属性编码 → 列（唯一归属时）
            // 跳过与 SQL 关键字同名的属性编码（如属性 desc/count/order），避免误伤关键字；
            // 这类属性只能通过点引用 Class.prop（4.1）安全改写。
            for (Map.Entry<String, String> e : barePropToCol.entrySet()) {
                if (SQL_KEYWORDS.contains(e.getKey().toLowerCase())) {
                    continue;
                }
                if (propOccur.getOrDefault(e.getKey(), 0) == 1 && e.getValue() != null) {
                    s = s.replaceAll("(?i)\\b" + Pattern.quote(e.getKey()) + "\\b",
                            Matcher.quoteReplacement(e.getValue()));
                }
            }
            // 4.4 其余裸类编码（如 ON 条件）→ 物理表
            // 跳过与 SQL 关键字同名的类编码（如类 Order 与关键字 ORDER、类 Group 与 GROUP）：
            // (?i)\bOrder\b 会误把 ORDER BY 改写为 <table> BY 破坏语法。这些类的合法出现
            // 已由 4.1（点引用 Class.prop）与 4.2（FROM/JOIN 前缀守卫）覆盖，无需在此裸替换。
            for (String c : referenced) {
                if (SQL_KEYWORDS.contains(c.toLowerCase())) {
                    continue;
                }
                s = s.replaceAll("(?i)\\b" + Pattern.quote(c) + "\\b",
                        Matcher.quoteReplacement(qualify(refMapping.get(c))));
            }
            return s;
        });
        return out;
    }

    /** 物理表全限定名（schema.table 或 table） */
    private String qualify(ClassMapping cm) {
        return StringUtils.hasText(cm.schemaName) ? cm.schemaName + "." + cm.tableName : cm.tableName;
    }

    /** 仅对单引号字面量之外的片段应用改写 */
    private String transformOutsideLiterals(String sql, Function<String, String> fn) {
        StringBuilder sb = new StringBuilder();
        int i = 0;
        while (i < sql.length()) {
            int q = sql.indexOf('\'', i);
            if (q < 0) {
                sb.append(fn.apply(sql.substring(i)));
                break;
            }
            sb.append(fn.apply(sql.substring(i, q)));
            int end = sql.indexOf('\'', q + 1);
            if (end < 0) {
                sb.append(sql.substring(q));
                break;
            }
            sb.append(sql, q, end + 1);
            i = end + 1;
        }
        return sb.toString();
    }

    /** 提取引号外的标识符 token */
    private List<String> extractIdents(String sql) {
        List<String> idents = new ArrayList<>();
        String stripped = transformOutsideLiterals(sql, seg -> seg.replaceAll("'[^']*'", " "));
        Matcher m = IDENT.matcher(stripped);
        while (m.find()) {
            idents.add(m.group());
        }
        return idents;
    }

    /** 判断引号外是否包含某标识符（词边界） */
    private boolean containsIdent(String sql, String ident) {
        String stripped = transformOutsideLiterals(sql, seg -> seg);
        Matcher m = Pattern.compile("(?i)\\b" + Pattern.quote(ident) + "\\b").matcher(stripped);
        return m.find();
    }

    /** 加载本体下启用映射（类ID → 映射信息，取优先级最高者） */
    private Map<Long, ClassMapping> loadMappings(Long ontologyId) {
        MappingConfigQuery q = new MappingConfigQuery();
        q.setOntologyId(ontologyId);
        q.setCurrent(1);
        q.setSize(500);
        List<MappingConfigVO> records = mappingService.page(q).getRecords();
        Map<Long, ClassMapping> best = new HashMap<>();
        Map<Long, Integer> bestPriority = new HashMap<>();
        for (MappingConfigVO vo : records) {
            if (vo.getStatus() != null && vo.getStatus() == 0) {
                continue;
            }
            Long classId = parseLong(vo.getClassId());
            Long dsId = parseLong(vo.getDatasourceId());
            if (classId == null || dsId == null || !StringUtils.hasText(vo.getTableName())) {
                continue;
            }
            int priority = vo.getPriority() == null ? 0 : vo.getPriority();
            if (bestPriority.containsKey(classId) && bestPriority.get(classId) > priority) {
                continue;
            }
            ClassMapping cm = new ClassMapping();
            cm.datasourceId = dsId;
            cm.schemaName = vo.getSchemaName();
            cm.tableName = vo.getTableName();
            cm.propToCol = parseColumnMappings(vo.getColumnMappings());
            best.put(classId, cm);
            bestPriority.put(classId, priority);
        }
        return best;
    }

    /** 解析 column_mappings JSON（propertyCode → columnName） */
    private Map<String, String> parseColumnMappings(String json) {
        Map<String, String> map = new HashMap<>();
        if (!StringUtils.hasText(json)) {
            return map;
        }
        try {
            List<Map<String, Object>> list = OBJECT_MAPPER.readValue(json,
                    new TypeReference<List<Map<String, Object>>>() {
                    });
            for (Map<String, Object> item : list) {
                Object pc = item.get("propertyCode");
                Object cn = item.get("columnName");
                if (pc == null) {
                    pc = item.get("property_code");
                }
                if (cn == null) {
                    cn = item.get("column_name");
                }
                if (pc != null && cn != null) {
                    map.put(String.valueOf(pc), String.valueOf(cn));
                }
            }
        } catch (Exception e) {
            log.warn("解析 column_mappings 失败: {}", e.getMessage());
        }
        return map;
    }

    /** 扁平化类树，填充 code → classId */
    private void flattenClasses(List<OntClassVO> classes, Map<String, Long> codeToClassId) {
        if (classes == null) {
            return;
        }
        for (OntClassVO c : classes) {
            Long id = parseLong(c.getId());
            if (id != null && StringUtils.hasText(c.getCode())) {
                codeToClassId.put(c.getCode(), id);
            }
            flattenClasses(c.getChildren(), codeToClassId);
        }
    }

    private Long parseLong(String s) {
        if (!StringUtils.hasText(s)) {
            return null;
        }
        try {
            return Long.valueOf(s.trim());
        } catch (NumberFormatException e) {
            return null;
        }
    }
}
