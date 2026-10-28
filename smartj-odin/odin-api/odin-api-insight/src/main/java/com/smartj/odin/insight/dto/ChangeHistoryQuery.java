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
package com.smartj.odin.insight.dto;

import com.smartj.common.core.domain.BasePageQuery;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.io.Serial;

/**
 * 变更历史分页查询条件。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
@EqualsAndHashCode(callSuper = true)
public class ChangeHistoryQuery extends BasePageQuery {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 实体类型：ontology/class/property/relation/rule/action/mapping/datasource */
    private String entityType;

    /** 实体ID（精确） */
    private Long entityId;

    /** 变更动作：create/update/delete/publish/revert */
    private String action;
}
