<template>
  <div class="sm-page">
    <!-- 检索区 -->
    <SmSearchForm
      :model="queryForm"
      @search="handleQuery"
      @reset="handleReset"
    >
      <el-form-item
        :label="$t('sysCloudRoute.routeId')"
        label-width="auto"
      >
        <el-input
          v-model="queryForm.routeId"
          :placeholder="$t('common.pleaseInput') + $t('sysCloudRoute.routeId')"
          clearable
        />
      </el-form-item>
      <el-form-item
        :label="$t('sysCloudRoute.enabled')"
        label-width="auto"
      >
        <el-select
          v-model="queryForm.enabled"
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
      :show-index="true"
      @page-change="loadData"
      @size-change="loadData"
    >
      <template #title>{{ $t('sysCloudRoute.list') }}</template>
      <template #header-actions>
        <el-button
          v-permission="'sys:cloudRoute:create'"
          type="primary"
          @click="handleAdd"
        >
          {{ $t('common.add') }}
        </el-button>
      </template>
      <template #enabled="{ row }">
        <el-tag
          :type="row.enabled === 1 ? 'success' : 'danger'"
          size="small"
        >
          {{ row.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}
        </el-tag>
      </template>
      <template #operate="{ row }">
        <el-button
          v-permission="'sys:cloudRoute:update'"
          link
          type="primary"
          @click="handleEdit(row)"
        >
          {{ $t('common.edit') }}
        </el-button>
        <el-button
          v-permission="'sys:cloudRoute:delete'"
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
      :title="isEdit ? $t('sysCloudRoute.editTitle') : $t('sysCloudRoute.addTitle')"
      width="640px"
      :draggable="true"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form
        ref="editFormRef"
        :model="editForm"
        :rules="editRules"
        label-width="100px"
      >
        <el-form-item
          :label="$t('sysCloudRoute.routeId')"
          prop="routeId"
        >
          <el-input
            v-model="editForm.routeId"
            placeholder="如: sys-service"
            :maxlength="64"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysCloudRoute.uri')"
          prop="uri"
        >
          <el-input
            v-model="editForm.uri"
            placeholder="如: http://localhost:8082"
            :maxlength="256"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysCloudRoute.predicates')"
          prop="predicates"
        >
          <el-input
            v-model="editForm.predicates"
            type="textarea"
            :rows="3"
            placeholder='如: [{"name":"Path","args":{"pattern":"/api/sys/**"}}]'
            :maxlength="2000"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysCloudRoute.filters')"
          prop="filters"
        >
          <el-input
            v-model="editForm.filters"
            type="textarea"
            :rows="2"
            placeholder='如: [{"name":"StripPrefix","args":{"parts":"1"}}]'
            :maxlength="2000"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="$t('sysCloudRoute.orderNum')">
          <el-input-number
            v-model="editForm.orderNum"
            :min="0"
            :max="9999"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="$t('sysCloudRoute.enabled')">
          <el-radio-group v-model="editForm.enabled">
            <el-radio :value="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :value="0">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('sysCloudRoute.remark')">
          <el-input
            v-model="editForm.remark"
            type="textarea"
            :rows="2"
            :placeholder="$t('common.pleaseInput') + $t('sysCloudRoute.remark')"
            :maxlength="256"
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
import { getCloudRoutePage, getCloudRouteById, createCloudRoute, updateCloudRoute, deleteCloudRoute } from '@/api/sys/cloud-route'
import type { CloudRouteVO, CloudRouteDTO } from '@/types/sys'

const { t } = useI18n()

const loading = ref(false)
const submitLoading = ref(false)
const tableData = ref<CloudRouteVO[]>([])
const total = ref(0)
const dialogVisible = ref(false)
const isEdit = ref(false)

const editFormRef = ref<FormInstance>()

const queryForm = reactive({
  current: 1,
  size: 10,
  routeId: '',
  enabled: undefined as number | undefined,
})

const columns = computed<TableColumn[]>(() => [
  { prop: 'routeId', label: t('sysCloudRoute.routeId'), minWidth: 140, align: 'left' },
  { prop: 'uri', label: t('sysCloudRoute.uri'), minWidth: 200, align: 'left' },
  { prop: 'orderNum', label: t('sysCloudRoute.orderNum'), width: 80, align: 'right' },
  { prop: 'enabled', label: t('sysCloudRoute.status'), width: 100, slot: 'enabled' },
  { prop: 'predicates', label: t('sysCloudRoute.predicates'), minWidth: 200, align: 'left' },
  { prop: 'createTime', label: t('sysCloudRoute.createTime'), width: 170 },
  { prop: 'remark', label: t('sysCloudRoute.remark'), minWidth: 150, align: 'left' },
])

const editForm = reactive<CloudRouteDTO>({
  id: undefined,
  routeId: '',
  uri: '',
  predicates: '',
  filters: '',
  orderNum: 0,
  enabled: 1,
  remark: '',
})

const editRules = computed<FormRules>(() => ({
  routeId: [{ required: true, message: t('common.pleaseInput') + t('sysCloudRoute.routeId'), trigger: 'blur' }],
  uri: [{ required: true, message: t('common.pleaseInput') + t('sysCloudRoute.uri'), trigger: 'blur' }],
  predicates: [
    { required: true, message: t('common.pleaseInput') + t('sysCloudRoute.predicates'), trigger: 'blur' },
    {
      validator(_: any, value: string, callback: any) {
        // 校验字段是否为合法json,value已经不是空串
        try {
          const parsed = JSON.parse(value)
          // 必须是数组，且每个元素是普通对象（非 null、非数组）
          if (Array.isArray(parsed) && parsed.every((item) => item !== null && typeof item === 'object' && !Array.isArray(item))) {
            callback()
          } else {
            callback(new Error(t('common.invalidJson')))
          }
        } catch {
          callback(new Error(t('common.invalidJson')))
        }
      },
      trigger: 'blur',
    },
  ],
  filters: [
    {
      validator(_: any, value: string, callback: any) {
        // 空值跳过校验
        if (!value || value.trim() === '') {
          callback()
          return
        }
        try {
          const parsed = JSON.parse(value)
          if (Array.isArray(parsed) && parsed.every((item) => item !== null && typeof item === 'object' && !Array.isArray(item))) {
            callback()
          } else {
            callback(new Error(t('common.invalidJson')))
          }
        } catch {
          callback(new Error(t('common.invalidJson')))
        }
      },
      trigger: 'blur',
    },
  ],
}))

async function loadData() {
  loading.value = true
  try {
    const res = await getCloudRoutePage(queryForm)
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
  queryForm.routeId = ''
  queryForm.enabled = undefined
  handleQuery()
}

function handleAdd() {
  isEdit.value = false
  Object.assign(editForm, { id: undefined, routeId: '', uri: '', predicates: '', filters: '', orderNum: 0, enabled: 1, remark: '' })
  dialogVisible.value = true
}

async function handleEdit(row: CloudRouteVO) {
  isEdit.value = true
  const detail = await getCloudRouteById(row.id)
  Object.assign(editForm, detail)
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!editFormRef.value) return
  const valid = await editFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  // 处理json空字段值
  if (editForm.filters === '') {
    editForm.filters = undefined
  }
  try {
    if (isEdit.value) {
      await updateCloudRoute(editForm)
      ElMessage.success(t('common.updateSuccess'))
    } else {
      await createCloudRoute(editForm)
      ElMessage.success(t('common.createSuccess'))
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: CloudRouteVO) {
  await ElMessageBox.confirm(t('sysCloudRoute.confirmDelete'), t('common.tip'), { type: 'warning' })
  await deleteCloudRoute(row.id)
  ElMessage.success(t('common.deleteSuccess'))
  loadData()
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.sm-text-ellipsis {
  display: inline-block;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}
</style>
