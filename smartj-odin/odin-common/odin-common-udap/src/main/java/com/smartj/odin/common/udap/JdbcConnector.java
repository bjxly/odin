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
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.sql.Connection;
import java.sql.DatabaseMetaData;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.ResultSetMetaData;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * 基于 JDBC {@link DatabaseMetaData} 的通用连接器实现。
 *
 * <p>使用短连接（{@link DriverManager}）完成连接测试与元数据内省，避免为一次性内省维护连接池。
 * 具体支持的数据源类型由 {@link JdbcUrlBuilder} 决定。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public class JdbcConnector implements Connector {

    private static final Logger log = LoggerFactory.getLogger(JdbcConnector.class);

    /**
     * 测试连接：建立短连接并读取数据库版本。
     */
    @Override
    public TestResult testConnection(ConnectorConfig config) {
        long start = System.currentTimeMillis();
        String url = JdbcUrlBuilder.build(config);
        try (Connection conn = DriverManager.getConnection(url, config.getUsername(), config.getPassword())) {
            String version = conn.getMetaData().getDatabaseProductVersion();
            return TestResult.ok(System.currentTimeMillis() - start, version);
        } catch (Exception e) {
            String msg = e.getMessage() != null ? e.getMessage() : e.getClass().getSimpleName();
            log.warn("数据源连接测试失败: url={}, err={}", url, msg);
            return TestResult.fail(System.currentTimeMillis() - start, msg);
        }
    }

    /**
     * 内省表清单（TABLE 与 VIEW）。
     */
    @Override
    public List<TableMeta> listTables(ConnectorConfig config) {
        List<TableMeta> tables = new ArrayList<>();
        String url = JdbcUrlBuilder.build(config);
        try (Connection conn = DriverManager.getConnection(url, config.getUsername(), config.getPassword())) {
            DatabaseMetaData md = conn.getMetaData();
            String schema = resolveSchema(config, conn);
            try (ResultSet rs = md.getTables(conn.getCatalog(), schema, "%", new String[]{"TABLE", "VIEW"})) {
                while (rs.next()) {
                    TableMeta t = new TableMeta();
                    t.setSchemaName(rs.getString("TABLE_SCHEM"));
                    t.setTableName(rs.getString("TABLE_NAME"));
                    t.setTableType(rs.getString("TABLE_TYPE"));
                    t.setComment(rs.getString("REMARKS"));
                    tables.add(t);
                }
            }
        } catch (Exception e) {
            throw new IllegalStateException("内省表清单失败: " + e.getMessage(), e);
        }
        return tables;
    }

    /**
     * 内省列，并标记主键。
     */
    @Override
    public List<ColumnMeta> listColumns(ConnectorConfig config, String schema, String tableName) {
        Map<String, ColumnMeta> byName = new LinkedHashMap<>();
        String url = JdbcUrlBuilder.build(config);
        try (Connection conn = DriverManager.getConnection(url, config.getUsername(), config.getPassword())) {
            DatabaseMetaData md = conn.getMetaData();
            String sch = schema != null ? schema : resolveSchema(config, conn);
            // 主键列集合
            List<String> pkCols = new ArrayList<>();
            try (ResultSet pk = md.getPrimaryKeys(conn.getCatalog(), sch, tableName)) {
                while (pk.next()) {
                    pkCols.add(pk.getString("COLUMN_NAME"));
                }
            }
            try (ResultSet rs = md.getColumns(conn.getCatalog(), sch, tableName, "%")) {
                while (rs.next()) {
                    ColumnMeta c = new ColumnMeta();
                    c.setColumnName(rs.getString("COLUMN_NAME"));
                    c.setDataType(rs.getString("TYPE_NAME"));
                    c.setNullable("YES".equalsIgnoreCase(rs.getString("IS_NULLABLE")));
                    c.setDefaultValue(rs.getString("COLUMN_DEF"));
                    c.setComment(rs.getString("REMARKS"));
                    int size = rs.getInt("COLUMN_SIZE");
                    Integer colSize = rs.wasNull() ? null : size;
                    c.setMaxLength(colSize);
                    c.setNumericPrecision(colSize);
                    int scale = rs.getInt("DECIMAL_DIGITS");
                    c.setNumericScale(rs.wasNull() ? null : scale);
                    c.setOrdinalPosition(rs.getInt("ORDINAL_POSITION"));
                    c.setPrimaryKey(pkCols.contains(c.getColumnName()));
                    byName.put(c.getColumnName(), c);
                }
            }
        } catch (Exception e) {
            throw new IllegalStateException("内省列失败: " + e.getMessage(), e);
        }
        return new ArrayList<>(byName.values());
    }

    /**
     * 内省外键。
     */
    @Override
    public List<ForeignKeyMeta> listForeignKeys(ConnectorConfig config, String schema, String tableName) {
        List<ForeignKeyMeta> fks = new ArrayList<>();
        String url = JdbcUrlBuilder.build(config);
        try (Connection conn = DriverManager.getConnection(url, config.getUsername(), config.getPassword())) {
            DatabaseMetaData md = conn.getMetaData();
            String sch = schema != null ? schema : resolveSchema(config, conn);
            try (ResultSet rs = md.getImportedKeys(conn.getCatalog(), sch, tableName)) {
                while (rs.next()) {
                    ForeignKeyMeta fk = new ForeignKeyMeta();
                    fk.setConstraintName(rs.getString("FK_NAME"));
                    fk.setColumnName(rs.getString("FKCOLUMN_NAME"));
                    fk.setRefSchema(rs.getString("PKTABLE_SCHEM"));
                    fk.setRefTable(rs.getString("PKTABLE_NAME"));
                    fk.setRefColumn(rs.getString("PKCOLUMN_NAME"));
                    fk.setDeleteRule(fkRuleName(rs.getShort("DELETE_RULE")));
                    fk.setUpdateRule(fkRuleName(rs.getShort("UPDATE_RULE")));
                    fks.add(fk);
                }
            }
        } catch (Exception e) {
            throw new IllegalStateException("内省外键失败: " + e.getMessage(), e);
        }
        return fks;
    }

    /**
     * 内省索引（按索引名聚合列）。
     */
    @Override
    public List<IndexMeta> listIndexes(ConnectorConfig config, String schema, String tableName) {
        Map<String, IndexMeta> byName = new LinkedHashMap<>();
        String url = JdbcUrlBuilder.build(config);
        try (Connection conn = DriverManager.getConnection(url, config.getUsername(), config.getPassword())) {
            DatabaseMetaData md = conn.getMetaData();
            String sch = schema != null ? schema : resolveSchema(config, conn);
            try (ResultSet rs = md.getIndexInfo(conn.getCatalog(), sch, tableName, false, true)) {
                while (rs.next()) {
                    short type = rs.getShort("TYPE");
                    if (type == DatabaseMetaData.tableIndexStatistic) {
                        continue; // 表统计信息行，非真实索引
                    }
                    String name = rs.getString("INDEX_NAME");
                    if (name == null) {
                        continue; // 统计信息行
                    }
                    IndexMeta idx = byName.computeIfAbsent(name, k -> {
                        IndexMeta m = new IndexMeta();
                        m.setIndexName(k);
                        m.setColumnNames(new ArrayList<>());
                        return m;
                    });
                    idx.setUnique(!rs.getBoolean("NON_UNIQUE"));
                    idx.setIndexType(indexTypeName(type));
                    idx.getColumnNames().add(rs.getString("COLUMN_NAME"));
                }
            }
        } catch (Exception e) {
            throw new IllegalStateException("内省索引失败: " + e.getMessage(), e);
        }
        return new ArrayList<>(byName.values());
    }

    /**
     * 执行只读查询：短连接 + PreparedStatement，限制返回行数与查询超时。
     */
    @Override
    public QueryResult query(ConnectorConfig config, String sql, int maxRows) {
        if (sql == null || !sql.trim().toUpperCase().startsWith("SELECT")) {
            throw new IllegalArgumentException("仅允许执行 SELECT 查询");
        }
        int limit = maxRows <= 0 ? 200 : maxRows;
        String url = JdbcUrlBuilder.build(config);
        QueryResult result = new QueryResult();
        result.setColumns(new ArrayList<>());
        result.setRows(new ArrayList<>());
        try (Connection conn = DriverManager.getConnection(url, config.getUsername(), config.getPassword())) {
            try (PreparedStatement ps = conn.prepareStatement(sql)) {
                int timeout = config.getQueryTimeoutS() <= 0 ? 60 : config.getQueryTimeoutS();
                ps.setQueryTimeout(timeout);
                try (ResultSet rs = ps.executeQuery()) {
                    ResultSetMetaData md = rs.getMetaData();
                    int colCount = md.getColumnCount();
                    for (int i = 1; i <= colCount; i++) {
                        result.getColumns().add(md.getColumnLabel(i));
                    }
                    int count = 0;
                    while (rs.next()) {
                        if (count >= limit) {
                            result.setTruncated(true);
                            break;
                        }
                        Map<String, Object> row = new LinkedHashMap<>();
                        for (int i = 1; i <= colCount; i++) {
                            row.put(md.getColumnLabel(i), rs.getObject(i));
                        }
                        result.getRows().add(row);
                        count++;
                    }
                    result.setRowCount(count);
                }
            }
        } catch (Exception e) {
            throw new IllegalStateException("执行查询失败: " + e.getMessage(), e);
        }
        return result;
    }

    /**
     * 解析内省使用的 schema：优先配置，其次连接的当前 schema/catalog。
     */
    private String resolveSchema(ConnectorConfig config, Connection conn) throws Exception {
        if (config.getSchemaName() != null && !config.getSchemaName().isBlank()) {
            return config.getSchemaName();
        }
        String type = config.getType() == null ? "" : config.getType().toLowerCase();
        // MySQL 用 catalog（数据库名）过滤，schema 传 null；PostgreSQL 用 schema
        if (type.contains("postgres")) {
            return conn.getSchema();
        }
        return null;
    }

    /**
     * 将 {@link DatabaseMetaData} 外键级联规则码映射为可读名称。
     *
     * @param rule importedKey* 常量
     * @return 规则名称（CASCADE/RESTRICT/SET NULL/NO ACTION/SET DEFAULT）
     */
    private String fkRuleName(short rule) {
        return switch (rule) {
            case DatabaseMetaData.importedKeyCascade -> "CASCADE";
            case DatabaseMetaData.importedKeyRestrict -> "RESTRICT";
            case DatabaseMetaData.importedKeySetNull -> "SET NULL";
            case DatabaseMetaData.importedKeyNoAction -> "NO ACTION";
            case DatabaseMetaData.importedKeySetDefault -> "SET DEFAULT";
            default -> "NO ACTION";
        };
    }

    /**
     * 将 {@link DatabaseMetaData#getIndexInfo} 的 TYPE 码映射为索引类型名称。
     *
     * @param type tableIndex* 常量
     * @return 索引类型（CLUSTERED/HASHED/OTHER），默认 BTREE
     */
    private String indexTypeName(short type) {
        return switch (type) {
            case DatabaseMetaData.tableIndexClustered -> "CLUSTERED";
            case DatabaseMetaData.tableIndexHashed -> "HASHED";
            case DatabaseMetaData.tableIndexOther -> "OTHER";
            default -> "BTREE";
        };
    }
}
