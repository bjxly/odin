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
package com.smartj.odin.datasource.controller;

import com.smartj.common.core.domain.PageResult;
import com.smartj.common.core.domain.Result;
import com.smartj.common.log.annotation.OperationLog;
import com.smartj.odin.datasource.dto.DatasourceDTO;
import com.smartj.odin.datasource.dto.DatasourceQuery;
import com.smartj.odin.datasource.service.DatasourceService;
import com.smartj.odin.datasource.vo.DatasourceTestResultVO;
import com.smartj.odin.datasource.vo.DatasourceVO;
import com.smartj.odin.datasource.vo.SchemaColumnVO;
import com.smartj.odin.datasource.vo.SchemaTableVO;
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

import java.util.List;

/**
 * 数据源管理 Controller。
 *
 * <p>提供数据源 CRUD、连接测试、元数据同步与浏览接口。所有接口受权限控制，
 * 写操作通过 @OperationLog 记录审计日志。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Tag(name = "数据源管理", description = "数据源 CRUD、连接测试、元数据同步与浏览")
@RestController
@RequestMapping("/odin/datasource")
@RequiredArgsConstructor
public class DatasourceController {

    private final DatasourceService datasourceService;

    /** 分页查询数据源列表 */
    @Operation(summary = "分页查询数据源列表")
    @GetMapping("/page")
    @PreAuthorize("@perm.has('odin:datasource:list')")
    public Result<PageResult<DatasourceVO>> page(DatasourceQuery query) {
        return Result.ok(datasourceService.page(query));
    }

    /** 查询全部数据源列表（不分页，用于下拉选择） */
    @Operation(summary = "查询全部数据源列表")
    @GetMapping("/list")
    @PreAuthorize("@perm.has('odin:datasource:list')")
    public Result<List<DatasourceVO>> list() {
        return Result.ok(datasourceService.listAll());
    }

    /** 查询数据源摘要统计 */
    @Operation(summary = "查询数据源摘要统计")
    @GetMapping("/summaries")
    @PreAuthorize("@perm.has('odin:datasource:list')")
    public Result<List<DatasourceVO>> summaries() {
        return Result.ok(datasourceService.listAll());
    }

    /** 根据ID查询数据源详情 */
    @Operation(summary = "根据ID查询数据源详情")
    @GetMapping("/{id}")
    @PreAuthorize("@perm.has('odin:datasource:query')")
    public Result<DatasourceVO> getById(@Parameter(name = "id", description = "主键ID", required = true) @PathVariable Long id) {
        return Result.ok(datasourceService.getById(id));
    }

    /** 新增数据源，返回新记录ID */
    @Operation(summary = "新增数据源")
    @PostMapping
    @PreAuthorize("@perm.has('odin:datasource:create')")
    @OperationLog(module = "数据源管理", action = "创建数据源")
    public Result<Long> create(@Valid @RequestBody DatasourceDTO dto) {
        return Result.ok(datasourceService.create(dto));
    }

    /** 更新数据源 */
    @Operation(summary = "更新数据源")
    @PutMapping
    @PreAuthorize("@perm.has('odin:datasource:update')")
    @OperationLog(module = "数据源管理", action = "更新数据源")
    public Result<Void> update(@Valid @RequestBody DatasourceDTO dto) {
        datasourceService.update(dto);
        return Result.ok();
    }

    /** 根据ID删除数据源 */
    @Operation(summary = "根据ID删除数据源")
    @DeleteMapping("/{id}")
    @PreAuthorize("@perm.has('odin:datasource:delete')")
    @OperationLog(module = "数据源管理", action = "删除数据源")
    public Result<Void> delete(@Parameter(name = "id", description = "主键ID", required = true) @PathVariable Long id) {
        datasourceService.delete(id);
        return Result.ok();
    }

    /** 测试数据源连接 */
    @Operation(summary = "测试数据源连接")
    @PostMapping("/{id}/test")
    @PreAuthorize("@perm.has('odin:datasource:test')")
    @OperationLog(module = "数据源管理", action = "测试连接")
    public Result<DatasourceTestResultVO> test(@Parameter(name = "id", description = "主键ID", required = true) @PathVariable Long id) {
        return Result.ok(datasourceService.testConnection(id));
    }

    /** 同步数据源元数据，返回同步的表数量 */
    @Operation(summary = "同步数据源元数据")
    @PostMapping("/{id}/sync")
    @PreAuthorize("@perm.has('odin:datasource:sync')")
    @OperationLog(module = "数据源管理", action = "同步元数据")
    public Result<Integer> sync(@Parameter(name = "id", description = "主键ID", required = true) @PathVariable Long id) {
        return Result.ok(datasourceService.syncMetadata(id));
    }

    /** 查询数据源已同步的表清单 */
    @Operation(summary = "查询数据源表清单")
    @GetMapping("/{id}/tables")
    @PreAuthorize("@perm.has('odin:datasource:query')")
    public Result<List<SchemaTableVO>> tables(@Parameter(name = "id", description = "数据源ID", required = true) @PathVariable Long id) {
        return Result.ok(datasourceService.listTables(id));
    }

    /** 查询表已同步的列清单 */
    @Operation(summary = "查询表列清单")
    @GetMapping("/table/{tableId}/columns")
    @PreAuthorize("@perm.has('odin:datasource:query')")
    public Result<List<SchemaColumnVO>> columns(@Parameter(name = "tableId", description = "表元数据ID", required = true) @PathVariable Long tableId) {
        return Result.ok(datasourceService.listColumns(tableId));
    }
}
