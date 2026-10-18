# TodoList 接口文档

## 通用约定

- Base URL：`http://localhost:8888`
- 请求格式：`Content-Type: application/json`
- 时间字段：Unix 时间戳（秒），由服务端生成
- **认证方式**：登录成功后服务端写入 `HttpOnly` Cookie（名为 `token`），后续请求由浏览器自动携带；服务端校验该 Token 是否存在于数据库且未过期

## 认证说明

登录成功后响应头包含：

```http
Set-Cookie: token=<64位hex>; Path=/; Max-Age=86400; HttpOnly; SameSite=Lax
```

| 属性 | 值 | 作用 |
|---|---|---|
| `HttpOnly` | true | 禁止 JS 读取，防 XSS 窃取凭据 |
| `SameSite` | Lax | 缓解 CSRF |
| `Path` | / | 全站生效 |
| `Max-Age` | 86400 | 与 Token 服务端有效期一致（24 小时） |

- Token 为**不透明字符串**（`crypto/rand` 生成 32 字节 → hex，共 64 字符），服务端存于 `tokens` 表，含 `expires_at`
- **退出登录会删除服务端记录**，旧 Token 立即失效（不是仅清除客户端 Cookie）
- 未携带凭据 / Token 不存在 / Token 已过期 → **401**

## 错误响应格式

- 错误类别由 **HTTP 状态码**表达（400 / 401 / 404 / 409 / 500）
- 响应体是 **JSON 字符串**（不是对象），内容为错误提示文本
- `Content-Type: application/json; charset=utf-8`

**示例**

```http
HTTP/1.1 404 Not Found
Content-Type: application/json; charset=utf-8
"备忘录不存在"
```

## 状态码约定

| 状态码 | 含义 | 触发场景 |
|---|---|---|
| 200 | 成功 | 响应成功 |
| 400 | 请求内容有问题 | 参数缺失/格式错误/标题为空/原密码错误 |
| 401 | 未认证或凭据失效 | 未登录、Token 不存在、Token 过期、登录失败（用户名或密码错误，统一提示） |
| 404 | 不存在或无权访问 | 所查询内容不存在；**访问他人待办也返回 404**（不暴露存在性） |
| 409 | 冲突 | 注册时用户名/邮箱已存在；修改邮箱时邮箱已被占用 |
| 500 | 服务器内部错误 | 数据库操作失败等 |

---

# 一、用户模块

## 1. 注册

`POST /user/register`　（无需登录）

**请求体**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| username | string | 是 | 用户名，全局唯一，不可修改 |
| password | string | 是 | 明文密码，服务端用 bcrypt 散列后存储 |
| email | string | 是 | 邮箱，全局唯一 |

**响应**：200 + 用户信息

```json
{"id":1,"username":"alice","email":"alice@example.com","created_at":1791538968,"updated_at":1791538968}
```

> 响应中**不包含** `password_hash` 字段。

**错误**
- 400 —— 用户名/密码/邮箱为空（body：`"用户名/密码/邮箱不能为空"`）
- 409 —— 用户名或邮箱已存在（body：`"用户名/邮箱已被注册"`）

**示例**
```bash
curl.exe -X POST http://localhost:8888/user/register -H "Content-Type: application/json" --% -d "{\"username\":\"alice\",\"password\":\"secret123\",\"email\":\"alice@example.com\"}"
```

## 2. 登录

`POST /user/login`　（无需登录）

**请求体**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |

**响应**：200 + 用户信息，并写入 `Set-Cookie`

**错误**
- 400 —— 用户名或密码为空
- 401 —— **用户名不存在 或 密码错误，返回同一个错误**（body：`"用户名或密码错误"`）

> 统一提示的目的：**防止账号枚举**。若两者提示不同，攻击者可据此探测系统内存在哪些用户名。

**示例**
```bash
curl.exe -i -c cookies.txt -X POST http://localhost:8888/user/login -H "Content-Type: application/json" --% -d "{\"username\":\"alice\",\"password\":\"secret123\"}"
```

## 3. 退出登录

`POST /user/logout`　（**需要登录**）

**响应**：200（空响应体），同时清除客户端 Cookie

- 服务端**删除该 Token 记录** → 旧 Token 立即失效
- 响应头包含 `Set-Cookie: token=; Path=/; Max-Age=0; HttpOnly`

**错误**
- 401 —— 未登录或凭据已失效

**示例**
```bash
curl.exe -i -b cookies.txt -X POST http://localhost:8888/user/logout
```

## 4. 查看个人信息

`GET /user/me`　（**需要登录**）

**响应**：200 + 当前登录用户的信息

**错误**：401 —— 未登录 / Token 失效

**示例**
```bash
curl.exe -i -b cookies.txt http://localhost:8888/user/me
```

## 5. 修改邮箱

`PUT /user/me`　（**需要登录**）

