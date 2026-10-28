# 多本体图谱设计文档

> 目标：一套控制台管理多个独立本体图谱，共享同一套数据源基础设施。

---

## 1. 架构总览

```
Workspace（工作区）
├── sources: SourceInfo[]              ← 全局共享，不绑定本体
├── schemas: SourceSchema[]            ← 物理 schema，全局共享
├── activeOntologyId: string           ← 当前激活的本体
└── ontologies: OntologyRecord[]       ← 本体注册表（轻量索引）
        │
        ├── OntologyRecord「进销存域」
        │     classes / relations / rules / layout
        │     classMaps[]              ← 该本体的映射
        │     changeLog[]
        │
        └── OntologyRecord「海军基地域」
              classes / relations / rules / layout
              classMaps[]
              changeLog[]
```

## 2. 核心原则

| 层 | 归属 | 说明 |
|---|---|---|
| 数据源 (Source) | 全局共享 | erp_pg / crm_dm / oa_kb 一套，所有本体共用 |
| 物理 Schema | 全局共享 | SourceSchema 按源组织，内省一次全局可用 |
| 本体 (Ontology) | 独立单元 | 每个本体有自己的类/关系/规则/布局/版本 |
| 映射 (ClassMap) | 绑定本体 | classId 在本体内唯一，映射随本体切换 |
| 查询 | 绑定本体 | 查询工作台针对当前激活本体 |

## 3. 数据模型

### 3.1 OntologyRecord（新增）

```ts
export interface OntologyRecord {
  id: string
  name: string
  description?: string
  version: number
  createdAt: number
  updatedAt: number
  // 本体内容
  classes: OntoClass[]
  relations: OntoRelation[]
  rules: OntoRule[]
  layout: Record<string, LayoutPos>
  // 映射（从 mapping store 绑定过来）
  classMaps: ClassMap[]
}
```

### 3.2 WorkspaceData（替代 OntologySnapshot）

```ts
export interface WorkspaceData {
  activeOntologyId: string
  sources: SourceInfo[]
  ontologies: OntologyRecord[]
}
```

### 3.3 OntologySnapshot → 废弃

原 `OntologySnapshot` 不再使用，拆为 `WorkspaceData` + `OntologyRecord`。

## 4. 存储布局

```
localStorage
├── ontomcp-workspace-v4         ← { activeOntologyId, sources[], ontologies[] }
├── ontomcp-schemas-v3           ← 物理 schema（已有，不变）
├── ontomcp-saved-queries-v1     ← 查询模板
└── ontomcp-query-history-v1     ← 查询历史
```

所有本体数据（含 classMaps）合并在 `workspace` 里，不再按本体拆 key。
因为本体数量有限（2-10），总量可控。

## 5. Store 架构

```
stores/
├── workspace.ts  ← 新增：本体注册表 + sources + activeOntology 切换
├── ontology.ts   ← 改造：从 workspace 派生，操作当前本体
├── mapping.ts    ← 改造：classMaps 绑定当前本体
├── query.ts      ← 小改：跟随 activeOntology
```

### 关键设计

`ontology.ts` 不再持有 snapshot ref，而是从 `workspace.activeOntology` 派生所有 computed：

```ts
const workspace = useWorkspaceStore()
const active = computed(() => workspace.activeOntology)
const classes = computed(() => active.value?.classes ?? [])
// save() 写回 workspace
```

现有页面用 `onto.classes` 等，行为不变。

## 6. Mock 本体

### 6.1 进销存域（现有数据迁移）

类：Customer, VipCustomer, Order, BigOrder, OrderLine, Product, Shipment
域：客户域, 销售域, 商品域, 履约域

### 6.2 海军基地域（新增）

类：
- **NavalBase**：海军基地（主类）
- **BaseFacility**：基地固定设施（码头/仓库/指挥所）
- **BaseEquipment**：基地装备（雷达/火炮/舰艇）
- **DeploymentSeq**：基地部署序列（部署计划/轮换周期）
- **LogisticsSupport**：后勤保障（补给/维修/医疗）
- **PatrolMission**：巡逻任务（派生类：近海巡逻/远海巡逻）

关系：
- NavalBase → BaseFacility (1:N)
- NavalBase → BaseEquipment (1:N)
- NavalBase → DeploymentSeq (1:N)
- NavalBase → LogisticsSupport (1:N)
- BaseEquipment → PatrolMission (N:N)

## 7. UI 变更

### 7.1 Header

右侧加本体切换下拉：
```
[▼ 进销存本体 v12 ▾]  保存 │ 重置
```

### 7.2 侧栏（图谱页）

底部加「本体管理」面板：
- 列表：名称 + 类数 + 版本 + 激活标记
- 操作：切换 / 新建 / 删除 / 克隆

### 7.3 新建本体对话框

- 名称（必填）
- 描述（选填）
- 克隆源（选填：从当前本体克隆）

## 8. 迁移策略

首次加载 workspace 数据时：
1. 检查 localStorage 是否有 `ontomcp-workspace-v4`
2. 如果没有，检查旧 `ontomcp-ontology-v2` + `ontomcp-mappings-v3`
3. 旧数据自动转为 `id: 'inventory'` 的本体（进销存域）
4. 同时创建 `id: 'naval'` 的海军基地本体（from mock）
5. sources 从旧 ontology snapshot 提取，写入 workspace
6. 清理旧 localStorage key

## 9. 实施步骤

| 步 | 内容 | 风险 |
|---|---|---|
| S1 | 类型层改造 | 低 |
| S2 | Mock 军事本体 | 低 |
| S3 | 新建 workspace store | 中 |
| S4 | 改造 ontology store（核心） | **高** |
| S5 | 改造 mapping store | 中 |
| S6 | UI 改造 | 低 |
| S7 | 迁移逻辑 | 中 |
| S8 | 构建验证 + README | 低 |

S4 是最大风险点：ontology store 所有函数要从操作 snapshot 切换为操作 workspace.activeOntology。
