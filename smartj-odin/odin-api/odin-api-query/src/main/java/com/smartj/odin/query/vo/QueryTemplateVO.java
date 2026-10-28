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
package com.smartj.odin.query.vo;

import lombok.Data;

import java.io.Serial;
import java.io.Serializable;

/**
 * 查询模板 VO（展示）。
 *
 * <p>ID 类字段以字符串输出，规避前端 JS 大整数（雪花ID）精度丢失。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class QueryTemplateVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID */
    private String id;
    /** 模板名称 */
    private String name;
    /** 描述 */
    private String description;
    /** 本体ID */
    private String ontologyId;
    /** 本体名称 */
    private String ontologyName;
    /** 本体逻辑 SQL */
    private String logicalSql;
    /** 参数定义（JSON） */
    private String parameters;
    /** 标签（JSON 数组） */
    private String tags;
    /** 使用次数 */
    private Integer usageCount;
    /** 状态 0=禁用 1=启用 */
    private Integer status;
    /** 创建时间 */
    private String createTime;
    /** 更新时间 */
    private String updateTime;
}
