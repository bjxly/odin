declare module '@smartx/dagflow' {
  import { DefineComponent } from 'vue'
  import { InjectionKey } from 'vue'

  // ==================== 类型定义 ====================

  export interface WorkflowData {
    id?: string
    name: string
    version?: number
    definition: string
    status?: 'DRAFT' | 'PUBLISHED' | 'ARCHIVED'
    spaceId?: string
    createBy?: string
    updateBy?: string
    createTime?: string
    updateTime?: string
  }

  export interface WorkflowListParams {
    current?: number
    size?: number
    name?: string
    status?: string
    spaceId?: string
  }

  export interface WorkflowListResponse {
    total: number
    records: WorkflowData[]
    current: number
    size: number
    pages: number
  }

  export interface WorkflowTestRequest {
    input: Record<string, unknown>
    stream?: boolean
    enablePersist?: boolean
    triggerBy?: string
  }

  export interface WorkflowTestResponse {
    success: boolean
    data?: Record<string, unknown>
    message?: string
    code?: number
  }

  export interface WorkflowChunk {
    type: 'node_start' | 'node_output' | 'node_complete' | 'log' | 'error' | 'done'
    nodeId?: string
    nodeName?: string
    nodeType?: string
    content?: string
    delta?: string
    data?: unknown
    inputs?: Record<string, unknown>
    outputs?: Record<string, unknown>
    timestamp?: number
  }

  export interface WorkflowStreamCallbacks {
    onMessage: (chunk: WorkflowChunk) => void
    onDone: (result: WorkflowTestResponse) => void
    onError: (error: Error) => void
    onNodeStart?: (nodeId: string, nodeName: string) => void
    onNodeComplete?: (nodeId: string, result: unknown) => void
  }

  export interface IWorkflowService {
    save(workflow: Omit<WorkflowData, 'id'>): Promise<{ id: string }>
    update(workflow: WorkflowData): Promise<void>
    get(id: string): Promise<WorkflowData>
    delete(ids: string[]): Promise<void>
    list(params: WorkflowListParams): Promise<WorkflowListResponse>
    getBySpaceId(spaceId: string): Promise<WorkflowData[]>
    test(workflowId: string, request: WorkflowTestRequest): Promise<WorkflowTestResponse>
    testStream(workflowId: string, request: WorkflowTestRequest, callbacks: WorkflowStreamCallbacks): AbortController
  }

  export interface DataSource {
    id: string
    name: string
    type: 'mysql' | 'postgresql' | 'sqlserver' | 'dm' | 'kingbase' | 'opengauss'
    host?: string
    port?: number
    database?: string
  }

  export interface ColumnInfo {
    name: string
    type: string
    comment?: string
    nullable: boolean
    primaryKey: boolean
  }

  export interface TableInfo {
    name: string
    comment?: string
    columns: ColumnInfo[]
  }

  export interface IDataSourceService {
    list(): Promise<DataSource[]>
    getTables(dataSourceId: string): Promise<TableInfo[]>
    getTableSchema(dataSourceId: string, tableName: string): Promise<TableInfo>
  }

  export interface FileUploadInfo {
    id: string
    originalName: string
    url: string
    size: number
    ext: string
    type: string
  }

  export interface IFileUploadService {
    upload(file: File): Promise<FileUploadInfo>
    readContent?(fileId: string): Promise<string>
  }

  export interface SmartxDagFlowOptions {
    services?: {
      workflow?: IWorkflowService
      dataSource?: IDataSourceService
      fileUpload?: IFileUploadService
    }
  }

  // ==================== 组件 ====================

  export const WorkflowDesigner: DefineComponent<{}, {}, any>

  // ==================== 注入 Key ====================

  export const WORKFLOW_SERVICE_KEY: InjectionKey<IWorkflowService>
  export const DATA_SOURCE_SERVICE_KEY: InjectionKey<IDataSourceService>
  export const FILE_UPLOAD_SERVICE_KEY: InjectionKey<IFileUploadService>

  // ==================== Hooks ====================

  export function useWorkflowService(): IWorkflowService | undefined
  export function useDataSourceService(): IDataSourceService | undefined
  export function useFileUploadService(): IFileUploadService | undefined

  // ==================== Store ====================

  export function useWorkflowStore(): any
  export function useNodeEditor(): any
  export function useVariableReference(): any

  // ==================== 插件 ====================

  export const SmartxDagFlowPlugin: {
    install(app: any, options?: SmartxDagFlowOptions): void
  }

  export default SmartxDagFlowPlugin
}

declare module '@smartx/dagflow/style.css' {
  // CSS 模块，无需导出
}
