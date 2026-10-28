<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import {
  ChatDotRound,
  Coin,
  Share,
  Connection,
  MagicStick,
  Guide,
  ArrowRight,
  Plus,
  SetUp,
} from '@element-plus/icons-vue'
import OntologyFormDialog from '@/views/odin/components/ontology/OntologyFormDialog.vue'
import { useOdinWorkspaceStore } from '@/stores/odin/workspace'
import { createOntology } from '@/api/odin/ontology'
import type { OntologyDTO, OntologySummary } from '@/types/odin'

/**
 * ODIN 品牌落地页（替代原任务调度 Dashboard）。
 *
 * 设计语言「北欧罗盘」：深夜蓝 × 极光青 × 黄铜金，
 * 衬线 display 字标 + 旋转虚线罗盘环 + 点阵底纹，
 * 全部颜色基于 --sm-* 变量 color-mix 派生，自动适配深色模式。
 */
const router = useRouter()
const { t } = useI18n()
const wsStore = useOdinWorkspaceStore()

onMounted(() => {
  void wsStore.loadOntologies()
})

/** 四大核心特性（2×2 网格） */
const features = computed(() => [
  {
    icon: Share,
    tone: 'primary',
    title: t('dashboard.featureOntologyTitle'),
    desc: t('dashboard.featureOntologyDesc'),
  },
  {
    icon: Connection,
    tone: 'success',
    title: t('dashboard.featureFederationTitle'),
    desc: t('dashboard.featureFederationDesc'),
  },
  {
    icon: MagicStick,
    tone: 'warning',
    title: t('dashboard.featureAiTitle'),
    desc: t('dashboard.featureAiDesc'),
  },
  {
    icon: Guide,
    tone: 'danger',
    title: t('dashboard.featureTraceTitle'),
    desc: t('dashboard.featureTraceDesc'),
  },
])

/** 三个场景推荐（优先读当前本体的预置问题，为空回退通用文案） */
const scenarios = computed(() => {
  const qs = wsStore.activeOntology?.presetQuestions ?? []
  if (qs.length) {
    return qs.slice(0, 3).map((q) => ({
      title: q,
      desc: wsStore.activeOntology?.name ?? '',
      example: q,
    }))
  }
  return [
    {
      title: t('dashboard.scenario1Title'),
      desc: t('dashboard.scenario1Desc'),
      example: t('dashboard.scenario1Example'),
    },
    {
      title: t('dashboard.scenario2Title'),
      desc: t('dashboard.scenario2Desc'),
      example: t('dashboard.scenario2Example'),
    },
    {
      title: t('dashboard.scenario3Title'),
      desc: t('dashboard.scenario3Desc'),
      example: t('dashboard.scenario3Example'),
    },
  ]
})

/** 本体卡片区：左列展示（分页：每页 4 个 = 2 列 × 2 行） */
const ontologies = computed<OntologySummary[]>(() => wsStore.ontologies)
const ontPage = ref(1)
const ONT_PAGE_SIZE = 4
const ontTotalPages = computed(() =>
  Math.max(1, Math.ceil(ontologies.value.length / ONT_PAGE_SIZE)),
)
const pagedOntologies = computed<OntologySummary[]>(() => {
  const start = (ontPage.value - 1) * ONT_PAGE_SIZE
  return ontologies.value.slice(start, start + ONT_PAGE_SIZE)
})
watch(ontTotalPages, (tp) => {
  if (ontPage.value > tp) ontPage.value = tp
})

function goToAssistant() {
  router.push('/odin/assistant')
}

function goToSources() {
  router.push('/odin/data-access/sources')
}

/** 场景推荐：跳转助手并预填示例问句 */
function askScenario(example: string) {
  router.push({ path: '/odin/assistant', query: { q: example } })
}

/** 本体卡片：设为当前本体后直达助手/治理 */
function goAssistantFor(o: OntologySummary) {
  wsStore.setActive(o.id)
  router.push('/odin/assistant')
}

function governOntology(o: OntologySummary) {
  wsStore.setActive(o.id)
  router.push('/odin/ontology')
}

// ===== 新建本体弹窗 =====
const ontDlgVisible = ref(false)

