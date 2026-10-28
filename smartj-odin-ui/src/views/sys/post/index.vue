<template>
  <div class="sm-page">
    <!-- 检索区 -->
    <SmSearchForm
      :model="queryForm"
      @search="handleQuery"
      @reset="handleReset"
    >
      <el-form-item
        :label="$t('sysPost.postName')"
        label-width="auto"
      >
        <el-input
          v-model="queryForm.postName"
          :placeholder="$t('common.pleaseInput') + $t('sysPost.postName')"
          clearable
        />
      </el-form-item>
      <el-form-item
        :label="$t('sysPost.postCode')"
        label-width="auto"
      >
        <el-input
          v-model="queryForm.postCode"
          :placeholder="$t('common.pleaseInput') + $t('sysPost.postCode')"
          clearable
        />
      </el-form-item>
      <el-form-item
        :label="$t('sysPost.status')"
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
      <template #title>{{ $t('sysPost.list') }}</template>
      <template #header-actions>
        <el-button
          v-permission="'sys:post:create'"
          type="primary"
          @click="handleAdd"
        >
          {{ $t('common.add') }}
        </el-button>
        <el-button
          v-permission="'sys:post:delete'"
          type="danger"
          :disabled="!selectedIds.length"
          @click="handleBatchDelete"
        >
          {{ $t('common.batchDelete') }}
        </el-button>
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
          v-permission="'sys:post:update'"
          link
          type="primary"
          @click="handleEdit(row)"
        >
          {{ $t('common.edit') }}
        </el-button>
        <el-button
          v-permission="'sys:post:delete'"
          link
          type="danger"
          @click="handleDelete(row)"
        >
          {{ $t('common.delete') }}
        </el-button>
      </template>
    </SmCrudTable>

    <!-- 新增/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? $t('sysPost.editTitle') : $t('sysPost.addTitle')"
      width="560px"
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
          :label="$t('sysPost.postCode')"
          prop="postCode"
        >
          <el-input
            v-model="editForm.postCode"
            :placeholder="$t('common.pleaseInput') + $t('sysPost.postCode')"
            :maxlength="50"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysPost.postName')"
          prop="postName"
        >
          <el-input
            v-model="editForm.postName"
            :placeholder="$t('common.pleaseInput') + $t('sysPost.postName')"
            :maxlength="50"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="$t('sysPost.sort')">
          <el-input-number
            v-model="editForm.sortOrder"
            :min="0"
            :max="9999"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="$t('sysPost.status')">
          <el-radio-group v-model="editForm.status">
            <el-radio :value="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :value="0">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('sysPost.remark')">
          <el-input
            v-model="editForm.remark"
            type="textarea"
            :rows="2"
            :placeholder="$t('common.pleaseInput') + $t('sysPost.remark')"
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
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import SmSearchForm from '@/components/SmSearchForm/index.vue'
import SmCrudTable from '@/components/SmCrudTable/index.vue'
import type { TableColumn } from '@/components/SmCrudTable/index.vue'
import { getPostPage, getPostById, createPost, updatePost, deletePost } from '@/api/sys/post'
import type { PostVO, PostDTO } from '@/types/sys'

const { t } = useI18n()

const loading = ref(false)
const submitLoading = ref(false)
const tableData = ref<PostVO[]>([])
const total = ref(0)
const selectedIds = ref<string[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)

const editFormRef = ref<FormInstance>()

const queryForm = reactive({
  current: 1,
  size: 10,
  postName: '',
  postCode: '',
  status: undefined as number | undefined,
})

const columns = computed<TableColumn[]>(() => [
  { prop: 'postCode', label: t('sysPost.postCode'), minWidth: 120, align: 'left' },
  { prop: 'postName', label: t('sysPost.postName'), minWidth: 150, align: 'left' },
  { prop: 'sortOrder', label: t('sysPost.sort'), width: 80, align: 'right' },
  { prop: 'status', label: t('sysPost.status'), width: 100, slot: 'status' },
  { prop: 'createTime', label: t('sysPost.createTime'), width: 170 },
  { prop: 'remark', label: t('sysPost.remark'), minWidth: 200, align: 'left' },
])

const editForm = reactive<PostDTO>({
  id: undefined,
  postCode: '',
  postName: '',
  sortOrder: 0,
  status: 1,
  remark: '',
})

const editRules = computed<FormRules>(() => ({
  postCode: [{ required: true, message: t('common.pleaseInput') + t('sysPost.postCode'), trigger: 'blur' }],
  postName: [{ required: true, message: t('common.pleaseInput') + t('sysPost.postName'), trigger: 'blur' }],
}))

async function loadData() {
  loading.value = true
  try {
    const res = await getPostPage(queryForm)
    tableData.value = res.records
    total.value = res.total
  } finally {
    loading.value = false
  }
}

function handleQuery() {
  queryForm.current = 1
  loadData()
}

function handleReset() {
  queryForm.postName = ''
  queryForm.postCode = ''
  queryForm.status = undefined
  handleQuery()
}

function handleSelectionChange(rows: PostVO[]) {
  selectedIds.value = rows.map((r) => r.id)
}

function handleAdd() {
  isEdit.value = false
  Object.assign(editForm, { id: undefined, postCode: '', postName: '', sortOrder: 0, status: 1, remark: '' })
  dialogVisible.value = true
}

async function handleEdit(row: PostVO) {
  isEdit.value = true
  const detail = await getPostById(row.id)
  Object.assign(editForm, detail)
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!editFormRef.value) return
  const valid = await editFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEdit.value) {
      await updatePost(editForm)
      ElMessage.success(t('common.updateSuccess'))
    } else {
      await createPost(editForm)
      ElMessage.success(t('common.createSuccess'))
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: PostVO) {
  await ElMessageBox.confirm(t('sysPost.confirmDelete'), t('common.tip'), { type: 'warning' })
  await deletePost(row.id)
  ElMessage.success(t('common.deleteSuccess'))
  loadData()
}

async function handleBatchDelete() {
  await ElMessageBox.confirm(t('sysPost.confirmDeletePosts', { count: selectedIds.value.length }), t('common.tip'), { type: 'warning' })
  for (let id of selectedIds.value) {
    await deletePost(id)
  }
  ElMessage.success(t('common.deleteSuccess'))
  loadData()
}

onMounted(() => {
  loadData()
})
</script>
