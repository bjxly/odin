<script setup lang="ts">
import { getCurrentInstance } from 'vue'
import { useI18n } from 'vue-i18n'
import { Delete, Plus, Folder } from '@element-plus/icons-vue'
import type { GroupNode } from './conditionTree'
import { newGroup, newCond } from './conditionTree'

/**
 * 递归条件组编辑器（AND / OR / NOT）。
 *
 * 支持子条件与嵌套子组的增删；本组件直接原地修改传入的 group 响应式对象
 * （递归子组件共享同一实例，避免深层 v-model 样板代码）。
 */
const props = defineProps<{
  group: GroupNode
  depth: number
  propertyOptions: { value: string; label: string }[]
  operators: string[]
}>()

const { t } = useI18n()

/** 递归自引用（Vue 3.15+ 稳定 API） */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const self: any = getCurrentInstance()!.type

function addCond() {
  props.group.items.push(newCond())
}

function addGroup() {
  props.group.groups.push(newGroup())
}

function removeCond(idx: number) {
  props.group.items.splice(idx, 1)
}

function removeGroup(idx: number) {
  props.group.groups.splice(idx, 1)
}
</script>

<template>
  <div class="cond-group" :class="{ 'is-root': depth === 0 }">
    <div class="cond-group__head">
      <el-select v-model="group.logic" size="small" class="cond-group__logic">
        <el-option :label="t('odinOntology.ruleLogicAnd')" value="and" />
        <el-option :label="t('odinOntology.ruleLogicOr')" value="or" />
        <el-option :label="t('odinOntology.ruleLogicNot')" value="not" />
      </el-select>
      <template v-if="depth > 0">
        <span class="cond-group__hint">{{ t('odinOntology.ruleCondGroup') }}</span>
        <span class="cond-group__spacer" />
      </template>
      <el-button size="small" :icon="Plus" @click="addCond">{{ t('odinOntology.ruleAddCond') }}</el-button>
      <el-button
        v-if="depth < 2"
        size="small"
        :icon="Folder"
        @click="addGroup"
      >
        {{ t('odinOntology.ruleAddGroup') }}
      </el-button>
    </div>

    <el-alert
      v-if="!group.items.length && !group.groups.length"
      type="info"
      :closable="false"
      :title="t('odinOntology.ruleCondEmptyTip')"
      class="cond-group__empty"
    />

    <div v-for="(c, i) in group.items" :key="c.id" class="cond-row">
      <el-select
        v-model="c.property"
        filterable
        allow-create
        default-first-option
        size="small"
        class="cond-row__prop"
        :placeholder="t('odinOntology.ruleCondProperty')"
      >
        <el-option v-for="p in propertyOptions" :key="p.value" :label="p.label" :value="p.value" />
      </el-select>
      <el-select v-model="c.operator" size="small" class="cond-row__op">
        <el-option v-for="op in operators" :key="op" :label="op" :value="op" />
      </el-select>
      <el-input
        v-model="c.value"
        size="small"
        class="cond-row__val"
        :maxlength="128"
        :placeholder="t('odinOntology.ruleCondValue')"
      />
      <el-button size="small" :icon="Delete" text type="danger" @click="removeCond(i)" />
    </div>

    <div v-for="(g, i) in group.groups" :key="g.id" class="cond-group__nested">
      <component :is="self" :group="g" :depth="depth + 1" :property-options="propertyOptions" :operators="operators" />
      <el-button
        class="cond-group__remove"
        size="small"
        :icon="Delete"
        text
        type="danger"
        @click="removeGroup(i)"
      >
        {{ t('common.delete') }}
      </el-button>
    </div>
  </div>
</template>

<style scoped>
.cond-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.cond-group.is-root {
  width: 100%;
}

.cond-group__head {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cond-group__logic {
  width: 88px;
  flex-shrink: 0;
}

.cond-group__hint {
  font-size: 12px;
  color: var(--sm-text-secondary, #909399);
}

.cond-group__spacer {
  flex: 1;
}

.cond-group__empty {
  margin: 0;
}

.cond-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cond-row__prop {
  flex: 1;
  min-width: 0;
}

.cond-row__op {
  width: 96px;
  flex-shrink: 0;
}

.cond-row__val {
  flex: 1;
  min-width: 0;
}

.cond-group__nested {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px 12px;
  background: color-mix(in srgb, var(--sm-bg-page, #f5f7fa) 60%, transparent);
  border: 1px dashed var(--sm-border-base, #dcdfe6);
  border-radius: 8px;
}

.cond-group__remove {
  align-self: flex-start;
}
</style>
