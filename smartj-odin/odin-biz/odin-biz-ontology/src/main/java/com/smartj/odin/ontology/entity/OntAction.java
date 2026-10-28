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
package com.smartj.odin.ontology.entity;

import com.mybatisflex.annotation.Table;
import com.smartj.common.core.domain.BaseSimpleEntity;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.io.Serial;

/**
 * 本体动作实体。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
@EqualsAndHashCode(callSuper = true)
@Table("odin_ont_action")
public class OntAction extends BaseSimpleEntity {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 所属本体ID */
    private Long ontologyId;
    /** 动作编码 */
    private String code;
    /** 动作名称 */
    private String name;
    /** 描述 */
    private String description;
    /** 动作类型 */
    private String actionType;
    /** 目标类ID */
    private Long targetClassId;
    /** 参数（JSON） */
    private String parameters;
    /** 前置条件（JSON） */
    private String preconditions;
    /** 是否启用 0=否 1=是 */
    private Integer enabled;
    /** 状态 0=禁用 1=启用 */
    private Integer status;
}
