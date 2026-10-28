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
 * 本体图谱 VO（节点 + 边，用于可视化渲染）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class OntologyGraphVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 图节点列表 */
    private List<GraphNode> nodes;
    /** 图边列表 */
    private List<GraphEdge> edges;

    /**
     * 图节点（对应本体类）。
     */
    @Data
    public static class GraphNode implements Serializable {

        @Serial
        private static final long serialVersionUID = 1L;

        /** 节点ID（类ID） */
        private String id;
        /** 节点标签（类名称） */
        private String label;
        /** 节点类型，固定为 "class" */
        private String type;
        /** 图标 */
        private String icon;
        /** 属性编码列表 */
        private List<String> properties;
        /** 属性明细（编码+名称+类型，供图谱节点展示） */
        private List<NodeProperty> propertyDetails;
    }

    /**
     * 图节点属性明细。
     */
    @Data
    public static class NodeProperty implements Serializable {

        @Serial
        private static final long serialVersionUID = 1L;

        /** 属性编码 */
        private String code;
        /** 属性名称 */
        private String name;
        /** 数据类型 */
        private String dataType;
    }

    /**
     * 图边（对应本体关系）。
     */
    @Data
    public static class GraphEdge implements Serializable {

        @Serial
        private static final long serialVersionUID = 1L;

        /** 边ID（关系ID） */
        private String id;
        /** 源节点ID */
        private String source;
        /** 目标节点ID */
        private String target;
        /** 边标签（关系名称） */
        private String label;
        /** 基数 */
        private String cardinality;
    }
}
