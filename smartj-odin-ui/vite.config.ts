import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import { visualizer } from 'rollup-plugin-visualizer'
import viteCompression from 'vite-plugin-compression'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig(({ mode }) => {
  return {
    plugins: [
      vue(),
      // 构建产物分析（仅 build 时生效）
      visualizer({
        open: false,
        gzipSize: true,
        brotliSize: true,
        filename: 'dist/stats.html',
      }),
      // Gzip 压缩（>10KB 的文件）
      viteCompression({
        algorithm: 'gzip',
        ext: '.gz',
        threshold: 10240,
        deleteOriginalAssets: false,
      }),
      // Brotli 压缩（压缩率更高，>10KB 的文件）
      viteCompression({
        algorithm: 'brotliCompress',
        ext: '.br',
        threshold: 10240,
        deleteOriginalAssets: false,
      }),
      // PWA 离线支持
      VitePWA({
        registerType: 'autoUpdate',
        // 预缓存大小上限保持默认 2MiB：超限大 chunk 不入 precache，仅告警不中断构建，运行时按需缓存
        showMaximumFileSizeToCacheInBytesWarning: true,
        manifest: {
          name: 'SmartJOne',
          short_name: 'SmartJOne',
          description: 'SmartJOne —— 企业级 Java 快速开发底座',
          theme_color: '#409eff',
          icons: [
            { src: '/pwa-192x192.png', sizes: '192x192', type: 'image/png' },
            { src: '/pwa-512x512.png', sizes: '512x512', type: 'image/png' },
          ],
          display: 'standalone',
          start_url: '/',
        },
        workbox: {
          globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
          // 忽略构建分析产物（非应用资源，不参与预缓存）
          globIgnores: ['stats.html'],
          runtimeCaching: [
            {
              // SSE 流式对话：NetworkOnly，禁止 Service Worker 缓存/缓冲，否则逐 token 流式会被打断
              urlPattern: /\/api\/odin\/assistant\/chat\/stream/,
              handler: 'NetworkOnly',
            },
            {
              urlPattern: /^\/api\//,
              handler: 'NetworkFirst',
              options: {
                cacheName: 'api-cache',
                cacheableResponse: { statuses: [0, 200] },
              },
            },
          ],
        },
      }),
    ],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, 'src'),
      },
      dedupe: ['vue', 'pinia', 'element-plus', '@vue-flow/core', '@vue-flow/background', '@vue-flow/controls', '@vue-flow/minimap'],
    },
    server: {
      port: 3200,
      open: true,
      proxy: {
        // SSO 社交登录 OAuth 跳转（页面级导航，仅匹配 /sso/social/{idpType} 路径）
        '^/sso/social/(?!callback)': {
          target: 'http://localhost:8000',
          changeOrigin: true,
        },
        '/api': {
          target: 'http://localhost:9000',
          changeOrigin: true,
        },
      },
    },
  }
})
