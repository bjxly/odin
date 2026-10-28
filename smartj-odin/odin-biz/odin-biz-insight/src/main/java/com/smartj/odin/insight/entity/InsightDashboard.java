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
package com.smartj.odin.insight.entity;

import com.mybatisflex.annotation.Table;
import com.smartj.common.core.domain.BaseSimpleEntity;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.io.Serial;

/**
 * 洞察看板实体。
 *
 * <p>承载 AI 助手对话中「另存为洞察看板」的图表快照：数据集列/行、图型与图表配置序列化落库，
 * 供洞察看板列表回看。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
@EqualsAndHashCode(callSuper = true)
@Table("odin_insight_dashboard")
public class InsightDashboard extends BaseSimpleEntity {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 看板名称 */
    private String name;
    /** 本体ID */
    private Long ontologyId;
    /** 图表标题 */
    private String title;
    /** 图型: bar/line/pie */
    private String chartType;
    /** 数据集列头（JSON） */
    private String columnsJson;
    /** 数据集行（字符串二维数组，JSON） */
    private String rowsJson;
    /** 图表配置快照（JSON） */
    private String chartOption;
    /** 来源 */
    private String source;
    /** 状态 0=禁用 1=启用 */
    private Integer status;
    /** 备注 */
    private String remark;
}
