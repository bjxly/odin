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
package com.smartj.odin.ontology.dto;

import com.smartj.common.core.domain.BasePageQuery;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.io.Serial;

/**
 * 本体变更日志查询条件。
 *
 * <p>字段与前端 {@code smartj-odin-ui/src/types/odin/trace.ts#ChangeLogQuery} 对齐。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
@EqualsAndHashCode(callSuper = true)
public class ChangeLogQuery extends BasePageQuery {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 本体ID（精确） */
    private Long ontologyId;

    /** 实体类型（精确）：ontology/class/property/relation/rule/action/mapping */
    private String entityType;

    /** 操作类型（精确）：create/update/delete/publish/revert */
    private String action;
}
