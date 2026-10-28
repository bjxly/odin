<template>
  <div class="map-dashboard">
    <!-- 地图容器 -->
    <div
      ref="mapContainer"
      class="map-dashboard__map"
    ></div>

    <!-- 顶部标题栏 -->
    <div class="map-dashboard__header">
      <div class="header__title">中国数据大屏</div>
      <div class="header__time">{{ currentTime }}</div>
    </div>

    <!-- 左侧面板 -->
    <div class="map-dashboard__left">
      <!-- 统计卡片 -->
      <div class="panel">
        <div class="panel__title">核心指标</div>
        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-card__value">{{ animatedPopulation }}</div>
            <div class="stat-card__label">总人口(万)</div>
          </div>
          <div class="stat-card">
            <div class="stat-card__value">{{ animatedGdp }}</div>
            <div class="stat-card__label">GDP(万亿)</div>
          </div>
          <div class="stat-card">
            <div class="stat-card__value">34</div>
            <div class="stat-card__label">省级行政区</div>
          </div>
          <div class="stat-card">
            <div class="stat-card__value">333</div>
            <div class="stat-card__label">地级市</div>
          </div>
        </div>
      </div>

      <!-- 饼图 -->
      <div class="panel">
        <div class="panel__title">区域人口分布</div>
        <div
          ref="pieChartRef"
          class="chart-container"
        ></div>
      </div>

      <!-- 折线图 -->
      <div class="panel">
        <div class="panel__title">GDP增长趋势</div>
        <div
          ref="lineChartRef"
          class="chart-container"
        ></div>
      </div>
    </div>

    <!-- 右侧面板 -->
    <div class="map-dashboard__right">
      <div class="panel">
        <div class="panel__title">GDP省份TOP10</div>
        <div
          ref="barChartRef"
          class="chart-container chart-container--tall"
        ></div>
      </div>
    </div>

    <!-- 图层切换 -->
    <div class="map-dashboard__controls">
      <div class="layer-switcher">
        <button
          v-for="layer in layers"
          :key="layer.value"
          :class="['layer-btn', { 'layer-btn--active': currentLayer === layer.value }]"
          @click="handleLayerChange(layer.value)"
        >
          {{ layer.label }}
        </button>
      </div>
    </div>

    <!-- 缩放控件 -->
    <div class="map-dashboard__zoom">
      <button
        class="zoom-btn"
        @click="zoomIn"
      >
        +
      </button>
      <button
        class="zoom-btn"
        @click="zoomOut"
      >
        −
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import * as echarts from 'echarts'
import { useMapLibre, type TiandituLayerType } from './useMapLibre'

const mapContainer = ref<HTMLElement>()
const pieChartRef = ref<HTMLElement>()
const lineChartRef = ref<HTMLElement>()
const barChartRef = ref<HTMLElement>()

const currentTime = ref('')
const currentLayer = ref<TiandituLayerType>('img')
const animatedPopulation = ref(0)
const animatedGdp = ref(0)

let pieChart: echarts.ECharts | null = null
let lineChart: echarts.ECharts | null = null
let barChart: echarts.ECharts | null = null
let timeTimer: number | null = null

const layers = [
  { label: '矢量', value: 'vec' as TiandituLayerType },
  { label: '卫星', value: 'img' as TiandituLayerType },
  { label: '地形', value: 'ter' as TiandituLayerType },
]

const { initMap, switchLayer, zoomIn, zoomOut } = useMapLibre(mapContainer, {
  center: [104, 35],
  zoom: 4,
})

// 模拟数据
const provinceData = [
  { name: '广东', gdp: 13.5, population: 12600 },
  { name: '江苏', gdp: 12.8, population: 8500 },
  { name: '山东', gdp: 9.2, population: 10100 },
  { name: '浙江', gdp: 8.3, population: 6500 },
  { name: '河南', gdp: 6.1, population: 9900 },
  { name: '四川', gdp: 5.7, population: 8400 },
  { name: '湖北', gdp: 5.5, population: 5800 },
  { name: '福建', gdp: 5.4, population: 4200 },
  { name: '湖南', gdp: 4.9, population: 6600 },
  { name: '安徽', gdp: 4.7, population: 6100 },
]

const regionPopulation = [
  { name: '华东', value: 42000 },
  { name: '华南', value: 18000 },
  { name: '华北', value: 17000 },
  { name: '华中', value: 22000 },
  { name: '西南', value: 19000 },
  { name: '西北', value: 10000 },
  { name: '东北', value: 9600 },
]

const gdpTrend = {
  years: ['2019', '2020', '2021', '2022', '2023'],
  values: [98.6, 101.4, 114.9, 121.0, 126.0],
}

