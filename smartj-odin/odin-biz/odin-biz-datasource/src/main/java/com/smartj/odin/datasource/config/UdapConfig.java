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
package com.smartj.odin.datasource.config;

import com.smartj.odin.common.udap.Connector;
import com.smartj.odin.common.udap.JdbcConnector;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * 统一数据接入（UDAP）连接器装配。
 *
 * <p>{@code odin-common-udap} 模块不依赖 Spring，故在本业务模块以 {@link Configuration} 声明
 * {@link Connector} Bean，交由容器管理并注入到服务层。连接器无状态，可安全复用单例。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Configuration
public class UdapConfig {

    /**
     * 声明基于 JDBC 的连接器单例。
     *
     * @return JDBC 连接器实现
     */
    @Bean
    public Connector jdbcConnector() {
        return new JdbcConnector();
    }
}
