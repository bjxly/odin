// ==================== 通知日志 ====================

/** 通知发送记录 VO */
export interface NoticeLogVO {
  id: string
  templateCode: string
  channel: string
  receiver: string
  title: string
  content: string
  sendStatus: number
  errorMessage: string
  providerId: string
  providerReqId: string
  readStatus: number
  readTime: string
  userId: string
  tenantId: string
  createTime: string
}

/** 通知发送请求 */
export interface NoticeSendRequest {
  templateCode: string
  channel: string
  receiver: string
  userId: string
  variables: Record<string, unknown>
  tenantId: string
}

// ==================== 通知模板 ====================

/** 通知模板 VO */
export interface NoticeTemplateVO {
  id?: string
  templateCode: string
  templateName: string
  channel: string
  titleTemplate: string
  contentTemplate: string
  variables: string
  status: number
  tenantId: string
  remark: string
  delFlag: number
  createBy: string
  createTime: string
  updateBy: string
  updateTime: string
}

// ==================== IM 配置 ====================

/** IM 配置 VO */
export interface ImConfigVO {
  id?: string
  providerId: string
  robotName: string
  webhookUrl: string
  secret: string
  atMobiles: string
  scene: string
  status: number
  tenantId: string
  remark: string
  createBy: string
  createTime: string
  updateBy: string
  updateTime: string
}
