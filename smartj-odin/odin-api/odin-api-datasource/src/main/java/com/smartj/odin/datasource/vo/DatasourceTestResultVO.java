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
package com.smartj.odin.datasource.vo;

import lombok.Data;

import java.io.Serial;
import java.io.Serializable;

/**
 * 数据源连接测试结果 VO。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class DatasourceTestResultVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 是否成功 */
    private Boolean success;
    /** 耗时毫秒 */
    private Long latencyMs;
    /** 数据库版本 */
    private String dbVersion;
    /** 错误信息 */
    private String errorMessage;
}
