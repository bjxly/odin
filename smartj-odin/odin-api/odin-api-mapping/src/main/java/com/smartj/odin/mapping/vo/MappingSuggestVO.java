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
package com.smartj.odin.mapping.vo;

import lombok.Data;

import java.io.Serial;
import java.io.Serializable;

/**
 * 映射建议 VO（基于名称相似度推荐的列-属性映射）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class MappingSuggestVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 物理列名 */
    private String columnName;
    /** 属性名称 */
    private String propertyName;
    /** 属性编码 */
    private String propertyCode;
    /** 置信度 (0~1) */
    private Double confidence;
    /** 匹配原因 */
    private String matchReason;
}
