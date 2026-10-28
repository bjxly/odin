import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useWorkspaceStore = defineStore('workspace', () => {
  // 全局UI状态
  const currentOntologyId = ref<number | null>(null)
  const sidebarCollapsed = ref(false)
  const appName = ref('SmartG ODIN')
  const loading = ref(false)
  const loaded = ref(false)
  const error = ref<string | null>(null)

  function setCurrentOntology(id: number | null) {
    currentOntologyId.value = id
  }

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  return {
    currentOntologyId,
    sidebarCollapsed,
    appName,
    loading,
    loaded,
    error,
    setCurrentOntology,
    toggleSidebar,
  }
})
