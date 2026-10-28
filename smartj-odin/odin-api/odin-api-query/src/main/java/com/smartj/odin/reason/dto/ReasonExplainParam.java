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
package com.smartj.odin.reason.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

import java.io.Serial;
import java.io.Serializable;
import java.util.List;

/**
 * 推理解释请求参数。
 *
 * <p>与前端 {@code src/types/odin/reason.ts} 的 ReasonExplainParam 契约一致，
 * 用于对指定本体下的某个类施加过滤条件并触发推理试运行。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class ReasonExplainParam implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 本体ID */
    @NotNull(message = "本体ID不能为空")
    private Long ontologyId;

    /** 类编码 */
    @NotBlank(message = "类编码不能为空")
    private String classCode;

    /** 过滤条件 */
    private List<Filter> filters;

    /** 过滤条件项 */
    @Data
    public static class Filter implements Serializable {
        @Serial
        private static final long serialVersionUID = 1L;
        /** 属性 */
        private String property;
        /** 操作符 */
        private String op;
        /** 值 */
        private Object value;
    }
}
