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
package com.smartj.odin.insight.service.impl;

import com.mybatisflex.core.paginate.Page;
import com.mybatisflex.core.query.QueryWrapper;
import com.smartj.common.core.domain.PageResult;
import com.smartj.common.core.exception.BusinessException;
import com.smartj.odin.insight.converter.InsightConverter;
import com.smartj.odin.insight.dto.ChangeHistoryQuery;
import com.smartj.odin.insight.entity.ChangeHistory;
import com.smartj.odin.insight.error.InsightErrorCode;
import com.smartj.odin.insight.mapper.ChangeHistoryMapper;
import com.smartj.odin.insight.service.ChangeHistoryService;
import com.smartj.odin.insight.vo.ChangeHistoryVO;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;

import java.util.List;
import java.util.stream.Collectors;

import static com.smartj.odin.insight.entity.table.ChangeHistoryTableDef.CHANGE_HISTORY;

/**
 * 变更历史服务实现。
 *
 * <p>分页审计查询为完整实现；变更回滚（依据快照反向恢复实体）留待 P5 接入，
 * 当前 {@link #revert(Long)} 为骨架实现，仅校验记录存在性。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Service
@RequiredArgsConstructor
public class ChangeHistoryServiceImpl implements ChangeHistoryService {

    private final ChangeHistoryMapper changeHistoryMapper;

    /** 分页查询变更历史，支持实体类型/实体ID/动作筛选，按创建时间倒序 */
    @Override
    @Transactional(readOnly = true)
    public PageResult<ChangeHistoryVO> page(ChangeHistoryQuery query) {
        QueryWrapper qw = QueryWrapper.create()
                .where(CHANGE_HISTORY.ENTITY_TYPE.eq(query.getEntityType()).when(StringUtils.hasText(query.getEntityType())))
                .and(CHANGE_HISTORY.ENTITY_ID.eq(query.getEntityId()).when(query.getEntityId() != null))
                .and(CHANGE_HISTORY.ACTION.eq(query.getAction()).when(StringUtils.hasText(query.getAction())))
                .orderBy(CHANGE_HISTORY.CREATE_TIME.desc());

        Page<ChangeHistory> page = changeHistoryMapper.paginate(Page.of(query.getCurrent(), query.getSize()), qw);
        List<ChangeHistoryVO> records = page.getRecords().stream()
                .map(InsightConverter::toChangeVO)
                .collect(Collectors.toList());
        return PageResult.of(records, page.getTotalRow(), page.getPageNumber(), page.getPageSize());
    }

    /** 回滚指定变更记录（依据前后快照反向恢复实体，回滚逻辑待 P5 接入） */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void revert(Long id) {
        loadChange(id);
        // TODO: P5 实现变更回滚逻辑（解析 beforeSnapshot → 按 entityType 反向恢复目标实体 → 追加 revert 变更记录）
    }

    /**
     * 加载变更历史实体，不存在抛业务异常。
     *
     * @param id 主键ID
     * @return 变更历史实体
     */
    private ChangeHistory loadChange(Long id) {
        ChangeHistory entity = changeHistoryMapper.selectOneById(id);
        if (entity == null) {
            throw new BusinessException(InsightErrorCode.CHANGE_NOT_FOUND);
        }
        return entity;
    }
}
