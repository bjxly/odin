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
package com.smartj.odin.query.entity;

import com.mybatisflex.annotation.Table;
import com.smartj.common.core.domain.BaseSimpleEntity;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.io.Serial;
import java.time.LocalDateTime;

/**
 * 意图缓存实体。
 *
 * <p>以 {@code hash(规范化NL + 本体版本)} 为键缓存逻辑 SQL 与物理执行计划哈希，
 * 用于收敛同问同答的确定性（默认关闭，P3 视需要开启）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
@EqualsAndHashCode(callSuper = true)
@Table("odin_intent_cache")
public class IntentCache extends BaseSimpleEntity {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 规范化自然语言哈希 */
    private String nlHash;
    /** 本体版本 */
    private String ontologyVersion;
    /** 缓存的本体逻辑 SQL */
    private String logicalSql;
    /** 物理执行计划哈希 */
    private String physicalPlanHash;
    /** 命中次数 */
    private Integer hitCount;
    /** 最近命中时间 */
    private LocalDateTime lastHitTime;
    /** 状态 0=禁用 1=启用 */
    private Integer status;
}
