/** 连接器驱动信息 */
export interface DriverInfo {
  id: string
  driverName: string
  driverClass: string
  urlPattern: string
  defaultPort?: number
  dbType: string
  description?: string
  version?: string
  builtin?: boolean
}

/** 连接器 VO */
export interface ConnectorVO {
  id: string
  name: string
  code: string
  driverClass: string
  urlPattern: string
  dbType: string
  defaultPort?: number
  description?: string
  status: number
  remark?: string
  createTime: string
  updateTime?: string
}

/** 连接器 DTO */
export interface ConnectorDTO {
  id?: string
  name: string
  code: string
  driverClass: string
  urlPattern: string
  dbType: string
  defaultPort?: number
  description?: string
  status?: number
  remark?: string
}

/** 连接器查询条件 */
export interface ConnectorQuery {
  current: number
  size: number
  name?: string
  code?: string
  dbType?: string
  status?: number
}
