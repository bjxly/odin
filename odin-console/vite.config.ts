import { fileURLToPath, URL } from 'node:url'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

/**
 * 第三方依赖分包策略（Vite 8 / rolldown 的 `output.codeSplitting.groups`）。
 *
 * 为什么需要显式分组：main.ts 全量注册 Element Plus（`app.use(ElementPlus)` +
 * 全量图标 + `element-plus/dist/index.css`），首屏必然要加载整个组件库。若不分组，
 * rolldown 的自动分包会把 vendor 与业务代码混进以业务模块命名的共享 chunk
 * （改造前：`index-*.js` 660KB、`_plugin-vue_export-helper-*.js` 273KB），
 * 既触发 >500KB 警告，也让业务代码一改就整包 vendor 缓存失效。
 *
 * ⚠️ 不要对 Element Plus 使用 `maxSize` 按体积硬切：Element Plus 的组件之间存在
 * 大量跨目录引用，按体积装箱会切断这些引用方向，产生**循环 chunk**（ESM 下
 * `components: { ElInput }` 这类模块作用域取值可能触发 TDZ 运行时崩溃）。
 * 下面的三段式切分是按 element-plus 模块依赖图的**拓扑层**划分的：
 * 依赖边的方向恒为「低层 → 高层」，因此按连续层分带后跨 chunk 的边必然单向，
 * 结构上不可能成环（已用依赖图 SCC 分析 + 构建产物 import 图双重验证）。
 *
 * 分层（level 越大越底层，箭头表示依赖方向）：
 *   element-plus       L0-L1  聚合器 es/index.mjs + 顶层组件（table / menu / upload / tour ...）
 *   element-plus-mid   L2-L4  中间层组件（select / tree / dialog / input / button / tooltip ...）
 *   element-plus-base  L4-L9  基础层（utils / hooks / constants / directives / locale
 *                             + popper / scrollbar / config-provider / form / icon 等被广泛复用的底座）
 *
 * 升级 element-plus 大版本后，组件依赖层次可能变化；若怀疑出现循环 chunk，
 * 重新按上述拓扑层规则核对这三个名单即可（`npm run build` 产物中若出现
 * chunk 间互相 import 即为循环）。
 */

/** 基础层：element-plus 的非组件基础设施目录 */
const EP_BASE_ES_DIRS = 'utils|hooks|constants|directives|locale|_virtual'
/** 基础层：被中间层/顶层广泛依赖的底座组件（含 L4 的 directives 同级底座） */
const EP_BASE_COMPONENTS = 'base|collection|config-provider|focus-trap|form|icon|popper|scrollbar|slot'
/** 中间层组件（L2-L4）；顶层组件走 catch-all，无需枚举 */
const EP_MID_COMPONENTS = [
  // L2
  'date-picker-panel', 'select', 'tree', 'cascader-panel', 'color-picker-panel', 'dialog', 'image-viewer',
  'input-number', 'progress', 'empty', 'popover', 'statistic', 'badge', 'row', 'divider', 'checkbox-group',
  'descriptions-item', 'skeleton-item',
  // L3
  'time-picker', 'select-v2', 'dropdown', 'checkbox', 'radio', 'text', 'collapse-transition', 'overlay',
  'option', 'option-group',
  // L4
  'virtual-list', 'input', 'tooltip', 'button', 'roving-focus-group', 'tag', 'button-group',
].join('|')

/** 正则统一用 `[\\/]` 匹配路径分隔符，兼容 Windows */
const NM = 'node_modules[\\\\/]'

const VENDOR_CHUNK_GROUPS = [
  {
    name: 'element-plus-icons',
    test: new RegExp(`${NM}@element-plus[\\\\/]icons-vue[\\\\/]`),
    priority: 60,
  },
  {
    // @vueuse/core 被 Element Plus、Vue Flow 与业务代码共同依赖，单独成组避免被任一方吞并
    name: 'vueuse',
    test: new RegExp(`${NM}@vueuse[\\\\/]`),
    priority: 55,
  },
  {
    // Vue Flow 及其传递依赖（d3-* / @dagrejs），保证图谱页只多加载一个 vendor chunk
    name: 'vue-flow',
    test: new RegExp(`${NM}(@vue-flow[\\\\/]|@dagrejs[\\\\/]|d3-|delaunator|internmap|robust-predicates)`),
    priority: 50,
  },
  {
    // `@vue[\\/]` 不会误匹配 `@vue-flow`（已被更高优先级组吃掉）；`vue[\\/]` 不会误匹配 `vue-router`
    name: 'vue-vendor',
    test: new RegExp(`${NM}(vue|@vue|vue-router|pinia|nostics)[\\\\/]`),
    priority: 45,
  },
  {
    name: 'axios',
    test: new RegExp(`${NM}axios[\\\\/]`),
    priority: 40,
  },
  {
    // zrender（echarts 渲染内核）+ tslib 单独成组：
    // 1) `includeDependenciesRecursively: false` ⇒ 分组只看模块 id，echarts 的传递依赖
    //    必须显式列出，否则会掉进 priority=1 的 vendor，撑大首屏 chunk 并破坏缓存边界；
    // 2) 与 echarts 合为一组会产出 >500KB 的单个 chunk，拆开后两块都在阈值内，
    //    且 zrender 版本远比 echarts 稳定，独立缓存命中更好。
    // 依赖方向单向：echarts → zrender → tslib，不会产生循环 chunk。
    // tslib 目前仅由 echarts/zrender 使用（已核对：@vue-flow 等首屏依赖不引用 tslib）。
    name: 'zrender',
    test: new RegExp(`${NM}(zrender|tslib)[\\\\/]`),
    priority: 43,
  },
  {
    // ECharts 仅被 `EChartsBlock.vue` 动态 import（助手 chart block），不进首屏。
    name: 'echarts',
    test: new RegExp(`${NM}echarts[\\\\/]`),
    priority: 42,
  },
  {
    name: 'element-plus-base',
    test: new RegExp(
      `${NM}element-plus[\\\\/]es[\\\\/](?:(${EP_BASE_ES_DIRS})[\\\\/]|components[\\\\/](${EP_BASE_COMPONENTS})[\\\\/])`,
    ),
    priority: 30,
  },
  {
    name: 'element-plus-mid',
    test: new RegExp(`${NM}element-plus[\\\\/]es[\\\\/]components[\\\\/](${EP_MID_COMPONENTS})[\\\\/]`),
    priority: 25,
  },
  {
    // 兜住 element-plus 剩余部分：聚合器 es/index.mjs + 顶层组件 + theme-chalk
    name: 'element-plus',
    test: new RegExp(`${NM}element-plus[\\\\/]`),
    priority: 20,
  },
  {
    // 其余第三方依赖（dayjs / lodash-es / @popperjs / async-validator 等）统一进 vendor。
    // 它们都是依赖图叶子，不会反向引用上层，天然无环。
    name: 'vendor',
    test: new RegExp(`${NM}`),
    priority: 1,
  },
]

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    host: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    // 页面级 chunk 已由 router 的动态 import 拆出；这里只负责把 vendor 从业务包中剥离
    rolldownOptions: {
      output: {
        codeSplitting: {
          // 关闭「递归吞并依赖」：否则优先级更高的组会把 vue / @vueuse 等底层依赖一并
          // 吸进自己的 chunk，导致分组名不副实、vendor 缓存边界混乱。关闭后「模块 → 分组」
          // 完全由 id 决定，与上面的拓扑分层分析一一对应，可推理、可验证。
          includeDependenciesRecursively: false,
          groups: VENDOR_CHUNK_GROUPS,
        },
      },
    },
  },
})