function openNewOntology() {
  ontDlgVisible.value = true
}

async function handleOntSave(data: OntologyDTO) {
  await createOntology(data)
  ElMessage.success(t('common.createSuccess'))
  await wsStore.loadOntologies()
  if (data.code) {
    const created = wsStore.ontologies.find((o) => o.code === data.code)
    if (created) wsStore.setActive(created.id)
  }
}
</script>

<template>
  <div class="odin-home">
    <!-- 大气层：点阵 + 极光晕 + 地平线 -->
    <div class="atmosphere" aria-hidden="true">
      <div class="aurora aurora-a" />
      <div class="aurora aurora-b" />
      <div class="horizon" />
    </div>

    <!-- 主区：品牌 + 控制台 -->
    <div class="home-hero">
    <!-- 左侧：品牌区 -->
    <section class="brand-side">
      <div class="brand-block">
        <div class="eyebrow rise" style="--d: 0">
          <span class="eyebrow-rune">◈</span>
          <span>{{ t('dashboard.eyebrow') }}</span>
        </div>

        <div class="wordmark-wrap rise" style="--d: 1">
          <div class="compass-ring" aria-hidden="true">
            <span class="ring ring-outer" />
            <span class="ring ring-inner" />
          </div>
          <h1 class="wordmark">ODIN</h1>
        </div>

        <p class="slogan rise" style="--d: 2">{{ t('dashboard.slogan') }}</p>
        <p class="slogan-en rise" style="--d: 3">{{ t('dashboard.sloganEn') }}</p>
        <div class="hairline rise" style="--d: 3" />
      </div>

      <!-- 特性卡片 2×2 -->
      <div class="feature-grid">
        <article
          v-for="(f, i) in features"
          :key="f.title"
          class="feature-card rise"
          :class="`tone-${f.tone}`"
          :style="{ '--d': 4 + i }"
        >
          <span class="feature-glyph">
            <el-icon :size="20"><component :is="f.icon" /></el-icon>
          </span>
          <div class="feature-text">
            <h4>{{ f.title }}</h4>
            <p>{{ f.desc }}</p>
          </div>
          <span class="feature-index">0{{ i + 1 }}</span>
        </article>
      </div>

      <!-- 本体工作台：卡片区（直达助手 / 治理 + 新建本体） -->
      <section class="ontology-rail rise" style="--d: 8">
        <header class="rail-head">
          <h3 class="rail-title">{{ t('dashboard.railTitle') }}</h3>
          <span class="rail-line" />
          <el-button type="primary" plain size="small" :icon="Plus" @click="openNewOntology">
            {{ t('dashboard.railNew') }}
          </el-button>
        </header>

        <div class="rail-grid">
          <article
            v-for="o in pagedOntologies"
            :key="o.id"
            class="ont-card"
            :class="{ 'is-active': o.id === wsStore.activeOntologyId }"
          >
            <div class="ont-card-head">
              <span class="ont-name">{{ o.name }}</span>
              <code class="ont-code">{{ o.code }}</code>
            </div>
            <p class="ont-desc">{{ o.description || t('dashboard.ontNoDesc') }}</p>
            <div class="ont-metrics">
              <span class="metric"><em>{{ o.classCount ?? 0 }}</em> {{ t('dashboard.ontClassCount') }}</span>
              <span class="metric"><em>{{ o.relationCount ?? 0 }}</em> {{ t('dashboard.ontRelationCount') }}</span>
            </div>
            <div class="ont-actions">
              <button class="ont-btn ont-btn-ask" type="button" @click="goAssistantFor(o)">
                <el-icon :size="15"><ChatDotRound /></el-icon>
                {{ t('dashboard.ontAskBtn') }}
              </button>
              <button class="ont-btn ont-btn-govern" type="button" @click="governOntology(o)">
                <el-icon :size="15"><SetUp /></el-icon>
                {{ t('dashboard.ontGovernBtn') }}
              </button>
            </div>
          </article>
        </div>

        <div v-if="ontologies.length > ONT_PAGE_SIZE" class="rail-pager">
          <el-pagination
            v-model:current-page="ontPage"
            layout="prev, pager, next"
            size="small"
            background
            :total="ontologies.length"
            :page-size="ONT_PAGE_SIZE"
          />
        </div>

        <el-empty
          v-if="!ontologies.length"
          :image-size="48"
          :description="t('dashboard.ontEmpty')"
        >
          <el-button type="primary" plain :icon="Plus" @click="openNewOntology">
            {{ t('dashboard.railNew') }}
          </el-button>
        </el-empty>
      </section>
    </section>

    <!-- 右侧：操作控制台 -->
    <aside class="console-side rise" style="--d: 2">
      <div class="console-panel">
        <header class="console-head">
          <span class="console-title">{{ t('dashboard.consoleTitle') }}</span>
          <span class="console-line" />
        </header>

        <!-- 主操作 -->
        <div class="action-stack">
          <button class="action-card action-primary" type="button" @click="goToAssistant">
            <span class="action-icon">
              <el-icon :size="26"><ChatDotRound /></el-icon>
            </span>
            <span class="action-text">
              <strong>{{ t('dashboard.startAsk') }}</strong>
              <small>{{ t('dashboard.startAskDesc') }}</small>
            </span>
            <span class="action-tag">{{ t('dashboard.recommendTag') }}</span>
            <el-icon class="action-arrow" :size="18"><ArrowRight /></el-icon>
          </button>

          <button class="action-card action-secondary" type="button" @click="goToSources">
            <span class="action-icon">
              <el-icon :size="26"><Coin /></el-icon>
            </span>
            <span class="action-text">
              <strong>{{ t('dashboard.dataAccess') }}</strong>
              <small>{{ t('dashboard.dataAccessDesc') }}</small>
            </span>
            <span class="action-tag">{{ t('dashboard.quickTag') }}</span>
            <el-icon class="action-arrow" :size="18"><ArrowRight /></el-icon>
          </button>
        </div>

        <!-- 场景推荐 -->
        <div class="scenario-block">
          <h3 class="block-title">{{ t('dashboard.scenarioTitle') }}</h3>
          <button
            v-for="(s, i) in scenarios"
            :key="s.title"
            class="scenario-item"
            type="button"
            @click="askScenario(s.example)"
          >
            <span class="scenario-no">{{ i + 1 }}</span>
            <span class="scenario-text">
              <strong>{{ s.title }}</strong>
              <small>{{ s.desc }}</small>
              <code>“{{ s.example }}”</code>
            </span>
            <el-icon class="scenario-arrow" :size="14"><ArrowRight /></el-icon>
          </button>
        </div>

        <!-- 快速开始 -->
        <footer class="quick-start">
          <span class="quick-icon">✦</span>
          <!-- eslint-disable-next-line vue/no-v-html -->
          <p v-html="t('dashboard.quickTip')" />
        </footer>
      </div>
    </aside>
    </div>

    <OntologyFormDialog v-model:visible="ontDlgVisible" @save="handleOntSave" />
  </div>
