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
package com.smartj.odin.insight.error;

import com.smartj.common.core.exception.ErrorCode;

/**
 * 洞察域错误码（区间 7500~7599）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public enum InsightErrorCode implements ErrorCode {

    /** 监控规则不存在 */
    MONITOR_NOT_FOUND(7500, "监控规则不存在"),
    /** 告警不存在 */
    ALERT_NOT_FOUND(7501, "告警不存在"),
    /** 告警已处理 */
    ALERT_ALREADY_HANDLED(7502, "告警已处理"),
    /** 监控执行失败 */
    MONITOR_RUN_FAILED(7503, "监控执行失败"),
    /** 变更记录不存在 */
    CHANGE_NOT_FOUND(7510, "变更记录不存在"),
    /** 变更回滚失败 */
    CHANGE_REVERT_FAILED(7511, "变更回滚失败"),
    /** 洞察看板不存在 */
    DASHBOARD_NOT_FOUND(7520, "洞察看板不存在");

    /** 错误码 */
    private final int code;
    /** 错误信息 */
    private final String message;

    InsightErrorCode(int code, String message) {
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
