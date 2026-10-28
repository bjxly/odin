import { ref, reactive } from 'vue'

export interface PaginationParams {
  current: number
  size: number
  [key: string]: any
}

export interface PageResult<T> {
  records: T[]
  total: number
}

export function usePagination<T = any>(fetchFn: (params: PaginationParams) => Promise<PageResult<T>>) {
  const loading = ref(false)
  const data = ref<T[]>([]) as ReturnType<typeof ref<T[]>>
  const total = ref(0)
  const query = reactive<PaginationParams>({ current: 1, size: 10 })

  async function load(params?: Partial<PaginationParams>) {
    if (params) {
      Object.assign(query, params)
    }
    loading.value = true
    try {
      const res = await fetchFn({ ...query })
      data.value = (res.records ?? []) as T[]
      total.value = res.total ?? 0
    } finally {
      loading.value = false
    }
  }

  function refresh() {
    query.current = 1
    return load()
  }

  function search(params?: Record<string, any>) {
    if (params) {
      Object.assign(query, params)
    }
    query.current = 1
    return load()
  }

  function onPageChange(page: number) {
    query.current = page
    return load()
  }

  function onSizeChange(size: number) {
    query.size = size
    query.current = 1
    return load()
  }

  return {
    loading,
    data,
    total,
    query,
    load,
    refresh,
    search,
    onPageChange,
    onSizeChange,
  }
}
