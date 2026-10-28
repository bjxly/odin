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
package com.smartj.odin.mapping.error;

import com.smartj.common.core.exception.ErrorCode;

/**
 * 映射域错误码（区间 7200~7299）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public enum MappingErrorCode implements ErrorCode {

    /** 映射配置不存在 */
    MAPPING_NOT_FOUND(7200, "映射配置不存在"),
    /** 该映射配置已存在 */
    MAPPING_ALREADY_EXISTS(7201, "该映射配置已存在"),
    /** 映射校验失败 */
    MAPPING_VALIDATE_FAILED(7202, "映射校验失败");

    /** 错误码 */
    private final int code;
    /** 错误信息 */
    private final String message;

    MappingErrorCode(int code, String message) {
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
