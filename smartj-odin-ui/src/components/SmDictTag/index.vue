<template>
  <el-tag
    v-if="label"
    :type="tagType"
    :effect="effect"
    disable-transitions
  >
    {{ label }}
  </el-tag>
  <span v-else>{{ value }}</span>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useDictStore } from './dictStore'

const props = withDefaults(
  defineProps<{
    dictType: string
    value: string | number
    effect?: 'dark' | 'light' | 'plain'
  }>(),
  {
    effect: 'light',
  }
)

const dictStore = useDictStore()

onMounted(() => {
  dictStore.loadDict(props.dictType)
})

const label = computed(() => dictStore.getDictLabel(props.dictType, props.value))
const tagType = computed(() => dictStore.getDictTagType(props.dictType, props.value) as any)
</script>
