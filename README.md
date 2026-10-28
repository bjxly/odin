# ODIN - Ontology-Driven Intelligent Navigator

本体驱动的智能数据洞察（SmartJ-Odin）

> 
> ODIN — Ontology-Driven Intelligent Navigator
> ODIN 北欧神话中全知之神，天然契合"基于本体洞察万物"的产品气质。面向业务概念，而非物理表 —— 用本体建模屏蔽底层数据源差异，让业务与技术都能用"客户、订单、产品"这样的业务语言直接查询、分析和洞察数据。

---

## 项目特点

### 核心理念

**"面向业务概念，而非物理表"**

传统数据访问方式需要了解底层数据库结构、表关系、字段含义。ODIN 通过本体建模，让用户直接用业务语言查询数据，无需关心数据存储在哪里、用什么格式。

### 核心价值

| 传统方式 | ODIN 方式 |
|---------|----------|
| 直接操作物理表 | 面向业务概念（本体） |
| 跨库查询困难 | 统一语义层，多源虚拟集成 |
| 新增数据源需改代码 | 插件化连接器，配置即接入 |
| 专家经验难以复用 | 规则声明，知识沉淀 |
| 需要数据同步（ETL） | 虚拟集成，实时访问源头 |
| 只有技术人员能用 | 业务人员自助查询 + AI 问数 |

### 关键特性

#### 1. 本体驱动（Ontology-Driven）
- **语义建模**：用类、属性、关系描述业务概念，图谱可视化编辑
- **知识沉淀**：业务规则、领域知识固化为本体（如 `VipCustomer(level='VIP')`）
- **坐标布局**：类节点位置三级优先级持久化（用户保存 > 种子设计 > 自动布局），支持拖动保存

#### 2. 虚拟集成（Virtual Integration）
- **不搬数据**：直接访问源头，避免数据冗余
- **实时内省**：应用启动时对激活数据源做运行时 Schema 内省，真实表结构动态获取
- **映射对齐**：本体属性 → 物理列的声明式映射，含值映射与转换表达式

#### 3. 多源与国产支持（Multi-Source & Localization）
- **统一访问协议（UDAP）**：插件化连接器抽象，屏蔽不同数据库方言
- **国产数据库**：达梦、金仓、openGauss、GaussDB 等（连接器可扩展）
- **信创环境**：支持国密、内网部署，元数据存于 MySQL

#### 4. AI 副驾（AI Copilot）
- **确定性优先**：LLM 只产出受约束的本体逻辑 SQL，经封闭词表校验后编译为物理 SQL，杜绝幻觉
- **工具编排**：基于 Agentscope-Java 的 ReAct Agent + Toolkit（问数、本体写操作草稿等）
- **可溯源**：每条回答关联物理 SQL、数据集快照与推理轨迹；写操作仅生成待人工审阅的草稿
- **智能问数出图**：AI 分析结论表可确定性解析成图，支持柱/折/饼、行列轴向切换、中文表头

### 与传统方案对比

| 维度 | 传统数仓 | 数据虚拟化 | **ODIN** |
|------|---------|-----------|----------|
| 数据同步 | 需要 ETL | 不需要 | **不需要** |
| 实时性 | T+1 | 实时 | **实时** |
| 语义层 | 弱（星型模型） | 无 | **强（本体）** |
| AI 集成 | 无 | 无 | **AI 副驾 + 确定性翻译** |
| 学习曲线 | 高 | 中 | **低（业务语言）** |

### 应用场景

#### 场景1：业务人员自助分析
> "查一下今年 VIP 客户的订单情况"
>
> ODIN：选择客户类 → 筛选 VIP 规则 → 关联订单 → 生成图表

#### 场景2：AI 辅助问数
> "总共有多少 VIP 客户？"
>
> ODIN：AI 理解本体词表 → 编写逻辑 SQL → 确定性编译执行 → 结论 + 可视化

#### 场景3：跨系统数据整合
> "把 MySQL 和 PostgreSQL 的数据统一起来看"
>
> ODIN：配置多数据源连接 → 本体建模 → 映射对齐 → 统一语义查询

---

## 项目结构

```
odin/
├── smartj-odin/              # 后端服务（Java 21 + Spring Boot，Maven 多模块）
│   ├── odin-common/          #   公共能力（UDAP 统一数据访问协议、Connector）
│   ├── odin-api/             #   各业务模块契约层（DTO / VO / Query）
│   ├── odin-biz/             #   各业务模块实现层（Controller / Service / Mapper）
│   └── odin-bootstrap/       #   Spring Boot 启动模块（端口 9000，context-path /api）
├── smartj-odin-ui/           # 前端应用（Vue 3 + TypeScript + Element Plus）
├── odin-console/             # 早期原型（Vue 3，概念验证与 UI 参考）
├── docker/                   # 多源数据库编排（MySQL / PostgreSQL / Redis + 初始化脚本）
├── config/                   # JWT 密钥等运行时配置
├── docs/superpowers/         # 迭代计划与设计 spec
├── 设计文档/                  # 架构方案、数据接入方案、数据库 Schema 脚本
└── README.md                 # 本文件
```

## 模块说明

### 1. smartj-odin（后端服务）

基于 Java 21 + Spring Boot 的 Maven 多模块工程，按"契约（api）/ 实现（biz）"分层，`odin-bootstrap` 聚合启动。

