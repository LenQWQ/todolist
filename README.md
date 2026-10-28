# TodoList

一个带用户体系与鉴权的待办服务：注册 / 登录 / 退出、个人信息查看与修改、密码修改、待办增删改查，数据按用户隔离；提供可在浏览器操作的网页，并支持用 Docker Compose 一条命令部署。

## 功能

### 用户与鉴权
- 注册（用户名 / 邮箱全局唯一，密码 bcrypt 散列存储）
- 登录（签发不透明 Token，写入 `HttpOnly` Cookie）
- 退出（**删除服务端 Token 记录**，旧凭据立即失效）
- 查看个人信息
- 修改邮箱（唯一性校验，允许改为自己当前邮箱）
- 修改密码（需校验原密码）
- 全局鉴权中间件：除注册/登录外，所有接口默认要求登录
- 登录失败统一返回 401「用户名或密码错误」，防止账号枚举

### 待办
- 创建、列表、详情、更新（整体更新）、删除
- 完成状态勾选
- 参数校验与错误分类（400 / 401 / 404 / 409 / 500）
- **数据隔离**：只能操作自己的待办，访问他人待办返回 404

### 网页
- 单页应用（原生 HTML + fetch，无前端框架）
- 注册、登录、退出、个人信息、改邮箱、改密码、待办的增删改查与状态切换
- 操作结果与错误提示统一展示

## 技术栈

| 层 | 选型 | 版本 |
|---|---|---|
| 语言 | Go | 1.26.1 |
| Web 框架 | go-zero | v1.10.3 |
| ORM | GORM | v1.31.2 |
| 数据库 | MySQL | 8.0 |
| 密码散列 | golang.org/x/crypto/bcrypt | v0.55.0 |
| 容器 | Docker / Compose | 29.6.1 / v5.2.0 |

## 目录结构

```text
.
├── todo.go                     # 入口：加载配置、初始化 svc、注册中间件与路由、挂载静态页面
├── todo.api                    # 接口契约（goctl 生成 types/handler/routes 的源文件）
├── etc/
│   ├── todo-api.yaml.example   # 本地配置样例（复制为 todo-api.yaml）
│   └── todo-api.docker.yaml    # 容器内配置（只含 ${MYSQL_DSN} 占位符）
├── internal/
│   ├── config/                 # 配置结构
│   ├── errs/                   # 业务错误（CodeError：状态码 + 提示）
│   ├── handler/                # HTTP 层（goctl 生成 + 手写错误处理）+ 测试
│   ├── logic/                  # 业务逻辑
│   ├── middleware/             # 鉴权中间件（Cookie → Token → user_id 注入 context）
│   ├── repository/             # 数据访问（接口 + GORM 实现 + 模型）
│   ├── svc/                    # 依赖注入容器
│   └── types/                  # 请求/响应类型（goctl 生成）
├── sql/init.sql                # 建表脚本（todos / users / tokens）
├── web/index.html              # 网页（静态资源，由 rest.WithFileServer 提供）
├── docs/api.md                 # 接口文档
├── docs/验证记录.md             # 浏览器操作 + Docker 验证记录
├── Dockerfile
├── compose.yaml
└── .env.example
```

## 快速开始

### 方式一：Docker Compose

前置：安装 Docker Desktop。

```powershell
# 1. 准备环境变量（密码不进镜像、不进 git）
Copy-Item .env.example .env
# 编辑 .env，设置 MYSQL_ROOT_PASSWORD

# 2. 构建并启动（app + mysql）
docker compose up -d --build

# 3. 查看状态（mysql 需显示 healthy 后 app 才启动）
docker compose ps

# 4. 看日志
docker compose logs -f app
```

浏览器打开 **http://localhost:8888/** 即可使用。

> 首次启动时 MySQL 会自动执行 `sql/init.sql` 建表（挂载到 `/docker-entrypoint-initdb.d/`），仅在数据卷为空时执行一次。

**停止与清理**

| 命令 | 效果 |
|---|---|
| `docker compose stop` / `start` | 停止 / 启动容器，**数据保留** |
| `docker compose down` | 删除容器与网络，**数据保留**（命名卷仍在） |
| `docker compose down -v` | ⚠️ **连数据卷一起删除，数据全部丢失** |

**进入数据库排查**（MySQL 未对宿主机暴露端口）

```powershell
docker compose exec mysql mysql -uroot -p -e "USE todo; SELECT id,username,email FROM users;"
```

### 方式二：本地运行

前置：Go 1.26+、MySQL 8.0。

```powershell
# 1. 建库并执行建表脚本
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS todo CHARACTER SET utf8mb4;"
mysql -u root -p todo -e "source sql/init.sql"

# 2. 准备配置
Copy-Item etc\todo-api.yaml.example etc\todo-api.yaml
# 编辑 etc\todo-api.yaml，把 Mysql.DataSource 改成你自己的账号密码

# 3. 运行
go run todo.go
```

浏览器打开 **http://localhost:8888/**。

## 配置说明

| 键 | 说明 |
|---|---|
| `Name` | 服务名（日志标识） |
| `Host` / `Port` | 监听地址。**容器内必须是 `0.0.0.0`**，本地可用 `127.0.0.1` |
| `Mysql.DataSource` | MySQL DSN，必须包含 `parseTime=True`（GORM 读写 `time.Time` 需要） |
| `Token.ExpireSeconds` | 登录凭据有效期（秒），同时用于 Cookie 的 `Max-Age` |
| `Token.CookieSecure` | Cookie 是否仅 HTTPS 发送。本地 `false`，**生产应设 `true`** |

