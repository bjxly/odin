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
package com.smartj.odin.ontology.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;
import lombok.Data;

import java.io.Serial;
import java.io.Serializable;

/**
 * 本体类 DTO（新增/更新入参）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class OntClassDTO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID（更新时必填） */
    private Long id;

    /** 所属本体ID */
    @NotNull(message = "所属本体ID不能为空")
    private Long ontologyId;

    /** 类编码 */
    @NotBlank(message = "类编码不能为空")
    @Size(max = 64, message = "类编码长度不能超过64")
    private String code;

    /** 类名称 */
    @NotBlank(message = "类名称不能为空")
    @Size(max = 128, message = "类名称长度不能超过128")
    private String name;

    /** 描述 */
    private String description;

    /** 父类ID */
    private Long parentClassId;

    /** 图标 */
    @Size(max = 64, message = "图标长度不能超过64")
    private String icon;

    /** 排序号 */
    private Integer sortOrder;

    /** 身份键（JSON） */
    private String identityKeys;

    /** 状态 0=禁用 1=启用 */
    private Integer status;

    /** 备注 */
    private String remark;
}
