# ODIN - Ontology-Driven Intelligent Navigator

本体驱动的智能数据导航系统

---

## 项目特点

### 核心理念

**"面向业务概念，而非物理表"**

传统数据访问方式需要了解底层数据库结构、表关系、字段含义。ODIN 通过本体建模，让用户直接用业务语言（客户、订单、产品）查询数据，无需关心数据存储在哪里、用什么格式。

### 核心价值

| 传统方式 | ODIN 方式 |
|---------|----------|
| 直接操作物理表 | 面向业务概念（本体） |
| 跨库查询困难 | 统一语义层，自动联邦 |
| 新增数据源需改代码 | 插件化连接器，配置即接入 |
| 专家经验难以复用 | 规则推理，知识沉淀 |
| 需要数据同步（ETL） | 虚拟集成，实时访问源头 |
| 只有技术人员能用 | 业务人员也能自助查询 |

### 关键特性

#### 1. 本体驱动（Ontology-Driven）
- **语义建模**：用类、属性、关系描述业务概念
- **知识沉淀**：业务规则、领域知识固化为本体
- **智能推理**：自动发现数据关系、推导新知识

#### 2. 虚拟集成（Virtual Integration）
- **不搬数据**：直接访问源头，避免数据冗余
- **实时洞察**：查询最新数据，无需等待同步
- **降低成本**：无需维护ETL管道和数据仓库

#### 3. 多源融合（Multi-Source Federation）
- **统一视图**：跨数据源的业务实体对齐
- **自动联邦**：查询时自动合并多源数据
- **智能优化**：谓词下推、本地物化、性能优化

#### 4. 国产化支持（Localization）
- **国产数据库**：达梦、金仓、openGauss、GaussDB等
- **信创环境**：支持国密、等保、内网部署
- **自主可控**：Go语言实现，单二进制部署

#### 5. AI 友好（AI-Friendly）
- **MCP 协议**：为 AI Agent 提供标准化接口
- **语义理解**：AI 可理解本体结构，智能生成查询
- **渐进披露**：控制 token 数量，优化上下文

### 与传统方案对比

| 维度 | 传统数仓 | 数据虚拟化 | **ODIN** |
|------|---------|-----------|----------|
| 数据同步 | 需要 ETL | 不需要 | **不需要** |
| 实时性 | T+1 | 实时 | **实时** |
| 语义层 | 弱（星型模型） | 无 | **强（本体）** |
| 跨源查询 | 困难 | 支持 | **智能联邦** |
| AI 集成 | 无 | 无 | **MCP 协议** |
| 学习曲线 | 高 | 中 | **低（业务语言）** |

### 应用场景

#### 场景1：跨系统数据洞察
> "我想看看 ERP 和 CRM 中客户数据的差异"
> 
> ODIN：定义客户本体 → 映射两个系统 → 自动联邦查询 → 对比分析

#### 场景2：业务人员自助分析
> "查一下今年 VIP 客户的订单情况"
> 
> ODIN：选择客户类 → 筛选 VIP 规则 → 关联订单 → 生成报表

#### 场景3：AI 辅助决策
> "帮我分析一下客户流失的原因"
> 
> ODIN：AI 理解本体 → 自动构建查询 → 多维分析 → 生成洞察

#### 场景4：信创环境数据整合
> "把达梦和金仓的数据统一起来"
> 
> ODIN：配置国产数据库连接 → 本体建模 → 自动映射 → 统一查询

---

## 项目结构

```
odin/
├── smartg-odin-api/          # 后端服务（Go + Gin）
├── smartg-odin-ui/           # 前端应用（Vue 3 + Element Plus）
├── odin-console/             # 原型设计（参考）
├── 设计文档/                  # 设计文档集合
└── README.md                 # 本文件
```

## 模块说明

### 1. smartg-odin-api（后端服务）

基于 Go + Gin 框架的后端服务，提供：

- **数据源管理**：统一数据访问协议（UDAP）
- **本体管理**：类、属性、关系、规则的CRUD
- **映射管理**：本体到物理数据源的映射配置
- **查询引擎**：本体查询 → 物理查询翻译
- **推理引擎**：规则展开、虚拟类
- **审计日志**：操作记录与变更历史

**技术栈**：
- Go 1.21+
- Gin Web框架
- DuckDB（本地分析引擎）
- SQLite/PostgreSQL（元数据存储）

### 2. smartg-odin-ui（前端应用）

基于 Vue 3 + Element Plus 的前端应用，提供：

- **本体图谱**：可视化编辑本体类、属性、关系
- **数据接入**：数据源配置、Schema浏览、映射管理
- **查询工作台**：本体查询构建、SQL预览、结果展示
- **推理验证**：规则试推理、路径高亮

**技术栈**：
- Vue 3 + TypeScript
- Element Plus UI组件库
- Vue Flow（图谱可视化）
- Pinia（状态管理）

### 3. odin-console（原型设计）

早期原型设计，用于：

- 验证产品概念
- UI/UX设计参考
- 功能需求梳理

### 4. 设计文档

包含所有设计文档：

- **技术方案**：Go MCP国产库数据融合网关、本体可视化管理、本体驱动数据访问
- **数据接入方案**：数据源管理、Schema管理、映射管理设计
- **核心模块设计**：SmartG-Odin数据集市核心模块完整设计
- **数据库脚本**：PostgreSQL和SQLite版本的表结构脚本

## 快速开始

### 后端开发

```bash
cd smartg-odin-api
go mod init smartg-odin
go mod tidy
go run cmd/odin/main.go
```

### 前端开发

```bash
cd smartg-odin-ui
npm install
npm run dev
```

### 数据库初始化

```bash
# SQLite（开发环境）
sqlite3 odin.db < 设计文档/odin_database_sqlite.sql

# PostgreSQL（生产环境）
psql -U postgres -d odin -f 设计文档/odin_database_schema.sql
```

## 核心功能

### 1. 数据源管理

支持多种数据源类型：
- 关系型数据库：PostgreSQL、MySQL、达梦、金仓等
- 文件系统：CSV、Excel、Parquet（通过DuckDB挂载）
- API接口：REST、GraphQL

### 2. 本体建模

- **类管理**：定义业务实体（客户、订单、产品等）
- **属性管理**：定义实体属性（名称、类型、约束等）
- **关系管理**：定义实体间关系（1:1、1:N、N:M）
- **规则管理**：定义业务规则（VIP客户、活跃用户等）

### 3. 映射配置

- 本体属性 → 物理列的映射
- 值映射（枚举值转换）
- 转换表达式（数据清洗）
- 置信度评估

### 4. 智能查询

- 本体查询 → SQL翻译
- 跨源联邦查询
- 查询优化建议
- 执行计划分析

### 5. 推理引擎

- 规则展开（虚拟类）
- 关系路径发现
- 语义对齐
- 置信度计算

## 开发规范

### 代码规范

- 后端：Go官方规范 + Uber Go Style Guide
- 前端：Vue 3 Composition API + TypeScript

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