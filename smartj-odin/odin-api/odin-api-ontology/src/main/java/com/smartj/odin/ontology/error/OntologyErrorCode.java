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
package com.smartj.odin.ontology.error;

import com.smartj.common.core.exception.ErrorCode;

/**
 * 本体域错误码（区间 7100~7199）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public enum OntologyErrorCode implements ErrorCode {

    /** 本体不存在 */
    ONTOLOGY_NOT_FOUND(7100, "本体不存在"),
    /** 本体编码已存在 */
    ONTOLOGY_CODE_EXISTS(7101, "本体编码已存在"),
    /** 本体类不存在 */
    CLASS_NOT_FOUND(7110, "本体类不存在"),
    /** 类编码已存在 */
    CLASS_CODE_EXISTS(7111, "类编码已存在"),
    /** 属性不存在 */
    PROPERTY_NOT_FOUND(7120, "属性不存在"),
    /** 属性编码已存在 */
    PROPERTY_CODE_EXISTS(7121, "属性编码已存在"),
    /** 关系不存在 */
    RELATION_NOT_FOUND(7130, "关系不存在"),
    /** 关系编码已存在 */
    RELATION_CODE_EXISTS(7131, "关系编码已存在"),
    /** 规则不存在 */
    RULE_NOT_FOUND(7140, "规则不存在"),
    /** 动作不存在 */
    ACTION_NOT_FOUND(7150, "动作不存在");

    /** 错误码 */
    private final int code;
    /** 错误信息 */
    private final String message;

    OntologyErrorCode(int code, String message) {
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
