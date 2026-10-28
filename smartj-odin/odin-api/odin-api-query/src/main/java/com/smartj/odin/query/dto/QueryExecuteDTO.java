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
package com.smartj.odin.query.dto;

import lombok.Data;

import java.io.Serial;
import java.io.Serializable;
import java.util.Map;

/**
 * 智能问数执行入参 DTO。
 *
 * <p>支持两种入口：①自然语言问题（naturalLanguage），服务端解析为逻辑 SQL；
 * ②直接给定本体逻辑 SQL（logicalSql）+ 参数（params）执行。二者至少提供其一，
 * 因此在契约层均不做强制非空校验，由服务层择一处理。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class QueryExecuteDTO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 自然语言问题（与 logicalSql 二选一） */
    private String naturalLanguage;

    /** 本体逻辑 SQL（与自然语言二选一，前端"执行逻辑SQL问数"入口使用） */
    private String logicalSql;

    /** 逻辑 SQL 参数（命名参数绑定，可选） */
    private Map<String, Object> params;

    /** 本体ID（限定语义上下文，可选） */
    private Long ontologyId;

    /** 助手会话ID（可选，用于串联多轮问数） */
    private Long sessionId;
}
