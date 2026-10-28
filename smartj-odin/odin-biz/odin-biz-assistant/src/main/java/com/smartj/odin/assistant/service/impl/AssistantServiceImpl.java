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
package com.smartj.odin.assistant.service.impl;

import com.mybatisflex.core.paginate.Page;
import com.mybatisflex.core.query.QueryWrapper;
import com.smartj.common.auth.util.SecurityUtils;
import com.smartj.common.core.domain.PageResult;
import com.smartj.common.core.exception.BusinessException;
import com.smartj.odin.assistant.ai.AssistantAiClient;
import com.smartj.odin.assistant.ai.config.AgentScopeProperties;
import com.smartj.odin.assistant.ai.tool.DatasetSink;
import com.smartj.odin.assistant.ai.tool.OntologyMutateTool;
import com.smartj.odin.assistant.ai.tool.OntologyQueryTool;
import com.smartj.odin.assistant.ai.tool.OntologySqlExecutor;
import com.smartj.odin.assistant.converter.AssistantConverter;
import com.smartj.odin.assistant.dto.ChatRequestDTO;
import com.smartj.odin.assistant.dto.SessionCreateDTO;
import com.smartj.odin.assistant.dto.SessionQuery;
import com.smartj.odin.assistant.entity.AssistantMessage;
import com.smartj.odin.assistant.entity.AssistantSession;
import com.smartj.odin.assistant.error.AssistantErrorCode;
import com.smartj.odin.assistant.mapper.AiDraftItemMapper;
import com.smartj.odin.assistant.mapper.AiDraftMapper;
import com.smartj.odin.assistant.mapper.AssistantMessageMapper;
import com.smartj.odin.assistant.mapper.AssistantSessionMapper;
import com.smartj.odin.assistant.service.AssistantService;
import com.smartj.odin.assistant.service.ChatContext;
import com.smartj.odin.assistant.vo.AssistantMessageVO;
import com.smartj.odin.assistant.vo.AssistantSessionVO;
import com.smartj.odin.assistant.vo.AssistantSkillVO;
import com.smartj.odin.assistant.vo.ChatResponseVO;
import com.smartj.odin.datasource.service.DatasourceService;
import com.smartj.odin.ontology.service.OntologyService;
import com.smartj.odin.ontology.vo.VocabularyVO;
import io.agentscope.core.tool.Toolkit;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

import static com.smartj.odin.assistant.entity.table.AssistantMessageTableDef.ASSISTANT_MESSAGE;
import static com.smartj.odin.assistant.entity.table.AssistantSessionTableDef.ASSISTANT_SESSION;

