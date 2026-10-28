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
package com.smartj.odin.insight.entity;

import com.mybatisflex.annotation.Table;
import com.smartj.common.core.domain.BaseSimpleEntity;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.io.Serial;

/**
 * 变更历史实体。
 *
 * <p>记录本体/映射/数据源等各类实体的变更快照，支持审计追溯与回滚。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
@EqualsAndHashCode(callSuper = true)
@Table("odin_change_history")
public class ChangeHistory extends BaseSimpleEntity {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 实体类型：ontology/class/property/relation/rule/action/mapping/datasource */
    private String entityType;
    /** 实体ID */
    private Long entityId;
    /** 变更动作：create/update/delete/publish/revert */
    private String action;
    /** 变更前快照（JSON） */
    private String beforeSnapshot;
    /** 变更后快照（JSON） */
    private String afterSnapshot;
    /** 操作人 */
    private String operator;
    /** 变更原因 */
    private String reason;
}
