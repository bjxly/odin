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
package com.smartj.odin.ontology.service.impl;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.mybatisflex.core.paginate.Page;
import com.mybatisflex.core.query.QueryWrapper;
import com.smartj.common.core.domain.PageResult;
import com.smartj.common.core.exception.BusinessException;
import com.smartj.odin.ontology.converter.OntologyConverter;
import com.smartj.odin.ontology.dto.OntActionDTO;
import com.smartj.odin.ontology.dto.OntClassDTO;
import com.smartj.odin.ontology.dto.OntPropertyDTO;
import com.smartj.odin.ontology.dto.OntRelationDTO;
import com.smartj.odin.ontology.dto.OntRuleDTO;
import com.smartj.odin.ontology.dto.OntologyDTO;
import com.smartj.odin.ontology.dto.OntologyQuery;
import com.smartj.odin.ontology.entity.OntAction;
import com.smartj.odin.ontology.entity.OntClass;
import com.smartj.odin.ontology.entity.OntClassProperty;
import com.smartj.odin.ontology.entity.OntDefinition;
import com.smartj.odin.ontology.entity.OntRelation;
import com.smartj.odin.ontology.entity.OntRule;
import com.smartj.odin.ontology.error.OntologyErrorCode;
import com.smartj.odin.ontology.mapper.OntActionMapper;
import com.smartj.odin.ontology.mapper.OntClassMapper;
import com.smartj.odin.ontology.mapper.OntClassPropertyMapper;
import com.smartj.odin.ontology.mapper.OntDefinitionMapper;
import com.smartj.odin.ontology.mapper.OntRelationMapper;
import com.smartj.odin.ontology.mapper.OntRuleMapper;
import com.smartj.odin.ontology.service.OntologyService;
import com.smartj.odin.ontology.vo.OntActionVO;
import com.smartj.odin.ontology.vo.OntClassVO;
import com.smartj.odin.ontology.vo.OntPropertyVO;
import com.smartj.odin.ontology.vo.OntRelationVO;
import com.smartj.odin.ontology.vo.OntRuleVO;
import com.smartj.odin.ontology.vo.OntologyGraphVO;
import com.smartj.odin.ontology.vo.OntologyVO;
import com.smartj.odin.ontology.vo.VocabularyVO;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import static com.smartj.odin.ontology.entity.table.OntActionTableDef.ONT_ACTION;
import static com.smartj.odin.ontology.entity.table.OntClassPropertyTableDef.ONT_CLASS_PROPERTY;
import static com.smartj.odin.ontology.entity.table.OntClassTableDef.ONT_CLASS;
import static com.smartj.odin.ontology.entity.table.OntDefinitionTableDef.ONT_DEFINITION;
import static com.smartj.odin.ontology.entity.table.OntRelationTableDef.ONT_RELATION;
import static com.smartj.odin.ontology.entity.table.OntRuleTableDef.ONT_RULE;

