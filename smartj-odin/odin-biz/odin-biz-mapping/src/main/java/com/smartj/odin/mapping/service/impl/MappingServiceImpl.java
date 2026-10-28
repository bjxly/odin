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
package com.smartj.odin.mapping.service.impl;

import com.mybatisflex.core.paginate.Page;
import com.mybatisflex.core.query.QueryWrapper;
import com.smartj.common.core.domain.PageResult;
import com.smartj.common.core.exception.BusinessException;
import com.smartj.odin.datasource.entity.Datasource;
import com.smartj.odin.datasource.entity.SchemaColumn;
import com.smartj.odin.datasource.entity.SchemaTable;
import com.smartj.odin.datasource.mapper.DatasourceMapper;
import com.smartj.odin.datasource.mapper.SchemaColumnMapper;
import com.smartj.odin.datasource.mapper.SchemaTableMapper;
import com.smartj.odin.mapping.converter.MappingConverter;
import com.smartj.odin.mapping.dto.MappingConfigDTO;
import com.smartj.odin.mapping.dto.MappingConfigQuery;
import com.smartj.odin.mapping.entity.MappingConfig;
import com.smartj.odin.mapping.error.MappingErrorCode;
import com.smartj.odin.mapping.mapper.MappingConfigMapper;
import com.smartj.odin.mapping.service.MappingService;
import com.smartj.odin.mapping.vo.MappingConfigVO;
import com.smartj.odin.mapping.vo.MappingSuggestVO;
import com.smartj.odin.ontology.entity.OntClass;
import com.smartj.odin.ontology.entity.OntClassProperty;
import com.smartj.odin.ontology.entity.OntDefinition;
import com.smartj.odin.ontology.mapper.OntClassMapper;
import com.smartj.odin.ontology.mapper.OntClassPropertyMapper;
import com.smartj.odin.ontology.mapper.OntDefinitionMapper;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.stream.Collectors;

import static com.smartj.odin.datasource.entity.table.DatasourceTableDef.DATASOURCE;
import static com.smartj.odin.datasource.entity.table.SchemaColumnTableDef.SCHEMA_COLUMN;
import static com.smartj.odin.datasource.entity.table.SchemaTableTableDef.SCHEMA_TABLE;
import static com.smartj.odin.mapping.entity.table.MappingConfigTableDef.MAPPING_CONFIG;
import static com.smartj.odin.ontology.entity.table.OntClassPropertyTableDef.ONT_CLASS_PROPERTY;
import static com.smartj.odin.ontology.entity.table.OntClassTableDef.ONT_CLASS;
import static com.smartj.odin.ontology.entity.table.OntDefinitionTableDef.ONT_DEFINITION;

