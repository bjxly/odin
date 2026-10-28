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
package com.smartj.odin.query.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

import java.io.Serial;
import java.io.Serializable;

/**
 * 查询试运行（dry-run）入参 DTO。
 *
 * <p>对给定的本体逻辑 SQL 做静态校验与物理 SQL 编译预览，不真正执行，用于建模与审阅阶段的验证。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class QueryDryRunDTO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 本体逻辑 SQL */
    @NotBlank(message = "逻辑SQL不能为空")
    private String logicalSql;

    /** 本体ID（限定语义上下文，可选） */
    private Long ontologyId;
}
