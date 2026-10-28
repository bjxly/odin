import { ref } from 'vue'
import type { NodePosition, LayoutOptions, LayoutResult } from '@/types/odin'

/** 图谱节点（布局输入） */
export interface LayoutNode {
  id: string
  label?: string
}

/** 图谱边（布局输入） */
export interface LayoutEdge {
  source: string
  target: string
}

const DEFAULT_OPTIONS: Required<LayoutOptions> = {
  direction: 'TB',
  nodeSpacingX: 180,
  layerSpacingY: 120,
  nodeWidth: 140,
  nodeHeight: 60,
  canvasWidth: 1200,
  canvasHeight: 800,
  collisionDetection: true,
  maxIterations: 50,
}

/**
 * 本体图谱分层布局组合函数。
 * 算法：BFS 深度分层 → barycenter 排序 → 螺旋碰撞补位
 */
export function useOntologyLayout(options?: LayoutOptions) {
  const config = { ...DEFAULT_OPTIONS, ...options }
  const positions = ref<Map<string, NodePosition>>(new Map())

  /**
   * 计算布局
   * @param nodes - 图谱节点列表
   * @param edges - 图谱边列表
   * @param direction - 布局方向覆盖（TB 纵向 / LR 横向），缺省用配置值
   * @returns 布局结果
   */
  function layout(nodes: LayoutNode[], edges: LayoutEdge[], direction?: 'TB' | 'LR'): LayoutResult {
    if (nodes.length === 0) {
      positions.value = new Map()
      return { positions: positions.value, layers: [], bounds: { minX: 0, minY: 0, maxX: 0, maxY: 0 } }
    }

    // Step 1: BFS 深度分层
    const layers = bfsLayering(nodes, edges)

    // Step 2: Barycenter 排序（减少交叉）
    const sortedLayers = barycenterSort(layers, edges)

    // Step 3: 分配坐标
    const posMap = assignPositions(sortedLayers, direction ?? config.direction)

    // Step 4: 螺旋碰撞补位
    if (config.collisionDetection) {
      resolveCollisions(posMap)
    }

    positions.value = posMap

    // 计算边界
    let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity
    posMap.forEach((p) => {
      if (p.x < minX) minX = p.x
      if (p.y < minY) minY = p.y
      if (p.x > maxX) maxX = p.x
      if (p.y > maxY) maxY = p.y
    })

    return {
      positions: posMap,
      layers: sortedLayers,
      bounds: {
        minX: minX === Infinity ? 0 : minX,
        minY: minY === Infinity ? 0 : minY,
        maxX: maxX === -Infinity ? 0 : maxX,
        maxY: maxY === -Infinity ? 0 : maxY,
      },
    }
  }

  /** BFS 分层：从根节点（入度为0）开始逐层展开 */
  function bfsLayering(nodes: LayoutNode[], edges: LayoutEdge[]): string[][] {
    const nodeIds = new Set(nodes.map((n) => n.id))
    const inDegree = new Map<string, number>()
    const children = new Map<string, string[]>()

    nodeIds.forEach((id) => {
      inDegree.set(id, 0)
      children.set(id, [])
    })

    edges.forEach((e) => {
      if (nodeIds.has(e.source) && nodeIds.has(e.target)) {
        inDegree.set(e.target, (inDegree.get(e.target) ?? 0) + 1)
        children.get(e.source)!.push(e.target)
      }
    })

    // 找到入度为0的根节点
    const roots = [...nodeIds].filter((id) => (inDegree.get(id) ?? 0) === 0)
    if (roots.length === 0) {
      // 有环，取第一个节点作为根
      roots.push(nodes[0].id)
    }

    const layers: string[][] = []
    const visited = new Set<string>()
    let queue = roots

    while (queue.length > 0) {
      layers.push([...queue])
      queue.forEach((id) => visited.add(id))
      const nextQueue: string[] = []
      for (const id of queue) {
        for (const child of children.get(id) ?? []) {
          if (!visited.has(child)) {
            visited.add(child)
            nextQueue.push(child)
          }
        }
      }
      queue = nextQueue
    }

    // 未被访问到的孤立节点放入最后一层
    const unvisited = [...nodeIds].filter((id) => !visited.has(id))
    if (unvisited.length > 0) {
      layers.push(unvisited)
    }

    return layers
  }

  /** Barycenter 排序：按邻接层重心排序减少交叉 */
  function barycenterSort(layers: string[][], edges: LayoutEdge[]): string[][] {
    if (layers.length <= 1) return layers

    const result = layers.map((l) => [...l])

    for (let iter = 0; iter < 4; iter++) {
      // 正向扫描
      for (let i = 1; i < result.length; i++) {
        result[i] = sortByBarycenter(result[i], result[i - 1], edges, 'source')
      }
      // 反向扫描
      for (let i = result.length - 2; i >= 0; i--) {
        result[i] = sortByBarycenter(result[i], result[i + 1], edges, 'target')
      }
    }

    return result
  }

  function sortByBarycenter(
    layer: string[],
    referenceLayer: string[],
    edges: LayoutEdge[],
    fixedEnd: 'source' | 'target',
  ): string[] {
    const refIndex = new Map<string, number>()
    referenceLayer.forEach((id, idx) => refIndex.set(id, idx))

    const barycenters = new Map<string, number>()
    for (const nodeId of layer) {
      const connected = edges
        .filter((e) => (fixedEnd === 'source' ? e.target === nodeId : e.source === nodeId))
        .map((e) => refIndex.get(fixedEnd === 'source' ? e.source : e.target))
        .filter((v): v is number => v !== undefined)
      if (connected.length > 0) {
        barycenters.set(nodeId, connected.reduce((a, b) => a + b, 0) / connected.length)
      } else {
        barycenters.set(nodeId, layer.indexOf(nodeId))
      }
    }

    return [...layer].sort((a, b) => (barycenters.get(a) ?? 0) - (barycenters.get(b) ?? 0))
  }

  /** 为分层结果分配坐标 */
  function assignPositions(layers: string[][], direction: 'TB' | 'LR'): Map<string, NodePosition> {
    const posMap = new Map<string, NodePosition>()
    const { nodeSpacingX, layerSpacingY, canvasWidth, canvasHeight, nodeWidth, nodeHeight } = config

    if (direction === 'LR') {
      // 横向：层级自左而右展开，同层节点自上而下堆叠
      const layerGapX = nodeWidth + Math.max(80, layerSpacingY - 40)
      const nodeGapY = nodeHeight + 40
      layers.forEach((layer, layerIdx) => {
        const totalHeight = (layer.length - 1) * nodeGapY
        const startY = (canvasHeight - totalHeight) / 2
        layer.forEach((nodeId, nodeIdx) => {
          posMap.set(nodeId, {
            x: 80 + layerIdx * layerGapX,
            y: startY + nodeIdx * nodeGapY,
          })
        })
      })
      return posMap
    }

    // 纵向（默认）：层级自上而下展开，同层节点水平居中排布
    layers.forEach((layer, layerIdx) => {
      const totalWidth = (layer.length - 1) * nodeSpacingX
      const startX = (canvasWidth - totalWidth) / 2
      layer.forEach((nodeId, nodeIdx) => {
        posMap.set(nodeId, {
          x: startX + nodeIdx * nodeSpacingX,
          y: 80 + layerIdx * layerSpacingY,
        })
      })
    })

    return posMap
  }

  /** 螺旋碰撞检测与补位 */
  function resolveCollisions(posMap: Map<string, NodePosition>): void {
    const { nodeWidth, nodeHeight, maxIterations } = config
    const minDx = nodeWidth + 20
    const minDy = nodeHeight + 20
    const entries = [...posMap.entries()]

    for (let iter = 0; iter < maxIterations; iter++) {
      let moved = false
      for (let i = 0; i < entries.length; i++) {
        for (let j = i + 1; j < entries.length; j++) {
          const [, posA] = entries[i]
          const [idB, posB] = entries[j]
          const dx = Math.abs(posA.x - posB.x)
          const dy = Math.abs(posA.y - posB.y)
          if (dx < minDx && dy < minDy) {
            // 螺旋偏移 B
            const angle = iter * 0.5 + j * 0.3
            const radius = Math.max(minDx - dx, minDy - dy) / 2 + 10
            posB.x += Math.cos(angle) * radius
            posB.y += Math.sin(angle) * radius
            posMap.set(idB, posB)
            moved = true
          }
        }
      }
      if (!moved) break
    }
  }

  return {
    layout,
    positions,
  }
}
