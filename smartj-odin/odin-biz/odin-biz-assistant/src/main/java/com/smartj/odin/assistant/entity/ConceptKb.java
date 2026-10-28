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
package com.smartj.odin.assistant.entity;

import com.mybatisflex.annotation.Table;
import com.smartj.common.core.domain.BaseSimpleEntity;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.io.Serial;

/**
 * 概念知识库实体。
 *
 * <p>沉淀业务概念/术语的规范释义，为问数与助手对话提供语义消歧与向量检索依据。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
@EqualsAndHashCode(callSuper = true)
@Table("odin_concept_kb")
public class ConceptKb extends BaseSimpleEntity {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 本体ID（NULL 表示全局） */
    private Long ontologyId;
    /** 概念编码 */
    private String conceptCode;
    /** 概念名称 */
    private String conceptName;
    /** 概念释义/摘要 */
    private String summary;
    /** 向量化状态：pending/embedded/failed */
    private String embeddingStatus;
    /** 来源类型：manual/ai/import */
    private String sourceType;
    /** 状态 0=禁用 1=启用 */
    private Integer status;
}