**请求体**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| email | string | 是 | 新邮箱，需全局唯一 |

**响应**：200 + **更新后**的用户信息（含新的 `updated_at`）

**错误**
- 400 —— 邮箱为空
- 409 —— 邮箱已被其他用户占用（body：`"邮箱已被注册"`）
- 401 —— 未登录

> 允许把邮箱改成自己当前的邮箱（幂等，返回 200）；用户名不可修改。

**示例**
```bash
curl.exe -b cookies.txt -X PUT http://localhost:8888/user/me -H "Content-Type: application/json" --% -d "{\"email\":\"new@example.com\"}"
```

## 6. 修改密码

`PUT /user/me/password`　（**需要登录**）

**请求体**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| old_password | string | 是 | 原密码，用于校验身份 |
| new_password | string | 是 | 新密码 |

**响应**：200（空响应体）

**错误**
- 400 —— 两个字段为空（`"用户名/密码/邮箱不能为空"`）
- 400 —— 原密码不正确（body：`"原密码错误"`）
- 401 —— 未登录

**示例**
```bash
curl.exe -b cookies.txt -X PUT http://localhost:8888/user/me/password -H "Content-Type: application/json" --% -d "{\"old_password\":\"secret123\",\"new_password\":\"newpass456\"}"
```

---

# 二、待办模块

> 以下所有接口**均需登录**。服务端从认证上下文（Cookie → Token → `user_id`）确定数据归属，**不接受请求体传入的 user_id**。
>
> **数据隔离**：用户只能查看/修改/删除**自己的**待办；访问他人的待办一律返回 **404**（不暴露该数据是否存在）。

## 1. 创建待办

`POST /todos`

**请求体**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| titles | string | 是 | 待办标题，不能为空 |
| content | string | 否 | 备忘录详细内容 |

**响应**：200 + Todo 对象（`user_id` 由服务端从登录态写入）

**错误**：400 —— 标题为空 / JSON 格式错误；401 —— 未登录

**示例**
```bash
curl.exe -b cookies.txt -X POST http://localhost:8888/todos -H "Content-Type: application/json" --% -d "{\"titles\":\"学 Go\",\"content\":\"内容\"}"
```

## 2. 查询详细备忘录

`GET /todos/:id`

**响应**：200 + Todo 详细信息

**错误**
- 400 —— `id` 不是数字
- 404 —— 不存在 **或 不属于当前用户**（body：`"备忘录不存在"`）
- 401 —— 未登录

**示例**
```bash
curl.exe -b cookies.txt -i http://localhost:8888/todos/1
```

## 3. 罗列全部备忘录

`GET /todos`

**响应**：200 + **当前用户的**待办列表

```json
{"todos":[{"id":5,"titles":"学 Go","done":false,"content":"内容","created_at":1792054307,"updated_at":1792054307}]}
```

**错误**：401 —— 未登录

**示例**
```bash
curl.exe -b cookies.txt -i http://localhost:8888/todos
```

## 4. 更新备忘录

`PUT /todos/:id`

**请求体**

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| titles | string | 是 | 要更改的标题 |
| content | string | 否 | 要更改的内容 |
| done | bool | 否 | 完成状态 |

**响应**：200 + 更新后的完整信息

> ⚠️ **整体更新语义**：未传的字段会被重置（例如只传 titles 不传 done，`done` 会变回 `false`）。想要部分更新请先查询完整数据再整体提交。

**错误**
- 400 —— 标题为空
- 404 —— 不存在或不属于当前用户
- 401 —— 未登录

**示例**
```bash
curl.exe -b cookies.txt -X PUT http://localhost:8888/todos/1 -H "Content-Type: application/json" --% -d "{\"titles\":\"改后的标题\",\"content\":\"改后的内容\",\"done\":true}"

# 只想改标题，没传 done
curl.exe -b cookies.txt -X PUT http://localhost:8888/todos/1 -H "Content-Type: application/json" --% -d "{\"titles\":\"新标题\"}"
# 结果：done 被重置为 false
```

## 5. 删除备忘录

`DELETE /todos/:id`

**响应**：200

**错误**
- 404 —— 不存在或不属于当前用户
- 401 —— 未登录

**示例**
```bash
curl.exe -b cookies.txt -X DELETE http://localhost:8888/todos/1
```

---

# 三、数据归属与权限

| 规则 | 实现位置 |
|---|---|
| 未登录不能访问受保护接口 | `internal/middleware/auth.go`（全局中间件 + 公开路径白名单，默认拒绝） |
| 身份来源可信 | 中间件从 Cookie 校验 Token 后，把 `user_id` 放入 context；logic 只从 context 取 |
| 只能操作自己的数据 | `TodoRepository` 的 5 个方法**全部要求传入 `userId`**，SQL 条件带 `user_id = ?` |
| 越权返回 404 | 查询条件不匹配 → 视为"不存在" |
