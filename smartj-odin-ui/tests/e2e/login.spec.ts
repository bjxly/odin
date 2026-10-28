import { test, expect } from '@playwright/test'
import { apiLogin, injectToken, collectPageErrors } from './helpers'

/**
 * login.spec —— 登录流程冒烟。
 *
 * 1. 登录页正常渲染（表单、验证码控件出现，无 JS 运行时异常）；
 * 2. 接口登录 + token 注入后可进入系统首页（等价于登录成功后的会话状态）。
 * 图形验证码无法自动识别，表单提交流程由人工/接口用例覆盖。
 */

test('登录页正常渲染', async ({ page }) => {
  const errors = collectPageErrors(page)
  await page.goto('/login')
  // 用户名/密码输入框与提交按钮可见
  await expect(page.locator('input').first()).toBeVisible({ timeout: 15_000 })
  await expect(page.locator('input[type="password"]')).toBeVisible()
  await expect(page.locator('button').first()).toBeVisible()
  expect(errors, `登录页存在 JS 异常: ${errors.join('; ')}`).toHaveLength(0)
})

test('token 注入后可进入系统首页', async ({ page, request }) => {
  const tokens = await apiLogin(request)
  await injectToken(page, tokens)
  const errors = collectPageErrors(page)
  await page.goto('/')
  // 未被重定向回登录页，且主界面渲染出内容
  await page.waitForLoadState('networkidle')
  expect(page.url()).not.toContain('/login')
  const text = (await page.locator('#app').innerText()).trim()
  expect(text.length, '首页疑似白屏').toBeGreaterThan(0)
  expect(errors, `首页存在 JS 异常: ${errors.join('; ')}`).toHaveLength(0)
})
