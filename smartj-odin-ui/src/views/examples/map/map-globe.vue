<template>
  <div class="map-globe">
    <!-- 地图容器 -->
    <div
      ref="mapContainer"
      class="map-globe__map"
    ></div>

    <!-- 顶部标题栏 -->
    <div class="map-globe__header">
      <div class="header__title">全球数据大屏</div>
      <div class="header__time">{{ currentTime }}</div>
    </div>

    <!-- 左侧面板 -->
    <div class="map-globe__left">
      <!-- 统计卡片 -->
      <div class="panel">
        <div class="panel__title">全球概览</div>
        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-card__value">197</div>
            <div class="stat-card__label">国家/地区</div>
          </div>
          <div class="stat-card">
            <div class="stat-card__value">{{ animatedPopulation }}</div>
            <div class="stat-card__label">人口(亿)</div>
          </div>
          <div class="stat-card">
            <div class="stat-card__value">{{ animatedTrade }}</div>
            <div class="stat-card__label">贸易额(万亿$)</div>
          </div>
          <div class="stat-card">
            <div class="stat-card__value">89</div>
            <div class="stat-card__label">航线城市</div>
          </div>
        </div>
      </div>

      <!-- 环形图 -->
      <div class="panel">
        <div class="panel__title">各大洲GDP占比</div>
        <div
          ref="donutChartRef"
          class="chart-container"
        ></div>
      </div>
    </div>

    <!-- 右侧面板 -->
    <div class="map-globe__right">
      <!-- 折线图 -->
      <div class="panel">
        <div class="panel__title">全球贸易趋势</div>
        <div
          ref="lineChartRef"
          class="chart-container"
        ></div>
      </div>

      <!-- 指标卡片 -->
      <div class="panel">
        <div class="panel__title">核心指标</div>
        <div class="indicator-list">
          <div
            v-for="item in indicators"
            :key="item.label"
            class="indicator-item"
          >
            <div class="indicator-item__label">{{ item.label }}</div>
            <div
              class="indicator-item__value"
              :style="{ color: item.color }"
            >
              {{ item.value }}
            </div>
            <div
              class="indicator-item__trend"
              :class="item.trend > 0 ? 'trend-up' : 'trend-down'"
            >
              {{ item.trend > 0 ? '+' : '' }}{{ item.trend }}%
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 图层切换 -->
    <div class="map-globe__controls">
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
    <div class="map-globe__zoom">
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
const donutChartRef = ref<HTMLElement>()
const lineChartRef = ref<HTMLElement>()

const currentTime = ref('')
const currentLayer = ref<TiandituLayerType>('img')
const animatedPopulation = ref(0)
const animatedTrade = ref(0)

let donutChart: echarts.ECharts | null = null
let lineChart: echarts.ECharts | null = null
let timeTimer: number | null = null
let rotationFrame: number | null = null
let isUserInteracting = false

const layers = [
  { label: '矢量', value: 'vec' as TiandituLayerType },
  { label: '卫星', value: 'img' as TiandituLayerType },
  { label: '地形', value: 'ter' as TiandituLayerType },
]

const indicators = [
  { label: '全球GDP', value: '$105万亿', color: '#4facfe', trend: 3.2 },
  { label: '国际贸易', value: '$32万亿', color: '#43e97b', trend: 5.8 },
  { label: '跨境投资', value: '$1.8万亿', color: '#f093fb', trend: -2.1 },
  { label: '数字经济', value: '$18万亿', color: '#fee140', trend: 12.5 },
]

// 城市坐标
const cities = [
  { name: '北京', coord: [116.4, 39.9] },
  { name: '上海', coord: [121.5, 31.2] },
  { name: '东京', coord: [139.7, 35.7] },
  { name: '纽约', coord: [-74.0, 40.7] },
  { name: '伦敦', coord: [-0.1, 51.5] },
  { name: '新加坡', coord: [103.8, 1.4] },
  { name: '迪拜', coord: [55.3, 25.3] },
  { name: '悉尼', coord: [151.2, -33.9] },
]

// 飞线数据
const flightLines = [
  { from: [116.4, 39.9], to: [-74.0, 40.7] }, // 北京 -> 纽约
  { from: [121.5, 31.2], to: [-0.1, 51.5] }, // 上海 -> 伦敦
  { from: [139.7, 35.7], to: [151.2, -33.9] }, // 东京 -> 悉尼
  { from: [103.8, 1.4], to: [55.3, 25.3] }, // 新加坡 -> 迪拜
  { from: [116.4, 39.9], to: [139.7, 35.7] }, // 北京 -> 东京
  { from: [-74.0, 40.7], to: [-0.1, 51.5] }, // 纽约 -> 伦敦
]

const { initMap, switchLayer, zoomIn, zoomOut, getMap } = useMapLibre(mapContainer, {
  center: [0, 20],
  zoom: 3,
  projection: 'globe',
})

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