function updateTime() {
  const now = new Date()
  currentTime.value = now.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function animateNumber(target: number, duration: number, callback: (val: number) => void) {
  const start = 0
  const startTime = performance.now()
  function update(currentTime: number) {
    const elapsed = currentTime - startTime
    const progress = Math.min(elapsed / duration, 1)
    const value = Math.floor(start + (target - start) * progress)
    callback(value)
    if (progress < 1) {
      requestAnimationFrame(update)
    }
  }
  requestAnimationFrame(update)
}

function initPieChart() {
  if (!pieChartRef.value) return
  pieChart = echarts.init(pieChartRef.value)
  pieChart.setOption({
    backgroundColor: 'transparent',
    tooltip: { trigger: 'item', formatter: '{b}: {c}万 ({d}%)' },
    legend: {
      orient: 'vertical',
      right: 10,
      top: 'center',
      textStyle: { color: '#a0aec0', fontSize: 11 },
    },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        center: ['35%', '50%'],
        avoidLabelOverlap: false,
        itemStyle: { borderRadius: 4, borderColor: '#0a1628', borderWidth: 2 },
        label: { show: false },
        emphasis: {
          label: { show: true, fontSize: 12, fontWeight: 'bold', color: '#fff' },
        },
        data: regionPopulation.map((item, index) => ({
          ...item,
          itemStyle: {
            color: ['#667eea', '#764ba2', '#f093fb', '#4facfe', '#43e97b', '#fa709a', '#fee140'][index],
          },
        })),
      },
    ],
  })
}

function initLineChart() {
  if (!lineChartRef.value) return
  lineChart = echarts.init(lineChartRef.value)
  lineChart.setOption({
    backgroundColor: 'transparent',
    tooltip: { trigger: 'axis', formatter: '{b}年: {c}万亿' },
    grid: { left: 50, right: 20, top: 20, bottom: 30 },
    xAxis: {
      type: 'category',
      data: gdpTrend.years,
      axisLine: { lineStyle: { color: '#2d3748' } },
      axisLabel: { color: '#a0aec0' },
    },
    yAxis: {
      type: 'value',
      axisLine: { show: false },
      splitLine: { lineStyle: { color: '#2d3748' } },
      axisLabel: { color: '#a0aec0' },
    },
    series: [
      {
        type: 'line',
        data: gdpTrend.values,
        smooth: true,
        symbol: 'circle',
        symbolSize: 8,
        lineStyle: { color: '#4facfe', width: 3 },
        itemStyle: { color: '#4facfe', borderColor: '#fff', borderWidth: 2 },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(79, 172, 254, 0.4)' },
            { offset: 1, color: 'rgba(79, 172, 254, 0.05)' },
          ]),
        },
      },
    ],
  })
}

function initBarChart() {
  if (!barChartRef.value) return
  barChart = echarts.init(barChartRef.value)
  const top10 = provinceData.slice(0, 10).reverse()
  barChart.setOption({
    backgroundColor: 'transparent',
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' }, formatter: '{b}: {c}万亿' },
    grid: { left: 70, right: 20, top: 10, bottom: 30 },
    xAxis: {
      type: 'value',
      axisLine: { show: false },
      splitLine: { lineStyle: { color: '#2d3748' } },
      axisLabel: { color: '#a0aec0' },
    },
    yAxis: {
      type: 'category',
      data: top10.map((d) => d.name),
      axisLine: { lineStyle: { color: '#2d3748' } },
      axisLabel: { color: '#a0aec0' },
    },
    series: [
      {
        type: 'bar',
        data: top10.map((d) => d.gdp),
        barWidth: 16,
        itemStyle: {
          borderRadius: [0, 4, 4, 0],
          color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
            { offset: 0, color: '#667eea' },
            { offset: 1, color: '#764ba2' },
          ]),
        },
      },
    ],
  })
}

function handleLayerChange(layer: TiandituLayerType) {
  currentLayer.value = layer
  switchLayer(layer)
}

