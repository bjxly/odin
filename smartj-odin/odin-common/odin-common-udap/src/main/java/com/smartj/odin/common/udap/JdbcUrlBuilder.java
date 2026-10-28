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

import com.smartj.odin.common.udap.model.ConnectorConfig;

import java.util.LinkedHashMap;
import java.util.Map;
import java.util.StringJoiner;

/**
 * JDBC URL 构造器。
 *
 * <p>根据数据源类型拼接标准 JDBC URL；当 {@link ConnectorConfig#getJdbcUrl()} 非空时直接返回。
 * 支持 mysql / postgresql / dameng / kingbase / gaussdb(opengauss) / oracle / sqlserver / sqlite / duckdb_file。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public final class JdbcUrlBuilder {

    /** 达梦默认端口 */
    private static final int DM_DEFAULT_PORT = 5236;
    /** 人大金仓默认端口 */
    private static final int KINGBASE_DEFAULT_PORT = 54321;
    /** GaussDB / openGauss 默认端口 */
    private static final int OPENGAUSS_DEFAULT_PORT = 5432;
    /** Oracle 默认端口 */
    private static final int ORACLE_DEFAULT_PORT = 1521;
    /** SQL Server 默认端口 */
    private static final int SQLSERVER_DEFAULT_PORT = 1433;

    private JdbcUrlBuilder() {
    }

    /**
     * 构造 JDBC URL。
     *
     * @param config 连接配置
     * @return 完整 JDBC URL
     */
    public static String build(ConnectorConfig config) {
        if (config.getJdbcUrl() != null && !config.getJdbcUrl().isBlank()) {
            return config.getJdbcUrl();
        }
        String type = config.getType() == null ? "" : config.getType().toLowerCase();
        return switch (type) {
            case "mysql" -> mysql(config);
            case "postgresql", "postgres" -> postgres(config);
            case "dameng", "dm" -> "jdbc:dm://" + hostPort(config, DM_DEFAULT_PORT) + "/" + nullToEmpty(config.getDbName());
            case "kingbase", "kingbase8" -> "jdbc:kingbase8://" + hostPort(config, KINGBASE_DEFAULT_PORT) + "/" + nullToEmpty(config.getDbName());
            case "gaussdb", "opengauss" -> "jdbc:opengauss://" + hostPort(config, OPENGAUSS_DEFAULT_PORT) + "/" + nullToEmpty(config.getDbName());
            case "oracle" -> "jdbc:oracle:thin:@//" + hostPort(config, ORACLE_DEFAULT_PORT) + "/" + nullToEmpty(config.getDbName());
            case "sqlserver", "mssql" -> "jdbc:sqlserver://" + hostPort(config, SQLSERVER_DEFAULT_PORT)
                    + ";databaseName=" + nullToEmpty(config.getDbName()) + ";encrypt=false";
            case "sqlite" -> "jdbc:sqlite:" + nullToEmpty(config.getDbName());
            case "duckdb_file", "duckdb" -> "jdbc:duckdb:" + nullToEmpty(config.getDbName());
            default -> throw new IllegalArgumentException("暂不支持的数据源类型: " + config.getType());
        };
    }

    /**
     * 拼接 host:port，端口为空时回落到指定默认端口。
     *
     * @param c           连接配置
     * @param defaultPort 默认端口
     * @return host:port 字符串
     */
    private static String hostPort(ConnectorConfig c, int defaultPort) {
        int port = c.getPort() == null ? defaultPort : c.getPort();
        return c.getHost() + ":" + port;
    }

    /**
     * 拼接 MySQL URL，附带时区/SSL/公钥检索等常用参数与自定义 options。
     */
    private static String mysql(ConnectorConfig c) {
        Map<String, String> params = new LinkedHashMap<>();
        params.put("useSSL", "false");
        params.put("serverTimezone", "Asia/Shanghai");
        params.put("allowPublicKeyRetrieval", "true");
        if (c.getOptions() != null) {
            params.putAll(c.getOptions());
        }
        String base = "jdbc:mysql://" + c.getHost() + ":" + c.getPort() + "/" + nullToEmpty(c.getDbName());
        return base + "?" + join(params);
    }

    /**
     * 拼接 PostgreSQL URL。
     */
    private static String postgres(ConnectorConfig c) {
        return "jdbc:postgresql://" + c.getHost() + ":" + c.getPort() + "/" + nullToEmpty(c.getDbName());
    }

    /**
     * 将参数 Map 拼为 k=v&k=v 形式。
     */
    private static String join(Map<String, String> params) {
        StringJoiner sj = new StringJoiner("&");
        params.forEach((k, v) -> sj.add(k + "=" + v));
        return sj.toString();
    }

    private static String nullToEmpty(String s) {
        return s == null ? "" : s;
    }
}
