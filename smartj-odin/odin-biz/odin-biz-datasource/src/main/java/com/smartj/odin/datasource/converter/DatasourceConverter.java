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
package com.smartj.odin.datasource.converter;

import com.smartj.common.core.util.BeanCopyUtil;
import com.smartj.odin.datasource.entity.Datasource;
import com.smartj.odin.datasource.vo.DatasourceVO;

/**
 * 数据源 Entity → VO 转换器。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public final class DatasourceConverter {

    private DatasourceConverter() {
    }

    /**
     * Entity → VO 转换（不含密码字段）。
     *
     * @param datasource 数据源实体
     * @return 数据源 VO
     */
    public static DatasourceVO toVO(Datasource datasource) {
        return BeanCopyUtil.copyProperties(datasource, DatasourceVO.class);
    }
}
