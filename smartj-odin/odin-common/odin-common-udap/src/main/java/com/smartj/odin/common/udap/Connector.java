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
package com.smartj.odin.common.udap;

import com.smartj.odin.common.udap.model.ColumnMeta;
import com.smartj.odin.common.udap.model.ConnectorConfig;
import com.smartj.odin.common.udap.model.ForeignKeyMeta;
import com.smartj.odin.common.udap.model.IndexMeta;
import com.smartj.odin.common.udap.model.QueryResult;
import com.smartj.odin.common.udap.model.TableMeta;
import com.smartj.odin.common.udap.model.TestResult;

import java.util.List;

/**
 * 数据源连接器 SPI。
 *
 * <p>定义连接测试与元数据内省的统一契约，屏蔽不同数据库差异。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public interface Connector {

    /**
     * 测试连接可用性。
     *
     * @param config 连接配置
     * @return 测试结果（含耗时与版本/错误信息）
     */
    TestResult testConnection(ConnectorConfig config);

    /**
     * 内省表清单。
     *
     * @param config 连接配置
     * @return 表元数据列表
     */
    List<TableMeta> listTables(ConnectorConfig config);

    /**
     * 内省指定表的列。
     *
     * @param config    连接配置
     * @param schema    模式名（可空）
     * @param tableName 表名
     * @return 列元数据列表（按序号排序）
     */
    List<ColumnMeta> listColumns(ConnectorConfig config, String schema, String tableName);

    /**
     * 内省指定表的外键。
     *
     * @param config    连接配置
     * @param schema    模式名（可空）
     * @param tableName 表名
     * @return 外键元数据列表
     */
    List<ForeignKeyMeta> listForeignKeys(ConnectorConfig config, String schema, String tableName);

    /**
     * 内省指定表的索引。
     *
     * @param config    连接配置
     * @param schema    模式名（可空）
     * @param tableName 表名
     * @return 索引元数据列表
     */
    List<IndexMeta> listIndexes(ConnectorConfig config, String schema, String tableName);

    /**
     * 执行只读查询（单源方言直查路径），返回结果集。
     *
     * <p>仅允许 SELECT 语句；通过 {@code maxRows} 限制返回行数防止大结果集拖垮服务。
     *
     * @param config  连接配置
     * @param sql     物理 SQL（仅 SELECT）
     * @param maxRows 最大返回行数
     * @return 查询结果（列名 + 行数据）
     */
    QueryResult query(ConnectorConfig config, String sql, int maxRows);
}
