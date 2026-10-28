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
import java.util.ArrayList;
import java.util.List;

/**
 * 推理试运行结果 VO。
 *
 * <p>与前端 {@code src/types/odin/reason.ts} 的 ReasonTraceResult 契约一致，
 * 包含规则触发、虚拟类展开、派生过滤与推导关系四组明细。
 * TODO: 契约对齐桩类型，真实推理引擎待 P3 接入。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class ReasonTraceResult implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 规则触发记录 */
    private List<RuleFire> rulesFired = new ArrayList<>();
    /** 虚拟类展开步骤 */
    private List<ExpansionStep> expansions = new ArrayList<>();
    /** 派生过滤条件 */
    private List<DerivedFilter> derivedFilters = new ArrayList<>();
    /** 推导关系 */
    private List<InferredRelation> inferredRelations = new ArrayList<>();

    /** 规则触发记录 */
    @Data
    public static class RuleFire implements Serializable {
        @Serial
        private static final long serialVersionUID = 1L;
        private String ruleId;
        private String ruleName;
        private String ruleType;
        private boolean conditionResult;
        private boolean derived;
        private String action;
        private String detail;
    }

    /** 虚拟类展开步骤 */
    @Data
    public static class ExpansionStep implements Serializable {
        @Serial
        private static final long serialVersionUID = 1L;
        private String from;
        private String to;
        private String reason;
    }

    /** 派生过滤条件 */
    @Data
    public static class DerivedFilter implements Serializable {
        @Serial
        private static final long serialVersionUID = 1L;
        private String property;
        private String op;
        private Object value;
        private String source;
    }

    /** 推导关系 */
    @Data
    public static class InferredRelation implements Serializable {
        @Serial
        private static final long serialVersionUID = 1L;
        private String name;
        private String label;
        private String fromClass;
        private String toClass;
    }
}
