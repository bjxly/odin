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
package com.smartj.odin.assistant.service.impl;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.mybatisflex.core.paginate.Page;
import com.mybatisflex.core.query.QueryWrapper;
import com.smartj.common.core.domain.PageResult;
import com.smartj.common.core.exception.BusinessException;
import com.smartj.odin.assistant.converter.AssistantConverter;
import com.smartj.odin.assistant.dto.DraftQuery;
import com.smartj.odin.assistant.dto.DraftReviewDTO;
import com.smartj.odin.assistant.entity.AiDraft;
import com.smartj.odin.assistant.entity.AiDraftItem;
import com.smartj.odin.assistant.error.AssistantErrorCode;
import com.smartj.odin.assistant.mapper.AiDraftItemMapper;
import com.smartj.odin.assistant.mapper.AiDraftMapper;
import com.smartj.odin.assistant.service.AiDraftService;
import com.smartj.odin.assistant.vo.AiDraftItemVO;
import com.smartj.odin.assistant.vo.AiDraftVO;
import com.smartj.odin.datasource.dto.DatasourceDTO;
import com.smartj.odin.datasource.service.DatasourceService;
import com.smartj.odin.mapping.dto.MappingConfigDTO;
import com.smartj.odin.mapping.service.MappingService;
import com.smartj.odin.ontology.dto.OntClassDTO;
import com.smartj.odin.ontology.dto.OntPropertyDTO;
import com.smartj.odin.ontology.service.OntologyService;
import com.smartj.odin.ontology.vo.OntClassVO;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import static com.smartj.odin.assistant.entity.table.AiDraftItemTableDef.AI_DRAFT_ITEM;
import static com.smartj.odin.assistant.entity.table.AiDraftTableDef.AI_DRAFT;

