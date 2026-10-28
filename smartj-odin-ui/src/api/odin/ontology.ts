import { get, post, put, del } from '@/api/request'
import type { PageResult } from '@/types'
import type {
  OntologyVO,
  OntologyDTO,
  OntologyQuery,
  OntClassVO,
  OntClassDTO,
  OntPropertyVO,
  OntPropertyDTO,
  OntRelationVO,
  OntRelationDTO,
  OntRuleVO,
  OntRuleDTO,
  OntActionVO,
  OntActionDTO,
  OntologyGraphVO,
  VocabularyVO,
  OntologySummary,
  NodePosition,
  ChangeLogVO,
  ChangeLogQuery,
} from '@/types/odin'

const BASE = '/odin/ontology'

// ===== 本体定义 =====

/** 分页查询本体 */
export function getOntologyPage(params: OntologyQuery) {
  return get<PageResult<OntologyVO>>(`${BASE}/page`, { params })
}

/** 查询全部启用本体（下拉用） */
export function getOntologyList() {
  return get<OntologyVO[]>(`${BASE}/list`)
}

/** 根据ID查询本体 */
export function getOntologyById(id: string) {
  return get<OntologyVO>(`${BASE}/${id}`)
}

/** 新增本体，返回ID */
export function createOntology(data: OntologyDTO) {
  return post<string>(BASE, data)
}

/** 更新本体 */
export function updateOntology(data: OntologyDTO) {
  return put<void>(BASE, data)
}

/** 删除本体 */
export function deleteOntology(id: string) {
  return del<void>(`${BASE}/${id}`)
}

/** 发布本体版本 */
export function publishOntology(id: string) {
  return post<void>(`${BASE}/${id}/publish`)
}

/** 查询本体图谱（类+关系可视化数据） */
export function getOntologyGraph(id: string) {
  return get<OntologyGraphVO>(`${BASE}/${id}/graph`)
}

/** 查询本体词表（LLM 问数用的逻辑 Schema 词汇） */
export function getOntologyVocabulary(id: string) {
  return get<VocabularyVO>(`${BASE}/${id}/vocabulary`)
}

// ===== 本体类 =====

/** 查询本体下类树 */
export function getOntClassTree(ontologyId: string) {
  return get<OntClassVO[]>(`${BASE}/${ontologyId}/classes`)
}

/** 根据ID查询本体类 */
export function getOntClassById(id: string) {
  return get<OntClassVO>(`${BASE}/class/${id}`)
}

/** 新增本体类，返回ID */
export function createOntClass(data: OntClassDTO) {
  return post<string>(`${BASE}/class`, data)
}

/** 更新本体类 */
export function updateOntClass(data: OntClassDTO) {
  return put<void>(`${BASE}/class`, data)
}

/** 删除本体类 */
export function deleteOntClass(id: string) {
  return del<void>(`${BASE}/class/${id}`)
}

// ===== 属性 =====

/** 查询类下属性列表 */
export function getOntProperties(classId: string) {
  return get<OntPropertyVO[]>(`${BASE}/class/${classId}/properties`)
}

/** 新增属性，返回ID */
export function createOntProperty(data: OntPropertyDTO) {
  return post<string>(`${BASE}/property`, data)
}

/** 更新属性 */
export function updateOntProperty(data: OntPropertyDTO) {
  return put<void>(`${BASE}/property`, data)
}

/** 删除属性 */
export function deleteOntProperty(id: string) {
  return del<void>(`${BASE}/property/${id}`)
}

// ===== 关系 =====

/** 查询本体下关系列表 */
export function getOntRelations(ontologyId: string) {
  return get<OntRelationVO[]>(`${BASE}/${ontologyId}/relations`)
}

/** 根据ID查询关系 */
export function getOntRelationById(id: string) {
  return get<OntRelationVO>(`${BASE}/relation/${id}`)
}

/** 新增关系，返回ID */
export function createOntRelation(data: OntRelationDTO) {
  return post<string>(`${BASE}/relation`, data)
}

/** 更新关系 */
export function updateOntRelation(data: OntRelationDTO) {
  return put<void>(`${BASE}/relation`, data)
}

/** 删除关系 */
export function deleteOntRelation(id: string) {
  return del<void>(`${BASE}/relation/${id}`)
}

// ===== 规则 =====

/** 查询本体下规则列表 */
export function getOntRules(ontologyId: string) {
  return get<OntRuleVO[]>(`${BASE}/${ontologyId}/rules`)
}

/** 新增规则，返回ID */
export function createOntRule(data: OntRuleDTO) {
  return post<string>(`${BASE}/rule`, data)
}

/** 更新规则 */
export function updateOntRule(data: OntRuleDTO) {
  return put<void>(`${BASE}/rule`, data)
}

/** 删除规则 */
export function deleteOntRule(id: string) {
  return del<void>(`${BASE}/rule/${id}`)
}

// ===== 动作 =====

/** 查询本体下动作列表 */
export function getOntActions(ontologyId: string) {
  return get<OntActionVO[]>(`${BASE}/${ontologyId}/actions`)
}

/** 新增动作，返回ID */
export function createOntAction(data: OntActionDTO) {
  return post<string>(`${BASE}/action`, data)
}

/** 更新动作 */
export function updateOntAction(data: OntActionDTO) {
  return put<void>(`${BASE}/action`, data)
}

/** 删除动作 */
export function deleteOntAction(id: string) {
  return del<void>(`${BASE}/action/${id}`)
}

// ===== 本体摘要/布局/导入导出 =====

/** 查询本体摘要列表（轻量信息，用于工作台切换器） */
export function getOntologySummaryList() {
  return get<OntologySummary[]>(`${BASE}/summaries`)
}

/** 保存本体布局坐标 */
export function saveOntologyLayout(id: string, layout: Record<string, NodePosition>) {
  return put<void>(`${BASE}/${id}/layout`, layout)
}

/** 导出本体（返回 JSON 字符串） */
export function exportOntology(id: string) {
  return get<string>(`${BASE}/${id}/export`)
}

/** 导入本体 */
export function importOntology(id: string, data: Record<string, unknown>) {
  return post<void>(`${BASE}/${id}/import`, data, { timeout: 120000 })
}

/** 查询变更日志 */
export function getChangeLog(params: ChangeLogQuery) {
  return get<PageResult<ChangeLogVO>>(`${BASE}/changes`, { params })
}