</template>

<style scoped>
/* ============ 页面骨架 ============ */
.odin-home {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-height: calc(100vh - var(--sm-header-height) - 48px);
  padding: var(--sm-spacing-lg) clamp(20px, 3vw, 48px);
  overflow: hidden auto;
  background:
    radial-gradient(
      1200px 600px at 12% -10%,
      color-mix(in srgb, var(--sm-color-primary) 14%, transparent),
      transparent 60%
    ),
    radial-gradient(
      900px 520px at 92% 110%,
      color-mix(in srgb, var(--sm-color-success) 10%, transparent),
      transparent 60%
    ),
    var(--sm-bg-page);
}

.home-hero {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: stretch;
  gap: var(--sm-spacing-lg);
  min-height: 0;
}

/* ============ 大气层 ============ */
.atmosphere {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.atmosphere::before {
  /* 点阵底纹 */
  content: '';
  position: absolute;
  inset: 0;
  background-image: radial-gradient(
    color-mix(in srgb, var(--sm-text-secondary) 22%, transparent) 1px,
    transparent 1px
  );
  background-size: 26px 26px;
  opacity: 0.5;
  mask-image: linear-gradient(160deg, #000 20%, transparent 75%);
}

.aurora {
  position: absolute;
  border-radius: 50%;
  filter: blur(90px);
  animation: drift 18s ease-in-out infinite alternate;
}

.aurora-a {
  width: 480px;
  height: 300px;
  top: -120px;
  left: 18%;
  background: color-mix(in srgb, var(--sm-color-primary) 22%, transparent);
}

.aurora-b {
  width: 420px;
  height: 260px;
  bottom: -100px;
  left: 42%;
  background: color-mix(in srgb, var(--sm-color-success) 16%, transparent);
  animation-delay: -9s;
}

.horizon {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent,
    color-mix(in srgb, var(--sm-color-primary) 45%, transparent),
    transparent
  );
}

@keyframes drift {
  from {
    transform: translateX(-40px) scale(1);
  }
  to {
    transform: translateX(60px) scale(1.12);
  }
}

/* ============ 入场动画（交错） ============ */
.rise {
  animation: rise 0.7s cubic-bezier(0.22, 1, 0.36, 1) both;
  animation-delay: calc(var(--d, 0) * 90ms);
}

@keyframes rise {
  from {
    opacity: 0;
    transform: translateY(22px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* ============ 左侧品牌区 ============ */
.brand-side {
  position: relative;
  z-index: 1;
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 8px 0;
}

.eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-size: var(--sm-font-size-sm);
  font-weight: 600;
  letter-spacing: 0.35em;
  text-transform: uppercase;
  color: var(--sm-text-secondary);
}

.eyebrow-rune {
  color: var(--sm-color-warning);
  letter-spacing: 0;
}

.wordmark-wrap {
  position: relative;
  margin-top: 8px;
}

.wordmark {
  position: relative;
  z-index: 1;
  margin: 0;
  font-family: Georgia, 'Times New Roman', 'Songti SC', 'SimSun', serif;
  font-size: clamp(48px, 6vw, 84px);
  font-weight: 700;
  line-height: 1;
  letter-spacing: 0.14em;
  text-indent: 0.14em; /* 抵消末字符字距，保持视觉居中 */
  background: linear-gradient(
    100deg,
    var(--sm-text-primary) 15%,
    var(--sm-color-primary) 55%,
    var(--sm-color-success) 90%
  );
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  color: transparent;
}

/* 旋转罗盘环 */
.compass-ring {
  position: absolute;
  top: 50%;
  right: 8%;
  width: 150px;
  height: 150px;
  transform: translateY(-50%);
  pointer-events: none;
}

.ring {
  position: absolute;
  inset: 0;
  border-radius: 50%;
}

.ring-outer {
  border: 1px dashed color-mix(in srgb, var(--sm-color-primary) 45%, transparent);
  animation: spin 48s linear infinite;
}

.ring-inner {
  inset: 34px;
  border: 1px solid color-mix(in srgb, var(--sm-color-warning) 40%, transparent);
  border-top-color: transparent;
  border-left-color: transparent;
  animation: spin 30s linear infinite reverse;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.slogan {
  margin: 12px 0 0;
  font-size: clamp(18px, 2vw, 26px);
  font-weight: 700;
  letter-spacing: 0.12em;
  color: var(--sm-text-primary);
}

.slogan-en {
  margin: 6px 0 0;
  font-family: Georgia, 'Times New Roman', serif;
  font-style: italic;
  font-size: var(--sm-font-size-lg);
  letter-spacing: 0.06em;
  color: var(--sm-text-secondary);
}

.hairline {
  width: 120px;
  height: 2px;
  margin-top: 14px;
  background: linear-gradient(
    90deg,
    var(--sm-color-warning),
    color-mix(in srgb, var(--sm-color-warning) 10%, transparent)
  );
}

/* ============ 特性卡片 ============ */
.feature-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--sm-spacing-md);
}

.feature-card {
  --tone: var(--sm-color-primary);
  position: relative;
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px 14px;
  overflow: hidden;
  background: color-mix(in srgb, var(--sm-bg-card) 82%, transparent);
  border: 1px solid var(--sm-border-light);
  border-radius: var(--sm-radius-lg);
  backdrop-filter: blur(8px);
  transition:
    transform 0.3s ease,
    border-color 0.3s ease,
    box-shadow 0.3s ease;
}

.feature-card.tone-success {
  --tone: var(--sm-color-success);
}
.feature-card.tone-warning {
  --tone: var(--sm-color-warning);
}
.feature-card.tone-danger {
  --tone: var(--sm-color-danger);
}

.feature-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, var(--tone), transparent 70%);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-3px);
  border-color: color-mix(in srgb, var(--tone) 45%, var(--sm-border-light));
  box-shadow: 0 10px 28px color-mix(in srgb, var(--tone) 16%, transparent);
}

