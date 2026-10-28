import type { MenuTreeVO } from '@/types/sys'

/**
 * 免登录模式下的静态菜单定义（对应 VITE_AUTH_ENABLED=false 的内部版）。
 *
 * <p>结构与后端 <code>getMyMenus()</code> 返回完全一致，路由守卫使用现有的
 * <code>generateRoutes()</code> 打平注册到 MainLayout 下。
 *
 * <p>约定：
 * <ul>
 *   <li><code>component</code> 为相对 <code>src/views/</code> 的路径（不带前后斜杠、不带 .vue 后缀），
 *       既支持 <code>"sys/user/index"</code> 也支持 <code>"dashboard/index"</code>；</li>
 *   <li>目录节点不写 <code>component</code>，只写 <code>children</code>；</li>
 *   <li>顶栏"首页"（<code>/dashboard</code>）由 permission store 自动前置，本文件无需重复声明；</li>
 *   <li>系统管理页（<code>views/sys/*</code>）在内部版中默认不列出，从菜单入口切断访问路径。</li>
 * </ul>
 *
 * <p>业务改造点：按内部系统的实际页面清单增删条目即可，无需触碰任何后端配置。
 */
export const STATIC_MENUS: MenuTreeVO[] = [
  // ==================== 示例：可按需替换为实际内部系统菜单 ====================
  // {
  //   path: '/demo',
  //   name: 'DemoGroup',
  //   meta: { title: '示例目录', icon: 'Folder' },
  //   children: [
  //     {
  //       path: '/demo/list',
  //       name: 'DemoList',
  //       component: 'demo/list/index',
  //       meta: { title: '示例列表', icon: 'List', keepAlive: true },
  //     },
  //     {
  //       path: '/demo/detail',
  //       name: 'DemoDetail',
  //       component: 'demo/detail/index',
  //       meta: { title: '示例详情', hidden: true },
  //     },
  //   ],
  // },
]
