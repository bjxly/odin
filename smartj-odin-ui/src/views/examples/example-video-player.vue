<template>
  <el-card header="SmVideoPlayer 示例">
    <el-row :gutter="24">
      <el-col :span="16">
        <h4>视频播放器</h4>
        <SmVideoPlayer
          :src="videoSrc"
          :poster="posterSrc"
          :autoplay="false"
          :muted="false"
        />
      </el-col>
      <el-col :span="8">
        <h4>配置信息</h4>
        <el-descriptions
          :column="1"
          border
          size="small"
        >
          <el-descriptions-item label="视频地址">
            <el-text truncated>{{ videoSrc }}</el-text>
          </el-descriptions-item>
          <el-descriptions-item label="自动播放">否</el-descriptions-item>
          <el-descriptions-item label="静音">否</el-descriptions-item>
          <el-descriptions-item label="格式">MP4</el-descriptions-item>
        </el-descriptions>

        <h4 style="margin-top: 16px">操作</h4>
        <el-space
          direction="vertical"
          fill
          style="width: 100%"
        >
          <el-button @click="videoSrc = 'https://www.w3schools.com/html/mov_bbb.mp4'">加载视频 A（Big Buck Bunny）</el-button>
          <el-button @click="videoSrc = 'https://www.w3schools.com/html/movie.mp4'">加载视频 B（Sample）</el-button>
        </el-space>
      </el-col>
    </el-row>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import SmVideoPlayer from '@/components/SmVideoPlayer/index.vue'

const videoSrc = ref('https://www.w3schools.com/html/mov_bbb.mp4')
const posterSrc = ref('')
</script>

<style scoped>
h4 {
  margin: 0 0 12px;
}
</style>
