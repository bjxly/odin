import { ref } from 'vue'
import { TOKEN_KEY } from '@/utils/constants'
import type { SSEEvent } from '@/types/odin'

/**
 * ODIN SSE 流式通信组合函数。
 * 使用原生 fetch + ReadableStream + TextDecoder({stream:true})，
 * Token 从 localStorage/sessionStorage 读取注入 Authorization header。
 * 支持 AbortController 中止。
 */
export function useOdinSSE() {
  const isStreaming = ref(false)
  let controller: AbortController | null = null

  /** 获取当前认证 Token */
  function getAuthToken(): string | null {
    return localStorage.getItem(TOKEN_KEY) || sessionStorage.getItem(TOKEN_KEY)
  }

  /**
   * 解析单个 SSE 事件块（按空行切分后的文本）
   */
  function parseSSEChunk(chunk: string): SSEEvent | null {
    const lines = chunk.split(/\r?\n/)
    let event = ''
    const dataLines: string[] = []
    for (const rawLine of lines) {
      const line = rawLine.endsWith('\r') ? rawLine.slice(0, -1) : rawLine
      if (!line) continue
      if (line.startsWith(':')) continue
      const idx = line.indexOf(':')
      let field: string
      let value: string
      if (idx === -1) {
        field = line
        value = ''
      } else {
        field = line.slice(0, idx)
        value = line.slice(idx + 1)
        if (value.startsWith(' ')) value = value.slice(1)
      }
      if (field === 'event') {
        event = value.trim()
      } else if (field === 'data') {
        dataLines.push(value)
      }
    }
    const raw = dataLines.join('\n')
    if (!event && !raw) return null
    let data: unknown = null
    if (raw) {
      try {
        data = JSON.parse(raw)
      } catch {
        data = raw
      }
    }
    return { event: event || 'message', data, raw }
  }

  /**
   * 建立 SSE 连接
   * @param url - 请求地址
   * @param body - POST 请求体
   * @param onEvent - 每收到一个事件的回调
   * @param onError - 错误回调
   * @param onComplete - 完成回调
   * @param signal - 可选的外部 AbortSignal
   */
  async function connect(
    url: string,
    body: unknown,
    onEvent: (event: SSEEvent) => void,
    onError?: (error: Error) => void,
    onComplete?: () => void,
    signal?: AbortSignal,
  ): Promise<void> {
    controller = new AbortController()
    const combinedSignal = signal
      ? composeSignals(signal, controller.signal)
      : controller.signal

    isStreaming.value = true

    let response: Response
    try {
      const token = getAuthToken()
      response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Accept: 'text/event-stream',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify(body),
        signal: combinedSignal,
      })
    } catch (e) {
      const err = e as Error
      if (err?.name === 'AbortError') {
        isStreaming.value = false
        onComplete?.()
        return
      }
      isStreaming.value = false
      onError?.(err || new Error('网络连接失败'))
      return
    }

    if (!response.ok || !response.body) {
      let message = `请求失败（HTTP ${response.status}）`
      try {
        const errData = await response.json()
        if (errData?.message) message = errData.message
      } catch { /* ignore */ }
      isStreaming.value = false
      onError?.(new Error(message))
      return
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder('utf-8')
    let buffer = ''

    try {
      for (;;) {
        const { done, value } = await reader.read()
        if (done) break
        if (value) buffer += decoder.decode(value, { stream: true })
        // 按空行切分完整事件
        for (;;) {
          const m = /\r?\n\r?\n/.exec(buffer)
          if (!m) break
          const chunk = buffer.slice(0, m.index)
          buffer = buffer.slice(m.index + m[0].length)
          if (chunk.trim()) {
            const ev = parseSSEChunk(chunk)
            if (ev) onEvent(ev)
          }
        }
      }
      // flush 解码器残留
      buffer += decoder.decode()
      if (buffer.trim()) {
        const ev = parseSSEChunk(buffer)
        if (ev) onEvent(ev)
      }
    } catch (e) {
      const err = e as Error
      if (err?.name !== 'AbortError') {
        onError?.(err || new Error('事件流中断'))
      }
    } finally {
      try { reader.releaseLock() } catch { /* ignore */ }
      isStreaming.value = false
      controller = null
      onComplete?.()
    }
  }

  /** 中止当前流 */
  function abort(): void {
    if (controller) {
      controller.abort()
      controller = null
    }
    isStreaming.value = false
  }

  return {
    connect,
    abort,
    isStreaming,
  }
}

/**
 * 组合两个 AbortSignal：任一触发即中止。
 */
function composeSignals(a: AbortSignal, b: AbortSignal): AbortSignal {
  const combined = new AbortController()
  const onAbort = () => combined.abort()
  if (a.aborted || b.aborted) {
    combined.abort()
    return combined.signal
  }
  a.addEventListener('abort', onAbort, { once: true })
  b.addEventListener('abort', onAbort, { once: true })
  return combined.signal
}
