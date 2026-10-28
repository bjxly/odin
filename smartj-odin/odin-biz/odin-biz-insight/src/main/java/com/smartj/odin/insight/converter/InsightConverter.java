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
package com.smartj.odin.insight.converter;

import com.smartj.common.core.util.BeanCopyUtil;
import com.smartj.odin.insight.entity.ChangeHistory;
import com.smartj.odin.insight.entity.InsightAlert;
import com.smartj.odin.insight.entity.InsightDashboard;
import com.smartj.odin.insight.entity.InsightMonitor;
import com.smartj.odin.insight.vo.ChangeHistoryVO;
import com.smartj.odin.insight.vo.InsightAlertVO;
import com.smartj.odin.insight.vo.InsightDashboardVO;
import com.smartj.odin.insight.vo.InsightMonitorVO;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

/**
 * 洞察域 Entity → VO 转换器。
 *
 * <p>统一将 ID 类字段转为字符串、时间字段格式化为 {@code yyyy-MM-dd HH:mm:ss}，
 * 规避前端大整数精度丢失并保证展示格式一致。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public final class InsightConverter {

    /** 统一时间格式 */
    private static final DateTimeFormatter FMT = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

    private InsightConverter() {
    }

    /**
     * 监控规则 Entity → VO。
     *
     * @param entity 监控规则实体
     * @return 监控规则 VO
     */
    public static InsightMonitorVO toMonitorVO(InsightMonitor entity) {
        if (entity == null) {
            return null;
        }
        InsightMonitorVO vo = BeanCopyUtil.copyProperties(entity, InsightMonitorVO.class);
        vo.setId(str(entity.getId()));
        vo.setOntologyId(str(entity.getOntologyId()));
        vo.setCreateTime(fmt(entity.getCreateTime()));
        return vo;
    }

    /**
     * 告警实例 Entity → VO。
     *
     * @param entity 告警实例实体
     * @return 告警实例 VO
     */
    public static InsightAlertVO toAlertVO(InsightAlert entity) {
        if (entity == null) {
            return null;
        }
        InsightAlertVO vo = BeanCopyUtil.copyProperties(entity, InsightAlertVO.class);
        vo.setId(str(entity.getId()));
        vo.setMonitorId(str(entity.getMonitorId()));
        vo.setTriggerTime(fmt(entity.getTriggerTime()));
        vo.setHandledTime(fmt(entity.getHandledTime()));
        vo.setCreateTime(fmt(entity.getCreateTime()));
        return vo;
    }

    /**
     * 变更历史 Entity → VO。
     *
     * @param entity 变更历史实体
     * @return 变更历史 VO
     */
    public static ChangeHistoryVO toChangeVO(ChangeHistory entity) {
        if (entity == null) {
            return null;
        }
        ChangeHistoryVO vo = BeanCopyUtil.copyProperties(entity, ChangeHistoryVO.class);
        vo.setId(str(entity.getId()));
        vo.setEntityId(str(entity.getEntityId()));
        vo.setCreateTime(fmt(entity.getCreateTime()));
        return vo;
    }

    /**
     * 洞察看板 Entity → VO。
     *
     * @param entity 洞察看板实体
     * @return 洞察看板 VO
     */
    public static InsightDashboardVO toDashboardVO(InsightDashboard entity) {
        if (entity == null) {
            return null;
        }
        InsightDashboardVO vo = BeanCopyUtil.copyProperties(entity, InsightDashboardVO.class);
        vo.setId(str(entity.getId()));
        vo.setOntologyId(str(entity.getOntologyId()));
        vo.setCreateTime(fmt(entity.getCreateTime()));
        return vo;
    }

    /**
     * Long → 字符串（null 安全）。
     *
     * @param value 数值
     * @return 字符串或 null
     */
    private static String str(Long value) {
        return value == null ? null : String.valueOf(value);
    }

    /**
     * LocalDateTime → 格式化字符串（null 安全）。
     *
     * @param time 时间
     * @return 格式化字符串或 null
     */
    private static String fmt(LocalDateTime time) {
        return time == null ? null : time.format(FMT);
    }
}
