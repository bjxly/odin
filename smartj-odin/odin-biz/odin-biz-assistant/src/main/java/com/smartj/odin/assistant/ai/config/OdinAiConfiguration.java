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

import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.scheduling.concurrent.CustomizableThreadFactory;

import java.util.concurrent.ExecutorService;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.ThreadPoolExecutor;
import java.util.concurrent.TimeUnit;

/**
 * ODIN AI 助手配置装配。
 *
 * <p>开启 {@link AgentScopeProperties}（{@code odin.ai.*}）的配置绑定。
 * 模型与 Agent 采用惰性/按请求构建，故此处不声明重量级 Bean，避免未配置密钥时影响启动。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Configuration
@EnableConfigurationProperties(AgentScopeProperties.class)
public class OdinAiConfiguration {

    /**
     * SSE 流式对话专用有界线程池。
     *
     * <p>因 {@code AssistantAiClient.chatStream} 内部 {@code blockLast()} 会阻塞至流终止，每个进行中的
     * 流式会话占用一个工作线程；以核心 4 / 最大 16 / 队列 100 限定并发，命名 {@code odin-sse-} 便于排障。
     * 饱和时由提交线程（Tomcat 请求线程）直跑（CallerRuns），避免无界堆积。
     *
     * @return 流式执行器
     */
    @Bean(name = "assistantStreamExecutor", destroyMethod = "shutdown")
    public ExecutorService assistantStreamExecutor() {
        return new ThreadPoolExecutor(
                4, 16, 60L, TimeUnit.SECONDS,
                new LinkedBlockingQueue<>(100),
                new CustomizableThreadFactory("odin-sse-"),
                new ThreadPoolExecutor.CallerRunsPolicy());
    }
}
