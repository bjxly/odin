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
import com.smartj.odin.insight.dto.ChangeHistoryQuery;
import com.smartj.odin.insight.service.ChangeHistoryService;
import com.smartj.odin.insight.vo.ChangeHistoryVO;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.RequiredArgsConstructor;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * 变更历史 Controller。
 *
 * <p>提供各类实体变更快照的分页审计查询与变更回滚接口。所有接口受权限控制，
 * 回滚操作通过 @OperationLog 记录审计日志。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Tag(name = "变更历史", description = "变更审计查询与回滚")
@RestController
@RequestMapping("/odin/changes")
@RequiredArgsConstructor
public class ChangeHistoryController {

    private final ChangeHistoryService changeHistoryService;

    /** 分页查询变更历史 */
    @Operation(summary = "分页查询变更历史")
    @GetMapping("/page")
    @PreAuthorize("@perm.has('odin:changes:list')")
    public Result<PageResult<ChangeHistoryVO>> page(ChangeHistoryQuery query) {
        return Result.ok(changeHistoryService.page(query));
    }

    /** 回滚指定变更记录 */
    @Operation(summary = "回滚指定变更记录")
    @PostMapping("/{id}/revert")
    @PreAuthorize("@perm.has('odin:changes:revert')")
    @OperationLog(module = "变更历史", action = "回滚变更")
    public Result<Void> revert(@Parameter(name = "id", description = "变更历史ID", required = true) @PathVariable Long id) {
        changeHistoryService.revert(id);
        return Result.ok();
    }
}