.feature-card:hover::before {
  opacity: 1;
}

.feature-glyph {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  flex-shrink: 0;
  border-radius: var(--sm-radius-lg);
  color: var(--tone);
  background: color-mix(in srgb, var(--tone) 12%, transparent);
  border: 1px solid color-mix(in srgb, var(--tone) 25%, transparent);
}

.feature-text h4 {
  margin: 2px 0 6px;
  font-size: var(--sm-font-size-base);
  font-weight: 700;
  letter-spacing: 0.04em;
  color: var(--sm-text-primary);
}

.feature-text p {
  margin: 0;
  font-size: var(--sm-font-size-xs);
  line-height: 1.6;
  color: var(--sm-text-secondary);
}

.feature-index {
  position: absolute;
  top: 12px;
  right: 14px;
  font-family: Georgia, 'Times New Roman', serif;
  font-size: var(--sm-font-size-xs);
  font-style: italic;
  color: color-mix(in srgb, var(--sm-text-secondary) 55%, transparent);
}

/* ============ 右侧控制台 ============ */
.console-side {
  position: relative;
  z-index: 1;
  width: 420px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.console-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
  max-height: calc(100vh - var(--sm-header-height) - 40px);
  overflow-y: auto;
  padding: 20px;
  background: var(--sm-bg-card);
  border: 1px solid var(--sm-border-light);
  border-radius: 16px;
  box-shadow: 0 18px 50px color-mix(in srgb, var(--sm-text-primary) 8%, transparent);
}

