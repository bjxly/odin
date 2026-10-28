import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getOntologySummaryList } from '@/api/odin/ontology'
import type { OntologySummary } from '@/types/odin'

export const useOdinWorkspaceStore = defineStore('odinWorkspace', () => {
  const activeOntologyId = ref<string>('')
  const ontologies = ref<OntologySummary[]>([])
  const dirty = ref(false)
  const lastSavedAt = ref<string | null>(null)
  const loading = ref(false)

  const activeOntology = computed(() =>
    ontologies.value.find((o) => o.id === activeOntologyId.value) ?? null,
  )

  async function loadOntologies(): Promise<OntologySummary[]> {
    loading.value = true
    try {
      ontologies.value = await getOntologySummaryList()
      // 若当前无选中本体且有列表，自动选中第一个
      if (!activeOntologyId.value && ontologies.value.length > 0) {
        activeOntologyId.value = ontologies.value[0].id
      }
    } catch (e) {
      console.error('[odinWorkspace] loadOntologies failed', e)
    } finally {
      loading.value = false
    }
    return ontologies.value
  }

  function setActive(id: string) {
    if (activeOntologyId.value !== id) {
      activeOntologyId.value = id
    }
  }

  function markDirty() {
    dirty.value = true
  }

  function save() {
    dirty.value = false
    lastSavedAt.value = new Date().toISOString()
  }

  function reset() {
    activeOntologyId.value = ''
    ontologies.value = []
    dirty.value = false
    lastSavedAt.value = null
  }

  return {
    activeOntologyId,
    ontologies,
    dirty,
    lastSavedAt,
    loading,
    activeOntology,
    loadOntologies,
    setActive,
    markDirty,
    save,
    reset,
  }
})
