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
package com.smartj.odin.ontology.converter;

import com.smartj.common.core.util.BeanCopyUtil;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.smartj.odin.ontology.entity.OntAction;
import com.smartj.odin.ontology.entity.OntClass;
import com.smartj.odin.ontology.entity.OntClassProperty;
import com.smartj.odin.ontology.entity.OntDefinition;
import com.smartj.odin.ontology.entity.OntRelation;
import com.smartj.odin.ontology.entity.OntRule;
import com.smartj.odin.ontology.vo.OntActionVO;
import com.smartj.odin.ontology.vo.OntClassVO;
import com.smartj.odin.ontology.vo.OntPropertyVO;
import com.smartj.odin.ontology.vo.OntRelationVO;
import com.smartj.odin.ontology.vo.OntRuleVO;
import com.smartj.odin.ontology.vo.OntologyVO;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.List;

/**
 * 本体域 Entity → VO 转换器。
 *
 * <p>VO 中 ID 字段统一为 String，时间字段统一为格式化字符串，
 * 因此不能直接依赖 BeanCopyUtil 的同名同类型拷贝，需手动转换。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public final class OntologyConverter {

    private static final DateTimeFormatter TIME_FMT = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

    /** preset_questions 等 JSON 列解析器（线程安全，复用单例） */
    private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();

    private OntologyConverter() {
    }

    /**
     * OntDefinition → OntologyVO。
     *
     * @param entity 本体定义实体
     * @return 本体 VO
     */
    public static OntologyVO toVO(OntDefinition entity) {
        if (entity == null) {
            return null;
        }
        OntologyVO vo = BeanCopyUtil.copyProperties(entity, OntologyVO.class);
        vo.setId(toStr(entity.getId()));
        vo.setPresetQuestions(parseStringList(entity.getPresetQuestions()));
        vo.setCreateTime(fmtTime(entity.getCreateTime()));
        vo.setUpdateTime(fmtTime(entity.getUpdateTime()));
        return vo;
    }

    /**
     * OntClass → OntClassVO。
     *
     * @param entity 本体类实体
     * @return 本体类 VO
     */
    public static OntClassVO toClassVO(OntClass entity) {
        if (entity == null) {
            return null;
        }
        OntClassVO vo = BeanCopyUtil.copyProperties(entity, OntClassVO.class);
        vo.setId(toStr(entity.getId()));
        vo.setOntologyId(toStr(entity.getOntologyId()));
        vo.setParentClassId(toStr(entity.getParentClassId()));
        vo.setCreateTime(fmtTime(entity.getCreateTime()));
        return vo;
    }

    /**
     * OntClassProperty → OntPropertyVO。
     *
     * @param entity 属性实体
     * @return 属性 VO
     */
    public static OntPropertyVO toPropertyVO(OntClassProperty entity) {
        if (entity == null) {
            return null;
        }
        OntPropertyVO vo = BeanCopyUtil.copyProperties(entity, OntPropertyVO.class);
        vo.setId(toStr(entity.getId()));
        vo.setClassId(toStr(entity.getClassId()));
        vo.setCreateTime(fmtTime(entity.getCreateTime()));
        return vo;
    }

    /**
     * OntRelation → OntRelationVO。
     *
     * @param entity 关系实体
     * @return 关系 VO
     */
    public static OntRelationVO toRelationVO(OntRelation entity) {
        if (entity == null) {
            return null;
        }
        OntRelationVO vo = BeanCopyUtil.copyProperties(entity, OntRelationVO.class);
        vo.setId(toStr(entity.getId()));
        vo.setOntologyId(toStr(entity.getOntologyId()));
        vo.setSourceClassId(toStr(entity.getSourceClassId()));
        vo.setTargetClassId(toStr(entity.getTargetClassId()));
        vo.setCreateTime(fmtTime(entity.getCreateTime()));
        return vo;
    }

    /**
     * OntRule → OntRuleVO。
     *
     * @param entity 规则实体
     * @return 规则 VO
     */
    public static OntRuleVO toRuleVO(OntRule entity) {
        if (entity == null) {
            return null;
        }
        OntRuleVO vo = BeanCopyUtil.copyProperties(entity, OntRuleVO.class);
        vo.setId(toStr(entity.getId()));
        vo.setOntologyId(toStr(entity.getOntologyId()));
        vo.setCreateTime(fmtTime(entity.getCreateTime()));
        return vo;
    }

    /**
     * OntAction → OntActionVO。
     *
     * @param entity 动作实体
     * @return 动作 VO
     */
    public static OntActionVO toActionVO(OntAction entity) {
        if (entity == null) {
            return null;
        }
        OntActionVO vo = BeanCopyUtil.copyProperties(entity, OntActionVO.class);
        vo.setId(toStr(entity.getId()));
        vo.setOntologyId(toStr(entity.getOntologyId()));
        vo.setTargetClassId(toStr(entity.getTargetClassId()));
        vo.setCreateTime(fmtTime(entity.getCreateTime()));
        return vo;
    }

    /** Long → String，null 安全 */
    private static String toStr(Long value) {
        return value == null ? null : String.valueOf(value);
    }

    /** JSON 字符串数组 → List<String>，空值/非法降级为空列表 */
    private static List<String> parseStringList(String json) {
        if (json == null || json.isBlank()) {
            return List.of();
        }
        try {
            List<String> list = OBJECT_MAPPER.readValue(json, new TypeReference<List<String>>() {
            });
            return list == null ? List.of() : list;
        } catch (Exception e) {
            return List.of();
        }
    }

    /** LocalDateTime → 格式化字符串，null 安全 */
    private static String fmtTime(LocalDateTime time) {
        return time == null ? null : time.format(TIME_FMT);
    }
}
