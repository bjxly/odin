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
package com.smartj.odin.query.controller;

import com.smartj.common.core.domain.PageResult;
import com.smartj.common.core.domain.Result;
import com.smartj.common.log.annotation.OperationLog;
import com.smartj.odin.query.dto.QueryDryRunDTO;
import com.smartj.odin.query.dto.QueryExecuteDTO;
import com.smartj.odin.query.dto.QueryHistoryQuery;
import com.smartj.odin.query.dto.QueryTemplateDTO;
import com.smartj.odin.query.dto.QueryTemplateQuery;
import com.smartj.odin.query.service.QueryService;
import com.smartj.odin.query.vo.QueryDryRunVO;
import com.smartj.odin.query.vo.QueryHistoryVO;
import com.smartj.odin.query.vo.QueryResultVO;
import com.smartj.odin.query.vo.QueryTemplateVO;
import com.smartj.odin.query.vo.QueryTraceVO;
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

import java.util.Map;

/**
 * 智能问数 Controller。
 *
 * <p>提供自然语言问数执行、试运行校验、逻辑 SQL 解析，以及查询历史/溯源浏览与查询模板管理接口。
 * 所有接口受权限控制，写操作通过 @OperationLog 记录审计日志。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Tag(name = "智能问数", description = "问数执行、试运行、历史溯源与查询模板管理")
@RestController
@RequestMapping("/odin/query")
@RequiredArgsConstructor
public class QueryController {

    private final QueryService queryService;

    /** 执行自然语言问数 */
    @Operation(summary = "执行自然语言问数")
    @PostMapping("/execute")
    @PreAuthorize("@perm.has('odin:query:execute')")
    @OperationLog(module = "智能问数", action = "执行问数")
    public Result<QueryResultVO> execute(@Valid @RequestBody QueryExecuteDTO dto) {
        return Result.ok(queryService.execute(dto));
    }

    /** 试运行本体逻辑 SQL */
    @Operation(summary = "试运行本体逻辑SQL")
    @PostMapping("/dry-run")
    @PreAuthorize("@perm.has('odin:query:execute')")
    public Result<QueryDryRunVO> dryRun(@Valid @RequestBody QueryDryRunDTO dto) {
        return Result.ok(queryService.dryRun(dto));
    }

    /** 将自然语言解析为本体逻辑 SQL */
    @Operation(summary = "自然语言解析为逻辑SQL")
    @PostMapping("/nl-parse")
    @PreAuthorize("@perm.has('odin:query:execute')")
    public Result<String> nlParse(@RequestBody QueryExecuteDTO dto) {
        return Result.ok(queryService.nlParse(dto.getNaturalLanguage(), dto.getOntologyId()));
    }

    /** 分页查询查询历史 */
    @Operation(summary = "分页查询查询历史")
    @GetMapping("/history/page")
    @PreAuthorize("@perm.has('odin:query:history')")
    public Result<PageResult<QueryHistoryVO>> historyPage(QueryHistoryQuery query) {
        return Result.ok(queryService.historyPage(query));
    }

    /** 查询指定历史的执行溯源 */
    @Operation(summary = "查询执行溯源")
    @GetMapping("/history/{historyId}/trace")
    @PreAuthorize("@perm.has('odin:query:history')")
    public Result<QueryTraceVO> trace(@Parameter(name = "historyId", description = "查询历史ID", required = true) @PathVariable Long historyId) {
        return Result.ok(queryService.getTrace(historyId));
    }

    /** 分页查询查询模板 */
    @Operation(summary = "分页查询查询模板")
    @GetMapping("/template/page")
    @PreAuthorize("@perm.has('odin:query:template')")
    public Result<PageResult<QueryTemplateVO>> templatePage(QueryTemplateQuery query) {
        return Result.ok(queryService.templatePage(query));
    }

    /** 根据ID查询查询模板详情 */
    @Operation(summary = "根据ID查询查询模板详情")
    @GetMapping("/template/{id}")
    @PreAuthorize("@perm.has('odin:query:template')")
    public Result<QueryTemplateVO> getTemplate(@Parameter(name = "id", description = "模板ID", required = true) @PathVariable Long id) {
        return Result.ok(queryService.getTemplate(id));
    }

    /** 新增查询模板，返回新记录ID */
    @Operation(summary = "新增查询模板")
    @PostMapping("/template")
    @PreAuthorize("@perm.has('odin:query:create')")
    @OperationLog(module = "智能问数", action = "创建查询模板")
    public Result<Long> createTemplate(@Valid @RequestBody QueryTemplateDTO dto) {
        return Result.ok(queryService.createTemplate(dto));
    }

    /** 更新查询模板 */
    @Operation(summary = "更新查询模板")
    @PutMapping("/template")
    @PreAuthorize("@perm.has('odin:query:create')")
    @OperationLog(module = "智能问数", action = "更新查询模板")
    public Result<Void> updateTemplate(@Valid @RequestBody QueryTemplateDTO dto) {
        queryService.updateTemplate(dto);
        return Result.ok();
    }

    /** 根据ID删除查询模板 */
    @Operation(summary = "删除查询模板")
    @DeleteMapping("/template/{id}")
    @PreAuthorize("@perm.has('odin:query:delete')")
    @OperationLog(module = "智能问数", action = "删除查询模板")
    public Result<Void> deleteTemplate(@Parameter(name = "id", description = "主键ID", required = true) @PathVariable Long id) {
        queryService.deleteTemplate(id);
        return Result.ok();
    }

    /** 按模板执行查询 */
    @Operation(summary = "按模板执行查询")
    @PostMapping("/template/{id}/execute")
    @PreAuthorize("@perm.has('odin:query:execute')")
    @OperationLog(module = "智能问数", action = "执行模板查询")
    public Result<QueryResultVO> executeTemplate(@Parameter(name = "id", description = "模板ID", required = true) @PathVariable Long id,
                                                  @RequestBody(required = false) Map<String, Object> params) {
        return Result.ok(queryService.executeTemplate(id, params));
    }
}
