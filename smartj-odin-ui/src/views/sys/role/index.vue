<template>
  <div class="sm-page">
    <!-- 检索区 -->
    <SmSearchForm
      :model="queryForm"
      @search="handleQuery"
      @reset="handleReset"
    >
      <el-form-item
        :label="$t('sysRole.roleName')"
        label-width="auto"
      >
        <el-input
          v-model="queryForm.roleName"
          :placeholder="$t('common.pleaseInput') + $t('sysRole.roleName')"
          clearable
        />
      </el-form-item>
      <el-form-item
        :label="$t('sysRole.roleCode')"
        label-width="auto"
      >
        <el-input
          v-model="queryForm.roleCode"
          :placeholder="$t('common.pleaseInput') + $t('sysRole.roleCode')"
          clearable
        />
      </el-form-item>
      <el-form-item
        :label="$t('sysRole.status')"
        label-width="auto"
      >
        <el-select
          v-model="queryForm.status"
          :placeholder="$t('common.pleaseSelect')"
          clearable
        >
          <el-option
            :label="$t('common.enabled')"
            :value="1"
          />
          <el-option
            :label="$t('common.disabled')"
            :value="0"
          />
        </el-select>
      </el-form-item>
    </SmSearchForm>

    <!-- 表格区 -->
    <SmCrudTable
      v-model:page="queryForm.current"
      v-model:page-size="queryForm.size"
      :data="tableData"
      :columns="columns"
      :total="total"
      :loading="loading"
      :show-selection="true"
      :show-index="true"
      @selection-change="handleSelectionChange"
      @page-change="loadData"
      @size-change="loadData"
    >
      <template #title>{{ $t('sysRole.list') }}</template>
      <template #header-actions>
        <el-button
          v-permission="'sys:role:create'"
          type="primary"
          @click="handleAdd"
        >
          {{ $t('common.add') }}
        </el-button>
        <el-button
          v-permission="'sys:role:delete'"
          type="danger"
          :disabled="!selectedIds.length"
          @click="handleBatchDelete"
        >
          {{ $t('common.batchDelete') }}
        </el-button>
      </template>
      <template #dataScope="{ row }">
        <span>{{ dataScopeMap[row.dataScope] || '-' }}</span>
      </template>
      <template #status="{ row }">
        <el-tag
          :type="row.status === 1 ? 'success' : 'danger'"
          size="small"
        >
          {{ row.status === 1 ? $t('common.enabled') : $t('common.disabled') }}
        </el-tag>
      </template>
      <template #operate="{ row }">
        <el-button
          v-permission="'sys:role:update'"
          link
          type="primary"
          @click="handleEdit(row)"
        >
          {{ $t('common.edit') }}
        </el-button>
        <el-button
          v-permission="'sys:role:delete'"
          link
          type="danger"
          :disabled="row.roleCode === 'super_admin'"
          @click="handleDelete(row)"
        >
          {{ $t('common.delete') }}
        </el-button>
      </template>
    </SmCrudTable>

    <!-- 新增/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? $t('sysRole.editTitle') : $t('sysRole.addTitle')"
      width="600px"
      :draggable="true"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form
        ref="editFormRef"
        :model="editForm"
        :rules="editRules"
        label-width="auto"
      >
        <el-form-item
          :label="$t('sysRole.roleName')"
          prop="roleName"
        >
          <el-input
            v-model="editForm.roleName"
            :placeholder="$t('common.pleaseInput') + $t('sysRole.roleName')"
            :maxlength="50"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysRole.roleCode')"
          prop="roleCode"
        >
          <el-input
            v-model="editForm.roleCode"
            :placeholder="$t('common.pleaseInput') + $t('sysRole.roleCode')"
            :disabled="isEdit"
            :maxlength="50"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="$t('sysRole.sortOrder')">
          <el-input-number
            v-model="editForm.sortOrder"
            :min="0"
            :max="9999"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="$t('sysRole.dataScope')">
          <el-select
            v-model="editForm.dataScope"
            style="width: 100%"
          >
            <el-option
              :label="$t('sysRole.scopeAll')"
              :value="1"
            />
            <el-option
              :label="$t('sysRole.scopeDept')"
              :value="2"
            />
            <el-option
              :label="$t('sysRole.scopeDeptAndChild')"
              :value="3"
            />
            <el-option
              :label="$t('sysRole.scopeSelf')"
              :value="4"
            />
            <el-option
              :label="$t('sysRole.scopeCustom')"
              :value="5"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('sysRole.status')">
          <el-radio-group v-model="editForm.status">
            <el-radio :value="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :value="0">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('sysRole.menuIds')">
          <div
            class="menu-tree-wrapper"
            style="width: 100%"
          >
            <el-tree
              ref="menuTreeRef"
              :data="menuTreeData"
              show-checkbox
              node-key="id"
              :props="{ label: 'menuName', children: 'children' }"
              default-expand-all
              :check-strictly="false"
            />
          </div>
        </el-form-item>
        <el-form-item :label="$t('sysRole.remark')">
          <el-input
            v-model="editForm.remark"
            type="textarea"
            :rows="2"
            :placeholder="$t('common.pleaseInput') + $t('sysRole.remark')"
            :maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ $t('common.cancelBtn') }}</el-button>
        <el-button
          type="primary"
          :loading="submitLoading"
          @click="handleSubmit"
        >
          {{ $t('common.confirmBtn') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import type { ElTree } from 'element-plus'
import SmSearchForm from '@/components/SmSearchForm/index.vue'
import SmCrudTable from '@/components/SmCrudTable/index.vue'
import type { TableColumn } from '@/components/SmCrudTable/index.vue'
import { getRolePage, createRole, updateRole, deleteRoles, getRoleById } from '@/api/sys/role'
import { getMenuTree } from '@/api/sys/menu'
import type { RoleVO, RoleDTO, MenuVO } from '@/types/sys'

const loading = ref(false)
const submitLoading = ref(false)
const { t } = useI18n()
const tableData = ref<RoleVO[]>([])
const total = ref(0)
const selectedIds = ref<string[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const menuTreeData = ref<MenuVO[]>([])

const editFormRef = ref<FormInstance>()
const menuTreeRef = ref<InstanceType<typeof ElTree>>()

const dataScopeMap: Record<number, string> = {
  1: t('sysRole.scopeAll'),
  2: t('sysRole.scopeDept'),
  3: t('sysRole.scopeDeptAndChild'),
  4: t('sysRole.scopeSelf'),
  5: t('sysRole.scopeCustom'),
}

const queryForm = reactive({
  current: 1,
  size: 10,
  roleName: '',
  roleCode: '',
  status: undefined as number | undefined,
})

const columns = computed<TableColumn[]>(() => [
  { prop: 'roleName', label: t('sysRole.roleName'), minWidth: 120, align: 'left' },
  { prop: 'roleCode', label: t('sysRole.roleCode'), minWidth: 120, align: 'left' },
  { prop: 'sortOrder', label: t('sysRole.sortOrder'), width: 80, align: 'right' },
  { prop: 'dataScope', label: t('sysRole.dataScope'), width: 110, slot: 'dataScope' },
  { prop: 'status', label: t('sysRole.status'), width: 100, slot: 'status' },
  { prop: 'createTime', label: t('sysRole.createTime'), width: 170 },
])

const editForm = reactive<RoleDTO>({
  id: undefined,
  roleName: '',
  roleCode: '',
  sortOrder: 0,
  dataScope: 1,
  status: 1,
  remark: '',
  menuIds: [],
})

const editRules = computed<FormRules>(() => ({
  roleName: [{ required: true, message: t('common.pleaseInput') + t('sysRole.roleName'), trigger: 'blur' }],
  roleCode: [{ required: true, message: t('common.pleaseInput') + t('sysRole.roleCode'), trigger: 'blur' }],
}))

async function loadData() {
  loading.value = true
  try {
    const res = await getRolePage(queryForm)
    tableData.value = res.records
    total.value = res.total
  } finally {
    loading.value = false
  }
}

async function loadMenuTree() {
  try {
    menuTreeData.value = await getMenuTree()
  } catch {
    /* ignore */
  }
}

function handleQuery() {
  queryForm.current = 1
  loadData()
}

function handleReset() {
  queryForm.roleName = ''
  queryForm.roleCode = ''
  queryForm.status = undefined
  handleQuery()
}

function handleSelectionChange(rows: RoleVO[]) {
  selectedIds.value = rows.map((r) => r.id)
}

function handleAdd() {
  isEdit.value = false
  Object.assign(editForm, { id: undefined, roleName: '', roleCode: '', sortOrder: 0, dataScope: 1, status: 1, remark: '', menuIds: [] })
  dialogVisible.value = true
  nextTick(() => menuTreeRef.value?.setCheckedKeys([]))
}

async function handleEdit(row: RoleVO) {
  isEdit.value = true
  const detail = await getRoleById(row.id)
  Object.assign(editForm, detail)
  dialogVisible.value = true
  nextTick(() => menuTreeRef.value?.setCheckedKeys(detail.menuIds || []))
}

async function handleSubmit() {
  if (!editFormRef.value) return
  const valid = await editFormRef.value.validate().catch(() => false)
  if (!valid) return

  editForm.menuIds = (menuTreeRef.value?.getCheckedKeys(false) as string[]) || []

  submitLoading.value = true
  try {
    if (isEdit.value) {
      await updateRole(editForm)
      ElMessage.success(t('common.updateSuccess'))
    } else {
      await createRole(editForm)
      ElMessage.success(t('common.createSuccess'))
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: RoleVO) {
  await ElMessageBox.confirm(t('sysRole.confirmDeleteRole'), t('common.tip'), { type: 'warning' })
  await deleteRoles(row.id)
  ElMessage.success(t('common.deleteSuccess'))
  loadData()
}

async function handleBatchDelete() {
  await ElMessageBox.confirm(t('sysRole.confirmDeleteRoles', { count: selectedIds.value.length }), t('common.tip'), { type: 'warning' })
  for (let id of selectedIds.value) {
    await deleteRoles(id)
  }
  ElMessage.success(t('common.deleteSuccess'))
  loadData()
}

onMounted(() => {
  loadData()
  loadMenuTree()
})
</script>

<style scoped>
.menu-tree-wrapper {
  max-height: 400px;
  overflow-y: auto;
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  padding: 8px;
  height: 300px;
}
</style>
