import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface DictItem {
  label: string
  value: string | number
  cssClass?: string
  tagType?: '' | 'success' | 'warning' | 'danger' | 'info'
}

const mockDictData: Record<string, DictItem[]> = {
  task_status: [
    { label: '待执行', value: 0, tagType: 'info' },
    { label: '运行中', value: 1, tagType: '' },
    { label: '成功', value: 2, tagType: 'success' },
    { label: '失败', value: 3, tagType: 'danger' },
    { label: '已终止', value: 4, tagType: 'warning' },
  ],
  task_type: [
    { label: 'Shell', value: 'shell', tagType: '' },
    { label: 'HTTP', value: 'http', tagType: 'success' },
    { label: 'SQL', value: 'sql', tagType: 'warning' },
    { label: 'Python', value: 'python', tagType: 'info' },
  ],
  trigger_type: [
    { label: '手动', value: 'manual', tagType: '' },
    { label: '定时', value: 'cron', tagType: 'success' },
    { label: 'API', value: 'api', tagType: 'warning' },
    { label: '子流程', value: 'subflow', tagType: 'info' },
  ],
  yes_no: [
    { label: '是', value: 1, tagType: 'success' },
    { label: '否', value: 0, tagType: 'danger' },
  ],
}

export const useDictStore = defineStore('dict', () => {
  const dictMap = ref<Map<string, DictItem[]>>(new Map())

  async function loadDict(dictType: string): Promise<DictItem[]> {
    if (dictMap.value.has(dictType)) {
      return dictMap.value.get(dictType)!
    }
    const data = mockDictData[dictType] || []
    dictMap.value.set(dictType, data)
    return data
  }

  function getDict(dictType: string): DictItem[] {
    return dictMap.value.get(dictType) || []
  }

  function getDictLabel(dictType: string, value: string | number): string {
    const items = dictMap.value.get(dictType) || []
    const item = items.find((i) => i.value === value || String(i.value) === String(value))
    return item?.label ?? String(value)
  }

  function getDictTagType(dictType: string, value: string | number): string {
    const items = dictMap.value.get(dictType) || []
    const item = items.find((i) => i.value === value || String(i.value) === String(value))
    return item?.tagType ?? ''
  }

  return { dictMap, loadDict, getDict, getDictLabel, getDictTagType }
})
