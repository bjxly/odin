<template>
  <el-dialog
    v-model="visible"
    :title="$t('platform.roleTitle', { tenant: tenantName })"
    width="860px"
    :draggable="true"
    :close-on-click-modal="false"
    destroy-on-close
  >
    <div class="platform-auth-tip">
      {{ $t('platform.roleTip') }}
    </div>
    <div class="platform-role-toolbar">
      <el-button
        type="primary"
        @click="handleAddRole"
      >
        {{ $t('common.add') }}
      </el-button>
      <el-button
        :loading="loading"
        @click="loadRoles"
      >
        {{ $t('platform.refresh') }}
      </el-button>
    </div>

    <el-table
      v-loading="loading"
      :data="roleList"
      border
      size="small"
      max-height="380"
    >
      <el-table-column
        prop="roleName"
        :label="$t('platform.roleName')"
        min-width="120"
      />
      <el-table-column
        prop="roleCode"
        :label="$t('platform.roleCode')"
        min-width="120"
      />
      <el-table-column
        prop="sortOrder"
        :label="$t('platform.sort')"
        width="70"
        align="right"
      />
      <el-table-column
        :label="$t('platform.status')"
        width="80"
      >
        <template #default="{ row }">
          <el-tag
            :type="row.status === 1 ? 'success' : 'danger'"
            size="small"
          >
            {{ row.status === 1 ? $t('common.enabled') : $t('common.disabled') }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column
        :label="$t('common.operation')"
        width="220"
        align="center"
      >
        <template #default="{ row }">
          <el-button
            link
            type="primary"
            @click="handleEditRole(row)"
          >
            {{ $t('common.edit') }}
          </el-button>
          <el-button
            link
            type="primary"
            @click="handleViewBoundUsers(row)"
          >
            {{ $t('platform.boundUsers') }}
          </el-button>
          <el-button
            link
            type="danger"
            @click="handleDeleteRole(row)"
          >
            {{ $t('common.delete') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 角色新增/编辑（含菜单权限勾选） -->
    <el-dialog
      v-model="roleDialogVisible"
      :title="roleIsEdit ? $t('platform.roleEditTitle') : $t('platform.roleAddTitle')"
      width="560px"
      :draggable="true"
      :close-on-click-modal="false"
      append-to-body
      destroy-on-close
    >
      <el-form
        ref="roleFormRef"
        :model="roleForm"
        :rules="roleRules"
        label-width="80px"
      >
        <el-form-item
          :label="$t('platform.roleName')"
          prop="roleName"
        >
          <el-input
            v-model="roleForm.roleName"
            :placeholder="$t('common.pleaseInput') + $t('platform.roleName')"
            :maxlength="50"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          :label="$t('platform.roleCode')"
          prop="roleCode"
        >
          <el-input
            v-model="roleForm.roleCode"
            :placeholder="$t('common.pleaseInput') + $t('platform.roleCode')"
            :maxlength="50"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="$t('platform.sort')">
          <el-input-number
            v-model="roleForm.sortOrder"
            :min="0"
          />
        </el-form-item>
        <el-form-item :label="$t('platform.status')">
          <el-radio-group v-model="roleForm.status">
            <el-radio :value="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :value="0">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('platform.menuPermission')">
          <div
            class="platform-menu-tree"
            style="width: 100%"
          >
            <el-tree
              ref="menuTreeRef"
              :data="menuTreeData"
              show-checkbox
              node-key="id"
              :props="{ label: 'menuName', children: 'children' }"
              default-expand-all
            />
          </div>
        </el-form-item>
        <el-form-item :label="$t('common.remark')">
          <el-input
            v-model="roleForm.remark"
            type="textarea"
            :rows="2"
            :maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleDialogVisible = false">{{ $t('common.cancelBtn') }}</el-button>
        <el-button
          type="primary"
          :loading="submitLoading"
          @click="handleRoleSubmit"
        >
          {{ $t('common.confirmBtn') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 角色绑定用户 -->
    <el-dialog
      v-model="boundVisible"
      :title="$t('platform.boundUsersTitle', { role: boundRoleName })"
      width="560px"
      :draggable="true"
      append-to-body
      destroy-on-close
    >
      <el-table
        v-loading="boundLoading"
        :data="boundUsers"
        border
        size="small"
        max-height="360"
      >
        <el-table-column
          prop="username"
          :label="$t('sysUser.username')"
          min-width="120"
        />
        <el-table-column
          prop="nickname"
          :label="$t('sysUser.nickname')"
          min-width="120"
        />
        <el-table-column
          prop="mobile"
          :label="$t('sysUser.mobile')"
          min-width="130"
        />
        <el-table-column
          :label="$t('platform.status')"
          width="80"
        >
          <template #default="{ row }">
            <el-tag
              :type="row.status === 1 ? 'success' : 'danger'"
              size="small"
            >
              {{ row.status === 1 ? $t('common.enabled') : $t('common.disabled') }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import type { ElTree } from 'element-plus'
import { getTenantRoles, createTenantRole, updateTenantRole, deleteTenantRole, getTenantRoleBoundUsers } from '@/api/sys/platform'
import { getMenuTree } from '@/api/sys/menu'
import type { RoleVO, RoleDTO, MenuVO, UserVO } from '@/types/sys'

const props = defineProps<{
  /** 弹窗开关（v-model） */
  modelValue: boolean
  /** 目标租户ID */
  tenantId?: string
  /** 目标租户名称（标题展示） */
  tenantName?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const { t } = useI18n()
const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const loading = ref(false)
const submitLoading = ref(false)
const roleList = ref<RoleVO[]>([])
const menuTreeData = ref<MenuVO[]>([])

const roleDialogVisible = ref(false)
const roleIsEdit = ref(false)
const roleFormRef = ref<FormInstance>()
const menuTreeRef = ref<InstanceType<typeof ElTree>>()

const roleForm = reactive<RoleDTO>({
  id: undefined,
  roleName: '',
  roleCode: '',
  sortOrder: 0,
  status: 1,
  remark: '',
  menuIds: [],
})

const roleRules = computed<FormRules>(() => ({
  roleName: [{ required: true, message: t('common.pleaseInput') + t('platform.roleName'), trigger: 'blur' }],
  roleCode: [{ required: true, message: t('common.pleaseInput') + t('platform.roleCode'), trigger: 'blur' }],
}))

const boundVisible = ref(false)
const boundLoading = ref(false)
const boundRoleName = ref('')
const boundUsers = ref<UserVO[]>([])

watch(visible, (open) => {
  if (open) loadRoles()
})

/** 加载目标租户角色列表。 */
async function loadRoles() {
  if (!props.tenantId) return
  loading.value = true
  try {
    roleList.value = await getTenantRoles(props.tenantId)
  } finally {
    loading.value = false
  }
}

/** 懒加载全局菜单树（权限勾选共用）。 */
async function loadMenuTree() {
  if (menuTreeData.value.length) return
  try {
    menuTreeData.value = await getMenuTree()
  } catch {
    /* ignore */
  }
}

function handleAddRole() {
  roleIsEdit.value = false
  Object.assign(roleForm, { id: undefined, roleName: '', roleCode: '', sortOrder: 0, status: 1, remark: '', menuIds: [] })
  roleDialogVisible.value = true
  loadMenuTree()
  nextTick(() => menuTreeRef.value?.setCheckedKeys([]))
}

function handleEditRole(row: RoleVO) {
  roleIsEdit.value = true
  Object.assign(roleForm, {
    id: row.id,
    roleName: row.roleName,
    roleCode: row.roleCode,
    sortOrder: row.sortOrder,
    status: row.status,
    remark: row.remark || '',
    menuIds: [...(row.menuIds || [])],
  })
  roleDialogVisible.value = true
  loadMenuTree().then(() => {
    nextTick(() => menuTreeRef.value?.setCheckedKeys(row.menuIds || []))
  })
}

async function handleRoleSubmit() {
  if (!roleFormRef.value || !props.tenantId) return
  const valid = await roleFormRef.value.validate().catch(() => false)
  if (!valid) return
  roleForm.menuIds = (menuTreeRef.value?.getCheckedKeys(false) as string[]) || []
  submitLoading.value = true
  try {
    if (roleIsEdit.value) {
      await updateTenantRole(props.tenantId, roleForm)
      ElMessage.success(t('common.updateSuccess'))
    } else {
      await createTenantRole(props.tenantId, roleForm)
      ElMessage.success(t('common.createSuccess'))
    }
    roleDialogVisible.value = false
    loadRoles()
  } finally {
    submitLoading.value = false
  }
}

async function handleDeleteRole(row: RoleVO) {
  if (!props.tenantId) return
  await ElMessageBox.confirm(t('platform.confirmDeleteRole'), t('common.tip'), { type: 'warning' })
  await deleteTenantRole(props.tenantId, row.id)
  ElMessage.success(t('common.deleteSuccess'))
  loadRoles()
}

/** 查看角色绑定用户。 */
async function handleViewBoundUsers(row: RoleVO) {
  if (!props.tenantId) return
  boundRoleName.value = row.roleName
  boundUsers.value = []
  boundVisible.value = true
  boundLoading.value = true
  try {
    boundUsers.value = await getTenantRoleBoundUsers(props.tenantId, row.id)
  } finally {
    boundLoading.value = false
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
.platform-role-toolbar {
  margin-bottom: 12px;
}
.platform-menu-tree {
  max-height: 320px;
  overflow-y: auto;
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  padding: 8px;
}
</style>
