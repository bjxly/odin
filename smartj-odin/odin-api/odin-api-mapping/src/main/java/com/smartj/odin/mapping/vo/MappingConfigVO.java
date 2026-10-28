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
package com.smartj.odin.mapping.vo;

import lombok.Data;

import java.io.Serial;
import java.io.Serializable;

/**
 * 映射配置 VO（展示，含关联名称）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class MappingConfigVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID */
    private String id;
    /** 所属本体ID */
    private String ontologyId;
    /** 本体名称 */
    private String ontologyName;
    /** 映射目标类ID */
    private String classId;
    /** 类名称 */
    private String className;
    /** 数据源ID */
    private String datasourceId;
    /** 数据源名称 */
    private String datasourceName;
    /** 物理表名 */
    private String tableName;
    /** 模式名 */
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
    private String remark;
    /** 创建时间 */
    private String createTime;
    /** 更新时间 */
    private String updateTime;
}
