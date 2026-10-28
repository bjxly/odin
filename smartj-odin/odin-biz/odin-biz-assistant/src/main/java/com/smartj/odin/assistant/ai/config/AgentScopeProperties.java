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
package com.smartj.odin.assistant.ai.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;

/**
 * ODIN AI 助手运行参数（配置前缀 {@code odin.ai}）。
 *
 * <p>驱动 Agentscope-Java {@code ReActAgent} + OpenAI 兼容模型（默认可直连 DeepSeek）。
 * {@code enabled=false} 或缺少 api-key/base-url/model 时，助手回退为占位回复，不影响应用启动。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
@ConfigurationProperties(prefix = "odin.ai")
public class AgentScopeProperties {

    /** 是否启用真实 AI 对话（false 时返回占位回复） */
    private boolean enabled = false;

    /** 模型提供方：deepseek / openai（决定消息 Formatter） */
    private String provider = "deepseek";

    /** OpenAI 兼容端点 BaseURL，如 https://api.deepseek.com/v1 */
    private String baseUrl = "https://api.deepseek.com/v1";

    /** API Key（建议用环境变量注入，勿硬编码入库） */
    private String apiKey = "";

    /** 模型名，如 deepseek-chat */
    private String model = "deepseek-chat";

    /** 采样温度（0=确定性解码） */
    private Double temperature = 0.0;

    /** ReAct 最大迭代轮数 */
    private Integer maxIters = 8;

    /** 注入模型的历史消息最大条数（多轮上下文） */
    private Integer maxHistory = 20;

    /** 单次对话阻塞超时（秒） */
    private Integer timeoutSeconds = 120;

    /**
     * 是否启用 token 级流式（{@code streamEvents} 路径）。
     * {@code true} 时 {@code /chat/stream} 逐 token 增量下发；{@code false} 回退为一次性全文单块下发。
     */
    private boolean stream = true;

    /** 系统提示词（定义助手角色与边界） */
    private String systemPrompt = """
            你是 ODIN（本体驱动的数据洞察系统）的 AI 助手。
            你的职责：帮助用户理解本体（类/属性/关系）、数据源与映射、智能问数与业务洞察的用法与概念。
            约束：
            1. 只依据用户提供的信息与通用数据知识作答，不臆造具体的业务数据、表结构或查询结果；
            2. 涉及具体数据查询、推理试运行等操作时，说明应前往对应功能页面执行，不伪造执行结果；
            3. 回答使用简体中文，条理清晰、简洁专业。
            """;
}
