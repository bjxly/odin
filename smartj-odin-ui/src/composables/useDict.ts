import { ref, computed } from 'vue'
import { getDictDataByType } from '@/api/sys/dict'
import type { DictDataVO } from '@/types/sys'

export interface DictItem {
  label: string
  value: string | number
  tagType?: string
  [key: string]: any
}

export function useDict(dictTypes: string[]) {
  const dictMap = ref<Record<string, DictItem[]>>({})
  const loading = ref(false)

  async function load() {
    if (!dictTypes.length) return
    loading.value = true
    try {
      const results = await Promise.all(
        dictTypes.map((type) =>
          getDictDataByType(type).then(
            (data: DictDataVO[]) => ({ type, data: Array.isArray(data) ? data : [] }),
            () => ({ type, data: [] as DictDataVO[] })
          )
        )
      )
      const map: Record<string, DictItem[]> = {}
      results.forEach(({ type, data }: { type: string; data: DictDataVO[] }) => {
        map[type] = data.map((item: any) => ({
          label: item.label || item.name || String(item.value),
          value: item.value,
          tagType: item.type || item.tagType,
        }))
      })
      dictMap.value = map
    } finally {
      loading.value = false
    }
  }

  function getLabel(type: string, value: string | number): string {
    const items = dictMap.value[type]
    if (!items) return String(value)
    const item = items.find((i) => i.value === value || String(i.value) === String(value))
    return item?.label ?? String(value)
  }

  function getTagType(type: string, value: string | number): string {
    const items = dictMap.value[type]
    if (!items) return ''
    const item = items.find((i) => i.value === value || String(i.value) === String(value))
    return item?.tagType ?? ''
  }

  const isEmpty = computed(() => Object.keys(dictMap.value).length === 0)

  return { dictMap, loading, load, getLabel, getTagType, isEmpty }
}
