<template>
  <el-card header="SmAvatarCrop 示例">
    <el-row
      :gutter="24"
      align="middle"
    >
      <el-col :span="8">
        <h4>点击裁剪头像</h4>
        <SmAvatarCrop
          v-model="avatarBase64"
          :size="128"
          shape="circle"
        />
      </el-col>
      <el-col :span="16">
        <h4>预览</h4>
        <div
          v-if="avatarBase64"
          class="preview-row"
        >
          <el-avatar
            :size="64"
            :src="avatarBase64"
          />
          <el-avatar
            :size="48"
            :src="avatarBase64"
          />
          <el-avatar
            :size="32"
            :src="avatarBase64"
          />
        </div>
        <el-empty
          v-else
          description="尚未上传头像"
          :image-size="60"
        />

        <h4 style="margin-top: 16px">Base64 值</h4>
        <el-input
          type="textarea"
          :model-value="avatarBase64 || '（空）'"
          :rows="4"
          readonly
          placeholder="裁剪后将显示 base64 数据"
        />
      </el-col>
    </el-row>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import SmAvatarCrop from '@/components/SmAvatarCrop/index.vue'

const avatarBase64 = ref('')
</script>

<style scoped>
h4 {
  margin: 0 0 12px;
}
.preview-row {
  display: flex;
  align-items: flex-end;
  gap: 16px;
}
</style>
