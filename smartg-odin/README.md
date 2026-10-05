# SmartG-Odin API

ODIN 后端服务，基于 Go + Gin 框架

## 项目结构

```
smartg-odin-api/
├── cmd/
│   ├── odin/                 # 主服务入口
│   └── odin-cli/             # CLI工具
├── internal/
│   ├── ds/                   # 数据源相关
│   │   ├── protocol/         # 统一数据访问协议
│   │   ├── connector/        # 数据连接器
│   │   ├── pool/             # 连接池管理
│   │   ├── cache/            # 缓存管理
│   │   ├── query/            # 查询引擎
│   │   └── stream/           # 流处理
│   ├── ont/                  # 本体相关
│   │   ├── model/            # 本体模型
│   │   ├── mapping/          # 映射配置
│   │   ├── reason/           # 推理引擎
│   │   └── store/            # 本体存储
│   ├── server/               # 服务层
│   │   ├── api/              # REST API
│   │   ├── mcp/              # MCP Server
│   │   └── websocket/        # WebSocket
│   └── common/               # 公共模块
│       ├── config/           # 配置管理
│       ├── logger/           # 日志
│       ├── audit/            # 审计
│       └── utils/            # 工具
├── configs/                  # 配置文件
├── docs/                     # 文档（Swagger）
└── test/                     # 测试
```

## 技术栈

- **语言**：Go 1.21+
- **Web框架**：Gin
- **数据库**：SQLite（开发）/ PostgreSQL（生产）
- **本地引擎**：DuckDB
- **SQL解析**：pg_query_go / tidb parser
- **配置管理**：Viper
- **日志**：slog
- **测试**：testify + testcontainers-go

## 快速开始

### 环境要求

- Go 1.21+
- SQLite 3.x
- DuckDB（可选）

### 安装依赖

```bash
go mod init smartg-odin
go mod tidy
```

### 配置

复制配置模板：

```bash
cp configs/config.example.yaml configs/config.yaml
```

编辑配置文件，设置数据库连接等参数。

### 运行

```bash
# 开发模式
go run cmd/odin/main.go -config configs/config.yaml

# 构建
go build -o bin/odin cmd/odin/main.go

# 运行
./bin/odin -config configs/config.yaml
```

### 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/ds/connector/...

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## API 文档

启动服务后访问：

- Swagger UI: http://localhost:8080/swagger/index.html
- Health Check: http://localhost:8080/api/v1/health

## 开发指南

### 添加新的数据源连接器

1. 实现 `connector.Connector` 接口
2. 在 `connector/registry.go` 中注册
3. 添加配置支持
4. 编写测试

### 添加新的本体规则

1. 在 `ont/rule/` 下定义规则类型
2. 在 `ont/reason/` 下实现推理逻辑
3. 在 `ont/model/` 下扩展数据模型
4. 编写测试

## 部署

### Docker

```bash
docker build -t smartg-odin-api .
docker run -p 8080:8080 smartg-odin-api
```

### 系统服务

```bash
# systemd 服务文件
sudo cp deploy/odin.service /etc/systemd/system/
sudo systemctl enable odin
sudo systemctl start odin
```

## 监控

- Prometheus metrics: http://localhost:8080/metrics
- Health check: http://localhost:8080/api/v1/health
- Ready check: http://localhost:8080/api/v1/ready