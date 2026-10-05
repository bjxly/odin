<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  ElAside,
  ElAlert,
  ElButton,
  ElCard,
  ElCheckbox,
  ElDescriptions,
  ElDescriptionsItem,
  ElDialog,
  ElEmpty,
  ElForm,
  ElFormItem,
  ElInput,
  ElInputNumber,
  ElMain,
  ElMessage,
  ElOption,
  ElPopconfirm,
  ElSelect,
  ElTable,
  ElTableColumn,
  ElTag,
  ElContainer,
} from 'element-plus'
import { useOntologyStore } from '@/stores/ontology'
import { useMappingStore } from '@/stores/mapping'
import { useDataSourceStore, toApiSourceType } from '@/stores/datasource'
import { SOURCE_TYPES, type SourceInfo, type SourceType } from '@/types/ontology'

const onto = useOntologyStore()
const maps = useMappingStore()
const ds = useDataSourceStore()

const selectedId = ref<string | null>(null)
const search = ref('')
const testing = ref(false)
const introspecting = ref(false)
const saving = ref(false)
const testResult = ref<{ status: string; latencyMs?: number; testMessage: string } | null>(null)

const allSources = computed(() => ds.sourceInfos)

const filteredSources = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return allSources.value
  return allSources.value.filter((s) =>
    [s.id, s.code ?? '', s.label ?? '', s.type, s.database ?? '', s.host ?? '', ...(s.tags ?? [])]
      .join(' ')
      .toLowerCase()
      .includes(q),
  )
})

const current = computed(() => allSources.value.find((s) => s.id === selectedId.value) ?? null)

// 列表异步加载完成后自动选中首个数据源
watch(
  allSources,
  (list) => {
    if (!list.length) {
      selectedId.value = null
      return
    }
    if (!selectedId.value || !list.some((s) => s.id === selectedId.value)) {
      selectedId.value = list[0].id
    }
  },
  { immediate: true },
)

const currentTables = computed(() => {
  if (!current.value) return []
  return ds.getSchema(current.value.id)?.tables ?? []
})

const mappedMaps = computed(() => {
  if (!current.value) return []
  return maps.mapsOfSource(current.value.id)
})

const mappedClassIds = computed(() => [...new Set(mappedMaps.value.map((m) => m.classId))])

const statusTag = computed(() => {
  const s = current.value?.status
  if (s === 'ok') return { type: 'success' as const, text: '正常' }
  if (s === 'warn') return { type: 'warning' as const, text: '告警' }
  if (s === 'down') return { type: 'danger' as const, text: '不可用' }
  return { type: 'info' as const, text: '未测试' }
})