@media (max-width: 1200px) {
  .console-panel {
    max-height: none;
  }
}

.console-head {
  display: flex;
  align-items: center;
  gap: 14px;
}

.console-title {
  font-size: var(--sm-font-size-sm);
  font-weight: 700;
  letter-spacing: 0.3em;
  text-transform: uppercase;
  color: var(--sm-text-secondary);
  white-space: nowrap;
}

.console-line {
  flex: 1;
  height: 1px;
  background: linear-gradient(
    90deg,
    var(--sm-border-base),
    transparent
  );
}

/* ============ 主操作卡片 ============ */
.action-stack {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.action-card {
  --accent: var(--sm-color-primary);
  position: relative;
  display: flex;
  align-items: center;
  gap: 16px;
  width: 100%;
  padding: 14px 18px;
  text-align: left;
  cursor: pointer;
  border: 1px solid transparent;
  border-radius: 14px;
  color: #fff;
  transition:
    transform 0.3s cubic-bezier(0.22, 1, 0.36, 1),
    box-shadow 0.3s ease;
}

.action-primary {
  background: linear-gradient(
    120deg,
    var(--sm-color-primary),
    color-mix(in srgb, var(--sm-color-primary) 55%, var(--sm-nav-bg-end))
  );
  box-shadow: 0 10px 30px color-mix(in srgb, var(--sm-color-primary) 30%, transparent);
}

.action-secondary {
  --accent: var(--sm-color-success);
  background: linear-gradient(
    120deg,
    var(--sm-color-success),
    color-mix(in srgb, var(--sm-color-success) 55%, var(--sm-nav-bg-start))
  );
  box-shadow: 0 10px 30px color-mix(in srgb, var(--sm-color-success) 26%, transparent);
}

.action-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 16px 40px color-mix(in srgb, var(--accent) 38%, transparent);
}

.action-card:active {
  transform: translateY(-1px);
}

.action-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  flex-shrink: 0;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.16);
  border: 1px solid rgba(255, 255, 255, 0.28);
  backdrop-filter: blur(6px);
}

