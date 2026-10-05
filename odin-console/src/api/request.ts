import axios, { type AxiosInstance, type AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'

/** 后端统一响应结构 */
export interface ApiResponse<T = unknown> {
  code: number
  data: T
  message: string
  success: boolean
  timestamp: string
  traceId: string
}

/** 业务错误 */
export class ApiError extends Error {
  code: number
  traceId: string
  constructor(code: number, message: string, traceId: string) {
    super(message)
    this.code = code
    this.traceId = traceId
  }
}

const instance: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
instance.interceptors.request.use(
  (config) => {
    // 可在此处添加鉴权头
    return config
  },
  (error) => Promise.reject(error),
)

// 响应拦截器：自动解包 data.data
instance.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const body = response.data
    if (body && typeof body === 'object' && 'code' in body) {
      if (body.code === 0) {
        // 将解包后的数据挂到 response.data，方便调用方直接使用
        ;(response as unknown as { data: unknown }).data = body.data
        return response
      }
      const err = new ApiError(body.code, body.message || '请求失败', body.traceId)
      ElMessage.error(err.message)
      return Promise.reject(err)
    }
    return response
  },
  (error) => {
    const msg = error?.response?.data?.message || error?.message || '网络错误'
    ElMessage.error(msg)
    return Promise.reject(error)
  },
)

export default instance
