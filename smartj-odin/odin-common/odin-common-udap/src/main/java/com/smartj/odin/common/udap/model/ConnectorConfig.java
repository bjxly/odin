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
package com.smartj.odin.common.udap.model;

import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.ToString;

import java.io.Serial;
import java.io.Serializable;
import java.util.Map;

/**
 * 连接器配置。
 *
 * <p>描述一个外部数据源的连接参数，供 {@link com.smartj.odin.common.udap.Connector} 建立连接与内省元数据使用。
 * 当 {@link #jdbcUrl} 非空时优先使用，否则由 type + host/port/dbName 拼接。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class ConnectorConfig implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 数据源类型：mysql/postgresql 等 */
    private String type;
    /** 主机地址 */
    private String host;
    /** 端口 */
    private Integer port;
    /** 数据库名 */
    private String dbName;
    /** 模式名（可空） */
    private String schemaName;
    /** 用户名 */
    private String username;
    /** 明文密码（由调用方解密后传入，连接器不持久化；排除出 toString/equals 防止泄露） */
    @ToString.Exclude
    @EqualsAndHashCode.Exclude
    private String password;
    /** 完整 JDBC URL（非空时优先） */
    private String jdbcUrl;
    /** 扩展参数（拼接为 URL query） */
    private Map<String, String> options;
    /** 连接超时秒 */
    private int connTimeoutS = 30;
    /** 查询超时秒 */
    private int queryTimeoutS = 60;
}