/**
 * 本体服务实现。
 *
 * <p>提供本体定义、类、属性、关系、规则、动作的完整 CRUD，
 * 以及图谱组装与词表导出能力。写操作使用事务保证一致性。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Service
@RequiredArgsConstructor
public class OntologyServiceImpl implements OntologyService {

    /** preset_questions 等 JSON 列序列化器（线程安全，复用单例） */
    private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();

    private final OntDefinitionMapper ontDefinitionMapper;
    private final OntClassMapper ontClassMapper;
    private final OntClassPropertyMapper ontClassPropertyMapper;
    private final OntRelationMapper ontRelationMapper;
    private final OntRuleMapper ontRuleMapper;
    private final OntActionMapper ontActionMapper;

    // ==================== 本体定义 ====================

    /** 分页查询本体，附加 classCount / relationCount */
    @Override
    @Transactional(readOnly = true)
    public PageResult<OntologyVO> page(OntologyQuery query) {
        QueryWrapper qw = QueryWrapper.create()
                .where(ONT_DEFINITION.CODE.like(query.getCode()).when(StringUtils.hasText(query.getCode())))
                .and(ONT_DEFINITION.NAME.like(query.getName()).when(StringUtils.hasText(query.getName())))
                .and(ONT_DEFINITION.STATUS.eq(query.getStatus()).when(query.getStatus() != null))
                .orderBy(ONT_DEFINITION.CREATE_TIME.desc());

        Page<OntDefinition> page = ontDefinitionMapper.paginate(Page.of(query.getCurrent(), query.getSize()), qw);
        List<OntologyVO> records = page.getRecords().stream()
                .map(e -> {
                    OntologyVO vo = OntologyConverter.toVO(e);
                    vo.setClassCount((int) ontClassMapper.selectCountByQuery(
                            QueryWrapper.create().where(ONT_CLASS.ONTOLOGY_ID.eq(e.getId()))));
                    vo.setRelationCount((int) ontRelationMapper.selectCountByQuery(
                            QueryWrapper.create().where(ONT_RELATION.ONTOLOGY_ID.eq(e.getId()))));
                    return vo;
                })
                .collect(Collectors.toList());
        return PageResult.of(records, page.getTotalRow(), page.getPageNumber(), page.getPageSize());
    }

    /** 查询全部启用本体（不分页，附带 classCount/relationCount） */
    @Override
    @Transactional(readOnly = true)
    public List<OntologyVO> listAll() {
        List<OntDefinition> list = ontDefinitionMapper.selectListByQuery(
                QueryWrapper.create()
                        .where(ONT_DEFINITION.STATUS.eq(1))
                        .orderBy(ONT_DEFINITION.CREATE_TIME.desc()));
        return list.stream()
                .map(e -> {
                    OntologyVO vo = OntologyConverter.toVO(e);
                    vo.setClassCount((int) ontClassMapper.selectCountByQuery(
                            QueryWrapper.create().where(ONT_CLASS.ONTOLOGY_ID.eq(e.getId()))));
                    vo.setRelationCount((int) ontRelationMapper.selectCountByQuery(
                            QueryWrapper.create().where(ONT_RELATION.ONTOLOGY_ID.eq(e.getId()))));
                    return vo;
                })
                .collect(Collectors.toList());
    }

    /** 根据ID查询本体详情 */
    @Override
    @Transactional(readOnly = true)
    public OntologyVO getById(Long id) {
        OntologyVO vo = OntologyConverter.toVO(loadEntity(id));
        vo.setClassCount((int) ontClassMapper.selectCountByQuery(
                QueryWrapper.create().where(ONT_CLASS.ONTOLOGY_ID.eq(id))));
        vo.setRelationCount((int) ontRelationMapper.selectCountByQuery(
                QueryWrapper.create().where(ONT_RELATION.ONTOLOGY_ID.eq(id))));
        return vo;
    }

    /** 新增本体：校验编码唯一 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public Long create(OntologyDTO dto) {
        OntDefinition existing = ontDefinitionMapper.selectOneByQuery(
                QueryWrapper.create().where(ONT_DEFINITION.CODE.eq(dto.getCode())));
        if (existing != null) {
            throw new BusinessException(OntologyErrorCode.ONTOLOGY_CODE_EXISTS);
        }
        OntDefinition entity = new OntDefinition();
        entity.setCode(dto.getCode());
        entity.setName(dto.getName());
        entity.setDescription(dto.getDescription());
        entity.setVersion(dto.getVersion());
        entity.setPresetQuestions(writeStringList(dto.getPresetQuestions()));
        entity.setStatus(dto.getStatus() == null ? 1 : dto.getStatus());
        entity.setRemark(dto.getRemark());
        ontDefinitionMapper.insert(entity);
        return entity.getId();
    }

    /** 更新本体：编码唯一性排除自身 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void update(OntologyDTO dto) {
        OntDefinition entity = loadEntity(dto.getId());
        OntDefinition sameCode = ontDefinitionMapper.selectOneByQuery(
                QueryWrapper.create().where(ONT_DEFINITION.CODE.eq(dto.getCode())));
        if (sameCode != null && !sameCode.getId().equals(dto.getId())) {
            throw new BusinessException(OntologyErrorCode.ONTOLOGY_CODE_EXISTS);
        }
        entity.setCode(dto.getCode());
        entity.setName(dto.getName());
        entity.setDescription(dto.getDescription());
        entity.setVersion(dto.getVersion());
        entity.setPresetQuestions(writeStringList(dto.getPresetQuestions()));
        if (dto.getStatus() != null) {
            entity.setStatus(dto.getStatus());
        }
        entity.setRemark(dto.getRemark());
        ontDefinitionMapper.update(entity);
    }

    /** 删除本体：级联删除类→属性、关系、规则、动作 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void delete(Long id) {
        loadEntity(id);
        List<OntClass> classes = ontClassMapper.selectListByQuery(
                QueryWrapper.create().where(ONT_CLASS.ONTOLOGY_ID.eq(id)));
        if (!classes.isEmpty()) {
            List<Long> classIds = classes.stream().map(OntClass::getId).collect(Collectors.toList());
            ontClassPropertyMapper.deleteByQuery(
                    QueryWrapper.create().where(ONT_CLASS_PROPERTY.CLASS_ID.in(classIds)));
            ontClassMapper.deleteByQuery(
                    QueryWrapper.create().where(ONT_CLASS.ONTOLOGY_ID.eq(id)));
        }
        ontRelationMapper.deleteByQuery(QueryWrapper.create().where(ONT_RELATION.ONTOLOGY_ID.eq(id)));
        ontRuleMapper.deleteByQuery(QueryWrapper.create().where(ONT_RULE.ONTOLOGY_ID.eq(id)));
        ontActionMapper.deleteByQuery(QueryWrapper.create().where(ONT_ACTION.ONTOLOGY_ID.eq(id)));
        ontDefinitionMapper.deleteById(id);
    }

    /** List<String> → JSON 字符串数组（供 preset_questions 列），空列表存 null */
    private String writeStringList(List<String> list) {
        if (list == null || list.isEmpty()) {
            return null;
        }
        try {
            return OBJECT_MAPPER.writeValueAsString(list);
        } catch (JsonProcessingException e) {
            return null;
        }
    }

    // ==================== 图谱 & 词表 ====================

    /** 获取本体图谱：类→节点，关系→边 */
    @Override
    @Transactional(readOnly = true)
    public OntologyGraphVO getGraph(Long ontologyId) {
        loadEntity(ontologyId);
        List<OntClass> classes = ontClassMapper.selectListByQuery(
                QueryWrapper.create().where(ONT_CLASS.ONTOLOGY_ID.eq(ontologyId)));
        List<OntRelation> relations = ontRelationMapper.selectListByQuery(
                QueryWrapper.create().where(ONT_RELATION.ONTOLOGY_ID.eq(ontologyId)));

        List<OntologyGraphVO.GraphNode> nodes = new ArrayList<>();
        for (OntClass c : classes) {
            OntologyGraphVO.GraphNode node = new OntologyGraphVO.GraphNode();
            node.setId(String.valueOf(c.getId()));
            node.setLabel(c.getName());
            node.setType("class");
            node.setIcon(c.getIcon());
            List<OntClassProperty> props = ontClassPropertyMapper.selectListByQuery(
                    QueryWrapper.create().where(ONT_CLASS_PROPERTY.CLASS_ID.eq(c.getId())));
            node.setProperties(props.stream().map(OntClassProperty::getCode).collect(Collectors.toList()));
            node.setPropertyDetails(props.stream().map(p -> {
                OntologyGraphVO.NodeProperty np = new OntologyGraphVO.NodeProperty();
                np.setCode(p.getCode());
                np.setName(p.getName());
                np.setDataType(p.getDataType());
                return np;
            }).collect(Collectors.toList()));
            nodes.add(node);
        }

        List<OntologyGraphVO.GraphEdge> edges = new ArrayList<>();
        for (OntRelation r : relations) {
            OntologyGraphVO.GraphEdge edge = new OntologyGraphVO.GraphEdge();
            edge.setId(String.valueOf(r.getId()));
            edge.setSource(String.valueOf(r.getSourceClassId()));
            edge.setTarget(String.valueOf(r.getTargetClassId()));
            edge.setLabel(r.getName());
            edge.setCardinality(r.getCardinality());
            edges.add(edge);
        }

        OntologyGraphVO graph = new OntologyGraphVO();
        graph.setNodes(nodes);
        graph.setEdges(edges);
        return graph;
    }

    /** 获取本体词表：组装类/属性/关系的逻辑 Schema 描述 */
    @Override
    @Transactional(readOnly = true)
    public VocabularyVO getVocabulary(Long ontologyId) {
        OntDefinition ont = loadEntity(ontologyId);
        List<OntClass> classes = ontClassMapper.selectListByQuery(
                QueryWrapper.create().where(ONT_CLASS.ONTOLOGY_ID.eq(ontologyId)));
        List<OntRelation> relations = ontRelationMapper.selectListByQuery(
                QueryWrapper.create().where(ONT_RELATION.ONTOLOGY_ID.eq(ontologyId)));

        Map<Long, String> classCodeMap = classes.stream()
                .collect(Collectors.toMap(OntClass::getId, OntClass::getCode));

        List<VocabularyVO.ClassVocab> classVocabs = new ArrayList<>();
        for (OntClass c : classes) {
            VocabularyVO.ClassVocab cv = new VocabularyVO.ClassVocab();
            cv.setCode(c.getCode());
            cv.setName(c.getName());
            cv.setDescription(c.getDescription());
            List<OntClassProperty> props = ontClassPropertyMapper.selectListByQuery(
                    QueryWrapper.create().where(ONT_CLASS_PROPERTY.CLASS_ID.eq(c.getId())));
            cv.setProperties(props.stream().map(p -> {
                VocabularyVO.PropertyVocab pv = new VocabularyVO.PropertyVocab();
                pv.setCode(p.getCode());
                pv.setName(p.getName());
                pv.setDataType(p.getDataType());
                pv.setDescription(p.getDescription());
                return pv;
            }).collect(Collectors.toList()));
            classVocabs.add(cv);
        }

        List<VocabularyVO.RelationVocab> relVocabs = relations.stream().map(r -> {
            VocabularyVO.RelationVocab rv = new VocabularyVO.RelationVocab();
            rv.setCode(r.getCode());
            rv.setName(r.getName());
            rv.setSourceClass(classCodeMap.getOrDefault(r.getSourceClassId(), ""));
            rv.setTargetClass(classCodeMap.getOrDefault(r.getTargetClassId(), ""));
            rv.setCardinality(r.getCardinality());
            rv.setJoinKeys(r.getJoinKeys());
            return rv;
        }).collect(Collectors.toList());

        VocabularyVO vo = new VocabularyVO();
        vo.setOntologyId(String.valueOf(ont.getId()));
        vo.setOntologyName(ont.getName());
        vo.setVersion(ont.getVersion());
        vo.setClasses(classVocabs);
        vo.setRelations(relVocabs);
        return vo;
    }

    // ==================== 类 ====================

    /** 查询本体下所有类（树形结构） */
    @Override
    @Transactional(readOnly = true)
    public List<OntClassVO> listClasses(Long ontologyId) {
        List<OntClass> classes = ontClassMapper.selectListByQuery(QueryWrapper.create()
                .where(ONT_CLASS.ONTOLOGY_ID.eq(ontologyId))
                .orderBy(ONT_CLASS.SORT_ORDER.asc()));

        Map<Long, OntClassVO> voMap = new LinkedHashMap<>();
        for (OntClass c : classes) {
            OntClassVO vo = OntologyConverter.toClassVO(c);
            vo.setChildren(new ArrayList<>());
            vo.setPropertyCount((int) ontClassPropertyMapper.selectCountByQuery(
                    QueryWrapper.create().where(ONT_CLASS_PROPERTY.CLASS_ID.eq(c.getId()))));
            voMap.put(c.getId(), vo);
        }

        List<OntClassVO> roots = new ArrayList<>();
        for (OntClass c : classes) {
            OntClassVO vo = voMap.get(c.getId());
            if (c.getParentClassId() == null) {
                roots.add(vo);
            } else {
                OntClassVO parent = voMap.get(c.getParentClassId());
                if (parent != null) {
                    parent.getChildren().add(vo);
                } else {
                    roots.add(vo);
                }
            }
        }
        return roots;
    }

    /** 新增类：校验同本体下编码唯一 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public Long createClass(OntClassDTO dto) {
        OntClass existing = ontClassMapper.selectOneByQuery(QueryWrapper.create()
                .where(ONT_CLASS.ONTOLOGY_ID.eq(dto.getOntologyId()))
                .and(ONT_CLASS.CODE.eq(dto.getCode())));
        if (existing != null) {
            throw new BusinessException(OntologyErrorCode.CLASS_CODE_EXISTS);
        }
        OntClass entity = new OntClass();
        entity.setOntologyId(dto.getOntologyId());
        entity.setCode(dto.getCode());
        entity.setName(dto.getName());
        entity.setDescription(dto.getDescription());
        entity.setParentClassId(dto.getParentClassId());
        entity.setIcon(dto.getIcon());
        entity.setSortOrder(dto.getSortOrder());
        entity.setIdentityKeys(dto.getIdentityKeys());
        entity.setStatus(dto.getStatus() == null ? 1 : dto.getStatus());
        entity.setRemark(dto.getRemark());
        ontClassMapper.insert(entity);
        return entity.getId();
    }

    /** 更新类 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void updateClass(OntClassDTO dto) {
        OntClass entity = loadClass(dto.getId());
        OntClass sameCode = ontClassMapper.selectOneByQuery(QueryWrapper.create()
                .where(ONT_CLASS.ONTOLOGY_ID.eq(dto.getOntologyId()))
                .and(ONT_CLASS.CODE.eq(dto.getCode())));
        if (sameCode != null && !sameCode.getId().equals(dto.getId())) {
            throw new BusinessException(OntologyErrorCode.CLASS_CODE_EXISTS);
        }
        entity.setCode(dto.getCode());
        entity.setName(dto.getName());
        entity.setDescription(dto.getDescription());
        entity.setParentClassId(dto.getParentClassId());
        entity.setIcon(dto.getIcon());
        entity.setSortOrder(dto.getSortOrder());
        entity.setIdentityKeys(dto.getIdentityKeys());
        if (dto.getStatus() != null) {
            entity.setStatus(dto.getStatus());
        }
        entity.setRemark(dto.getRemark());
        ontClassMapper.update(entity);
    }

    /** 删除类：级联删除属性 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void deleteClass(Long classId) {
        loadClass(classId);
        ontClassPropertyMapper.deleteByQuery(
                QueryWrapper.create().where(ONT_CLASS_PROPERTY.CLASS_ID.eq(classId)));
        ontClassMapper.deleteById(classId);
    }

    // ==================== 属性 ====================

    /** 查询指定类的属性列表 */
    @Override
    @Transactional(readOnly = true)
    public List<OntPropertyVO> listProperties(Long classId) {
        List<OntClassProperty> list = ontClassPropertyMapper.selectListByQuery(QueryWrapper.create()
                .where(ONT_CLASS_PROPERTY.CLASS_ID.eq(classId))
                .orderBy(ONT_CLASS_PROPERTY.SORT_ORDER.asc()));
        return list.stream().map(OntologyConverter::toPropertyVO).collect(Collectors.toList());
    }

    /** 新增属性：校验同类下编码唯一 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public Long createProperty(OntPropertyDTO dto) {
        OntClassProperty existing = ontClassPropertyMapper.selectOneByQuery(QueryWrapper.create()
                .where(ONT_CLASS_PROPERTY.CLASS_ID.eq(dto.getClassId()))
                .and(ONT_CLASS_PROPERTY.CODE.eq(dto.getCode())));
        if (existing != null) {
            throw new BusinessException(OntologyErrorCode.PROPERTY_CODE_EXISTS);
        }
        OntClassProperty entity = new OntClassProperty();
        entity.setClassId(dto.getClassId());
        entity.setCode(dto.getCode());
        entity.setName(dto.getName());
        entity.setDataType(dto.getDataType());
        entity.setDescription(dto.getDescription());
        entity.setRequired(dto.getRequired());
        entity.setDefaultValue(dto.getDefaultValue());
        entity.setEnumValues(dto.getEnumValues());
        entity.setUnit(dto.getUnit());
        entity.setSortOrder(dto.getSortOrder());
        entity.setStatus(dto.getStatus() == null ? 1 : dto.getStatus());
        ontClassPropertyMapper.insert(entity);
        return entity.getId();
    }

    /** 更新属性 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void updateProperty(OntPropertyDTO dto) {
        OntClassProperty entity = loadProperty(dto.getId());
        OntClassProperty sameCode = ontClassPropertyMapper.selectOneByQuery(QueryWrapper.create()
                .where(ONT_CLASS_PROPERTY.CLASS_ID.eq(dto.getClassId()))
                .and(ONT_CLASS_PROPERTY.CODE.eq(dto.getCode())));
        if (sameCode != null && !sameCode.getId().equals(dto.getId())) {
            throw new BusinessException(OntologyErrorCode.PROPERTY_CODE_EXISTS);
        }
        entity.setCode(dto.getCode());
        entity.setName(dto.getName());
        entity.setDataType(dto.getDataType());
        entity.setDescription(dto.getDescription());
        entity.setRequired(dto.getRequired());
        entity.setDefaultValue(dto.getDefaultValue());
        entity.setEnumValues(dto.getEnumValues());
        entity.setUnit(dto.getUnit());
        entity.setSortOrder(dto.getSortOrder());
        if (dto.getStatus() != null) {
            entity.setStatus(dto.getStatus());
        }
        ontClassPropertyMapper.update(entity);
    }

    /** 删除属性 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void deleteProperty(Long id) {
        loadProperty(id);
        ontClassPropertyMapper.deleteById(id);
    }

    // ==================== 关系 ====================

    /** 查询本体下所有关系，附加源/目标类名 */
    @Override
    @Transactional(readOnly = true)
    public List<OntRelationVO> listRelations(Long ontologyId) {
        List<OntRelation> relations = ontRelationMapper.selectListByQuery(QueryWrapper.create()
                .where(ONT_RELATION.ONTOLOGY_ID.eq(ontologyId))
                .orderBy(ONT_RELATION.CREATE_TIME.asc()));
        // 批量获取类名映射
        List<OntClass> classes = ontClassMapper.selectListByQuery(
                QueryWrapper.create().where(ONT_CLASS.ONTOLOGY_ID.eq(ontologyId)));
        Map<Long, String> classNameMap = classes.stream()
                .collect(Collectors.toMap(OntClass::getId, OntClass::getName));

        return relations.stream().map(r -> {
            OntRelationVO vo = OntologyConverter.toRelationVO(r);
            vo.setSourceClassName(classNameMap.getOrDefault(r.getSourceClassId(), ""));
            vo.setTargetClassName(classNameMap.getOrDefault(r.getTargetClassId(), ""));
            return vo;
        }).collect(Collectors.toList());
    }

    /** 新增关系：校验同本体下编码唯一 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public Long createRelation(OntRelationDTO dto) {
        OntRelation existing = ontRelationMapper.selectOneByQuery(QueryWrapper.create()
                .where(ONT_RELATION.ONTOLOGY_ID.eq(dto.getOntologyId()))
                .and(ONT_RELATION.CODE.eq(dto.getCode())));
        if (existing != null) {
            throw new BusinessException(OntologyErrorCode.RELATION_CODE_EXISTS);
        }
        OntRelation entity = new OntRelation();
        entity.setOntologyId(dto.getOntologyId());
        entity.setCode(dto.getCode());
        entity.setName(dto.getName());
        entity.setDescription(dto.getDescription());
        entity.setSourceClassId(dto.getSourceClassId());
        entity.setTargetClassId(dto.getTargetClassId());
        entity.setCardinality(dto.getCardinality());
        entity.setJoinKeys(dto.getJoinKeys());
        entity.setInverseName(dto.getInverseName());
        entity.setStatus(dto.getStatus() == null ? 1 : dto.getStatus());
        entity.setRemark(dto.getRemark());
        ontRelationMapper.insert(entity);
        return entity.getId();
    }

    /** 更新关系 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void updateRelation(OntRelationDTO dto) {
        OntRelation entity = loadRelation(dto.getId());
        OntRelation sameCode = ontRelationMapper.selectOneByQuery(QueryWrapper.create()
                .where(ONT_RELATION.ONTOLOGY_ID.eq(dto.getOntologyId()))
                .and(ONT_RELATION.CODE.eq(dto.getCode())));
        if (sameCode != null && !sameCode.getId().equals(dto.getId())) {
            throw new BusinessException(OntologyErrorCode.RELATION_CODE_EXISTS);
        }
        entity.setCode(dto.getCode());
        entity.setName(dto.getName());
        entity.setDescription(dto.getDescription());
        entity.setSourceClassId(dto.getSourceClassId());
        entity.setTargetClassId(dto.getTargetClassId());
        entity.setCardinality(dto.getCardinality());
        entity.setJoinKeys(dto.getJoinKeys());
        entity.setInverseName(dto.getInverseName());
        if (dto.getStatus() != null) {
            entity.setStatus(dto.getStatus());
        }
        entity.setRemark(dto.getRemark());
        ontRelationMapper.update(entity);
    }

    /** 删除关系 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void deleteRelation(Long id) {
        loadRelation(id);
        ontRelationMapper.deleteById(id);
    }

    // ==================== 规则 ====================

    /** 查询本体下所有规则 */
    @Override
    @Transactional(readOnly = true)
    public List<OntRuleVO> listRules(Long ontologyId) {
        List<OntRule> list = ontRuleMapper.selectListByQuery(QueryWrapper.create()
                .where(ONT_RULE.ONTOLOGY_ID.eq(ontologyId))
                .orderBy(ONT_RULE.PRIORITY.asc()));
        return list.stream().map(OntologyConverter::toRuleVO).collect(Collectors.toList());
    }

    /** 新增规则 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public Long createRule(OntRuleDTO dto) {
        OntRule entity = new OntRule();
        entity.setOntologyId(dto.getOntologyId());
        entity.setCode(dto.getCode());
        entity.setName(dto.getName());
        entity.setDescription(dto.getDescription());
        entity.setRuleType(dto.getRuleType());
        entity.setExpression(dto.getExpression());
        entity.setPriority(dto.getPriority());
        entity.setEnabled(dto.getEnabled() == null ? 1 : dto.getEnabled());
        entity.setStatus(dto.getStatus() == null ? 1 : dto.getStatus());
        ontRuleMapper.insert(entity);
        return entity.getId();
    }

    /** 更新规则 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void updateRule(OntRuleDTO dto) {
        OntRule entity = loadRule(dto.getId());
        entity.setCode(dto.getCode());
        entity.setName(dto.getName());
        entity.setDescription(dto.getDescription());
        entity.setRuleType(dto.getRuleType());
        entity.setExpression(dto.getExpression());
        entity.setPriority(dto.getPriority());
        if (dto.getEnabled() != null) {
            entity.setEnabled(dto.getEnabled());
        }
        if (dto.getStatus() != null) {
            entity.setStatus(dto.getStatus());
        }
        ontRuleMapper.update(entity);
    }

    /** 删除规则 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void deleteRule(Long id) {
        loadRule(id);
        ontRuleMapper.deleteById(id);
    }

    // ==================== 动作 ====================

    /** 查询本体下所有动作，附加目标类名 */
    @Override
    @Transactional(readOnly = true)
    public List<OntActionVO> listActions(Long ontologyId) {
        List<OntAction> actions = ontActionMapper.selectListByQuery(QueryWrapper.create()
                .where(ONT_ACTION.ONTOLOGY_ID.eq(ontologyId))
                .orderBy(ONT_ACTION.CREATE_TIME.asc()));
        List<OntClass> classes = ontClassMapper.selectListByQuery(
                QueryWrapper.create().where(ONT_CLASS.ONTOLOGY_ID.eq(ontologyId)));
        Map<Long, String> classNameMap = classes.stream()
                .collect(Collectors.toMap(OntClass::getId, OntClass::getName));

        return actions.stream().map(a -> {
            OntActionVO vo = OntologyConverter.toActionVO(a);
            if (a.getTargetClassId() != null) {
                vo.setTargetClassName(classNameMap.getOrDefault(a.getTargetClassId(), ""));
            }
            return vo;
        }).collect(Collectors.toList());
    }

    /** 新增动作 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public Long createAction(OntActionDTO dto) {
        OntAction entity = new OntAction();
        entity.setOntologyId(dto.getOntologyId());
        entity.setCode(dto.getCode());
        entity.setName(dto.getName());
        entity.setDescription(dto.getDescription());
        entity.setActionType(dto.getActionType());
        entity.setTargetClassId(dto.getTargetClassId());
        entity.setParameters(dto.getParameters());
        entity.setPreconditions(dto.getPreconditions());
        entity.setEnabled(dto.getEnabled() == null ? 1 : dto.getEnabled());
        entity.setStatus(dto.getStatus() == null ? 1 : dto.getStatus());
        ontActionMapper.insert(entity);
        return entity.getId();
    }

    /** 更新动作 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void updateAction(OntActionDTO dto) {
        OntAction entity = loadAction(dto.getId());
        entity.setCode(dto.getCode());
        entity.setName(dto.getName());
        entity.setDescription(dto.getDescription());
        entity.setActionType(dto.getActionType());
        entity.setTargetClassId(dto.getTargetClassId());
        entity.setParameters(dto.getParameters());
        entity.setPreconditions(dto.getPreconditions());
        if (dto.getEnabled() != null) {
            entity.setEnabled(dto.getEnabled());
        }
        if (dto.getStatus() != null) {
            entity.setStatus(dto.getStatus());
        }
        ontActionMapper.update(entity);
    }

    /** 删除动作 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void deleteAction(Long id) {
        loadAction(id);
        ontActionMapper.deleteById(id);
    }

    // ==================== 详情 / 发布 / 导出 ====================

    /** 根据ID查询类详情，附加属性数量 */
    @Override
    @Transactional(readOnly = true)
    public OntClassVO getClassById(Long classId) {
        OntClass entity = loadClass(classId);
        OntClassVO vo = OntologyConverter.toClassVO(entity);
        vo.setPropertyCount((int) ontClassPropertyMapper.selectCountByQuery(
                QueryWrapper.create().where(ONT_CLASS_PROPERTY.CLASS_ID.eq(classId))));
        return vo;
    }

    /** 根据ID查询关系详情，附加源/目标类名 */
    @Override
    @Transactional(readOnly = true)
    public OntRelationVO getRelationById(Long relationId) {
        OntRelation r = loadRelation(relationId);
        OntRelationVO vo = OntologyConverter.toRelationVO(r);
        OntClass source = r.getSourceClassId() == null ? null : ontClassMapper.selectOneById(r.getSourceClassId());
        OntClass target = r.getTargetClassId() == null ? null : ontClassMapper.selectOneById(r.getTargetClassId());
        vo.setSourceClassName(source != null ? source.getName() : "");
        vo.setTargetClassName(target != null ? target.getName() : "");
        return vo;
    }

    /** 发布本体：置为启用状态 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void publish(Long id) {
        OntDefinition entity = loadEntity(id);
        entity.setStatus(1);
        ontDefinitionMapper.update(entity);
    }

    /** 导出本体：组装定义 + 类（含属性）+ 关系 + 规则 + 动作为可序列化结构 */
    @Override
    @Transactional(readOnly = true)
    public Map<String, Object> exportOntology(Long id) {
        OntologyVO ontology = getById(id);
        List<OntClass> classes = ontClassMapper.selectListByQuery(QueryWrapper.create()
                .where(ONT_CLASS.ONTOLOGY_ID.eq(id))
                .orderBy(ONT_CLASS.SORT_ORDER.asc()));

        List<Map<String, Object>> classExports = new ArrayList<>();
        for (OntClass c : classes) {
            Map<String, Object> cm = new LinkedHashMap<>();
            cm.put("id", String.valueOf(c.getId()));
            cm.put("code", c.getCode());
            cm.put("name", c.getName());
            cm.put("description", c.getDescription());
            cm.put("parentClassId", c.getParentClassId() == null ? null : String.valueOf(c.getParentClassId()));
            cm.put("icon", c.getIcon());
            cm.put("sortOrder", c.getSortOrder());
            cm.put("identityKeys", c.getIdentityKeys());
            cm.put("properties", listProperties(c.getId()));
            classExports.add(cm);
        }

        Map<String, Object> export = new LinkedHashMap<>();
        export.put("formatVersion", "1.0");
        export.put("ontology", ontology);
        export.put("classes", classExports);
        export.put("relations", listRelations(id));
        export.put("rules", listRules(id));
        export.put("actions", listActions(id));
        return export;
    }

    // ==================== 私有辅助 ====================

    private OntDefinition loadEntity(Long id) {
        OntDefinition entity = ontDefinitionMapper.selectOneById(id);
        if (entity == null) {
            throw new BusinessException(OntologyErrorCode.ONTOLOGY_NOT_FOUND);
        }
        return entity;
    }

    private OntClass loadClass(Long id) {
        OntClass entity = ontClassMapper.selectOneById(id);
        if (entity == null) {
            throw new BusinessException(OntologyErrorCode.CLASS_NOT_FOUND);
        }
        return entity;
    }

    private OntClassProperty loadProperty(Long id) {
        OntClassProperty entity = ontClassPropertyMapper.selectOneById(id);
        if (entity == null) {
            throw new BusinessException(OntologyErrorCode.PROPERTY_NOT_FOUND);
        }
        return entity;
    }

    private OntRelation loadRelation(Long id) {
        OntRelation entity = ontRelationMapper.selectOneById(id);
        if (entity == null) {
            throw new BusinessException(OntologyErrorCode.RELATION_NOT_FOUND);
        }
        return entity;
    }

    private OntRule loadRule(Long id) {
        OntRule entity = ontRuleMapper.selectOneById(id);
        if (entity == null) {
            throw new BusinessException(OntologyErrorCode.RULE_NOT_FOUND);
        }
        return entity;
    }

    private OntAction loadAction(Long id) {
        OntAction entity = ontActionMapper.selectOneById(id);
        if (entity == null) {
            throw new BusinessException(OntologyErrorCode.ACTION_NOT_FOUND);
        }
        return entity;
    }
}
