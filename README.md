# Wallet Service

Go 语言实现的数字钱包服务，提供 gRPC 和 REST 双协议 API，支持钱包管理、转账、存取款及会话认证等功能。

## 系统架构

### 技术栈

| 组件 | 技术选型 |
|------|---------|
| 语言 | Go 1.25 |
| 依赖注入 | uber/fx |
| RPC 框架 | gRPC + gRPC-Gateway |
| HTTP 框架 | Gin |
| 日志 | zap |
| 密码加密 | bcrypt |
| ID 生成 | Snowflake 雪花算法 |
| Proto 管理 | buf |
| 存储 | 内存（sync.RWMutex 并发安全） |

### 分层架构

项目采用 DDD（领域驱动设计）分层架构，各层职责清晰：

```
Handler (gRPC/REST)  ──>  Service  ──>  Domain  ──>  Repository
      ↑                                                   ↑
  中间件（认证/日志/恢复）                          内存存储实现
```

- **Handler 层**：gRPC 服务实现，通过 gRPC-Gateway 同时暴露 REST API
- **Service 层**：应用服务，编排业务流程（WalletService、AuthService）
- **Domain 层**：核心业务逻辑与实体定义（Wallet、Transaction、Session）
- **Repository 层**：数据访问抽象接口及内存存储实现

### 项目目录

```
wallet/
├── api/
│   ├── grpc/v1/                # gRPC 生成代码（pb.go / grpc.pb.go / gw.pb.go）
│   └── proto/v1/               # Protobuf 定义文件
│       ├── wallet.proto        # 服务与消息定义
│       ├── buf.yaml            # buf 模块配置
│       └── buf.gen.yaml        # buf 代码生成配置
├── cmd/
│   └── server/main.go          # 应用入口，fx 模块组装
├── internal/
│   ├── config/                 # 配置管理（YAML + 环境变量）
│   ├── domain/
│   │   ├── entity/             # 领域实体
│   │   │   ├── wallet.go       # 钱包实体（ID、名称、余额、状态、版本号）
│   │   │   ├── transaction.go  # 交易记录实体（转账/存款/取款）
│   │   │   └── session.go      # 会话实体（令牌、过期时间）
│   │   ├── session/            # 会话领域逻辑（创建/验证/清理）
│   │   └── wallet/             # 钱包领域逻辑（存取款/转账/余额管理）
│   ├── handler/grpc/           # gRPC Handler 实现
│   ├── middleware/             # 中间件
│   │   ├── grpc.go             # gRPC 认证拦截器（Bearer Token）
│   │   └── gin.go              # Gin 日志中间件
│   ├── repository/
│   │   ├── interface.go        # Repository 接口定义
│   │   └── memory/             # 内存存储实现（线程安全）
│   ├── server/                 # 服务器启动与生命周期管理
│   └── service/                # 应用服务层
│       ├── auth.go             # 认证服务（登录/登出/会话验证）
│       └── wallet.go           # 钱包服务（CRUD/转账/交易记录）
├── pkg/
│   ├── errors/                 # 统一错误码定义（1xxx~5xxx）
│   ├── idgen/                  # ID 生成器接口与工厂
│   │   └── snowflake/          # Snowflake 实现（42位时间戳+6位机器码+12位序列号）
│   ├── logger/                 # 日志抽象与 zap 实现
│   └── pwd/                    # 密码哈希工具（bcrypt）
├── design.md                   # 原始需求文档
├── go.mod
└── go.sum
```

## 核心功能

### 钱包管理

- **创建钱包**：指定名称和密码，生成唯一钱包 ID（Snowflake，前缀 `W`），初始余额为 0
- **查询钱包**：通过钱包 ID 获取钱包信息（ID、名称、余额）
- **更新钱包**：修改钱包名称或密码

### 资金操作

- **转账**：从源钱包向目标钱包转移指定金额（需认证），失败时自动回滚
- **存款/取款**：直接向钱包存入或提取资金

### 交易记录

- 每笔操作生成交易记录（ID 前缀 `T`），包含类型、状态、时间等信息
- 支持按钱包 ID 分页查询交易历史

### 认证系统

- **登录**：通过钱包 ID + 密码登录，获取会话令牌（Token）
- **登出**：使会话令牌失效
- **会话管理**：令牌有 TTL（默认 24 小时），后台定期清理过期会话

### 设计要点

| 特性 | 说明 |
|------|------|
| 幂等性 | 转账/存取款支持幂等键（idempotency_key），防止重复操作 |
| 乐观锁 | 钱包实体含版本号字段，支持并发控制 |
| 金额精度 | 余额和金额使用 `int64`（单位：分），避免浮点精度问题 |
| 并发安全 | 内存存储使用 `sync.RWMutex` 读写锁 |
| 优雅关闭 | HTTP/gRPC 服务器支持优雅关闭（5 秒超时） |
| Panic 恢复 | gRPC 和 Gin 均配置了 panic recovery 中间件 |

## API 参考

### REST API（通过 gRPC-Gateway 代理）

