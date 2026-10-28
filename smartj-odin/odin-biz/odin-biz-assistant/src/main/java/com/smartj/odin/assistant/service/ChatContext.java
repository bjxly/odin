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
package com.smartj.odin.assistant.service;

import com.smartj.odin.assistant.ai.AssistantAiClient;
import com.smartj.odin.assistant.ai.tool.DatasetSink;
import com.smartj.odin.assistant.ai.tool.OntologyMutateTool;
import io.agentscope.core.tool.Toolkit;

import java.util.List;

/**
 * 流式对话上下文：由 {@link AssistantService#prepareChat} 在请求线程内构建，交给异步流式阶段消费。
 *
 * <p>把全部依赖登录态的准备工作（会话、用户消息落库、词表、提示词、工具集、上下文）前置完成，
 * 使流式阶段与末尾回写只需持有这些显式数据；{@code aiEnabled=false} 时 {@code fallbackContent}
 * 携带占位回复，由流式层回退为单文本块下发。
 *
 * @param traceId         本轮追踪ID
 * @param sessionId       会话ID
 * @param userId          当前登录用户ID（用于异步线程审计与断言日志）
 * @param ontologyId      当前本体ID（可空）
 * @param sysPrompt       系统提示词（可空）
 * @param toolkit         工具集（可空）
 * @param conversation    注入模型的多轮上下文（末条为当前用户消息）
 * @param mutateTool      本轮写工具实例（用于流结束后收集草稿ID，可空）
 * @param datasetSink     数据集侧信道（问数工具产出快照的接收/转发，可空）
 * @param aiEnabled       AI 是否可用
 * @param fallbackContent AI 不可用时的占位回复（可用时为 null）
 * @author SmartG-Odin
 * @since 1.0.0
 */
public record ChatContext(
        String traceId,
        Long sessionId,
        Long userId,
        Long ontologyId,
        String sysPrompt,
        Toolkit toolkit,
        List<AssistantAiClient.HistoryTurn> conversation,
        OntologyMutateTool mutateTool,
        DatasetSink datasetSink,
        boolean aiEnabled,
        String fallbackContent
) {
}