/**
 * 助手服务实现。
 *
 * <p>对话经 Agentscope-java {@code ReActAgent} + OpenAI 兼容模型（默认 DeepSeek）生成真实回复，
 * 携带同会话历史实现多轮上下文；当 {@code odin.ai.enabled=false}、缺少密钥或模型调用失败时，
 * 回退为占位回复，保证功能可用且不影响会话持久化。技能列表、会话/消息管理为完整实现。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class AssistantServiceImpl implements AssistantService {

    /** 会话默认标题最大长度（首条问题截取前 N 个字作为会话名） */
    private static final int TITLE_MAX_LEN = 20;

    /** AI 未启用或调用失败时的占位回复 */
    private static final String FALLBACK_REPLY = "AI 助手功能待接入（P3）。当前已记录您的问题，稍后将由智能体为您解答。";

    private final AssistantSessionMapper assistantSessionMapper;
    private final AssistantMessageMapper assistantMessageMapper;
    private final AssistantAiClient aiClient;
    private final AgentScopeProperties aiProps;
    private final OntologyService ontologyService;
    private final OntologySqlExecutor ontologySqlExecutor;
    private final DatasourceService datasourceService;
    private final AiDraftMapper aiDraftMapper;
    private final AiDraftItemMapper aiDraftItemMapper;

    /** 处理一次对话：创建/获取会话，保存用户消息，经 AI 生成助手回复（未启用则回退占位） */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public ChatResponseVO chat(ChatRequestDTO dto) {
        long start = System.currentTimeMillis();

        AssistantSession session = resolveSession(dto);
        insertUserMessage(session.getId(), dto.getMessage());

        // 经 Agentscope-java ReActAgent 生成回复（携带同会话历史多轮上下文 + 本体词表/问数工具）；未启用或失败则回退占位
        ReplyOutcome outcome = generateReply(session.getId(), dto.getOntologyId());
        String replyContent = outcome.content();
        AssistantMessage assistantMsg = new AssistantMessage();
        assistantMsg.setSessionId(session.getId());
        assistantMsg.setRole("assistant");
        assistantMsg.setContent(replyContent);
        assistantMsg.setContentType("text");
        assistantMsg.setLatencyMs((int) (System.currentTimeMillis() - start));
        assistantMessageMapper.insert(assistantMsg);

        // 更新会话消息计数（用户 + 助手）
        int count = session.getMessageCount() == null ? 0 : session.getMessageCount();
        session.setMessageCount(count + 2);
        assistantSessionMapper.update(session);

        ChatResponseVO vo = new ChatResponseVO();
        vo.setSessionId(String.valueOf(session.getId()));
        vo.setMessageId(String.valueOf(assistantMsg.getId()));
        vo.setContent(replyContent);
        vo.setContentType("text");
        vo.setDone(true);
        // 本轮若产生 AI 草稿，回传首个草稿ID与明细数，前端据此提示「有草稿待审阅」
        if (!outcome.draftIds().isEmpty()) {
            vo.setDraftId(String.valueOf(outcome.draftIds().get(0)));
            vo.setDraftItemCount(outcome.draftItemCount());
        }
        return vo;
    }

    /**
     * 为流式对话做准备（请求线程、事务内）：会话/用户消息落库 + 词表/提示词/工具集/上下文构建。
     */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public ChatContext prepareChat(ChatRequestDTO dto) {
        String traceId = java.util.UUID.randomUUID().toString().replace("-", "");
        AssistantSession session = resolveSession(dto);
        insertUserMessage(session.getId(), dto.getMessage());

        if (!aiClient.isEnabled()) {
            return new ChatContext(traceId, session.getId(), SecurityUtils.getUserId(), dto.getOntologyId(),
                    null, null, List.of(), null, null, false, FALLBACK_REPLY);
        }

        DatasetSink datasetSink = new DatasetSink(traceId);
        Tooling tooling = buildTooling(dto.getOntologyId(), datasetSink);
        List<AssistantAiClient.HistoryTurn> conversation = buildConversation(session.getId());
        return new ChatContext(traceId, session.getId(), SecurityUtils.getUserId(), dto.getOntologyId(),
                tooling.sysPrompt(), tooling.toolkit(), conversation, tooling.mutateTool(), datasetSink, true, null);
    }

    /**
     * 流式结束后回写助手消息（异步线程，由调用方传播 SecurityContext 后进入本事务）。
     */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public Long persistAssistantReply(ChatContext ctx, String content, String metadataJson, int latencyMs, Integer tokensUsed) {
        AssistantMessage assistantMsg = new AssistantMessage();
        assistantMsg.setSessionId(ctx.sessionId());
        assistantMsg.setRole("assistant");
        assistantMsg.setContent(content);
        assistantMsg.setContentType("markdown");
        assistantMsg.setMetadata(metadataJson);
        assistantMsg.setLatencyMs(latencyMs);
        assistantMsg.setTokensUsed(tokensUsed);
        assistantMessageMapper.insert(assistantMsg);

        AssistantSession session = assistantSessionMapper.selectOneById(ctx.sessionId());
        if (session != null) {
            int count = session.getMessageCount() == null ? 0 : session.getMessageCount();
            session.setMessageCount(count + 2);
            assistantSessionMapper.update(session);
        }
        return assistantMsg.getId();
    }

    /** 获取或创建会话（与 {@link #chat} 行为一致，供同步与流式复用）。 */
    private AssistantSession resolveSession(ChatRequestDTO dto) {
        if (dto.getSessionId() == null) {
            AssistantSession session = new AssistantSession();
            session.setUserId(SecurityUtils.getUserId());
            session.setTitle(buildTitle(dto.getMessage()));
            session.setOntologyId(dto.getOntologyId());
            session.setStatus("active");
            session.setMessageCount(0);
            assistantSessionMapper.insert(session);
            return session;
        }
        AssistantSession session = loadSession(dto.getSessionId());
        // 首条消息进入默认命名会话时，用问题内容作为标题
        // （前端「新建会话」按钮会先建一个占位会话，此处补标题，避免一直显示「新会话」）
        int count = session.getMessageCount() == null ? 0 : session.getMessageCount();
        boolean defaultTitle = !StringUtils.hasText(session.getTitle()) || "新会话".equals(session.getTitle());
        if (count == 0 && defaultTitle) {
            session.setTitle(buildTitle(dto.getMessage()));
            assistantSessionMapper.update(session);
        }
        return session;
    }

    /** 落库用户消息。 */
    private void insertUserMessage(Long sessionId, String message) {
        AssistantMessage userMsg = new AssistantMessage();
        userMsg.setSessionId(sessionId);
        userMsg.setRole("user");
        userMsg.setContent(message);
        userMsg.setContentType("text");
        assistantMessageMapper.insert(userMsg);
    }

    /** 获取助手技能列表（硬编码，P3 起由 Agentscope 工具注册表动态生成） */
    @Override
    public List<AssistantSkillVO> getSkills() {
        List<AssistantSkillVO> skills = new ArrayList<>();
        skills.add(buildSkill("data_query", "数据查询", "基于本体以自然语言查询多源业务数据", "search"));
        skills.add(buildSkill("ontology_browse", "本体浏览", "浏览类、属性与关系构成的本体图谱", "share"));
        skills.add(buildSkill("mapping_suggest", "映射建议", "为本体与物理数据源智能推荐字段映射", "link"));
        skills.add(buildSkill("data_insight", "数据洞察", "配置监控规则并对业务异常进行预警解读", "trend"));
        return skills;
    }

    /** 分页查询会话，支持用户/本体/状态筛选 */
    @Override
    @Transactional(readOnly = true)
    public PageResult<AssistantSessionVO> sessionPage(SessionQuery query) {
        QueryWrapper qw = QueryWrapper.create()
                .where(ASSISTANT_SESSION.USER_ID.eq(query.getUserId()).when(query.getUserId() != null))
                .and(ASSISTANT_SESSION.ONTOLOGY_ID.eq(query.getOntologyId()).when(query.getOntologyId() != null))
                .and(ASSISTANT_SESSION.STATUS.eq(query.getStatus()).when(StringUtils.hasText(query.getStatus())))
                .orderBy(ASSISTANT_SESSION.CREATE_TIME.desc());

        Page<AssistantSession> page = assistantSessionMapper.paginate(Page.of(query.getCurrent(), query.getSize()), qw);
        List<AssistantSessionVO> records = page.getRecords().stream()
                .map(AssistantConverter::toSessionVO)
                .collect(Collectors.toList());
        return PageResult.of(records, page.getTotalRow(), page.getPageNumber(), page.getPageSize());
    }

    /** 查询指定会话的消息列表（按创建时间升序） */
    @Override
    @Transactional(readOnly = true)
    public List<AssistantMessageVO> getMessages(Long sessionId) {
        List<AssistantMessage> messages = assistantMessageMapper.selectListByQuery(QueryWrapper.create()
                .where(ASSISTANT_MESSAGE.SESSION_ID.eq(sessionId))
                .orderBy(ASSISTANT_MESSAGE.CREATE_TIME.asc())
                .orderBy(ASSISTANT_MESSAGE.ID.asc()));
        return messages.stream()
                .map(AssistantConverter::toMessageVO)
                .collect(Collectors.toList());
    }

    /** 创建会话，返回主键ID */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public Long createSession(SessionCreateDTO dto) {
        AssistantSession session = new AssistantSession();
        session.setUserId(SecurityUtils.getUserId());
        session.setTitle(StringUtils.hasText(dto.getTitle()) ? dto.getTitle() : "新会话");
        session.setOntologyId(dto.getOntologyId());
        session.setStatus("active");
        session.setMessageCount(0);
        assistantSessionMapper.insert(session);
        return session.getId();
    }

    /** 删除会话，并级联删除其消息 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void deleteSession(Long sessionId) {
        loadSession(sessionId);
        assistantMessageMapper.deleteByQuery(QueryWrapper.create()
                .where(ASSISTANT_MESSAGE.SESSION_ID.eq(sessionId)));
        assistantSessionMapper.deleteById(sessionId);
    }

    /**
     * 加载会话实体，不存在抛业务异常。
     *
     * @param sessionId 会话ID
     * @return 会话实体
     */
    private AssistantSession loadSession(Long sessionId) {
        AssistantSession session = assistantSessionMapper.selectOneById(sessionId);
        if (session == null) {
            throw new BusinessException(AssistantErrorCode.SESSION_NOT_FOUND);
        }
        return session;
    }

    /**
     * 由首条用户消息构造会话标题（超长截断）。
     *
     * @param message 用户消息
     * @return 会话标题
     */
    private String buildTitle(String message) {
        if (!StringUtils.hasText(message)) {
            return "新会话";
        }
        String trimmed = message.trim();
        return trimmed.length() > TITLE_MAX_LEN ? trimmed.substring(0, TITLE_MAX_LEN) : trimmed;
    }

    /**
     * 一次 AI 对话的产出。
     *
     * @param content        助手回复文本
     * @param draftIds       本轮产生的草稿ID（可能为空）
     * @param draftItemCount 本轮草稿明细项总数
     */
    private record ReplyOutcome(String content, List<Long> draftIds, int draftItemCount) {
    }

    /**
     * 经 Agentscope ReActAgent 生成回复。
     *
     * <p>绑定本体时注入词表系统提示词，并注册只读问数工具与 HITL 写工具；
     * 未启用 AI 或调用异常时回退占位回复，不影响会话持久化。
     *
     * @param sessionId  会话ID
     * @param ontologyId 当前本体ID（可空）
     * @return 回复产出（含本轮草稿信息）
     */
    private ReplyOutcome generateReply(Long sessionId, Long ontologyId) {
        if (!aiClient.isEnabled()) {
            return new ReplyOutcome(FALLBACK_REPLY, List.of(), 0);
        }
        try {
            Tooling tooling = buildTooling(ontologyId, null);
            String reply = aiClient.chat(buildConversation(sessionId), tooling.sysPrompt(), tooling.toolkit());
            String content = StringUtils.hasText(reply) ? reply : FALLBACK_REPLY;
            OntologyMutateTool mutateTool = tooling.mutateTool();
            if (mutateTool == null) {
                return new ReplyOutcome(content, List.of(), 0);
            }
            return new ReplyOutcome(content, mutateTool.getProducedDraftIds(), mutateTool.getProducedItemCount());
        } catch (Exception e) {
            log.error("AI 助手对话失败，回退占位回复: sessionId={}", sessionId, e);
            return new ReplyOutcome(FALLBACK_REPLY, List.of(), 0);
        }
    }

    /**
     * 一次对话的工具装配产出。
     *
     * @param sysPrompt  系统提示词（未绑定本体或词表为空时为 null，由调用方回退默认）
     * @param toolkit    工具集（同上，为 null 表示无工具）
     * @param mutateTool 写工具实例（用于流/调用后收集草稿ID，可空）
     */
    private record Tooling(String sysPrompt, Toolkit toolkit, OntologyMutateTool mutateTool) {
    }

    /**
     * 按当前本体构建系统提示词与工具集（只读问数工具 + HITL 写工具）。
     *
     * <p>未绑定本体或词表为空时返回全 null，调用方使用默认提示词且不注册工具。
     *
     * @param ontologyId  本体ID（可空）
     * @param datasetSink 数据集侧信道（可空；流式路径传入以便问数结果随流下发，同步路径传 null）
     * @return 工具装配产出
     */
    private Tooling buildTooling(Long ontologyId, DatasetSink datasetSink) {
        if (ontologyId == null) {
            return new Tooling(null, null, null);
        }
        VocabularyVO vocab = loadVocabularySafely(ontologyId);
        if (vocab == null) {
            return new Tooling(null, null, null);
        }
        String sysPrompt = buildOntologyPrompt(vocab);
        Toolkit toolkit = new Toolkit();
        toolkit.registerTool(new OntologyQueryTool(ontologySqlExecutor, ontologyId, datasetSink));
        OntologyMutateTool mutateTool = new OntologyMutateTool(aiDraftMapper, aiDraftItemMapper,
                ontologyService, datasourceService, ontologyId, vocab.getOntologyName());
        toolkit.registerTool(mutateTool);
        return new Tooling(sysPrompt, toolkit, mutateTool);
    }

    /**
     * 安全加载本体词表：异常或空词表返回 null（调用方回退默认提示词、不注册工具）。
     *
     * @param ontologyId 本体ID
     * @return 词表 VO 或 null
     */
    private VocabularyVO loadVocabularySafely(Long ontologyId) {
        try {
            VocabularyVO vocab = ontologyService.getVocabulary(ontologyId);
            if (vocab == null || vocab.getClasses() == null || vocab.getClasses().isEmpty()) {
                return null;
            }
            return vocab;
        } catch (Exception e) {
            log.warn("加载本体词表失败 ontologyId={}", ontologyId, e);
            return null;
        }
    }

    /**
     * 构建含本体词表的系统提示词：默认提示 + 本体逻辑 Schema（类/属性/关系）+ 问数与写操作指引。
     *
     * @param vocab 本体词表（非空）
     * @return 系统提示词
     */
    private String buildOntologyPrompt(VocabularyVO vocab) {
        StringBuilder sb = new StringBuilder();
        if (StringUtils.hasText(aiProps.getSystemPrompt())) {
            sb.append(aiProps.getSystemPrompt()).append("\n\n");
        }
        sb.append("当前本体：").append(vocab.getOntologyName()).append("\n");
        sb.append("本体逻辑 Schema（编写逻辑 SQL 时只能使用以下类编码与属性编码）：\n");
        for (VocabularyVO.ClassVocab c : vocab.getClasses()) {
            sb.append("- 类 ").append(c.getCode()).append("（").append(c.getName()).append("）属性：");
            if (c.getProperties() != null && !c.getProperties().isEmpty()) {
                List<String> props = new ArrayList<>();
                for (VocabularyVO.PropertyVocab p : c.getProperties()) {
                    props.add(p.getCode() + "(" + p.getName() + "," + p.getDataType() + ")");
                }
                sb.append(String.join(", ", props));
            } else {
                sb.append("（无）");
            }
            sb.append("\n");
        }
        if (vocab.getRelations() != null && !vocab.getRelations().isEmpty()) {
            sb.append("关系（可用于 JOIN）：\n");
            for (VocabularyVO.RelationVocab r : vocab.getRelations()) {
                sb.append("- ").append(r.getSourceClass()).append(" -[")
                        .append(r.getCode()).append("]-> ").append(r.getTargetClass()).append("\n");
            }
        }
        sb.append("\n问数指引：当用户询问数据时，请基于上述本体逻辑 Schema 编写逻辑 SQL（表=类编码、列=属性编码），")
                .append("调用 run_ontology_query 工具执行并基于返回的真实数据作答；不要编造数据，不要使用物理表名/列名。")
                .append("每次只执行回答该问题所需的最小查询，通常一条 SQL 就足够；禁止顺手执行与问题无关的补充查询（除非用户明确要求额外分析）。")
                .append("编写 SQL 时请为聚合列和计算列指定中文别名（如 COUNT(*) AS 客户数、SUM(amount) AS 累计金额），让结果表头直接以中文展示。");
        sb.append("\n写操作指引：当用户要求新建/修改本体类、属性、类-物理表映射或数据源时，")
                .append("调用对应的 propose_create_class / propose_create_property / propose_create_mapping / ")
                .append("propose_create_datasource 工具生成草稿。这些工具只产生待人工审阅的草稿，不会立即生效；")
                .append("调用后必须如实告知用户「已生成草稿，请在草稿审阅面板逐项批准后发布」，")
                .append("严禁声称已直接创建或修改成功。数据源不要询问或接收密码。");
        sb.append("\n图表推荐指引：展示明细数据时鼓励在正文中使用 Markdown 表格。图表应优先基于正文中支撑最终结论的分析表（而不是原始查询结果），")
                .append("确保图文口径一致：当分析表就是结论口径时，在回复最后一行输出机器可读标记指明要成图的表、图型与列索引：")
                .append("[[chart_hint dataId=tbl-xxxxxxxx-1 type=bar x=1 y=3,4]]。其中 dataId 按正文表格出现顺序从 1 编号（tbl-为固定前缀，xxxxxxxx 可省略）；")
                .append("type 取 bar/line/pie；x 为类别列的 0 基列索引；y 为一个或多个指标列 0 基列索引（逗号分隔）。")
                .append("若类别在表头上（如本月/上月、环比多期对比），额外加空格 t=1 表示转置后以首行为类别。")
                .append("类别轴应选语义明确的名称列（如客户名称）而非编号/ID 列；若没有任何表适合成图或图会与结论矛盾，宁可不输出该标记，只留表格。")
                .append("不确定时不输出该标记。");
        return sb.toString();
    }

    /**
     * 构建注入模型的对话上下文：按时间升序取本会话消息，尾部截断至 {@code odin.ai.max-history} 条。
     *
     * @param sessionId 会话ID
     * @return 有序对话轮次（末条为当前用户消息）
     */
    private List<AssistantAiClient.HistoryTurn> buildConversation(Long sessionId) {
        List<AssistantMessage> messages = assistantMessageMapper.selectListByQuery(QueryWrapper.create()
                .where(ASSISTANT_MESSAGE.SESSION_ID.eq(sessionId))
                .orderBy(ASSISTANT_MESSAGE.CREATE_TIME.asc())
                .orderBy(ASSISTANT_MESSAGE.ID.asc()));
        int maxHistory = aiProps.getMaxHistory() == null ? 20 : aiProps.getMaxHistory();
        if (maxHistory > 0 && messages.size() > maxHistory) {
            messages = messages.subList(messages.size() - maxHistory, messages.size());
        }
        List<AssistantAiClient.HistoryTurn> turns = new ArrayList<>();
        for (AssistantMessage message : messages) {
            if (!StringUtils.hasText(message.getContent())) {
                continue;
            }
            turns.add(new AssistantAiClient.HistoryTurn(message.getRole(), message.getContent()));
        }
        return turns;
    }

    /**
     * 构造技能 VO。
     *
     * @param code        技能编码
     * @param name        技能名称
     * @param description 技能描述
     * @param icon        图标标识
     * @return 技能 VO
     */
    private AssistantSkillVO buildSkill(String code, String name, String description, String icon) {
        AssistantSkillVO skill = new AssistantSkillVO();
        skill.setCode(code);
        skill.setName(name);
        skill.setDescription(description);
        skill.setIcon(icon);
        return skill;
    }
}
