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
package com.smartj.odin.assistant.ai;

import com.smartj.odin.assistant.ai.config.AgentScopeProperties;
import io.agentscope.core.ReActAgent;
import io.agentscope.core.event.AgentEvent;
import io.agentscope.core.message.Msg;
import io.agentscope.core.message.MsgRole;
import io.agentscope.core.model.GenerateOptions;
import io.agentscope.core.tool.Toolkit;
import io.agentscope.extensions.model.openai.OpenAIChatModel;
import io.agentscope.extensions.model.openai.compat.deepseek.DeepSeekFormatter;
import io.agentscope.extensions.model.openai.formatter.OpenAIChatFormatter;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;

import java.time.Duration;
import java.util.ArrayList;
import java.util.List;
import java.util.function.Consumer;

/**
 * AI 助手对话客户端（Agentscope-Java）。
 *
 * <p>基于 OpenAI 兼容 {@link OpenAIChatModel}（默认 DeepSeek）与 {@link ReActAgent} 完成一次对话。
 * {@link ReActAgent} 非线程安全，故每次请求经工厂方法新建一个实例；{@link OpenAIChatModel} 与
 * {@link Toolkit} 可跨请求共享（Toolkit 在 agent 构建时深拷贝）。
 *
 * <p>未启用（{@code odin.ai.enabled=false} 或缺少密钥）时 {@link #isEnabled()} 返回 false，
 * 由调用方回退为占位回复。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Slf4j
@Component
@RequiredArgsConstructor
public class AssistantAiClient {

    private final AgentScopeProperties props;

    /** 惰性构建、跨请求共享的模型实例——非流式（{@code stream(false)}），服务同步 {@code /chat}。 */
    private volatile OpenAIChatModel modelNonStream;

    /** 惰性构建、跨请求共享的模型实例——流式（{@code stream(true)}），服务 {@code streamEvents} token 级增量。 */
    private volatile OpenAIChatModel modelStream;

    /**
     * 一轮对话消息（角色 + 文本）。role 取 {@code user} / {@code assistant}。
     *
     * @param role    角色
     * @param content 文本内容
     */
    public record HistoryTurn(String role, String content) {
    }

    /** AI 对话是否可用（已启用且 base-url/api-key/model 均非空）。 */
    public boolean isEnabled() {
        return props.isEnabled()
                && StringUtils.hasText(props.getApiKey())
                && StringUtils.hasText(props.getBaseUrl())
                && StringUtils.hasText(props.getModel());
    }

    /**
     * 执行一次对话，返回助手回复文本。
     *
     * @param conversation 有序对话（含历史与当前用户问题，最后一条须为用户消息）
     * @param systemPrompt 系统提示词（可空，空则用配置默认）；用于注入本体词表等上下文
     * @param toolkit      工具集（可空，空则无工具）；用于注册本体问数等 @Tool
     * @return 助手回复文本；无有效回复时返回 {@code null}
     */
    public String chat(List<HistoryTurn> conversation, String systemPrompt, Toolkit toolkit) {
        OpenAIChatModel chatModel = model(false);
        int maxIters = props.getMaxIters() == null ? 8 : props.getMaxIters();
        int timeoutSeconds = props.getTimeoutSeconds() == null ? 120 : props.getTimeoutSeconds();

        String sysPrompt = StringUtils.hasText(systemPrompt) ? systemPrompt : props.getSystemPrompt();
        Toolkit effectiveToolkit = toolkit != null ? toolkit : new Toolkit();

        ReActAgent agent = ReActAgent.builder()
                .name("OdinAssistant")
                .sysPrompt(sysPrompt)
                .model(chatModel)
                .toolkit(effectiveToolkit)
                .maxIters(maxIters)
                .build();
        try {
            List<Msg> msgs = toMsgs(conversation);
            Msg reply = agent.call(msgs).block(Duration.ofSeconds(timeoutSeconds));
            String text = reply == null ? null : reply.getTextContent();
            return StringUtils.hasText(text) ? text : null;
        } finally {
            closeQuietly(agent);
        }
    }

    /**
     * 以 token 级流式执行一次对话，将 Agentscope 增量事件逐个回调给 {@code onEvent}。
     *
     * <p>复用与 {@link #chat} 相同的 {@link ReActAgent} 内核（{@code streamEvents} 与 {@code call}
     * 共用 buildAgentStream），但使用 {@code stream(true)} 模型实例并逐事件下发。本方法阻塞直到
     * 流终止（完成/超时/异常）；{@code onEvent} 在反应式发射线程上被回调，调用方需保证线程安全。
     *
     * @param conversation 有序对话（含历史与当前用户问题，末条须为用户消息）
     * @param systemPrompt 系统提示词（可空，空则用配置默认）
     * @param toolkit      工具集（可空，空则无工具）
     * @param onEvent      事件回调（每个 {@link AgentEvent} 触发一次）
     * @param timeout      整体超时（可空，空则用配置 {@code timeoutSeconds}）
     */
    public void chatStream(List<HistoryTurn> conversation, String systemPrompt, Toolkit toolkit,
                           Consumer<AgentEvent> onEvent, Duration timeout) {
        OpenAIChatModel chatModel = model(true);
        int maxIters = props.getMaxIters() == null ? 8 : props.getMaxIters();
        int timeoutSeconds = props.getTimeoutSeconds() == null ? 120 : props.getTimeoutSeconds();
        Duration effectiveTimeout = timeout != null ? timeout : Duration.ofSeconds(timeoutSeconds);

        String sysPrompt = StringUtils.hasText(systemPrompt) ? systemPrompt : props.getSystemPrompt();
        Toolkit effectiveToolkit = toolkit != null ? toolkit : new Toolkit();

        ReActAgent agent = ReActAgent.builder()
                .name("OdinAssistant")
                .sysPrompt(sysPrompt)
                .model(chatModel)
                .toolkit(effectiveToolkit)
                .maxIters(maxIters)
                .build();
        try {
            List<Msg> msgs = toMsgs(conversation);
            agent.streamEvents(msgs)
                    .timeout(effectiveTimeout)
                    .doOnNext(onEvent::accept)
                    .blockLast();
        } finally {
            closeQuietly(agent);
        }
    }

    /**
     * 惰性构建并缓存 OpenAI 兼容模型（DeepSeek 用 {@link DeepSeekFormatter}，其余用标准 Formatter）。
     *
     * <p>{@link OpenAIChatModel} 实例的 stream 模式不可混用，故按 {@code stream} 值分别缓存两个实例：
     * 非流式服务同步 {@code /chat}，流式服务 {@code /chat/stream} 的 token 级增量。
     *
     * @param stream 是否需要流式模型实例
     * @return 模型实例
     */
    private OpenAIChatModel model(boolean stream) {
        if (stream) {
            OpenAIChatModel local = this.modelStream;
            if (local == null) {
                synchronized (this) {
                    local = this.modelStream;
                    if (local == null) {
                        local = buildModel(true);
                        this.modelStream = local;
                    }
                }
            }
            return local;
        }
        OpenAIChatModel local = this.modelNonStream;
        if (local == null) {
            synchronized (this) {
                local = this.modelNonStream;
                if (local == null) {
                    local = buildModel(false);
                    this.modelNonStream = local;
                }
            }
        }
        return local;
    }

    /**
     * 按 stream 模式构建模型实例并记录初始化日志。
     *
     * @param stream 是否流式
     * @return 模型实例
     */
    private OpenAIChatModel buildModel(boolean stream) {
        OpenAIChatFormatter formatter = "openai".equalsIgnoreCase(props.getProvider())
                ? new OpenAIChatFormatter()
                : new DeepSeekFormatter();
        GenerateOptions options = GenerateOptions.builder()
                .temperature(props.getTemperature())
                .build();
        OpenAIChatModel built = OpenAIChatModel.builder()
                .apiKey(props.getApiKey())
                .baseUrl(props.getBaseUrl())
                .modelName(props.getModel())
                .stream(stream)
                .formatter(formatter)
                .nativeStructuredOutputWithTools(false)
                .generateOptions(options)
                .build();
        log.info("ODIN AI 模型已初始化: provider={}, model={}, baseUrl={}, stream={}",
                props.getProvider(), props.getModel(), props.getBaseUrl(), stream);
        return built;
    }

    /**
     * 将对话轮次转换为 Agentscope {@link Msg} 列表（跳过空内容）。
     *
     * @param conversation 对话轮次
     * @return Msg 列表
     */
    private List<Msg> toMsgs(List<HistoryTurn> conversation) {
        List<Msg> msgs = new ArrayList<>();
        if (conversation == null) {
            return msgs;
        }
        for (HistoryTurn turn : conversation) {
            if (turn == null || !StringUtils.hasText(turn.content())) {
                continue;
            }
            boolean assistant = "assistant".equalsIgnoreCase(turn.role());
            MsgRole role = assistant ? MsgRole.ASSISTANT : MsgRole.USER;
            msgs.add(Msg.builder()
                    .name(assistant ? "assistant" : "user")
                    .role(role)
                    .textContent(turn.content())
                    .build());
        }
        return msgs;
    }

    /**
     * 静默关闭 Agent，释放其占用资源。
     *
     * @param agent 待关闭的 Agent
     */
    private void closeQuietly(ReActAgent agent) {
        try {
            agent.close();
        } catch (Exception e) {
            log.debug("关闭 ReActAgent 失败: {}", e.getMessage());
        }
    }
}
