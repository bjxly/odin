<template>
  <el-config-provider :locale="elementLocale">
    <router-view />
  </el-config-provider>
</template>

<script setup lang="ts">
import { ElConfigProvider } from 'element-plus'
import { useElementLocale } from '@/composables/useElementLocale'

const { elementLocale } = useElementLocale()
</script>

<style>
html,
body,
#app {
  background-color: #ffffff;
}
</style>
