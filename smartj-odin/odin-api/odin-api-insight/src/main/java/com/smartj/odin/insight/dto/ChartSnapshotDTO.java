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
package com.smartj.odin.insight.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;
import lombok.Data;

import java.io.Serial;
import java.io.Serializable;
import java.util.List;

/**
 * 图表快照 DTO（AI 助手「另存为洞察看板」入参）。
 *
 * <p>前端把当前数据集（列/行）、图型与图表配置序列化后提交，后端落库为洞察看板。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class ChartSnapshotDTO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 看板名称 */
    @NotBlank(message = "看板名称不能为空")
    @Size(max = 128, message = "看板名称长度不能超过128")
    private String name;

    /** 所属本体ID（以字符串下发规避 JS 大数精度，后端转 Long） */
    private Long ontologyId;

    /** 图表标题 */
    @Size(max = 256, message = "图表标题长度不能超过256")
    private String title;

    /** 图型: bar/line/pie */
    private String chartType;

    /** 数据集列头 */
    private List<String> columns;

    /** 数据集行（单元格一律字符串） */
    private List<List<String>> rows;

    /** 图表配置快照（选择/聚合等，原样存储） */
    private Object chartOption;
}
