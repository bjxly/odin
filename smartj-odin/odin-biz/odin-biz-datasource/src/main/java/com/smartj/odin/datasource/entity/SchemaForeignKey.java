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
package com.smartj.odin.datasource.entity;

import com.mybatisflex.annotation.Table;
import com.smartj.common.core.domain.BaseSimpleEntity;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.io.Serial;

/**
 * 外键元数据实体。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
@EqualsAndHashCode(callSuper = true)
@Table("odin_schema_foreign_key")
public class SchemaForeignKey extends BaseSimpleEntity {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 表元数据ID */
    private Long tableId;
    /** 约束名 */
    private String constraintName;
    /** 列名 */
    private String columnName;
    /** 引用模式 */
    private String refSchema;
    /** 引用表 */
    private String refTable;
    /** 引用列 */
    private String refColumn;
    /** 删除级联策略 */
    private String onDelete;
    /** 更新级联策略 */
    private String onUpdate;
}
