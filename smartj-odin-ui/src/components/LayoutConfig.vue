<template>
  <el-drawer
    v-model="drawerVisible"
    :title="$t('layout.configTitle')"
    direction="rtl"
    size="280px"
    :append-to-body="true"
    :z-index="2000"
  >
    <!-- 布局切换 -->
    <el-divider content-position="center">{{ $t('layout.layoutSwitch') }}</el-divider>
    <div class="layout-selector">
      <div
        v-for="layout in layouts"
        :key="layout.value"
        class="layout-thumb"
        :class="{ active: themeStore.layoutMode === layout.value }"
        @click="changeLayout(layout.value)"
      >
        <div class="thumb-inner">
          <!-- 左侧布局：左色块 -->
          <template v-if="layout.value === 'sidebar'">
            <div
              class="t-left"
              style="top: 0; bottom: 0"
            ></div>
          </template>
          <!-- 左顶布局：顶色条 + 左色块 -->
          <template v-else-if="layout.value === 'sidebar-top'">
            <div class="t-top"></div>
            <div
              class="t-left"
              style="top: 10px; bottom: 0"
            ></div>
          </template>
          <!-- 横向布局：顶色条 -->
          <template v-else-if="layout.value === 'horizontal'">
            <div class="t-top"></div>
          </template>
          <!-- 分栏布局：窄左栏 + 次左栏 -->
          <template v-else-if="layout.value === 'columns'">
            <div
              class="t-col"
              style="top: 0; bottom: 0"
            ></div>
            <div
              class="t-col2"
              style="top: 0; bottom: 0"
            ></div>
          </template>
        </div>
        <span>{{ $t(layout.label) }}</span>
      </div>
    </div>

    <!-- 全局主题 -->
    <el-divider content-position="center">{{ $t('layout.globalTheme') }}</el-divider>

    <div class="config-item">
      <span class="config-label">{{ $t('layout.darkMode') }}</span>
      <el-switch
        v-model="themeStore.isDark"
        @change="themeStore.applyTheme()"
      />
    </div>

    <div class="config-item">
      <span class="config-label">{{ $t('layout.watermark') }}</span>
      <el-switch
        v-model="themeStore.watermark"
        @change="themeStore.saveConfig()"
      />
    </div>

    <div class="config-item">
      <span class="config-label">{{ $t('layout.navIcon') }}</span>
      <el-switch
        v-model="themeStore.showNavIcon"
        @change="themeStore.saveConfig()"
      />
    </div>

    <div class="config-item">
      <span class="config-label">{{ $t('layout.tagsView') }}</span>
      <el-switch
        v-model="themeStore.showTagsView"
        @change="themeStore.saveConfig()"
      />
    </div>

    <!-- 导航栏配色 -->
    <el-divider content-position="center">{{ $t('layout.navColor') }}</el-divider>

    <div class="config-item">
      <span class="config-label">{{ $t('layout.startColor') }}</span>
      <el-color-picker
        v-model="themeStore.navBgStart"
        @change="themeStore.applyTheme()"
      />
    </div>

    <div class="config-item">
      <span class="config-label">{{ $t('layout.endColor') }}</span>
      <el-color-picker
        v-model="themeStore.navBgEnd"
        @change="themeStore.applyTheme()"
      />
    </div>

    <!-- 预设方案 -->
    <div class="preset-label">{{ $t('layout.presets') }}</div>
    <div class="preset-list">
      <el-tooltip
        v-for="preset in presets"
        :key="preset.name"
        :content="$t(preset.name)"
        placement="top"
      >
        <div
          class="preset-item"
          :style="{ background: `linear-gradient(90deg, ${preset.start}, ${preset.end})` }"
          @click="applyPreset(preset)"
        />
      </el-tooltip>
    </div>

    <!-- 一键重置 -->
    <el-divider />
    <el-button
      type="warning"
      plain
      style="width: 100%"
      @click="handleReset"
    >
      {{ $t('layout.resetConfig') }}
    </el-button>
  </el-drawer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useThemeStore } from '@/stores/theme'
import { ElMessageBox } from 'element-plus'

const { t } = useI18n()

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ (e: 'update:visible', val: boolean): void }>()

const themeStore = useThemeStore()

const drawerVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
})

const layouts = [
  { value: 'sidebar', label: 'layout.sidebar' },
  { value: 'sidebar-top', label: 'layout.sidebarTop' },
  { value: 'horizontal', label: 'layout.horizontal' },
  { value: 'columns', label: 'layout.columns' },
]

function changeLayout(mode: string) {
  themeStore.layoutMode = mode
  themeStore.saveConfig()
}

const presets = [
  { name: 'layout.presetDefault', start: '#1d1e3a', end: '#2d2f5e' },
  { name: 'layout.presetDeepSpace', start: '#0f2027', end: '#2c5364' },
  { name: 'layout.presetPurple', start: '#2b1055', end: '#d53369' },
  { name: 'layout.presetGreen', start: '#134e5e', end: '#71b280' },
  { name: 'layout.presetCarbon', start: '#232526', end: '#414345' },
  { name: 'layout.presetNavy', start: '#141e30', end: '#243b55' },
]

function applyPreset(preset: { start: string; end: string }) {
  themeStore.navBgStart = preset.start
  themeStore.navBgEnd = preset.end
  themeStore.applyTheme()
}

function handleReset() {
  ElMessageBox.confirm(t('layout.resetConfirm'), t('common.tip'), {
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
    type: 'warning',
  })
    .then(() => {
      themeStore.resetConfig()
    })
    .catch(() => {})
}
</script>

<style scoped>
.config-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
}
.config-label {
  font-size: 14px;
  color: var(--sm-text-regular);
}
.preset-label {
  font-size: 13px;
  color: var(--sm-text-secondary);
  margin-top: 12px;
  margin-bottom: 8px;
}
.preset-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.preset-item {
  width: 60px;
  height: 28px;
  border-radius: 4px;
  cursor: pointer;
  border: 2px solid transparent;
  transition: border-color 0.2s;
}
.preset-item:hover {
  border-color: var(--sm-color-primary);
}
</style>
