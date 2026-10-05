/**
 * ECharts 按需引入集中处。
 *
 * 设计要点：
 * 1. **只注册用到的模块**：聚合结果的图表形态固定为 bar / line / pie 三种，
 *    配合 grid（直角坐标系）、tooltip（悬浮明细）、legend（多度量图例）即可，
 *    渲染器只用 Canvas。全量 `import * as echarts from 'echarts'` 会把 ~1MB
 *    的图表/组件全家桶拖进产物，与 code-splitting 的目标相悖。
 * 2. **本模块只被 `EChartsBlock.vue` 动态 import**，因此 echarts + zrender
 *    会被拆成独立的异步 chunk（见 vite.config.ts 的 `echarts` 分组），
 *    不进首屏、不影响首屏 vendor 的缓存边界。
 * 3. `use()` 是幂等的全局注册，多次 import 本模块不会重复注册。
 */
import * as echarts from 'echarts/core'
import { BarChart, LineChart, PieChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import type { BarSeriesOption, LineSeriesOption, PieSeriesOption } from 'echarts/charts'
import type {
  GridComponentOption,
  LegendComponentOption,
  TooltipComponentOption,
} from 'echarts/components'
import type { ComposeOption } from 'echarts/core'

echarts.use([
  BarChart,
  LineChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  CanvasRenderer,
])

/**
 * 由已注册模块组合出的 option 类型。
 * 用它而不是 `EChartsOption`，可让 TS 在写错未注册模块（如 `radar`）时直接报错。
 */
export type ChartOption = ComposeOption<
  | BarSeriesOption
  | LineSeriesOption
  | PieSeriesOption
  | GridComponentOption
  | TooltipComponentOption
  | LegendComponentOption
>

/** `echarts.init()` 返回的实例类型（`EChartsType` 在 core 中重命名为 `ECharts`） */
export type EChartsInstance = echarts.ECharts

export default echarts
