import { get, post, put, del } from '../request'
import type { PageResult } from '@/types'
import type { PostVO, PostDTO, PostQuery } from '@/types/sys'

export function getPostPage(params: PostQuery) {
  return get<PageResult<PostVO>>('/sys/post/page', { params })
}

export function getPostById(id: string) {
  return get<PostVO>(`/sys/post/${id}`)
}

export function createPost(data: PostDTO) {
  return post<string>('/sys/post', data)
}

export function updatePost(data: PostDTO) {
  return put<void>('/sys/post', data)
}

export function deletePost(id: string) {
  return del<void>(`/sys/post/${id}`)
}
