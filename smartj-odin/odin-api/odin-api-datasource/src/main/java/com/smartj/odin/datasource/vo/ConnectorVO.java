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

/**
 * 连接器 VO（展示）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class ConnectorVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID */
    private String id;
    /** 连接器名称 */
    private String name;
    /** 连接器编码 */
    private String code;
    /** 驱动类名 */
    private String driverClass;
    /** URL 模板 */
    private String urlPattern;
    /** 数据库类型 */
    private String dbType;
    /** 默认端口 */
    private Integer defaultPort;
    /** 描述 */
    private String description;
    /** 状态 0=禁用 1=启用 */
    private Integer status;
    /** 备注 */
    private String remark;
    /** 创建时间 */
    private String createTime;
    /** 更新时间 */
    private String updateTime;
}
