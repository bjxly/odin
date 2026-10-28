import { test, expect } from '@playwright/test'
import { apiLogin, injectToken, collectPageErrors } from './helpers'

/**
 * crud-smoke.spec —— 核心页面 CRUD 走查冒烟。
 *
 * 挑用户管理、角色管理两个核心页面：
 * - 列表表格渲染出数据行；
 * - 点击"新增"弹出对话框并可关闭（不实际提交，避免污染数据）。
 * 完整增改删链路由接口用例（api-test/cases）覆盖。
 */

test.beforeEach(async ({ page, request }) => {
  const tokens = await apiLogin(request)
  await injectToken(page, tokens)
})

async function crudWalk(page: any, path: string, pageName: string) {
  const errors = collectPageErrors(page)
  await page.goto(path, { waitUntil: 'domcontentloaded' })
  await page.waitForLoadState('networkidle')

  // 1. 表格渲染出行
  const rows = page.locator('.el-table__body tr')
  await expect(rows.first(), `${pageName}：表格无数据行`).toBeVisible({ timeout: 20_000 })

  // 2. 打开"新增"对话框
  const addBtn = page.getByRole('button', { name: /新增|添加/ }).first()
  if (await addBtn.isVisible().catch(() => false)) {
    await addBtn.click()
    const dialog = page.locator('.el-dialog:visible, .el-drawer:visible').first()
    await expect(dialog, `${pageName}：新增对话框未弹出`).toBeVisible({ timeout: 10_000 })
    // 3. 关闭对话框（点取消或关闭按钮）
    const cancelBtn = dialog.getByRole('button', { name: /取消|关闭/ }).first()
    if (await cancelBtn.isVisible().catch(() => false)) {
      await cancelBtn.click()
    } else {
      await page.keyboard.press('Escape')
    }
  }

  expect(errors, `${pageName}存在 JS 异常: ${errors.join('; ')}`).toHaveLength(0)
}

test('用户管理页面 CRUD 走查', async ({ page }) => {
  await crudWalk(page, '/sys/user', '用户管理')
})

test('角色管理页面 CRUD 走查', async ({ page }) => {
  await crudWalk(page, '/sys/role', '角色管理')
})
