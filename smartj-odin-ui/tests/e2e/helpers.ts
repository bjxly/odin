import { Page, APIRequestContext, expect } from '@playwright/test'

/**
 * E2E 测试辅助：绕过登录表单（图形验证码无法自动识别），
 * 直接调 bootstrap 登录接口拿 token，注入 localStorage 后进入系统。
 *
 * Token 存储约定（与 src/utils/constants.ts 保持一致）：
 * - smartj_remember_me = '1'  → token 存 localStorage
 * - smartj_access_token / smartj_refresh_token
 */

const API_BASE = process.env.E2E_API_BASE || 'http://localhost:9000/api'
const ADMIN_USER = process.env.E2E_USERNAME || 'admin'
const ADMIN_PASS = process.env.E2E_PASSWORD || 'admin123'

export interface LoginTokens {
  token: string
  refreshToken: string
}

/** 通过后端接口登录，获取 token */
export async function apiLogin(request: APIRequestContext): Promise<LoginTokens> {
  const res = await request.post(`${API_BASE}/auth/login`, {
    data: { username: ADMIN_USER, password: ADMIN_PASS },
  })
  expect(res.ok(), `登录接口 HTTP ${res.status()}`).toBeTruthy()
  const body = await res.json()
  expect(body.code, `登录业务失败: ${body.message}`).toBe(0)
  return { token: body.data.token, refreshToken: body.data.refreshToken }
}

/** 将 token 注入 localStorage（在页面任何脚本执行前生效） */
export async function injectToken(page: Page, tokens: LoginTokens): Promise<void> {
  await page.addInitScript(
    ([token, refreshToken]) => {
      localStorage.setItem('smartj_remember_me', '1')
      localStorage.setItem('smartj_access_token', token)
      localStorage.setItem('smartj_refresh_token', refreshToken)
    },
    [tokens.token, tokens.refreshToken],
  )
}

/** 收集页面运行时异常（pageerror），返回数组引用供断言 */
export function collectPageErrors(page: Page): string[] {
  const errors: string[] = []
  page.on('pageerror', (e) => errors.push(e.message))
  return errors
}

/** 从后端菜单接口取全部叶子路由路径 */
export async function fetchMenuPaths(request: APIRequestContext, token: string): Promise<string[]> {
  const res = await request.get(`${API_BASE}/sys/menu/me`, {
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!res.ok()) return []
  const body = await res.json()
  const paths: string[] = []
  const walk = (nodes: any[], parent: string) => {
    for (const n of nodes || []) {
      let p: string = n.path || n.routePath || ''
      if (p && !p.startsWith('/')) p = `${parent.replace(/\/$/, '')}/${p}`
      const children = n.children || []
      if (children.length > 0) {
        walk(children, p || parent)
      } else if (p && !p.includes('http')) {
        paths.push(p)
      }
    }
  }
  walk(body.data || [], '')
  return [...new Set(paths)]
}
