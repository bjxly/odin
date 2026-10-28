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
package com.smartj.odin.ontology.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;
import lombok.Data;

import java.io.Serial;
import java.io.Serializable;
import java.util.List;

/**
 * 本体定义 DTO（新增/更新入参）。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Data
public class OntologyDTO implements Serializable {

    @Serial
    private static final long serialVersionUID = 1L;

    /** 主键ID（更新时必填） */
    private Long id;

    /** 本体编码 */
    @NotBlank(message = "本体编码不能为空")
    @Size(max = 64, message = "本体编码长度不能超过64")
    private String code;

    /** 本体名称 */
    @NotBlank(message = "本体名称不能为空")
    @Size(max = 128, message = "本体名称长度不能超过128")
    private String name;

    /** 描述 */
    @Size(max = 2000, message = "描述长度不能超过2000")
    private String description;

    /** 版本号 */
    @Size(max = 32, message = "版本号长度不能超过32")
    private String version;

    /** 助手预置问题列表 */
    private List<String> presetQuestions;

    /** 状态 0=禁用 1=启用 */
    private Integer status;

    /** 备注 */
    @Size(max = 500, message = "备注长度不能超过500")
    private String remark;
}