所有 REST 端点前缀为 `/v1/`（gRPC-Gateway 路由），使用 JSON 格式。

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/wallets` | 创建钱包 | 否 |
| GET | `/wallets/{wallet_id}` | 查询钱包 | 否 |
| PUT | `/wallets/{wallet_id}` | 更新钱包 | 否 |
| POST | `/wallets/transfer` | 转账 | 是 |
| GET | `/wallets/transactions/{wallet_id}` | 查询交易记录 | 否 |
| POST | `/auth/login` | 登录 | 否 |
| POST | `/auth/logout` | 登出 | 否 |
| GET | `/healthy` | 健康检查 | 否 |

#### 创建钱包

```bash
curl -X POST http://localhost:8080/v1/wallets \
  -H "Content-Type: application/json" \
  -d '{"name": "my-wallet", "passwd": "123456"}'
```

响应：
```json
{"walletId": "W1234567890"}
```

#### 查询钱包

```bash
curl http://localhost:8080/v1/wallets/{wallet_id}
```

响应：
```json
{"id": "W1234567890", "name": "my-wallet", "balance": "0"}
```

#### 登录

```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"walletId": "W1234567890", "passwd": "123456"}'
```

响应：
```json
{"token": "abcdef1234567890...", "expiredAt": "2025-01-02T00:00:00Z"}
```

#### 转账（需认证）

```bash
curl -X POST http://localhost:8080/v1/wallets/transfer \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "fromWalletId": "W111",
    "toWalletId": "W222",
    "amount": "1000",
    "idempotencyKey": "unique-key-001"
  }'
```

响应：
```json
{"transactionId": "T9876543210"}
```

#### 查询交易记录

```bash
curl "http://localhost:8080/v1/wallets/transactions/{wallet_id}?offset=0&limit=10"
```

### gRPC API

gRPC 服务定义在 `api/proto/v1/wallet.proto`，服务名 `wallet.v1.WalletService`，默认监听端口 `9090`。

支持 gRPC 反射，可使用 grpcurl 等工具直接调用：

```bash
# 列出服务
grpcurl -plaintext localhost:9090 list

# 创建钱包
grpcurl -plaintext -d '{"name": "my-wallet", "passwd": "123456"}' \
  localhost:9090 wallet.v1.WalletService/CreateWallet

# 转账（需在 metadata 中携带 token）
grpcurl -plaintext \
  -H "authorization: Bearer <token>" \
  -d '{"from_wallet_id": "W111", "to_wallet_id": "W222", "amount": 1000, "idempotency_key": "key-001"}' \
  localhost:9090 wallet.v1.WalletService/Transfer
```

## 错误码

| 范围 | 类别 | 示例 |
|------|------|------|
| 1xxx | 通用错误 | 1001 内部错误, 1002 参数无效 |
| 2xxx | 钱包错误 | 2001 钱包不存在, 2003 钱包已冻结, 2004 密码错误 |
| 3xxx | 交易错误 | 3001 余额不足, 3005 重复交易 |
| 4xxx | 认证错误 | 4001 未认证, 4002 会话过期, 4003 无效令牌 |
| 5xxx | 并发错误 | 5001 获取锁超时, 5002 数据已被修改 |

## 快速开始

### 前置条件

- Go 1.25+
- （可选）buf CLI — 仅在需要重新生成 protobuf 代码时使用

### 启动服务

```bash
go run ./cmd/server
```

服务启动后：
- HTTP REST API：`http://localhost:8080`
- gRPC API：`localhost:9090`

### 配置

支持三种配置方式（优先级从高到低）：

1. **环境变量**

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `APP_ENV` | 运行环境 | development |
| `APP_HTTP_PORT` | HTTP 端口 | 8080 |
| `APP_GRPC_PORT` | gRPC 端口 | 9090 |
| `APP_MACHINE_ID` | 机器码 (0-63) | 1 |

2. **YAML 配置文件**（自动搜索 `config.yaml`、`config/config.yaml`、`/etc/wallet/config.yaml`）

```yaml
app:
  env: development
  http_port: 8080
  grpc_port: 9090
  machine_id: 1
session:
  ttl: 24h
```

3. **内置默认值**

### 重新生成 Protobuf 代码

```bash
cd api/proto/v1
buf generate
```

## 测试

### 功能测试流程

```bash
# 1. 启动服务
go run ./cmd/server

# 2. 创建两个钱包
curl -s -X POST http://localhost:8080/v1/wallets \
  -H "Content-Type: application/json" \
  -d '{"name": "alice", "passwd": "pass123"}'

curl -s -X POST http://localhost:8080/v1/wallets \
  -H "Content-Type: application/json" \
  -d '{"name": "bob", "passwd": "pass456"}'

# 3. 登录获取 token
curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"walletId": "<alice_wallet_id>", "passwd": "pass123"}'

# 4. 使用 token 进行转账
curl -s -X POST http://localhost:8080/v1/wallets/transfer \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "fromWalletId": "<alice_wallet_id>",
    "toWalletId": "<bob_wallet_id>",
    "amount": "500",
    "idempotencyKey": "test-transfer-001"
  }'

# 5. 查询余额
curl -s http://localhost:8080/v1/wallets/<alice_wallet_id>
curl -s http://localhost:8080/v1/wallets/<bob_wallet_id>

# 6. 查询交易记录
curl -s "http://localhost:8080/v1/wallets/transactions/<alice_wallet_id>?offset=0&limit=10"
```


## 编译部署
```shell
docker compose up -d
````
服务默认开放 8080 http服务端口 和 9090 grpc服务端口

**自定义配置**: docker compose 运行的容器自动挂载运行目录下的config目录到容器内/app/config，服务自动加载/app/config/config.yaml
若要自定义配置，复制 config.example.yaml 为 config.yaml 并修改相关配置项
