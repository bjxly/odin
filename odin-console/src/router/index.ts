import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: () => import('@/pages/Home.vue'), meta: { title: '首页', icon: 'home' } },
    { path: '/sources', name: 'sources', component: () => import('@/pages/SourceManager.vue'), meta: { title: '数据源管理', icon: 'database' } },
    { path: '/ontology', name: 'ontology', component: () => import('@/pages/OntologyGraph.vue'), meta: { title: '本体图谱', icon: 'share' } },
    { path: '/mapping', name: 'mapping', component: () => import('@/pages/MappingWorkbench.vue'), meta: { title: '映射工作台', icon: 'link' } },
    { path: '/query', name: 'query', component: () => import('@/pages/QueryLab.vue'), meta: { title: '查询实验室', icon: 'search' } },
    { path: '/data-access', name: 'dataAccess', component: () => import('@/pages/DataAccess.vue'), meta: { title: '数据接入', icon: 'plug' } },
    { path: '/connectors', name: 'connectors', component: () => import('@/pages/ConnectorManager.vue'), meta: { title: '连接器管理', icon: 'box' } },
    { path: '/reason', name: 'reason', component: () => import('@/pages/ReasonLab.vue'), meta: { title: '推理实验室', icon: 'cpu' } },
    { path: '/assistant', name: 'assistant', component: () => import('@/pages/Assistant.vue'), meta: { title: 'AI 助手', icon: 'message' } },
  ],
})

export default router
