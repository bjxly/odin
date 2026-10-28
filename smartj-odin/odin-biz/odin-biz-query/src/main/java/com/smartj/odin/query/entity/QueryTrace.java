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
package com.smartj.odin.query.entity;

import com.mybatisflex.annotation.Table;
import com.smartj.common.core.domain.BaseSimpleEntity;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.io.Serial;

/**
 * 查询溯源步骤实体。
 *
 * <p>一条查询历史对应多条溯源步骤，按 {@code stepIndex} 升序还原执行链路。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
@EqualsAndHashCode(callSuper = true)
@Table("odin_query_trace")
public class QueryTrace extends BaseSimpleEntity {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 关联查询历史ID */
    private Long historyId;
    /** 步骤序号 */
    private Integer stepIndex;
    /** 步骤类型：parse/compile/pushdown/source_query/merge 等 */
    private String stepType;
    /** 数据源ID */
    private Long datasourceId;
    /** 该步骤执行的 SQL 文本 */
    private String sqlText;
    /** 该步骤返回行数 */
    private Integer rowCount;
    /** 该步骤耗时毫秒 */
    private Integer latencyMs;
    /** 步骤详情(JSON 或说明文本) */
    private String details;
}
