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
package com.smartj.odin.insight.service;

import com.smartj.common.core.domain.PageResult;
import com.smartj.odin.insight.dto.AlertHandleDTO;
import com.smartj.odin.insight.dto.AlertQuery;
import com.smartj.odin.insight.dto.ChartSnapshotDTO;
import com.smartj.odin.insight.dto.InsightDashboardQuery;
import com.smartj.odin.insight.dto.InsightMonitorDTO;
import com.smartj.odin.insight.dto.InsightMonitorQuery;
import com.smartj.odin.insight.vo.InsightAlertVO;
import com.smartj.odin.insight.vo.InsightDashboardVO;
import com.smartj.odin.insight.vo.InsightMonitorVO;

/**
 * 数据洞察服务接口。
 *
 * <p>覆盖监控规则 CRUD、规则调度执行（P5 接入规则引擎），以及告警的分页查询、详情与人工处理闭环。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public interface InsightService {

    /**
     * 分页查询监控规则，附带告警统计。
     *
     * @param query 查询条件
     * @return 分页结果
     */
    PageResult<InsightMonitorVO> monitorPage(InsightMonitorQuery query);

    /**
     * 根据ID查询监控规则详情。
     *
     * @param id 主键ID
     * @return 监控规则 VO
     */
    InsightMonitorVO getMonitor(Long id);

    /**
     * 新增监控规则，返回主键ID。
     *
     * @param dto 监控规则 DTO
     * @return 新记录ID
     */
    Long createMonitor(InsightMonitorDTO dto);

    /**
     * 更新监控规则。
     *
     * @param dto 监控规则 DTO
     */
    void updateMonitor(InsightMonitorDTO dto);

    /**
     * 根据ID删除监控规则，级联删除其告警。
     *
     * @param id 主键ID
     */
    void deleteMonitor(Long id);

    /**
     * 手动触发监控规则执行，返回执行结果描述。
     *
     * @param id 监控规则ID
     * @return 执行结果描述
     */
    String runMonitor(Long id);

    /**
     * 分页查询告警。
     *
     * @param query 查询条件
     * @return 分页结果
     */
    PageResult<InsightAlertVO> alertPage(AlertQuery query);

    /**
     * 根据ID查询告警详情。
     *
     * @param id 主键ID
     * @return 告警 VO
     */
    InsightAlertVO getAlert(Long id);

    /**
     * 处理告警，标记为已处理。
     *
     * @param id  告警ID
     * @param dto 处理入参
     */
    void handleAlert(Long id, AlertHandleDTO dto);

    /**
     * AI 助手图表快照另存为洞察看板，返回新记录ID。
     *
     * @param dto 图表快照
     * @return 新看板ID
     */
    Long createFromChart(ChartSnapshotDTO dto);

    /**
     * 分页查询洞察看板。
     *
     * @param query 查询条件
     * @return 分页结果
     */
    PageResult<InsightDashboardVO> dashboardPage(InsightDashboardQuery query);

    /**
     * 根据ID删除洞察看板。
     *
     * @param id 主键ID
     */
    void deleteDashboard(Long id);
}
