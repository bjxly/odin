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
package com.smartj.odin.query.error;

import com.smartj.common.core.exception.ErrorCode;

/**
 * 查询域错误码（区间 7300~7399）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public enum QueryErrorCode implements ErrorCode {

    /** 查询模板不存在 */
    QUERY_TEMPLATE_NOT_FOUND(7300, "查询模板不存在"),
    /** 查询执行失败 */
    QUERY_EXECUTE_FAILED(7301, "查询执行失败"),
    /** 查询历史不存在 */
    QUERY_HISTORY_NOT_FOUND(7302, "查询历史不存在"),
    /** 自然语言解析失败 */
    QUERY_PARSE_FAILED(7303, "自然语言解析失败");

    /** 错误码 */
    private final int code;
    /** 错误信息 */
    private final String message;

    QueryErrorCode(int code, String message) {
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
