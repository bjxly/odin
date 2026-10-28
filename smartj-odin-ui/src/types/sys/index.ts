// ==================== User ====================
export interface SocialAccountVO {
  id: string
  idpType: string
  openId: string
  nickname: string
  avatarUrl?: string
  createTime: string
}

export interface UserVO {
  id: string
  username: string
  nickname: string
  mobile?: string
  email?: string
  avatar?: string
  gender?: number
  userType?: number
  status: number
  deptId?: string
  deptName?: string
  roleIds?: string[]
  postIds?: string[]
  createTime: string
  remark?: string
  /** 账号来源: LOCAL/GITHUB/WECHAT/WECOM/DINGTALK/FEISHU 等 */
  source?: string
  /** 绑定的社交账号列表 */
  socialAccounts?: SocialAccountVO[]
}

export interface UserDTO {
  id?: string
  username: string
  nickname: string
  password?: string
  mobile?: string
  email?: string
  gender?: number
  status?: number
  deptId?: string
  roleIds?: string[]
  postIds?: string[]
  remark?: string
}

export interface UserQuery {
  current: number
  size: number
  username?: string
  nickname?: string
  mobile?: string
  status?: number
  deptId?: string
}

// ==================== Role ====================
export interface RoleVO {
  id: string
  roleName: string
  roleCode: string
  sortOrder: number
  dataScope?: number
  status: number
  remark?: string
  menuIds?: string[]
  createTime: string
}

export interface RoleDTO {
  id?: string
  roleName: string
  roleCode: string
  sortOrder?: number
  dataScope?: number
  status?: number
  remark?: string
  menuIds?: string[]
}

export interface RoleQuery {
  current: number
  size: number
  roleName?: string
  roleCode?: string
  status?: number
}

// ==================== Menu ====================
export interface MenuVO {
  id: string
  parentId: string
  menuName: string
  menuType: number // 1-目录 2-菜单 3-按钮
  path?: string
  component?: string
  permission?: string
  icon?: string
  sortOrder: number
  status: number
  visible?: number
  keepAlive?: number
  createTime: string
  children?: MenuVO[]
}

export interface MenuDTO {
  id?: string
  parentId?: string
  menuName: string
  menuType: number
  path?: string
  component?: string
  permission?: string
  icon?: string
  sortOrder?: number
  status?: number
  visible?: number
  keepAlive?: number
}

export interface MenuTreeVO {
  path: string
  component?: string
  name?: string
  redirect?: string
  meta?: {
    title: string
    icon?: string
    keepAlive?: boolean
    hidden?: boolean
    permission?: string
    affix?: boolean
  }
  children?: MenuTreeVO[]
}

// ==================== Dept ====================
export interface DeptVO {
  id: string
  parentId: string
  deptName: string
  sortOrder: number
  leader?: string
  phone?: string
  email?: string
  status: number
  createTime: string
  children?: DeptVO[]
}

export interface DeptDTO {
  id?: string
  parentId?: string
  deptName: string
  sortOrder?: number
  leader?: string
  phone?: string
  email?: string
  status?: number
}

export interface DeptQuery {
  deptName?: string
  status?: number
}

// ==================== DictType ====================
export interface DictTypeVO {
  id: string
  dictName: string
  dictType: string
  status: number
  remark?: string
  createTime: string
}

export interface DictTypeDTO {
  id?: string
  dictName: string
  dictType: string
  status?: number
  remark?: string
}

export interface DictTypeQuery {
  current: number
  size: number
  dictName?: string
  dictType?: string
  status?: number
}

// ==================== DictData ====================
export interface DictDataVO {
  id: string
  dictType: string
  dictLabel: string
  dictValue: string
  sort: number
  cssClass?: string
  listClass?: string
  status: number
  remark?: string
  createTime: string
}

export interface DictDataDTO {
  id?: string
  dictType: string
  dictLabel: string
  dictValue: string
  sortOrder?: number
  cssClass?: string
  listClass?: string
  status?: number
  remark?: string
}

export interface DictDataQuery {
  current: number
  size: number
  dictType: string
  dictLabel?: string
  status?: number
}

// ==================== Tenant ====================
export interface TenantVO {
  id: string
  name: string
  code: string
  packageId?: string
  expireTime?: string
  contact?: string
  phone?: string
  address?: string
  status: number
  remark?: string
  createTime: string
}