function formatTime(ts?: number) {
  if (!ts) return '—'
  const d = new Date(ts)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function connHint(s: SourceInfo) {
  if (!s.host) return '未配置主机'
  const port = s.port ? `:${s.port}` : ''
  const db = s.database ? `/${s.database}` : ''
  return `${s.type}://${s.host}${port}${db}`
}

function selectSource(id: string) {
  selectedId.value = id
  testResult.value = null
}

// ---- 新建 / 编辑 ----
const dlgOpen = ref(false)
const dlgMode = ref<'create' | 'edit'>('create')

const form = reactive({
  id: '',
  label: '',
  type: 'postgres' as SourceType,
  host: '',
  port: 5432,
  database: '',
  user: '',
  password: '',
  defaultSchema: 'public',
  readOnly: true,
  tagsText: '',
  description: '',
})

const DEFAULT_PORTS: Partial<Record<SourceType, number>> = {
  postgres: 5432,
  dameng: 5236,
  kingbase: 54321,
  opengauss: 5432,
  gaussdb: 5432,
  mysql: 3306,
  oracle: 1521,
  sqlite: 0,
  sqlserver: 1433,
  clickhouse: 9000,
}

function openCreate() {
  dlgMode.value = 'create'
  form.id = ''
  form.label = ''
  form.type = 'postgres'
  form.host = ''
  form.port = 5432
  form.database = ''
  form.user = ''
  form.password = ''
  form.defaultSchema = 'public'
  form.readOnly = true
  form.tagsText = ''
  form.description = ''
  dlgOpen.value = true
}

function openEdit() {
  if (!current.value) return
  const s = current.value
  dlgMode.value = 'edit'
  form.id = s.code ?? s.id
  form.label = s.label ?? ''
  form.type = s.type
  form.host = s.host ?? ''
  form.port = s.port ?? DEFAULT_PORTS[s.type] ?? 5432
  form.database = s.database ?? ''
  form.user = s.user ?? ''
  // 密码出于安全考虑不回显，留空表示不修改
  form.password = ''
  form.defaultSchema = s.defaultSchema ?? 'public'
  form.readOnly = s.readOnly ?? true
  form.tagsText = (s.tags ?? []).join(', ')
  form.description = s.description ?? ''
  dlgOpen.value = true
}

watch(
  () => form.type,
  (t) => {
    if (dlgMode.value === 'create' && !form.port) {
      form.port = DEFAULT_PORTS[t] ?? 5432
    }
  },
)

function parseTags(): string[] {
  return form.tagsText
    .split(/[,，]/)
    .map((s) => s.trim())
    .filter(Boolean)
}

/** 表单 → 后端 DataSourceVO 载荷 */
function buildPayload() {
  return {
    name: form.label || form.id,
    type: toApiSourceType(form.type),
    host: form.host || undefined,
    port: form.port || undefined,
    database_name: form.database || undefined,
    username: form.user || undefined,
    password: form.password || undefined,
    schema_name: form.defaultSchema || undefined,
    tags: parseTags().join(','),
    description: form.description || undefined,
  }
}

async function submitForm() {
  if (saving.value) return
  saving.value = true
  try {
    if (dlgMode.value === 'create') {
      const created = await ds.createDatasource({
        code: form.id || undefined,
        ...buildPayload(),
      })
      if (created?.id != null) selectedId.value = String(created.id)
      ElMessage.success('数据源已创建，建议先「测试连接」再「内省 Schema」')
    } else if (current.value) {
      await ds.updateDatasource(Number(current.value.id), buildPayload())
      ElMessage.success('数据源已更新')
    }
    dlgOpen.value = false
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : String(e))
  } finally {
    saving.value = false
  }
}

// ---- 测试 / 内省 / 删除 ----

async function runTest() {
  if (!current.value) return
  testing.value = true
  testResult.value = null
  try {
    const r = await ds.testDatasource(Number(current.value.id))
    testResult.value = { status: r.status, latencyMs: r.latencyMs, testMessage: r.testMessage }
    if (r.status === 'ok') ElMessage.success(r.testMessage)
    else if (r.status === 'warn') ElMessage.warning(r.testMessage)
    else ElMessage.error(r.testMessage)
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    testResult.value = { status: 'down', testMessage: msg }
    ElMessage.error(msg)
  } finally {
    testing.value = false
  }
}

async function runIntrospect() {
  if (!current.value) return
  if (current.value.status === 'down') {
    ElMessage.warning('源不可达，请先测试连接')
    return
  }
  introspecting.value = true
  try {
    const res = await ds.fetchSchema(Number(current.value.id))
    if (!res) {
      ElMessage.error('内省失败，未获取到表结构')
      return
    }
    const count = res.tables?.length ?? 0
    ElMessage.success(`内省完成：${count} 张表`)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : String(e))
  } finally {
    introspecting.value = false
  }
}

async function removeCurrent() {
  if (!current.value) return
  const id = current.value.id
  const label = current.value.code ?? current.value.label ?? id
  try {
    await ds.deleteDatasource(Number(id))
    await maps.fetchMappings({ ontology_id: onto.currentOntologyId ?? undefined })
    testResult.value = null
    ElMessage.success(`已删除数据源 ${label}，相关映射已清理`)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : String(e))
  }
}

onMounted(() => {
  void ds.fetchDatasources()
  if (onto.currentOntologyId != null && !maps.classMaps.length) {
    void maps.fetchMappings({ ontology_id: onto.currentOntologyId })
  }
})

