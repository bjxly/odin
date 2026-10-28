/**
 * 格式化日期时间
 */
export function formatDate(date: string | Date | undefined, fmt = 'yyyy-MM-dd HH:mm:ss'): string {
  if (!date) return '-'
  const d = typeof date === 'string' ? new Date(date) : date
  if (isNaN(d.getTime())) return '-'

  const o: Record<string, number> = {
    'M+': d.getMonth() + 1,
    'd+': d.getDate(),
    'H+': d.getHours(),
    'm+': d.getMinutes(),
    's+': d.getSeconds(),
    'q+': Math.floor((d.getMonth() + 3) / 3),
    S: d.getMilliseconds(),
  }

  if (/(y+)/.test(fmt)) {
    fmt = fmt.replace(RegExp.$1, (d.getFullYear() + '').substr(4 - RegExp.$1.length))
  }

  for (const k in o) {
    if (new RegExp('(' + k + ')').test(fmt)) {
      const val = o[k] < 10 && RegExp.$1.length > 1 ? '0' + o[k] : o[k]
      fmt = fmt.replace(RegExp.$1, val as string)
    }
  }

  return fmt
}

/**
 * 格式化耗时（毫秒 -> 可读字符串）
 */
export function formatDuration(ms: number | undefined): string {
  if (ms === undefined || ms === null || ms < 0) return '-'
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
  const minutes = Math.floor(ms / 60000)
  const seconds = Math.floor((ms % 60000) / 1000)
  return `${minutes}m ${seconds}s`
}
