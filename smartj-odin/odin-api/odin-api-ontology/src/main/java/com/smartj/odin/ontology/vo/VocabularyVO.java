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
package com.smartj.odin.ontology.vo;

import lombok.Data;

import java.io.Serial;
import java.io.Serializable;
import java.util.List;

/**
 * 本体词表 VO（供 LLM 上下文注入，描述本体逻辑 Schema）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class VocabularyVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 本体ID */
    private String ontologyId;
    /** 本体名称 */
    private String ontologyName;
    /** 版本号 */
    private String version;
    /** 类词表列表 */
    private List<ClassVocab> classes;
    /** 关系词表列表 */
    private List<RelationVocab> relations;

    /**
     * 类词表条目。
     */
    @Data
    public static class ClassVocab implements Serializable {

        @Serial
        private static final long serialVersionUID = 1L;

        /** 类编码 */
        private String code;
        /** 类名称 */
        private String name;
        /** 描述 */
        private String description;
        /** 属性列表 */
        private List<PropertyVocab> properties;
    }

    /**
     * 属性词表条目。
     */
    @Data
    public static class PropertyVocab implements Serializable {

        @Serial
        private static final long serialVersionUID = 1L;

        /** 属性编码 */
        private String code;
        /** 属性名称 */
        private String name;
        /** 数据类型 */
        private String dataType;
        /** 描述 */
        private String description;
    }

    /**
     * 关系词表条目。
     */
    @Data
    public static class RelationVocab implements Serializable {

        @Serial
        private static final long serialVersionUID = 1L;

        /** 关系编码 */
        private String code;
        /** 关系名称 */
        private String name;
        /** 源类编码 */
        private String sourceClass;
        /** 目标类编码 */
        private String targetClass;
        /** 基数 */
        private String cardinality;
        /** 关联键（JSON） */
        private String joinKeys;
    }
}
