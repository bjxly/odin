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
package com.smartj.odin.assistant.service;

import com.smartj.common.core.domain.PageResult;
import com.smartj.odin.assistant.dto.DraftQuery;
import com.smartj.odin.assistant.dto.DraftReviewDTO;
import com.smartj.odin.assistant.vo.AiDraftVO;

/**
 * AI 草稿服务接口（HITL 审阅与发布）。
 *
 * <p>草稿由 AI 产出，经逐项审阅后发布为正式资源；发布/回滚的事务落地在 P4 完善。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public interface AiDraftService {

    /**
     * 分页查询 AI 草稿。
     *
     * @param query 查询条件
     * @return 分页结果
     */
    PageResult<AiDraftVO> draftPage(DraftQuery query);

    /**
     * 查询草稿详情（含明细项）。
     *
     * @param id 草稿ID
     * @return 草稿 VO（items 已填充）
     */
    AiDraftVO getDraft(Long id);

    /**
     * 审阅单条草稿明细项。
     *
     * @param draftId 草稿ID
     * @param itemId  明细项ID
     * @param dto     审阅入参
     */
    void reviewItem(Long draftId, Long itemId, DraftReviewDTO dto);

    /**
     * 发布草稿（将已通过的明细项落地为正式资源）。
     *
     * @param draftId 草稿ID
     */
    void publishDraft(Long draftId);

    /**
     * 驳回草稿。
     *
     * @param draftId 草稿ID
     * @param reason  驳回原因
     */
    void rejectDraft(Long draftId, String reason);
}
