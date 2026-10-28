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
package com.smartj.odin.mapping.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;
import lombok.Data;

import java.io.Serial;
import java.io.Serializable;

/**
 * 映射配置 DTO（新增/更新入参）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class MappingConfigDTO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID（更新时必填） */
    private Long id;

    /** 所属本体ID */
    @NotNull(message = "所属本体ID不能为空")
    private Long ontologyId;

    /** 映射目标类ID */
    @NotNull(message = "映射目标类ID不能为空")
    private Long classId;

    /** 数据源ID */
    @NotNull(message = "数据源ID不能为空")
    private Long datasourceId;

    /** 物理表名 */
    @NotBlank(message = "物理表名不能为空")
    @Size(max = 128, message = "物理表名长度不能超过128")
    private String tableName;

    /** 模式名 */
    @Size(max = 128, message = "模式名长度不能超过128")
    private String schemaName;

    /** 列映射（JSON） */
    private String columnMappings;

    /** 过滤条件 */
    private String filterCondition;

    /** 优先级 */
    private Integer priority;

    /** 状态 0=禁用 1=启用 */
    private Integer status;

    /** 备注 */
    @Size(max = 500, message = "备注长度不能超过500")
    private String remark;
}
