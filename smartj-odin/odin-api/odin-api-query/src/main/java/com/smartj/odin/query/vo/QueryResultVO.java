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
import java.util.List;
import java.util.Map;

/**
 * 智能问数执行结果 VO。
 *
 * <p>承载结果集（列名 + 行数据）、执行耗时、编译得到的逻辑/物理 SQL 以及溯源标识，供前端渲染与追溯。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class QueryResultVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 是否执行成功 */
    private Boolean success;
    /** 结果列名（有序） */
    private List<String> columns;
    /** 结果行数据（列名 → 值） */
    private List<Map<String, Object>> rows;
    /** 结果行数 */
    private Integer rowCount;
    /** 执行耗时毫秒 */
    private Integer latencyMs;
    /** 本体逻辑 SQL */
    private String logicalSql;
    /** 编译下推后的物理 SQL */
    private String physicalSql;
    /** 全链路溯源ID */
    private String traceId;
    /** 失败时的错误信息 */
    private String errorMessage;
}
