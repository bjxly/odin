<template>
  <div class="sm-page">
    <div class="sm-card">
      <h2 class="text-lg font-semibold mb-4">SmCountTo 数字滚动动画示例</h2>
      <p class="text-sm text-sm-text-secondary mb-4">使用 requestAnimationFrame 实现流畅的数字滚动动画。</p>

      <div class="grid grid-cols-4 gap-4 mb-6">
        <div class="text-center p-4 rounded bg-sm-bg-page">
          <div class="text-2xl font-bold text-sm-color-primary">
            <SmCountTo
              :end-val="12580"
              :duration="2000"
              separator=","
            />
          </div>
          <div class="text-xs text-sm-text-secondary mt-2">普通数字</div>
        </div>
        <div class="text-center p-4 rounded bg-sm-bg-page">
          <div class="text-2xl font-bold text-sm-color-success">
            <SmCountTo
              :end-val="88.56"
              :duration="2000"
              :decimals="2"
              suffix="%"
            />
          </div>
          <div class="text-xs text-sm-text-secondary mt-2">带小数 + 后缀</div>
        </div>
        <div class="text-center p-4 rounded bg-sm-bg-page">
          <div class="text-2xl font-bold text-sm-color-warning">
            <SmCountTo
              ref="countRef1"
              :end-val="9999"
              :duration="3000"
              prefix="$ "
              :autoplay="false"
            />
          </div>
          <div class="text-xs text-sm-text-secondary mt-2">前缀 + 手动触发</div>
        </div>
        <div class="text-center p-4 rounded bg-sm-bg-page">
          <div class="text-2xl font-bold text-sm-color-danger">
            <SmCountTo
              :end-val="1234567"
              :duration="2500"
              separator=","
            />
          </div>
          <div class="text-xs text-sm-text-secondary mt-2">大数字千分位</div>
        </div>
      </div>

      <div class="flex gap-3 mb-4">
        <el-button
          type="primary"
          @click="replay"
        >
          重新播放全部
        </el-button>
        <el-button @click="triggerManual">手动触发 $</el-button>
      </div>

      <div class="flex gap-3 flex-wrap">
        <el-input-number
          v-model="targetVal"
          :min="0"
          :max="999999"
          class="w-40"
        />
        <el-button
          type="primary"
          @click="applyValue"
        >
          应用到第一个
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import SmCountTo from '@/components/SmCountTo/index.vue'

const countRef1 = ref<InstanceType<typeof SmCountTo>>()
const targetVal = ref(5000)

function replay() {
  // Force re-render by toggling key (handled by parent via v-if)
  location.reload()
}

function triggerManual() {
  countRef1.value?.start()
}

function applyValue() {
  // Changing endVal will auto-trigger replay due to the watch
  location.reload()
}
</script>
