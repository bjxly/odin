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
 * 本体类 VO（展示，支持树形结构）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class OntClassVO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID */
    private String id;
    /** 所属本体ID */
    private String ontologyId;
    /** 类编码 */
    private String code;
    /** 类名称 */
    private String name;
    /** 描述 */
    private String description;
    /** 父类ID */
    private String parentClassId;
    /** 图标 */
    private String icon;
    /** 排序号 */
    private Integer sortOrder;
    /** 身份键（JSON） */
    private String identityKeys;
    /** 状态 0=禁用 1=启用 */
    private Integer status;
    /** 备注 */
    private String remark;
    /** 创建时间 */
    private String createTime;
    /** 子类列表 */
    private List<OntClassVO> children;
    /** 属性数量 */
    private Integer propertyCount;
}
