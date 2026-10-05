<script setup lang="ts">
import {
  ElTag,
  ElIcon,
} from 'element-plus'
import {
  Connection,
  Share,
  MagicStick,
  TrendCharts,
  Star,
} from '@element-plus/icons-vue'

const emit = defineEmits<{
  goToAssistant: []
  goToDataAccess: []
}>()

// 核心特性
const features = [
  {
    icon: Share,
    title: '本体驱动',
    desc: '以业务概念为核心，而非技术表结构',
    color: '#409eff',
  },
  {
    icon: Connection,
    title: '多源统一',
    desc: '一个接口访问所有数据源',
    color: '#67c23a',
  },
  {
    icon: MagicStick,
    title: '智能映射',
    desc: '自动推断本体与数据的关联',
    color: '#e6a23c',
  },
  {
    icon: TrendCharts,
    title: '数据洞察',
    desc: '发现数据背后的业务逻辑',
    color: '#f56c6c',
  },
]

// 应用场景
const scenarios = [
  {
    icon: '🔗',
    title: '跨库数据查询',
    desc: 'ERP + CRM + OA 联合分析',
    example: '查询华东区VIP客户今年的订单总额',
  },
  {
    icon: '🔍',
    title: '多因素问题诊断',
    desc: '订单延迟原因分析',
    example: '分析订单延迟的可能原因',
  },
  {
    icon: '📊',
    title: '业务指标监控',
    desc: '实时监控关键指标',
    example: '本月销售额环比增长情况',
  },
]
</script>