function initDonutChart() {
  if (!donutChartRef.value) return
  donutChart = echarts.init(donutChartRef.value)
  donutChart.setOption({
    backgroundColor: 'transparent',
    tooltip: { trigger: 'item', formatter: '{b}: {c}万亿$ ({d}%)' },
    legend: {
      orient: 'vertical',
      right: 10,
      top: 'center',
      textStyle: { color: '#a0aec0', fontSize: 11 },
    },
    series: [
      {
        type: 'pie',
        radius: ['45%', '75%'],
        center: ['35%', '50%'],
        avoidLabelOverlap: false,
        itemStyle: { borderRadius: 6, borderColor: '#0a1628', borderWidth: 2 },
        label: { show: false },
        emphasis: {
          label: { show: true, fontSize: 12, fontWeight: 'bold', color: '#fff' },
        },
        data: [
          { name: '亚洲', value: 38, itemStyle: { color: '#667eea' } },
          { name: '北美', value: 25, itemStyle: { color: '#4facfe' } },
          { name: '欧洲', value: 22, itemStyle: { color: '#43e97b' } },
          { name: '南美', value: 5, itemStyle: { color: '#f093fb' } },
          { name: '非洲', value: 3, itemStyle: { color: '#fa709a' } },
          { name: '大洋洲', value: 2, itemStyle: { color: '#fee140' } },
        ],
      },
    ],
  })
}

function initLineChart() {
  if (!lineChartRef.value) return
  lineChart = echarts.init(lineChartRef.value)
  lineChart.setOption({
    backgroundColor: 'transparent',
    tooltip: { trigger: 'axis' },
    grid: { left: 50, right: 20, top: 20, bottom: 30 },
    xAxis: {
      type: 'category',
      data: ['2019', '2020', '2021', '2022', '2023'],
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
        name: '出口',
        type: 'line',
        data: [19.5, 17.8, 21.2, 24.5, 25.8],
        smooth: true,
        lineStyle: { color: '#4facfe', width: 2 },
        itemStyle: { color: '#4facfe' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(79, 172, 254, 0.3)' },
            { offset: 1, color: 'rgba(79, 172, 254, 0.05)' },
          ]),
        },
      },
      {
        name: '进口',
        type: 'line',
        data: [18.2, 16.5, 20.1, 23.2, 24.5],
        smooth: true,
        lineStyle: { color: '#43e97b', width: 2 },
        itemStyle: { color: '#43e97b' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(67, 233, 123, 0.3)' },
            { offset: 1, color: 'rgba(67, 233, 123, 0.05)' },
          ]),
        },
      },
    ],
  })
}

function generateArcPoints(from: number[], to: number[], numPoints: number = 50): number[][] {
  const points: number[][] = []
  const dist = Math.sqrt(Math.pow(to[0] - from[0], 2) + Math.pow(to[1] - from[1], 2))
  const elevation = dist * 0.3

  for (let i = 0; i <= numPoints; i++) {
    const t = i / numPoints
    const lon = from[0] + (to[0] - from[0]) * t
    const lat = from[1] + (to[1] - from[1]) * t
    const arcHeight = Math.sin(t * Math.PI) * elevation
    points.push([lon, lat, arcHeight])
  }
  return points
}

function addFlightLines(map: any) {
  const lineCoords: number[][] = []
  for (const flight of flightLines) {
    const arcPoints = generateArcPoints(flight.from, flight.to, 40)
    lineCoords.push(...arcPoints, [null as any, null as any, null as any])
  }

  map.addSource('flight-lines', {
    type: 'geojson',
    data: {
      type: 'FeatureCollection',
      features: [
        {
          type: 'Feature',
          geometry: {
            type: 'LineString',
            coordinates: lineCoords.filter((c) => c[0] !== null),
          },
        },
      ],
    },
  })

  map.addLayer({
    id: 'flight-lines-layer',
    type: 'line',
    source: 'flight-lines',
    paint: {
      'line-color': '#4facfe',
      'line-width': 2,
      'line-opacity': 0.8,
    },
  })
}

function addCityMarkers(map: any) {
  map.addSource('cities', {
    type: 'geojson',
    data: {
      type: 'FeatureCollection',
      features: cities.map((city) => ({
        type: 'Feature',
        geometry: {
          type: 'Point',
          coordinates: city.coord,
        },
        properties: {
          name: city.name,
        },
      })),
    },
  })

  // 外圈脉冲
  map.addLayer({
    id: 'cities-pulse',
    type: 'circle',
    source: 'cities',
    paint: {
      'circle-radius': 8,
      'circle-color': 'rgba(79, 172, 254, 0.3)',
      'circle-stroke-width': 2,
      'circle-stroke-color': 'rgba(79, 172, 254, 0.8)',
    },
  })

  // 内圈实心
  map.addLayer({
    id: 'cities-dot',
    type: 'circle',
    source: 'cities',
    paint: {
      'circle-radius': 4,
      'circle-color': '#4facfe',
      'circle-stroke-width': 1,
      'circle-stroke-color': '#fff',
    },
  })
}

