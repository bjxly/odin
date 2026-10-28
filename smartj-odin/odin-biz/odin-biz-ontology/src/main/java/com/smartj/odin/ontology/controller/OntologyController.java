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
package com.smartj.odin.ontology.controller;

import com.smartj.common.core.domain.PageResult;
import com.smartj.common.core.domain.Result;
import com.smartj.common.log.annotation.OperationLog;
import com.smartj.odin.ontology.dto.ChangeLogQuery;
import com.smartj.odin.ontology.dto.OntActionDTO;
import com.smartj.odin.ontology.dto.OntClassDTO;
import com.smartj.odin.ontology.dto.OntPropertyDTO;
import com.smartj.odin.ontology.dto.OntRelationDTO;
import com.smartj.odin.ontology.dto.OntRuleDTO;
import com.smartj.odin.ontology.dto.OntologyDTO;
import com.smartj.odin.ontology.dto.OntologyQuery;
import com.smartj.odin.ontology.service.OntologyService;
import com.smartj.odin.ontology.vo.ChangeLogVO;
import com.smartj.odin.ontology.vo.OntActionVO;
import com.smartj.odin.ontology.vo.OntClassVO;
import com.smartj.odin.ontology.vo.OntPropertyVO;
import com.smartj.odin.ontology.vo.OntRelationVO;
import com.smartj.odin.ontology.vo.OntRuleVO;
import com.smartj.odin.ontology.vo.OntologyGraphVO;
import com.smartj.odin.ontology.vo.OntologyVO;
import com.smartj.odin.ontology.vo.VocabularyVO;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;
import java.util.Map;

