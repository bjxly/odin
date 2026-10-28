import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { test } from 'node:test'
import { runInNewContext } from 'node:vm'
import ts from 'typescript'

const require = createRequire(import.meta.url)

// 在内存中转译真实 Store，只替换 API 和界面依赖，不启动服务器或写入用户数据。
function createStore(response) {
  const auth = { setToken() {}, setTenantId() {} }
  const permissions = { hasRoutes: true, clearPermission() {} }
  const modules = {
    pinia: { defineStore: (_name, setup) => setup },
    vue: require('vue'),
    'element-plus': { ElMessage: { success() {} } },
    '@/api/sso': { switchTenant: async () => response },
    '@/stores/auth': { useAuthStore: () => auth },
    '@/stores/permission': { usePermissionStore: () => permissions },
    '@/stores/tagsView': { useTagsViewStore: () => ({ removeAll() {} }) },
    '@/stores/keepAlive': { useKeepAliveStore: () => ({ clearCache() {} }) },
    '@/utils/constants': { SSO_PUBLIC_CONFIG_KEY: 'test' },
    '@/i18n': { default: { global: { t: (key) => key } } },
  }
  const source = readFileSync(new URL('../../src/stores/user.ts', import.meta.url), 'utf8')
  const { outputText } = ts.transpileModule(source, {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
  })
  const exports = {}
  runInNewContext(outputText, {
    exports,
    require: (name) => {
      assert.ok(name in modules, `未模拟的依赖：${name}`)
      return modules[name]
    },
    setTimeout() {},
  })
  return exports.useUserStore()
}

test('切换租户后同步角色、权限并保留雪花 ID 精度', async () => {
  const response = {
    token: 'access',
    refreshToken: 'refresh',
    user: {
      id: '9007199254740993',
      username: 'admin',
      tenantId: '9007199254740995',
      roles: ['tenant_reader'],
      permissions: ['order:read'],
    },
  }
  const store = createStore(response)
  await store.switchTenant(response.user.tenantId)
  assert.deepEqual(store.userInfo.value.roles, response.user.roles)
  assert.deepEqual(store.userInfo.value.permissions, response.user.permissions)
  assert.equal(store.userInfo.value.id, response.user.id)
  assert.equal(store.userInfo.value.tenantId, response.user.tenantId)
})
