import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  getOntologyGraph,
  getOntClassTree,
  getOntRelations,
  getOntRules,
  createOntClass,
  updateOntClass,
  deleteOntClass,
  createOntRelation,
  updateOntRelation,
  deleteOntRelation,
  saveOntologyLayout,
  getChangeLog,
} from '@/api/odin/ontology'
import type {
  OntClassVO,
  OntClassDTO,
  OntRelationVO,
  OntRelationDTO,
  OntRuleVO,
  OntologyGraphVO,
  NodePosition,
  ChangeLogVO,
  ChangeLogQuery,
} from '@/types/odin'

export const useOdinOntologyStore = defineStore('odinOntology', () => {
  const classes = ref<OntClassVO[]>([])
  const relations = ref<OntRelationVO[]>([])
  const rules = ref<OntRuleVO[]>([])
  const graph = ref<OntologyGraphVO | null>(null)
  const layoutData = ref<Map<string, NodePosition>>(new Map())
  const focusedClassId = ref<string>('')
  const changeLog = ref<ChangeLogVO[]>([])
  const loading = ref(false)

  /** 统计信息 */
  const stats = computed(() => {
    const classCount = classes.value.length
    const relationCount = relations.value.length
    // 属性数：从 graph nodes 的 properties 字段统计
    const propertyCount = graph.value?.nodes?.reduce(
      (sum, n) => sum + (n.properties?.length ?? 0), 0,
    ) ?? 0
    const coverageRate = classCount > 0
      ? Math.round((classes.value.filter((c) => (c.propertyCount ?? 0) > 0).length / classCount) * 100)
      : 0
    return { classCount, relationCount, propertyCount, coverageRate }
  })

  /** 加载本体图谱数据（类 + 关系 + 规则） */
  async function loadGraph(ontologyId: string): Promise<void> {
    loading.value = true
    try {
      const [graphData, classTree, relationList, ruleList] = await Promise.all([
        getOntologyGraph(ontologyId),
        getOntClassTree(ontologyId),
        getOntRelations(ontologyId),
        getOntRules(ontologyId),
      ])
      graph.value = graphData
      classes.value = classTree
      relations.value = relationList
      rules.value = ruleList
      // 从 graph nodes 提取布局坐标（如果有）
      const positions = new Map<string, NodePosition>()
      // graph nodes 可能携带 position 信息，暂存备用
      layoutData.value = positions
    } catch (e) {
      console.error('[odinOntology] loadGraph failed', e)
    } finally {
      loading.value = false
    }
  }

  /** 加载变更日志 */
  async function loadChangeLog(ontologyId: string): Promise<void> {
    try {
      const params: ChangeLogQuery = { current: 1, size: 50, ontologyId }
      const res = await getChangeLog(params)
      changeLog.value = res?.records ?? []
    } catch (e) {
      console.error('[odinOntology] loadChangeLog failed', e)
    }
  }

  /** 保存布局 */
  async function saveLayout(ontologyId: string): Promise<void> {
    try {
      const layout: Record<string, NodePosition> = {}
      layoutData.value.forEach((pos, key) => { layout[key] = pos })
      await saveOntologyLayout(ontologyId, layout)
    } catch (e) {
      console.error('[odinOntology] saveLayout failed', e)
    }
  }

  /** 聚焦到某个类 */
  function focusClass(id: string) {
    focusedClassId.value = id
  }

  /** 新增类 */
  async function addClass(data: OntClassDTO): Promise<string> {
    const id = await createOntClass(data)
    return id
  }

  /** 更新类 */
  async function updateClass(data: OntClassDTO): Promise<void> {
    await updateOntClass(data)
  }

  /** 删除类 */
  async function deleteClass(id: string): Promise<void> {
    await deleteOntClass(id)
  }

  /** 新增关系 */
  async function addRelation(data: OntRelationDTO): Promise<string> {
    const id = await createOntRelation(data)
    return id
  }

  /** 更新关系 */
  async function updateRelation(data: OntRelationDTO): Promise<void> {
    await updateOntRelation(data)
  }

  /** 删除关系 */
  async function deleteRelation(id: string): Promise<void> {
    await deleteOntRelation(id)
  }

  /** 重置 */
  function reset() {
    classes.value = []
    relations.value = []
    rules.value = []
    graph.value = null
    layoutData.value = new Map()
    focusedClassId.value = ''
    changeLog.value = []
  }

  return {
    classes,
    relations,
    rules,
    graph,
    layoutData,
    focusedClassId,
    changeLog,
    loading,
    stats,
    loadGraph,
    loadChangeLog,
    saveLayout,
    focusClass,
    addClass,
    updateClass,
    deleteClass,
    addRelation,
    updateRelation,
    deleteRelation,
    reset,
  }
})
