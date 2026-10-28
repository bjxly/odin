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
package com.smartj.odin.common.udap.model;

import lombok.Data;

import java.io.Serial;
import java.io.Serializable;
import java.util.List;
import java.util.Map;

/**
 * 连接器查询结果（列名 + 行数据）。
 *
 * <p>用于联邦执行层「单源方言直查」路径返回物理查询结果集。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class QueryResult implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 列名列表（按结果集顺序） */
    private List<String> columns;

    /** 行数据（列名 → 值） */
    private List<Map<String, Object>> rows;

    /** 实际返回行数 */
    private int rowCount;

    /** 是否因达到 maxRows 而被截断 */
    private boolean truncated;
}
