const loadedUrls = new Set<string>()

/**
 * 懒加载 iconfont font-class 模式的 CSS 文件。
 * 相同 URL 只会被插入一次，避免重复请求。
 *
 * @param url CSS 文件地址，例如 /iconfont/iconfont.css
 */
export function useIconfontStyle(url?: string) {
  if (!url || loadedUrls.has(url)) return

  loadedUrls.add(url)
  const link = document.createElement('link')
  link.rel = 'stylesheet'
  link.href = url
  document.head.appendChild(link)
}
