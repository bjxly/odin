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

import io.agentscope.core.tool.Tool;
import io.agentscope.core.tool.ToolParam;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/**
 * 本体问数工具（Agentscope @Tool）：把本体逻辑 SQL 交给 {@link OntologySqlExecutor} 执行。
 *
 * <p>非单例 Bean：每次对话请求按当前本体ID新建实例并注册进 Toolkit，保证工具闭包绑定正确本体。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public class OntologyQueryTool {

    /** 结果展示最大行数（超出截断，避免 token 爆炸） */
    private static final int DISPLAY_MAX_ROWS = 50;

    /** 侧信道快照行上限（供图表消费，超出截断并标记） */
    private static final int DATASET_MAX_ROWS = 200;

    private final OntologySqlExecutor executor;
    private final Long ontologyId;
    private final DatasetSink datasetSink;

    public OntologyQueryTool(OntologySqlExecutor executor, Long ontologyId) {
        this(executor, ontologyId, null);
    }

    public OntologyQueryTool(OntologySqlExecutor executor, Long ontologyId, DatasetSink datasetSink) {
        this.executor = executor;
        this.ontologyId = ontologyId;
        this.datasetSink = datasetSink;
    }

    /**
     * 执行本体逻辑 SQL 并返回结果集文本。
     *
     * @param logicalSql 本体逻辑 SQL（仅使用本体类/属性编码）
     * @return 结果文本（物理SQL + 列 + 行），失败时返回错误说明
     */
    @Tool(name = "run_ontology_query",
            description = "针对当前本体执行一条逻辑 SQL 查询并返回真实数据。SQL 中只能使用本体词表里的类编码（作为表）"
                    + "与属性编码（作为列），不要使用物理表名/列名。返回编译后的物理SQL与查询结果行。")
    public String runOntologyQuery(
            @ToolParam(name = "logical_sql", required = true,
                    description = "本体逻辑 SQL，仅 SELECT，表用类编码、列用属性编码，例如：SELECT name FROM Customer WHERE level = 'VIP' LIMIT 5")
            String logicalSql) {
        OntologySqlExecutor.ExecOutcome outcome = executor.execute(ontologyId, logicalSql, 200);
        if (!outcome.isSuccess()) {
            return "查询失败：" + outcome.getError();
        }
        // 侧信道：把结构化结果（列全量、行截断、单元格转字符串）随流下发，供前端图表卡片消费
        publishDataset(outcome);
        StringBuilder sb = new StringBuilder();
        sb.append("物理SQL: ").append(outcome.getPhysicalSql()).append('\n');
        List<String> columns = outcome.getColumns();
        sb.append("列: ").append(String.join(", ", columns)).append('\n');
        List<Map<String, Object>> rows = outcome.getRows();
        sb.append("行数: ").append(rows.size()).append('\n');
        int show = Math.min(rows.size(), DISPLAY_MAX_ROWS);
        for (int i = 0; i < show; i++) {
            Map<String, Object> row = rows.get(i);
            StringBuilder line = new StringBuilder();
            for (String col : columns) {
                if (line.length() > 0) {
                    line.append(" | ");
                }
                line.append(col).append('=').append(row.get(col));
            }
            sb.append(line).append('\n');
        }
        if (rows.size() > show) {
            sb.append("...（仅展示前 ").append(show).append(" 行，共 ").append(rows.size()).append(" 行）\n");
        }
        return sb.toString();
    }

    /**
     * 将执行结果转为 {@link DatasetSnapshot} 并投递到侧信道（未配置 sink 时跳过）。
     *
     * @param outcome 执行结果（已成功）
     */
    private void publishDataset(OntologySqlExecutor.ExecOutcome outcome) {
        if (datasetSink == null) {
            return;
        }
        List<String> columns = outcome.getColumns();
        List<Map<String, Object>> rawRows = outcome.getRows();
        int total = rawRows == null ? 0 : rawRows.size();
        int limit = Math.min(total, DATASET_MAX_ROWS);
        List<List<String>> rows = new ArrayList<>(limit);
        for (int i = 0; i < limit; i++) {
            Map<String, Object> row = rawRows.get(i);
            List<String> line = new ArrayList<>(columns.size());
            for (String col : columns) {
                Object v = row.get(col);
                line.add(v == null ? "" : String.valueOf(v));
            }
            rows.add(line);
        }
        datasetSink.publish(new DatasetSnapshot(
                datasetSink.nextDataId(),
                null,
                new ArrayList<>(columns),
                rows,
                total,
                total > limit,
                outcome.getPhysicalSql()));
    }
}
