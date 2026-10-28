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
package com.smartj.odin.datasource.vo;

import lombok.Data;

import java.io.Serial;
import java.io.Serializable;
import java.time.LocalDateTime;

/**
 * 数据源 VO（展示，不含密码）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class DatasourceVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID */
    private Long id;
    /** 数据源编码 */
    private String code;
    /** 名称 */
    private String name;
    /** 类型 */
    private String type;
    /** 主机 */
    private String host;
    /** 端口 */
    private Integer port;
    /** 数据库名 */
    private String dbName;
    /** 模式名 */
    private String schemaName;
    /** 用户名 */
    private String username;
    /** 完整 JDBC URL */
    private String jdbcUrl;
    /** 连接选项(JSON) */
    private String options;
    /** 标签(JSON数组) */
    private String tags;
    /** 连接池最大连接数 */
    private Integer poolMaxSize;
    /** 连接池最小空闲 */
    private Integer poolMinIdle;
    /** 连接最大存活时间(秒) */
    private Integer poolMaxLifeS;
    /** 连接超时(秒) */
    private Integer connTimeoutS;
    /** 查询超时(秒) */
    private Integer queryTimeoutS;
    /** 状态 0=禁用 1=启用 */
    private Integer status;
    /** 备注 */
    private String remark;
    /** 创建时间 */
    private LocalDateTime createTime;
    /** 更新时间 */
    private String updateTime;
}