/**
 * 映射配置服务实现。
 *
 * <p>提供映射配置 CRUD、完整性校验与智能列映射推荐能力。
 * 跨模块依赖本体域（类/属性）与数据源域（表/列元数据）的 Mapper。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Service
@RequiredArgsConstructor
public class MappingServiceImpl implements MappingService {

    /** column_mappings JSON 解析器（线程安全，复用单例） */
    private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();

    private final MappingConfigMapper mappingConfigMapper;
    private final OntDefinitionMapper ontDefinitionMapper;
    private final OntClassMapper ontClassMapper;
    private final OntClassPropertyMapper ontClassPropertyMapper;
    private final DatasourceMapper datasourceMapper;
    private final SchemaTableMapper schemaTableMapper;
    private final SchemaColumnMapper schemaColumnMapper;

    /** 分页查询映射配置，关联查询 ontologyName / className / datasourceName */
    @Override
    @Transactional(readOnly = true)
    public PageResult<MappingConfigVO> page(MappingConfigQuery query) {
        QueryWrapper qw = QueryWrapper.create()
                .where(MAPPING_CONFIG.ONTOLOGY_ID.eq(query.getOntologyId()).when(query.getOntologyId() != null))
                .and(MAPPING_CONFIG.CLASS_ID.eq(query.getClassId()).when(query.getClassId() != null))
                .and(MAPPING_CONFIG.DATASOURCE_ID.eq(query.getDatasourceId()).when(query.getDatasourceId() != null))
                .and(MAPPING_CONFIG.TABLE_NAME.like(query.getTableName()).when(StringUtils.hasText(query.getTableName())))
                .and(MAPPING_CONFIG.STATUS.eq(query.getStatus()).when(query.getStatus() != null))
                .orderBy(MAPPING_CONFIG.CREATE_TIME.desc());

        Page<MappingConfig> page = mappingConfigMapper.paginate(Page.of(query.getCurrent(), query.getSize()), qw);
        List<MappingConfigVO> records = page.getRecords().stream()
                .map(e -> {
                    MappingConfigVO vo = MappingConverter.toVO(e);
                    enrichNames(vo, e);
                    return vo;
                })
                .collect(Collectors.toList());
        return PageResult.of(records, page.getTotalRow(), page.getPageNumber(), page.getPageSize());
    }

    /** 根据ID查询映射配置详情 */
    @Override
    @Transactional(readOnly = true)
    public MappingConfigVO getById(Long id) {
        MappingConfig entity = loadEntity(id);
        MappingConfigVO vo = MappingConverter.toVO(entity);
        enrichNames(vo, entity);
        return vo;
    }

    /** 新增映射配置：UK 校验（ontologyId + classId + datasourceId + tableName） */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public Long create(MappingConfigDTO dto) {
        MappingConfig existing = mappingConfigMapper.selectOneByQuery(QueryWrapper.create()
                .where(MAPPING_CONFIG.ONTOLOGY_ID.eq(dto.getOntologyId()))
                .and(MAPPING_CONFIG.CLASS_ID.eq(dto.getClassId()))
                .and(MAPPING_CONFIG.DATASOURCE_ID.eq(dto.getDatasourceId()))
                .and(MAPPING_CONFIG.TABLE_NAME.eq(dto.getTableName())));
        if (existing != null) {
            throw new BusinessException(MappingErrorCode.MAPPING_ALREADY_EXISTS);
        }
        MappingConfig entity = new MappingConfig();
        copyDtoToEntity(dto, entity);
        mappingConfigMapper.insert(entity);
        return entity.getId();
    }

    /** 更新映射配置 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void update(MappingConfigDTO dto) {
        MappingConfig entity = loadEntity(dto.getId());
        MappingConfig sameUk = mappingConfigMapper.selectOneByQuery(QueryWrapper.create()
                .where(MAPPING_CONFIG.ONTOLOGY_ID.eq(dto.getOntologyId()))
                .and(MAPPING_CONFIG.CLASS_ID.eq(dto.getClassId()))
                .and(MAPPING_CONFIG.DATASOURCE_ID.eq(dto.getDatasourceId()))
                .and(MAPPING_CONFIG.TABLE_NAME.eq(dto.getTableName())));
        if (sameUk != null && !sameUk.getId().equals(dto.getId())) {
            throw new BusinessException(MappingErrorCode.MAPPING_ALREADY_EXISTS);
        }
        copyDtoToEntity(dto, entity);
        mappingConfigMapper.update(entity);
    }

    /** 删除映射配置 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void delete(Long id) {
        loadEntity(id);
        mappingConfigMapper.deleteById(id);
    }

    /** 校验映射完整性：检查 columnMappings 中的列是否存在于数据源 schema 中 */
    @Override
    @Transactional(readOnly = true)
    public String validate(Long id) {
        MappingConfig config = loadEntity(id);
        if (!StringUtils.hasText(config.getColumnMappings())) {
            return "列映射为空，无需校验";
        }

        SchemaTable table = schemaTableMapper.selectOneByQuery(QueryWrapper.create()
                .where(SCHEMA_TABLE.DATASOURCE_ID.eq(config.getDatasourceId()))
                .and(SCHEMA_TABLE.TABLE_NAME.eq(config.getTableName())));
        if (table == null) {
            throw new BusinessException(MappingErrorCode.MAPPING_VALIDATE_FAILED);
        }

        List<SchemaColumn> columns = schemaColumnMapper.selectListByQuery(
                QueryWrapper.create().where(SCHEMA_COLUMN.TABLE_ID.eq(table.getId())));
        Set<String> columnNames = columns.stream()
                .map(SchemaColumn::getColumnName)
                .collect(Collectors.toSet());

        // 解析 column_mappings JSON（规范键 columnName，兼容历史 column_name / column）
        List<String> missingColumns = new ArrayList<>();
        List<Map<String, Object>> items;
        try {
            items = OBJECT_MAPPER.readValue(config.getColumnMappings(),
                    new TypeReference<List<Map<String, Object>>>() {
                    });
        } catch (Exception e) {
            throw new BusinessException(MappingErrorCode.MAPPING_VALIDATE_FAILED);
        }
        for (Map<String, Object> item : items) {
            Object col = item.get("columnName");
            if (col == null) {
                col = item.get("column_name");
            }
            if (col == null) {
                col = item.get("column");
            }
            if (col == null) {
                continue;
            }
            String colName = String.valueOf(col);
            if (!colName.isEmpty() && !columnNames.contains(colName)) {
                missingColumns.add(colName);
            }
        }

        if (missingColumns.isEmpty()) {
            return "校验通过，所有映射列均存在于数据源 schema 中";
        }
        return "校验失败，以下列不存在于数据源 schema 中：" + String.join(", ", missingColumns);
    }

    /** 基于名称相似度推荐列-属性映射（按数据源ID+表名定位物理表） */
    @Override
    @Transactional(readOnly = true)
    public List<MappingSuggestVO> suggest(Long ontologyId, Long classId, Long datasourceId, String tableName) {
        // 获取类的属性列表
        List<OntClassProperty> properties = ontClassPropertyMapper.selectListByQuery(
                QueryWrapper.create().where(ONT_CLASS_PROPERTY.CLASS_ID.eq(classId)));
        if (properties.isEmpty()) {
            return Collections.emptyList();
        }

        // 获取目标表的列列表
        SchemaTable table = schemaTableMapper.selectOneByQuery(QueryWrapper.create()
                .where(SCHEMA_TABLE.DATASOURCE_ID.eq(datasourceId))
                .and(SCHEMA_TABLE.TABLE_NAME.eq(tableName)));
        if (table == null) {
            return Collections.emptyList();
        }
        List<SchemaColumn> columns = schemaColumnMapper.selectListByQuery(
                QueryWrapper.create().where(SCHEMA_COLUMN.TABLE_ID.eq(table.getId())));
        return buildSuggestions(properties, columns);
    }

    /** 基于名称相似度推荐列-属性映射（前端契约：按 classId + 物理表元数据ID 直接定位） */
    @Override
    @Transactional(readOnly = true)
    public List<MappingSuggestVO> suggestByTable(Long classId, Long tableId) {
        if (classId == null || tableId == null) {
            return Collections.emptyList();
        }
        List<OntClassProperty> properties = ontClassPropertyMapper.selectListByQuery(
                QueryWrapper.create().where(ONT_CLASS_PROPERTY.CLASS_ID.eq(classId)));
        if (properties.isEmpty()) {
            return Collections.emptyList();
        }
        List<SchemaColumn> columns = schemaColumnMapper.selectListByQuery(
                QueryWrapper.create().where(SCHEMA_COLUMN.TABLE_ID.eq(tableId)));
        return buildSuggestions(properties, columns);
    }

    /** 对每对 (property, column) 计算名称相似度，过滤低置信度并按置信度降序返回 */
    private List<MappingSuggestVO> buildSuggestions(List<OntClassProperty> properties, List<SchemaColumn> columns) {
        if (columns == null || columns.isEmpty()) {
            return Collections.emptyList();
        }
        List<MappingSuggestVO> suggestions = new ArrayList<>();
        for (OntClassProperty prop : properties) {
            for (SchemaColumn col : columns) {
                double confidence = computeSimilarity(prop.getCode(), col.getColumnName());
                if (confidence > 0.3) {
                    MappingSuggestVO vo = new MappingSuggestVO();
                    vo.setColumnName(col.getColumnName());
                    vo.setPropertyName(prop.getName());
                    vo.setPropertyCode(prop.getCode());
                    vo.setConfidence(Math.round(confidence * 100.0) / 100.0);
                    vo.setMatchReason(confidence >= 0.9 ? "名称包含" : "相似度匹配");
                    suggestions.add(vo);
                }
            }
        }
        suggestions.sort((a, b) -> Double.compare(b.getConfidence(), a.getConfidence()));
        return suggestions;
    }

    // ==================== 私有辅助 ====================

    private MappingConfig loadEntity(Long id) {
        MappingConfig entity = mappingConfigMapper.selectOneById(id);
        if (entity == null) {
            throw new BusinessException(MappingErrorCode.MAPPING_NOT_FOUND);
        }
        return entity;
    }

    private void copyDtoToEntity(MappingConfigDTO dto, MappingConfig entity) {
        entity.setOntologyId(dto.getOntologyId());
        entity.setClassId(dto.getClassId());
        entity.setDatasourceId(dto.getDatasourceId());
        entity.setTableName(dto.getTableName());
        entity.setSchemaName(dto.getSchemaName());
        entity.setColumnMappings(dto.getColumnMappings());
        entity.setFilterCondition(dto.getFilterCondition());
        entity.setPriority(dto.getPriority());
        entity.setStatus(dto.getStatus() == null ? 1 : dto.getStatus());
        entity.setRemark(dto.getRemark());
    }

    /** 为 VO 填充关联名称（本体名、类名、数据源名） */
    private void enrichNames(MappingConfigVO vo, MappingConfig entity) {
        if (entity.getOntologyId() != null) {
            OntDefinition ont = ontDefinitionMapper.selectOneById(entity.getOntologyId());
            if (ont != null) {
                vo.setOntologyName(ont.getName());
            }
        }
        if (entity.getClassId() != null) {
            OntClass cls = ontClassMapper.selectOneById(entity.getClassId());
            if (cls != null) {
                vo.setClassName(cls.getName());
            }
        }
        if (entity.getDatasourceId() != null) {
            Datasource ds = datasourceMapper.selectOneById(entity.getDatasourceId());
            if (ds != null) {
                vo.setDatasourceName(ds.getName());
            }
        }
    }

    /**
     * 计算两个名称的相似度（0~1）。
     *
     * <p>规则：完全匹配=1.0；一方包含另一方=0.9；否则用 Levenshtein 距离归一化。
     *
     * @param s1 名称1
     * @param s2 名称2
     * @return 相似度
     */
    private double computeSimilarity(String s1, String s2) {
        if (s1 == null || s2 == null) {
            return 0.0;
        }
        String n1 = normalize(s1);
        String n2 = normalize(s2);
        if (n1.equals(n2)) {
            return 1.0;
        }
        if (n1.contains(n2) || n2.contains(n1)) {
            return 0.9;
        }
        int maxLen = Math.max(n1.length(), n2.length());
        if (maxLen == 0) {
            return 1.0;
        }
        int dist = levenshteinDistance(n1, n2);
        return 1.0 - (double) dist / maxLen;
    }

    /** 名称归一化：小写 + 去除下划线和连字符 */
    private String normalize(String s) {
        return s.toLowerCase().replace("_", "").replace("-", "");
    }

    /** 计算 Levenshtein 编辑距离 */
    private int levenshteinDistance(String s1, String s2) {
        int len1 = s1.length();
        int len2 = s2.length();
        int[][] dp = new int[len1 + 1][len2 + 1];
        for (int i = 0; i <= len1; i++) {
            dp[i][0] = i;
        }
        for (int j = 0; j <= len2; j++) {
            dp[0][j] = j;
        }
        for (int i = 1; i <= len1; i++) {
            for (int j = 1; j <= len2; j++) {
                int cost = s1.charAt(i - 1) == s2.charAt(j - 1) ? 0 : 1;
                dp[i][j] = Math.min(Math.min(dp[i - 1][j] + 1, dp[i][j - 1] + 1), dp[i - 1][j - 1] + cost);
            }
        }
        return dp[len1][len2];
    }
}
