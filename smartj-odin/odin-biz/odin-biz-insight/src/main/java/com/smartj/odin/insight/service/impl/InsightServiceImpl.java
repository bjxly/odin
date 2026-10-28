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
package com.smartj.odin.insight.service.impl;

import com.mybatisflex.core.paginate.Page;
import com.mybatisflex.core.query.QueryWrapper;
import tools.jackson.databind.ObjectMapper;
import com.smartj.common.auth.util.SecurityUtils;
import com.smartj.common.core.domain.PageResult;
import com.smartj.common.core.exception.BusinessException;
import com.smartj.odin.insight.converter.InsightConverter;
import com.smartj.odin.insight.dto.AlertHandleDTO;
import com.smartj.odin.insight.dto.AlertQuery;
import com.smartj.odin.insight.dto.ChartSnapshotDTO;
import com.smartj.odin.insight.dto.InsightDashboardQuery;
import com.smartj.odin.insight.dto.InsightMonitorDTO;
import com.smartj.odin.insight.dto.InsightMonitorQuery;
import com.smartj.odin.insight.entity.InsightAlert;
import com.smartj.odin.insight.entity.InsightDashboard;
import com.smartj.odin.insight.entity.InsightMonitor;
import com.smartj.odin.insight.error.InsightErrorCode;
import com.smartj.odin.insight.mapper.InsightAlertMapper;
import com.smartj.odin.insight.mapper.InsightDashboardMapper;
import com.smartj.odin.insight.mapper.InsightMonitorMapper;
import com.smartj.odin.insight.service.InsightService;
import com.smartj.odin.insight.vo.InsightAlertVO;
import com.smartj.odin.insight.vo.InsightDashboardVO;
import com.smartj.odin.insight.vo.InsightMonitorVO;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.stream.Collectors;

import static com.smartj.odin.insight.entity.table.InsightAlertTableDef.INSIGHT_ALERT;
import static com.smartj.odin.insight.entity.table.InsightDashboardTableDef.INSIGHT_DASHBOARD;
import static com.smartj.odin.insight.entity.table.InsightMonitorTableDef.INSIGHT_MONITOR;

