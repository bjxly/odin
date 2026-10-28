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
package com.smartj.odin.datasource.vo;

import lombok.Data;

import java.io.Serial;
import java.io.Serializable;

/**
 * 列元数据 VO。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class SchemaColumnVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID */
    private Long id;
    /** 列名 */
    private String columnName;
    /** 源库原始类型 */
    private String dataType;
    /** 归一类型 */
    private String canonicalType;
    /** 是否可空 0=否 1=是 */
    private Integer nullable;
    /** 默认值 */
    private String defaultValue;
    /** 列注释 */
    private String comment;
    /** 最大长度 */
    private Integer maxLength;
    /** 数值精度 */
    private Integer numericPrecision;
    /** 数值小数位 */
    private Integer numericScale;
    /** 列序号 */
    private Integer ordinalPosition;
    /** 是否主键 0=否 1=是 */
    private Integer isPrimaryKey;
    /** 是否唯一 0=否 1=是 */
    private Integer isUnique;
    /** 列同义词(JSON) */
    private String synonyms;
    /** 状态 0=禁用 1=启用 */
    private Integer status;
}