容器环境下 `Mysql.DataSource` 写成 `${MYSQL_DSN}`，由环境变量注入（依赖 `conf.UseEnv()`），因此**密码不会进入镜像**。

## 接口文档

见 [docs/api.md](docs/api.md)，覆盖：

- 认证方式（Cookie 属性、Token 生命周期）
- 错误响应格式与状态码约定
- 用户模块 6 个接口
- 待办模块 5 个接口
- 数据归属与权限规则

## 数据库

| 表 | 说明 |
|---|---|
| `users` | id、username（唯一、不可改）、email（唯一、可改）、password_hash、created_at、updated_at |
| `tokens` | id、token（唯一索引）、user_id（索引）、expires_at、created_at |
| `todos` | id、user_id（索引）、titles、content、done、created_at、updated_at |

建表脚本：`sql/init.sql`。

## 测试

```powershell
cd projects/todo
go test ./internal/handler/ -v
```

**共 37 个用例**，覆盖：

- 待办 CRUD、非法参数、不存在
- 注册（成功 / 用户名重复 / 邮箱重复 / 空字段）
- 登录（成功 / 密码错误 / 用户不存在且提示一致）
- 鉴权（未登录 401 / Token 过期 401 / **退出后旧 Token 失效**）
- 个人信息（查看 / 改邮箱成功 / 邮箱冲突 409 / 改为自己邮箱幂等）
- 改密码（旧密码错误 400 / 成功且新旧密码校验结果相反）
- **越权**（他人列表不可见 / 他人详情、更新、删除均 404 且数据未被篡改）

测试使用内存假仓库（`fakeTodoRepo` / `fakeUserRepo` / `fakeTokenRepo`），假仓库**同样实现 `user_id` 过滤**，避免"假实现比真实现宽松"导致测试失真。

### 集成测试（连真实 MySQL）

单元测试使用假仓库，**不会执行真实 SQL**——`First(&id)` 这类写错的查询、缺失的 `WHERE` 条件、`RowsAffected` 语义错误，单测全绿也发现不了。因此另有一组连接真实 MySQL 的集成测试，覆盖仓库层的完整往返。

```powershell
# 1. 建一个独立的测试库（不要用开发库，测试会清表）
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS todo_test CHARACTER SET utf8mb4;"

# 2. 设置 DSN 并运行
$env:TEST_MYSQL_DSN="root:你的密码@tcp(127.0.0.1:3306)/todo_test?charset=utf8mb4&parseTime=True"
go test ./internal/repository/ -run Integration -v
```

- 未设置 `TEST_MYSQL_DSN` 时自动 **skip**（CI 无数据库也能跑其余测试）
- 覆盖：`Create → FindById → FindAll → Update → Delete` 往返、按 id 精确查询、他用户数据不可见/不可改/不可删、重复删除返回"不存在"

> 实测有效的证据：把 `FindById` 里的 `First(&todo)` 改回 `First(&id)` 后，**37 个单元测试仍然全过**，而集成测试立即失败并指出具体原因。

## 验证记录

见 [docs/验证记录.md](docs/验证记录.md)：包含浏览器操作记录、Docker 构建/启动记录、数据持久化三连测。

## 已知问题 / 限制

1. **建表有两套来源**：`sql/init.sql`（Docker 首次启动 / 手动执行）与 GORM `AutoMigrate`（应用启动时）。两者存在结构漂移风险，后续应统一为一种。
2. **无刷新 Token**：Token 有效期 24 小时，过期后需重新登录，没有续期机制。
3. **改密码后其他设备的旧 Token 不失效**：当前只更新密码散列，未清理该用户的其他会话。生产环境建议在改密后删除该用户全部 Token。
4. **无集成测试**：现有测试通过假仓库验证 HTTP 层逻辑，未连真实 MySQL 验证 GORM 行为（如唯一约束冲突、数据库故障）。
5. **网页为最小实现**：仅保证功能可用，无视觉设计。
6. 前端静态文件与 API 同源部署，未单独处理 CORS。

## 安全说明

已实现：

- 密码使用 **bcrypt** 散列存储，不落库明文；接口响应不返回散列
- 登录凭据为 **`crypto/rand`** 生成的 256 bit 随机值，不可预测；服务端存库并带过期时间
- 凭据放 **`HttpOnly` Cookie**，JS 无法读取，降低 XSS 窃取风险
- 登录失败**统一提示**，避免账号枚举
- 鉴权中间件**默认拒绝**（白名单仅注册/登录），新增接口默认受保护
- **越权访问返回 404**，不暴露数据是否存在
- 所有数据库查询按 `user_id` 过滤，用户身份只来自认证上下文，**不信任请求体**
- 网页渲染用户内容使用 `textContent`，避免 XSS
- 密码、密钥通过环境变量注入，`.env` 与本地配置均被 `.gitignore` 排除

部署到公网前还需：

- 启用 HTTPS，并将 `Token.CookieSecure` 设为 `true`
- 更换默认数据库密码，限制数据库端口暴露
- 增加登录失败频率限制 / 账号锁定
- 为 Token 增加定期清理过期记录的任务
