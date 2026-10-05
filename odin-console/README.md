# OntoMCP Console（V5 多本体管理）

本体可视化管理台：**Vue 3 + TypeScript + Vue Flow + Element Plus + Pinia**。

当前 **V5**：**多本体图谱** + 映射工作台 + 源管理 + 查询工作台。

## 快速开始

```bash
cd ontomcp-console
npm install
npm run dev
```

## 功能

### 多本体图谱（核心）
- **Header 本体切换器**：下拉选择不同本体，实时切换图谱 / 映射 / 查询上下文
- **新建本体**：名称 + 描述 + 可从现有本体克隆
- **删除本体**：至少保留一个
- **Mock 示例**：
  - **进销存域**：Customer / Order / Product / VipCustomer / BigOrder 等 7 类，5 关系，2 规则
  - **海军基地域**：NavalBase / BaseFacility / BaseEquipment / DeploymentSeq / LogisticsSupport / PatrolMission 等 10 类，6 关系，4 规则

### 图谱编辑
- 类/属性/关系 CRUD、拖拽布局、画布左上工具栏

### 映射工作台
- 按类对照 / 按源浏览建议 / 校验 / coverage
- 映射绑定当前本体，切换本体时自动切换映射上下文

### 源管理
- 全局共享数据源（跨本体），新增/编辑/删除/测试/内省
- 删除本体不影响数据源

### 查询工作台
- 自然语言 → 本体查询（支持海军域关键词）
- 推理展开、Dry-run SQL、mock 结果集
- 预设模板含海军域示例

## 架构

```
workspace store（全局）
├── sources: SourceInfo[]           ← 共享基础设施
├── schemas: SourceSchema[]         ← 物理 schema
└── ontologies: OntologyRecord[]    ← 多本体
        ├── 进销存域（classes + relations + rules + classMaps）
        └── 海军基地域（classes + relations + rules + classMaps）

ontology store（派生式）
└── 从 workspace.activeOntology 派生所有 computed
    页面代码零改动：onto.classes / onto.relations 行为不变

mapping store（派生式）
└── 从 workspace.activeOntology.classMaps 派生
```

## 目录

```
src/
  mock/
    ontology.ts        # 进销存本体 + 全局 sources
    naval.ts           # 海军基地本体 mock
    schema.ts          # 物理 schema（含海军库）
    query.ts           # 样例行 + NL 解析 + 预置模板
  types/
    ontology.ts        # OntologyRecord / WorkspaceData
    mapping.ts
    query.ts
  stores/
    workspace.ts       ← 新增：多本体注册表
    ontology.ts        ← 改造：派生式
    mapping.ts         ← 改造：绑定当前本体
    query.ts           ← 小改
  pages/
    OntologyGraph.vue
    MappingWorkbench.vue
    SourceManager.vue
    QueryLab.vue
  App.vue              ← 本体切换器 + 新建对话框
```

## 试玩建议

1. Header 右侧下拉切换到「海军基地域」→ 图谱变成海军基地图
2. 点「映射」→ 看到海军基地的 ClassMap（已预置 5 个映射）
3. 点「查询」→ 模板「现役海军基地清单」→ Dry-run 看 SQL
4. 自然语言输入「查询青岛基地的装备和后勤保障」→ 解析为本体查询
5. 切回「进销存域」→ 图谱恢复客户/订单图
6. 点「新建」→ 从「进销存域」克隆一个新本体
