/**
 * 规则条件树的前端编辑模型与 expression JSON（递归分组结构）互转工具。
 *
 * expression.condition 约定：
 * - 叶子：{ type, property, operator, value }（type 由操作符推断：threshold / value_match）
 * - 分组：{ logic: 'and'|'or'|'not', rules: (叶子|分组)[] }，单条件时折叠回叶子以兼容旧数据
 */

/** 叶子条件（编辑模型） */
export interface CondNode {
  id: number
  property: string
  operator: string
  /** 编辑态始终为字符串，序列化时再转数值 */
  value: string
}

/** 条件组（编辑模型，递归） */
export interface GroupNode {
  id: number
  logic: 'and' | 'or' | 'not'
  items: CondNode[]
  groups: GroupNode[]
}

let uid = 0

export function nextNodeId(): number {
  return ++uid
}

export function newCond(): CondNode {
  return { id: nextNodeId(), property: '', operator: 'eq', value: '' }
}

export function newGroup(logic: GroupNode['logic'] = 'and'): GroupNode {
  return { id: nextNodeId(), logic, items: [], groups: [] }
}

/** 依据操作符推断 condition.type（阈值比较 vs 值匹配） */
export function condTypeOf(op: string): string {
  return ['gt', 'gte', 'lt', 'lte'].includes(op) ? 'threshold' : 'value_match'
}

/** 数值型字符串转 number，其余保持字符串（与种子数据 amount>50000 的数值比较一致） */
export function coerceValue(v: string): string | number {
  if (v !== '' && !Number.isNaN(Number(v))) return Number(v)
  return v
}

/** 编辑模型 → expression JSON 字符串（完全无条件时返回 null = 不修改原表达式） */
export function serializeCondition(root: GroupNode, actionType: string, actionTarget: string): string | null {
  const flat = flattenGroup(root)
  if (flat == null) return null
  return JSON.stringify({ condition: flat, action: { type: actionType, target_class: actionTarget } })
}

/** 递归展平条件组；完全空时返回 null */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function flattenGroup(node: GroupNode): any | null {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const parts: any[] = []
  for (const c of node.items) {
    const prop = c.property.trim()
    if (!prop) continue
    parts.push({
      type: condTypeOf(c.operator),
      property: prop,
      operator: c.operator,
      value: coerceValue(c.value.trim()),
    })
  }
  for (const g of node.groups) {
    const sub = flattenGroup(g)
    if (sub != null) parts.push(sub)
  }
  if (parts.length === 0) return null
  if (parts.length === 1 && node.logic === 'and') return parts[0]
  return { logic: node.logic, rules: parts }
}

/** expression JSON 字符串 → 编辑模型根组（无法识别的节点跳过） */
export function parseCondition(exprJson?: string | null): GroupNode {
  const root = newGroup()
  if (!exprJson) return root
  let parsed: Record<string, any> | null = null
  try {
    parsed = JSON.parse(exprJson)
  } catch {
    return root
  }
  const cond = parsed?.condition
  if (cond) applyNode(root, cond)
  return root
}

/** expression JSON 字符串 → action.target_class（兼容驼峰） */
export function parseTargetClass(exprJson?: string | null): string {
  if (!exprJson) return ''
  try {
    const expr = JSON.parse(exprJson)
    return String(expr?.action?.target_class ?? expr?.action?.targetClass ?? '')
  } catch {
    return ''
  }
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function applyNode(group: GroupNode, node: any): void {
  if (!node || typeof node !== 'object') return
  if (Array.isArray(node.rules)) {
    // 分组节点：logic 缺省按 and 处理（兼容 { rules: [...] } 老写法）
    const logic = node.logic === 'or' || node.logic === 'not' ? node.logic : 'and'
    if (group.logic === 'and' && group.items.length === 0 && group.groups.length === 0) {
      // 根组首次填充时继承外层 logic，避免多包一层
      group.logic = logic
      node.rules.forEach((child: any) => applyNode(group, child))
    } else {
      const sub = newGroup(logic)
      node.rules.forEach((child: any) => applyNode(sub, child))
      group.groups.push(sub)
    }
  } else if (typeof node.property === 'string') {
    group.items.push({
      id: nextNodeId(),
      property: node.property,
      operator: typeof node.operator === 'string' ? node.operator : 'eq',
      value: node.value != null ? String(node.value) : '',
    })
  }
}
