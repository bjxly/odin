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
package com.smartj.odin.insight.service;

import com.smartj.common.core.domain.PageResult;
import com.smartj.odin.insight.dto.ChangeHistoryQuery;
import com.smartj.odin.insight.vo.ChangeHistoryVO;

/**
 * 变更历史服务接口。
 *
 * <p>提供各类实体变更快照的分页审计查询与变更回滚能力（回滚逻辑 P5 接入）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public interface ChangeHistoryService {

    /**
     * 分页查询变更历史。
     *
     * @param query 查询条件
     * @return 分页结果
     */
    PageResult<ChangeHistoryVO> page(ChangeHistoryQuery query);

    /**
     * 回滚指定变更记录。
     *
     * @param id 变更历史ID
     */
    void revert(Long id);
}
