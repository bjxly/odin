/** 本体摘要（轻量本体信息，用于切换器） */
export interface OntologySummary {
  id: string
  code: string
  name: string
  description?: string
  version?: string
  status: number
  classCount?: number
  relationCount?: number
  lastModified?: string
  /** 助手预置问题列表 */
  presetQuestions?: string[]
}

/** 工作台保存状态 */
export interface WorkspaceSaveState {
  lastSavedAt: string | null
  dirty: boolean
}
