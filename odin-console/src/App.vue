<script setup lang="ts">
import {
  ElContainer,
  ElHeader,
  ElMain,
  ElAside,
  ElMenu,
  ElMenuItem,
  ElBadge,
  ElButton,
  ElPopconfirm,
  ElMessage,
  ElSelect,
  ElOption,
  ElDialog,
  ElForm,
  ElFormItem,
  ElInput,
  ElTag,
  ElIcon,
  ElLoading,
} from 'element-plus'
import { Share, Connection, DataLine, Coin, Search, HomeFilled, ChatDotRound, MagicStick } from '@element-plus/icons-vue'
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import TraceDrawer from '@/components/trace/TraceDrawer.vue'
import { useOntologyStore } from '@/stores/ontology'
import { useMappingStore } from '@/stores/mapping'
import { useDataSourceStore } from '@/stores/datasource'

const route = useRoute()
const router = useRouter()
const onto = useOntologyStore()
const maps = useMappingStore()
const ds = useDataSourceStore()
const appReady = ref(false)

// Route-based active menu index
const routeToMenuIndex: Record<string, string> = {
  '/': 'home',
  '/assistant': 'assistant',
  '/data-access': 'data-access',
  '/ontology': 'graph',
  '/connectors': 'connectors',
  '/reason': 'reason',
  '/query': 'query',
  '/sources': 'sources',
  '/mapping': 'mapping',
}

const active = computed(() => routeToMenuIndex[route.path] || 'home')

// Page navigation methods (for child component events)
const goToAssistant = () => {
  router.push('/assistant')
}

const goBackFromAssistant = () => {
  router.back()
}

const goToDataAccess = () => {
  router.push('/data-access')
}

const mappedClassIds = computed(() => new Set(maps.classMaps.map((m) => m.classId)))

const unmappedCount = computed(
  () => onto.classes.filter((c) => !mappedClassIds.value.has(c.id) && !c.abstract).length,
)

const anyDirty = computed(() => onto.dirty || onto.layoutDirty || maps.dirty)

/** 顶部「未保存」计数：变更日志条数（布局变更固定占 1 条）+ 映射未保存 */
const pendingCount = computed(() => {
  const logged = onto.changeLog.length
  const ontoOnly = onto.dirty && logged === 0 ? 1 : 0
  return logged + ontoOnly + (maps.dirty ? 1 : 0)
})

async function onSave() {
  try {
    if (maps.dirty) maps.saveMappings()
    if (onto.dirty || onto.layoutDirty) {
      // save() 会先将图谱坐标写回 ont_definition.layout_json
      const v = await onto.save()
      ElMessage.success(`已保存，本体版本 v${v}`)
    } else {
      ElMessage.success('已保存')
    }
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : String(e))
  }
}

async function onReset() {
  const loadingInstance = ElLoading.service({
    lock: true,
    text: '正在重新加载...',
    background: 'rgba(255, 255, 255, 0.8)',
  })
  try {
    await onto.fetchOntologies()
    if (onto.currentOntologyId != null) {
      await onto.fetchOntology(onto.currentOntologyId)
    }
    await ds.fetchDatasources()
    await maps.fetchMappings({ ontology_id: onto.currentOntologyId ?? undefined })
    onto.dirty = false
    onto.layoutDirty = false
    onto.changeLog = []
    maps.dirty = false
    ElMessage.info('已放弃未保存修改，恢复为服务端数据')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : String(e))
  } finally {
    loadingInstance.close()
  }
}

// ---- 新建本体 ----
const newDlg = ref(false)
const newForm = reactive({
  id: '',
  name: '',
  description: '',
  cloneFrom: '',
})

function openNew() {
  newForm.id = ''
  newForm.name = ''
  newForm.description = ''
  newForm.cloneFrom = ''
  newDlg.value = true
}

async function doCreate() {
  try {
    const rec = await onto.createOntology({
      name: newForm.name || '新本体',
      description: newForm.description || undefined,
    })
    if (rec?.id) {
      onto.switchOntology(String(rec.id))
    }
    newDlg.value = false
    ElMessage.success('本体已创建')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : String(e))
  }
}

