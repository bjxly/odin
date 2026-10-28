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
package com.smartj.odin.query.vo;

import lombok.Data;

import java.io.Serial;
import java.io.Serializable;
import java.util.List;

/**
 * 查询试运行结果 VO。
 *
 * <p>返回逻辑 SQL 是否通过校验、编译得到的物理 SQL、预估行数与告警列表，不真正执行数据读取。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class QueryDryRunVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 是否校验通过 */
    private Boolean valid;
    /** 本体逻辑 SQL */
    private String logicalSql;
    /** 编译下推后的物理 SQL */
    private String physicalSql;
    /** 预估行数 */
    private Integer estimatedRows;
    /** 校验告警（如 JOIN 扇出、成本超阈值等） */
    private List<String> warnings;
    /** 校验失败时的错误信息 */
    private String errorMessage;
}
