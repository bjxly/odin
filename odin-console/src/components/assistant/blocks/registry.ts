import type { Component } from 'vue'

/**
 * 渲染块注册表（设计文档 §8.6）。
 *
 * 助手消息由若干 Block 组成，每个 Block 按 `type` 分发到对应渲染器组件。
 * 新增块类型（如 pivot / markdown / file）只需调用 `registerBlockRenderer`
 * 注册一个组件，**无需改动 BlockRenderer 或消息流核心逻辑**。
 *
 * 约定：渲染器组件接收 `block: BlockVO` prop，并通过 `action` 事件向上冒泡交互
 * （澄清选项、打开溯源抽屉等），由页面统一处理。
 */
const registry = new Map<string, Component>()

function normalize(type: unknown): string {
  return String(type ?? '').trim().toLowerCase()
}

/** 注册（或覆盖）一个块渲染器 */
export function registerBlockRenderer(type: string, component: Component): void {
  const key = normalize(type)
  if (!key || !component) return
  registry.set(key, component)
}

/** 批量注册 */
export function registerBlockRenderers(map: Record<string, Component>): void {
  for (const [type, comp] of Object.entries(map)) registerBlockRenderer(type, comp)
}

/** 取渲染器；未注册返回 undefined（由调用方回落到 UnknownBlock） */
export function getBlockRenderer(type: unknown): Component | undefined {
  return registry.get(normalize(type))
}

export function hasBlockRenderer(type: unknown): boolean {
  return registry.has(normalize(type))
}

/** 已注册的块类型列表（按注册顺序） */
export function listBlockTypes(): string[] {
  return [...registry.keys()]
}

export function unregisterBlockRenderer(type: string): boolean {
  return registry.delete(normalize(type))
}
