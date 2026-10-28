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
package com.smartj.odin.datasource.entity;

import com.mybatisflex.annotation.Table;
import com.smartj.common.core.domain.BaseSimpleEntity;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.io.Serial;

/**
 * 数据源配置实体。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
@EqualsAndHashCode(callSuper = true)
@Table("odin_datasource")
public class Datasource extends BaseSimpleEntity {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 数据源编码(唯一) */
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
    /** Jasypt 加密后的密码 */
    private String passwordEnc;
    /** 完整 JDBC URL(优先) */
    private String jdbcUrl;
    /** 扩展参数(JSON) */
    private String options;
    /** 连接池最大连接数 */
    private Integer poolMaxSize;
    /** 连接池最小空闲 */
    private Integer poolMinIdle;
    /** 连接最大存活秒 */
    private Integer poolMaxLifeS;
    /** 连接超时秒 */
    private Integer connTimeoutS;
    /** 查询超时秒 */
    private Integer queryTimeoutS;
    /** 标签数组(JSON) */
    private String tags;
    /** 状态 0=禁用 1=启用 */
    private Integer status;
    /** 备注 */
    private String remark;
}