function loadWorldBoundaries(map: any) {
  fetch('https://geo.datav.aliyun.com/areas_v3/bound/world.json')
    .then((res) => res.json())
    .then((geojson) => {
      map.addSource('world', {
        type: 'geojson',
        data: geojson,
      })

      map.addLayer(
        {
          id: 'world-fill',
          type: 'fill',
          source: 'world',
          paint: {
            'fill-color': 'rgba(102, 126, 234, 0.12)',
            'fill-outline-color': 'rgba(102, 126, 234, 0.5)',
          },
        },
        'flight-lines-layer'
      )

      map.addLayer(
        {
          id: 'world-line',
          type: 'line',
          source: 'world',
          paint: {
            'line-color': 'rgba(102, 126, 234, 0.6)',
            'line-width': 1,
          },
        },
        'flight-lines-layer'
      )

      // Hover 交互
      let hoveredId: string | null = null
      map.on('mousemove', 'world-fill', (e: any) => {
        if (e.features && e.features.length > 0) {
          if (hoveredId !== null) {
            map.setFeatureState({ source: 'world', id: hoveredId }, { hover: false })
          }
          hoveredId = e.features[0].id as string
          map.setFeatureState({ source: 'world', id: hoveredId }, { hover: true })
          map.getCanvas().style.cursor = 'pointer'
        }
      })

      map.on('mouseleave', 'world-fill', () => {
        if (hoveredId !== null) {
          map.setFeatureState({ source: 'world', id: hoveredId }, { hover: false })
        }
        hoveredId = null
        map.getCanvas().style.cursor = ''
      })
    })
    .catch((err) => console.error('Failed to load world boundaries:', err))
}

function startAutoRotation() {
  const map = getMap()
  if (!map) return

  let bearing = map.getBearing()

  function rotate() {
    if (!isUserInteracting && map) {
      bearing += 0.1
      map.setBearing(bearing % 360)
    }
    rotationFrame = requestAnimationFrame(rotate)
  }

  rotationFrame = requestAnimationFrame(rotate)
}

function handleLayerChange(layer: TiandituLayerType) {
  currentLayer.value = layer
  switchLayer(layer)
}

onMounted(() => {
  updateTime()
  timeTimer = window.setInterval(updateTime, 1000)

  const map = initMap('img')

  map.on('load', () => {
    loadWorldBoundaries(map)
    addFlightLines(map)
    addCityMarkers(map)
    startAutoRotation()
  })

  // 用户交互时暂停旋转
  map.on('mousedown', () => {
    isUserInteracting = true
  })
  map.on('mouseup', () => {
    setTimeout(() => {
      isUserInteracting = false
    }, 3000)
  })

  initDonutChart()
  initLineChart()

  animateNumber(80, 2000, (val) => {
    animatedPopulation.value = val
  })
  animateNumber(32, 2000, (val) => {
    animatedTrade.value = val
  })
})

onBeforeUnmount(() => {
  if (timeTimer) clearInterval(timeTimer)
  if (rotationFrame) cancelAnimationFrame(rotationFrame)
  donutChart?.dispose()
  lineChart?.dispose()
})
</script>

<style scoped>
.map-globe {
  position: relative;
  width: 100%;
  height: 90vh;
  overflow: hidden;
  background: #0a1628;
}

.map-globe__map {
  position: absolute;
  inset: 0;
  z-index: 0;
}

.map-globe__header {
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

.map-globe__left {
  position: absolute;
  top: 70px;
  left: 16px;
  width: 320px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  z-index: 10;
}

.map-globe__right {
  position: absolute;
  top: 70px;
  right: 16px;
  width: 320px;
  display: flex;
  flex-direction: column;
  gap: 12px;
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
  height: 200px;
}

.indicator-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.indicator-item {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  background: rgba(102, 126, 234, 0.06);
  border-radius: 6px;
  border-left: 3px solid transparent;
}

.indicator-item__label {
  flex: 1;
  font-size: 13px;
  color: #a0aec0;
}

.indicator-item__value {
  font-size: 16px;
  font-weight: 700;
  font-family: 'Courier New', monospace;
  margin-right: 12px;
}

.indicator-item__trend {
  font-size: 12px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 4px;
}

.trend-up {
  color: #43e97b;
  background: rgba(67, 233, 123, 0.15);
}

.trend-down {
  color: #fa709a;
  background: rgba(250, 112, 154, 0.15);
}

.map-globe__controls {
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

.map-globe__zoom {
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
  .map-globe__left,
  .map-globe__right {
    display: none;
  }
}
</style>
