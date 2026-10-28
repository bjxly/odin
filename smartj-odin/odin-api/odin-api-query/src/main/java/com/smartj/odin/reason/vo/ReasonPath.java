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
package com.smartj.odin.reason.vo;

import lombok.Data;

import java.io.Serial;
import java.io.Serializable;
import java.util.ArrayList;
import java.util.List;

/**
 * 推理路径 VO。
 *
 * <p>与前端 {@code src/types/odin/reason.ts} 的 ReasonPath 契约一致，
 * 描述源类到目标类的推理路径节点、跳数与途经关系。
 * TODO: 契约对齐桩类型，真实路径推导引擎待 P3 接入。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class ReasonPath implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 源类编码 */
    private String sourceClass;
    /** 目标类编码 */
    private String targetClass;
    /** 路径节点 */
    private List<PathNode> nodes = new ArrayList<>();
    /** 跳数 */
    private Integer hops;
    /** 途经关系 */
    private List<String> relations = new ArrayList<>();

    /** 推理路径节点 */
    @Data
    public static class PathNode implements Serializable {
        @Serial
        private static final long serialVersionUID = 1L;
        /** 类编码 */
        private String classCode;
        /** 类名称 */
        private String className;
        /** 深度 */
        private Integer depth;
        /** 到达该节点的原因 */
        private String reason;
    }
}
