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
package com.smartj.odin.insight.vo;

import lombok.Data;

import java.io.Serial;
import java.io.Serializable;

/**
 * 洞察看板 VO（展示）。
 *
 * <p>列/行/图表配置以 JSON 字符串透传，前端解析后重建图表卡片。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class InsightDashboardVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID */
    private String id;
    /** 看板名称 */
    private String name;
    /** 本体ID */
    private String ontologyId;
    /** 图表标题 */
    private String title;
    /** 图型: bar/line/pie */
    private String chartType;
    /** 数据集列头（JSON 字符串） */
    private String columnsJson;
    /** 数据集行（JSON 字符串） */
    private String rowsJson;
    /** 图表配置快照（JSON 字符串） */
    private String chartOption;
    /** 来源 */
    private String source;
    /** 状态 0=禁用 1=启用 */
    private Integer status;
    /** 备注 */
    private String remark;
    /** 创建时间 */
    private String createTime;
}
