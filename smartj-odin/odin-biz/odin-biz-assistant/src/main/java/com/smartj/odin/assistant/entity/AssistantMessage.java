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
package com.smartj.odin.assistant.entity;

import com.mybatisflex.annotation.Table;
import com.smartj.common.core.domain.BaseSimpleEntity;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.io.Serial;

/**
 * 助手消息实体。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
@EqualsAndHashCode(callSuper = true)
@Table("odin_assistant_message")
public class AssistantMessage extends BaseSimpleEntity {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 会话ID */
    private Long sessionId;
    /** 角色：user/assistant/tool/system */
    private String role;
    /** 消息内容 */
    private String content;
    /** 内容类型：text/markdown/table 等 */
    private String contentType;
    /** 元数据(JSON) */
    private String metadata;
    /** 消耗 token 数 */
    private Integer tokensUsed;
    /** 生成耗时毫秒 */
    private Integer latencyMs;
}
