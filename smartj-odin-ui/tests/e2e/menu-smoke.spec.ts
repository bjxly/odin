import { test, expect } from '@playwright/test'
import { apiLogin, injectToken, fetchMenuPaths, collectPageErrors } from './helpers'

/**
 * menu-smoke.spec —— 菜单页面遍历冒烟。
 *
 * 登录后从菜单接口取全部叶子路由，逐个打开页面，断言：
 * - 无 JS 运行时异常（pageerror）；
 * - 页面非白屏（#app 有可见内容）；
 * - 未被重定向回登录页。
 */

test('遍历全部菜单页面无白屏无JS异常', async ({ page, request }) => {
  test.setTimeout(15 * 60_000) // 页面多，放宽整体超时

  const tokens = await apiLogin(request)
  const paths = await fetchMenuPaths(request, tokens.token)
  expect(paths.length, '菜单接口未返回任何路由').toBeGreaterThan(0)
  console.log(`共 ${paths.length} 个菜单页面待遍历`)

  await injectToken(page, tokens)
  const failures: string[] = []

  for (const p of paths) {
    const errors = collectPageErrors(page)
    try {
      await page.goto(p, { waitUntil: 'domcontentloaded', timeout: 30_000 })
      // 等待异步渲染（表格/表单首屏请求）
      await page.waitForTimeout(1500)
      if (page.url().includes('/login')) {
        failures.push(`${p} → 被重定向回登录页`)
        continue
      }
      const text = (await page.locator('#app').innerText()).trim()
      // canvas 类页面（如 Three.js 3D 场景）无文本但非白屏，按 canvas 存在判定
      const canvasCount = await page.locator('#app canvas').count()
      if (text.length === 0 && canvasCount === 0) failures.push(`${p} → 白屏（无可见内容）`)
      if (errors.length > 0) failures.push(`${p} → JS 异常: ${errors.slice(0, 2).join('; ')}`)
    } catch (e: any) {
      failures.push(`${p} → 打开失败: ${e.message?.slice(0, 120)}`)
    }
  }

  expect(failures, `以下页面冒烟失败:\n${failures.join('\n')}`).toHaveLength(0)
})
