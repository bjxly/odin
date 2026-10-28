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
package com.smartj.odin.mapping.converter;

import com.smartj.common.core.util.BeanCopyUtil;
import com.smartj.odin.mapping.entity.MappingConfig;
import com.smartj.odin.mapping.vo.MappingConfigVO;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

/**
 * 映射域 Entity → VO 转换器。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public final class MappingConverter {

    private static final DateTimeFormatter TIME_FMT = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

    private MappingConverter() {
    }

    /**
     * MappingConfig → MappingConfigVO。
     *
     * @param entity 映射配置实体
     * @return 映射配置 VO
     */
    public static MappingConfigVO toVO(MappingConfig entity) {
        if (entity == null) {
            return null;
        }
        MappingConfigVO vo = BeanCopyUtil.copyProperties(entity, MappingConfigVO.class);
        vo.setId(toStr(entity.getId()));
        vo.setOntologyId(toStr(entity.getOntologyId()));
        vo.setClassId(toStr(entity.getClassId()));
        vo.setDatasourceId(toStr(entity.getDatasourceId()));
        vo.setCreateTime(fmtTime(entity.getCreateTime()));
        vo.setUpdateTime(fmtTime(entity.getUpdateTime()));
        return vo;
    }

    /** Long → String，null 安全 */
    private static String toStr(Long value) {
        return value == null ? null : String.valueOf(value);
    }

    /** LocalDateTime → 格式化字符串，null 安全 */
    private static String fmtTime(LocalDateTime time) {
        return time == null ? null : time.format(TIME_FMT);
    }
}
