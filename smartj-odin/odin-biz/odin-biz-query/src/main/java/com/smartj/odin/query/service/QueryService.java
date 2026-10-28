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
package com.smartj.odin.query.service;

import com.smartj.common.core.domain.PageResult;
import com.smartj.odin.query.dto.QueryDryRunDTO;
import com.smartj.odin.query.dto.QueryExecuteDTO;
import com.smartj.odin.query.dto.QueryHistoryQuery;
import com.smartj.odin.query.dto.QueryTemplateDTO;
import com.smartj.odin.query.dto.QueryTemplateQuery;
import com.smartj.odin.query.vo.QueryDryRunVO;
import com.smartj.odin.query.vo.QueryHistoryVO;
import com.smartj.odin.query.vo.QueryResultVO;
import com.smartj.odin.query.vo.QueryTemplateVO;
import com.smartj.odin.query.vo.QueryTraceVO;

import java.util.Map;

/**
 * 智能问数服务接口。
 *
 * <p>覆盖自然语言问数执行、试运行校验、逻辑 SQL 解析，以及查询历史/溯源浏览与查询模板的增删改查。
 * AI 解析、Calcite 编译与 DuckDB 联邦执行在 P3 接入，当前为持久化骨架。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public interface QueryService {

    /**
     * 执行自然语言问数，记录查询历史并返回结果集。
     *
     * @param dto 执行入参
     * @return 查询结果 VO
     */
    QueryResultVO execute(QueryExecuteDTO dto);

    /**
     * 对本体逻辑 SQL 做试运行校验与物理 SQL 编译预览（不真正执行）。
     *
     * @param dto 试运行入参
     * @return 试运行结果 VO
     */
    QueryDryRunVO dryRun(QueryDryRunDTO dto);

    /**
     * 将自然语言解析为本体逻辑 SQL。
     *
     * @param naturalLanguage 自然语言问题
     * @param ontologyId      本体ID
     * @return 本体逻辑 SQL
     */
    String nlParse(String naturalLanguage, Long ontologyId);

    /**
     * 分页查询查询历史。
     *
     * @param query 查询条件
     * @return 分页结果
     */
    PageResult<QueryHistoryVO> historyPage(QueryHistoryQuery query);

    /**
     * 查询指定历史的执行溯源。
     *
     * @param historyId 查询历史ID
     * @return 溯源 VO
     */
    QueryTraceVO getTrace(Long historyId);

    /**
     * 根据ID查询查询模板详情。
     *
     * @param id 模板ID
     * @return 模板 VO
     */
    QueryTemplateVO getTemplate(Long id);

    /**
     * 分页查询查询模板。
     *
     * @param query 查询条件
     * @return 分页结果
     */
    PageResult<QueryTemplateVO> templatePage(QueryTemplateQuery query);

    /**
     * 新增查询模板，返回主键ID。
     *
     * @param dto 模板 DTO
     * @return 新记录ID
     */
    Long createTemplate(QueryTemplateDTO dto);

    /**
     * 更新查询模板。
     *
     * @param dto 模板 DTO
     */
    void updateTemplate(QueryTemplateDTO dto);

    /**
     * 根据ID删除查询模板。
     *
     * @param id 主键ID
     */
    void deleteTemplate(Long id);

    /**
     * 按模板执行查询。
     *
     * @param id     模板ID
     * @param params 模板参数
     * @return 查询结果 VO
     */
    QueryResultVO executeTemplate(Long id, Map<String, Object> params);
}