async function onDeleteOntology() {
  if (onto.ontologies.length <= 1) {
    ElMessage.warning('不能删除最后一个本体')
    return
  }
  try {
    if (onto.currentOntologyId != null) {
      await onto.deleteOntology(onto.currentOntologyId)
      ElMessage.success('本体已删除')
    }
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : String(e))
  }
}

function onSwitch(id: string) {
  try {
    onto.switchOntology(id)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : String(e))
  }
}

// 初始化：加载本体列表
onMounted(async () => {
  const loadingInstance = ElLoading.service({
    lock: true,
    text: '正在加载工作区...',
    background: 'rgba(255, 255, 255, 0.8)',
  })
  try {
    await onto.fetchOntologies()
    // Auto-select first ontology if available
    if (onto.ontologies.length > 0 && !onto.currentOntologyId) {
      onto.switchOntology(String(onto.ontologies[0].id))
    }
    // Load datasources + mappings in parallel (best-effort, pages refresh on demand)
    await Promise.all([
      ds.fetchDatasources().catch(() => {}),
      maps.fetchMappings({ ontology_id: onto.currentOntologyId ?? undefined }).catch(() => {}),
    ])
    appReady.value = true
  } catch (e) {
    console.error('[App] init failed', e)
    ElMessage.error('加载工作区失败')
    appReady.value = true
  } finally {
    loadingInstance.close()
  }
})
</script>

<template>
  <ElContainer class="app-root">
    <ElHeader class="app-header">
      <div class="brand">
        <span class="logo">O</span>
        <div>
          <div class="title">ODIN Console</div>
          <div class="sub">Ontology-Driven Intelligent Navigator · V1.0</div>
        </div>
      </div>
      <div class="meta">
        <!-- 本体切换器 -->
        <div class="onto-switch">
          <span class="meta-label">本体</span>
          <ElSelect
            :model-value="onto.activeOntologyId"
            size="small"
            style="width: 180px"
            @change="onSwitch"
          >
            <ElOption
              v-for="r in onto.ontologyRegistry"
              :key="r.id"
              :value="r.id"
              :label="r.name"
            >
              <div class="onto-option">
                <span>{{ r.name }}</span>
                <span class="onto-meta">{{ r.classCount }} 类 · v{{ r.version }}</span>
              </div>
            </ElOption>
          </ElSelect>
          <ElButton size="small" text @click="openNew">新建</ElButton>
          <ElPopconfirm title="删除当前本体？此操作不可恢复" @confirm="onDeleteOntology">
            <template #reference>
              <ElButton size="small" text type="danger">删</ElButton>
            </template>
          </ElPopconfirm>
        </div>

        <ElTag size="small" effect="plain" type="info">v{{ onto.version }}</ElTag>
        <ElBadge :value="unmappedCount" type="warning" class="badge">
          <span class="meta-item">未映射</span>
        </ElBadge>
        <ElBadge v-if="anyDirty" :value="pendingCount" type="danger" class="badge">
          <span class="meta-item">未保存</span>
        </ElBadge>
        <ElButton size="small" type="primary" :disabled="!anyDirty" @click="onSave">
          保存
        </ElButton>
        <ElPopconfirm title="放弃未保存修改并重新加载服务端数据？" @confirm="onReset">
          <template #reference>
            <ElButton size="small">重置</ElButton>
          </template>
        </ElPopconfirm>
      </div>
    </ElHeader>

    <ElContainer class="body">
      <ElAside width="88px" class="nav">
        <ElMenu :default-active="active" :collapse="false" class="nav-menu">
          <ElMenuItem index="home" @click="router.push('/')">
            <el-icon><HomeFilled /></el-icon>
            <span>首页</span>
          </ElMenuItem>
          <ElMenuItem index="assistant" @click="router.push('/assistant')">
            <el-icon><ChatDotRound /></el-icon>
            <span>助手</span>
          </ElMenuItem>
          <ElMenuItem index="data-access" @click="router.push('/data-access')">
            <el-icon><Connection /></el-icon>
            <span>数据接入</span>
          </ElMenuItem>
          <ElMenuItem index="sources" @click="router.push('/sources')">
            <el-icon><Coin /></el-icon>
            <span>数据源</span>
          </ElMenuItem>
          <ElMenuItem index="graph" @click="router.push('/ontology')">
            <el-icon><Share /></el-icon>
            <span>图谱</span>
          </ElMenuItem>
          <ElMenuItem index="mapping" @click="router.push('/mapping')">
            <el-icon><DataLine /></el-icon>
            <span>映射</span>
          </ElMenuItem>
          <ElMenuItem index="connectors" @click="router.push('/connectors')">
            <el-icon><Connection /></el-icon>
            <span>连接器</span>
          </ElMenuItem>
          <ElMenuItem index="reason" @click="router.push('/reason')">
            <el-icon><MagicStick /></el-icon>
            <span>推理</span>
          </ElMenuItem>
          <ElMenuItem index="query" @click="router.push('/query')">
            <el-icon><Search /></el-icon>
            <span>查询</span>
          </ElMenuItem>
        </ElMenu>
      </ElAside>

      <ElMain class="main">
        <router-view v-slot="{ Component }">
          <component
            :is="Component"
            @go-to-assistant="goToAssistant"
            @go-to-data-access="goToDataAccess"
            @go-back="goBackFromAssistant"
          />
        </router-view>
      </ElMain>
    </ElContainer>

    <ElDialog v-model="newDlg" title="新建本体" width="480px">
      <ElForm label-width="80px" size="small">
        <ElFormItem label="本体 ID" required>
          <ElInput v-model="newForm.id" placeholder="如 supply-chain" />
        </ElFormItem>
        <ElFormItem label="名称" required>
          <ElInput v-model="newForm.name" placeholder="如 供应链域" />
        </ElFormItem>
        <ElFormItem label="描述">
          <ElInput v-model="newForm.description" type="textarea" :rows="2" />
        </ElFormItem>
        <ElFormItem label="克隆源">
          <ElSelect v-model="newForm.cloneFrom" clearable style="width: 100%" placeholder="留空则创建空白本体">
            <ElOption
              v-for="r in onto.ontologyRegistry"
              :key="r.id"
              :value="r.id"
              :label="`${r.name}（${r.classCount} 类）`"
            />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="newDlg = false">取消</ElButton>
        <ElButton type="primary" @click="doCreate">创建</ElButton>
      </template>
    </ElDialog>

    <!-- 全链路溯源抽屉：由 trace store 驱动，任意页面调用 open(traceId) 即可唤起 -->
    <TraceDrawer />
  </ElContainer>
