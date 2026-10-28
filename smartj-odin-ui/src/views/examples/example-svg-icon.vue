<template>
  <div class="example-page">
    <h2>SmSvgIcon 示例</h2>
    <p class="example-desc">展示 Element Plus 图标的不同尺寸与颜色组合。</p>

    <el-card shadow="never">
      <template #header>
        <span>图标网格 - 不同尺寸</span>
      </template>
      <el-row :gutter="24">
        <el-col
          v-for="icon in icons"
          :key="icon.name + icon.size"
          :span="4"
          class="icon-cell"
        >
          <div class="icon-item">
            <SmSvgIcon
              :name="icon.name"
              :size="icon.size"
              :color="icon.color"
            />
            <div class="icon-label">{{ icon.name }}</div>
            <div class="icon-meta">{{ icon.size }}px</div>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <el-card
      shadow="never"
      style="margin-top: 16px"
    >
      <template #header>
        <span>图标网格 - 不同颜色</span>
      </template>
      <el-row :gutter="24">
        <el-col
          v-for="item in colorSamples"
          :key="item.color"
          :span="4"
          class="icon-cell"
        >
          <div class="icon-item">
            <SmSvgIcon
              name="Timer"
              :size="28"
              :color="item.color"
            />
            <div class="icon-label">{{ item.label }}</div>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <el-card
      shadow="never"
      style="margin-top: 16px"
    >
      <template #header>
        <span>自定义尺寸</span>
      </template>
      <div class="size-demo">
        <div
          v-for="s in [14, 20, 28, 36, 48]"
          :key="s"
          class="size-item"
        >
          <SmSvgIcon
            name="Setting"
            :size="s"
            color="var(--el-color-primary)"
          />
          <span class="size-label">{{ s }}px</span>
        </div>
      </div>
    </el-card>

    <el-card
      shadow="never"
      style="margin-top: 16px"
    >
      <template #header>
        <span>iconfont font-class 模式</span>
      </template>
      <p class="example-desc">默认读取 public/iconfont/iconfont.css，也可通过 VITE_ICONFONT_CSS_URL 覆盖；name 以 ifc- 开头。</p>
      <div class="size-demo">
        <div class="size-item">
          <SmSvgIcon
            name="ifc-feishu"
            :size="24"
            color="var(--el-color-primary)"
          />
          <span class="size-label">ifc-feishu</span>
        </div>
        <div class="size-item">
          <SmSvgIcon
            name="ifc-github-fill"
            :size="32"
            color="var(--el-color-success)"
          />
          <span class="size-label">ifc-github-fill</span>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import SmSvgIcon from '@/components/SmSvgIcon/index.vue'

const icons = [
  { name: 'Timer', size: 16, color: '' },
  { name: 'Setting', size: 16, color: '' },
  { name: 'Search', size: 16, color: '' },
  { name: 'Timer', size: 24, color: '' },
  { name: 'Setting', size: 24, color: '' },
  { name: 'Search', size: 24, color: '' },
  { name: 'Timer', size: 32, color: 'var(--el-color-primary)' },
  { name: 'Setting', size: 32, color: 'var(--el-color-success)' },
  { name: 'Search', size: 32, color: 'var(--el-color-danger)' },
]

const colorSamples = [
  { label: '默认', color: '' },
  { label: 'Primary', color: 'var(--el-color-primary)' },
  { label: 'Success', color: 'var(--el-color-success)' },
  { label: 'Warning', color: 'var(--el-color-warning)' },
  { label: 'Danger', color: 'var(--el-color-danger)' },
  { label: 'Info', color: 'var(--el-color-info)' },
]
</script>

<style scoped>
.example-page {
  height: 100%;
  overflow-y: auto;
  padding: 20px;
  box-sizing: border-box;
}
.example-desc {
  color: var(--el-text-color-secondary);
  margin-bottom: 16px;
}
.icon-cell {
  margin-bottom: 16px;
}
.icon-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px 0;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
}
.icon-label {
  margin-top: 8px;
  font-size: 12px;
  color: var(--el-text-color-regular);
}
.icon-meta {
  font-size: 11px;
  color: var(--el-text-color-placeholder);
}
.size-demo {
  display: flex;
  align-items: flex-end;
  gap: 24px;
}
.size-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}
.size-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
