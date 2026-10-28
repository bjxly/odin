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
 * 查询历史实体。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
@EqualsAndHashCode(callSuper = true)
@Table("odin_query_history")
public class QueryHistory extends BaseSimpleEntity {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 助手会话ID */
    private Long sessionId;
    /** 用户ID */
    private Long userId;
    /** 自然语言问题 */
    private String naturalLanguage;
    /** 本体逻辑 SQL */
    private String logicalSql;
    /** 物理 SQL */
    private String physicalSql;
    /** 结果摘要 */
    private String resultSummary;
    /** 执行耗时毫秒 */
    private Integer latencyMs;
    /** 执行状态：success/error */
    private String status;
    /** 错误信息 */
    private String errorMessage;
}
