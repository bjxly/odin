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
package com.smartj.odin.assistant.converter;

import com.smartj.common.core.util.BeanCopyUtil;
import com.smartj.odin.assistant.entity.AiDraft;
import com.smartj.odin.assistant.entity.AiDraftItem;
import com.smartj.odin.assistant.entity.AssistantMessage;
import com.smartj.odin.assistant.entity.AssistantSession;
import com.smartj.odin.assistant.vo.AiDraftItemVO;
import com.smartj.odin.assistant.vo.AiDraftVO;
import com.smartj.odin.assistant.vo.AssistantMessageVO;
import com.smartj.odin.assistant.vo.AssistantSessionVO;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

/**
 * 助手域 Entity → VO 转换器。
 *
 * <p>统一将 ID 类字段转为字符串、时间字段格式化为 {@code yyyy-MM-dd HH:mm:ss}，
 * 规避前端大整数精度丢失并保证展示格式一致。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public final class AssistantConverter {

    /** 统一时间格式 */
    private static final DateTimeFormatter FMT = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

    private AssistantConverter() {
    }

    /**
     * 会话 Entity → VO。
     *
     * @param entity 会话实体
     * @return 会话 VO
     */
    public static AssistantSessionVO toSessionVO(AssistantSession entity) {
        if (entity == null) {
            return null;
        }
        AssistantSessionVO vo = BeanCopyUtil.copyProperties(entity, AssistantSessionVO.class);
        vo.setId(str(entity.getId()));
        vo.setUserId(str(entity.getUserId()));
        vo.setOntologyId(str(entity.getOntologyId()));
        vo.setCreateTime(fmt(entity.getCreateTime()));
        vo.setUpdateTime(fmt(entity.getUpdateTime()));
        return vo;
    }

    /**
     * 消息 Entity → VO。
     *
     * @param entity 消息实体
     * @return 消息 VO
     */
    public static AssistantMessageVO toMessageVO(AssistantMessage entity) {
        if (entity == null) {
            return null;
        }
        AssistantMessageVO vo = BeanCopyUtil.copyProperties(entity, AssistantMessageVO.class);
        vo.setId(str(entity.getId()));
        vo.setSessionId(str(entity.getSessionId()));
        vo.setCreateTime(fmt(entity.getCreateTime()));
        return vo;
    }

    /**
     * AI 草稿头 Entity → VO（不含明细项）。
     *
     * @param entity 草稿头实体
     * @return 草稿头 VO
     */
    public static AiDraftVO toDraftVO(AiDraft entity) {
        if (entity == null) {
            return null;
        }
        AiDraftVO vo = BeanCopyUtil.copyProperties(entity, AiDraftVO.class);
        vo.setId(str(entity.getId()));
        vo.setOntologyId(str(entity.getOntologyId()));
        vo.setCreateTime(fmt(entity.getCreateTime()));
        return vo;
    }

    /**
     * AI 草稿明细项 Entity → VO。
     *
     * @param entity 草稿明细项实体
     * @return 草稿明细项 VO
     */
    public static AiDraftItemVO toDraftItemVO(AiDraftItem entity) {
        if (entity == null) {
            return null;
        }
        AiDraftItemVO vo = BeanCopyUtil.copyProperties(entity, AiDraftItemVO.class);
        vo.setId(str(entity.getId()));
        vo.setDraftId(str(entity.getDraftId()));
        vo.setTargetId(str(entity.getTargetId()));
        vo.setCreateTime(fmt(entity.getCreateTime()));
        return vo;
    }

    /**
     * Long → 字符串（null 安全）。
     *
     * @param value 数值
     * @return 字符串或 null
     */
    private static String str(Long value) {
        return value == null ? null : String.valueOf(value);
    }

    /**
     * LocalDateTime → 格式化字符串（null 安全）。
     *
     * @param time 时间
     * @return 格式化字符串或 null
     */
    private static String fmt(LocalDateTime time) {
        return time == null ? null : time.format(FMT);
    }
}
