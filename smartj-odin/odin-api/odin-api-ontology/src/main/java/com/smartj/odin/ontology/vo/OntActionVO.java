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

/**
 * 本体动作 VO（展示，含目标类名）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class OntActionVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID */
    private String id;
    /** 所属本体ID */
    private String ontologyId;
    /** 动作编码 */
    private String code;
    /** 动作名称 */
    private String name;
    /** 描述 */
    private String description;
    /** 动作类型 */
    private String actionType;
    /** 目标类ID */
    private String targetClassId;
    /** 目标类名称 */
    private String targetClassName;
    /** 参数（JSON） */
    private String parameters;
    /** 前置条件（JSON） */
    private String preconditions;
    /** 是否启用 0=否 1=是 */
    private Integer enabled;
    /** 状态 0=禁用 1=启用 */
    private Integer status;
    /** 创建时间 */
    private String createTime;
}
