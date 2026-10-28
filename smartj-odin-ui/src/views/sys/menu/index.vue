<template>
  <div class="sm-page">
    <!-- 检索区 -->
    <SmSearchForm
      :model="queryForm"
      @search="handleQuery"
      @reset="handleReset"
    >
      <el-form-item
        :label="$t('sysMenu.menuName')"
        label-width="auto"
      >
        <el-input
          v-model="queryForm.menuName"
          :placeholder="$t('common.pleaseInput') + $t('sysMenu.menuName')"
          clearable
        />
      </el-form-item>
      <el-form-item
        :label="$t('sysMenu.status')"
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
    <div class="sm-card">
      <div class="sm-table-header">
        <div class="sm-table-title">
          <span>{{ $t('sysMenu.list') }}</span>
        </div>
        <div class="sm-table-header-actions">
          <el-button-group>
            <el-button @click="toggleExpandAll">
              {{ expandAll ? $t('common.collapse') : $t('common.expand') }}
            </el-button>
            <el-button
              v-permission="'sys:menu:create'"
              type="primary"
              @click="handleAdd()"
            >
              {{ $t('common.add') }}
            </el-button>
          </el-button-group>
        </div>
      </div>

      <!-- 树形表格 -->
      <el-table
        v-if="refreshTable"
        v-loading="loading"
        :data="filteredData"
        row-key="id"
        border
        stripe
        :default-expand-all="expandAll"
        :tree-props="{ children: 'children' }"
      >
        <el-table-column
          prop="menuName"
          :label="$t('sysMenu.menuName')"
          min-width="200"
        />
        <el-table-column
          :label="$t('sysMenu.icon')"
          width="80"
          align="center"
        >
          <template #default="{ row }">
            <el-icon v-if="row.icon">
              <component :is="row.icon" />
            </el-icon>
          </template>
        </el-table-column>
        <el-table-column
          prop="sortOrder"
          :label="$t('sysMenu.sort')"
          width="80"
          align="center"
        />
        <el-table-column
          prop="path"
          :label="$t('sysMenu.path')"
        />
        <el-table-column
          prop="permission"
          :label="$t('sysMenu.permission')"
          width="180"
        />
        <el-table-column
          prop="component"
          :label="$t('sysMenu.component')"
        />
        <el-table-column
          :label="$t('sysMenu.menuType')"
          width="110"
          align="center"
        >
          <template #default="{ row }">
            <el-tag
              v-if="row.menuType === 1"
              type="primary"
              size="small"
            >
              {{ $t('sysMenu.dir') }}
            </el-tag>
            <el-tag
              v-else-if="row.menuType === 2"
              type="success"
              size="small"
            >
              {{ $t('sysMenu.page') }}
            </el-tag>
            <el-tag
              v-else-if="row.menuType === 3"
              type="warning"
              size="small"
            >
              {{ $t('sysMenu.button') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('sysMenu.visible')"
          width="80"
          align="center"
        >
          <template #default="{ row }">
            <el-tag
              v-if="row.menuType !== 3"
              :type="row.visible === 1 ? 'success' : 'info'"
              size="small"
            >
              {{ row.visible === 1 ? $t('sysMenu.show') : $t('sysMenu.hide') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="status"
          :label="$t('sysMenu.status')"
          width="100"
          align="center"
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
          fixed="right"
          align="center"
        >
          <template #default="{ row }">
            <el-button
              v-permission="'sys:menu:update'"
              link
              type="primary"
              @click="handleEdit(row)"
            >
              {{ $t('common.edit') }}
            </el-button>
            <el-button
              v-if="row.menuType !== 2"
              v-permission="'sys:menu:create'"
              link
              type="primary"
              @click="handleAdd(row.id)"
            >
              {{ $t('common.add') }}
            </el-button>
            <el-button
              v-permission="'sys:menu:delete'"
              link
              type="danger"
              @click="handleDelete(row)"
            >
              {{ $t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 新增/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? $t('sysMenu.editTitle') : $t('sysMenu.addTitle')"
      width="680px"
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
          :label="$t('sysMenu.parentId')"
          prop="parentId"
        >
          <el-tree-select
            v-model="editForm.parentId"
            :data="menuTreeOptions"
            :props="{ label: 'menuName', value: 'id', children: 'children' }"
            :placeholder="$t('common.pleaseSelect') + $t('sysMenu.parentId')"
            check-strictly
            filterable
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysMenu.menuType')"
          prop="menuType"
        >
          <el-radio-group v-model="editForm.menuType">
            <el-radio :value="1">{{ $t('sysMenu.dir') }}</el-radio>
            <el-radio :value="2">{{ $t('sysMenu.page') }}</el-radio>
            <el-radio :value="3">{{ $t('sysMenu.button') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item
          :label="$t('sysMenu.menuName')"
          prop="menuName"
        >
          <el-input
            v-model="editForm.menuName"
            :placeholder="$t('common.pleaseInput') + $t('sysMenu.menuName')"
            :maxlength="50"
            show-word-limit
            clearable
          />
        </el-form-item>
        <el-form-item
          v-if="editForm.menuType !== 3"
          :label="$t('sysMenu.icon')"
        >
          <el-input
            v-model="editForm.icon"
            :placeholder="$t('common.pleaseInput') + $t('sysMenu.icon')"
            :maxlength="100"
            show-word-limit
            clearable
          />
        </el-form-item>
        <el-form-item
          v-if="editForm.menuType !== 3"
          :label="$t('sysMenu.path')"
          prop="path"
        >
          <el-input
            v-model="editForm.path"
            :placeholder="$t('common.pleaseInput') + $t('sysMenu.path')"
            :maxlength="200"
            show-word-limit
            clearable
          />
        </el-form-item>
        <el-form-item
          v-if="editForm.menuType === 2"
          :label="$t('sysMenu.component')"
          prop="component"
        >
          <el-input
            v-model="editForm.component"
            :placeholder="$t('common.pleaseInput') + $t('sysMenu.component')"
            :maxlength="200"
            show-word-limit
            clearable
          />
        </el-form-item>
        <el-form-item
          v-if="editForm.menuType === 3"
          :label="$t('sysMenu.permission')"
          prop="permission"
        >
          <el-input
            v-model="editForm.permission"
            :placeholder="$t('common.pleaseInput') + $t('sysMenu.permission')"
            :maxlength="200"
            show-word-limit
            clearable
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysMenu.sort')"
          prop="sortOrder"
        >
          <el-input-number
            v-model="editForm.sortOrder"
            :min="0"
            :max="9999"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item
          v-if="editForm.menuType !== 3"
          :label="$t('sysMenu.visible')"
        >
          <el-radio-group v-model="editForm.visible">
            <el-radio :value="1">{{ $t('sysMenu.show') }}</el-radio>
            <el-radio :value="0">{{ $t('sysMenu.hide') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('sysMenu.status')">
          <el-radio-group v-model="editForm.status">
            <el-radio :value="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :value="0">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
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
import SmSearchForm from '@/components/SmSearchForm/index.vue'
import { getMenuTree, getMenuById, createMenu, updateMenu, deleteMenu } from '@/api/sys/menu'
import type { MenuVO, MenuDTO } from '@/types/sys'

const { t } = useI18n()

const loading = ref(false)
const submitLoading = ref(false)
const tableData = ref<MenuVO[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const expandAll = ref(true)
const refreshTable = ref(true)

const editFormRef = ref<FormInstance>()

const queryForm = reactive({
  menuName: '',
  status: undefined as number | undefined,
})

const editForm = reactive<MenuDTO>({
  id: undefined,
  parentId: '0',
  menuName: '',
  menuType: 1,
  path: '',
  component: '',
  permission: '',
  icon: '',
  sortOrder: 0,
  status: 1,
  visible: 1,
  keepAlive: 1,
})

const editRules = computed<FormRules>(() => ({
  menuName: [{ required: true, message: t('sysMenu.searchPlaceholder'), trigger: 'blur' }],
  menuType: [{ required: true, message: t('common.pleaseSelect') + t('sysMenu.menuType'), trigger: 'change' }],
}))

// 菜单树选项（包含根目录）
const menuTreeOptions = computed(() => {
  return [{ id: '0', menuName: t('sysMenu.root'), children: tableData.value }]
})

// 前端过滤树数据
const filteredData = ref<MenuVO[]>([])

function filterTree(nodes: MenuVO[]): MenuVO[] {
  const result: MenuVO[] = []
  for (const node of nodes) {
    const children = node.children ? filterTree(node.children) : []
    const nameMatch = !queryForm.menuName || node.menuName.includes(queryForm.menuName)
    const statusMatch = queryForm.status === undefined || node.status === queryForm.status
    if ((nameMatch && statusMatch) || children.length > 0) {
      result.push({ ...node, children })
    }
  }
  return result
}

async function loadData() {
  loading.value = true
  try {
    tableData.value = await getMenuTree()
    // 更新过滤树
    filteredData.value = tableData.value
  } finally {
    loading.value = false
  }
}

function handleQuery() {
  // 根据当前 queryForm 过滤数据
  filteredData.value = filterTree(tableData.value)
}

function handleReset() {
  queryForm.menuName = ''
  queryForm.status = undefined
  // 数据树变更为原始内容
  filteredData.value = tableData.value
}

function toggleExpandAll() {
  expandAll.value = !expandAll.value
  refreshTable.value = false
  nextTick(() => {
    refreshTable.value = true
  })
}

function handleAdd(parentId?: string) {
  isEdit.value = false
  Object.assign(editForm, {
    id: undefined,
    parentId: parentId || '0',
    menuName: '',
    menuType: '1',
    path: '',
    component: '',
    permission: '',
    icon: '',
    sortOrder: 0,
    status: 1,
    visible: 1,
    keepAlive: 1,
  })
  dialogVisible.value = true
}

async function handleEdit(row: MenuVO) {
  isEdit.value = true
  try {
    const detail = await getMenuById(row.id)
    Object.assign(editForm, { ...detail })
  } catch {
    Object.assign(editForm, { ...row })
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!editFormRef.value) return
  const valid = await editFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEdit.value) {
      await updateMenu(editForm)
      ElMessage.success(t('common.updateSuccess'))
    } else {
      await createMenu(editForm)
      ElMessage.success(t('common.createSuccess'))
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: MenuVO) {
  await ElMessageBox.confirm(t('sysMenu.confirmDelete'), t('common.tip'), { type: 'warning' })
  await deleteMenu(row.id)
  ElMessage.success(t('common.deleteSuccess'))
  loadData()
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.sm-crud-table {
  width: 100%;
}
.sm-table-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--sm-bg-card);
  border-radius: var(--sm-radius-md) var(--sm-radius-md) 0 0;
  margin-bottom: 0;
}
.sm-table-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--sm-text-primary);
}
.sm-table-header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
