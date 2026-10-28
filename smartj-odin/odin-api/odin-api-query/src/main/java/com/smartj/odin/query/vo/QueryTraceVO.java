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

/**
 * 查询全链路溯源 VO（展示）。
 *
 * <p>按步骤展开一次问数的执行链路（解析、编译、下推、各源查询、合并等），供前端可视化追溯。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class QueryTraceVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 溯源ID */
    private String id;
    /** 关联查询历史ID */
    private String historyId;
    /** 溯源步骤列表（按 stepIndex 升序） */
    private List<TraceStep> steps;

    /**
     * 溯源步骤。
     */
    @Data
    public static class TraceStep implements Serializable {

        @Serial
        private static final long serialVersionUID = 1L;

        /** 步骤序号 */
        private Integer stepIndex;
        /** 步骤类型：parse/compile/pushdown/source_query/merge 等 */
        private String stepType;
        /** 数据源ID */
        private String datasourceId;
        /** 数据源名称 */
        private String datasourceName;
        /** 该步骤执行的 SQL 文本 */
        private String sqlText;
        /** 该步骤返回行数 */
        private Integer rowCount;
        /** 该步骤耗时毫秒 */
        private Integer latencyMs;
        /** 步骤详情（JSON 或说明文本） */
        private String details;
    }
}
