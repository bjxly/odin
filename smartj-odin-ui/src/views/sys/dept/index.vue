<template>
  <div class="sm-page">
    <!-- 检索区 -->
    <SmSearchForm
      :model="queryForm"
      @search="loadData"
      @reset="handleReset"
    >
      <el-form-item
        :label="$t('sysDept.deptName')"
        label-width="auto"
      >
        <el-input
          v-model="queryForm.deptName"
          :placeholder="$t('common.pleaseInput') + $t('sysDept.deptName')"
          clearable
        />
      </el-form-item>
      <el-form-item
        :label="$t('sysDept.status')"
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
        <span class="sm-table-title">{{ $t('sysDept.list') }}</span>
        <div class="sm-table-header-actions">
          <el-button-group>
            <el-button @click="toggleExpandAll">{{ expandAll ? $t('common.collapse') : $t('common.expand') }}</el-button>
            <el-button
              v-permission="'sys:dept:create'"
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
        :data="tableData"
        row-key="id"
        border
        stripe
        :default-expand-all="expandAll"
        :tree-props="{ children: 'children' }"
      >
        <el-table-column
          prop="deptName"
          :label="$t('sysDept.deptName')"
          min-width="200"
        />
        <el-table-column
          prop="leader"
          :label="$t('sysDept.leader')"
          min-width="120"
        />
        <el-table-column
          prop="sortOrder"
          :label="$t('sysDept.sort')"
          width="80"
          align="right"
        />
        <el-table-column
          prop="status"
          :label="$t('sysDept.status')"
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
              v-permission="'sys:dept:create'"
              link
              type="primary"
              @click="handleAdd(row.id)"
            >
              {{ $t('common.add') }}
            </el-button>
            <el-button
              v-permission="'sys:dept:update'"
              link
              type="primary"
              @click="handleEdit(row)"
            >
              {{ $t('common.edit') }}
            </el-button>
            <el-button
              v-permission="'sys:dept:delete'"
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
      :title="isEdit ? $t('sysDept.editTitle') : $t('sysDept.addTitle')"
      width="520px"
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
          :label="$t('sysDept.parentId')"
          prop="parentId"
        >
          <el-tree-select
            v-model="editForm.parentId"
            :data="deptTreeOptions"
            :props="{ label: 'deptName', value: 'id', children: 'children' }"
            :placeholder="$t('common.pleaseSelect') + $t('sysDept.parentId')"
            check-strictly
            filterable
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysDept.deptName')"
          prop="deptName"
        >
          <el-input
            v-model="editForm.deptName"
            :placeholder="$t('common.pleaseInput') + $t('sysDept.deptName')"
            :maxlength="50"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="$t('sysDept.sort')">
          <el-input-number
            v-model="editForm.sortOrder"
            :min="0"
            :max="9999"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="$t('sysDept.leader')">
          <el-input
            v-model="editForm.leader"
            :placeholder="$t('common.pleaseInput') + $t('sysDept.leader')"
            :maxlength="50"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="$t('sysDept.phone')">
          <el-input
            v-model="editForm.phone"
            :placeholder="$t('common.pleaseInput') + $t('sysDept.phone')"
            :maxlength="50"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="$t('sysDept.email')">
          <el-input
            v-model="editForm.email"
            :placeholder="$t('common.pleaseInput') + $t('sysDept.email')"
            :maxlength="100"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="$t('sysDept.status')">
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
import { ref, reactive, onMounted, nextTick, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import SmSearchForm from '@/components/SmSearchForm/index.vue'
import { getDeptTree, createDept, updateDept, deleteDept } from '@/api/sys/dept'
import type { DeptVO, DeptDTO } from '@/types/sys'

const { t } = useI18n()

const loading = ref(false)
const submitLoading = ref(false)
const tableData = ref<DeptVO[]>([])
const deptTreeOptions = ref<DeptVO[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const expandAll = ref(true)
const refreshTable = ref(true)

const editFormRef = ref<FormInstance>()

const queryForm = reactive({
  deptName: '',
  status: undefined as number | undefined,
})

const editForm = reactive<DeptDTO>({
  id: undefined,
  parentId: '0',
  deptName: '',
  sortOrder: 0,
  leader: '',
  phone: '',
  email: '',
  status: 1,
})

const editRules = computed<FormRules>(() => ({
  deptName: [{ required: true, message: t('sysDept.deptSearch'), trigger: 'blur' }],
}))

async function loadData() {
  loading.value = true
  try {
    const data = await getDeptTree(queryForm)
    tableData.value = data
    deptTreeOptions.value = [{ id: '0', parentId: '', deptName: t('sysDept.root'), sortOrder: 0, status: 1, createTime: '', children: data }]
  } finally {
    loading.value = false
  }
}

function handleReset() {
  queryForm.deptName = ''
  queryForm.status = undefined
  loadData()
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
  Object.assign(editForm, { id: undefined, parentId: parentId || '0', deptName: '', sortOrder: 0, leader: '', phone: '', email: '', status: 1 })
  dialogVisible.value = true
}

function handleEdit(row: DeptVO) {
  isEdit.value = true
  Object.assign(editForm, { ...row })
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!editFormRef.value) return
  const valid = await editFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEdit.value) {
      await updateDept(editForm)
      ElMessage.success(t('common.updateSuccess'))
    } else {
      await createDept(editForm)
      ElMessage.success(t('common.createSuccess'))
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: DeptVO) {
  await ElMessageBox.confirm(t('sysDept.confirmDelete'), t('common.tip'), { type: 'warning' })
  await deleteDept(row.id)
  ElMessage.success(t('common.deleteSuccess'))
  loadData()
}

onMounted(() => {
  loadData()
})
</script>