.action-text {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.action-text strong {
  font-size: var(--sm-font-size-xl);
  font-weight: 700;
  letter-spacing: 0.06em;
}

.action-text small {
  font-size: var(--sm-font-size-xs);
  line-height: 1.5;
  opacity: 0.85;
}

.action-tag {
  position: absolute;
  top: -9px;
  right: 16px;
  padding: 2px 10px;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.1em;
  color: var(--sm-text-primary);
  background: color-mix(in srgb, var(--sm-color-warning) 85%, #fff);
  border-radius: 999px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.18);
}

.action-arrow {
  flex-shrink: 0;
  opacity: 0.75;
  transition: transform 0.3s ease;
}

.action-card:hover .action-arrow {
  transform: translateX(6px);
  opacity: 1;
}

/* ============ 场景推荐 ============ */
.scenario-block {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.block-title {
  margin: 0;
  font-size: var(--sm-font-size-sm);
  font-weight: 700;
  letter-spacing: 0.2em;
  color: var(--sm-text-secondary);
}

.scenario-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  cursor: pointer;
  text-align: left;
  background: var(--sm-bg-page);
  border: 1px solid var(--sm-border-lighter);
  border-radius: 12px;
  transition:
    border-color 0.25s ease,
    transform 0.25s ease,
    box-shadow 0.25s ease;
}

.scenario-item:hover {
  transform: translateX(4px);
  border-color: color-mix(in srgb, var(--sm-color-primary) 50%, var(--sm-border-light));
  box-shadow: 0 6px 18px color-mix(in srgb, var(--sm-color-primary) 12%, transparent);
}

.scenario-no {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  flex-shrink: 0;
  font-family: Georgia, 'Times New Roman', serif;
  font-size: var(--sm-font-size-sm);
  font-weight: 700;
  font-style: italic;
  color: var(--sm-color-primary);
  background: color-mix(in srgb, var(--sm-color-primary) 10%, transparent);
  border: 1px solid color-mix(in srgb, var(--sm-color-primary) 25%, transparent);
  border-radius: 8px;
}

.scenario-text {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.scenario-text strong {
  font-size: var(--sm-font-size-sm);
  font-weight: 600;
  color: var(--sm-text-primary);
}

.scenario-text small {
  font-size: var(--sm-font-size-xs);
  color: var(--sm-text-secondary);
}

.scenario-text code {
  margin-top: 4px;
  padding: 3px 8px;
  font-size: 11px;
  font-family: 'JetBrains Mono', Consolas, monospace;
  color: var(--sm-color-primary);
  background: color-mix(in srgb, var(--sm-color-primary) 8%, transparent);
  border: 1px solid color-mix(in srgb, var(--sm-color-primary) 18%, transparent);
  border-radius: 6px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.scenario-arrow {
  flex-shrink: 0;
  color: var(--sm-text-placeholder);
  transition: transform 0.25s ease;
}

.scenario-item:hover .scenario-arrow {
  transform: translateX(4px);
  color: var(--sm-color-primary);
}

/* ============ 快速开始 ============ */
.quick-start {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 14px 16px;
  background: color-mix(in srgb, var(--sm-color-warning) 7%, var(--sm-bg-page));
  border: 1px dashed color-mix(in srgb, var(--sm-color-warning) 40%, transparent);
  border-radius: 12px;
}

.quick-icon {
  color: var(--sm-color-warning);
  font-size: var(--sm-font-size-lg);
  line-height: 1.4;
}

.quick-start p {
  margin: 0;
  font-size: var(--sm-font-size-xs);
  line-height: 1.7;
  color: var(--sm-text-regular);
}

.quick-start :deep(strong) {
  color: var(--sm-color-primary);
  font-weight: 700;
}

/* ============ 响应式 ============ */
@media (max-width: 1200px) {
  .odin-home {
    flex-direction: column;
    align-items: stretch;
  }

  .console-side {
    width: 100%;
  }

  .console-panel {
    max-height: none;
  }

  .compass-ring {
    display: none;
  }
}

@media (max-width: 640px) {
  .feature-grid {
    grid-template-columns: 1fr;
  }
}

@media (prefers-reduced-motion: reduce) {
  .rise,
  .aurora,
  .ring-outer,
  .ring-inner {
    animation: none;
  }
}

/* ============ 本体工作台区（左列卡片区） ============ */
.ontology-rail {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rail-head {
  display: flex;
  align-items: center;
  gap: 14px;
}

.rail-title {
  margin: 0;
  font-size: var(--sm-font-size-base);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--sm-text-secondary);
  white-space: nowrap;
}

.rail-line {
  flex: 1;
  height: 1px;
  background: linear-gradient(90deg, var(--sm-border-base), transparent);
}

.rail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--sm-spacing-md);
}

.rail-pager {
  display: flex;
  justify-content: center;
  margin-top: 2px;
}

.ont-card {
  --tone: var(--sm-color-primary);
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 14px 16px;
  text-align: left;
  background: color-mix(in srgb, var(--sm-bg-card) 88%, transparent);
  border: 1px solid var(--sm-border-light);
  border-radius: var(--sm-radius-lg);
  backdrop-filter: blur(8px);
  transition:
    transform 0.3s ease,
    border-color 0.3s ease,
    box-shadow 0.3s ease;
}

.ont-card:hover {
  transform: translateY(-3px);
  border-color: color-mix(in srgb, var(--tone) 45%, var(--sm-border-light));
  box-shadow: 0 10px 28px color-mix(in srgb, var(--tone) 16%, transparent);
}

.ont-card.is-active {
  border-color: color-mix(in srgb, var(--tone) 55%, transparent);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--tone) 22%, transparent);
}

