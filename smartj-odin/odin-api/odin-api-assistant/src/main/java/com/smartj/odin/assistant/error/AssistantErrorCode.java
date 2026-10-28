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
package com.smartj.odin.assistant.error;

import com.smartj.common.core.exception.ErrorCode;

/**
 * 助手域错误码（区间 7400~7499）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public enum AssistantErrorCode implements ErrorCode {

    /** 会话不存在 */
    SESSION_NOT_FOUND(7400, "会话不存在"),
    /** 消息发送失败 */
    MESSAGE_SEND_FAILED(7401, "消息发送失败"),
    /** AI草稿不存在 */
    DRAFT_NOT_FOUND(7410, "AI草稿不存在"),
    /** 草稿已审阅 */
    DRAFT_ALREADY_REVIEWED(7411, "草稿已审阅"),
    /** 草稿发布失败 */
    DRAFT_PUBLISH_FAILED(7412, "草稿发布失败");

    /** 错误码 */
    private final int code;
    /** 错误信息 */
    private final String message;

    AssistantErrorCode(int code, String message) {
        this.code = code;
        this.message = message;
    }

    @Override
    public int getCode() {
        return code;
    }

    @Override
    public String getMessage() {
        return message;
    }
}
