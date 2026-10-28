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
package com.smartj.odin.ontology.vo;

import lombok.Data;

import java.io.Serial;
import java.io.Serializable;

/**
 * 本体变更日志 VO（展示）。
 *
 * <p>字段与前端 {@code smartj-odin-ui/src/types/odin/trace.ts#ChangeLogVO} 一一对应。
 * 当前 odin-biz-ontology 未依赖 odin-biz-insight，故本体维度的变更历史以桩形式返回，
 * 待后续将 ChangeHistory 通用能力下沉到 odin-common 后再切换为真实数据源。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class ChangeLogVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 变更历史ID */
    private String id;
    /** 本体ID */
    private String ontologyId;
    /** 实体类型：ontology/class/property/relation/rule/action/mapping */
    private String entityType;
    /** 实体ID */
    private String entityId;
    /** 操作：create/update/delete/publish/revert */
    private String action;
    /** 变更前快照（JSON 字符串） */
    private String beforeSnapshot;
    /** 变更后快照（JSON 字符串） */
    private String afterSnapshot;
    /** 操作人账号 */
    private String operator;
    /** 操作人姓名 */
    private String operatorName;
    /** 变更原因 */
    private String reason;
    /** 创建时间 */
    private String createTime;
}