.ont-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.ont-name {
  font-size: var(--sm-font-size-base);
  font-weight: 700;
  color: var(--sm-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ont-code {
  flex-shrink: 0;
  padding: 2px 8px;
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 11px;
  color: var(--tone);
  background: color-mix(in srgb, var(--tone) 10%, transparent);
  border-radius: 6px;
}

.ont-desc {
  margin: 0;
  font-size: var(--sm-font-size-xs);
  line-height: 1.6;
  color: var(--sm-text-secondary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 2.6em;
}

.ont-metrics {
  display: flex;
  gap: 16px;
}

.ont-metrics .metric {
  font-size: var(--sm-font-size-xs);
  color: var(--sm-text-secondary);
}

.ont-metrics .metric em {
  font-style: normal;
  font-weight: 700;
  color: var(--sm-text-primary);
  margin-right: 2px;
}

.ont-actions {
  display: flex;
  gap: 10px;
  margin-top: auto;
}

.ont-btn {
  flex: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 12px;
  font-size: var(--sm-font-size-xs);
  font-weight: 600;
  cursor: pointer;
  border-radius: 10px;
  border: 1px solid transparent;
  transition:
    background 0.2s ease,
    transform 0.1s ease;
}

.ont-btn:active {
  transform: scale(0.98);
}

.ont-btn-ask {
  color: #fff;
  background: linear-gradient(120deg, var(--sm-color-primary), color-mix(in srgb, var(--sm-color-primary) 60%, var(--sm-nav-bg-end)));
}

.ont-btn-ask:hover {
  filter: brightness(1.06);
}

.ont-btn-govern {
  color: var(--sm-text-primary);
  background: var(--sm-bg-page);
  border-color: var(--sm-border-lighter);
}

.ont-btn-govern:hover {
  border-color: color-mix(in srgb, var(--tone) 40%, var(--sm-border-light));
  color: var(--tone);
}

.ont-card-new {
  --tone: var(--sm-color-success);
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 150px;
  cursor: pointer;
  color: var(--sm-text-secondary);
  background: color-mix(in srgb, var(--sm-bg-page) 60%, transparent);
  border: 1px dashed color-mix(in srgb, var(--tone) 45%, var(--sm-border-base));
}

.ont-card-new span {
  font-size: var(--sm-font-size-sm);
  font-weight: 600;
}

.ont-card-new:hover {
  color: var(--tone);
  border-color: var(--tone);
  background: color-mix(in srgb, var(--tone) 8%, transparent);
}

@media (max-width: 1200px) {
  .home-hero {
    flex-direction: column;
    align-items: stretch;
  }

  .console-side {
    width: 100%;
  }
}

@media (max-width: 1100px) {
  .feature-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .rail-grid {
    grid-template-columns: 1fr;
  }
}
</style>
