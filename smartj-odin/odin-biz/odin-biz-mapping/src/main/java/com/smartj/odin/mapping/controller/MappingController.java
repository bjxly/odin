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
package com.smartj.odin.mapping.controller;

import com.smartj.common.core.domain.PageResult;
import com.smartj.common.core.domain.Result;
import com.smartj.common.log.annotation.OperationLog;
import com.smartj.odin.mapping.dto.MappingConfigDTO;
import com.smartj.odin.mapping.dto.MappingConfigQuery;
import com.smartj.odin.mapping.service.MappingService;
import com.smartj.odin.mapping.vo.MappingConfigVO;
import com.smartj.odin.mapping.vo.MappingSuggestVO;
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
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

/**
 * 映射管理 Controller。
 *
 * <p>提供映射配置 CRUD、完整性校验与智能列映射推荐接口。
 * 所有接口受权限控制，写操作通过 @OperationLog 记录审计日志。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Tag(name = "映射管理", description = "映射配置 CRUD、校验与智能推荐")
@RestController
@RequestMapping("/odin/mapping")
@RequiredArgsConstructor
public class MappingController {

    private final MappingService mappingService;

    /** 分页查询映射配置列表 */
    @Operation(summary = "分页查询映射配置列表")
    @GetMapping("/page")
    @PreAuthorize("@perm.has('odin:mapping:list')")
    public Result<PageResult<MappingConfigVO>> page(MappingConfigQuery query) {
        return Result.ok(mappingService.page(query));
    }

    /** 根据ID查询映射配置详情 */
    @Operation(summary = "根据ID查询映射配置详情")
    @GetMapping("/{id}")
    @PreAuthorize("@perm.has('odin:mapping:query')")
    public Result<MappingConfigVO> getById(@Parameter(name = "id", description = "主键ID", required = true) @PathVariable Long id) {
        return Result.ok(mappingService.getById(id));
    }

    /** 新增映射配置 */
    @Operation(summary = "新增映射配置")
    @PostMapping
    @PreAuthorize("@perm.has('odin:mapping:create')")
    @OperationLog(module = "映射管理", action = "创建映射配置")
    public Result<Long> create(@Valid @RequestBody MappingConfigDTO dto) {
        return Result.ok(mappingService.create(dto));
    }

    /** 更新映射配置 */
    @Operation(summary = "更新映射配置")
    @PutMapping
    @PreAuthorize("@perm.has('odin:mapping:update')")
    @OperationLog(module = "映射管理", action = "更新映射配置")
    public Result<Void> update(@Valid @RequestBody MappingConfigDTO dto) {
        mappingService.update(dto);
        return Result.ok();
    }

    /** 删除映射配置 */
    @Operation(summary = "删除映射配置")
    @DeleteMapping("/{id}")
    @PreAuthorize("@perm.has('odin:mapping:delete')")
    @OperationLog(module = "映射管理", action = "删除映射配置")
    public Result<Void> delete(@Parameter(name = "id", description = "主键ID", required = true) @PathVariable Long id) {
        mappingService.delete(id);
        return Result.ok();
    }

    /** 校验映射完整性 */
    @Operation(summary = "校验映射完整性")
    @PostMapping("/{id}/validate")
    @PreAuthorize("@perm.has('odin:mapping:query')")
    @OperationLog(module = "映射管理", action = "校验映射")
    public Result<String> validate(@Parameter(name = "id", description = "映射配置ID", required = true) @PathVariable Long id) {
        return Result.ok(mappingService.validate(id));
    }

    /** 智能推荐列映射（前端契约：classId + tableId） */
    @Operation(summary = "智能推荐列映射")
    @GetMapping("/suggest")
    @PreAuthorize("@perm.has('odin:mapping:query')")
    public Result<List<MappingSuggestVO>> suggest(
            @Parameter(name = "classId", description = "类ID", required = true) @RequestParam Long classId,
            @Parameter(name = "tableId", description = "物理表元数据ID", required = true) @RequestParam Long tableId) {
        return Result.ok(mappingService.suggestByTable(classId, tableId));
    }
}
