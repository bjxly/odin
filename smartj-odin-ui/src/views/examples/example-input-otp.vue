<template>
  <div class="sm-page">
    <div class="sm-card">
      <h2 class="text-lg font-semibold mb-4">SmInputOtp 验证码输入组件示例</h2>
      <p class="text-sm text-sm-text-secondary mb-4">支持 4/6 位验证码输入、自动聚焦跳转、粘贴填充、键盘导航，参考 NaiveUI InputOtp 设计。</p>

      <!-- 6 位数字验证码 -->
      <h3 class="text-base font-medium mt-6 mb-3">6 位数字验证码（默认）</h3>
      <SmInputOtp v-model="otp1" @complete="onComplete" />
      <div class="mt-2 text-sm">
        <span v-if="otp1.length === 6" class="text-sm-color-success">✓ 输入完成：{{ otp1 }}</span>
        <span v-else class="text-sm-text-secondary">请输入 6 位数字</span>
      </div>

      <!-- 4 位验证码 -->
      <h3 class="text-base font-medium mt-6 mb-3">4 位验证码</h3>
      <SmInputOtp v-model="otp2" :length="4" @complete="onComplete" />
      <div class="mt-2 text-sm">
        <span v-if="otp2.length === 4" class="text-sm-color-success">✓ 输入完成：{{ otp2 }}</span>
        <span v-else class="text-sm-text-secondary">请输入 4 位数字</span>
      </div>

      <!-- 不同大小 -->
      <h3 class="text-base font-medium mt-6 mb-3">不同大小</h3>
      <div class="flex flex-col gap-4">
        <SmInputOtp v-model="otp3" :length="4" size="36px" :gap="6" />
        <SmInputOtp v-model="otp3" :length="4" size="48px" :gap="10" />
        <SmInputOtp v-model="otp3" :length="4" size="56px" :gap="12" />
      </div>

      <!-- 支持字母 -->
      <h3 class="text-base font-medium mt-6 mb-3">支持字母 + 数字</h3>
      <SmInputOtp v-model="otp4" :length="6" :numeric-only="false" />
      <div class="mt-2 text-sm text-sm-text-secondary">
        当前输入：{{ otp4 || '(空)' }}
      </div>

      <!-- 禁用状态 -->
      <h3 class="text-base font-medium mt-6 mb-3">禁用状态</h3>
      <SmInputOtp v-model="otp5" disabled />

      <!-- 通过 ref 控制 -->
      <h3 class="text-base font-medium mt-6 mb-3">通过 ref 控制</h3>
      <div class="flex gap-3 mb-4">
        <el-button size="small" @click="otpRef?.focus()">聚焦</el-button>
        <el-button size="small" @click="otpRef?.clear()">清空</el-button>
        <el-button size="small" type="danger" @click="otpRef?.setError('验证码错误，请重试')">
          设置错误
        </el-button>
      </div>
      <SmInputOtp ref="otpRef" v-model="otp6" :length="6" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import SmInputOtp from '@/components/SmInputOtp/index.vue'

const otp1 = ref('')
const otp2 = ref('')
const otp3 = ref('1234')
const otp4 = ref('ab12cd')
const otp5 = ref('987654')
const otp6 = ref('')
const otpRef = ref<InstanceType<typeof SmInputOtp>>()

function onComplete(val: string) {
  console.log('OTP 输入完成:', val)
}
</script>
