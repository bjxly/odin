<template>
  <el-dialog
    v-model="visible"
    :title="$t('platform.authTitle')"
    width="720px"
    :draggable="true"
    :close-on-click-modal="false"
    destroy-on-close
  >
    <div class="platform-auth-tip">
      {{ $t('platform.authTip') }}
    </div>

    <!-- 授权行：租户 + 角色 + 默认 -->
    <div
      v-for="(row, index) in grantRows"
      :key="index"
      class="platform-auth-row"
    >
      <el-select
        v-model="row.tenantId"
        :placeholder="$t('common.pleaseSelect') + $t('platform.tenant')"
        filterable
        class="platform-auth-tenant"
      >
        <el-option
          v-for="t in tenantOptions"
          :key="t.id"
          :label="t.name"
          :value="t.id"
          :disabled="t.id !== row.tenantId && usedTenantIds.includes(t.id)"
        />
      </el-select>
      <el-select
        v-model="row.roleIds"
        multiple
        collapse-tags
        collapse-tags-tooltip
        :placeholder="$t('common.pleaseSelect') + $t('platform.role')"
        :disabled="!row.tenantId"
        :loading="roleLoadingMap[row.tenantId] || false"
        class="platform-auth-roles"
        @visible-change="(v: boolean) => v && row.tenantId && loadRoleOptions(row.tenantId)"
      >
        <el-option
          v-for="r in roleOptionsMap[row.tenantId] || []"
          :key="r.id"
          :label="r.roleName"
          :value="r.id"
        />
      </el-select>
      <el-radio
        :model-value="defaultTenantId"
        :value="row.tenantId"
        @change="defaultTenantId = row.tenantId || undefined"
      >
        {{ $t('platform.setDefault') }}
      </el-radio>
      <el-button
        link
        type="danger"
        @click="removeRow(index)"
      >
        {{ $t('common.delete') }}
      </el-button>
    </div>
    <el-button
      type="primary"
      link
      @click="grantRows.push({ tenantId: '', roleIds: [] })"
    >
      + {{ $t('platform.addTenantRow') }}
    </el-button>

    <template #footer>
      <el-button @click="visible = false">{{ $t('common.cancelBtn') }}</el-button>
      <el-button
        type="primary"
        :loading="submitLoading"
        @click="handleSubmit"
      >
        {{ $t('common.confirmBtn') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { getTenantPage } from '@/api/sys/tenant'
import { getPlatformUserById, saveUserAuthorization, getTenantRoles } from '@/api/sys/platform'
import type { TenantGrantDTO, RoleVO, TenantVO } from '@/types/sys'

const props = defineProps<{
  /** 弹窗开关（v-model） */
  modelValue: boolean
  /** 目标账号ID */
  userId?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'success'): void
}>()

const { t } = useI18n()
const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const submitLoading = ref(false)
/** 乐观锁基版本，回传给后端防覆盖 */
const version = ref(0)
const defaultTenantId = ref<string | undefined>(undefined)
/** 授权行（tenantId 为空的新行提交时被过滤） */
const grantRows = reactive<TenantGrantDTO[]>([])
const tenantOptions = ref<TenantVO[]>([])
/** 每个租户的可选项角色缓存（按打开时懒加载） */
const roleOptionsMap = reactive<Record<string, RoleVO[]>>({})
const roleLoadingMap = reactive<Record<string, boolean>>({})

/** 已被占用的租户ID（下拉中互斥） */
const usedTenantIds = computed(() => grantRows.map((r) => r.tenantId).filter(Boolean))

watch(visible, async (open) => {
  if (!open || !props.userId) return
  await loadAuth()
})

/** 移除授权行；若默认租户被移除则同步清空默认选择。 */
function removeRow(index: number) {
  const [row] = grantRows.splice(index, 1)
  if (row && defaultTenantId.value === row.tenantId) {
    defaultTenantId.value = undefined
  }
}

/** 拉取租户全集并回显账号当前授权。 */
async function loadAuth() {
  grantRows.splice(0, grantRows.length)
  Object.keys(roleOptionsMap).forEach((k) => delete roleOptionsMap[k])
  const [detail, tenantPage] = await Promise.all([
    getPlatformUserById(props.userId!),
    getTenantPage({ current: 1, size: 500 }),
  ])
  tenantOptions.value = tenantPage.records
  version.value = detail.version
  defaultTenantId.value = detail.defaultTenantId
  // 保留历史归属（含失效租户行），失效租户角色只读展示、提交时后端负责移除
  for (const grant of detail.tenants) {
    grantRows.push({ tenantId: grant.tenantId, roleIds: [...grant.roleIds] })
    if (grant.available) loadRoleOptions(grant.tenantId)
  }
  if (!grantRows.length) {
    grantRows.push({ tenantId: '', roleIds: [] })
  }
}

/** 懒加载指定租户的角色选项（供角色下拉回显名称与新勾选）。 */
async function loadRoleOptions(tenantId: string) {
  if (!tenantId || roleOptionsMap[tenantId]) return
  roleLoadingMap[tenantId] = true
  try {
    roleOptionsMap[tenantId] = await getTenantRoles(tenantId)
  } finally {
    roleLoadingMap[tenantId] = false
  }
}

async function handleSubmit() {
  const rows = grantRows.filter((r) => r.tenantId)
  if (new Set(rows.map((r) => r.tenantId)).size !== rows.length) {
    ElMessage.warning(t('platform.duplicateTenant'))
    return
  }
  if (defaultTenantId.value && !rows.some((r) => r.tenantId === defaultTenantId.value)) {
    ElMessage.warning(t('platform.defaultMustBeGranted'))
    return
  }
  submitLoading.value = true
  try {
    await saveUserAuthorization(props.userId!, {
      version: version.value,
      defaultTenantId: defaultTenantId.value,
      tenants: rows.map((r) => ({ tenantId: r.tenantId, roleIds: r.roleIds })),
    })
    ElMessage.success(t('platform.authSaveSuccess'))
    visible.value = false
    emit('success')
  } finally {
    submitLoading.value = false
  }
}
</script>

<style scoped>
.platform-auth-tip {
  margin-bottom: 12px;
  color: var(--sm-text-secondary, #909399);
  font-size: 12px;
  line-height: 1.5;
}
.platform-auth-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}
.platform-auth-tenant {
  width: 200px;
  flex-shrink: 0;
}
.platform-auth-roles {
  flex: 1;
  min-width: 0;
}
</style>
