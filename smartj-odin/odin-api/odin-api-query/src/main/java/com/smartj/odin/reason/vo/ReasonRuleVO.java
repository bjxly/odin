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
package com.smartj.odin.reason.vo;

import lombok.Data;

import java.io.Serial;
import java.io.Serializable;

/**
 * 推理规则 VO（展示）。
 *
 * <p>与前端 {@code src/types/odin/reason.ts} 的 ReasonRuleVO 契约一致。
 * TODO: 当前为契约对齐桩类型，待后续接入独立的推理规则 Entity/Service/Mapper 后填充真实数据。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class ReasonRuleVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID */
    private String id;
    /** 本体ID */
    private String ontologyId;
    /** 规则名称 */
    private String name;
    /** 规则编码 */
    private String code;
    /** 规则描述 */
    private String description;
    /** 规则类型 */
    private String ruleType;
    /** 规则表达式 */
    private String expression;
    /** 条件 JSON */
    private String conditionJson;
    /** 动作 JSON */
    private String actionJson;
    /** 优先级 */
    private Integer priority;
    /** 是否启用 0=禁用 1=启用 */
    private Integer enabled;
    /** 状态 0=禁用 1=启用 */
    private Integer status;
    /** 创建时间 */
    private String createTime;
    /** 更新时间 */
    private String updateTime;
}
