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
package com.smartj.odin.assistant.stream;

import com.smartj.odin.assistant.ai.AssistantAiClient;
import com.smartj.odin.assistant.ai.config.AgentScopeProperties;
import com.smartj.odin.assistant.ai.tool.DatasetSnapshot;
import com.smartj.odin.assistant.ai.tool.MarkdownTableParser;
import com.smartj.odin.assistant.dto.ChatRequestDTO;
import com.smartj.odin.assistant.dto.ChatStreamRequestDTO;
import com.smartj.odin.assistant.service.AssistantService;
import com.smartj.odin.assistant.service.ChatContext;
import io.agentscope.core.event.AgentEvent;
import io.agentscope.core.event.AgentResultEvent;
import io.agentscope.core.event.ModelCallEndEvent;
import io.agentscope.core.event.TextBlockDeltaEvent;
import io.agentscope.core.event.TextBlockEndEvent;
import io.agentscope.core.event.ThinkingBlockDeltaEvent;
import io.agentscope.core.event.ThinkingBlockEndEvent;
import io.agentscope.core.event.ToolCallStartEvent;
import io.agentscope.core.message.Msg;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.security.core.context.SecurityContext;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;
import tools.jackson.databind.ObjectMapper;

import java.io.IOException;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.regex.Pattern;

