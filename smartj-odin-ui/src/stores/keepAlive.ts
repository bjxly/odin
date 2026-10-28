import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

const STORAGE_KEY = 'smartj-keepalive'

function restoreViews(): string[] {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY)
    return raw ? JSON.parse(raw) : []
  } catch {
    return []
  }
}

export const useKeepAliveStore = defineStore('keepAlive', () => {
  const cachedViews = ref<string[]>(restoreViews())

  watch(
    cachedViews,
    (views) => {
      try {
        sessionStorage.setItem(STORAGE_KEY, JSON.stringify(views))
      } catch {
        // ignore
      }
    },
    { deep: true }
  )

  function addCache(name: string) {
    if (name && !cachedViews.value.includes(name)) {
      cachedViews.value.push(name)
    }
  }

  function removeCache(name: string) {
    const idx = cachedViews.value.indexOf(name)
    if (idx > -1) {
      cachedViews.value.splice(idx, 1)
    }
  }

  function clearCache() {
    cachedViews.value = []
  }

  /** 清除持久化数据（登出时调用） */
  function clearPersisted() {
    try {
      sessionStorage.removeItem(STORAGE_KEY)
    } catch {
      // ignore
    }
  }

  return { cachedViews, addCache, removeCache, clearCache, clearPersisted }
})
