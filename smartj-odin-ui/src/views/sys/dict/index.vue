<template>
  <div class="sm-page">
    <div class="sm-split-layout">
      <!-- 左侧: 字典类型 -->
      <div class="sm-split-left sm-card dict-type-panel">
        <div class="sm-table-header">
          <span class="sm-table-title">{{ $t('sysDict.dictType') }}</span>
          <div class="sm-table-header-actions">
            <el-button
              v-permission="'sys:dict:create'"
              type="primary"
              link
              @click="handleAddType"
            >
              <el-icon><Plus /></el-icon>
            </el-button>
          </div>
        </div>
        <el-input
          v-model="typeKeyword"
          :placeholder="$t('common.pleaseInput') + $t('sysDict.dictType')"
          clearable
          class="type-search"
        />
        <div class="type-list">
          <div
            v-for="item in filteredTypeList"
            :key="item.id"
            class="type-item"
            :class="{ active: currentType?.id === item.id }"
            @click="selectType(item)"
          >
            <div class="type-item-content">
              <div class="type-name">{{ item.dictName }}</div>
              <div class="type-code">{{ item.dictType }}</div>
            </div>
            <div class="type-actions">
              <el-button
                v-permission="'sys:dict:update'"
                link
                size="small"
                @click.stop="handleEditType(item)"
              >
                <el-icon><Edit /></el-icon>
              </el-button>
              <el-button
                v-permission="'sys:dict:delete'"
                link
                size="small"
                type="danger"
                @click.stop="handleDeleteType(item)"
              >
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧: 字典数据 -->
      <div class="sm-split-right sm-card">
        <template v-if="currentType">
          <SmCrudTable
            v-model:page="dataQuery.current"
            v-model:page-size="dataQuery.size"
            :data="dataTableData"
            :columns="dataColumns"
            :total="dataTotal"
            :loading="dataLoading"
            :show-index="true"
            @page-change="loadDictData"
            @size-change="loadDictData"
          >
            <template #title>{{ currentType.dictName }} - {{ $t('sysDict.dictItem') }}</template>
            <template #header-actions>
              <el-button
                v-permission="'sys:dict:create'"
                type="primary"
                @click="handleAddData"
              >
                {{ $t('common.add') }}
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
                v-permission="'sys:dict:update'"
                link
                type="primary"
                @click="handleEditData(row)"
              >
                {{ $t('common.edit') }}
              </el-button>
              <el-button
                v-permission="'sys:dict:delete'"
                link
                type="danger"
                @click="handleDeleteData(row)"
              >
                {{ $t('common.delete') }}
              </el-button>
            </template>
          </SmCrudTable>
        </template>
        <div
          v-else
          class="empty-hint"
        >
          <el-empty :description="$t('sysDict.selectTypeHint')" />
        </div>
      </div>
    </div>

    <!-- 字典类型弹窗 -->
    <el-dialog
      v-model="typeDialogVisible"
      :title="isEditType ? $t('sysDict.editTitle') : $t('sysDict.addTitle')"
      width="480px"
      :draggable="true"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form
        ref="typeFormRef"
        :model="typeForm"
        :rules="typeRules"
        label-width="auto"
      >
        <el-form-item
          :label="$t('sysDict.dictName')"
          prop="dictName"
        >
          <el-input
            v-model="typeForm.dictName"
            :placeholder="$t('common.pleaseInput') + $t('sysDict.dictName')"
            :maxlength="100"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysDict.dictType')"
          prop="dictType"
        >
          <el-input
            v-model="typeForm.dictType"
            :placeholder="$t('common.pleaseInput') + $t('sysDict.dictType')"
            :disabled="isEditType"
            :maxlength="100"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="$t('sysDict.status')">
          <el-radio-group v-model="typeForm.status">
            <el-radio :value="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :value="0">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('sysDict.remark')">
          <el-input
            v-model="typeForm.remark"
            type="textarea"
            :rows="2"
            :maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="typeDialogVisible = false">{{ $t('common.cancelBtn') }}</el-button>
        <el-button
          type="primary"
          :loading="submitLoading"
          @click="handleTypeSubmit"
        >
          {{ $t('common.confirmBtn') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 字典数据弹窗 -->
    <el-dialog
      v-model="dataDialogVisible"
      :title="isEditData ? $t('sysDict.editDataTitle') : $t('sysDict.addDataTitle')"
      width="480px"
      :draggable="true"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form
        ref="dataFormRef"
        :model="dataForm"
        :rules="dataRules"
        label-width="auto"
      >
        <el-form-item
          :label="$t('sysDict.dictLabel')"
          prop="dictLabel"
        >
          <el-input
            v-model="dataForm.dictLabel"
            :placeholder="$t('common.pleaseInput') + $t('sysDict.dictLabel')"
            :maxlength="100"
            show-word-limit
          />
        </el-form-item>
        <el-form-item
          :label="$t('sysDict.dictValue')"
          prop="dictValue"
        >
          <el-input
            v-model="dataForm.dictValue"
            :placeholder="$t('common.pleaseInput') + $t('sysDict.dictValue')"
            :maxlength="100"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="$t('sysDict.sort')">
          <el-input-number
            v-model="dataForm.sortOrder"
            :min="0"
            :max="9999"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="$t('sysDict.status')">
          <el-radio-group v-model="dataForm.status">
            <el-radio :value="1">{{ $t('common.enabled') }}</el-radio>
            <el-radio :value="0">{{ $t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="$t('sysDict.remark')">
          <el-input
            v-model="dataForm.remark"
            type="textarea"
            :rows="2"
            :maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dataDialogVisible = false">{{ $t('common.cancelBtn') }}</el-button>
        <el-button
          type="primary"
          :loading="submitLoading"
          @click="handleDataSubmit"
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
import { Plus, Edit, Delete } from '@element-plus/icons-vue'
import SmCrudTable from '@/components/SmCrudTable/index.vue'
import type { TableColumn } from '@/components/SmCrudTable/index.vue'
import { getDictTypeList, createDictType, updateDictType, deleteDictTypes } from '@/api/sys/dict'
import { getDictDataPage, createDictData, updateDictData, deleteDictData } from '@/api/sys/dict'
import type { DictTypeVO, DictTypeDTO, DictDataVO, DictDataDTO } from '@/types/sys'

const { t } = useI18n()

const submitLoading = ref(false)
const dataLoading = ref(false)

// ==================== 字典类型 ====================
const typeList = ref<DictTypeVO[]>([])
const typeKeyword = ref('')
const currentType = ref<DictTypeVO | null>(null)
const typeDialogVisible = ref(false)
const isEditType = ref(false)
const typeFormRef = ref<FormInstance>()

const filteredTypeList = computed(() => {
  if (!typeKeyword.value) return typeList.value
  const kw = typeKeyword.value.toLowerCase()
  return typeList.value.filter((t) => t.dictName.toLowerCase().includes(kw) || t.dictType.toLowerCase().includes(kw))
})

const typeForm = reactive<DictTypeDTO>({
  id: undefined,
  dictName: '',
  dictType: '',
  status: 1,
  remark: '',
})

const typeRules = computed<FormRules>(() => ({
  dictName: [{ required: true, message: t('common.pleaseInput') + t('sysDict.dictName'), trigger: 'blur' }],
  dictType: [{ required: true, message: t('common.pleaseInput') + t('sysDict.dictType'), trigger: 'blur' }],
}))

async function loadTypeList() {
  try {
    typeList.value = await getDictTypeList()
  } catch {
    /* ignore */
  }
}

function selectType(item: DictTypeVO) {
  currentType.value = item
  dataQuery.current = 1
  dataQuery.dictType = item.dictType
  loadDictData()
}

function handleAddType() {
  isEditType.value = false
  Object.assign(typeForm, { id: undefined, dictName: '', dictType: '', status: 1, remark: '' })
  typeDialogVisible.value = true
}

function handleEditType(item: DictTypeVO) {
  isEditType.value = true
  Object.assign(typeForm, { ...item })
  typeDialogVisible.value = true
}

async function handleTypeSubmit() {
  if (!typeFormRef.value) return
  const valid = await typeFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEditType.value) {
      await updateDictType(typeForm)
      ElMessage.success(t('common.updateSuccess'))
    } else {
      await createDictType(typeForm)
      ElMessage.success(t('common.createSuccess'))
    }
    typeDialogVisible.value = false
    loadTypeList()
  } finally {
    submitLoading.value = false
  }
}

async function handleDeleteType(item: DictTypeVO) {
  await ElMessageBox.confirm(t('sysDict.confirmDelete'), t('common.tip'), { type: 'warning' })
  await deleteDictTypes([item.id])
  ElMessage.success(t('common.deleteSuccess'))
  if (currentType.value?.id === item.id) {
    currentType.value = null
  }
  loadTypeList()
}

// ==================== 字典数据 ====================
const dataTableData = ref<DictDataVO[]>([])
const dataTotal = ref(0)
const dataDialogVisible = ref(false)
const isEditData = ref(false)
const dataFormRef = ref<FormInstance>()

const dataQuery = reactive({
  current: 1,
  size: 10,
  dictType: '',
})

const dataColumns = computed<TableColumn[]>(() => [
  { prop: 'dictLabel', label: t('sysDict.dictLabel'), minWidth: 120, align: 'left' },
  { prop: 'dictValue', label: t('sysDict.dictValue'), minWidth: 120, align: 'left' },
  { prop: 'sortOrder', label: t('sysDict.sort'), width: 80, align: 'right' },
  { prop: 'status', label: t('sysDict.status'), width: 100, slot: 'status' },
])

const dataForm = reactive<DictDataDTO>({
  id: undefined,
  dictType: '',
  dictLabel: '',
  dictValue: '',
  sortOrder: 0,
  status: 1,
  remark: '',
})

const dataRules = computed<FormRules>(() => ({
  dictLabel: [{ required: true, message: t('common.pleaseInput') + t('sysDict.dictLabel'), trigger: 'blur' }],
  dictValue: [{ required: true, message: t('common.pleaseInput') + t('sysDict.dictValue'), trigger: 'blur' }],
}))

async function loadDictData() {
  if (!currentType.value) return
  dataLoading.value = true
  try {
    const res = await getDictDataPage(dataQuery)
    dataTableData.value = res.records
    dataTotal.value = res.total
  } finally {
    dataLoading.value = false
  }
}

function handleAddData() {
  if (!currentType.value) return
  isEditData.value = false
  Object.assign(dataForm, { id: undefined, dictType: currentType.value.dictType, dictLabel: '', dictValue: '', sortOrder: 0, status: 1, remark: '' })
  dataDialogVisible.value = true
}

function handleEditData(row: DictDataVO) {
  isEditData.value = true
  Object.assign(dataForm, { ...row })
  dataDialogVisible.value = true
}

async function handleDataSubmit() {
  if (!dataFormRef.value) return
  const valid = await dataFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEditData.value) {
      await updateDictData(dataForm)
      ElMessage.success(t('common.updateSuccess'))
    } else {
      await createDictData(dataForm)
      ElMessage.success(t('common.createSuccess'))
    }
    dataDialogVisible.value = false
    loadDictData()
  } finally {
    submitLoading.value = false
  }
}

async function handleDeleteData(row: DictDataVO) {
  await ElMessageBox.confirm(t('sysDict.confirmDeleteItem'), t('common.tip'), { type: 'warning' })
  await deleteDictData(row.id)
  ElMessage.success(t('common.deleteSuccess'))
  loadDictData()
}

onMounted(() => {
  loadTypeList()
})
</script>

<style scoped>
.dict-type-panel {
  flex: 0 0 320px !important;
  display: flex;
  flex-direction: column;
}

.type-search {
  margin-bottom: 12px;
}

.type-list {
  flex: 1;
  overflow-y: auto;
}

.type-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  border-radius: var(--sm-radius-md);
  cursor: pointer;
  transition: background 0.2s;
}

.type-item:hover {
  background: var(--sm-bg-hover);
}
.type-item.active {
  background: #ecf5ff;
  color: var(--sm-color-primary);
}

.type-item-content {
  flex: 1;
  min-width: 0;
}
.type-name {
  font-size: 14px;
  font-weight: 500;
}
.type-code {
  font-size: 12px;
  color: var(--sm-text-secondary);
  margin-top: 2px;
}
.type-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s;
}
.type-item:hover .type-actions {
  opacity: 1;
}

.empty-hint {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}
</style>