/**
 * AI 草稿服务实现（HITL 审阅与发布）。
 *
 * <p>提供草稿分页、详情（含明细项）、逐项审阅与发布/驳回。发布时在单事务内把已批准明细项
 * 落地为正式资源（本体类 / 属性 / 类-物理表映射 / 数据源），失败整体回滚。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Service
@RequiredArgsConstructor
public class AiDraftServiceImpl implements AiDraftService {

    /** 明细项待审阅状态 */
    private static final String ITEM_PENDING = "pending";
    /** 明细项已批准状态 */
    private static final String ITEM_APPROVED = "approved";
    /** 草稿已发布状态 */
    private static final String DRAFT_PUBLISHED = "published";
    /** 草稿已驳回状态 */
    private static final String DRAFT_REJECTED = "rejected";
    /** 快照 JSON 解析器 */
    private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();
    /** 快照 Map 类型 */
    private static final TypeReference<Map<String, Object>> SNAPSHOT_TYPE = new TypeReference<>() {
    };

    private final AiDraftMapper aiDraftMapper;
    private final AiDraftItemMapper aiDraftItemMapper;
    private final OntologyService ontologyService;
    private final MappingService mappingService;
    private final DatasourceService datasourceService;

    /** 分页查询 AI 草稿，支持类型/本体/状态筛选 */
    @Override
    @Transactional(readOnly = true)
    public PageResult<AiDraftVO> draftPage(DraftQuery query) {
        QueryWrapper qw = QueryWrapper.create()
                .where(AI_DRAFT.DRAFT_TYPE.eq(query.getDraftType()).when(StringUtils.hasText(query.getDraftType())))
                .and(AI_DRAFT.ONTOLOGY_ID.eq(query.getOntologyId()).when(query.getOntologyId() != null))
                .and(AI_DRAFT.STATUS.eq(query.getStatus()).when(StringUtils.hasText(query.getStatus())))
                .orderBy(AI_DRAFT.CREATE_TIME.desc());

        Page<AiDraft> page = aiDraftMapper.paginate(Page.of(query.getCurrent(), query.getSize()), qw);
        List<AiDraftVO> records = page.getRecords().stream()
                .map(AssistantConverter::toDraftVO)
                .collect(Collectors.toList());
        return PageResult.of(records, page.getTotalRow(), page.getPageNumber(), page.getPageSize());
    }

    /** 查询草稿详情，附带明细项列表与明细数量 */
    @Override
    @Transactional(readOnly = true)
    public AiDraftVO getDraft(Long id) {
        AiDraft draft = loadDraft(id);
        AiDraftVO vo = AssistantConverter.toDraftVO(draft);
        List<AiDraftItem> items = aiDraftItemMapper.selectListByQuery(QueryWrapper.create()
                .where(AI_DRAFT_ITEM.DRAFT_ID.eq(id))
                .orderBy(AI_DRAFT_ITEM.ID.asc()));
        List<AiDraftItemVO> itemVOs = items.stream()
                .map(AssistantConverter::toDraftItemVO)
                .collect(Collectors.toList());
        vo.setItems(itemVOs);
        vo.setItemCount(itemVOs.size());
        return vo;
    }

    /** 审阅单条明细项：仅 pending 状态可审阅，写入审阅结论与意见 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void reviewItem(Long draftId, Long itemId, DraftReviewDTO dto) {
        loadDraft(draftId);
        AiDraftItem item = aiDraftItemMapper.selectOneById(itemId);
        if (item == null || !draftId.equals(item.getDraftId())) {
            throw new BusinessException(AssistantErrorCode.DRAFT_NOT_FOUND);
        }
        if (!ITEM_PENDING.equals(item.getReviewStatus())) {
            throw new BusinessException(AssistantErrorCode.DRAFT_ALREADY_REVIEWED);
        }
        item.setReviewStatus(dto.getReviewStatus());
        item.setReviewerComment(dto.getComment());
        aiDraftItemMapper.update(item);
    }

    /**
     * 发布草稿：把已批准（approved）明细项在单事务内落地为正式资源。
     *
     * <p>支持 class / property / mapping / datasource 四类明细；任一明细落地失败即整体回滚，
     * 草稿状态保持 draft，可修正后重新发布。落地成功后回写明细项 {@code targetId} 便于溯源。
     *
     * @param draftId 草稿ID
     */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void publishDraft(Long draftId) {
        AiDraft draft = loadDraft(draftId);
        if (DRAFT_PUBLISHED.equals(draft.getStatus())) {
            throw new BusinessException(AssistantErrorCode.DRAFT_PUBLISH_FAILED, "草稿已发布，请勿重复操作");
        }
        if (DRAFT_REJECTED.equals(draft.getStatus())) {
            throw new BusinessException(AssistantErrorCode.DRAFT_PUBLISH_FAILED, "草稿已驳回，不可发布");
        }
        List<AiDraftItem> items = aiDraftItemMapper.selectListByQuery(QueryWrapper.create()
                .where(AI_DRAFT_ITEM.DRAFT_ID.eq(draftId))
                .and(AI_DRAFT_ITEM.REVIEW_STATUS.eq(ITEM_APPROVED))
                .orderBy(AI_DRAFT_ITEM.ID.asc()));
        if (items.isEmpty()) {
            throw new BusinessException(AssistantErrorCode.DRAFT_PUBLISH_FAILED, "没有已批准的明细项可发布");
        }
        // 类编码 → 类ID 缓存：同一草稿内先建类再加属性/映射时避免重复查询
        Map<String, Long> classIdByCode = new HashMap<>();
        for (AiDraftItem item : items) {
            applyItem(draft, item, classIdByCode);
        }
        draft.setStatus(DRAFT_PUBLISHED);
        String base = StringUtils.hasText(draft.getSummary()) ? draft.getSummary() : "";
        draft.setSummary(base + (base.isEmpty() ? "" : " | ") + "已发布 " + items.size() + " 项");
        aiDraftMapper.update(draft);
    }

    /**
     * 按明细项类型分发落地逻辑。
     *
     * @param draft         草稿头
     * @param item          明细项
     * @param classIdByCode 类编码→类ID 缓存
     */
    private void applyItem(AiDraft draft, AiDraftItem item, Map<String, Long> classIdByCode) {
        Map<String, Object> after = readSnapshot(item.getAfterSnapshot());
        String itemType = item.getItemType() == null ? "" : item.getItemType();
        switch (itemType) {
            case "class" -> applyClass(draft, item, after, classIdByCode);
            case "property" -> applyProperty(draft, item, after, classIdByCode);
            case "mapping" -> applyMapping(draft, item, after, classIdByCode);
            case "datasource" -> applyDatasource(item, after);
            default -> throw new BusinessException(AssistantErrorCode.DRAFT_PUBLISH_FAILED,
                    "暂不支持发布的明细类型: " + itemType);
        }
    }

    /**
     * 落地本体类。
     *
     * @param draft         草稿头
     * @param item          明细项
     * @param after         变更后快照
     * @param classIdByCode 类编码→类ID 缓存
     */
    private void applyClass(AiDraft draft, AiDraftItem item, Map<String, Object> after, Map<String, Long> classIdByCode) {
        String code = strOf(after.get("code"));
        if (!StringUtils.hasText(code)) {
            throw new BusinessException(AssistantErrorCode.DRAFT_PUBLISH_FAILED, "类编码为空，无法发布");
        }
        OntClassDTO dto = new OntClassDTO();
        dto.setOntologyId(longOf(after.get("ontologyId"), draft.getOntologyId()));
        dto.setCode(code);
        dto.setName(strOf(after.get("name")));
        dto.setDescription(strOf(after.get("description")));
        // sort_order 为 NOT NULL 列，而下层 Service 是直拷不兜底，发布路径必须自带默认值
        dto.setSortOrder(intOf(after.get("sortOrder")) == null ? 0 : intOf(after.get("sortOrder")));
        dto.setStatus(1);
        Long classId = ontologyService.createClass(dto);
        classIdByCode.put(code, classId);
        markApplied(item, classId);
    }

    /**
     * 落地本体属性（需先解析所属类ID）。
     *
     * @param draft         草稿头
     * @param item          明细项
     * @param after         变更后快照
     * @param classIdByCode 类编码→类ID 缓存
     */
    private void applyProperty(AiDraft draft, AiDraftItem item, Map<String, Object> after, Map<String, Long> classIdByCode) {
        Long ontologyId = longOf(after.get("ontologyId"), draft.getOntologyId());
        OntPropertyDTO dto = new OntPropertyDTO();
        dto.setClassId(resolveClassId(ontologyId, strOf(after.get("classCode")), classIdByCode));
        dto.setCode(strOf(after.get("code")));
        dto.setName(strOf(after.get("name")));
        dto.setDataType(strOf(after.get("dataType")));
        dto.setDescription(strOf(after.get("description")));
        dto.setDefaultValue(strOf(after.get("defaultValue")));
        dto.setUnit(strOf(after.get("unit")));
        // required / sort_order 为 NOT NULL 列，快照缺省时须兜底，否则插入报 Column 'required' cannot be null
        Integer required = intOf(after.get("required"));
        dto.setRequired(required == null ? 0 : required);
        Integer sortOrder = intOf(after.get("sortOrder"));
        dto.setSortOrder(sortOrder == null ? 0 : sortOrder);
        dto.setStatus(1);
        markApplied(item, ontologyService.createProperty(dto));
    }

    /**
     * 落地类-物理表映射。
     *
     * @param draft         草稿头
     * @param item          明细项
     * @param after         变更后快照
     * @param classIdByCode 类编码→类ID 缓存
     */
    private void applyMapping(AiDraft draft, AiDraftItem item, Map<String, Object> after, Map<String, Long> classIdByCode) {
        Long ontologyId = longOf(after.get("ontologyId"), draft.getOntologyId());
        MappingConfigDTO dto = new MappingConfigDTO();
        dto.setOntologyId(ontologyId);
        dto.setClassId(resolveClassId(ontologyId, strOf(after.get("classCode")), classIdByCode));
        dto.setDatasourceId(longOf(after.get("datasourceId"), null));
        dto.setTableName(strOf(after.get("tableName")));
        dto.setColumnMappings(writeJson(after.get("columnMappings")));
        dto.setPriority(100);
        dto.setStatus(1);
        markApplied(item, mappingService.create(dto));
    }

    /**
     * 落地数据源（密码留空，需人工在数据源管理页补录后才能测试连接）。
     *
     * @param item  明细项
     * @param after 变更后快照
     */
    private void applyDatasource(AiDraftItem item, Map<String, Object> after) {
        DatasourceDTO dto = new DatasourceDTO();
        dto.setCode(strOf(after.get("code")));
        dto.setName(strOf(after.get("name")));
        dto.setType(strOf(after.get("type")));
        dto.setHost(strOf(after.get("host")));
        dto.setPort(intOf(after.get("port")));
        dto.setDbName(strOf(after.get("dbName")));
        dto.setSchemaName(strOf(after.get("schemaName")));
        dto.setUsername(strOf(after.get("username")));
        dto.setStatus(1);
        dto.setRemark("由 AI 草稿发布创建，密码待人工补录");
        markApplied(item, datasourceService.create(dto));
    }

    /**
     * 回写明细项落地结果（目标实体ID）。
     *
     * @param item     明细项
     * @param targetId 新建资源ID
     */
    private void markApplied(AiDraftItem item, Long targetId) {
        item.setTargetId(targetId);
        aiDraftItemMapper.update(item);
    }

    /**
     * 解析类编码为类ID：先查缓存，未命中则拉取本体类树扁平化查找。
     *
     * @param ontologyId    本体ID
     * @param classCode     类编码
     * @param classIdByCode 类编码→类ID 缓存（会被填充）
     * @return 类ID
     */
    private Long resolveClassId(Long ontologyId, String classCode, Map<String, Long> classIdByCode) {
        if (StringUtils.hasText(classCode) && classIdByCode.containsKey(classCode)) {
            return classIdByCode.get(classCode);
        }
        if (ontologyId == null || !StringUtils.hasText(classCode)) {
            throw new BusinessException(AssistantErrorCode.DRAFT_PUBLISH_FAILED, "缺少本体ID或类编码，无法解析目标类");
        }
        Map<String, Long> found = new HashMap<>();
        flattenClasses(ontologyService.listClasses(ontologyId), found);
        classIdByCode.putAll(found);
        Long classId = classIdByCode.get(classCode);
        if (classId == null) {
            throw new BusinessException(AssistantErrorCode.DRAFT_PUBLISH_FAILED, "本体中不存在类：" + classCode);
        }
        return classId;
    }

    /**
     * 递归扁平化类树为「类编码 → 类ID」。
     *
     * @param nodes 类节点列表
     * @param out   输出映射
     */
    private void flattenClasses(List<OntClassVO> nodes, Map<String, Long> out) {
        if (nodes == null) {
            return;
        }
        for (OntClassVO node : nodes) {
            Long id = longOf(node.getId(), null);
            if (node.getCode() != null && id != null) {
                out.put(node.getCode(), id);
            }
            flattenClasses(node.getChildren(), out);
        }
    }

    /**
     * 解析变更后快照 JSON，非法或为空返回空 Map。
     *
     * @param json 快照 JSON
     * @return 快照 Map
     */
    private Map<String, Object> readSnapshot(String json) {
        if (!StringUtils.hasText(json)) {
            return Map.of();
        }
        try {
            Map<String, Object> map = OBJECT_MAPPER.readValue(json, SNAPSHOT_TYPE);
            return map == null ? Map.of() : map;
        } catch (Exception e) {
            throw new BusinessException(AssistantErrorCode.DRAFT_PUBLISH_FAILED, "草稿快照解析失败，无法发布");
        }
    }

    /**
     * 将列映射对象序列化回 JSON 字符串。
     *
     * @param value 列映射对象
     * @return JSON 字符串，空值返回 null
     */
    private String writeJson(Object value) {
        if (value == null) {
            return null;
        }
        try {
            return OBJECT_MAPPER.writeValueAsString(value);
        } catch (Exception e) {
            throw new BusinessException(AssistantErrorCode.DRAFT_PUBLISH_FAILED, "列映射序列化失败，无法发布");
        }
    }

    /**
     * 取字符串值（null 安全，空白串归一为 null）。
     *
     * @param o 原始值
     * @return 字符串或 null
     */
    private String strOf(Object o) {
        if (o == null) {
            return null;
        }
        String s = String.valueOf(o);
        return StringUtils.hasText(s) ? s : null;
    }

    /**
     * 取 Long 值（兼容字符串形式的ID）。
     *
     * @param o            原始值
     * @param defaultValue 解析失败时的兜底值
     * @return Long 或兜底值
     */
    private Long longOf(Object o, Long defaultValue) {
        if (o instanceof Number n) {
            return n.longValue();
        }
        if (o != null) {
            try {
                return Long.valueOf(String.valueOf(o).trim());
            } catch (NumberFormatException ignored) {
                return defaultValue;
            }
        }
        return defaultValue;
    }

    /**
     * 取 Integer 值（兼容字符串形式的数字）。
     *
     * @param o 原始值
     * @return Integer 或 null
     */
    private Integer intOf(Object o) {
        Long v = longOf(o, null);
        return v == null ? null : v.intValue();
    }

    /** 驳回草稿：推进状态为 rejected，并记录驳回原因 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void rejectDraft(Long draftId, String reason) {
        AiDraft draft = loadDraft(draftId);
        draft.setStatus(DRAFT_REJECTED);
        if (StringUtils.hasText(reason)) {
            String base = StringUtils.hasText(draft.getSummary()) ? draft.getSummary() : "";
            draft.setSummary(base + (base.isEmpty() ? "" : " | ") + "驳回原因: " + reason);
        }
        aiDraftMapper.update(draft);
    }

    /**
     * 加载草稿实体，不存在抛业务异常。
     *
     * @param id 草稿ID
     * @return 草稿实体
     */
    private AiDraft loadDraft(Long id) {
        AiDraft draft = aiDraftMapper.selectOneById(id);
        if (draft == null) {
            throw new BusinessException(AssistantErrorCode.DRAFT_NOT_FOUND);
        }
        return draft;
    }
}