</template>

<style scoped>
.app-root {
  height: 100vh;
  background: #f1f5f9;
}

.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #0f172a;
  color: #e2e8f0;
  height: 56px !important;
  padding: 0 16px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
}

.logo {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: linear-gradient(135deg, #4c8bf5, #22d3ee);
  display: grid;
  place-items: center;
  font-weight: 700;
  color: #fff;
}

.title {
  font-size: 14px;
  font-weight: 600;
  line-height: 1.2;
}

.sub {
  font-size: 11px;
  color: #94a3b8;
}

.meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.meta-item {
  font-size: 12px;
  color: #cbd5e1;
}

.meta-label {
  font-size: 12px;
  color: #94a3b8;
  flex-shrink: 0;
}

.onto-switch {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.08);
  padding: 4px 10px;
  border-radius: 8px;
}

.onto-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.onto-meta {
  font-size: 11px;
  color: #94a3b8;
}

.body {
  min-height: 0;
}

.nav {
  background: #fff;
  border-right: 1px solid #e2e8f0;
  padding-top: 8px;
}

.nav-menu {
  border-right: none;
  height: 100%;
}

.nav-menu :deep(.el-menu-item) {
  height: 64px;
  line-height: 1.2;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 0 !important;
  color: #475569;
}

.nav-menu :deep(.el-menu-item .el-icon) {
  margin: 0;
  font-size: 18px;
}

.nav-menu :deep(.el-menu-item span) {
  font-size: 12px;
}

.nav-menu :deep(.el-menu-item.is-active) {
  color: #2563eb;
  background: #eff6ff;
}

.nav-menu :deep(.el-menu-item.is-disabled) {
  color: #cbd5e1;
}

.main {
  padding: 0;
  min-height: 0;
  overflow: auto;
}
</style>
