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

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.function.Consumer;

/**
 * 数据集侧信道：问数工具产出快照的接收与转发枢纽。
 *
 * <p>工具执行线程（反应式线程）调用 {@link #publish} 投递快照；{@code handler} 由流式编排线程在
 * 订阅前设置，用于即时下发 SSE {@code dataset} 事件。无论 handler 是否就绪，快照都会累积到
 * {@link #collected()}，供流结束时写入 {@code metadata} 做历史还原。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
public class DatasetSink {

    /** dataId 前缀：{@code ds-{traceId前8}-} */
    private final String prefix;

    /** 本轮序号 */
    private final AtomicInteger seq = new AtomicInteger();

    /** 已收集快照（线程安全） */
    private final List<DatasetSnapshot> collected = Collections.synchronizedList(new ArrayList<>());

    /** 即时转发回调（可空，异步线程在流开始前设置） */
    private volatile Consumer<DatasetSnapshot> handler;

    public DatasetSink(String traceId) {
        String head = traceId == null ? "anon" : (traceId.length() >= 8 ? traceId.substring(0, 8) : traceId);
        this.prefix = "ds-" + head;
    }

    /** 生成下一个数据集ID。 */
    public String nextDataId() {
        return prefix + "-" + seq.incrementAndGet();
    }

    /** 设置即时转发回调。 */
    public void setHandler(Consumer<DatasetSnapshot> handler) {
        this.handler = handler;
    }

    /** 投递一个快照：先累积，再触发回调（若已就绪）。 */
    public void publish(DatasetSnapshot snapshot) {
        if (snapshot == null) {
            return;
        }
        collected.add(snapshot);
        Consumer<DatasetSnapshot> h = this.handler;
        if (h != null) {
            h.accept(snapshot);
        }
    }

    /** 返回本轮已收集的全部快照副本。 */
    public List<DatasetSnapshot> collected() {
        synchronized (collected) {
            return new ArrayList<>(collected);
        }
    }
}
