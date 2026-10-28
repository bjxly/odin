<template>
  <el-card header="SmTreeSelect 示例">
    <el-row :gutter="24">
      <el-col :span="12">
        <h4>单选模式</h4>
        <p class="example-desc">
          当前选中：
          <el-tag
            v-if="singleValue"
            type="info"
          >
            {{ singleValue }}
          </el-tag>
          <span v-else>无</span>
        </p>
        <SmTreeSelect
          v-model="singleValue"
          :data="orgTree"
          placeholder="请选择部门"
        />
      </el-col>
      <el-col :span="12">
        <h4>多选模式</h4>
        <p class="example-desc">
          当前选中：
          <el-tag
            v-for="v in multiValue"
            :key="v"
            closable
            @close="removeMulti(v)"
          >
            {{ v }}
          </el-tag>
          <span v-if="!multiValue.length">无</span>
        </p>
        <SmTreeSelect
          v-model="multiValue"
          :data="orgTree"
          multiple
          placeholder="请选择部门"
        />
      </el-col>
    </el-row>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import SmTreeSelect from '@/components/SmTreeSelect/index.vue'

const singleValue = ref<string | number | undefined>(undefined)
const multiValue = ref<(string | number)[]>([])

const orgTree = [
  {
    id: 'company',
    label: 'SmartX 科技有限公司',
    children: [
      {
        id: 'tech',
        label: '技术中心',
        children: [
          { id: 'frontend', label: '前端开发组' },
          { id: 'backend', label: '后端开发组' },
          { id: 'qa', label: '质量保障组' },
          { id: 'devops', label: '运维组' },
        ],
      },
      {
        id: 'product',
        label: '产品中心',
        children: [
          { id: 'pm', label: '产品策划组' },
          { id: 'ui', label: 'UI 设计组' },
        ],
      },
      {
        id: 'hr',
        label: '人力资源部',
        children: [
          { id: 'recruit', label: '招聘组' },
          { id: 'training', label: '培训组' },
        ],
      },
      {
        id: 'sales',
        label: '销售部',
        children: [
          { id: 'north', label: '华北区域' },
          { id: 'south', label: '华南区域' },
          { id: 'overseas', label: '海外业务' },
        ],
      },
    ],
  },
]

function removeMulti(id: string | number) {
  multiValue.value = multiValue.value.filter((v) => v !== id)
}
</script>

<style scoped>
.example-desc {
  margin: 8px 0 12px;
  color: var(--el-text-color-regular);
}
h4 {
  margin: 0 0 4px;
}
</style>
