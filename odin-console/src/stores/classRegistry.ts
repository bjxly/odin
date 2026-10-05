/**
 * 类 ID 解析注册表
 *
 * 背景：后端 `onto_class` 以数字主键标识类，而前端 `OntoClass.id` 使用类名
 * （relations 的 domain/range、layout 坐标、页面选择器全部以类名为键）。
 * mapping store 需要把 `MappingVO.class_id`（数字）还原为类名，但 mapping store
 * 不能直接 import ontology store（会形成 ontology → datasource → mapping → ontology 的环）。
 *
 * 因此由 ontology store 在初始化时注册解析函数，mapping store 只依赖本模块。
 */

export type ClassIdResolver = (numericId: number) => string | undefined

let resolver: ClassIdResolver | null = null

/** 由 ontology store 在 setup 时调用 */
export function registerClassIdResolver(fn: ClassIdResolver) {
  resolver = fn
}

/** 数字 class_id → 前端类名；未注册或未命中时回退为数字字符串 */
export function resolveClassName(numericId: number | undefined | null): string {
  if (numericId == null) return ''
  return resolver?.(numericId) ?? String(numericId)
}

/**
 * 映射关系提供者（反向依赖）
 *
 * ontology store 需要知道每个类被哪些数据源映射（`OntoClass.mappedSources`），
 * 但同样不能 import mapping store。由 mapping store 注册只读快照提供者。
 */
export interface ClassMapRef {
  classId: string
  sourceId: string
  primary?: boolean
}

export type ClassMapsProvider = () => ClassMapRef[]

let mapsProvider: ClassMapsProvider | null = null

/** 由 mapping store 在 setup 时调用 */
export function registerClassMapsProvider(fn: ClassMapsProvider) {
  mapsProvider = fn
}

/** 读取全部映射引用；未注册时返回空数组 */
export function getClassMapRefs(): ClassMapRef[] {
  try {
    return mapsProvider?.() ?? []
  } catch {
    return []
  }
}