function loadProvinceBoundaries(map: any) {
  fetch('https://geo.datav.aliyun.com/areas_v3/bound/100000_full.json')
    .then((res) => res.json())
    .then((geojson) => {
      map.addSource('provinces', {
        type: 'geojson',
        data: geojson,
      })

      map.addLayer({
        id: 'provinces-fill',
        type: 'fill',
        source: 'provinces',
        paint: {
          'fill-color': 'rgba(102, 126, 234, 0.15)',
          'fill-outline-color': 'rgba(102, 126, 234, 0.6)',
        },
      })

      map.addLayer({
        id: 'provinces-line',
        type: 'line',
        source: 'provinces',
        paint: {
          'line-color': 'rgba(102, 126, 234, 0.8)',
          'line-width': 1.5,
        },
      })

      // Hover 交互
      let hoveredId: string | null = null
      map.on('mousemove', 'provinces-fill', (e: any) => {
        if (e.features && e.features.length > 0) {
          if (hoveredId !== null) {
            map.setFeatureState({ source: 'provinces', id: hoveredId }, { hover: false })
          }
          hoveredId = e.features[0].id as string
          map.setFeatureState({ source: 'provinces', id: hoveredId }, { hover: true })
          map.getCanvas().style.cursor = 'pointer'
        }
      })

      map.on('mouseleave', 'provinces-fill', () => {
        if (hoveredId !== null) {
          map.setFeatureState({ source: 'provinces', id: hoveredId }, { hover: false })
        }
        hoveredId = null
        map.getCanvas().style.cursor = ''
      })
    })
    .catch((err) => console.error('Failed to load province boundaries:', err))
}

onMounted(() => {
  updateTime()
  timeTimer = window.setInterval(updateTime, 1000)

  const map = initMap('img')

  map.on('load', () => {
    loadProvinceBoundaries(map)
  })

  initPieChart()
  initLineChart()
  initBarChart()

  animateNumber(141000, 2000, (val) => {
    animatedPopulation.value = val
  })
  animateNumber(126, 2000, (val) => {
    animatedGdp.value = val
  })
})

onBeforeUnmount(() => {
  if (timeTimer) clearInterval(timeTimer)
  pieChart?.dispose()
  lineChart?.dispose()
  barChart?.dispose()
})
</script>

<style scoped>
.map-dashboard {
  position: relative;
  width: 100%;
  height: 90vh;
  overflow: hidden;
  background: #0a1628;
}

.map-dashboard__map {
  position: absolute;
  inset: 0;
  z-index: 0;
}

.map-dashboard__header {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  background: linear-gradient(180deg, rgba(10, 22, 40, 0.95) 0%, rgba(10, 22, 40, 0) 100%);
  z-index: 10;
}

.header__title {
  font-size: 20px;
  font-weight: 700;
  color: #fff;
  letter-spacing: 2px;
}

.header__time {
  font-size: 14px;
  color: #a0aec0;
  font-family: 'Courier New', monospace;
}

.map-dashboard__left {
  position: absolute;
  top: 70px;
  left: 16px;
  width: 320px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  z-index: 10;
}

.map-dashboard__right {
  position: absolute;
  top: 70px;
  right: 16px;
  width: 320px;
  z-index: 10;
}

.panel {
  background: rgba(10, 22, 40, 0.85);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(102, 126, 234, 0.2);
  border-radius: 8px;
  padding: 16px;
}

.panel__title {
  font-size: 14px;
  font-weight: 600;
  color: #e2e8f0;
  margin-bottom: 12px;
  padding-left: 8px;
  border-left: 3px solid #667eea;
}

.stats-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.stat-card {
  text-align: center;
  padding: 12px 8px;
  background: rgba(102, 126, 234, 0.08);
  border-radius: 6px;
}

.stat-card__value {
  font-size: 22px;
  font-weight: 700;
  color: #4facfe;
  font-family: 'Courier New', monospace;
}

.stat-card__label {
  font-size: 11px;
  color: #a0aec0;
  margin-top: 4px;
}

.chart-container {
  width: 100%;
  height: 180px;
}

.chart-container--tall {
  height: 400px;
}

.map-dashboard__controls {
  position: absolute;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 10;
}

.layer-switcher {
  display: flex;
  gap: 8px;
  background: rgba(10, 22, 40, 0.85);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(102, 126, 234, 0.2);
  border-radius: 8px;
  padding: 8px 12px;
}

.layer-btn {
  padding: 6px 16px;
  font-size: 13px;
  color: #a0aec0;
  background: transparent;
  border: 1px solid rgba(102, 126, 234, 0.3);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.layer-btn:hover {
  color: #fff;
  border-color: #667eea;
}

.layer-btn--active {
  color: #fff;
  background: #667eea;
  border-color: #667eea;
}

.map-dashboard__zoom {
  position: absolute;
  right: 16px;
  bottom: 24px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  z-index: 10;
}

.zoom-btn {
  width: 36px;
  height: 36px;
  font-size: 20px;
  font-weight: 700;
  color: #fff;
  background: rgba(10, 22, 40, 0.85);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(102, 126, 234, 0.2);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.zoom-btn:hover {
  background: #667eea;
  border-color: #667eea;
}

@media (max-width: 1200px) {
  .map-dashboard__left,
  .map-dashboard__right {
    display: none;
  }
}
</style>
