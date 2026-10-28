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
package com.smartj.odin.datasource.dto;

import com.smartj.common.core.domain.BasePageQuery;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.io.Serial;

/**
 * 连接器分页查询条件。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
@EqualsAndHashCode(callSuper = true)
public class ConnectorQuery extends BasePageQuery {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 名称（模糊） */
    private String name;

    /** 编码（模糊） */
    private String code;

    /** 数据库类型（精确） */
    private String dbType;

    /** 状态 0=禁用 1=启用 */
    private Integer status;
}
