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
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * JdbcUrlBuilder 单元测试。
 */
class JdbcUrlBuilderTest {

    @Test
    void buildsMysqlUrl() {
        ConnectorConfig c = new ConnectorConfig();
        c.setType("mysql");
        c.setHost("localhost");
        c.setPort(13306);
        c.setDbName("smartj");
        String url = JdbcUrlBuilder.build(c);
        assertTrue(url.startsWith("jdbc:mysql://localhost:13306/smartj?"));
        assertTrue(url.contains("serverTimezone=Asia/Shanghai"));
    }

    @Test
    void buildsPostgresUrl() {
        ConnectorConfig c = new ConnectorConfig();
        c.setType("postgresql");
        c.setHost("localhost");
        c.setPort(15432);
        c.setDbName("demo");
        String url = JdbcUrlBuilder.build(c);
        assertEquals("jdbc:postgresql://localhost:15432/demo", url);
    }

    @Test
    void prefersExplicitJdbcUrl() {
        ConnectorConfig c = new ConnectorConfig();
        c.setType("mysql");
        c.setJdbcUrl("jdbc:mysql://dbhost:3306/x");
        assertEquals("jdbc:mysql://dbhost:3306/x", JdbcUrlBuilder.build(c));
    }
}