| 业务模块 | 职责 |
|---------|------|
| **datasource** | 数据源与连接器管理（UDAP 统一访问、运行时 Schema 内省、连通性测试） |
| **ontology** | 本体管理（类 / 属性 / 关系 / 规则 CRUD、图谱与布局坐标持久化、词表输出） |
| **mapping** | 映射管理（本体属性 ↔ 物理列映射、值映射、置信度） |
| **query** | 智能查询（本体逻辑 → 物理 SQL 编译执行、结果集） |
| **reason** | 推理验证（规则展开、虚拟类过滤、路径校验） |
| **assistant** | AI 副驾（ReAct Agent、问数工具、SSE 流式、草稿审阅） |
| **insight** | 洞察看板（图表保存、变更历史） |

**技术栈**：
- Java 21、Spring Boot 4（parent = `smartj-parent`，依赖经 `smartj-dependencies` BOM 管理）
- MyBatis-Flex（APT 生成 TableDef）+ HikariCP
- Jackson 3、Spring Security + OAuth2 Resource Server、Jasypt 加密
- Agentscope-Java（AI 编排：ReActAgent + Toolkit）
- 元数据 / 应用库：MySQL 8.4；缓存：Redis

### 2. smartj-odin-ui（前端应用）

基于 Vue 3 + Element Plus 的单页应用，`src/views/odin/` 下按业务域组织：

| 页面 | 说明 |
|------|------|
| **本体治理（ontology）** | 图谱可视化编辑类、属性、关系、规则；条件组结构化构造器 |
| **数据接入（data-access）** | 数据源、连接器、映射管理、推理验证 |
| **智能问数（assistant）** | AI 副驾对话、SSE 流式、结论出图、草稿审阅 |
| **洞察看板（insight）** | 图表保存与管理 |
| **首页（dashboard）** | 本体卡片、场景推荐、快捷入口 |

**技术栈**：
- Vue 3.5 + TypeScript + Vite
- Element Plus + Tailwind CSS
- Vue Flow（图谱可视化）、ECharts（图表）
- Pinia（状态管理）、vue-i18n（国际化）

### 3. odin-console（原型设计）

早期原型，用于验证产品概念、UI/UX 设计与功能需求梳理。

### 4. 设计文档

- **架构方案**：本体驱动数据洞察完整设计、Java 重构技术方案（Spring Boot 4 + Agentscope-Java）
- **技术方案**：OntoMCP 本体可视化 / 本体驱动数据访问、国产库数据融合网关
- **数据接入方案**：数据源、Schema、映射管理设计
- **数据库脚本**：`odin_database_schema.sql`（PostgreSQL）、`odin_database_sqlite.sql`（SQLite）

## 快速开始

### 环境要求

- JDK 21
- Maven 3.6+
- Node.js + npm
- Docker（用于启动 MySQL / PostgreSQL / Redis）

### 1. 启动数据库（Docker）

```bash
cd docker
docker compose up -d
```

默认端口映射：

| 服务 | 容器 | 宿主机端口 |
|------|------|-----------|
| MySQL 8.4 | odin-mysql | 13306 |
| PostgreSQL 17 | odin-pgsql | 15432 |
| Redis | odin-redis | 16379 |

> 初始化脚本位于 `docker/init-db/`（mysql / postgres），仅在数据目录为空时执行；改动已运行容器的数据需手动应用。

### 2. 启动后端

```bash
cd smartj-odin
mvn -s <你的 settings.xml> clean install
# 运行 odin-bootstrap（Spring Boot，端口 9000，context-path /api）
```

> 应用启动时会创建数据源元数据、本体映射，并对激活数据源做运行时 Schema 内省。

### 3. 启动前端

```bash
cd smartj-odin-ui
npm install
npm run dev      # Vite 开发服务，/api 代理至 localhost:9000
```

生产构建：`npm run build`（先 `vue-tsc` 类型检查再打包）。

## 核心功能

### 1. 数据源管理
- 关系型数据库：MySQL、PostgreSQL、达梦、金仓等（连接器可扩展）
- 连通性测试、运行时 Schema 内省、目录浏览

### 2. 本体建模
- **类 / 属性 / 关系**：定义业务实体及其语义
- **规则管理**：声明式业务规则（阈值 / 值匹配 / 条件组 AND/OR/NOT）
- **图谱设计**：Vue Flow 可视化编辑，布局坐标持久化与自动布局

### 3. 映射配置
- 本体属性 → 物理列映射、值映射、转换表达式、置信度评估

### 4. 智能查询
- 本体逻辑 SQL → 物理 SQL 确定性编译执行
- 结果集供图表卡片消费，支持聚合与可视化

### 5. 推理验证
- 规则展开、虚拟类过滤、关系路径校验

### 6. AI 副驾
- 自然语言问数，ReAct Agent 编排工具，SSE 流式返回
- 写操作生成待审阅草稿（本体类/属性/映射/数据源），人工批准后发布
- 分析结论表可确定性解析成图，图文口径一致

## 开发规范

### 代码规范

- 后端：Java 21 + Spring Boot 分层约定（Controller / Service / Mapper，api 契约 + biz 实现）
- 前端：Vue 3 Composition API（`<script setup>`）+ TypeScript + Element Plus 规范

### 提交规范

```
feat: 新功能
fix: 修复bug
docs: 文档更新
style: 代码格式调整
refactor: 重构
test: 测试相关
chore: 构建/工具链更新
```

### 分支策略

- `main`：生产分支
- `develop`：开发分支
- `feature/*`：功能分支
- `hotfix/*`：紧急修复分支

## 联系方式

- 项目负责人：[待填写]
- 技术文档：[设计文档/](./设计文档/)
- 问题反馈：[待创建Issue模板]

## 许可证

本项目采用 [Apache License 2.0](./LICENSE) 许可证。

```
Copyright 2024 ODIN Project

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
