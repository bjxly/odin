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

import com.smartj.common.core.domain.PageResult;
import com.smartj.odin.assistant.dto.ChatRequestDTO;
import com.smartj.odin.assistant.dto.SessionCreateDTO;
import com.smartj.odin.assistant.dto.SessionQuery;
import com.smartj.odin.assistant.vo.AssistantMessageVO;
import com.smartj.odin.assistant.vo.AssistantSessionVO;
import com.smartj.odin.assistant.vo.AssistantSkillVO;
import com.smartj.odin.assistant.vo.ChatResponseVO;

import java.util.List;

/**
 * 助手服务接口。
 *
 * <p>覆盖对话、技能列表与会话/消息管理。对话当前为持久化骨架（保存用户消息并返回占位助手消息），
 * Agentscope-java HarnessAgent 在 P3 接入后改为 SSE 流式。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public interface AssistantService {

    /**
     * 处理一次对话：创建/获取会话，保存用户消息，返回助手回复。
     *
     * @param dto 对话请求
     * @return 对话响应
     */
    ChatResponseVO chat(ChatRequestDTO dto);

    /**
     * 为流式对话做准备（在请求线程内、携带 SecurityContext 与事务执行）：
     * 获取/创建会话、落库用户消息、加载本体词表、构建系统提示词与工具集、拼装多轮上下文。
     *
     * <p>把需要登录态的写操作全部前置到此处，流式阶段（异步线程）不再持有事务，仅末尾回写助手消息。
     *
     * @param dto 对话请求
     * @return 流式上下文（含会话、提示词、工具集、上下文轮次；AI 未启用时携带占位回复）
     */
    ChatContext prepareChat(ChatRequestDTO dto);

    /**
     * 流式结束后回写助手消息（在异步线程、由调用方传播 SecurityContext 后于事务内执行）。
     *
     * @param ctx         {@link #prepareChat} 产出的上下文
     * @param content     聚合后的助手回复全文
     * @param metadataJson 块结构 JSON（用于历史还原），可空
     * @param latencyMs   生成耗时（毫秒）
     * @param tokensUsed  消耗 token 数，可空
     * @return 新插入的助手消息ID
     */
    Long persistAssistantReply(ChatContext ctx, String content, String metadataJson, int latencyMs, Integer tokensUsed);

    /**
     * 获取助手可编排的技能列表。
     *
     * @return 技能 VO 列表
     */
    List<AssistantSkillVO> getSkills();

    /**
     * 分页查询会话。
     *
     * @param query 查询条件
     * @return 分页结果
     */
    PageResult<AssistantSessionVO> sessionPage(SessionQuery query);

    /**
     * 查询指定会话的消息列表（按时间升序）。
     *
     * @param sessionId 会话ID
     * @return 消息 VO 列表
     */
    List<AssistantMessageVO> getMessages(Long sessionId);

    /**
     * 创建会话，返回主键ID。
     *
     * @param dto 会话创建入参
     * @return 新会话ID
     */
    Long createSession(SessionCreateDTO dto);

    /**
     * 删除会话（级联删除其消息）。
     *
     * @param sessionId 会话ID
     */
    void deleteSession(Long sessionId);
}
