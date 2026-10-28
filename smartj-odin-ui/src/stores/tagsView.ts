import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import type { RouteLocationNormalized, Router } from 'vue-router'

export interface TagView {
  path: string
  fullPath: string
  title: string
  name?: string
  affix?: boolean
}

const STORAGE_KEY = 'smartj-tagsview'

/** 从 sessionStorage 恢复非 affix 的标签列表 */
function restoreTags(): TagView[] {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY)
    if (raw) {
      const parsed: TagView[] = JSON.parse(raw)
      return parsed.filter((t) => t.path && !t.affix)
    }
  } catch {
    // ignore
  }
  return []
}

/** 将非 affix 标签同步到 sessionStorage */
function persistTags(tags: TagView[]) {
  try {
    const nonAffix = tags.filter((t) => !t.affix)
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(nonAffix))
  } catch {
    // ignore
  }
}

export const useTagsViewStore = defineStore('tagsView', () => {
  const visitedTags = ref<TagView[]>(restoreTags())

  // 自动同步到 sessionStorage
  watch(
    visitedTags,
    (tags) => {
      persistTags(tags)
    },
    { deep: true }
  )

  function addTag(route: RouteLocationNormalized, router?: Router) {
    // 忽略无名路由（如重定向）
    if (!route.name) return

    // 从 route.meta 或 router.getRoutes() 回退查找 affix 标记
    let isAffix = !!route.meta?.affix
    if (!isAffix && router) {
      const configRoute = router.getRoutes().find((r) => r.path === route.path)
      isAffix = !!configRoute?.meta?.affix
    }

    const exists = visitedTags.value.find((t) => t.path === route.path)
    if (exists) {
      exists.fullPath = route.fullPath
      if (isAffix) exists.affix = true
      return
    }
    visitedTags.value.push({
      path: route.path,
      fullPath: route.fullPath,
      title: (route.meta?.title as string) || (route.name as string) || route.path,
      name: route.name as string,
      affix: isAffix,
    })
  }

  function removeTag(path: string): TagView | undefined {
    const tag = visitedTags.value.find((t) => t.path === path)
    if (!tag || tag.affix) return undefined
    const idx = visitedTags.value.indexOf(tag)
    visitedTags.value.splice(idx, 1)
    // 返回最后一个标签，用于关闭当前激活标签后跳转
    return visitedTags.value[visitedTags.value.length - 1]
  }

  function removeOthers(path: string) {
    visitedTags.value = visitedTags.value.filter((t) => t.path === path || t.affix)
  }

  function removeAll(): TagView | undefined {
    visitedTags.value = visitedTags.value.filter((t) => t.affix)
    return visitedTags.value[visitedTags.value.length - 1]
  }

  function initAffixTags(routes: Array<{ path: string; name?: string; meta?: Record<string, any> }>) {
    routes.forEach((route) => {
      if (route.meta?.affix) {
        const exists = visitedTags.value.find((t) => t.path === route.path)
        if (!exists) {
          visitedTags.value.push({
            path: route.path,
            fullPath: route.path,
            title: (route.meta?.title as string) || (route.name as string) || route.path,
            name: route.name as string,
            affix: true,
          })
        } else {
          // 已有标签（如从 sessionStorage 恢复的），同步 affix 标记
          exists.affix = true
        }
      }
    })
  }

  /** 过滤掉路径不在已注册路由中的标签（菜单动态变更后防止死链） */
  function restoreValidTags(registeredPaths: Set<string>) {
    visitedTags.value = visitedTags.value.filter((t) => t.affix || registeredPaths.has(t.path))
  }

  /** 清除持久化数据（登出时调用） */
  function clearPersisted() {
    try {
      sessionStorage.removeItem(STORAGE_KEY)
    } catch {
      // ignore
    }
  }

  return {
    visitedTags,
    addTag,
    removeTag,
    removeOthers,
    removeAll,
    initAffixTags,
    restoreValidTags,
    clearPersisted,
  }
})
