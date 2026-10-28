import request from './request'
import type {
  OntologyRecord,
  SourceInfo,
} from '@/types/ontology'
import type { SourceSchema } from '@/types/mapping'

/** 后端 GET /workspace 返回的数据结构 */
export interface WorkspacePayload {
  activeOntologyId: string
  sources: SourceInfo[]
  ontologies: OntologyRecord[]
  schemas: SourceSchema[]
}

/** 加载完整工作区 */
export async function fetchWorkspace(): Promise<WorkspacePayload> {
  const res = await request.get<WorkspacePayload>('/workspace')
  return res.data
}

/** 保存完整工作区 */
export async function saveWorkspace(payload: WorkspacePayload): Promise<void> {
  await request.put('/workspace', payload)
}

/** 健康检查 */
export async function healthCheck(): Promise<{ status: string }> {
  const res = await request.get<{ status: string }>('/health')
  return res.data
}
