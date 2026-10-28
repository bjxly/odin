import { ref, watch, computed } from 'vue'
import type { TableColumn, ColumnPreference } from './types'

const STORAGE_PREFIX = 'sm-table-pref:'

function load(key: string): Record<string, ColumnPreference> | null {
  try {
    const raw = localStorage.getItem(STORAGE_PREFIX + key)
    if (!raw) return null
    const parsed = JSON.parse(raw)
    if (typeof parsed === 'object' && parsed !== null) return parsed
    return null
  } catch {
    return null
  }
}

function save(key: string, data: Record<string, ColumnPreference>) {
  try {
    localStorage.setItem(STORAGE_PREFIX + key, JSON.stringify(data))
  } catch {
    // storage full or disabled
  }
}

export function useColumnPreference(key: string | undefined, columns: () => TableColumn[]) {
  const preferences = ref<Record<string, ColumnPreference>>({})

  // 构建默认配置（基于 columns 的初始状态）
  function buildDefaults(): Record<string, ColumnPreference> {
    const cols = columns()
    const result: Record<string, ColumnPreference> = {}
    cols.forEach((col, idx) => {
      const prop = col.prop || col.label
      result[prop] = {
        prop,
        visible: col.visible !== false,
        order: idx,
      }
    })
    return result
  }

  // 初始化：优先从 localStorage 加载，否则使用默认值
  function init() {
    if (key) {
      const saved = load(key)
      if (saved) {
        // 合并保存的配置与当前 columns（处理列增减变化）
        const defaults = buildDefaults()
        const merged: Record<string, ColumnPreference> = {}
        // 先放入所有当前列默认值
        Object.assign(merged, defaults)
        // 用保存的配置覆盖（仅覆盖仍然存在的列）
        for (const prop of Object.keys(saved)) {
          if (merged[prop]) {
            merged[prop] = saved[prop]
          }
        }
        preferences.value = merged
        return
      }
    }
    preferences.value = buildDefaults()
  }

  init()

  // 当 columns 变化时（如新增/删除列），重新合并
  watch(
    columns,
    () => {
      const defaults = buildDefaults()
      const merged = { ...preferences.value }
      // 新增列
      for (const prop of Object.keys(defaults)) {
        if (!merged[prop]) {
          merged[prop] = defaults[prop]
        }
      }
      // 删除已经不存在的列
      for (const prop of Object.keys(merged)) {
        if (!defaults[prop]) {
          delete merged[prop]
        }
      }
      preferences.value = merged
      persist()
    },
    { deep: true }
  )

  // 持久化
  function persist() {
    if (key) {
      save(key, preferences.value)
    }
  }

  // 可见 + 按 order 排序的列列表
  const visibleColumns = computed(() => {
    const cols = columns()
    const pref = preferences.value
    const result = cols
      .filter((col) => {
        const prop = col.prop || col.label
        return pref[prop]?.visible !== false
      })
      .sort((a, b) => {
        const propA = a.prop || a.label
        const propB = b.prop || b.label
        return (pref[propA]?.order ?? 0) - (pref[propB]?.order ?? 0)
      })
    return result
  })

  // 所有列配置列表（用于设置面板）
  const columnConfigs = computed(() => {
    const cols = columns()
    const pref = preferences.value
    return cols
      .map((col, idx) => {
        const prop = col.prop || col.label
        return {
          prop,
          label: col.label,
          visible: pref[prop]?.visible ?? col.visible !== false,
          order: pref[prop]?.order ?? idx,
        }
      })
      .sort((a, b) => a.order - b.order)
  })

  // 切换列可见性
  function toggleColumn(prop: string) {
    if (preferences.value[prop]) {
      preferences.value[prop].visible = !preferences.value[prop].visible
      persist()
    }
  }

  // 更新列顺序（拖拽后）
  function updateOrder(orderedProps: string[]) {
    orderedProps.forEach((prop, idx) => {
      if (preferences.value[prop]) {
        preferences.value[prop].order = idx
      }
    })
    persist()
  }

  // 重置为默认
  function resetToDefaults() {
    preferences.value = buildDefaults()
    if (key) {
      localStorage.removeItem(STORAGE_PREFIX + key)
    }
  }

  return {
    preferences,
    visibleColumns,
    columnConfigs,
    toggleColumn,
    updateOrder,
    resetToDefaults,
  }
}
