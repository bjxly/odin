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
package com.smartj.odin.datasource.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;
import lombok.Data;

import java.io.Serial;
import java.io.Serializable;

/**
 * 数据源 DTO（新增/更新入参）。
 *
 * <p>{@code password} 为明文入参，服务端加密后存入 {@code password_enc}；更新时留空表示不修改密码。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class DatasourceDTO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID（更新时必填） */
    private Long id;

    /** 数据源编码 */
    @NotBlank(message = "数据源编码不能为空")
    @Size(max = 64, message = "数据源编码长度不能超过64")
    private String code;

    /** 名称 */
    @NotBlank(message = "数据源名称不能为空")
    @Size(max = 128, message = "名称长度不能超过128")
    private String name;

    /** 类型：mysql/postgresql 等 */
    @NotBlank(message = "数据源类型不能为空")
    @Size(max = 32, message = "类型长度不能超过32")
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

    /** 明文密码（更新时留空表示不修改） */
    private String password;

    /** 完整 JDBC URL（优先） */
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
    @Size(max = 500, message = "备注长度不能超过500")
    private String remark;
}
