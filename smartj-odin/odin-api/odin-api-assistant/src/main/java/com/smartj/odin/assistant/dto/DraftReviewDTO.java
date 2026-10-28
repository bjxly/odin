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
package com.smartj.odin.assistant.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Pattern;
import lombok.Data;

import java.io.Serial;
import java.io.Serializable;

/**
 * AI 草稿明细审阅 DTO（HITL 逐项审阅）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class DraftReviewDTO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 审阅结论：approved/rejected */
    @NotBlank(message = "审阅结论不能为空")
    @Pattern(regexp = "approved|rejected", message = "审阅结论仅支持 approved/rejected")
    private String reviewStatus;

    /** 审阅意见 */
    private String comment;
}