/**
 * 助手 SSE 流式对话编排。
 *
 * <p>请求线程内完成 {@link AssistantService#prepareChat}（会话/用户消息落库、词表、提示词、工具集、上下文），
 * 再把 {@code streamEvents} 的逐 token 事件翻译为向后兼容的 SSE 事件写入 {@link SseEmitter}；流结束后把
 * 聚合全文与块结构 JSON（{@code metadata}）经 {@link AssistantService#persistAssistantReply} 回写。
 *
 * <p>异步线程无 SecurityContext，故在提交任务前捕获、任务内重设、{@code finally} 清理，保证末尾落库的
 * 审计字段与租户隔离正确。客户端断开（{@code IOException}）后静默停止下发，但继续消费流以完成落库。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Slf4j
@Service
public class AssistantStreamService {

    /** SSE 空闲超时（毫秒），与既有端点保持一致 */
    private static final long SSE_TIMEOUT_MS = 300_000L;

    /** AI 不可用或无有效回复时的占位全文（与同步端点语义一致） */
    private static final String FALLBACK_REPLY =
            "AI 助手功能待接入（P3）。当前已记录您的问题，稍后将由智能体为您解答。";

    /** 正文中的图表推荐标记：[[chart_hint dataId=.. type=.. x=.. y=..[,..] [t=1]]]（键值对任意顺序，t 可选） */
    private static final Pattern CHART_HINT = Pattern.compile("\\[\\[chart_hint\\s+([^\\]]+)\\]\\]");

    private final AssistantService assistantService;
    private final AssistantAiClient aiClient;
    private final AgentScopeProperties props;
    private final ExecutorService streamExecutor;
    private final ObjectMapper objectMapper;

    public AssistantStreamService(AssistantService assistantService,
                                  AssistantAiClient aiClient,
                                  AgentScopeProperties props,
                                  @Qualifier("assistantStreamExecutor") ExecutorService streamExecutor,
                                  ObjectMapper objectMapper) {
        this.assistantService = assistantService;
        this.aiClient = aiClient;
        this.props = props;
        this.streamExecutor = streamExecutor;
        this.objectMapper = objectMapper;
    }

    /**
     * 处理一次流式对话：请求线程准备上下文并捕获登录态，异步线程订阅事件流并下发 SSE。
     *
     * @param dto 流式请求
     * @return 已提交异步任务的 emitter
     */
    public SseEmitter stream(ChatStreamRequestDTO dto) {
        ChatRequestDTO chatReq = new ChatRequestDTO();
        chatReq.setMessage(dto.getMessage());
        chatReq.setOntologyId(dto.getOntologyId());
        if (StringUtils.hasText(dto.getSessionId())) {
            chatReq.setSessionId(Long.valueOf(dto.getSessionId().trim()));
        }

        // 请求线程：登录态完备，完成全部前置落库与上下文构建
        ChatContext ctx = assistantService.prepareChat(chatReq);
        SecurityContext securityContext = SecurityContextHolder.getContext();

        SseEmitter emitter = new SseEmitter(SSE_TIMEOUT_MS);
        ClientGate gate = new ClientGate();
        emitter.onCompletion(gate::disconnect);
        emitter.onTimeout(gate::disconnect);
        emitter.onError((ex) -> gate.disconnect());

        streamExecutor.submit(() -> runStream(ctx, emitter, gate, securityContext));
        return emitter;
    }

    /**
     * 异步线程主体：重设登录态 → 订阅事件流 → 聚合落库 → 收尾。
     */
    private void runStream(ChatContext ctx, SseEmitter emitter, ClientGate gate, SecurityContext securityContext) {
        long start = System.currentTimeMillis();
        SecurityContextHolder.setContext(securityContext);
        Accumulator acc = new Accumulator();
        try {
            Map<String, Object> meta = new LinkedHashMap<>();
            meta.put("traceId", ctx.traceId());
            meta.put("sessionId", String.valueOf(ctx.sessionId()));
            meta.put("skillId", "data_query");
            sendEvent(emitter, gate, "meta", meta);

            // 数据集侧信道：工具执行线程回调 → 即时下发 dataset 事件
            if (ctx.datasetSink() != null) {
                ctx.datasetSink().setHandler((snapshot) -> sendEvent(emitter, gate, "dataset", toDatasetMap(snapshot)));
            }

            String content;
            if (!ctx.aiEnabled() || !props.isStream()) {
                // AI 关闭或流式回退：走一次性全文单块（向后兼容旧协议）
                content = StringUtils.hasText(ctx.fallbackContent())
                        ? ctx.fallbackContent()
                        : runNonStream(ctx, acc);
                Map<String, Object> block = new LinkedHashMap<>();
                block.put("type", "text");
                block.put("content", content);
                sendEvent(emitter, gate, "block", block);
            } else {
                aiClient.chatStream(ctx.conversation(), ctx.sysPrompt(), ctx.toolkit(),
                        (event) -> translate(emitter, gate, acc, event), null);
                content = acc.mergedText();
                if (!StringUtils.hasText(content)) {
                    content = acc.resultText;
                }
                if (!StringUtils.hasText(content)) {
                    content = FALLBACK_REPLY;
                }
            }

            // 分析表侧信道：确定性解析正文 Markdown 结论表 → 合成数据集（图与结论口径一致）
            List<DatasetSnapshot> tables = MarkdownTableParser.extract(content, ctx.traceId());
            for (DatasetSnapshot t : tables) {
                sendEvent(emitter, gate, "dataset", toTableDatasetMap(t));
            }

            // 提取并剔除正文中的 chart_hint 标记，按 dataId 关联数据集/分析表下发
            List<Map<String, Object>> hints = extractChartHints(ctx, tables, emitter, gate, acc, content);
            content = acc.cleanContent(content);
            finishAndPersist(ctx, emitter, gate, content, buildMetadataJson(ctx, tables, acc, content, hints), start, acc);
        } catch (Exception e) {
            log.error("流式助手处理失败: traceId={}", ctx.traceId(), e);
            sendError(emitter, gate, e);
            safeComplete(emitter);
        } finally {
            SecurityContextHolder.clearContext();
        }
    }

    /**
     * 从正文提取 {@code [[chart_hint ...]]} 标记，校验后写入累加器并下发 chart_hint 事件。
     *
     * <p>校验：dataId 必须命中本轮数据集或分析表、type∈{bar,line,pie}、x/y 索引合法（转置提示跳过，
     * 由前端在转置视图上终校）；不合法则丢弃（前端回退启发式）。
     */
    private List<Map<String, Object>> extractChartHints(ChatContext ctx, List<DatasetSnapshot> tables,
                                                        SseEmitter emitter, ClientGate gate,
                                                        Accumulator acc, String rawContent) {
        List<Map<String, Object>> valid = new ArrayList<>();
        java.util.regex.Matcher m = CHART_HINT.matcher(rawContent == null ? "" : rawContent);
        while (m.find()) {
            Map<String, String> kv = parseHintAttrs(m.group(1));
            String dataId = kv.get("dataId");
            String type = kv.get("type");
            if (!StringUtils.hasText(dataId) || !("bar".equals(type) || "line".equals(type) || "pie".equals(type))) {
                continue;
            }
            Integer x = parseInt(kv.get("x"));
            List<Integer> ys = parseIndexList(kv.get("y"));
            if (x == null || ys.isEmpty()) {
                continue;
            }
            boolean transposed = "1".equals(kv.get("t"));
            DatasetSnapshot target = resolveHintTarget(ctx, tables, dataId);
            if (target != null) {
                dataId = target.dataId();
            }
            Integer cols = target == null ? null : (target.columns() == null ? 0 : target.columns().size());
            if (cols != null && !transposed) {
                if (x < 0 || x >= cols) {
                    continue;
                }
                final int xFinal = x;
                boolean bad = ys.stream().anyMatch(y -> y < 0 || y >= cols || y == xFinal);
                if (bad) {
                    continue;
                }
            }
            Map<String, Object> hint = new LinkedHashMap<>();
            hint.put("dataId", dataId);
            hint.put("type", type);
            hint.put("xIndex", x);
            hint.put("yIndices", ys);
            if (transposed) {
                hint.put("transposed", true);
            }
            valid.add(hint);
            acc.chartHints.put(dataId, hint);
            sendEvent(emitter, gate, "chart_hint", hint);
        }
        return valid;
    }

    /** 解析 chart_hint 标记体内的键值对（{@code key=value}，空格分隔，value 不含空格）。 */
    private static Map<String, String> parseHintAttrs(String body) {
        Map<String, String> kv = new LinkedHashMap<>();
        for (String token : body.trim().split("\\s+")) {
            int eq = token.indexOf('=');
            if (eq > 0) {
                kv.put(token.substring(0, eq), token.substring(eq + 1));
            }
        }
        return kv;
    }

    private static Integer parseInt(String s) {
        if (s == null) {
            return null;
        }
        try {
            return Integer.parseInt(s.trim());
        } catch (NumberFormatException e) {
            return null;
        }
    }

    private static List<Integer> parseIndexList(String s) {
        List<Integer> list = new ArrayList<>();
        if (s == null) {
            return list;
        }
        for (String p : s.split(",")) {
            Integer v = parseInt(p);
            if (v != null) {
                list.add(v);
            }
        }
        return list;
    }

    /**
     * 解析提示目标：查询数据集按 dataId 精确命中；分析表先精确匹配，再按尾序号 {@code -N} 回退
     *（模型不知道 traceId 前缀，常把 dataId 写成 {@code tbl-xxxxxxxx-N} 占位形式）。
     */
    private DatasetSnapshot resolveHintTarget(ChatContext ctx, List<DatasetSnapshot> tables, String dataId) {
        if (dataId == null) {
            return null;
        }
        if (dataId.startsWith(MarkdownTableParser.TABLE_ID_PREFIX)) {
            for (DatasetSnapshot t : tables) {
                if (dataId.equals(t.dataId())) {
                    return t;
                }
            }
            int idx = dataId.lastIndexOf('-');
            if (idx > 0) {
                Integer n = parseInt(dataId.substring(idx + 1));
                if (n != null && n >= 1 && n <= tables.size()) {
                    return tables.get(n - 1);
                }
            }
            return null;
        }
        return findDataset(ctx, dataId);
    }

    /** 按 dataId 在本轮数据集中查找快照。 */
    private DatasetSnapshot findDataset(ChatContext ctx, String dataId) {
        if (ctx.datasetSink() == null || dataId == null) {
            return null;
        }
        for (DatasetSnapshot s : ctx.datasetSink().collected()) {
            if (dataId.equals(s.dataId())) {
                return s;
            }
        }
        return null;
    }

    /** 数据集快照 → SSE dataset 事件 payload。 */
    private Map<String, Object> toDatasetMap(DatasetSnapshot s) {
        Map<String, Object> m = new LinkedHashMap<>();
        m.put("dataId", s.dataId());
        m.put("id", s.dataId());
        m.put("title", s.title());
        m.put("columns", s.columns());
        m.put("rows", s.rows());
        m.put("rowCount", s.rowCount());
        m.put("truncated", s.truncated());
        m.put("physicalSql", s.physicalSql());
        return m;
    }

    /** 分析表数据集 → SSE dataset 事件 payload（source=table 供前端区分口径优先级）。 */
    private Map<String, Object> toTableDatasetMap(DatasetSnapshot s) {
        Map<String, Object> m = toDatasetMap(s);
        m.put("source", "table");
        return m;
    }

    /**
     * 非流式回退（{@code odin.ai.stream=false}）：同步调用一次取全文，填入累加器供落库。
     */
    private String runNonStream(ChatContext ctx, Accumulator acc) {
        try {
            String reply = aiClient.chat(ctx.conversation(), ctx.sysPrompt(), ctx.toolkit());
            if (StringUtils.hasText(reply)) {
                acc.appendText("full", reply);
                return reply;
            }
        } catch (Exception e) {
            log.error("非流式助手回退失败: traceId={}", ctx.traceId(), e);
        }
        return FALLBACK_REPLY;
    }

    /**
     * 把单个 Agentscope 事件翻译为 SSE 事件并更新累加器。
     */
    private void translate(SseEmitter emitter, ClientGate gate, Accumulator acc, AgentEvent event) {
        switch (event.getType()) {
            case TEXT_BLOCK_DELTA -> {
                TextBlockDeltaEvent e = (TextBlockDeltaEvent) event;
                acc.appendText(e.getBlockId(), e.getDelta());
                Map<String, Object> data = new LinkedHashMap<>();
                data.put("replyId", e.getReplyId());
                data.put("blockId", e.getBlockId());
                data.put("delta", e.getDelta());
                sendEvent(emitter, gate, "text_delta", data);
            }
            case TEXT_BLOCK_END -> {
                TextBlockEndEvent e = (TextBlockEndEvent) event;
                Map<String, Object> data = new LinkedHashMap<>();
                data.put("replyId", e.getReplyId());
                data.put("blockId", e.getBlockId());
                sendEvent(emitter, gate, "text_end", data);
            }
            case THINKING_BLOCK_DELTA -> {
                ThinkingBlockDeltaEvent e = (ThinkingBlockDeltaEvent) event;
                acc.appendThinking(e.getBlockId(), e.getDelta());
                Map<String, Object> data = new LinkedHashMap<>();
                data.put("replyId", e.getReplyId());
                data.put("blockId", e.getBlockId());
                data.put("delta", e.getDelta());
                sendEvent(emitter, gate, "thinking_delta", data);
            }
            case THINKING_BLOCK_END -> {
                ThinkingBlockEndEvent e = (ThinkingBlockEndEvent) event;
                Map<String, Object> data = new LinkedHashMap<>();
                data.put("blockId", e.getBlockId());
                sendEvent(emitter, gate, "thinking_end", data);
            }
            case TOOL_CALL_START -> {
                ToolCallStartEvent e = (ToolCallStartEvent) event;
                Map<String, Object> data = new LinkedHashMap<>();
                data.put("toolName", e.getToolCallName());
                sendEvent(emitter, gate, "tool_call", data);
            }
            case MODEL_CALL_END -> {
                ModelCallEndEvent e = (ModelCallEndEvent) event;
                if (e.getUsage() != null) {
                    acc.usageInput = e.getUsage().getInputTokens();
                    acc.usageOutput = e.getUsage().getOutputTokens();
                }
            }
            case AGENT_RESULT -> {
                AgentResultEvent e = (AgentResultEvent) event;
                Msg result = e.getResult();
                if (result != null && StringUtils.hasText(result.getTextContent())) {
                    acc.resultText = result.getTextContent();
                }
            }
            default -> {
                // 其余事件（START/工具结果/数据块等）本期不下发，保持协议精简
            }
        }
    }

    /**
     * 流末尾：落库助手消息（全文 + metadata），下发 draft/done 并完成 emitter。
     */
    private void finishAndPersist(ChatContext ctx, SseEmitter emitter, ClientGate gate,
                                  String content, String metadataJson, long start, Accumulator acc) {
        Integer tokens = acc.usageInput + acc.usageOutput > 0 ? acc.usageInput + acc.usageOutput : null;
        Long assistantMsgId = null;
        try {
            assistantMsgId = assistantService.persistAssistantReply(ctx, content, metadataJson,
                    (int) (System.currentTimeMillis() - start), tokens);
        } catch (Exception e) {
            log.error("流式助手回写失败: traceId={}", ctx.traceId(), e);
        }

        if (ctx.mutateTool() != null && !ctx.mutateTool().getProducedDraftIds().isEmpty()) {
            Map<String, Object> draft = new LinkedHashMap<>();
            draft.put("draftId", String.valueOf(ctx.mutateTool().getProducedDraftIds().get(0)));
            draft.put("itemCount", ctx.mutateTool().getProducedItemCount());
            sendEvent(emitter, gate, "draft", draft);
        }

        Map<String, Object> done = new LinkedHashMap<>();
        done.put("sessionId", String.valueOf(ctx.sessionId()));
        done.put("messageId", assistantMsgId == null ? null : String.valueOf(assistantMsgId));
        done.put("traceId", ctx.traceId());
        Map<String, Object> usage = new LinkedHashMap<>();
        usage.put("inputTokens", acc.usageInput);
        usage.put("outputTokens", acc.usageOutput);
        done.put("usage", usage);
        done.put("latencyMs", System.currentTimeMillis() - start);
        sendEvent(emitter, gate, "done", done);
        safeComplete(emitter);
    }

    /**
     * 构建块结构 JSON（供历史还原）：{@code blocks}（reasoning 在前、text 在后）+ {@code datasets}（含分析表）+ {@code chartHints}。
     */
    private String buildMetadataJson(ChatContext ctx, List<DatasetSnapshot> tables, Accumulator acc,
                                     String content, List<Map<String, Object>> hints) {
        List<Map<String, Object>> blocks = new ArrayList<>();
        for (Map.Entry<String, StringBuilder> en : acc.thinkingBlocks.entrySet()) {
            Map<String, Object> payload = new LinkedHashMap<>();
            payload.put("blockId", en.getKey());
            payload.put("text", en.getValue().toString());
            payload.put("collapsed", true);
            Map<String, Object> block = new LinkedHashMap<>();
            block.put("type", "reasoning");
            block.put("payload", payload);
            blocks.add(block);
        }
        for (Map.Entry<String, StringBuilder> en : acc.textBlocksCopy().entrySet()) {
            Map<String, Object> payload = new LinkedHashMap<>();
            payload.put("blockId", en.getKey());
            payload.put("text", acc.cleanContent(en.getValue().toString()));
            Map<String, Object> block = new LinkedHashMap<>();
            block.put("type", "text");
            block.put("payload", payload);
            blocks.add(block);
        }
        if (blocks.isEmpty() && StringUtils.hasText(content)) {
            Map<String, Object> payload = new LinkedHashMap<>();
            payload.put("text", content);
            Map<String, Object> block = new LinkedHashMap<>();
            block.put("type", "text");
            block.put("payload", payload);
            blocks.add(block);
        }
        Map<String, Object> root = new LinkedHashMap<>();
        root.put("blocks", blocks);
        List<Map<String, Object>> datasets = new ArrayList<>();
        if (ctx != null && ctx.datasetSink() != null && !ctx.datasetSink().collected().isEmpty()) {
            for (DatasetSnapshot s : ctx.datasetSink().collected()) {
                datasets.add(toDatasetMap(s));
            }
        }
        for (DatasetSnapshot t : tables) {
            datasets.add(toTableDatasetMap(t));
        }
        if (!datasets.isEmpty()) {
            root.put("datasets", datasets);
        }
        if (!acc.chartHints.isEmpty()) {
            root.put("chartHints", acc.chartHints);
        }
        try {
            return objectMapper.writeValueAsString(root);
        } catch (Exception e) {
            log.warn("metadata 序列化失败，降级为 null: {}", e.getMessage());
            return null;
        }
    }

    /**
     * 下发一个 SSE 事件；客户端断开时静默跳过后续下发（不抛出，以便继续消费流完成落库）。
     */
    private void sendEvent(SseEmitter emitter, ClientGate gate, String name, Map<String, Object> data) {
        if (gate.isGone()) {
            return;
        }
        try {
            emitter.send(SseEmitter.event().name(name).data(data));
        } catch (IOException e) {
            gate.disconnect();
            log.debug("SSE 客户端已断开，停止下发: {}", e.getMessage());
        } catch (IllegalStateException e) {
            gate.disconnect();
        }
    }

    /**
     * 下发错误事件（替代 done）。
     */
    private void sendError(SseEmitter emitter, ClientGate gate, Exception e) {
        Map<String, Object> err = new LinkedHashMap<>();
        err.put("message", StringUtils.hasText(e.getMessage()) ? e.getMessage() : "助手处理失败");
        sendEvent(emitter, gate, "error", err);
    }

    /**
     * 安全完成 emitter（忽略已完成/已断开）。
     */
    private void safeComplete(SseEmitter emitter) {
        try {
            emitter.complete();
        } catch (Exception ignored) {
            // 已完成或客户端断开，忽略
        }
    }

    /**
     * 客户端连接闸门：断开后停止下发，但流继续消费以完成落库。
     */
    private static final class ClientGate {
        private volatile boolean gone;

        boolean isGone() {
            return gone;
        }

        void disconnect() {
            this.gone = true;
        }
    }

    /**
     * 事件累加器：按 blockId 聚合正文与思考增量，记录 usage 与最终结果文本。
     *
     * <p>事件在反应式线程上串行回调，末尾在异步线程读取；方法 {@code synchronized} 保证发布安全。
     */
    private static final class Accumulator {
        private final LinkedHashMap<String, StringBuilder> textBlocks = new LinkedHashMap<>();
        private final LinkedHashMap<String, StringBuilder> thinkingBlocks = new LinkedHashMap<>();
        private final LinkedHashMap<String, Object> chartHints = new LinkedHashMap<>();
        private volatile String resultText;
        private volatile int usageInput;
        private volatile int usageOutput;

        synchronized void appendText(String blockId, String delta) {
            if (delta == null) {
                return;
            }
            String id = blockId == null ? "default" : blockId;
            textBlocks.computeIfAbsent(id, k -> new StringBuilder()).append(delta);
        }

        synchronized void appendThinking(String blockId, String delta) {
            if (delta == null) {
                return;
            }
            String id = blockId == null ? "default" : blockId;
            thinkingBlocks.computeIfAbsent(id, k -> new StringBuilder()).append(delta);
        }

        synchronized String mergedText() {
            StringBuilder all = new StringBuilder();
            for (StringBuilder sb : textBlocks.values()) {
                if (all.length() > 0) {
                    all.append("\n\n");
                }
                all.append(sb);
            }
            return all.toString();
        }

        /** 剔除正文中的 chart_hint 标记行及其遗留空白。 */
        String cleanContent(String raw) {
            if (raw == null) {
                return null;
            }
            String cleaned = CHART_HINT.matcher(raw).replaceAll("");
            // 收拾标记被剔除后可能留下的多余空行
            cleaned = cleaned.replaceAll("[ 	]*\\n+[ 	]*\\n+[ 	]*", "\n\n").trim();
            return cleaned;
        }

        synchronized LinkedHashMap<String, StringBuilder> textBlocksCopy() {
            LinkedHashMap<String, StringBuilder> copy = new LinkedHashMap<>();
            for (Map.Entry<String, StringBuilder> en : textBlocks.entrySet()) {
                copy.put(en.getKey(), new StringBuilder(en.getValue()));
            }
            return copy;
        }
    }
}
