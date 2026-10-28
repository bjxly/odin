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
package com.smartj.odin.query.service.impl;

import com.mybatisflex.core.paginate.Page;
import com.mybatisflex.core.query.QueryWrapper;
import com.smartj.common.auth.util.SecurityUtils;
import com.smartj.common.core.domain.PageResult;
import com.smartj.common.core.exception.BusinessException;
import com.smartj.odin.query.converter.QueryConverter;
import com.smartj.odin.query.dto.QueryDryRunDTO;
import com.smartj.odin.query.dto.QueryExecuteDTO;
import com.smartj.odin.query.dto.QueryHistoryQuery;
import com.smartj.odin.query.dto.QueryTemplateDTO;
import com.smartj.odin.query.dto.QueryTemplateQuery;
import com.smartj.odin.query.entity.QueryHistory;
import com.smartj.odin.query.entity.QueryTemplate;
import com.smartj.odin.query.entity.QueryTrace;
import com.smartj.odin.query.error.QueryErrorCode;
import com.smartj.odin.query.mapper.QueryHistoryMapper;
import com.smartj.odin.query.mapper.QueryTemplateMapper;
import com.smartj.odin.query.mapper.QueryTraceMapper;
import com.smartj.odin.query.service.QueryService;
import com.smartj.odin.query.vo.QueryDryRunVO;
import com.smartj.odin.query.vo.QueryHistoryVO;
import com.smartj.odin.query.vo.QueryResultVO;
import com.smartj.odin.query.vo.QueryTemplateVO;
import com.smartj.odin.query.vo.QueryTraceVO;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import static com.smartj.odin.query.entity.table.QueryHistoryTableDef.QUERY_HISTORY;
import static com.smartj.odin.query.entity.table.QueryTemplateTableDef.QUERY_TEMPLATE;
import static com.smartj.odin.query.entity.table.QueryTraceTableDef.QUERY_TRACE;

