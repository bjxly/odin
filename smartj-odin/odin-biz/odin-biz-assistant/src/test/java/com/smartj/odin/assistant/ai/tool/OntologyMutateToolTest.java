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

import com.smartj.odin.assistant.entity.AiDraft;
import com.smartj.odin.assistant.entity.AiDraftItem;
import com.smartj.odin.assistant.mapper.AiDraftItemMapper;
import com.smartj.odin.assistant.mapper.AiDraftMapper;
import com.smartj.odin.datasource.service.DatasourceService;
import com.smartj.odin.datasource.vo.DatasourceVO;
import com.smartj.odin.ontology.service.OntologyService;
import com.smartj.odin.ontology.vo.VocabularyVO;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.ArgumentCaptor;

import java.util.List;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * {@link OntologyMutateTool} 单元测试：只验校验分支与草稿落库参数，不触真实数据库。
 */
class OntologyMutateToolTest {

    private AiDraftMapper draftMapper;
    private AiDraftItemMapper itemMapper;
    private OntologyService ontologyService;
    private DatasourceService datasourceService;
    private OntologyMutateTool tool;

    /** 构造仅含 Customer(name) 一个类的词表，并模拟 insert 回填主键 */
    @BeforeEach
    void setUp() {
        draftMapper = mock(AiDraftMapper.class);
        itemMapper = mock(AiDraftItemMapper.class);
        ontologyService = mock(OntologyService.class);
        datasourceService = mock(DatasourceService.class);

        VocabularyVO.PropertyVocab p = new VocabularyVO.PropertyVocab();
        p.setCode("name");
        p.setName("姓名");
        VocabularyVO.ClassVocab c = new VocabularyVO.ClassVocab();
        c.setCode("Customer");
        c.setName("客户");
        c.setProperties(List.of(p));
        VocabularyVO vocab = new VocabularyVO();
        vocab.setOntologyName("测试本体");
        vocab.setClasses(List.of(c));
        when(ontologyService.getVocabulary(1L)).thenReturn(vocab);

        // MyBatis-Flex insert 后回填主键，此处模拟该行为
        when(draftMapper.insert(any(AiDraft.class))).thenAnswer(inv -> {
            inv.getArgument(0, AiDraft.class).setId(100L);
            return 1;
        });

        tool = new OntologyMutateTool(draftMapper, itemMapper, ontologyService, datasourceService, 1L, "测试本体");
    }

    @Test
    void proposeCreateClassWritesDraftAndItem() {
        String out = tool.proposeCreateClass("Order", "订单", "销售订单");

        assertTrue(out.contains("草稿"), out);
        ArgumentCaptor<AiDraft> draftCap = ArgumentCaptor.forClass(AiDraft.class);
        verify(draftMapper).insert(draftCap.capture());
        assertEquals("ontology", draftCap.getValue().getDraftType());
        assertEquals("draft", draftCap.getValue().getStatus());
        assertEquals(Integer.valueOf(1), draftCap.getValue().getCreatedByAi());
        assertEquals(Long.valueOf(1L), draftCap.getValue().getOntologyId());

        ArgumentCaptor<AiDraftItem> itemCap = ArgumentCaptor.forClass(AiDraftItem.class);
        verify(itemMapper).insert(itemCap.capture());
        AiDraftItem item = itemCap.getValue();
        assertEquals(Long.valueOf(100L), item.getDraftId());
        assertEquals("class", item.getItemType());
        assertEquals("add", item.getAction());
        assertEquals("pending", item.getReviewStatus());
        assertTrue(item.getAfterSnapshot().contains("Order"), item.getAfterSnapshot());

        assertEquals(List.of(100L), tool.getProducedDraftIds());
        assertEquals(1, tool.getProducedItemCount());
    }

    @Test
    void duplicateClassCodeIsRejectedWithoutDraft() {
        String out = tool.proposeCreateClass("Customer", "客户", null);

        assertTrue(out.contains("已存在"), out);
        verify(draftMapper, never()).insert(any(AiDraft.class));
        verify(itemMapper, never()).insert(any(AiDraftItem.class));
        assertTrue(tool.getProducedDraftIds().isEmpty());
    }

    @Test
    void propertyOfUnknownClassIsRejected() {
        String out = tool.proposeCreateProperty("Order", "amount", "金额", "decimal");

        assertTrue(out.contains("不存在"), out);
        verify(itemMapper, never()).insert(any(AiDraftItem.class));
    }

    @Test
    void mappingWithUnknownPropertyCodeIsRejected() {
        String out = tool.proposeCreateMapping("Customer", "7", "t_customer",
                "[{\"propertyCode\":\"phone\",\"columnName\":\"tel\"}]");

        assertTrue(out.contains("不存在"), out);
        verify(itemMapper, never()).insert(any(AiDraftItem.class));
    }

    @Test
    void mappingWithValidColumnsProposesItem() {
        DatasourceVO ds = new DatasourceVO();
        ds.setId(7L);
        when(datasourceService.getById(7L)).thenReturn(ds);

        String out = tool.proposeCreateMapping("Customer", "7", "t_customer",
                "[{\"propertyCode\":\"name\",\"columnName\":\"cust_name\"}]");

        assertTrue(out.contains("草稿"), out);
        ArgumentCaptor<AiDraftItem> itemCap = ArgumentCaptor.forClass(AiDraftItem.class);
        verify(itemMapper).insert(itemCap.capture());
        assertEquals("mapping", itemCap.getValue().getItemType());
        assertEquals("mapping", itemCapForDraftType());
    }

    /** 断言草稿类型为 mapping（读取首次 insert 的草稿） */
    private String itemCapForDraftType() {
        ArgumentCaptor<AiDraft> draftCap = ArgumentCaptor.forClass(AiDraft.class);
        verify(draftMapper).insert(draftCap.capture());
        return draftCap.getValue().getDraftType();
    }

    @Test
    void unsupportedDatasourceTypeIsRejected() {
        String out = tool.proposeCreateDatasource("ds_mongo", "Mongo库", "mongodb", "db.local", "27017", "app", null, "sa");

        assertTrue(out.contains("不支持"), out);
        verify(draftMapper, never()).insert(any(AiDraft.class));
    }

    @Test
    void datasourceDraftNeverCarriesPassword() {
        String out = tool.proposeCreateDatasource("ds_erp", "ERP库", "mysql", "db.local", "3306", "erp", null, "erp_user");

        assertTrue(out.contains("密码"), out);
        ArgumentCaptor<AiDraftItem> itemCap = ArgumentCaptor.forClass(AiDraftItem.class);
        verify(itemMapper).insert(itemCap.capture());
        assertEquals("datasource", itemCap.getValue().getItemType());
        assertTrue(!itemCap.getValue().getAfterSnapshot().contains("password"), itemCap.getValue().getAfterSnapshot());
    }
}
