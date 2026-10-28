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
package com.smartj.odin.datasource.error;

import com.smartj.common.core.exception.ErrorCode;

/**
 * 数据源域错误码（区间 7000~7099）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public enum DatasourceErrorCode implements ErrorCode {

    /** 数据源不存在 */
    DATASOURCE_NOT_FOUND(7000, "数据源不存在"),
    /** 数据源编码已存在 */
    DATASOURCE_CODE_EXISTS(7001, "数据源编码已存在"),
    /** 连接测试失败 */
    DATASOURCE_TEST_FAILED(7002, "数据源连接测试失败"),
    /** 元数据同步失败 */
    DATASOURCE_SYNC_FAILED(7003, "元数据同步失败");

    /** 错误码 */
    private final int code;
    /** 错误信息 */
    private final String message;

    DatasourceErrorCode(int code, String message) {
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
