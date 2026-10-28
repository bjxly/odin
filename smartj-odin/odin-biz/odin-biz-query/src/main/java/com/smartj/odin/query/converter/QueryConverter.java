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
package com.smartj.odin.query.converter;

import com.smartj.common.core.util.BeanCopyUtil;
import com.smartj.odin.query.entity.QueryHistory;
import com.smartj.odin.query.entity.QueryTemplate;
import com.smartj.odin.query.entity.QueryTrace;
import com.smartj.odin.query.vo.QueryHistoryVO;
import com.smartj.odin.query.vo.QueryTemplateVO;
import com.smartj.odin.query.vo.QueryTraceVO;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

/**
 * 查询域 Entity → VO 转换器。
 *
 * <p>统一将 ID 类字段转为字符串、时间字段格式化为 {@code yyyy-MM-dd HH:mm:ss}，
 * 规避前端大整数精度丢失并保证展示格式一致。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public final class QueryConverter {

    /** 统一时间格式 */
    private static final DateTimeFormatter FMT = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

    private QueryConverter() {
    }

    /**
     * 查询模板 Entity → VO。
     *
     * @param entity 查询模板实体
     * @return 查询模板 VO
     */
    public static QueryTemplateVO toTemplateVO(QueryTemplate entity) {
        if (entity == null) {
            return null;
        }
        QueryTemplateVO vo = BeanCopyUtil.copyProperties(entity, QueryTemplateVO.class);
        vo.setId(str(entity.getId()));
        vo.setOntologyId(str(entity.getOntologyId()));
        vo.setCreateTime(fmt(entity.getCreateTime()));
        vo.setUpdateTime(fmt(entity.getUpdateTime()));
        return vo;
    }

    /**
     * 查询历史 Entity → VO。
     *
     * @param entity 查询历史实体
     * @return 查询历史 VO
     */
    public static QueryHistoryVO toHistoryVO(QueryHistory entity) {
        if (entity == null) {
            return null;
        }
        QueryHistoryVO vo = BeanCopyUtil.copyProperties(entity, QueryHistoryVO.class);
        vo.setId(str(entity.getId()));
        vo.setSessionId(str(entity.getSessionId()));
        vo.setUserId(str(entity.getUserId()));
        vo.setCreateTime(fmt(entity.getCreateTime()));
        return vo;
    }

    /**
     * 查询溯源 Entity → 溯源步骤 VO。
     *
     * @param entity 查询溯源实体
     * @return 溯源步骤 VO
     */
    public static QueryTraceVO.TraceStep toTraceStep(QueryTrace entity) {
        if (entity == null) {
            return null;
        }
        QueryTraceVO.TraceStep step = BeanCopyUtil.copyProperties(entity, QueryTraceVO.TraceStep.class);
        step.setDatasourceId(str(entity.getDatasourceId()));
        return step;
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
