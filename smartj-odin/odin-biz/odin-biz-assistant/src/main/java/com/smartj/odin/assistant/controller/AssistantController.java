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
package com.smartj.odin.assistant.controller;

import com.smartj.common.core.domain.PageResult;
import com.smartj.common.core.domain.Result;
import com.smartj.common.log.annotation.OperationLog;
import com.smartj.odin.assistant.dto.ChatRequestDTO;
import com.smartj.odin.assistant.dto.ChatStreamRequestDTO;
import com.smartj.odin.assistant.dto.SessionCreateDTO;
import com.smartj.odin.assistant.dto.SessionQuery;
import com.smartj.odin.assistant.service.AssistantService;
import com.smartj.odin.assistant.stream.AssistantStreamService;
import com.smartj.odin.assistant.vo.AssistantMessageVO;
import com.smartj.odin.assistant.vo.AssistantSessionVO;
import com.smartj.odin.assistant.vo.AssistantSkillVO;
import com.smartj.odin.assistant.vo.ChatResponseVO;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.MediaType;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.util.List;

/**
 * 助手 Controller。
 *
 * <p>提供对话、技能列表与会话/消息管理接口。对话当前为一次性完整响应，
 * P3 接入 Agentscope-java 后改为 SSE 流式（{@code produces = text/event-stream}）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Tag(name = "AI 助手", description = "对话、技能列表与会话管理")
@RestController
@RequestMapping("/odin/assistant")
@RequiredArgsConstructor
public class AssistantController {

    private final AssistantService assistantService;
    private final AssistantStreamService streamService;

    /** 发送对话消息（P3 起改为 SSE 流式返回） */
    @Operation(summary = "发送对话消息")
    @PostMapping("/chat")
    @PreAuthorize("@perm.has('odin:assistant:chat')")
    @OperationLog(module = "AI助手", action = "对话")
    public Result<ChatResponseVO> chat(@Valid @RequestBody ChatRequestDTO dto) {
        return Result.ok(assistantService.chat(dto));
    }

    /**
     * 发送对话消息（SSE 流式，token 级）。
     *
     * <p>事件协议与前端 {@code useOdinSSE} / assistant store 对齐：{@code meta} 首帧、
     * {@code text_delta}/{@code text_end} 正文增量、{@code thinking_delta}/{@code thinking_end} 思考增量、
     * {@code tool_call} 工具占位、{@code block} 整块兜底、{@code draft} 草稿角标、{@code done}/{@code error} 收尾。
     * 具体订阅与事件翻译由 {@link AssistantStreamService} 在专用线程池完成（SecurityContext 跨线程传递）。
     */
    @Operation(summary = "发送对话消息（SSE 流式）")
    @PostMapping(value = "/chat/stream", produces = MediaType.TEXT_EVENT_STREAM_VALUE)
    @PreAuthorize("@perm.has('odin:assistant:chat')")
    public SseEmitter chatStream(@Valid @RequestBody ChatStreamRequestDTO dto) {
        return streamService.stream(dto);
    }

    /** 获取助手技能列表 */
    @Operation(summary = "获取助手技能列表")
    @GetMapping("/skills")
    @PreAuthorize("@perm.has('odin:assistant:skills')")
    public Result<List<AssistantSkillVO>> skills() {
        return Result.ok(assistantService.getSkills());
    }

    /** 分页查询会话列表 */
    @Operation(summary = "分页查询会话列表")
    @GetMapping("/session/page")
    @PreAuthorize("@perm.has('odin:assistant:session')")
    public Result<PageResult<AssistantSessionVO>> sessionPage(SessionQuery query) {
        return Result.ok(assistantService.sessionPage(query));
    }

    /** 查询会话列表（不分页，可按本体过滤） */
    @Operation(summary = "查询会话列表")
    @GetMapping("/sessions")
    @PreAuthorize("@perm.has('odin:assistant:session')")
    public Result<List<AssistantSessionVO>> sessions(
            @Parameter(name = "ontologyId", description = "本体ID，不传则返回全部")
            @RequestParam(required = false) Long ontologyId) {
        SessionQuery query = new SessionQuery();
        query.setCurrent(1);
        query.setSize(9999);
        query.setOntologyId(ontologyId);
        return Result.ok(assistantService.sessionPage(query).getRecords());
    }

    /** 查询会话消息列表 */
    @Operation(summary = "查询会话消息列表")
    @GetMapping("/session/{sid}/messages")
    @PreAuthorize("@perm.has('odin:assistant:session')")
    public Result<List<AssistantMessageVO>> messages(@Parameter(name = "sid", description = "会话ID", required = true) @PathVariable Long sid) {
        return Result.ok(assistantService.getMessages(sid));
    }

    /** 创建会话，返回新会话ID */
    @Operation(summary = "创建会话")
    @PostMapping("/session")
    @PreAuthorize("@perm.has('odin:assistant:session')")
    @OperationLog(module = "AI助手", action = "创建会话")
    public Result<Long> createSession(@RequestBody SessionCreateDTO dto) {
        return Result.ok(assistantService.createSession(dto));
    }

    /** 删除会话（级联删除消息） */
    @Operation(summary = "删除会话")
    @DeleteMapping("/session/{sid}")
    @PreAuthorize("@perm.has('odin:assistant:session')")
    @OperationLog(module = "AI助手", action = "删除会话")
    public Result<Void> deleteSession(@Parameter(name = "sid", description = "会话ID", required = true) @PathVariable Long sid) {
        assistantService.deleteSession(sid);
        return Result.ok();
    }
}
