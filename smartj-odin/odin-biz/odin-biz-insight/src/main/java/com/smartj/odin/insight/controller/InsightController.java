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
package com.smartj.odin.insight.controller;

import com.smartj.common.core.domain.PageResult;
import com.smartj.common.core.domain.Result;
import com.smartj.common.log.annotation.OperationLog;
import com.smartj.odin.insight.dto.AlertHandleDTO;
import com.smartj.odin.insight.dto.AlertQuery;
import com.smartj.odin.insight.dto.ChartSnapshotDTO;
import com.smartj.odin.insight.dto.InsightDashboardQuery;
import com.smartj.odin.insight.dto.InsightMonitorDTO;
import com.smartj.odin.insight.dto.InsightMonitorQuery;
import com.smartj.odin.insight.service.InsightService;
import com.smartj.odin.insight.vo.InsightAlertVO;
import com.smartj.odin.insight.vo.InsightDashboardVO;
import com.smartj.odin.insight.vo.InsightMonitorVO;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * 数据洞察 Controller。
 *
 * <p>提供监控规则 CRUD、规则执行，以及告警查询与人工处理接口。所有接口受权限控制，
 * 写操作通过 @OperationLog 记录审计日志。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Tag(name = "数据洞察", description = "监控规则管理、规则执行与告警处理")
@RestController
@RequestMapping("/odin/insight")
@RequiredArgsConstructor
public class InsightController {

    private final InsightService insightService;

    /** 分页查询监控规则列表 */
    @Operation(summary = "分页查询监控规则列表")
    @GetMapping("/monitor/page")
    @PreAuthorize("@perm.has('odin:insight:list')")
    public Result<PageResult<InsightMonitorVO>> monitorPage(InsightMonitorQuery query) {
        return Result.ok(insightService.monitorPage(query));
    }

    /** 根据ID查询监控规则详情 */
    @Operation(summary = "根据ID查询监控规则详情")
    @GetMapping("/monitor/{id}")
    @PreAuthorize("@perm.has('odin:insight:query')")
    public Result<InsightMonitorVO> getMonitor(@Parameter(name = "id", description = "主键ID", required = true) @PathVariable Long id) {
        return Result.ok(insightService.getMonitor(id));
    }

    /** 新增监控规则，返回新记录ID */
    @Operation(summary = "新增监控规则")
    @PostMapping("/monitor")
    @PreAuthorize("@perm.has('odin:insight:create')")
    @OperationLog(module = "数据洞察", action = "创建监控规则")
    public Result<Long> createMonitor(@Valid @RequestBody InsightMonitorDTO dto) {
        return Result.ok(insightService.createMonitor(dto));
    }

    /** 更新监控规则 */
    @Operation(summary = "更新监控规则")
    @PutMapping("/monitor")
    @PreAuthorize("@perm.has('odin:insight:update')")
    @OperationLog(module = "数据洞察", action = "更新监控规则")
    public Result<Void> updateMonitor(@Valid @RequestBody InsightMonitorDTO dto) {
        insightService.updateMonitor(dto);
        return Result.ok();
    }

    /** 根据ID删除监控规则 */
    @Operation(summary = "根据ID删除监控规则")
    @DeleteMapping("/monitor/{id}")
    @PreAuthorize("@perm.has('odin:insight:delete')")
    @OperationLog(module = "数据洞察", action = "删除监控规则")
    public Result<Void> deleteMonitor(@Parameter(name = "id", description = "主键ID", required = true) @PathVariable Long id) {
        insightService.deleteMonitor(id);
        return Result.ok();
    }

    /** 手动触发监控规则执行 */
    @Operation(summary = "手动触发监控规则执行")
    @PostMapping("/monitor/{id}/run")
    @PreAuthorize("@perm.has('odin:insight:run')")
    @OperationLog(module = "数据洞察", action = "执行监控规则")
    public Result<String> runMonitor(@Parameter(name = "id", description = "监控规则ID", required = true) @PathVariable Long id) {
        return Result.ok(insightService.runMonitor(id));
    }

    /** 分页查询告警列表 */
    @Operation(summary = "分页查询告警列表")
    @GetMapping("/alert/page")
    @PreAuthorize("@perm.has('odin:insight:list')")
    public Result<PageResult<InsightAlertVO>> alertPage(AlertQuery query) {
        return Result.ok(insightService.alertPage(query));
    }

    /** 根据ID查询告警详情 */
    @Operation(summary = "根据ID查询告警详情")
    @GetMapping("/alert/{id}")
    @PreAuthorize("@perm.has('odin:insight:query')")
    public Result<InsightAlertVO> getAlert(@Parameter(name = "id", description = "告警ID", required = true) @PathVariable Long id) {
        return Result.ok(insightService.getAlert(id));
    }

    /** 处理告警，标记为已处理 */
    @Operation(summary = "处理告警")
    @PutMapping("/alert/{id}/handle")
    @PreAuthorize("@perm.has('odin:insight:handle')")
    @OperationLog(module = "数据洞察", action = "处理告警")
    public Result<Void> handleAlert(@Parameter(name = "id", description = "告警ID", required = true) @PathVariable Long id,
                                    @Valid @RequestBody AlertHandleDTO dto) {
        insightService.handleAlert(id, dto);
        return Result.ok();
    }

    /** AI 助手图表快照另存为洞察看板，返回新记录ID */
    @Operation(summary = "图表另存为洞察看板")
    @PostMapping("/dashboard")
    @PreAuthorize("@perm.has('odin:insight:create')")
    @OperationLog(module = "数据洞察", action = "另存为洞察看板")
    public Result<Long> createFromChart(@Valid @RequestBody ChartSnapshotDTO dto) {
        return Result.ok(insightService.createFromChart(dto));
    }

    /** 分页查询洞察看板 */
    @Operation(summary = "分页查询洞察看板")
    @GetMapping("/dashboard/page")
    @PreAuthorize("@perm.has('odin:insight:list')")
    public Result<PageResult<InsightDashboardVO>> dashboardPage(InsightDashboardQuery query) {
        return Result.ok(insightService.dashboardPage(query));
    }

    /** 根据ID删除洞察看板 */
    @Operation(summary = "删除洞察看板")
    @DeleteMapping("/dashboard/{id}")
    @PreAuthorize("@perm.has('odin:insight:delete')")
    @OperationLog(module = "数据洞察", action = "删除洞察看板")
    public Result<Void> deleteDashboard(@Parameter(name = "id", description = "看板ID", required = true) @PathVariable Long id) {
        insightService.deleteDashboard(id);
        return Result.ok();
    }
}