/**
 * 数据洞察服务实现。
 *
 * <p>监控规则 CRUD、告警查询与人工处理为完整实现；监控规则的调度执行（规则引擎）留待 P5 接入，
 * 当前 {@link #runMonitor(Long)} 为骨架实现。列表页附带告警统计（累计告警数、最近告警时间）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class InsightServiceImpl implements InsightService {

    /** 统一时间格式 */
    private static final DateTimeFormatter FMT = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

    private final InsightMonitorMapper insightMonitorMapper;
    private final InsightAlertMapper insightAlertMapper;
    private final InsightDashboardMapper insightDashboardMapper;
    private final ObjectMapper objectMapper;

    /** 分页查询监控规则，批量附加告警统计（累计告警数、最近告警时间） */
    @Override
    @Transactional(readOnly = true)
    public PageResult<InsightMonitorVO> monitorPage(InsightMonitorQuery query) {
        QueryWrapper qw = QueryWrapper.create()
                .where(INSIGHT_MONITOR.NAME.like(query.getName()).when(StringUtils.hasText(query.getName())))
                .and(INSIGHT_MONITOR.ONTOLOGY_ID.eq(query.getOntologyId()).when(query.getOntologyId() != null))
                .and(INSIGHT_MONITOR.ENABLED.eq(query.getEnabled()).when(query.getEnabled() != null))
                .and(INSIGHT_MONITOR.STATUS.eq(query.getStatus()).when(query.getStatus() != null))
                .orderBy(INSIGHT_MONITOR.CREATE_TIME.desc());

        Page<InsightMonitor> page = insightMonitorMapper.paginate(Page.of(query.getCurrent(), query.getSize()), qw);
        List<InsightMonitor> records = page.getRecords();

        // 批量加载当前页监控的告警，按 monitorId 分组统计，规避 N+1
        List<Long> monitorIds = records.stream().map(InsightMonitor::getId).collect(Collectors.toList());
        Map<Long, List<InsightAlert>> alertMap = monitorIds.isEmpty() ? Collections.emptyMap()
                : insightAlertMapper.selectListByQuery(QueryWrapper.create()
                        .where(INSIGHT_ALERT.MONITOR_ID.in(monitorIds)))
                .stream().collect(Collectors.groupingBy(InsightAlert::getMonitorId));

        List<InsightMonitorVO> vos = records.stream().map(m -> {
            InsightMonitorVO vo = InsightConverter.toMonitorVO(m);
            applyAlertStats(vo, alertMap.getOrDefault(m.getId(), Collections.emptyList()));
            return vo;
        }).collect(Collectors.toList());
        return PageResult.of(vos, page.getTotalRow(), page.getPageNumber(), page.getPageSize());
    }

    /** 根据ID查询监控规则详情，附带告警统计 */
    @Override
    @Transactional(readOnly = true)
    public InsightMonitorVO getMonitor(Long id) {
        InsightMonitor entity = loadMonitor(id);
        InsightMonitorVO vo = InsightConverter.toMonitorVO(entity);
        List<InsightAlert> alerts = insightAlertMapper.selectListByQuery(
                QueryWrapper.create().where(INSIGHT_ALERT.MONITOR_ID.eq(id)));
        applyAlertStats(vo, alerts);
        return vo;
    }

    /** 新增监控规则，返回主键ID */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public Long createMonitor(InsightMonitorDTO dto) {
        InsightMonitor entity = new InsightMonitor();
        copyDtoToEntity(dto, entity);
        insightMonitorMapper.insert(entity);
        return entity.getId();
    }

    /** 更新监控规则：先查后改 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void updateMonitor(InsightMonitorDTO dto) {
        InsightMonitor entity = loadMonitor(dto.getId());
        copyDtoToEntity(dto, entity);
        insightMonitorMapper.update(entity);
    }

    /** 删除监控规则，并级联删除其告警 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void deleteMonitor(Long id) {
        loadMonitor(id);
        insightAlertMapper.deleteByQuery(QueryWrapper.create().where(INSIGHT_ALERT.MONITOR_ID.eq(id)));
        insightMonitorMapper.deleteById(id);
    }

    /** 手动触发监控规则执行，返回执行结果描述（规则引擎调度执行待 P5 接入） */
    @Override
    public String runMonitor(Long id) {
        loadMonitor(id);
        // TODO: P5 实现规则引擎调度执行（编译规则表达式→执行指标查询→命中阈值产出告警→通知分发）
        return "监控执行功能待实现（P5）";
    }

    /** 分页查询告警，支持监控/严重程度/处理状态筛选，并富化监控名称 */
    @Override
    @Transactional(readOnly = true)
    public PageResult<InsightAlertVO> alertPage(AlertQuery query) {
        QueryWrapper qw = QueryWrapper.create()
                .where(INSIGHT_ALERT.MONITOR_ID.eq(query.getMonitorId()).when(query.getMonitorId() != null))
                .and(INSIGHT_ALERT.SEVERITY.eq(query.getSeverity()).when(StringUtils.hasText(query.getSeverity())))
                .and(INSIGHT_ALERT.HANDLED.eq(query.getHandled()).when(query.getHandled() != null))
                .orderBy(INSIGHT_ALERT.TRIGGER_TIME.desc());

        Page<InsightAlert> page = insightAlertMapper.paginate(Page.of(query.getCurrent(), query.getSize()), qw);
        List<InsightAlert> records = page.getRecords();

        // 批量加载监控名称，规避 N+1
        Map<Long, String> monitorNames = loadMonitorNames(records);
        List<InsightAlertVO> vos = records.stream().map(a -> {
            InsightAlertVO vo = InsightConverter.toAlertVO(a);
            vo.setMonitorName(monitorNames.get(a.getMonitorId()));
            return vo;
        }).collect(Collectors.toList());
        return PageResult.of(vos, page.getTotalRow(), page.getPageNumber(), page.getPageSize());
    }

    /** 根据ID查询告警详情，富化监控名称 */
    @Override
    @Transactional(readOnly = true)
    public InsightAlertVO getAlert(Long id) {
        InsightAlert entity = loadAlert(id);
        InsightAlertVO vo = InsightConverter.toAlertVO(entity);
        if (entity.getMonitorId() != null) {
            InsightMonitor monitor = insightMonitorMapper.selectOneById(entity.getMonitorId());
            vo.setMonitorName(monitor == null ? null : monitor.getName());
        }
        return vo;
    }

    /** 处理告警：标记为已处理并记录处理人/时间/说明，重复处理抛业务异常 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void handleAlert(Long id, AlertHandleDTO dto) {
        InsightAlert entity = loadAlert(id);
        if (entity.getHandled() != null && entity.getHandled() == 1) {
            throw new BusinessException(InsightErrorCode.ALERT_ALREADY_HANDLED);
        }
        entity.setHandled(1);
        entity.setHandledBy(SecurityUtils.getUsername());
        entity.setHandledTime(LocalDateTime.now());
        entity.setHandleNote(dto.getHandleNote());
        insightAlertMapper.update(entity);
    }

    /**
     * 加载监控规则实体，不存在抛业务异常。
     *
     * @param id 主键ID
     * @return 监控规则实体
     */
    private InsightMonitor loadMonitor(Long id) {
        InsightMonitor entity = insightMonitorMapper.selectOneById(id);
        if (entity == null) {
            throw new BusinessException(InsightErrorCode.MONITOR_NOT_FOUND);
        }
        return entity;
    }

    /**
     * 加载告警实体，不存在抛业务异常。
     *
     * @param id 主键ID
     * @return 告警实体
     */
    private InsightAlert loadAlert(Long id) {
        InsightAlert entity = insightAlertMapper.selectOneById(id);
        if (entity == null) {
            throw new BusinessException(InsightErrorCode.ALERT_NOT_FOUND);
        }
        return entity;
    }

    /**
     * 依据告警集合为监控 VO 填充累计告警数与最近告警时间。
     *
     * @param vo     监控规则 VO
     * @param alerts 该监控的告警集合
     */
    private void applyAlertStats(InsightMonitorVO vo, List<InsightAlert> alerts) {
        vo.setAlertCount(alerts == null ? 0 : alerts.size());
        if (alerts == null || alerts.isEmpty()) {
            return;
        }
        alerts.stream()
                .map(InsightAlert::getTriggerTime)
                .filter(Objects::nonNull)
                .max(LocalDateTime::compareTo)
                .ifPresent(t -> vo.setLastAlertTime(t.format(FMT)));
    }

    /**
     * 批量加载告警所属监控的名称映射（monitorId → name）。
     *
     * @param alerts 告警集合
     * @return 监控名称映射
     */
    private Map<Long, String> loadMonitorNames(List<InsightAlert> alerts) {
        List<Long> monitorIds = alerts.stream()
                .map(InsightAlert::getMonitorId)
                .filter(Objects::nonNull)
                .distinct()
                .collect(Collectors.toList());
        if (monitorIds.isEmpty()) {
            return Collections.emptyMap();
        }
        List<InsightMonitor> monitors = insightMonitorMapper.selectListByQuery(
                QueryWrapper.create().where(INSIGHT_MONITOR.ID.in(monitorIds)));
        return monitors.stream()
                .collect(Collectors.toMap(InsightMonitor::getId, InsightMonitor::getName, (a, b) -> a));
    }

    /**
     * 将 DTO 拷贝到监控规则实体。
     *
     * @param dto    监控规则 DTO
     * @param entity 目标实体
     */
    private void copyDtoToEntity(InsightMonitorDTO dto, InsightMonitor entity) {
        entity.setName(dto.getName());
        entity.setDescription(dto.getDescription());
        entity.setOntologyId(dto.getOntologyId());
        entity.setRuleExpression(dto.getRuleExpression());
        entity.setScheduleCron(dto.getScheduleCron());
        entity.setThresholds(dto.getThresholds());
        entity.setNotificationChannels(dto.getNotificationChannels());
        entity.setEnabled(dto.getEnabled() == null ? 1 : dto.getEnabled());
        entity.setStatus(dto.getStatus() == null ? 1 : dto.getStatus());
        entity.setRemark(dto.getRemark());
    }

    /** AI 助手图表快照另存为洞察看板：列/行/配置序列化为 JSON 字符串落库。 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public Long createFromChart(ChartSnapshotDTO dto) {
        InsightDashboard entity = new InsightDashboard();
        entity.setName(dto.getName());
        entity.setOntologyId(dto.getOntologyId());
        entity.setTitle(dto.getTitle());
        entity.setChartType(dto.getChartType());
        entity.setColumnsJson(toJson(dto.getColumns()));
        entity.setRowsJson(toJson(dto.getRows()));
        entity.setChartOption(toJson(dto.getChartOption()));
        entity.setSource("assistant_chart");
        entity.setStatus(1);
        insightDashboardMapper.insert(entity);
        return entity.getId();
    }

    /** 分页查询洞察看板。 */
    @Override
    @Transactional(readOnly = true)
    public PageResult<InsightDashboardVO> dashboardPage(InsightDashboardQuery query) {
        QueryWrapper qw = QueryWrapper.create()
                .where(INSIGHT_DASHBOARD.NAME.like(query.getName()).when(StringUtils.hasText(query.getName())))
                .and(INSIGHT_DASHBOARD.ONTOLOGY_ID.eq(query.getOntologyId()).when(query.getOntologyId() != null))
                .and(INSIGHT_DASHBOARD.STATUS.eq(query.getStatus()).when(query.getStatus() != null))
                .orderBy(INSIGHT_DASHBOARD.CREATE_TIME.desc());

        Page<InsightDashboard> page = insightDashboardMapper.paginate(
                Page.of(query.getCurrent(), query.getSize()), qw);
        List<InsightDashboardVO> vos = page.getRecords().stream()
                .map(InsightConverter::toDashboardVO)
                .collect(Collectors.toList());
        return PageResult.of(vos, page.getTotalRow(), page.getPageNumber(), page.getPageSize());
    }

    /** 根据ID删除洞察看板。 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void deleteDashboard(Long id) {
        InsightDashboard entity = insightDashboardMapper.selectOneById(id);
        if (entity == null) {
            throw new BusinessException(InsightErrorCode.DASHBOARD_NOT_FOUND);
        }
        insightDashboardMapper.deleteById(id);
    }

    /** 对象 → JSON 字符串（null 安全；序列化失败降级为 null）。 */
    private String toJson(Object value) {
        if (value == null) {
            return null;
        }
        try {
            return objectMapper.writeValueAsString(value);
        } catch (Exception e) {
            log.warn("图表快照序列化失败: {}", e.getMessage());
            return null;
        }
    }
}
