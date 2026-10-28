<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  ElDrawer,
  ElEmpty,
  ElTag,
  ElDescriptions,
  ElDescriptionsItem,
  ElAlert,
  ElTable,
  ElTableColumn,
  ElButton,
  ElPopconfirm,
  ElDivider,
  ElMessage,
} from 'element-plus'
import { Edit, Delete, Plus } from '@element-plus/icons-vue'
import { useOntologyStore } from '@/stores/ontology'
import { useMappingStore } from '@/stores/mapping'
import { useDataSourceStore } from '@/stores/datasource'
import type { MappingVO } from '@/types/api'
import ClassFormDialog from './ClassFormDialog.vue'
import PropertyFormDialog from './PropertyFormDialog.vue'
import RelationFormDialog from './RelationFormDialog.vue'

const store = useOntologyStore()
const maps = useMappingStore()
const dsStore = useDataSourceStore()

const classDlg = ref(false)
const propDlg = ref(false)
const relDlg = ref(false)
const propMode = ref<'create' | 'edit'>('create')
const editingPropId = ref<string | null>(null)

const visible = computed({
  get: () => !!store.focusedClass,
  set: (v: boolean) => {
    if (!v) store.focus(null)
  },
})

const cls = computed(() => store.focusedClass)

const propsList = computed(() => cls.value?.properties ?? [])
const relations = computed(() => (cls.value ? store.relationsOf(cls.value.id) : []))
const rules = computed(() => (cls.value ? store.rulesOf(cls.value.id) : []))

// ---- 物理来源清单（跨数据源，只读汇总）----
interface SourceEntry {
  key: string
  sourceName: string
  sourceType: string
  table: string
  propCount: number
  props: string[]
}

const classSources = ref<MappingVO[]>([])
const classSourcesLoading = ref(false)

/** 解析后端返回的属性映射明细（数组优先，其次 JSON 字符串） */
function parseProps(m: MappingVO): { name: string; column: string }[] {
  let list: any[] = []
  if (Array.isArray(m.property_mappings) && m.property_mappings.length) {
    list = m.property_mappings
  } else {
    const raw = m.property_mappings_json || m.mappings_json
    if (raw) {
      try {
        const parsed = JSON.parse(raw)
        list = Array.isArray(parsed) ? parsed : (parsed?.property_mappings ?? [])
      } catch {
        list = []
      }
    }
  }
  return list
    .map((p: any) => ({
      name: String(p?.property_name ?? p?.propertyName ?? ''),
      column: String(p?.column_name ?? p?.columnName ?? ''),
    }))
    .filter((p) => p.name || p.column)
}

const sourceEntries = computed<SourceEntry[]>(() =>
  classSources.value.map((m, i) => {
    const sid = String(m.datasource_id ?? '')
    const info = sid ? dsStore.sourceInfo(sid) : undefined
    const list = parseProps(m)
    return {
      key: String(m.id ?? i),
      sourceName: m.datasource_name || info?.label || (sid ? `数据源#${sid}` : '未知数据源'),
      sourceType: String(m.datasource_type || info?.type || '').toUpperCase(),
      table: m.source_schema ? `${m.source_schema}.${m.source_table}` : m.source_table || '—',
      propCount: list.length || m.mapped_count || 0,
      props: list.map((p) => p.name || p.column).slice(0, 5),
    }
  }),
)

async function loadClassSources() {
  const id = cls.value?.numericId
  if (id == null) {
    classSources.value = []
    return
  }
  if (!dsStore.sources.length && !dsStore.loading) void dsStore.fetchDatasources()
  classSourcesLoading.value = true
  try {
    classSources.value = await maps.queryMappings({ class_id: id })
  } finally {
    classSourcesLoading.value = false
  }
}

watch(() => cls.value?.numericId, () => { void loadClassSources() }, { immediate: true })

/** 点击来源条目：仅做提示，不做深跳转 */
function hintSource(e: SourceEntry) {
  const name = cls.value?.label || cls.value?.id || '该类'
  const detail = e.props.length
    ? `，已映射：${e.props.join('、')}${e.propCount > e.props.length ? ' 等' : ''}`
    : ''
  ElMessage.info(
    `「${name}」的 ${e.propCount} 个属性来自 ${e.sourceName}.${e.table}${detail}。可在「映射工作台 → 查看映射」中调整。`,
  )
}

function relText(rel: {
  domain: string
  range: string
  cardinality: string
  label: string
}) {
  if (!cls.value) return rel.label
  const dir = rel.domain === cls.value.id ? '→' : '←'
  const other = rel.domain === cls.value.id ? rel.range : rel.domain
  return `${dir} ${other}（${rel.cardinality}）`
}

