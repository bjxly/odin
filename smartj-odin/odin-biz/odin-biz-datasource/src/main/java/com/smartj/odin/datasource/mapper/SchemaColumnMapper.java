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
package com.smartj.odin.datasource.mapper;

import com.mybatisflex.core.BaseMapper;
import com.smartj.odin.datasource.entity.SchemaColumn;
import org.apache.ibatis.annotations.Mapper;

/**
 * 列元数据 Mapper。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Mapper
public interface SchemaColumnMapper extends BaseMapper<SchemaColumn> {
}