export interface TenantDTO {
  id?: string
  name: string
  code: string
  /** 新增必填，编辑时不重新复制角色。 */
  roleId?: string
  /** 新增时关联创建人并绑定租户角色，默认开启。 */
  associateCurrentUser?: boolean
  packageId?: string
  expireTime?: string
  contact?: string
  phone?: string
  address?: string
  status?: number
  remark?: string
}

export interface TenantQuery {
  current: number
  size: number
  name?: string
  code?: string
  status?: number
}

// ==================== Platform（全部用户 / 跨租户授权） ====================
/** 账号在单个租户内的授权展示对象 */
export interface TenantGrantVO {
  tenantId: string
  tenantName?: string
  /** 成员状态（0=停用, 1=启用） */
  status?: number
  /** 租户当前是否有效（有效才可勾选角色） */
  available: boolean
  /** 该租户内已绑定的角色ID */
  roleIds: string[]
}

/** 全局账号详情（含租户授权回显） */
export interface PlatformUserVO extends UserVO {
  defaultTenantId?: string
  defaultTenantName?: string
  /** 乐观锁版本，保存授权时必须回传 */
  version: number
  tenants: TenantGrantVO[]
}

/** 授权提交：单个租户归属及角色 */
export interface TenantGrantDTO {
  tenantId: string
  roleIds: string[]
}

/** 授权提交：一次保存全部归属、角色与默认租户 */
export interface UserTenantAuthorizationDTO {
  version: number
  defaultTenantId?: string
  tenants: TenantGrantDTO[]
}

// ==================== Config ====================
export interface ConfigVO {
  id: string
  configName: string
  configKey: string
  configValue: string
  configType: number
  remark?: string
  createTime: string
}

export interface ConfigDTO {
  id?: string
  configName: string
  configKey: string
  configValue: string
  configType?: number
  remark?: string
}

export interface ConfigQuery {
  current: number
  size: number
  configName?: string
  configKey?: string
  configType?: number
}

// ==================== SmsConfig ====================
export interface SysSmsConfigVO {
  id: string
  providerId: string
  accessKeyId: string
  hasSecret: boolean
  signName: string
  region: string
  appId: string
  sender: string
  templateConfig: string
  isDefault: number
  enabled: number
  remark: string
  createTime: string
  updateTime: string
}

export interface SysSmsConfigDTO {
  id?: string
  providerId: string
  accessKeyId?: string
  accessKeySecret?: string
  signName?: string
  region?: string
  appId?: string
  sender?: string
  templateConfig?: string
  isDefault?: number
  enabled?: number
  remark?: string
}

export interface SysSmsConfigQuery {
  current: number
  size: number
  providerId?: string
  enabled?: number
}

// ==================== Post ====================
export interface PostVO {
  id: string
  postCode: string
  postName: string
  sortOrder: number
  status: number
  remark?: string
  createTime: string
}

export interface PostDTO {
  id?: string
  postCode: string
  postName: string
  sortOrder?: number
  status?: number
  remark?: string
}

export interface PostQuery {
  current: number
  size: number
  postCode?: string
  postName?: string
  status?: number
}

// ==================== OperLog ====================
export interface OperLogVO {
  id: string
  module: string
  action: string
  method?: string
  requestUrl?: string
  requestMethod?: string
  responseCode?: number
  operatorName?: string
  operatorIp?: string
  duration?: number
  traceId?: string
  createTime: string
}

export interface OperLogQuery {
  current: number
  size: number
  module?: string
  action?: string
  operatorName?: string
  startTime?: string
  endTime?: string
}

// ==================== AuditLog ====================
export interface AuditLogVO {
  id: string
  target: string
  targetId?: string
  action: string
  beforeData?: string
  afterData?: string
  operatorName?: string
  traceId?: string
  createTime: string
}

export interface AuditLogQuery {
  current: number
  size: number
  target?: string
  action?: string
  operatorName?: string
  startTime?: string
  endTime?: string
}

// ==================== CloudRoute ====================
export interface CloudRouteVO {
  id: string
  routeId: string
  uri: string
  predicates: string
  filters?: string
  orderNum: number
  enabled: number
  createTime: string
  remark?: string
}

export interface CloudRouteDTO {
  id?: string
  routeId: string
  uri: string
  predicates: string
  filters?: string
  orderNum?: number
  enabled?: number
  remark?: string
}

export interface CloudRouteQuery {
  current: number
  size: number
  routeId?: string
  enabled?: number
}
