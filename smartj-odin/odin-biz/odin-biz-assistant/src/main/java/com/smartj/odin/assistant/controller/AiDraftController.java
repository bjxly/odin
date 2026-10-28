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
package com.smartj.odin.assistant.controller;

import com.smartj.common.core.domain.PageResult;
import com.smartj.common.core.domain.Result;
import com.smartj.common.log.annotation.OperationLog;
import com.smartj.odin.assistant.dto.DraftQuery;
import com.smartj.odin.assistant.dto.DraftReviewDTO;
import com.smartj.odin.assistant.service.AiDraftService;
import com.smartj.odin.assistant.vo.AiDraftVO;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

/**
 * AI 草稿 Controller（HITL 审阅与发布）。
 *
 * <p>提供草稿分页、详情、逐项审阅与发布/驳回接口。所有接口受权限控制，
 * 写操作通过 @OperationLog 记录审计日志。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Tag(name = "AI 草稿", description = "AI 草稿的分页、详情、审阅与发布")
@RestController
@RequestMapping("/odin/ai/drafts")
@RequiredArgsConstructor
public class AiDraftController {

    private final AiDraftService aiDraftService;

    /** 分页查询 AI 草稿 */
    @Operation(summary = "分页查询AI草稿")
    @GetMapping("/page")
    @PreAuthorize("@perm.has('odin:assistant:draft')")
    public Result<PageResult<AiDraftVO>> page(DraftQuery query) {
        return Result.ok(aiDraftService.draftPage(query));
    }

    /** 查询草稿详情（含明细项） */
    @Operation(summary = "查询草稿详情")
    @GetMapping("/{id}")
    @PreAuthorize("@perm.has('odin:assistant:draft')")
    public Result<AiDraftVO> getDraft(@Parameter(name = "id", description = "草稿ID", required = true) @PathVariable Long id) {
        return Result.ok(aiDraftService.getDraft(id));
    }

    /** 审阅单条草稿明细项 */
    @Operation(summary = "审阅草稿明细项")
    @PostMapping("/{id}/items/{itemId}/review")
    @PreAuthorize("@perm.has('odin:assistant:draft')")
    @OperationLog(module = "AI草稿", action = "审阅明细项")
    public Result<Void> reviewItem(@Parameter(name = "id", description = "草稿ID", required = true) @PathVariable Long id,
                                   @Parameter(name = "itemId", description = "明细项ID", required = true) @PathVariable Long itemId,
                                   @Valid @RequestBody DraftReviewDTO dto) {
        aiDraftService.reviewItem(id, itemId, dto);
        return Result.ok();
    }

    /** 发布草稿 */
    @Operation(summary = "发布草稿")
    @PostMapping("/{id}/publish")
    @PreAuthorize("@perm.has('odin:assistant:draft')")
    @OperationLog(module = "AI草稿", action = "发布草稿")
    public Result<Void> publish(@Parameter(name = "id", description = "草稿ID", required = true) @PathVariable Long id) {
        aiDraftService.publishDraft(id);
        return Result.ok();
    }

    /** 驳回草稿 */
    @Operation(summary = "驳回草稿")
    @PostMapping("/{id}/reject")
    @PreAuthorize("@perm.has('odin:assistant:draft')")
    @OperationLog(module = "AI草稿", action = "驳回草稿")
    public Result<Void> reject(@Parameter(name = "id", description = "草稿ID", required = true) @PathVariable Long id,
                               @Parameter(name = "reason", description = "驳回原因") @RequestParam(required = false) String reason) {
        aiDraftService.rejectDraft(id, reason);
        return Result.ok();
    }
}
