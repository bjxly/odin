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
package com.smartj.odin.assistant.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

import java.io.Serial;
import java.io.Serializable;

/**
 * 助手流式对话请求 DTO。
 *
 * <p>专用于 SSE 端点：前端会话ID为字符串（新建会话时传空串 {@code ""}），
 * 故此处以 {@link String} 承载，避免空串到 {@link Long} 的类型转换歧义；
 * 由 Controller 负责将非空串解析为 Long 后转交 {@link ChatRequestDTO}。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class ChatStreamRequestDTO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 用户消息内容 */
    @NotBlank(message = "消息内容不能为空")
    private String message;

    /** 会话ID字符串（为空串则新建会话） */
    private String sessionId;

    /** 本体ID（限定语义上下文） */
    private Long ontologyId;
}
