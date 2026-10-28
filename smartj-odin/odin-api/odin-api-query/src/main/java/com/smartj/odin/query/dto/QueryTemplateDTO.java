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
package com.smartj.odin.query.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;
import lombok.Data;

import java.io.Serial;
import java.io.Serializable;

/**
 * 查询模板 DTO（新增/更新入参）。
 *
 * <p>模板固化"自然语言 → 本体逻辑 SQL"的映射，可由 AI 产出后经 HITL 审阅晋升，用于收敛问数抖动、提升确定性。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class QueryTemplateDTO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID（更新时必填） */
    private Long id;

    /** 模板名称 */
    @NotBlank(message = "模板名称不能为空")
    @Size(max = 128, message = "模板名称长度不能超过128")
    private String name;

    /** 描述 */
    private String description;

    /** 本体ID */
    private Long ontologyId;

    /** 本体逻辑 SQL */
    private String logicalSql;

    /** 参数定义（JSON） */
    private String parameters;

    /** 标签（JSON 数组） */
    private String tags;

    /** 状态 0=禁用 1=启用 */
    private Integer status;
}