const typeLabel: Record<string, string> = {
  postgres: 'PostgreSQL',
  dameng: '达梦 DM',
  kingbase: '人大金仓',
  opengauss: 'openGauss',
  gaussdb: 'GaussDB',
  mysql: 'MySQL',
  oracle: 'Oracle',
  sqlite: 'SQLite',
  sqlserver: 'SQL Server',
  clickhouse: 'ClickHouse',
}
</script>

<template>
  <div class="source-page">
    <ElContainer class="layout">
      <ElAside width="280px" class="list-panel">
        <div class="list-head">
          <h2>数据源</h2>
          <ElButton size="small" type="primary" @click="openCreate">新建</ElButton>
        </div>
        <ElInput
          v-model="search"
          size="small"
          placeholder="搜索 ID / 名称 / 类型 / 标签…"
          clearable
          class="search"
        />
        <ul v-loading="ds.loading" class="src-list">
          <li
            v-for="s in filteredSources"
            :key="s.id"
            :class="{ active: s.id === selectedId }"
            @click="selectSource(s.id)"
          >
            <div class="row1">
              <span class="dot" :class="s.status"></span>
              <span class="sid">{{ s.label || s.id }}</span>
              <ElTag size="small" effect="plain">{{ s.type }}</ElTag>
            </div>
            <div class="row2">
              <code>{{ s.code || s.id }}</code>
              <span class="meta">{{ s.tables ?? 0 }} 表</span>
              <span v-if="s.readOnly" class="ro">只读</span>
            </div>
          </li>
          <li v-if="!filteredSources.length" class="empty-item">无匹配数据源</li>
        </ul>
      </ElAside>

      <ElMain class="detail">
        <ElEmpty v-if="!current" description="请选择或新建数据源" />

        <template v-else>
          <div class="detail-head">
            <div>
              <h2>
                {{ current.label || current.id }}
                <ElTag :type="statusTag.type" size="small">{{ statusTag.text }}</ElTag>
                <ElTag size="small" effect="plain" type="info">{{ typeLabel[current.type] ?? current.type }}</ElTag>
              </h2>
              <p class="sub">
                <code>{{ connHint(current) }}</code>
                <span v-if="current.tags?.length" class="tags">
                  <ElTag v-for="t in current.tags" :key="t" size="small" type="info" effect="plain">
                    {{ t }}
                  </ElTag>
                </span>
              </p>
            </div>
            <div class="actions">
              <ElButton size="small" type="primary" :loading="testing" @click="runTest">
                测试连接
              </ElButton>
              <ElButton size="small" :loading="introspecting" @click="runIntrospect">
                内省 Schema
              </ElButton>
              <ElButton size="small" @click="openEdit">编辑</ElButton>
              <ElPopconfirm
                title="删除源将同时清理该源的全部类映射，确认？"
                @confirm="removeCurrent"
              >
                <template #reference>
                  <ElButton size="small" type="danger" plain>删除</ElButton>
                </template>
              </ElPopconfirm>
            </div>
          </div>

          <ElAlert
            v-if="testResult"
            :type="testResult.status === 'ok' ? 'success' : testResult.status === 'warn' ? 'warning' : 'error'"
            :title="testResult.testMessage"
            :closable="true"
            show-icon
            style="margin-bottom: 12px"
            @close="testResult = null"
          />

          <ElCard shadow="never" class="block">
            <template #header>连接信息</template>
            <ElDescriptions :column="3" size="small" border>
              <ElDescriptionsItem label="源 ID">
                <code>{{ current.code || current.id }}</code>
              </ElDescriptionsItem>
              <ElDescriptionsItem label="类型">{{ typeLabel[current.type] ?? current.type }}</ElDescriptionsItem>
              <ElDescriptionsItem label="状态">{{ statusTag.text }}</ElDescriptionsItem>
              <ElDescriptionsItem label="主机">{{ current.host || '—' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="端口">{{ current.port ?? '—' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="数据库">{{ current.database || '—' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="用户">{{ current.user || '—' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="默认 Schema">{{ current.defaultSchema || '—' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="访问模式">{{ current.readOnly ? '只读' : '读写' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="最近测试">{{ formatTime(current.lastTestAt) }}</ElDescriptionsItem>
              <ElDescriptionsItem label="最近内省">{{ formatTime(current.lastIntrospectAt) }}</ElDescriptionsItem>
              <ElDescriptionsItem label="延迟">
                {{ current.latencyMs != null ? `${current.latencyMs} ms` : '—' }}
              </ElDescriptionsItem>
              <ElDescriptionsItem label="说明" :span="3">
                {{ current.description || '—' }}
              </ElDescriptionsItem>
            </ElDescriptions>
          </ElCard>

          <ElCard shadow="never" class="block">
            <template #header>
              <div class="card-head">
                <span>物理表（{{ currentTables.length }}）</span>
                <span class="muted">
                  {{ current.lastIntrospectAt ? `内省于 ${formatTime(current.lastIntrospectAt)}` : '尚未内省' }}
                </span>
              </div>
            </template>
            <ElEmpty
              v-if="!currentTables.length"
              description="暂无表结构，点上方「内省 Schema」拉取"
              :image-size="60"
            />
            <ElTable v-else :data="currentTables" size="small" border max-height="320">
              <ElTableColumn prop="name" label="表名" min-width="140">
                <template #default="{ row }">
                  <code>{{ row.schema ? `${row.schema}.` : '' }}{{ row.name }}</code>
                </template>
              </ElTableColumn>
              <ElTableColumn prop="comment" label="注释" min-width="140" />
              <ElTableColumn label="列数" width="70" align="center">
                <template #default="{ row }">{{ row.columns.length }}</template>
              </ElTableColumn>
              <ElTableColumn label="主键" min-width="100">
                <template #default="{ row }">
                  <code v-if="row.columns.find((c: { isPk?: boolean }) => c.isPk)">
                    {{ row.columns.find((c: { isPk?: boolean }) => c.isPk)?.name }}
                  </code>
                  <span v-else class="muted">—</span>
                </template>
              </ElTableColumn>
            </ElTable>
          </ElCard>

          <ElCard shadow="never" class="block">
            <template #header>
              <div class="card-head">
                <span>已映射本体类（{{ mappedClassIds.length }}）</span>
                <span class="muted">共 {{ mappedMaps.length }} 条 ClassMap</span>
              </div>
            </template>
            <ElEmpty
              v-if="!mappedMaps.length"
              description="该源尚未映射任何本体类，可到「映射」工作台按源建映射"
              :image-size="60"
            />
            <ElTable v-else :data="mappedMaps" size="small" border>
              <ElTableColumn prop="classId" label="本体类" min-width="120">
                <template #default="{ row }">
                  {{ onto.classes.find((c) => c.id === row.classId)?.label ?? row.classId }}
                  <code class="cls-id">{{ row.classId }}</code>
                </template>
              </ElTableColumn>
              <ElTableColumn label="物理表" min-width="140">
                <template #default="{ row }">
                  <code>{{ row.schema ? `${row.schema}.` : '' }}{{ row.table }}</code>
                </template>
              </ElTableColumn>
              <ElTableColumn label="业务键" min-width="120">
                <template #default="{ row }">
                  <code>{{ row.keyColumns.join(', ') || '—' }}</code>
                </template>
              </ElTableColumn>
              <ElTableColumn label="属性映射" width="90" align="center">
                <template #default="{ row }">{{ row.properties.length }}</template>
              </ElTableColumn>
              <ElTableColumn label="主映射" width="80" align="center">
                <template #default="{ row }">
                  <ElTag v-if="row.primary" size="small" type="success">是</ElTag>
                  <span v-else class="muted">—</span>
                </template>
              </ElTableColumn>
            </ElTable>
          </ElCard>
        </template>
      </ElMain>
    </ElContainer>

    <ElDialog
      v-model="dlgOpen"
      :title="dlgMode === 'create' ? '新建数据源' : '编辑数据源'"
      width="560px"
      destroy-on-close
    >
      <ElForm label-width="96px" size="small">
        <ElFormItem label="源 ID" required>
          <ElInput
            v-model="form.id"
            :disabled="dlgMode === 'edit'"
            placeholder="如 erp_pg、crm_dm"
          />
        </ElFormItem>
        <ElFormItem label="显示名">
          <ElInput v-model="form.label" placeholder="如 ERP 主库" />
        </ElFormItem>
        <ElFormItem label="类型" required>
          <ElSelect v-model="form.type" style="width: 100%">
            <ElOption
              v-for="t in SOURCE_TYPES"
              :key="t"
              :label="typeLabel[t] ?? t"
              :value="t"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="主机">
          <ElInput v-model="form.host" placeholder="IP 或域名" />
        </ElFormItem>
        <ElFormItem label="端口">
          <ElInputNumber v-model="form.port" :min="0" :max="65535" />
        </ElFormItem>
        <ElFormItem label="数据库">
          <ElInput v-model="form.database" />
        </ElFormItem>
        <ElFormItem label="用户">
          <ElInput v-model="form.user" placeholder="数据库用户名" />
        </ElFormItem>
        <ElFormItem label="密码">
          <ElInput
            v-model="form.password"
            type="password"
            show-password
            :placeholder="dlgMode === 'edit' ? '留空则不修改' : '数据库密码'"
          />
        </ElFormItem>
        <ElFormItem label="默认 Schema">
          <ElInput v-model="form.defaultSchema" placeholder="public / CRM …" />
        </ElFormItem>
        <ElFormItem label="只读接入">
          <ElCheckbox v-model="form.readOnly" />
        </ElFormItem>
        <ElFormItem label="标签">
          <ElInput v-model="form.tagsText" placeholder="逗号分隔，如 prod, 信创" />
        </ElFormItem>
        <ElFormItem label="说明">
          <ElInput v-model="form.description" type="textarea" :rows="2" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dlgOpen = false">取消</ElButton>
        <ElButton type="primary" :loading="saving" @click="submitForm">
          {{ dlgMode === 'create' ? '创建' : '保存' }}
        </ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<style scoped>
.source-page {
  height: 100%;
  background: #f8fafc;
}

.layout {
  height: 100%;
}

.list-panel {
  background: #fff;
  border-right: 1px solid #e2e8f0;
  padding: 12px;
  overflow: auto;
}

.list-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.list-head h2 {
  margin: 0;
  font-size: 16px;
}

.search {
  margin-bottom: 10px;
}

.src-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.src-list li {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 8px 10px;
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}

.src-list li:hover {
  border-color: #93c5fd;
}

.src-list li.active {
  border-color: #4c8bf5;
  background: #eff6ff;
}

.src-list .row1 {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
}

.src-list .sid {
  font-weight: 600;
  font-size: 13px;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.src-list .row2 {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11px;
  color: #64748b;
}

.src-list .ro {
  color: #2563eb;
  background: #eff6ff;
  padding: 0 4px;
  border-radius: 3px;
}

.src-list .empty-item {
  border: none;
  color: #9ca3af;
  font-size: 12px;
  cursor: default;
  text-align: center;
  padding: 16px 0;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
  background: #cbd5e1;
}

.dot.ok {
  background: #22c55e;
}

.dot.warn {
  background: #f59e0b;
}

.dot.down {
  background: #ef4444;
}

.dot.unknown {
  background: #94a3b8;
}

.detail {
  padding: 16px 20px 32px;
  overflow: auto;
}

.detail-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 12px;
}

.detail-head h2 {
  margin: 0 0 4px;
  font-size: 18px;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.sub {
  margin: 0;
  font-size: 12px;
  color: #64748b;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.tags {
  display: inline-flex;
  gap: 4px;
}

.actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.block {
  margin-bottom: 12px;
}

.card-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.muted {
  color: #9ca3af;
  font-size: 12px;
}

.cls-id {
  margin-left: 6px;
  font-size: 11px;
  color: #94a3b8;
}
</style>