/**
 * 智能问数服务实现。
 *
 * <p>问数执行、试运行与逻辑 SQL 解析当前为持久化骨架：完整落库查询历史/溯源，
 * AI 解析、Calcite 编译下推与 DuckDB 联邦执行留待 P3 接入。查询历史、溯源浏览与模板 CRUD 为完整实现。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Service
@RequiredArgsConstructor
public class QueryServiceImpl implements QueryService {

    private final QueryTemplateMapper queryTemplateMapper;
    private final QueryHistoryMapper queryHistoryMapper;
    private final QueryTraceMapper queryTraceMapper;

    /** 执行自然语言问数：落库查询历史后返回结果集（AI/编译/联邦执行待 P3 接入） */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public QueryResultVO execute(QueryExecuteDTO dto) {
        long start = System.currentTimeMillis();

        // TODO: P3 接入 AI 解析 + Calcite 编译 + DuckDB 联邦执行
        // 前端可直接传入本体逻辑 SQL；缺省时回退到自然语言解析
        String logicalSql = dto.getLogicalSql();
        if (logicalSql == null || logicalSql.isBlank()) {
            logicalSql = nlParse(dto.getNaturalLanguage(), dto.getOntologyId());
        }

        QueryHistory history = new QueryHistory();
        history.setSessionId(dto.getSessionId());
        history.setUserId(SecurityUtils.getUserId());
        history.setNaturalLanguage(dto.getNaturalLanguage());
        history.setLogicalSql(logicalSql);
        history.setStatus("error");
        history.setErrorMessage("AI 问数执行功能待接入（P3）");
        history.setResultSummary("空结果（功能待接入）");
        history.setLatencyMs((int) (System.currentTimeMillis() - start));
        queryHistoryMapper.insert(history);

        QueryResultVO vo = new QueryResultVO();
        vo.setSuccess(false);
        vo.setColumns(new ArrayList<>());
        vo.setRows(new ArrayList<>());
        vo.setRowCount(0);
        vo.setLatencyMs(history.getLatencyMs());
        vo.setLogicalSql(logicalSql);
        vo.setTraceId(String.valueOf(history.getId()));
        vo.setErrorMessage(history.getErrorMessage());
        return vo;
    }

    /** 试运行：对逻辑 SQL 做基本语法校验（非空 + 含 SELECT），返回校验结论（编译与成本估算待 P3 接入） */
    @Override
    public QueryDryRunVO dryRun(QueryDryRunDTO dto) {
        String sql = dto.getLogicalSql();
        QueryDryRunVO vo = new QueryDryRunVO();
        vo.setLogicalSql(sql);

        boolean basicValid = StringUtils.hasText(sql)
                && sql.toUpperCase().contains("SELECT");
        if (!basicValid) {
            vo.setValid(false);
            vo.setErrorMessage("逻辑SQL非法：必须为包含 SELECT 的查询语句");
            vo.setWarnings(new ArrayList<>());
            return vo;
        }

        // TODO: P3 接入 Calcite 编译下推、schema 幻觉/关联/成本闸门等多层确定性校验
        vo.setValid(true);
        vo.setPhysicalSql(sql);
        vo.setEstimatedRows(0);
        List<String> warnings = new ArrayList<>();
        warnings.add("试运行仅完成基础语法校验，编译与成本估算待 P3 接入");
        vo.setWarnings(warnings);
        return vo;
    }

    /** 自然语言 → 本体逻辑 SQL（LLM 生成待 P3 接入，暂返回空串） */
    @Override
    public String nlParse(String naturalLanguage, Long ontologyId) {
        // TODO: P3 接入 LLM 本体逻辑 SQL 生成（仅面向本体逻辑 Schema，配 AST 校验）
        return "";
    }

    /** 分页查询查询历史，支持会话/用户/状态筛选 */
    @Override
    @Transactional(readOnly = true)
    public PageResult<QueryHistoryVO> historyPage(QueryHistoryQuery query) {
        QueryWrapper qw = QueryWrapper.create()
                .where(QUERY_HISTORY.SESSION_ID.eq(query.getSessionId()).when(query.getSessionId() != null))
                .and(QUERY_HISTORY.USER_ID.eq(query.getUserId()).when(query.getUserId() != null))
                .and(QUERY_HISTORY.STATUS.eq(query.getStatus()).when(StringUtils.hasText(query.getStatus())))
                .orderBy(QUERY_HISTORY.CREATE_TIME.desc());

        Page<QueryHistory> page = queryHistoryMapper.paginate(Page.of(query.getCurrent(), query.getSize()), qw);
        List<QueryHistoryVO> records = page.getRecords().stream()
                .map(QueryConverter::toHistoryVO)
                .collect(Collectors.toList());
        return PageResult.of(records, page.getTotalRow(), page.getPageNumber(), page.getPageSize());
    }

    /** 查询指定历史的执行溯源，按步骤序号升序还原链路 */
    @Override
    @Transactional(readOnly = true)
    public QueryTraceVO getTrace(Long historyId) {
        List<QueryTrace> traces = queryTraceMapper.selectListByQuery(QueryWrapper.create()
                .where(QUERY_TRACE.HISTORY_ID.eq(historyId))
                .orderBy(QUERY_TRACE.STEP_INDEX.asc()));

        QueryTraceVO vo = new QueryTraceVO();
        vo.setHistoryId(historyId == null ? null : String.valueOf(historyId));
        vo.setId(traces.isEmpty() ? null : String.valueOf(traces.get(0).getId()));
        vo.setSteps(traces.stream()
                .map(QueryConverter::toTraceStep)
                .collect(Collectors.toList()));
        return vo;
    }

    /** 根据ID查询查询模板详情 */
    @Override
    @Transactional(readOnly = true)
    public QueryTemplateVO getTemplate(Long id) {
        return QueryConverter.toTemplateVO(loadTemplate(id));
    }

    /** 分页查询查询模板，支持名称模糊、本体/状态精确筛选 */
    @Override
    @Transactional(readOnly = true)
    public PageResult<QueryTemplateVO> templatePage(QueryTemplateQuery query) {
        QueryWrapper qw = QueryWrapper.create()
                .where(QUERY_TEMPLATE.NAME.like(query.getName()).when(StringUtils.hasText(query.getName())))
                .and(QUERY_TEMPLATE.ONTOLOGY_ID.eq(query.getOntologyId()).when(query.getOntologyId() != null))
                .and(QUERY_TEMPLATE.STATUS.eq(query.getStatus()).when(query.getStatus() != null))
                .orderBy(QUERY_TEMPLATE.CREATE_TIME.desc());

        Page<QueryTemplate> page = queryTemplateMapper.paginate(Page.of(query.getCurrent(), query.getSize()), qw);
        List<QueryTemplateVO> records = page.getRecords().stream()
                .map(QueryConverter::toTemplateVO)
                .collect(Collectors.toList());
        return PageResult.of(records, page.getTotalRow(), page.getPageNumber(), page.getPageSize());
    }

    /** 新增查询模板，返回主键ID */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public Long createTemplate(QueryTemplateDTO dto) {
        QueryTemplate entity = new QueryTemplate();
        copyDtoToEntity(dto, entity);
        entity.setUsageCount(0);
        queryTemplateMapper.insert(entity);
        return entity.getId();
    }

    /** 更新查询模板：先查后改 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void updateTemplate(QueryTemplateDTO dto) {
        QueryTemplate entity = loadTemplate(dto.getId());
        copyDtoToEntity(dto, entity);
        queryTemplateMapper.update(entity);
    }

    /** 根据ID删除查询模板 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void deleteTemplate(Long id) {
        loadTemplate(id);
        queryTemplateMapper.deleteById(id);
    }

    /** 按模板执行查询：加载模板逻辑SQL，填充参数后委托 execute 流程（AI/编译/联邦执行待 P3 接入） */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public QueryResultVO executeTemplate(Long id, Map<String, Object> params) {
        QueryTemplate template = loadTemplate(id);
        // 增加使用计数
        template.setUsageCount(template.getUsageCount() == null ? 1 : template.getUsageCount() + 1);
        queryTemplateMapper.update(template);

        // TODO: P3 接入后，将模板 logicalSql + params 编译为物理 SQL 并执行
        QueryExecuteDTO dto = new QueryExecuteDTO();
        dto.setNaturalLanguage(null);
        dto.setOntologyId(template.getOntologyId());
        dto.setSessionId(null);

        QueryResultVO vo = new QueryResultVO();
        vo.setSuccess(false);
        vo.setColumns(new ArrayList<>());
        vo.setRows(new ArrayList<>());
        vo.setRowCount(0);
        vo.setLatencyMs(0);
        vo.setLogicalSql(template.getLogicalSql());
        vo.setErrorMessage("模板执行功能待接入（P3）");
        return vo;
    }

    /**
     * 加载查询模板实体，不存在抛业务异常。
     *
     * @param id 主键ID
     * @return 查询模板实体
     */
    private QueryTemplate loadTemplate(Long id) {
        QueryTemplate entity = queryTemplateMapper.selectOneById(id);
        if (entity == null) {
            throw new BusinessException(QueryErrorCode.QUERY_TEMPLATE_NOT_FOUND);
        }
        return entity;
    }

    /**
     * 将 DTO 拷贝到模板实体（不含使用次数等运行时字段）。
     *
     * @param dto    模板 DTO
     * @param entity 目标实体
     */
    private void copyDtoToEntity(QueryTemplateDTO dto, QueryTemplate entity) {
        entity.setName(dto.getName());
        entity.setDescription(dto.getDescription());
        entity.setOntologyId(dto.getOntologyId());
        entity.setLogicalSql(dto.getLogicalSql());
        entity.setParameters(dto.getParameters());
        entity.setTags(dto.getTags());
        entity.setStatus(dto.getStatus() == null ? 1 : dto.getStatus());
    }
}
