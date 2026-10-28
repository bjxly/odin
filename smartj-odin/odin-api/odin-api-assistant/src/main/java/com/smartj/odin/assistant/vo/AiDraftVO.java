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
import java.util.List;

/**
 * AI 草稿头 VO（展示）。
 *
 * <p>列表场景仅返回头部字段；详情场景（{@code getDraft}）额外填充 {@link #items} 明细项，供 HITL 逐项审阅。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class AiDraftVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID */
    private String id;
    /** 草稿类型：datasource/ontology/mapping/action/insight */
    private String draftType;
    /** 本体ID */
    private String ontologyId;
    /** 本体名称 */
    private String ontologyName;
    /** 草稿状态：draft/approved/rejected/published */
    private String status;
    /** 是否由 AI 生成 0=否 1=是 */
    private Integer createdByAi;
    /** AI 说明/摘要 */
    private String summary;
    /** 明细项数量 */
    private Integer itemCount;
    /** 创建时间 */
    private String createTime;
    /** 明细项列表（详情场景填充） */
    private List<AiDraftItemVO> items;
}
