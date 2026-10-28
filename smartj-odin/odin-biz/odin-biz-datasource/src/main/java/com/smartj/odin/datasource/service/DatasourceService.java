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
package com.smartj.odin.datasource.service;

import com.smartj.common.core.domain.PageResult;
import com.smartj.odin.common.udap.model.ConnectorConfig;
import com.smartj.odin.datasource.dto.DatasourceDTO;
import com.smartj.odin.datasource.dto.DatasourceQuery;
import com.smartj.odin.datasource.vo.DatasourceTestResultVO;
import com.smartj.odin.datasource.vo.DatasourceVO;
import com.smartj.odin.datasource.vo.SchemaColumnVO;
import com.smartj.odin.datasource.vo.SchemaTableVO;

import java.util.List;

/**
 * 数据源服务接口。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public interface DatasourceService {

    /**
     * 分页查询数据源。
     *
     * @param query 查询条件
     * @return 分页结果
     */
    PageResult<DatasourceVO> page(DatasourceQuery query);

    /**
     * 查询全部数据源（不分页，用于下拉选择）。
     *
     * @return 数据源 VO 列表
     */
    List<DatasourceVO> listAll();

    /**
     * 根据ID查询数据源详情。
     *
     * @param id 主键ID
     * @return 数据源 VO
     */
    DatasourceVO getById(Long id);

    /**
     * 新增数据源，返回主键ID。
     *
     * @param dto 数据源 DTO
     * @return 新记录ID
     */
    Long create(DatasourceDTO dto);

    /**
     * 更新数据源。
     *
     * @param dto 数据源 DTO
     */
    void update(DatasourceDTO dto);

    /**
     * 根据ID删除数据源（级联删除其元数据）。
     *
     * @param id 主键ID
     */
    void delete(Long id);

    /**
     * 测试指定数据源连接，并记录测试结果。
     *
     * @param id 数据源ID
     * @return 测试结果 VO
     */
    DatasourceTestResultVO testConnection(Long id);

    /**
     * 同步指定数据源的表/列/外键/索引元数据。
     *
     * @param id 数据源ID
     * @return 同步的表数量
     */
    int syncMetadata(Long id);

    /**
     * 查询指定数据源已同步的表清单。
     *
     * @param datasourceId 数据源ID
     * @return 表元数据 VO 列表
     */
    List<SchemaTableVO> listTables(Long datasourceId);

    /**
     * 查询指定表已同步的列清单。
     *
     * @param tableId 表元数据ID
     * @return 列元数据 VO 列表
     */
    List<SchemaColumnVO> listColumns(Long tableId);

    /**
     * 构建指定数据源的连接器配置（含解密后的明文密码），供联邦执行/问数直查复用。
     *
     * @param id 数据源ID
     * @return 连接器配置
     */
    ConnectorConfig getConnectorConfig(Long id);
}
