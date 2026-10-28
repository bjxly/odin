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
package com.smartj.odin.datasource.entity;

import com.mybatisflex.annotation.Table;
import com.smartj.common.core.domain.BaseSimpleEntity;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.io.Serial;
import java.time.LocalDateTime;

/**
 * 数据源测试记录实体。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
@EqualsAndHashCode(callSuper = true)
@Table("odin_datasource_test")
public class DatasourceTest extends BaseSimpleEntity {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 数据源ID */
    private Long datasourceId;
    /** 测试时间 */
    private LocalDateTime testTime;
    /** 是否成功 0=否 1=是 */
    private Integer success;
    /** 耗时毫秒 */
    private Integer latencyMs;
    /** 数据库版本 */
    private String dbVersion;
    /** 错误信息 */
    private String errorMessage;
}