function openCreateProp() {
  propMode.value = 'create'
  editingPropId.value = null
  propDlg.value = true
}

function openEditProp(id: string) {
  propMode.value = 'edit'
  editingPropId.value = id
  propDlg.value = true
}

function removeProp(id: string) {
  if (!cls.value) return
  store.deleteProperty(cls.value.id, id)
  ElMessage.success('属性已删除')
}

function removeRel(id: string) {
  store.deleteRelation(id)
  ElMessage.success('关系已删除')
}

function removeClass() {
  if (!cls.value) return
  store.deleteClass(cls.value.id)
  ElMessage.success('类已删除（含关联关系）')
  store.focus(null)
}

</script>

<template>
  <ElDrawer
    v-model="visible"
    :title="cls ? `${cls.label}（${cls.id}）` : ''"
    size="520px"
    destroy-on-close
  >
    <template v-if="cls">
      <div class="toolbar">
        <ElButton size="small" :icon="Edit" @click="classDlg = true">编辑类</ElButton>
        <ElButton size="small" :icon="Plus" @click="openCreateProp">加属性</ElButton>
        <ElButton size="small" :icon="Plus" @click="relDlg = true">加关系</ElButton>
        <ElPopconfirm title="删除该类及其所有关系？" @confirm="removeClass">
          <template #reference>
            <ElButton size="small" type="danger" :icon="Delete">删除类</ElButton>
          </template>
        </ElPopconfirm>
      </div>

      <ElAlert
        v-if="!sourceEntries.length && !cls.abstract && !classSourcesLoading"
        type="warning"
        title="该类尚未建立物理映射"
        :closable="false"
        style="margin-bottom: 12px"
      />
      <ElAlert
        v-if="cls.abstract"
        type="info"
        title="规则虚拟类，查询时展开为父类条件"
        :closable="false"
        style="margin-bottom: 12px"
      />

      <ElDescriptions :column="1" border size="small">
        <ElDescriptionsItem label="ID">
          <code>{{ cls.id }}</code>
        </ElDescriptionsItem>
        <ElDescriptionsItem label="标签">{{ cls.label }}</ElDescriptionsItem>
        <ElDescriptionsItem v-if="cls.domain" label="领域">{{ cls.domain }}</ElDescriptionsItem>
        <ElDescriptionsItem v-if="cls.description" label="描述">
          {{ cls.description }}
        </ElDescriptionsItem>
        <ElDescriptionsItem v-if="cls.synonyms?.length" label="同义词">
          <ElTag v-for="s in cls.synonyms" :key="s" size="small" style="margin: 2px">
            {{ s }}
          </ElTag>
        </ElDescriptionsItem>
        <ElDescriptionsItem v-if="cls.subclassOf" label="父类">
          {{ cls.subclassOf }}
        </ElDescriptionsItem>
      </ElDescriptions>

      <div class="section-head">
        <h4 class="section-title">数据属性（{{ propsList.length }}）</h4>
        <ElButton size="small" text type="primary" @click="openCreateProp">添加</ElButton>
      </div>
      <ElTable :data="propsList" size="small" border>
        <ElTableColumn prop="label" label="属性" min-width="100" show-overflow-tooltip />
        <ElTableColumn prop="range" label="类型" width="80" />
        <ElTableColumn label="标记" width="88">
          <template #default="{ row }">
            <ElTag v-if="row.isKey" size="small" type="warning">键</ElTag>
            <ElTag v-if="row.sensitive" size="small" type="danger">敏</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="110" fixed="right">
          <template #default="{ row }">
            <div class="ops-cell">
              <ElButton size="small" text @click="openEditProp(row.id)">编辑</ElButton>
              <ElPopconfirm title="删除属性？" @confirm="removeProp(row.id)">
                <template #reference>
                  <ElButton size="small" text type="danger">删除</ElButton>
                </template>
              </ElPopconfirm>
            </div>
          </template>
        </ElTableColumn>
      </ElTable>

      <div class="section-head">
        <h4 class="section-title">关系（{{ relations.length }}）</h4>
        <ElButton size="small" text type="primary" @click="relDlg = true">添加</ElButton>
      </div>
      <ul class="rel-list">
        <li v-for="r in relations" :key="r.id">
          <div class="rel-row">
            <strong>{{ r.label }}</strong>
            <span>
              <ElButton size="small" text type="danger" @click="removeRel(r.id)">删</ElButton>
            </span>
          </div>
          <span class="rel-meta">
            {{ relText(r) }}
            <ElTag v-if="r.crossSource" size="small" type="warning">跨源</ElTag>
          </span>
        </li>
      </ul>

      <div class="section-head">
        <h4 class="section-title">物理来源（{{ sourceEntries.length }}）</h4>
        <ElButton
          size="small"
          text
          type="primary"
          :loading="classSourcesLoading"
          @click="loadClassSources"
        >
          刷新
        </ElButton>
      </div>
      <ul v-if="sourceEntries.length" v-loading="classSourcesLoading" class="src-entry-list">
        <li
          v-for="e in sourceEntries"
          :key="e.key"
          class="src-entry"
          title="点击查看该来源的映射概况"
          @click="hintSource(e)"
        >
          <span class="se-main">
            <span class="se-source">{{ e.sourceName }}</span>
            <span class="se-dot">.</span>
            <code class="se-table">{{ e.table }}</code>
          </span>
          <span class="se-side">
            <ElTag v-if="e.sourceType" size="small" effect="plain" type="info">
              {{ e.sourceType }}
            </ElTag>
            <ElTag size="small" effect="plain" type="success">{{ e.propCount }} 属性</ElTag>
          </span>
        </li>
      </ul>
      <div v-else class="no-mapping">
        <ElEmpty
          :description="classSourcesLoading ? '加载中…' : '该类暂无物理来源映射'"
          :image-size="48"
        />
        <p class="no-mapping-tip">
          可在「数据接入 → 映射配置」中选择数据源与物理表，为该类建立映射
        </p>
      </div>

      <template v-if="rules.length">
        <ElDivider style="margin: 16px 0 8px" />
        <h4 class="section-title">关联规则</h4>
        <ul class="rule-list">
          <li v-for="r in rules" :key="r.id">
            <strong>{{ r.label }}</strong>
            <code>{{ r.condition }}</code>
            <p>{{ r.description }}</p>
          </li>
        </ul>
      </template>
    </template>

    <ClassFormDialog v-model:open="classDlg" mode="edit" :class-id="cls?.id" />
    <PropertyFormDialog
      v-model:open="propDlg"
      :class-id="cls?.id ?? ''"
      :mode="propMode"
      :prop-id="editingPropId"
    />
    <RelationFormDialog v-model:open="relDlg" :domain-id="cls?.id" />
  </ElDrawer>
