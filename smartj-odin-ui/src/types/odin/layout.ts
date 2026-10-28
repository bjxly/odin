/** 节点坐标 */
export interface NodePosition {
  x: number
  y: number
}

/** 布局选项 */
export interface LayoutOptions {
  /** 布局方向：TB 纵向（自上而下），LR 横向（自左而右） */
  direction?: 'TB' | 'LR'
  /** 水平间距 */
  nodeSpacingX?: number
  /** 垂直间距 */
  layerSpacingY?: number
  /** 节点宽度 */
  nodeWidth?: number
  /** 节点高度 */
  nodeHeight?: number
  /** 画布宽度 */
  canvasWidth?: number
  /** 画布高度 */
  canvasHeight?: number
  /** 是否启用碰撞检测 */
  collisionDetection?: boolean
  /** 最大迭代轮次 */
  maxIterations?: number
}

/** 布局结果 */
export interface LayoutResult {
  positions: Map<string, NodePosition>
  layers: string[][]
  bounds: { minX: number; minY: number; maxX: number; maxY: number }
}
