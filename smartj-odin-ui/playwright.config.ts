import { defineConfig } from '@playwright/test'

/**
 * SmartJ smartj-ui E2E 冒烟测试配置。
 *
 * 前置：后端 bootstrap(:9000) 与前端 dev server(:3200) 均需由使用者
 * 自行启动，本配置不自动拉起任何服务；前端未启动时由 api-test 探测并在报告中说明。
 */
export default defineConfig({
  testDir: './tests/e2e',
  timeout: 60_000,
  retries: 0,
  workers: 1,
  reporter: [['list']],
  use: {
    baseURL: process.env.E2E_BASE_URL || 'http://localhost:3200',
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
    locale: 'zh-CN',
  },
})
