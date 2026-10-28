<template>
  <!-- 租户切换器：仅当公开配置开关开启、租户数 > 1 且当前租户名可解析时渲染，否则不渲染任何 DOM -->
  <el-dropdown
    v-if="visible"
    class="ml-3 mr-1"
    trigger="click"
    @command="handleSwitch"
  >
    <el-button
      size="small"
      style="color: #ffffff; border-color: rgba(255, 255, 255, 0.3); background: transparent"
      :title="t('layout.currentTenant') + ': ' + currentTenantName"
    >
      {{ currentTenantName }}
      <el-icon class="ml-1"><ArrowDown /></el-icon>
    </el-button>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item
          v-for="tenant in tenants"
          :key="tenant.id"
          :command="tenant.id"
          :disabled="switching || tenant.id === currentTenantId"
        >
          {{ tenant.name }}
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { getSsoPublicConfig, getUserTenants } from '@/api/sso'
import type { TenantOptionVO } from '@/types/auth'

const { t } = useI18n()
const userStore = useUserStore()

/** 切换器功能开关（公开配置 ui.tenant.switcher.enabled，值为 "1" 时开启） */
const enabled = ref(false)
/** 当前用户可切换的租户列表 */
const tenants = ref<TenantOptionVO[]>([])
/** 切换中标记，防止重复触发 */
const switching = ref(false)

/** 当前租户 ID */
const currentTenantId = computed(() => userStore.userInfo?.tenantId ?? '')

/**
 * 当前租户名称
 * <p>
 * 优先取 userInfo.tenantName；缺失时按租户 ID 从租户列表反查；
 * 仍无法解析时返回空串（此时不渲染切换器）。
 */
const currentTenantName = computed(() => userStore.userInfo?.tenantName || tenants.value.find((item) => item.id === currentTenantId.value)?.name || '')

/** 显示条件：开关开启 且 租户数 > 1 且 当前租户名可解析，任一不满足则不渲染 */
const visible = computed(() => enabled.value && tenants.value.length > 1 && !!currentTenantName.value)

onMounted(async () => {
  try {
    // 复用 SSO 公开配置加载链路（sessionStorage 缓存命中时不发起额外请求）
    const cfg = await getSsoPublicConfig()
    enabled.value = cfg['ui.tenant.switcher.enabled'] === '1'
    if (!enabled.value) return
    // 开关开启时拉取可切换租户列表
    tenants.value = await getUserTenants()
  } catch {
    // 获取配置或租户列表失败时不渲染切换器
    enabled.value = false
  }
})

/** 切换到选中租户（成功后由 userStore 提示并整页刷新重建路由与菜单） */
async function handleSwitch(tenantId: string) {
  if (switching.value || tenantId === currentTenantId.value) return
  switching.value = true
  try {
    await userStore.switchTenant(tenantId)
  } catch {
    // 切换失败：恢复交互并提示（具体错误信息由请求拦截器统一弹出）
    switching.value = false
    ElMessage.error(t('layout.tenantSwitchFailed'))
  }
}
</script>
