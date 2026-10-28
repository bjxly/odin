// ==================== PayOrder ====================

export interface PayOrderVO {
  id: string
  orderNo: string
  payNo: string
  channel: string
  payMethod: string
  subject: string
  totalAmount: number
  currency: string
  status: number
  channelTradeNo?: string
  payTime?: string
  expireTime?: string
  refundAmount: number
  createTime: string
  remark?: string
}

export interface PayOrderDTO {
  orderNo: string
  channel: string
  payMethod: string
  subject: string
  totalAmount: number
  currency?: string
  bizModule?: string
  bizId?: string
  attach?: string
  returnUrl?: string
}

export interface PayResultVO {
  payNo: string
  channel: string
  payMethod: string
  channelOrderNo?: string
  payParams?: string
  qrCodeUrl?: string
  payUrl?: string
}

export interface PayOrderQuery {
  current: number
  size: number
  channel?: string
  status?: number
  orderNo?: string
  payNo?: string
}

// ==================== PayRefund ====================

export interface PayRefundDTO {
  payNo: string
  refundAmount: number
  refundReason: string
}

export interface PayRefundVO {
  id: string
  refundNo: string
  payNo: string
  channelRefundNo?: string
  refundAmount: number
  refundReason: string
  status: number
  refundTime?: string
  createTime: string
}

// ==================== PayChannelConfig ====================

export interface PayChannelConfigVO {
  id: string
  channel: string
  configName: string
  appId: string
  mchId?: string
  notifyUrl?: string
  status: number
  createTime: string
  remark?: string
}

export interface PayChannelConfigDTO {
  id?: string
  channel: string
  configName: string
  appId: string
  mchId?: string
  apiKey?: string
  privateKey?: string
  publicKey?: string
  notifyUrl?: string
  status?: number
  remark?: string
}

export interface PayChannelConfigQuery {
  current: number
  size: number
  channel?: string
  status?: number
}
