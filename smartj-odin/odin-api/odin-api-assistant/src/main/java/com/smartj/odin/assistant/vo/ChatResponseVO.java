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
package com.smartj.odin.assistant.vo;

import lombok.Data;

import java.io.Serial;
import java.io.Serializable;

/**
 * 助手对话响应 VO。
 *
 * <p>P2 为一次性完整响应；P3 接入 Agentscope-java 后改为 SSE 流式，按 {@code done} 标记结束。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class ChatResponseVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 会话ID */
    private String sessionId;
    /** 助手消息ID */
    private String messageId;
    /** 回复内容 */
    private String content;
    /** 内容类型：text/markdown/table 等 */
    private String contentType;
    /** 是否结束（流式场景标记最后一帧） */
    private Boolean done;
    /** 本轮 AI 产生的草稿ID（无草稿时为 null；多草稿时取首个，其余由前端刷新列表获得） */
    private String draftId;
    /** 本轮草稿明细项数量（无草稿时为 null） */
    private Integer draftItemCount;
}
