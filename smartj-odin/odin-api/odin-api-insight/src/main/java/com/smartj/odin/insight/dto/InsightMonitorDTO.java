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

/**
 * 洞察监控规则 DTO（新增/更新入参）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class InsightMonitorDTO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID（更新时必填） */
    private Long id;

    /** 监控名称 */
    @NotBlank(message = "监控名称不能为空")
    @Size(max = 128, message = "监控名称长度不能超过128")
    private String name;

    /** 描述 */
    private String description;

    /** 本体ID */
    private Long ontologyId;

    /** 规则表达式（阈值判定，如 count > 0） */
    private String ruleExpression;

    /** 调度 cron 表达式 */
    @Size(max = 64, message = "调度cron长度不能超过64")
    private String scheduleCron;

    /** 阈值配置（JSON） */
    private String thresholds;

    /** 通知渠道（JSON） */
    private String notificationChannels;

    /** 是否启用 0=停用 1=启用 */
    private Integer enabled;

    /** 状态 0=禁用 1=启用 */
    private Integer status;

    /** 备注 */
    @Size(max = 500, message = "备注长度不能超过500")
    private String remark;
}
