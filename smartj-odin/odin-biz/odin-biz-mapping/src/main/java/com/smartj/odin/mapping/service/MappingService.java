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
package com.smartj.odin.mapping.service;

import com.smartj.common.core.domain.PageResult;
import com.smartj.odin.mapping.dto.MappingConfigDTO;
import com.smartj.odin.mapping.dto.MappingConfigQuery;
import com.smartj.odin.mapping.vo.MappingConfigVO;
import com.smartj.odin.mapping.vo.MappingSuggestVO;

import java.util.List;

/**
 * 映射配置服务接口。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public interface MappingService {

    /**
     * 分页查询映射配置。
     *
     * @param query 查询条件
     * @return 分页结果
     */
    PageResult<MappingConfigVO> page(MappingConfigQuery query);

    /**
     * 根据ID查询映射配置详情。
     *
     * @param id 主键ID
     * @return 映射配置 VO
     */
    MappingConfigVO getById(Long id);

    /**
     * 新增映射配置，返回主键ID。
     *
     * @param dto 映射配置 DTO
     * @return 新记录ID
     */
    Long create(MappingConfigDTO dto);

    /**
     * 更新映射配置。
     *
     * @param dto 映射配置 DTO
     */
    void update(MappingConfigDTO dto);

    /**
     * 删除映射配置。
     *
     * @param id 主键ID
     */
    void delete(Long id);

    /**
     * 校验映射完整性：检查列映射中的列是否存在于数据源 schema 中。
     *
     * @param id 映射配置ID
     * @return 校验结果描述
     */
    String validate(Long id);

    /**
     * 基于名称相似度推荐列-属性映射。
     *
     * @param ontologyId   本体ID
     * @param classId      类ID
     * @param datasourceId 数据源ID
     * @param tableName    物理表名
     * @return 映射建议列表
     */
    List<MappingSuggestVO> suggest(Long ontologyId, Long classId, Long datasourceId, String tableName);

    /**
     * 按类ID与物理表元数据ID直接推荐列映射（前端 suggest 契约）。
     *
     * @param classId 本体类ID
     * @param tableId 物理表元数据ID（schema_table.id）
     * @return 列映射建议列表
     */
    List<MappingSuggestVO> suggestByTable(Long classId, Long tableId);
}
