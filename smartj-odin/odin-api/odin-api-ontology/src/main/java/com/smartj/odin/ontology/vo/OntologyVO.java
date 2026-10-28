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
import java.util.List;

/**
 * 本体定义 VO（展示）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class OntologyVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID */
    private String id;
    /** 本体编码 */
    private String code;
    /** 本体名称 */
    private String name;
    /** 描述 */
    private String description;
    /** 版本号 */
    private String version;
    /** 助手预置问题列表 */
    private List<String> presetQuestions;
    /** 状态 0=禁用 1=启用 */
    private Integer status;
    /** 备注 */
    private String remark;
    /** 创建时间 */
    private String createTime;
    /** 更新时间 */
    private String updateTime;
    /** 类数量 */
    private Integer classCount;
    /** 关系数量 */
    private Integer relationCount;
}
