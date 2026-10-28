<template>
  <div class="examples-page">
    <h2 class="examples-title">组件示例</h2>
    <el-row :gutter="16">
      <el-col
        v-for="item in componentList"
        :key="item.path"
        :xs="24"
        :sm="12"
        :md="8"
        :lg="6"
      >
        <el-card
          class="examples-card"
          shadow="hover"
          @click="router.push(item.path)"
        >
          <div class="examples-card__icon">
            <el-icon :size="32"><component :is="item.icon" /></el-icon>
          </div>
          <div class="examples-card__name">{{ item.name }}</div>
          <div class="examples-card__desc">{{ item.desc }}</div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import {
  Picture,
  Upload,
  PriceTag,
  Search,
  Grid,
  Timer,
  Connection,
  Brush,
  Edit,
  VideoPlay,
  Crop,
  Document,
  Monitor,
  List,
  View,
  Menu,
  DArrowRight,
  DataAnalysis,
  Tickets,
  Coin,
  Setting,
  Star,
  Calendar,
  PictureFilled,
  Iphone,
} from '@element-plus/icons-vue'

const router = useRouter()

const componentList = [
  { name: 'SmSvgIcon', path: '/examples/svg-icon', icon: Picture, desc: 'SVG 图标组件' },
  { name: 'SmDictTag', path: '/examples/dict-tag', icon: PriceTag, desc: '字典标签组件' },
  { name: 'SmSearchForm', path: '/examples/search-form', icon: Search, desc: '搜索表单栏' },
  { name: 'SmCrudTable', path: '/examples/crud-table', icon: Grid, desc: '表格增删改查' },
  { name: 'SmCrontab', path: '/examples/crontab', icon: Timer, desc: 'Cron 表达式选择器' },
  { name: 'SmTreeSelect', path: '/examples/tree-select', icon: Connection, desc: '树形选择器' },
  { name: 'SmFileUpload', path: '/examples/file-upload', icon: Upload, desc: '文件上传增强' },
  { name: 'SmAvatarCrop', path: '/examples/avatar-crop', icon: Crop, desc: '头像裁剪上传' },
  { name: 'SmColorPicker', path: '/examples/color-picker', icon: Brush, desc: '颜色选择器增强' },
  { name: 'SmJsonEditor', path: '/examples/json-editor', icon: Edit, desc: 'JSON 编辑器' },
  { name: 'SmVideoPlayer', path: '/examples/video-player', icon: VideoPlay, desc: '视频播放器' },
  { name: 'SmRichText', path: '/examples/rich-text', icon: Document, desc: '富文本编辑器' },
  { name: 'SmCodeEditor', path: '/examples/code-editor', icon: Monitor, desc: '代码编辑器' },
  { name: 'SmIconSelect', path: '/examples/icon-select', icon: List, desc: '图标选择器' },
  { name: 'SmImagePreview', path: '/examples/image-preview', icon: View, desc: '图片预览' },
  { name: 'SmContextMenu', path: '/examples/context-menu', icon: Menu, desc: '右键菜单' },
  { name: 'SmStepsForm', path: '/examples/steps-form', icon: DArrowRight, desc: '分步表单' },
  { name: 'SmDrawerForm', path: '/examples/drawer-form', icon: DataAnalysis, desc: '抽屉表单' },
  { name: 'SmCountTo', path: '/examples/count-to', icon: Coin, desc: '数字滚动动画' },
  { name: 'SmScrollbar', path: '/examples/scrollbar', icon: Tickets, desc: '统一滚动条' },
  { name: 'SmEmpty', path: '/examples/empty', icon: Setting, desc: '空状态占位' },
  { name: 'SmWatermark', path: '/examples/watermark', icon: Picture, desc: '水印组件' },
  { name: 'SmRate', path: '/examples/rate', icon: Star, desc: '评分组件' },
  { name: 'SmCalendar', path: '/examples/calendar', icon: Calendar, desc: '日历组件' },
  { name: 'SmQRCode', path: '/examples/qr-code', icon: PictureFilled, desc: '二维码生成' },
  { name: 'SmInputOtp', path: '/examples/input-otp', icon: Iphone, desc: '验证码输入' },
]
</script>

<style scoped>
.examples-page {
  height: 100%;
  overflow-y: auto;
  padding: 24px;
  box-sizing: border-box;
}
.examples-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--sm-text-primary);
  margin-bottom: 24px;
}
.examples-card {
  margin-bottom: 16px;
  cursor: pointer;
  transition: transform 0.2s;
  text-align: center;
}
.examples-card:hover {
  transform: translateY(-2px);
}
.examples-card__icon {
  color: var(--sm-color-primary);
  margin-bottom: 12px;
}
.examples-card__name {
  font-size: 15px;
  font-weight: 600;
  color: var(--sm-text-primary);
  margin-bottom: 4px;
}
.examples-card__desc {
  font-size: 13px;
  color: var(--sm-text-secondary);
}
</style>
