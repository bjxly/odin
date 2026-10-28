import { get, post } from '../request'

/** 获取文件预签名URL */
export function getPresignUrl(fileId: number): Promise<string> {
  return get<string>(`/oss/presign/${fileId}`)
}

/** 上传文件 */
export function uploadFile(data: FormData): Promise<unknown> {
  return post('/oss/upload', data, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}
