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
 * 本体类属性 VO（展示）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class OntPropertyVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID */
    private String id;
    /** 所属类ID */
    private String classId;
    /** 属性编码 */
    private String code;
    /** 属性名称 */
    private String name;
    /** 数据类型 */
    private String dataType;
    /** 描述 */
    private String description;
    /** 是否必填 0=否 1=是 */
    private Integer required;
    /** 默认值 */
    private String defaultValue;
    /** 枚举值（JSON） */
    private String enumValues;
    /** 单位 */
    private String unit;
    /** 排序号 */
    private Integer sortOrder;
    /** 状态 0=禁用 1=启用 */
    private Integer status;
    /** 创建时间 */
    private String createTime;
}
