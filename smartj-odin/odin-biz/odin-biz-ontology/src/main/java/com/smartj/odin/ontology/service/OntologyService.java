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
package com.smartj.odin.ontology.service;

import com.smartj.common.core.domain.PageResult;
import com.smartj.odin.ontology.dto.OntActionDTO;
import com.smartj.odin.ontology.dto.OntClassDTO;
import com.smartj.odin.ontology.dto.OntPropertyDTO;
import com.smartj.odin.ontology.dto.OntRelationDTO;
import com.smartj.odin.ontology.dto.OntRuleDTO;
import com.smartj.odin.ontology.dto.OntologyDTO;
import com.smartj.odin.ontology.dto.OntologyQuery;
import com.smartj.odin.ontology.vo.OntActionVO;
import com.smartj.odin.ontology.vo.OntClassVO;
import com.smartj.odin.ontology.vo.OntPropertyVO;
import com.smartj.odin.ontology.vo.OntRelationVO;
import com.smartj.odin.ontology.vo.OntRuleVO;
import com.smartj.odin.ontology.vo.OntologyGraphVO;
import com.smartj.odin.ontology.vo.OntologyVO;
import com.smartj.odin.ontology.vo.VocabularyVO;

import java.util.List;
import java.util.Map;

/**
 * 本体服务接口。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public interface OntologyService {

    /**
     * 分页查询本体列表。
     *
     * @param query 查询条件
     * @return 分页结果
     */
    PageResult<OntologyVO> page(OntologyQuery query);

    /**
     * 查询全部启用本体（不分页，用于下拉选择/摘要）。
     *
     * @return 本体 VO 列表
     */
    List<OntologyVO> listAll();

    /**
     * 根据ID查询本体详情。
     *
     * @param id 主键ID
     * @return 本体 VO
     */
    OntologyVO getById(Long id);

    /**
     * 新增本体，返回主键ID。
     *
     * @param dto 本体 DTO
     * @return 新记录ID
     */
    Long create(OntologyDTO dto);

    /**
     * 更新本体。
     *
     * @param dto 本体 DTO
     */
    void update(OntologyDTO dto);

    /**
     * 删除本体（级联删除类、属性、关系、规则、动作）。
     *
     * @param id 主键ID
     */
    void delete(Long id);

    /**
     * 获取本体图谱（节点 + 边）。
     *
     * @param ontologyId 本体ID
     * @return 图谱 VO
     */
    OntologyGraphVO getGraph(Long ontologyId);

    /**
     * 获取本体词表（供 LLM 上下文注入）。
     *
     * @param ontologyId 本体ID
     * @return 词表 VO
     */
    VocabularyVO getVocabulary(Long ontologyId);

    /**
     * 查询本体下所有类（树形结构）。
     *
     * @param ontologyId 本体ID
     * @return 类 VO 树形列表
     */
    List<OntClassVO> listClasses(Long ontologyId);

    /**
     * 新增本体类，返回主键ID。
     *
     * @param dto 类 DTO
     * @return 新记录ID
     */
    Long createClass(OntClassDTO dto);

    /**
     * 更新本体类。
     *
     * @param dto 类 DTO
     */
    void updateClass(OntClassDTO dto);

    /**
     * 删除本体类（级联删除属性）。
     *
     * @param classId 类ID
     */
    void deleteClass(Long classId);

    /**
     * 查询指定类的属性列表。
     *
     * @param classId 类ID
     * @return 属性 VO 列表
     */
    List<OntPropertyVO> listProperties(Long classId);

    /**
     * 新增属性，返回主键ID。
     *
     * @param dto 属性 DTO
     * @return 新记录ID
     */
    Long createProperty(OntPropertyDTO dto);

    /**
     * 更新属性。
     *
     * @param dto 属性 DTO
     */
    void updateProperty(OntPropertyDTO dto);

    /**
     * 删除属性。
     *
     * @param id 属性ID
     */
    void deleteProperty(Long id);

    /**
     * 查询本体下所有关系（含源/目标类名）。
     *
     * @param ontologyId 本体ID
     * @return 关系 VO 列表
     */
    List<OntRelationVO> listRelations(Long ontologyId);

    /**
     * 新增关系，返回主键ID。
     *
     * @param dto 关系 DTO
     * @return 新记录ID
     */
    Long createRelation(OntRelationDTO dto);

    /**
     * 更新关系。
     *
     * @param dto 关系 DTO
     */
    void updateRelation(OntRelationDTO dto);

    /**
     * 删除关系。
     *
     * @param id 关系ID
     */
    void deleteRelation(Long id);

    /**
     * 查询本体下所有规则。
     *
     * @param ontologyId 本体ID
     * @return 规则 VO 列表
     */
    List<OntRuleVO> listRules(Long ontologyId);

    /**
     * 新增规则，返回主键ID。
     *
     * @param dto 规则 DTO
     * @return 新记录ID
     */
    Long createRule(OntRuleDTO dto);

    /**
     * 更新规则。
     *
     * @param dto 规则 DTO
     */
    void updateRule(OntRuleDTO dto);

    /**
     * 删除规则。
     *
     * @param id 规则ID
     */
    void deleteRule(Long id);

    /**
     * 查询本体下所有动作（含目标类名）。
     *
     * @param ontologyId 本体ID
     * @return 动作 VO 列表
     */
    List<OntActionVO> listActions(Long ontologyId);

    /**
     * 新增动作，返回主键ID。
     *
     * @param dto 动作 DTO
     * @return 新记录ID
     */
    Long createAction(OntActionDTO dto);

    /**
     * 更新动作。
     *
     * @param dto 动作 DTO
     */
    void updateAction(OntActionDTO dto);

    /**
     * 删除动作。
     *
     * @param id 动作ID
     */
    void deleteAction(Long id);

    /**
     * 根据ID查询本体类详情（含属性数量）。
     *
     * @param classId 类ID
     * @return 类 VO
     */
    OntClassVO getClassById(Long classId);

    /**
     * 根据ID查询关系详情（含源/目标类名）。
     *
     * @param relationId 关系ID
     * @return 关系 VO
     */
    OntRelationVO getRelationById(Long relationId);

    /**
     * 发布本体（置为启用状态）。
     *
     * @param id 本体ID
     */
    void publish(Long id);

    /**
     * 导出本体为可序列化结构（定义 + 类/属性 + 关系 + 规则 + 动作）。
     *
     * @param id 本体ID
     * @return 导出结构
     */
    Map<String, Object> exportOntology(Long id);
}
