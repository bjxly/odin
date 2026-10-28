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

import org.junit.jupiter.api.Test;

import java.util.List;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * {@link MarkdownTableParser} 回归测试。
 *
 * <p>守护「正文分析表 → 图表数据集」的确定性解析：表头/分隔行/数据行识别、单元格归一化
 * （加粗、千分位、百分号、破折号占位转空串），以及多表编号。解析错误会直接导致前端出图口径偏离结论。
 */
class MarkdownTableParserTest {

    @Test
    void parsesVipSummaryTable() {
        String content = """
                结论如下：

                | 客户编号 | 客户名称 | 客户等级 | 订单数 | 订单总金额 |
                |---|---|---|---|---|
                | 1 | 张伟 | VIP | 3 | 140,400.00 |
                | 3 | 王强 | VIP | 2 | 31,300.00 |
                | **合计** | — | — | **8** | **391,500.00** |

                补充说明。
                """;
        List<DatasetSnapshot> tables = MarkdownTableParser.extract(content, "trace-abcdef1234");
        assertEquals(1, tables.size());
        DatasetSnapshot t = tables.get(0);
        assertTrue(t.dataId().startsWith("tbl-"));
        assertEquals(List.of("客户编号", "客户名称", "客户等级", "订单数", "订单总金额"), t.columns());
        assertEquals(3, t.rows().size());
        // 千分位剥离、加粗剥离、破折号转空串
        assertEquals(List.of("1", "张伟", "VIP", "3", "140400.00"), t.rows().get(0));
        assertEquals(List.of("合计", "", "", "8", "391500.00"), t.rows().get(2));
    }

    @Test
    void stripsPercentAndHandlesMultipleTables() {
        String content = """
                | 指标 | 本月 | 上月 | 环比 |
                | --- | --- | --- | --- |
                | 销售额 | 328,140.00 | 389,600.00 | -15.78% |

                中间文字

                | 类别 | 数量 |
                |---|---|
                | A | 10 |
                """;
        List<DatasetSnapshot> tables = MarkdownTableParser.extract(content, "trace-abcdef1234");
        assertEquals(2, tables.size());
        assertEquals("-15.78", tables.get(0).rows().get(0).get(3));
        assertEquals(List.of("类别", "数量"), tables.get(1).columns());
    }

    @Test
    void ignoresNonTables() {
        String content = "这是一段普通文本，包含 a | b 但不是表格。\n\n- 列表项 X | Y";
        assertTrue(MarkdownTableParser.extract(content, "trace").isEmpty());
    }
}
