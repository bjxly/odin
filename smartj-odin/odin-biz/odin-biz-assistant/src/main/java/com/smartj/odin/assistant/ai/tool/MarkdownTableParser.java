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

import org.springframework.util.StringUtils;

import java.util.ArrayList;
import java.util.List;

/**
 * 正文 Markdown 分析表解析器：把 AI 分析结论表确定性转成图表可消费的 {@link DatasetSnapshot}。
 *
 * <p>设计动机：图表若直接消费原始查询数据集，会与模型跨源合并/筛选后的结论口径不一致
 * （如"仅 VIP 客户"结论配了"全部客户"图）。分析表本身就是结论口径，且表格由模型写入正文一次、
 * 此处只解析不加工，数字无二次幻觉风险。
 *
 * <p>解析规则：表头行 + {@code |---|---|} 分隔行 + 数据行；单元格做确定性归一化
 * （剥离 Markdown 加粗、千分位逗号、百分号、空白与 {@code —} 转空串），使数值列可被前端启发式识别。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public final class MarkdownTableParser {

    /** 单轮最多解析的分析表数量（防异常长文） */
    private static final int MAX_TABLES = 5;

    /** 单表最多保留的数据行数 */
    private static final int MAX_ROWS = 100;

    /** dataId 前缀：chart_hint 以此识别表数据集 */
    public static final String TABLE_ID_PREFIX = "tbl-";

    private MarkdownTableParser() {
    }

    /**
     * 从正文提取全部 Markdown 表格并转为数据集快照。
     *
     * @param content  助手回复全文（含或不含 chart_hint 标记均可）
     * @param traceId  本轮追踪ID（用于生成稳定的 dataId 前缀）
     * @return 表数据集列表（id 形如 {@code tbl-{traceId前8}-N}，N 从 1 起）
     */
    public static List<DatasetSnapshot> extract(String content, String traceId) {
        List<DatasetSnapshot> tables = new ArrayList<>();
        if (!StringUtils.hasText(content)) {
            return tables;
        }
        String head = traceId == null ? "anon" : (traceId.length() >= 8 ? traceId.substring(0, 8) : traceId);
        String[] lines = content.replace("\r\n", "\n").replace('\r', '\n').split("\n", -1);
        int idx = 0;
        int tableNo = 0;
        while (idx < lines.length && tableNo < MAX_TABLES) {
            int[] end = matchTable(lines, idx);
            if (end == null) {
                idx++;
                continue;
            }
            // end[0]=分隔行下标（表头在 end[0]-1 不成立时已在 matchTable 内保证），数据行到 end[1)
            List<String> header = splitRow(lines[end[0] - 1]);
            List<List<String>> rows = new ArrayList<>();
            for (int i = end[0] + 1; i < end[1] && rows.size() < MAX_ROWS; i++) {
                List<String> cells = splitRow(lines[i]);
                if (cells.size() != header.size()) {
                    break;
                }
                rows.add(cells);
            }
            if (header.size() >= 2 && !rows.isEmpty()) {
                tableNo++;
                String dataId = TABLE_ID_PREFIX + head + "-" + tableNo;
                List<String> columns = new ArrayList<>(header.size());
                for (String h : header) {
                    columns.add(normalize(h));
                }
                List<List<String>> rowVals = new ArrayList<>(rows.size());
                for (List<String> r : rows) {
                    List<String> line = new ArrayList<>(r.size());
                    for (String c : r) {
                        line.add(normalize(c));
                    }
                    rowVals.add(line);
                }
                tables.add(new DatasetSnapshot(dataId, "分析表 " + tableNo, columns, rowVals, rowVals.size(), false, null));
            }
            idx = end[1];
        }
        return tables;
    }

    /**
     * 自 {@code start+1} 行起尝试匹配「表头行 + 分隔行」结构。
     *
     * @return {@code [分隔行下标, 数据行结束下标(不含))}；不匹配返回 null
     */
    private static int[] matchTable(String[] lines, int start) {
        int sep = start + 1;
        if (sep >= lines.length || !isSeparatorRow(lines[start]) || !isSeparator(lines[sep])) {
            return null;
        }
        int i = sep + 1;
        int firstCols = splitRow(lines[start]).size();
        while (i < lines.length) {
            String line = lines[i];
            if (!isDataRow(line) || splitRow(line).size() != firstCols) {
                break;
            }
            i++;
        }
        // 至少要有一行数据才构成有意义的分析表
        if (i <= sep + 1) {
            return null;
        }
        return new int[]{sep, i};
    }

    /** 候选表头行：含竖线、至少两个非空单元格、非分隔行 */
    private static boolean isSeparatorRow(String line) {
        return isDataRow(line) && !isSeparator(line);
    }

    /** 数据/表头行形态：trim 后以 | 开头或同时含两个 | */
    private static boolean isDataRow(String line) {
        String s = line.trim();
        if (s.isEmpty() || !s.contains("|")) {
            return false;
        }
        return s.startsWith("|") || s.chars().filter(c -> c == '|').count() >= 2;
    }

    /** 分隔行：仅由 | : - 空格构成且至少含一个 - */
    private static boolean isSeparator(String line) {
        String s = line.trim();
        if (s.isEmpty() || !s.contains("-") || !s.contains("|")) {
            return false;
        }
        return s.chars().allMatch(c -> c == '|' || c == ':' || c == '-' || c == ' ');
    }

    /** 拆分表格行为单元格列表（去掉首尾竖线产生的空串） */
    private static List<String> splitRow(String line) {
        String s = line.trim();
        if (s.startsWith("|")) {
            s = s.substring(1);
        }
        if (s.endsWith("|")) {
            s = s.substring(0, s.length() - 1);
        }
        List<String> cells = new ArrayList<>();
        for (String part : s.split("\\|", -1)) {
            cells.add(part.trim());
        }
        return cells;
    }

    /** 单元格归一化：剥离加粗、千分位、百分号；破折号占位转空串（前端按空值处理，不当 0）。 */
    private static String normalize(String cell) {
        if (cell == null) {
            return "";
        }
        String s = cell.trim();
        s = s.replace("**", "").replace("__", "");
        if (s.equals("—") || s.equals("-") || s.equals("–")) {
            return "";
        }
        s = s.replaceAll("(?<=\\d),(?=\\d{3}(?:\\D|$))", "");
        if (s.matches("[-+]?\\d+(?:\\.\\d+)?%")) {
            s = s.substring(0, s.length() - 1);
        }
        return s.trim();
    }
}
