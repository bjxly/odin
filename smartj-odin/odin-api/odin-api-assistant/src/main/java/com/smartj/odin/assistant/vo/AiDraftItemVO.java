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
package com.smartj.odin.assistant.vo;

import lombok.Data;

import java.io.Serial;
import java.io.Serializable;

/**
 * AI 草稿明细项 VO（展示）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class AiDraftItemVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID */
    private String id;
    /** 所属草稿ID */
    private String draftId;
    /** 明细项类型：class/property/relation/mapping 等 */
    private String itemType;
    /** 变更动作：add/update/delete */
    private String action;
    /** 目标实体ID */
    private String targetId;
    /** 变更前快照（JSON） */
    private String beforeSnapshot;
    /** 变更后内容（JSON，AI 建议） */
    private String afterSnapshot;
    /** 审阅状态：pending/approved/rejected */
    private String reviewStatus;
    /** 审阅意见 */
    private String reviewerComment;
    /** 创建时间 */
    private String createTime;
}
