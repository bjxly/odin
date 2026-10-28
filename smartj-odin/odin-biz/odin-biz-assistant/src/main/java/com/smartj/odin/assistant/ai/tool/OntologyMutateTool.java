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
import com.smartj.odin.assistant.entity.AiDraft;
import com.smartj.odin.assistant.entity.AiDraftItem;
import com.smartj.odin.assistant.mapper.AiDraftItemMapper;
import com.smartj.odin.assistant.mapper.AiDraftMapper;
import com.smartj.odin.common.udap.JdbcUrlBuilder;
import com.smartj.odin.common.udap.model.ConnectorConfig;
import com.smartj.odin.datasource.service.DatasourceService;
import com.smartj.odin.datasource.vo.DatasourceVO;
import com.smartj.odin.ontology.service.OntologyService;
import com.smartj.odin.ontology.vo.VocabularyVO;
import io.agentscope.core.tool.Tool;
import io.agentscope.core.tool.ToolParam;
import lombok.extern.slf4j.Slf4j;
import org.springframework.util.StringUtils;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;

/**
 * 本体写操作工具（HITL 草稿模式）。
 *
 * <p>把 AI 的建模/映射/数据源建议写入 {@code odin_ai_draft} / {@code odin_ai_draft_item} 草稿，
 * <b>绝不</b>直接修改本体、映射、数据源业务表；人工在草稿审阅面板逐项批准后，
 * 由 {@code AiDraftService.publishDraft} 在单事务内落地为正式资源。
 *
 * <p>非单例：每次对话请求按当前本体新建实例并注册进 {@code Toolkit}，
 * 实例内收集本轮产生的草稿ID，供 SSE {@code draft} 事件回传前端刷新未读角标。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Slf4j
public class OntologyMutateTool {

    /** 快照 JSON 序列化器 */
    private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();

    /** 列映射 JSON 反序列化类型 */
    private static final TypeReference<List<Map<String, Object>>> COLUMN_MAPPINGS_TYPE = new TypeReference<>() {
    };

    /** 草稿状态：待审阅 */
    private static final String STATUS_DRAFT = "draft";
    /** 明细项审阅状态：待审阅 */
    private static final String REVIEW_PENDING = "pending";
    /** 变更动作：新增 */
    private static final String ACTION_ADD = "add";

    private final AiDraftMapper draftMapper;
    private final AiDraftItemMapper itemMapper;
    private final OntologyService ontologyService;
    private final DatasourceService datasourceService;
    private final Long ontologyId;
    private final String ontologyName;

    /** 本轮对话产生的草稿ID（按 draftType 分组） */
    private final List<Long> producedDraftIds = new ArrayList<>();
    /** draftType → 草稿ID（同轮同类型建议合并进同一草稿） */
    private final Map<String, Long> draftIdByType = new LinkedHashMap<>();
    /** draftType → 已累积明细项数 */
    private final Map<String, Integer> itemCountByType = new LinkedHashMap<>();

    /**
     * 构造写工具实例。
     *
     * @param draftMapper       草稿头 Mapper
     * @param itemMapper        草稿明细 Mapper
     * @param ontologyService   本体服务（读词表做校验）
     * @param datasourceService 数据源服务（校验数据源存在性）
     * @param ontologyId        当前对话绑定的本体ID
     * @param ontologyName      当前本体名称（用于草稿摘要）
     */
    public OntologyMutateTool(AiDraftMapper draftMapper, AiDraftItemMapper itemMapper,
                              OntologyService ontologyService, DatasourceService datasourceService,
                              Long ontologyId, String ontologyName) {
        this.draftMapper = draftMapper;
        this.itemMapper = itemMapper;
        this.ontologyService = ontologyService;
        this.datasourceService = datasourceService;
        this.ontologyId = ontologyId;
        this.ontologyName = ontologyName;
    }

    /**
     * 本轮产生的草稿ID列表（供 SSE draft 事件回传）。
     *
     * @return 草稿ID列表（不可变副本）
     */
    public List<Long> getProducedDraftIds() {
        return List.copyOf(producedDraftIds);
    }

    /**
     * 本轮产生的草稿明细项总数。
     *
     * @return 明细项数量
     */
    public int getProducedItemCount() {
        return itemCountByType.values().stream().mapToInt(Integer::intValue).sum();
    }

    /**
     * 提议新建本体类（生成待审阅草稿）。
     *
     * @param classCode   类编码
     * @param className   类名称
     * @param description 类描述（可空）
     * @return 结果说明文本
     */
    @Tool(name = "propose_create_class",
            description = "提议在当前本体中新建一个类。该操作只生成待人工审阅的草稿，不会立即生效；"
                    + "调用成功后请告知用户已生成草稿、需在草稿审阅面板批准后发布。")
    public String proposeCreateClass(
            @ToolParam(name = "class_code", required = true, description = "类编码，英文标识，如 Order") String classCode,
            @ToolParam(name = "class_name", required = true, description = "类名称，中文可读名，如 订单") String className,
            @ToolParam(name = "description", required = false, description = "类的用途描述") String description) {
        if (ontologyId == null) {
            return "当前对话未绑定本体，无法提议建模变更";
        }
        if (!StringUtils.hasText(classCode) || !StringUtils.hasText(className)) {
            return "类编码与类名称不能为空";
        }
        VocabularyVO vocab = loadVocabulary();
        if (vocab == null) {
            return "本体词表加载失败，无法校验类编码";
        }
        String code = classCode.trim();
        if (findClass(vocab, code) != null) {
            return "类编码已存在：" + code;
        }
        Map<String, Object> after = new LinkedHashMap<>();
        after.put("ontologyId", String.valueOf(ontologyId));
        after.put("code", code);
        after.put("name", className.trim());
        after.put("description", description);
        Long draftId = draftOf("ontology", "AI 提议的本体建模变更");
        addItem(draftId, "ontology", "class", ACTION_ADD, after);
        return "已生成草稿#" + draftId + "：新建类 " + code + "（" + className.trim() + "），待人工审阅后发布。";
    }

    /**
     * 提议为指定类新建属性（生成待审阅草稿）。
     *
     * @param classCode    所属类编码
     * @param propertyCode 属性编码
     * @param propertyName 属性名称
     * @param dataType     数据类型（string/int/decimal/date/boolean 等）
     * @return 结果说明文本
     */
    @Tool(name = "propose_create_property",
            description = "提议为当前本体的某个类新增一个属性。只生成待人工审阅的草稿，不会立即生效。")
    public String proposeCreateProperty(
            @ToolParam(name = "class_code", required = true, description = "所属类编码，必须是本体词表中已存在的类") String classCode,
            @ToolParam(name = "property_code", required = true, description = "属性编码，英文标识，如 amount") String propertyCode,
            @ToolParam(name = "property_name", required = true, description = "属性名称，如 金额") String propertyName,
            @ToolParam(name = "data_type", required = false, description = "数据类型，默认 string") String dataType) {
        if (ontologyId == null) {
            return "当前对话未绑定本体，无法提议建模变更";
        }
        if (!StringUtils.hasText(classCode) || !StringUtils.hasText(propertyCode) || !StringUtils.hasText(propertyName)) {
            return "类编码、属性编码与属性名称不能为空";
        }
        VocabularyVO vocab = loadVocabulary();
        if (vocab == null) {
            return "本体词表加载失败，无法校验类与属性";
        }
        VocabularyVO.ClassVocab clazz = findClass(vocab, classCode.trim());
        if (clazz == null) {
            return "本体中不存在类：" + classCode + "；可用类：" + String.join(", ", classCodes(vocab));
        }
        String propCode = propertyCode.trim();
        if (clazz.getProperties() != null) {
            for (VocabularyVO.PropertyVocab p : clazz.getProperties()) {
                if (propCode.equals(p.getCode())) {
                    return "类 " + clazz.getCode() + " 下属性已存在：" + propCode;
                }
            }
        }
        Map<String, Object> after = new LinkedHashMap<>();
        after.put("ontologyId", String.valueOf(ontologyId));
        after.put("classCode", clazz.getCode());
        after.put("code", propCode);
        after.put("name", propertyName.trim());
        after.put("dataType", StringUtils.hasText(dataType) ? dataType.trim() : "string");
        Long draftId = draftOf("ontology", "AI 提议的本体建模变更");
        addItem(draftId, "ontology", "property", ACTION_ADD, after);
        return "已生成草稿#" + draftId + "：为类 " + clazz.getCode() + " 新增属性 " + propCode + "，待人工审阅后发布。";
    }

    /**
     * 提议新建「本体类 → 物理表」映射（生成待审阅草稿）。
     *
     * @param classCode          本体类编码
     * @param datasourceId       数据源ID
     * @param tableName          物理表名
     * @param columnMappingsJson 列映射 JSON 数组字符串
     * @return 结果说明文本
     */
    @Tool(name = "propose_create_mapping",
            description = "提议为本体类建立到物理表的映射（含列映射）。只生成待人工审阅的草稿，不会立即生效。"
                    + "column_mappings 为 JSON 数组，每项形如 {\"propertyCode\":\"name\",\"columnName\":\"cust_name\"}。")
    public String proposeCreateMapping(
            @ToolParam(name = "class_code", required = true, description = "本体类编码") String classCode,
            @ToolParam(name = "datasource_id", required = true, description = "目标数据源ID（数字字符串）") String datasourceId,
            @ToolParam(name = "table_name", required = true, description = "物理表名，如 t_customer") String tableName,
            @ToolParam(name = "column_mappings", required = true,
                    description = "列映射 JSON 数组，如 [{\"propertyCode\":\"name\",\"columnName\":\"cust_name\"}]") String columnMappingsJson) {
        if (ontologyId == null) {
            return "当前对话未绑定本体，无法提议映射";
        }
        VocabularyVO vocab = loadVocabulary();
        if (vocab == null) {
            return "本体词表加载失败，无法校验映射";
        }
        VocabularyVO.ClassVocab clazz = StringUtils.hasText(classCode) ? findClass(vocab, classCode.trim()) : null;
        if (clazz == null) {
            return "本体中不存在类：" + classCode + "；可用类：" + String.join(", ", classCodes(vocab));
        }
        if (!StringUtils.hasText(tableName)) {
            return "物理表名不能为空";
        }
        // 先校验列映射（纯词表校验，避免无谓的数据库访问）
        List<Map<String, Object>> columns = parseColumnMappings(columnMappingsJson);
        if (columns.isEmpty()) {
            return "列映射不能为空，且必须是 JSON 数组，每项含 propertyCode 与 columnName";
        }
        Set<String> props = new LinkedHashSet<>();
        if (clazz.getProperties() != null) {
            for (VocabularyVO.PropertyVocab p : clazz.getProperties()) {
                props.add(p.getCode());
            }
        }
        for (Map<String, Object> col : columns) {
            String propCode = str(col.get("propertyCode"));
            if (!StringUtils.hasText(str(col.get("columnName")))) {
                return "列映射缺少 columnName：" + columnMappingsJson;
            }
            if (!props.contains(propCode)) {
                return "类 " + clazz.getCode() + " 下属性不存在：" + propCode + "；可用属性：" + String.join(", ", props);
            }
        }
        Long dsId = parseLong(datasourceId);
        if (dsId == null) {
            return "数据源ID非法：" + datasourceId;
        }
        DatasourceVO ds = safeGetDatasource(dsId);
        if (ds == null) {
            return "数据源不存在：" + dsId;
        }
        Map<String, Object> after = new LinkedHashMap<>();
        after.put("ontologyId", String.valueOf(ontologyId));
        after.put("classCode", clazz.getCode());
        after.put("datasourceId", String.valueOf(dsId));
        after.put("tableName", tableName.trim());
        after.put("columnMappings", columns);
        Long draftId = draftOf("mapping", "AI 提议的类-物理表映射");
        addItem(draftId, "mapping", "mapping", ACTION_ADD, after);
        return "已生成草稿#" + draftId + "：类 " + clazz.getCode() + " → " + tableName.trim()
                + "（" + columns.size() + " 个列映射），待人工审阅后发布。";
    }

    /**
     * 提议新建数据源（生成待审阅草稿，不含密码）。
     *
     * @param code       数据源编码
     * @param name       数据源名称
     * @param type       数据源类型
     * @param host       主机
     * @param port       端口
     * @param dbName     数据库名
     * @param schemaName 模式名（可空）
     * @param username   用户名
     * @return 结果说明文本
     */
    @Tool(name = "propose_create_datasource",
            description = "提议新建一个数据源连接。出于安全考虑不收集密码，只生成待人工审阅的草稿；"
                    + "发布后需人工在数据源管理页补录密码才能测试连接。支持的类型：mysql/postgresql/oracle/sqlserver/dameng(dm)/kingbase/gaussdb(opengauss)/sqlite/duckdb。")
    public String proposeCreateDatasource(
            @ToolParam(name = "code", required = true, description = "数据源编码，英文唯一标识，如 ds_erp") String code,
            @ToolParam(name = "name", required = true, description = "数据源名称，如 ERP生产库") String name,
            @ToolParam(name = "type", required = true, description = "数据源类型，如 mysql") String type,
            @ToolParam(name = "host", required = true, description = "主机地址；sqlite/duckdb 可填 local") String host,
            @ToolParam(name = "port", required = false, description = "端口，如 3306") String port,
            @ToolParam(name = "db_name", required = true, description = "数据库名；sqlite/duckdb 填文件路径") String dbName,
            @ToolParam(name = "schema_name", required = false, description = "模式名（PostgreSQL/达梦等可选）") String schemaName,
            @ToolParam(name = "username", required = false, description = "连接用户名") String username) {
        if (!StringUtils.hasText(code) || !StringUtils.hasText(name) || !StringUtils.hasText(type)) {
            return "数据源编码、名称与类型不能为空";
        }
        if (!StringUtils.hasText(dbName)) {
            return "数据库名不能为空";
        }
        Integer portNum = parsePort(port);
        // 复用 JdbcUrlBuilder 判定类型是否受支持（单一事实来源，避免重复维护类型清单）
        ConnectorConfig probe = new ConnectorConfig();
        probe.setType(type.trim());
        probe.setHost(StringUtils.hasText(host) ? host.trim() : "local");
        probe.setPort(portNum);
        probe.setDbName(dbName.trim());
        try {
            JdbcUrlBuilder.build(probe);
        } catch (IllegalArgumentException e) {
            return "不支持的数据源类型：" + type;
        }
        Map<String, Object> after = new LinkedHashMap<>();
        after.put("code", code.trim());
        after.put("name", name.trim());
        after.put("type", type.trim());
        after.put("host", probe.getHost());
        after.put("port", portNum);
        after.put("dbName", dbName.trim());
        after.put("schemaName", StringUtils.hasText(schemaName) ? schemaName.trim() : null);
        after.put("username", StringUtils.hasText(username) ? username.trim() : null);
        Long draftId = draftOf("datasource", "AI 提议的数据源接入");
        addItem(draftId, "datasource", "datasource", ACTION_ADD, after);
        return "已生成草稿#" + draftId + "：新建数据源 " + name.trim() + "（" + type.trim() + "）。"
                + "出于安全考虑草稿不含密码，发布后请到数据源管理页补录密码再测试连接。";
    }

    /**
     * 读取当前本体词表，失败返回 null。
     *
     * @return 词表 VO 或 null
     */
    private VocabularyVO loadVocabulary() {
        if (ontologyId == null) {
            return null;
        }
        try {
            return ontologyService.getVocabulary(ontologyId);
        } catch (Exception e) {
            log.warn("加载本体词表失败 ontologyId={}", ontologyId, e);
            return null;
        }
    }

    /**
     * 按编码查找类词表条目。
     *
     * @param vocab 词表
     * @param code  类编码
     * @return 类词表条目，不存在返回 null
     */
    private VocabularyVO.ClassVocab findClass(VocabularyVO vocab, String code) {
        if (vocab == null || vocab.getClasses() == null) {
            return null;
        }
        for (VocabularyVO.ClassVocab c : vocab.getClasses()) {
            if (code.equals(c.getCode())) {
                return c;
            }
        }
        return null;
    }

    /**
     * 收集词表中的全部类编码（用于错误提示）。
     *
     * @param vocab 词表
     * @return 类编码集合
     */
    private Set<String> classCodes(VocabularyVO vocab) {
        Set<String> codes = new LinkedHashSet<>();
        if (vocab != null && vocab.getClasses() != null) {
            for (VocabularyVO.ClassVocab c : vocab.getClasses()) {
                codes.add(c.getCode());
            }
        }
        return codes;
    }

    /**
     * 解析列映射 JSON，非法或为空返回空列表。
     *
     * @param json 列映射 JSON 数组字符串
     * @return 列映射列表
     */
    private List<Map<String, Object>> parseColumnMappings(String json) {
        if (!StringUtils.hasText(json)) {
            return List.of();
        }
        try {
            List<Map<String, Object>> list = OBJECT_MAPPER.readValue(json.trim(), COLUMN_MAPPINGS_TYPE);
            return list == null ? List.of() : list;
        } catch (Exception e) {
            log.warn("列映射 JSON 解析失败: {}", json, e);
            return List.of();
        }
    }

    /**
     * 安全读取数据源（异常视为不存在）。
     *
     * @param id 数据源ID
     * @return 数据源 VO 或 null
     */
    private DatasourceVO safeGetDatasource(Long id) {
        try {
            return datasourceService.getById(id);
        } catch (Exception e) {
            log.warn("读取数据源失败 id={}", id, e);
            return null;
        }
    }

    /**
     * 获取或创建指定类型的草稿头（同轮同类型复用）。
     *
     * @param draftType 草稿类型
     * @param summary   草稿摘要
     * @return 草稿ID
     */
    private Long draftOf(String draftType, String summary) {
        Long existing = draftIdByType.get(draftType);
        if (existing != null) {
            return existing;
        }
        AiDraft draft = new AiDraft();
        draft.setDraftType(draftType);
        draft.setOntologyId(ontologyId);
        draft.setStatus(STATUS_DRAFT);
        draft.setCreatedByAi(1);
        draft.setSummary(summary + (StringUtils.hasText(ontologyName) ? "（本体：" + ontologyName + "）" : ""));
        draftMapper.insert(draft);
        Long id = draft.getId();
        draftIdByType.put(draftType, id);
        itemCountByType.put(draftType, 0);
        producedDraftIds.add(id);
        return id;
    }

    /**
     * 追加一条待审阅明细项。
     *
     * @param draftId   草稿ID
     * @param draftType 草稿类型（用于计数）
     * @param itemType  明细项类型
     * @param action    变更动作
     * @param after     变更后快照
     */
    private void addItem(Long draftId, String draftType, String itemType, String action, Map<String, Object> after) {
        AiDraftItem item = new AiDraftItem();
        item.setDraftId(draftId);
        item.setItemType(itemType);
        item.setAction(action);
        item.setReviewStatus(REVIEW_PENDING);
        item.setAfterSnapshot(toJson(after));
        itemMapper.insert(item);
        itemCountByType.merge(draftType, 1, Integer::sum);
    }

    /**
     * 快照 Map 序列化为 JSON 字符串。
     *
     * @param map 快照内容
     * @return JSON 字符串，失败返回空对象
     */
    private String toJson(Map<String, Object> map) {
        try {
            return OBJECT_MAPPER.writeValueAsString(map);
        } catch (Exception e) {
            log.warn("草稿快照序列化失败", e);
            return "{}";
        }
    }

    /**
     * 解析 Long。
     *
     * @param s 字符串
     * @return Long 或 null
     */
    private Long parseLong(String s) {
        try {
            return StringUtils.hasText(s) ? Long.valueOf(s.trim()) : null;
        } catch (NumberFormatException e) {
            return null;
        }
    }

    /**
     * 解析端口。
     *
     * @param s 端口字符串
     * @return Integer 或 null
     */
    private Integer parsePort(String s) {
        Long v = parseLong(s);
        return v == null ? null : v.intValue();
    }

    /**
     * 取对象字符串值（null 安全）。
     *
     * @param o 对象
     * @return 字符串或 null
     */
    private String str(Object o) {
        return o == null ? null : String.valueOf(o);
    }
}