<template>
  <div class="home-page">
    <!-- 左侧：品牌与特性 -->
    <div class="left-section">
      <div class="brand-area">
        <div class="logo-badge">
          <ElTag type="success" effect="dark" size="large">
            <ElIcon><Star /></ElIcon> ODIN
          </ElTag>
        </div>
        <h1 class="main-title">
          基于本体的<br><span class="highlight">数据洞察</span>平台
        </h1>
        <p class="subtitle">
          Ontology-Driven Intelligent Navigator
        </p>
        <p class="desc">
          让数据访问变得简单、智能、统一<br>
          用业务语言与数据对话
        </p>
      </div>

      <!-- 核心特性 -->
      <div class="features-grid">
        <div v-for="feature in features" :key="feature.title" class="feature-item">
          <div class="feature-icon" :style="{ background: feature.color + '15', color: feature.color }">
            <ElIcon><component :is="feature.icon" /></ElIcon>
          </div>
          <div class="feature-info">
            <h4>{{ feature.title }}</h4>
            <p>{{ feature.desc }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 右侧：操作入口 -->
    <div class="right-section">
      <!-- 主操作卡片 -->
      <div class="action-cards">
        <div class="action-card primary" @click="emit('goToAssistant')">
          <div class="card-glow"></div>
          <div class="card-inner">
            <div class="card-icon-wrapper">
              <div class="card-icon">🤖</div>
            </div>
            <div class="card-content">
              <h3>智能助手</h3>
              <p>开始对话，了解 ODIN 的核心理念和功能</p>
              <span class="card-tag">推荐入口</span>
            </div>
            <div class="card-arrow">→</div>
          </div>
        </div>

        <div class="action-card secondary" @click="emit('goToDataAccess')">
          <div class="card-glow"></div>
          <div class="card-inner">
            <div class="card-icon-wrapper">
              <div class="card-icon">🔌</div>
            </div>
            <div class="card-content">
              <h3>数据接入</h3>
              <p>配置数据源，开始接入您的数据</p>
              <span class="card-tag">快速开始</span>
            </div>
            <div class="card-arrow">→</div>
          </div>
        </div>
      </div>

      <!-- 应用场景 -->
      <div class="scenarios-section">
        <h3 class="scenarios-title">
          <span class="title-icon">🎯</span>
          <span>典型应用场景</span>
        </h3>
        <div class="scenarios-list">
          <div v-for="scenario in scenarios" :key="scenario.title" class="scenario-item">
            <div class="scenario-number">{{ scenarios.indexOf(scenario) + 1 }}</div>
            <div class="scenario-info">
              <h4>{{ scenario.title }}</h4>
              <p>{{ scenario.desc }}</p>
              <div class="scenario-example">
                <code>"{{ scenario.example }}"</code>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 快速开始提示 -->
      <div class="quick-start">
        <div class="quick-start-icon">💡</div>
        <p>点击 <strong>智能助手</strong> 开始了解 ODIN，或直接进入 <strong>数据接入</strong> 配置数据源</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.home-page {
  height: 100%;
  display: flex;
  background: linear-gradient(135deg, #e8f4fd 0%, #d1e9ff 50%, #f0f7ff 100%);
  overflow: hidden;
}

/* 左侧区域 */
.left-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 60px 40px 60px 60px;
  color: #1a365d;
}

.brand-area {
  margin-bottom: 40px;
}

.logo-badge {
  margin-bottom: 20px;
}

.main-title {
  font-size: 42px;
  font-weight: 700;
  line-height: 1.2;
  margin: 0 0 16px;
  color: #1a365d;
}

.highlight {
  background: linear-gradient(90deg, #2b6cb0, #3182ce);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.subtitle {
  font-size: 18px;
  color: #4a6fa5;
  margin: 0 0 12px;
  font-weight: 500;
}

.desc {
  font-size: 16px;
  color: #718096;
  margin: 0;
  line-height: 1.6;
}

/* 特性网格 */
.features-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: rgba(255, 255, 255, 0.85);
  border-radius: 12px;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(66, 153, 225, 0.2);
  box-shadow: 0 2px 8px rgba(66, 153, 225, 0.08);
  transition: all 0.3s ease;
}

.feature-item:hover {
  background: rgba(255, 255, 255, 0.95);
  box-shadow: 0 4px 16px rgba(66, 153, 225, 0.15);
  transform: translateY(-2px);
}

.feature-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
  background: white !important;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.feature-info h4 {
  font-size: 15px;
  font-weight: 600;
  margin: 0 0 4px;
  color: #2d3748;
}

.feature-info p {
  font-size: 13px;
  margin: 0;
  color: #718096;
}

/* 右侧区域 */
.right-section {
  width: 480px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 24px 0 0 24px;
  padding: 40px;
  display: flex;
  flex-direction: column;
  box-shadow: -4px 0 24px rgba(0, 0, 0, 0.06);
  backdrop-filter: blur(20px);
}

/* 操作卡片 */
.action-cards {
  display: flex;
  flex-direction: column;
  gap: 20px;
  margin-bottom: 36px;
}

.action-card {
  position: relative;
  border-radius: 20px;
  cursor: pointer;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}

.action-card:hover {
  transform: translateY(-4px);
}

.action-card.primary {
  background: linear-gradient(135deg, #3182ce 0%, #2b6cb0 100%);
  box-shadow: 0 8px 32px rgba(49, 130, 206, 0.25);
}

.action-card.primary:hover {
  box-shadow: 0 16px 48px rgba(49, 130, 206, 0.35);
}

.action-card.secondary {
  background: linear-gradient(135deg, #38a89d 0%, #2c9c8f 100%);
  box-shadow: 0 8px 32px rgba(56, 168, 157, 0.25);
}

.action-card.secondary:hover {
  box-shadow: 0 16px 48px rgba(56, 168, 157, 0.35);
}

.card-glow {
  position: absolute;
  top: -50%;
  right: -50%;
  width: 100%;
  height: 100%;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.15) 0%, transparent 70%);
  opacity: 0;
  transition: opacity 0.4s;
}

.action-card:hover .card-glow {
  opacity: 1;
}

.card-inner {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 28px 24px;
  position: relative;
  z-index: 1;
}

.card-icon-wrapper {
  width: 72px;
  height: 72px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.card-icon {
  font-size: 36px;
}

.card-content {
  flex: 1;
}

.card-content h3 {
  font-size: 22px;
  font-weight: 700;
  color: white;
  margin: 0 0 8px;
  letter-spacing: 0.5px;
}

.card-content p {
  font-size: 15px;
  color: rgba(255, 255, 255, 0.9);
  margin: 0 0 10px;
  line-height: 1.5;
}

.card-tag {
  display: inline-block;
  padding: 4px 12px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 20px;
  font-size: 12px;
  color: white;
  font-weight: 500;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.card-arrow {
  font-size: 28px;
  color: rgba(255, 255, 255, 0.7);
  transition: all 0.3s;
  flex-shrink: 0;
}

.action-card:hover .card-arrow {
  transform: translateX(8px);
  color: white;
}

/* 应用场景 */
.scenarios-section {
  flex: 1;
  min-height: 0;
}

.scenarios-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 20px;
}

.title-icon {
  font-size: 24px;
}

.scenarios-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.scenario-item {
  display: flex;
  gap: 16px;
  padding: 18px;
  background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
  border-radius: 16px;
  border: 1px solid #e4e7ed;
  transition: all 0.3s ease;
}

.scenario-item:hover {
  border-color: #409eff;
  box-shadow: 0 4px 16px rgba(64, 158, 255, 0.1);
  transform: translateX(4px);
}

.scenario-number {
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, #3182ce 0%, #2b6cb0 100%);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 700;
  color: white;
  flex-shrink: 0;
}

.scenario-info {
  flex: 1;
}

.scenario-info h4 {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 6px;
}

.scenario-info p {
  font-size: 13px;
  color: #606266;
  margin: 0 0 10px;
  line-height: 1.4;
}

.scenario-example {
  display: inline-block;
}

.scenario-example code {
  font-size: 13px;
  color: #409eff;
  background: #ecf5ff;
  padding: 6px 14px;
  border-radius: 8px;
  font-family: 'Monaco', 'Menlo', monospace;
  border: 1px solid #d9ecff;
}

/* 快速开始 */
.quick-start {
  margin-top: auto;
  padding: 16px 20px;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  border: 1px solid rgba(66, 153, 225, 0.2);
  box-shadow: 0 2px 8px rgba(66, 153, 225, 0.08);
}

.quick-start-icon {
  font-size: 24px;
  flex-shrink: 0;
}

.quick-start p {
  font-size: 14px;
  color: #4a5568;
  margin: 0;
  line-height: 1.6;
}

.quick-start strong {
  color: #2b6cb0;
  font-weight: 600;
}
</style>
