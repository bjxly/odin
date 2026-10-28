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
package com.smartj.odin.assistant.ai.tool;

import java.util.List;

/**
 * 数据集快照：问数工具执行成功后，经侧信道（{@link DatasetSink}）随流下发的结构化结果。
 *
 * <p>与前端 {@code QueryDataset} 对齐：{@code rows} 仅数据行、不含表头，单元格统一转字符串
 * （规避 JS 大数精度与前端二次格式化）；行数上限截断并标记 {@code truncated}。
 *
 * @param dataId      数据集ID（{@code ds-{traceId前8}-{序号}}）
 * @param title       标题（可空，前端图表卡片默认名）
 * @param columns     表头列名（全量）
 * @param rows        数据行（每行与 columns 等长，单元格为字符串；最多 200 行）
 * @param rowCount    原始总行数
 * @param truncated   是否发生行截断
 * @param physicalSql 编译后的物理 SQL（可空，供溯源）
 * @author SmartG-Odin
 * @since 1.0.0
 */
public record DatasetSnapshot(
        String dataId,
        String title,
        List<String> columns,
        List<List<String>> rows,
        int rowCount,
        boolean truncated,
        String physicalSql
) {
}