/**
 * 本体管理 Controller。
 *
 * <p>提供本体定义、类、属性、关系、规则、动作的完整 CRUD 接口，
 * 以及图谱与词表查询。所有接口受权限控制，写操作通过 @OperationLog 记录审计日志。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Tag(name = "本体管理", description = "本体定义、类、属性、关系、规则、动作 CRUD 及图谱/词表")
@RestController
@RequestMapping("/odin/ontology")
@RequiredArgsConstructor
public class OntologyController {

    private final OntologyService ontologyService;

    // ==================== 本体定义 ====================

    /** 分页查询本体列表 */
    @Operation(summary = "分页查询本体列表")
    @GetMapping("/page")
    @PreAuthorize("@perm.has('odin:ontology:list')")
    public Result<PageResult<OntologyVO>> page(OntologyQuery query) {
        return Result.ok(ontologyService.page(query));
    }

    /** 查询全部启用本体（不分页，用于下拉选择） */
    @Operation(summary = "查询全部启用本体列表")
    @GetMapping("/list")
    @PreAuthorize("@perm.has('odin:ontology:list')")
    public Result<List<OntologyVO>> list() {
        return Result.ok(ontologyService.listAll());
    }

    /** 查询本体摘要列表（轻量信息，用于工作台切换器） */
    @Operation(summary = "查询本体摘要列表")
    @GetMapping("/summaries")
    @PreAuthorize("@perm.has('odin:ontology:list')")
    public Result<List<OntologyVO>> summaries() {
        return Result.ok(ontologyService.listAll());
    }

    /**
     * 分页查询本体变更日志。
     *
     * <p>必须声明在 {@code /{id}} 之前，否则 Spring MVC 会把字面量 "changes" 当作 id
     * 匹配到 {@link #getById(Long)}，触发 Long 类型转换失败并返回 400。
     *
     * <p>odin-biz-ontology 未依赖 odin-biz-insight，无法直接复用 ChangeHistoryService，
     * 故此处以空分页桩返回；待变更历史通用能力下沉到 odin-common 后再接入真实数据源。
     */
    @Operation(summary = "分页查询本体变更日志")
    @GetMapping("/changes")
    @PreAuthorize("@perm.has('odin:ontology:list')")
    public Result<PageResult<ChangeLogVO>> changes(ChangeLogQuery query) {
        return Result.ok(PageResult.empty(query.getCurrent(), query.getSize()));
    }

    /** 根据ID查询本体详情 */
    @Operation(summary = "根据ID查询本体详情")
    @GetMapping("/{id}")
    @PreAuthorize("@perm.has('odin:ontology:query')")
    public Result<OntologyVO> getById(@Parameter(name = "id", description = "主键ID", required = true) @PathVariable Long id) {
        return Result.ok(ontologyService.getById(id));
    }

    /** 新增本体 */
    @Operation(summary = "新增本体")
    @PostMapping
    @PreAuthorize("@perm.has('odin:ontology:create')")
    @OperationLog(module = "本体管理", action = "创建本体")
    public Result<Long> create(@Valid @RequestBody OntologyDTO dto) {
        return Result.ok(ontologyService.create(dto));
    }

    /** 更新本体 */
    @Operation(summary = "更新本体")
    @PutMapping
    @PreAuthorize("@perm.has('odin:ontology:update')")
    @OperationLog(module = "本体管理", action = "更新本体")
    public Result<Void> update(@Valid @RequestBody OntologyDTO dto) {
        ontologyService.update(dto);
        return Result.ok();
    }

    /** 删除本体 */
    @Operation(summary = "删除本体")
    @DeleteMapping("/{id}")
    @PreAuthorize("@perm.has('odin:ontology:delete')")
    @OperationLog(module = "本体管理", action = "删除本体")
    public Result<Void> delete(@Parameter(name = "id", description = "主键ID", required = true) @PathVariable Long id) {
        ontologyService.delete(id);
        return Result.ok();
    }

    /** 获取本体图谱 */
    @Operation(summary = "获取本体图谱")
    @GetMapping("/{id}/graph")
    @PreAuthorize("@perm.has('odin:ontology:query')")
    public Result<OntologyGraphVO> graph(@Parameter(name = "id", description = "本体ID", required = true) @PathVariable Long id) {
        return Result.ok(ontologyService.getGraph(id));
    }

    /** 获取本体词表 */
    @Operation(summary = "获取本体词表")
    @GetMapping("/{id}/vocabulary")
    @PreAuthorize("@perm.has('odin:ontology:query')")
    public Result<VocabularyVO> vocabulary(@Parameter(name = "id", description = "本体ID", required = true) @PathVariable Long id) {
        return Result.ok(ontologyService.getVocabulary(id));
    }

    /** 发布本体版本 */
    @Operation(summary = "发布本体")
    @PostMapping("/{id}/publish")
    @PreAuthorize("@perm.has('odin:ontology:update')")
    @OperationLog(module = "本体管理", action = "发布本体")
    public Result<Void> publish(@Parameter(name = "id", description = "本体ID", required = true) @PathVariable Long id) {
        ontologyService.publish(id);
        return Result.ok();
    }

    /** 导出本体（返回可序列化结构，前端自行 JSON 化预览/下载） */
    @Operation(summary = "导出本体")
    @GetMapping("/{id}/export")
    @PreAuthorize("@perm.has('odin:ontology:query')")
    public Result<Map<String, Object>> export(@Parameter(name = "id", description = "本体ID", required = true) @PathVariable Long id) {
        return Result.ok(ontologyService.exportOntology(id));
    }

    /**
     * 保存本体图谱布局坐标。
     *
     * <p>当前 ont_class 无坐标持久化列，且前端图谱坐标由客户端 BFS 布局算法实时重算，
     * 故此端点为兼容契约的空操作桩，仅返回成功；待引入布局持久化列后再接入真实写入。
     */
    @Operation(summary = "保存本体布局坐标")
    @PutMapping("/{id}/layout")
    @PreAuthorize("@perm.has('odin:ontology:update')")
    public Result<Void> saveLayout(@Parameter(name = "id", description = "本体ID", required = true) @PathVariable Long id,
                                   @RequestBody(required = false) Map<String, Object> layout) {
        return Result.ok();
    }

    /**
     * 导入本体。
     *
     * <p>当前为兼容契约的空操作桩，仅校验本体存在并返回成功；
     * 完整的结构校验与增量合并导入待后续迭代实现。
     */
    @Operation(summary = "导入本体")
    @PostMapping("/{id}/import")
    @PreAuthorize("@perm.has('odin:ontology:update')")
    @OperationLog(module = "本体管理", action = "导入本体")
    public Result<Void> importOntology(@Parameter(name = "id", description = "本体ID", required = true) @PathVariable Long id,
                                       @RequestBody(required = false) Map<String, Object> data) {
        ontologyService.getById(id);
        return Result.ok();
    }

    // ==================== 类 ====================

    /** 查询本体下所有类（树形） */
    @Operation(summary = "查询本体类列表（树形）")
    @GetMapping("/{id}/classes")
    @PreAuthorize("@perm.has('odin:ontology:query')")
    public Result<List<OntClassVO>> listClasses(@Parameter(name = "id", description = "本体ID", required = true) @PathVariable Long id) {
        return Result.ok(ontologyService.listClasses(id));
    }

    /** 根据ID查询类详情 */
    @Operation(summary = "根据ID查询本体类")
    @GetMapping("/class/{id}")
    @PreAuthorize("@perm.has('odin:ontology:query')")
    public Result<OntClassVO> getClassById(@Parameter(name = "id", description = "类ID", required = true) @PathVariable Long id) {
        return Result.ok(ontologyService.getClassById(id));
    }

    /** 新增类（ontologyId 由请求体携带） */
    @Operation(summary = "新增本体类")
    @PostMapping("/class")
    @PreAuthorize("@perm.has('odin:ontology:create')")
    @OperationLog(module = "本体管理", action = "创建类")
    public Result<Long> createClass(@Valid @RequestBody OntClassDTO dto) {
        return Result.ok(ontologyService.createClass(dto));
    }

    /** 更新类（id 由请求体携带） */
    @Operation(summary = "更新本体类")
    @PutMapping("/class")
    @PreAuthorize("@perm.has('odin:ontology:update')")
    @OperationLog(module = "本体管理", action = "更新类")
    public Result<Void> updateClass(@Valid @RequestBody OntClassDTO dto) {
        ontologyService.updateClass(dto);
        return Result.ok();
    }

    /** 删除类 */
    @Operation(summary = "删除本体类")
    @DeleteMapping("/class/{id}")
    @PreAuthorize("@perm.has('odin:ontology:delete')")
    @OperationLog(module = "本体管理", action = "删除类")
    public Result<Void> deleteClass(@Parameter(name = "id", description = "类ID", required = true) @PathVariable Long id) {
        ontologyService.deleteClass(id);
        return Result.ok();
    }

    // ==================== 属性 ====================

    /** 查询类的属性列表 */
    @Operation(summary = "查询类属性列表")
    @GetMapping("/class/{classId}/properties")
    @PreAuthorize("@perm.has('odin:ontology:query')")
    public Result<List<OntPropertyVO>> listProperties(@Parameter(name = "classId", description = "类ID", required = true) @PathVariable Long classId) {
        return Result.ok(ontologyService.listProperties(classId));
    }

    /** 新增属性（classId 由请求体携带） */
    @Operation(summary = "新增类属性")
    @PostMapping("/property")
    @PreAuthorize("@perm.has('odin:ontology:create')")
    @OperationLog(module = "本体管理", action = "创建属性")
    public Result<Long> createProperty(@Valid @RequestBody OntPropertyDTO dto) {
        return Result.ok(ontologyService.createProperty(dto));
    }

    /** 更新属性（id 由请求体携带） */
    @Operation(summary = "更新类属性")
    @PutMapping("/property")
    @PreAuthorize("@perm.has('odin:ontology:update')")
    @OperationLog(module = "本体管理", action = "更新属性")
    public Result<Void> updateProperty(@Valid @RequestBody OntPropertyDTO dto) {
        ontologyService.updateProperty(dto);
        return Result.ok();
    }

    /** 删除属性 */
    @Operation(summary = "删除类属性")
    @DeleteMapping("/property/{id}")
    @PreAuthorize("@perm.has('odin:ontology:delete')")
    @OperationLog(module = "本体管理", action = "删除属性")
    public Result<Void> deleteProperty(@Parameter(name = "id", description = "属性ID", required = true) @PathVariable Long id) {
        ontologyService.deleteProperty(id);
        return Result.ok();
    }

    // ==================== 关系 ====================

    /** 查询本体下所有关系 */
    @Operation(summary = "查询本体关系列表")
    @GetMapping("/{id}/relations")
    @PreAuthorize("@perm.has('odin:ontology:query')")
    public Result<List<OntRelationVO>> listRelations(@Parameter(name = "id", description = "本体ID", required = true) @PathVariable Long id) {
        return Result.ok(ontologyService.listRelations(id));
    }

    /** 根据ID查询关系详情 */
    @Operation(summary = "根据ID查询本体关系")
    @GetMapping("/relation/{id}")
    @PreAuthorize("@perm.has('odin:ontology:query')")
    public Result<OntRelationVO> getRelationById(@Parameter(name = "id", description = "关系ID", required = true) @PathVariable Long id) {
        return Result.ok(ontologyService.getRelationById(id));
    }

    /** 新增关系（ontologyId 由请求体携带） */
    @Operation(summary = "新增本体关系")
    @PostMapping("/relation")
    @PreAuthorize("@perm.has('odin:ontology:create')")
    @OperationLog(module = "本体管理", action = "创建关系")
    public Result<Long> createRelation(@Valid @RequestBody OntRelationDTO dto) {
        return Result.ok(ontologyService.createRelation(dto));
    }

    /** 更新关系（id 由请求体携带） */
    @Operation(summary = "更新本体关系")
    @PutMapping("/relation")
    @PreAuthorize("@perm.has('odin:ontology:update')")
    @OperationLog(module = "本体管理", action = "更新关系")
    public Result<Void> updateRelation(@Valid @RequestBody OntRelationDTO dto) {
        ontologyService.updateRelation(dto);
        return Result.ok();
    }

    /** 删除关系 */
    @Operation(summary = "删除本体关系")
    @DeleteMapping("/relation/{id}")
    @PreAuthorize("@perm.has('odin:ontology:delete')")
    @OperationLog(module = "本体管理", action = "删除关系")
    public Result<Void> deleteRelation(@Parameter(name = "id", description = "关系ID", required = true) @PathVariable Long id) {
        ontologyService.deleteRelation(id);
        return Result.ok();
    }

    // ==================== 规则 ====================

    /** 查询本体下所有规则 */
    @Operation(summary = "查询本体规则列表")
    @GetMapping("/{id}/rules")
    @PreAuthorize("@perm.has('odin:ontology:query')")
    public Result<List<OntRuleVO>> listRules(@Parameter(name = "id", description = "本体ID", required = true) @PathVariable Long id) {
        return Result.ok(ontologyService.listRules(id));
    }

    /** 新增规则（ontologyId 由请求体携带） */
    @Operation(summary = "新增本体规则")
    @PostMapping("/rule")
    @PreAuthorize("@perm.has('odin:ontology:create')")
    @OperationLog(module = "本体管理", action = "创建规则")
    public Result<Long> createRule(@Valid @RequestBody OntRuleDTO dto) {
        return Result.ok(ontologyService.createRule(dto));
    }

    /** 更新规则（id 由请求体携带） */
    @Operation(summary = "更新本体规则")
    @PutMapping("/rule")
    @PreAuthorize("@perm.has('odin:ontology:update')")
    @OperationLog(module = "本体管理", action = "更新规则")
    public Result<Void> updateRule(@Valid @RequestBody OntRuleDTO dto) {
        ontologyService.updateRule(dto);
        return Result.ok();
    }

    /** 删除规则 */
    @Operation(summary = "删除本体规则")
    @DeleteMapping("/rule/{id}")
    @PreAuthorize("@perm.has('odin:ontology:delete')")
    @OperationLog(module = "本体管理", action = "删除规则")
    public Result<Void> deleteRule(@Parameter(name = "id", description = "规则ID", required = true) @PathVariable Long id) {
        ontologyService.deleteRule(id);
        return Result.ok();
    }

    // ==================== 动作 ====================

    /** 查询本体下所有动作 */
    @Operation(summary = "查询本体动作列表")
    @GetMapping("/{id}/actions")
    @PreAuthorize("@perm.has('odin:ontology:query')")
    public Result<List<OntActionVO>> listActions(@Parameter(name = "id", description = "本体ID", required = true) @PathVariable Long id) {
        return Result.ok(ontologyService.listActions(id));
    }

    /** 新增动作（ontologyId 由请求体携带） */
    @Operation(summary = "新增本体动作")
    @PostMapping("/action")
    @PreAuthorize("@perm.has('odin:ontology:create')")
    @OperationLog(module = "本体管理", action = "创建动作")
    public Result<Long> createAction(@Valid @RequestBody OntActionDTO dto) {
        return Result.ok(ontologyService.createAction(dto));
    }

    /** 更新动作（id 由请求体携带） */
    @Operation(summary = "更新本体动作")
    @PutMapping("/action")
    @PreAuthorize("@perm.has('odin:ontology:update')")
    @OperationLog(module = "本体管理", action = "更新动作")
    public Result<Void> updateAction(@Valid @RequestBody OntActionDTO dto) {
        ontologyService.updateAction(dto);
        return Result.ok();
    }

    /** 删除动作 */
    @Operation(summary = "删除本体动作")
    @DeleteMapping("/action/{id}")
    @PreAuthorize("@perm.has('odin:ontology:delete')")
    @OperationLog(module = "本体管理", action = "删除动作")
    public Result<Void> deleteAction(@Parameter(name = "id", description = "动作ID", required = true) @PathVariable Long id) {
        ontologyService.deleteAction(id);
        return Result.ok();
    }
}
