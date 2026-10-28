/**
 * 消息 Markdown 渲染：marked(GFM) + DOMPurify。
 *
 * <p>代码高亮由调用组件在渲染后对 {@code pre code} 元素做后处理
 * （marked v12 无内置 highlight），见 TextBlock.vue。
 *
 * <p>流式口径：{@link balanceFences} 会把未闭合的围栏代码块自动补 ``` ，
 * 避免 token 逐字到达时代码块渲染抖动。
 */
import { marked } from 'marked'
import DOMPurify from 'dompurify'

// gfm:true 是表格（| a | b | + |---|---|）渲染的必要条件；breaks:true 让单换行生效
marked.setOptions({ gfm: true, breaks: true })

/** 流式中未闭合的围栏代码块补齐（偶数个 ``` 才平衡） */
function balanceFences(text: string): string {
  const fences = (text.match(/```/g) || []).length
  return fences % 2 === 1 ? `${text}\n\`\`\`` : text
}

/** markdown → 安全 HTML（经 DOMPurify 净化，防 XSS） */
export function renderMarkdown(text: string): string {
  if (!text) return ''
  const html = marked.parse(balanceFences(text)) as string
  return DOMPurify.sanitize(html, {
    ADD_ATTR: ['target'],
  })
}
