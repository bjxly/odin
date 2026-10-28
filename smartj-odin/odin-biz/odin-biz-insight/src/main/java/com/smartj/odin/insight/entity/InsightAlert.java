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
import java.time.LocalDateTime;

/**
 * 洞察告警实例实体。
 *
 * <p>监控规则命中阈值后生成的告警记录，支持人工处理闭环。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
@EqualsAndHashCode(callSuper = true)
@Table("odin_insight_alert")
public class InsightAlert extends BaseSimpleEntity {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 监控规则ID */
    private Long monitorId;
    /** 触发时间 */
    private LocalDateTime triggerTime;
    /** 严重程度：info/warning/critical */
    private String severity;
    /** 告警标题 */
    private String title;
    /** 告警详情 */
    private String detail;
    /** 触发时指标值(JSON) */
    private String metricValues;
    /** 是否已处理 0=未处理 1=已处理 */
    private Integer handled;
    /** 处理人 */
    private String handledBy;
    /** 处理时间 */
    private LocalDateTime handledTime;
    /** 处理说明 */
    private String handleNote;
}