</template>

<style scoped>
.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 12px;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 16px 0 8px;
}

.section-title {
  margin: 0;
  font-size: 13px;
  color: #374151;
  border-left: 3px solid #4c8bf5;
  padding-left: 8px;
}

.rel-list,
.rule-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.rel-list li,
.rule-list li {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  padding: 8px 10px;
  font-size: 12px;
}

.rel-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.rel-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 4px;
  color: #64748b;
}

.rule-list code {
  display: inline-block;
  margin: 4px 0;
  background: #f1f5f9;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 11px;
}

.rule-list p {
  margin: 0;
  color: #6b7280;
}

.src-entry-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.src-entry {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 7px 10px;
  background: #f7fbff;
  border: 1px solid #bee3f8;
  border-left: 3px solid #3182ce;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.src-entry:hover {
  background: #ebf8ff;
  border-color: #90cdf4;
  border-left-color: #2b6cb0;
  box-shadow: 0 2px 8px rgba(43, 108, 176, 0.12);
}

.se-main {
  display: flex;
  align-items: baseline;
  gap: 2px;
  min-width: 0;
  overflow: hidden;
}

.se-source {
  font-size: 12px;
  font-weight: 600;
  color: #1a365d;
  white-space: nowrap;
}

.se-dot {
  color: #90cdf4;
  font-weight: 700;
}

.se-table {
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 11.5px;
  color: #2b6cb0;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  padding: 0 5px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.se-side {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.ops-cell {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 0;
  white-space: nowrap;
}

.ops-cell .el-button {
  padding-left: 6px;
  padding-right: 6px;
}

.no-mapping {
  text-align: center;
  padding: 8px 0 4px;
}

.no-mapping-tip {
  margin: 0;
  font-size: 11.5px;
  color: #7b9cc4;
  line-height: 1.6;
}

:deep(.el-drawer__header) {
  background: linear-gradient(135deg, #2b6cb0 0%, #3182ce 100%);
  color: white;
  margin-bottom: 0;
  padding: 20px 24px;
}

:deep(.el-drawer__title) {
  color: white;
  font-weight: 600;
}

:deep(.el-drawer__close-btn) {
  color: white;
}

:deep(.el-drawer__close-btn:hover) {
  color: rgba(255, 255, 255, 0.8);
}
</style>
