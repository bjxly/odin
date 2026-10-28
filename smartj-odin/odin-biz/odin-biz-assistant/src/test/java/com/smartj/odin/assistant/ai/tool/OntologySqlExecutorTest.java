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
package com.smartj.odin.assistant.ai.tool;

import com.smartj.common.core.domain.PageResult;
import com.smartj.odin.common.udap.Connector;
import com.smartj.odin.common.udap.model.ConnectorConfig;
import com.smartj.odin.common.udap.model.QueryResult;
import com.smartj.odin.datasource.service.DatasourceService;
import com.smartj.odin.mapping.service.MappingService;
import com.smartj.odin.mapping.vo.MappingConfigVO;
import com.smartj.odin.ontology.service.OntologyService;
import com.smartj.odin.ontology.vo.OntClassVO;
import com.smartj.odin.ontology.vo.VocabularyVO;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.ArgumentCaptor;

import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyInt;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * {@link OntologySqlExecutor} 逻辑→物理编译回归测试。
 *
 * <p>重点守护 FROM/JOIN 类编码替换：分组反向引用 {@code $1} 必须保留为活引用，
 * 仅对物理表名做 quoteReplacement。历史缺陷曾把 {@code "$1" + qualify} 整体 quote，
 * 导致物理 SQL 出现字面量 {@code FROM $1public.customers}，被数据库当作绑定参数而报语法错。
 */
class OntologySqlExecutorTest {

    private OntologyService ontologyService;
    private MappingService mappingService;
    private DatasourceService datasourceService;
    private Connector connector;
    private OntologySqlExecutor executor;

    /** 构造本体 Customer(name,address) → public.customers(cust_name,addr) 的单源映射 */
    @BeforeEach
    void setUp() {
        ontologyService = mock(OntologyService.class);
        mappingService = mock(MappingService.class);
        datasourceService = mock(DatasourceService.class);
        connector = mock(Connector.class);

        VocabularyVO.PropertyVocab name = new VocabularyVO.PropertyVocab();
        name.setCode("name");
        name.setName("姓名");
        VocabularyVO.PropertyVocab address = new VocabularyVO.PropertyVocab();
        address.setCode("address");
        address.setName("地址");
        VocabularyVO.ClassVocab customer = new VocabularyVO.ClassVocab();
        customer.setCode("Customer");
        customer.setName("客户");
        customer.setProperties(List.of(name, address));
        VocabularyVO vocab = new VocabularyVO();
        vocab.setOntologyName("测试本体");
        vocab.setClasses(List.of(customer));
        when(ontologyService.getVocabulary(1L)).thenReturn(vocab);

        OntClassVO cls = new OntClassVO();
        cls.setId("10");
        cls.setCode("Customer");
        when(ontologyService.listClasses(1L)).thenReturn(List.of(cls));

        MappingConfigVO mapping = new MappingConfigVO();
        mapping.setClassId("10");
        mapping.setDatasourceId("6001");
        mapping.setSchemaName("public");
        mapping.setTableName("customers");
        mapping.setColumnMappings("[{\"propertyCode\":\"name\",\"columnName\":\"cust_name\"},"
                + "{\"propertyCode\":\"address\",\"columnName\":\"addr\"}]");
        mapping.setStatus(1);
        mapping.setPriority(100);
        when(mappingService.page(any())).thenReturn(PageResult.of(List.of(mapping), 1L, 1L, 10L));

        when(datasourceService.getConnectorConfig(6001L)).thenReturn(mock(ConnectorConfig.class));
        QueryResult qr = new QueryResult();
        qr.setColumns(List.of("cust_name"));
        qr.setRows(List.of(Map.of("cust_name", "张三")));
        when(connector.query(any(), anyString(), anyInt())).thenReturn(qr);

        executor = new OntologySqlExecutor(ontologyService, mappingService, datasourceService, connector);
    }

    @Test
    void fromClauseCompilesToQualifiedTableWithoutBackreferenceLeak() {
        OntologySqlExecutor.ExecOutcome outcome = executor.execute(1L, "SELECT name FROM Customer LIMIT 5", 100);

        assertTrue(outcome.isSuccess(), outcome.getError());
        ArgumentCaptor<String> sqlCap = ArgumentCaptor.forClass(String.class);
        verify(connector).query(any(), sqlCap.capture(), anyInt());
        String physical = sqlCap.getValue();

        // 反向引用未泄漏为字面量 $1
        assertFalse(physical.contains("$1"), physical);
        // FROM 后为 schema.table 全限定名，属性编码已翻译为物理列
        assertTrue(physical.contains("FROM public.customers"), physical);
        assertTrue(physical.contains("cust_name"), physical);
        assertFalse(physical.contains("Customer"), physical);
    }

    /**
     * 类编码与 SQL 关键字同名（如类 {@code Order} 与关键字 {@code ORDER}）时，
     * 4.4 步的裸类名替换不得误伤 {@code ORDER BY} 语法。历史缺陷会把 {@code ORDER BY}
     * 改写成 {@code smartj.orders BY}，导致数据库报语法错。
     */
    @Test
    void classCodeCollidingWithKeywordDoesNotCorruptOrderBy() {
        OntClassVO order = new OntClassVO();
        order.setId("20");
        order.setCode("Order");
        when(ontologyService.listClasses(1L)).thenReturn(List.of(order));

        VocabularyVO.PropertyVocab orderDate = new VocabularyVO.PropertyVocab();
        orderDate.setCode("orderDate");
        orderDate.setName("下单日期");
        VocabularyVO.PropertyVocab amount = new VocabularyVO.PropertyVocab();
        amount.setCode("amount");
        amount.setName("金额");
        VocabularyVO.ClassVocab orderCls = new VocabularyVO.ClassVocab();
        orderCls.setCode("Order");
        orderCls.setName("订单");
        orderCls.setProperties(List.of(orderDate, amount));
        VocabularyVO vocab = new VocabularyVO();
        vocab.setOntologyName("订单本体");
        vocab.setClasses(List.of(orderCls));
        when(ontologyService.getVocabulary(1L)).thenReturn(vocab);

        MappingConfigVO mapping = new MappingConfigVO();
        mapping.setClassId("20");
        mapping.setDatasourceId("6001");
        mapping.setSchemaName("smartj");
        mapping.setTableName("orders");
        mapping.setColumnMappings("[{\"propertyCode\":\"orderDate\",\"columnName\":\"order_date\"},"
                + "{\"propertyCode\":\"amount\",\"columnName\":\"amount\"}]");
        mapping.setStatus(1);
        mapping.setPriority(100);
        when(mappingService.page(any())).thenReturn(PageResult.of(List.of(mapping), 1L, 1L, 10L));

        String logical = "SELECT Order.orderDate, Order.amount FROM Order "
                + "ORDER BY Order.orderDate DESC LIMIT 8";
        OntologySqlExecutor.ExecOutcome outcome = executor.execute(1L, logical, 100);

        assertTrue(outcome.isSuccess(), outcome.getError());
        ArgumentCaptor<String> sqlCap = ArgumentCaptor.forClass(String.class);
        verify(connector).query(any(), sqlCap.capture(), anyInt());
        String physical = sqlCap.getValue();

        // ORDER BY 关键字保留，未被替换成 "<table> BY"
        assertTrue(physical.contains("ORDER BY"), physical);
        assertFalse(physical.contains("smartj.orders BY"), physical);
        // FROM 指向物理表，点引用列已 schema 限定
        assertTrue(physical.contains("FROM smartj.orders"), physical);
        assertTrue(physical.contains("smartj.orders.order_date"), physical);
    }
}
