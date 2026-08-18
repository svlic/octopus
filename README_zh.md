<div align="center">

<img src="web/public/logo.svg" alt="Octopus Logo" width="120" height="120">

### Octopus

**为个人打造的简单、美观、优雅的 LLM API 聚合服务**

简体中文 | [English](README.md)

</div>


## ✨ 特性

- 🔀 **多渠道聚合** - 统一接入 OpenAI Chat、OpenAI Responses、Anthropic、Gemini、火山引擎（Volcengine）渠道
- 🔄 **协议互转** - 对外支持 OpenAI Chat Completions、OpenAI Responses、Anthropic Messages，并在上游间自动转换
- 📁 **分组手动路由** - 以分组名作为客户端 `model`，手动指定当前上游项，并为每项设置思考强度
- 💰 **价格同步** - 从 models.dev 自动同步模型价格，支持手动覆盖
- 🔃 **模型同步** - 自动与渠道同步可用模型列表
- 📊 **数据统计** - 全面的请求统计、Token 消耗、费用追踪
- 🎨 **优雅界面** - 简洁美观的 Web 管理面板
- 🗄️ **多数据库支持** - 支持 SQLite、MySQL、PostgreSQL


## 🚀 快速开始

### 🐳 Docker 部署

#### 使用官方预构建镜像

直接运行：

```bash
docker run -d --name octopus -v /path/to/data:/app/data -p 8080:8080 bestrui/octopus
```

或者使用 Docker Compose：

```bash
wget https://raw.githubusercontent.com/bestruirui/octopus/refs/heads/master/docker-compose.yml
docker compose up -d
```

> 上述方式使用 Docker Hub 上的 `bestrui/octopus` 预构建镜像，不会包含你在其他分支或本地源码中的修改。

#### 使用当前分支源码通过 Docker Compose 部署

以下流程会先构建当前检出分支的前端和 Go 二进制，再由 Docker Compose 构建本地镜像。以 `dyna` 分支和 Linux AMD64 服务器为例。

**环境要求：**

- Git
- Go 1.26.4
- Node.js 18+
- pnpm
- Docker 与 Docker Compose

1. 克隆仓库并切换到需要部署的分支：

```bash
git clone -b dyna https://github.com/svlic/octopus.git
cd octopus
```

如果已经克隆过上游仓库，先添加 fork 为单独的远程仓库：

```bash
git remote add svlic https://github.com/svlic/octopus.git
git fetch svlic
git switch dyna
git pull --ff-only svlic dyna
```

如果 `svlic` 远程仓库已经存在，则不需要再次执行 `git remote add`。通过前面的 `git clone -b dyna ...` 命令首次克隆时，后续使用 `git pull --ff-only origin dyna` 更新。

2. 构建需要嵌入 Go 二进制的前端静态文件：

```bash
cd web
pnpm install --frozen-lockfile
pnpm run build
cd ..
```

3. 为服务器架构构建 Go 二进制：

Linux AMD64（常见的 Intel/AMD 服务器）：

```bash
mkdir -p build/docker/linux/amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -tags=jsoniter -o build/docker/linux/amd64/octopus .
```

Linux ARM64：

```bash
mkdir -p build/docker/linux/arm64
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -tags=jsoniter -o build/docker/linux/arm64/octopus .
```

可以通过 `uname -m` 查看服务器架构：`x86_64` 对应 `linux/amd64`，`aarch64` 或 `arm64` 对应 `linux/arm64`。

4. 在项目根目录新建或替换 `docker-compose.yml`。以下示例适用于 Linux AMD64；ARM64 服务器将 `linux/amd64` 改为 `linux/arm64`：

```yaml
services:
  octopus:
    build:
      context: .
      dockerfile: scripts/dockerfiles/Dockerfile.debian
      args:
        TARGETPLATFORM: linux/amd64
    image: octopus:dyna
    container_name: octopus
    ports:
      - "8080:8080"
    volumes:
      - "/path/to/data:/app/data"
    restart: unless-stopped
```

请将 `/path/to/data` 替换为宿主机上实际的数据目录，然后启动：

```bash
docker compose up -d --build
```

查看状态和日志：

```bash
docker compose ps
docker compose logs -f octopus
```

以后同步 `dyna` 分支的新代码并重新部署。使用 `git clone -b dyna ...` 克隆的仓库执行：

```bash
git pull --ff-only origin dyna

cd web
pnpm install --frozen-lockfile
pnpm run build
cd ..

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -tags=jsoniter -o build/docker/linux/amd64/octopus .

docker compose up -d --build
```

ARM64 服务器在更新命令中同样使用 `GOARCH=arm64` 和 `build/docker/linux/arm64/octopus`。`/app/data` 已挂载到宿主机，因此重新构建和创建容器不会删除已有配置与数据库；仍建议在升级前备份宿主机数据目录。

### 📦 从 Release 下载

从 [Releases](https://github.com/bestruirui/octopus/releases) 下载对应平台的二进制文件，然后运行：

```bash
./octopus start
```

### 🛠️ 源码运行

**环境要求：**
- Go 1.26.4
- Node.js 18+
- pnpm

```bash
# 克隆项目
git clone https://github.com/bestruirui/octopus.git
cd octopus
# 构建前端
cd web && pnpm install && pnpm run build && cd ..
# 启动后端服务
go run main.go start 
```

> 💡 **提示**：前端构建产物会被嵌入到 Go 二进制文件中，所以必须先构建前端再启动后端。

**开发模式**

```bash
cd web && pnpm install && VITE_PROXY_TARGET="http://127.0.0.1:8080" pnpm run dev
## 新建终端,启动后端服务
go run main.go start
## 访问前端地址
http://localhost:5173
```

### 🔐 默认账户

首次启动后，访问 http://localhost:8080 使用以下默认账户登录管理面板：

- **用户名**：`admin`
- **密码**：`admin`

> ⚠️ **安全提示**：请在首次登录后立即修改默认密码。

### 📝 配置文件

配置文件默认位于 `data/config.json`，首次启动时自动生成。

**完整配置示例：**

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080
  },
  "database": {
    "type": "sqlite",
    "path": "data/data.db"
  },
  "log": {
    "level": "info"
  }
}
```

**配置项说明：**

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `server.host` | 监听地址 | `0.0.0.0` |
| `server.port` | 服务端口 | `8080` |
| `database.type` | 数据库类型 | `sqlite` |
| `database.path` | 数据库连接地址 | `data/data.db` |
| `log.level` | 日志级别 | `info` |

**数据库配置：**

支持三种数据库：

| 类型 | `database.type` | `database.path` 格式 |
|------|-----------------|---------------------|
| SQLite | `sqlite` | `data/data.db` |
| MySQL | `mysql` | `user:password@tcp(host:port)/dbname` |
| PostgreSQL | `postgres` | `postgresql://user:password@host:port/dbname?sslmode=disable` |

**MySQL 配置示例：**

```json
{
  "database": {
    "type": "mysql",
    "path": "root:password@tcp(127.0.0.1:3306)/octopus"
  }
}
```

**PostgreSQL 配置示例：**

```json
{
  "database": {
    "type": "postgres",
    "path": "postgresql://user:password@localhost:5432/octopus?sslmode=disable"
  }
}
```

> 💡 **提示**：MySQL 和 PostgreSQL 需要先手动创建数据库，程序会自动创建表结构。

**环境变量：**

所有配置项均可通过环境变量覆盖，格式为 `OCTOPUS_` + 配置路径（用 `_` 连接）：

| 环境变量 | 对应配置项 |
|----------|-----------|
| `OCTOPUS_SERVER_PORT` | `server.port` |
| `OCTOPUS_SERVER_HOST` | `server.host` |
| `OCTOPUS_DATABASE_TYPE` | `database.type` |
| `OCTOPUS_DATABASE_PATH` | `database.path` |
| `OCTOPUS_LOG_LEVEL` | `log.level` |


## 📸 界面预览

### 🖥️ 桌面端

<div align="center">
<table>
<tr>
<td align="center"><b>首页</b></td>
<td align="center"><b>渠道</b></td>
<td align="center"><b>分组</b></td>
</tr>
<tr>
<td><img src="web/public/screenshot/desktop-home.png" alt="首页" width="400"></td>
<td><img src="web/public/screenshot/desktop-channel.png" alt="渠道" width="400"></td>
<td><img src="web/public/screenshot/desktop-group.png" alt="分组" width="400"></td>
</tr>
<tr>
<td align="center"><b>价格</b></td>
<td align="center"><b>日志</b></td>
<td align="center"><b>设置</b></td>
</tr>
<tr>
<td><img src="web/public/screenshot/desktop-price.png" alt="价格" width="400"></td>
<td><img src="web/public/screenshot/desktop-log.png" alt="日志" width="400"></td>
<td><img src="web/public/screenshot/desktop-setting.png" alt="设置" width="400"></td>
</tr>
</table>
</div>

### 📱 移动端

<div align="center">
<table>
<tr>
<td align="center"><b>首页</b></td>
<td align="center"><b>渠道</b></td>
<td align="center"><b>分组</b></td>
<td align="center"><b>价格</b></td>
<td align="center"><b>日志</b></td>
<td align="center"><b>设置</b></td>
</tr>
<tr>
<td><img src="web/public/screenshot/mobile-home.png" alt="移动端首页" width="140"></td>
<td><img src="web/public/screenshot/mobile-channel.png" alt="移动端渠道" width="140"></td>
<td><img src="web/public/screenshot/mobile-group.png" alt="移动端分组" width="140"></td>
<td><img src="web/public/screenshot/mobile-price.png" alt="移动端价格" width="140"></td>
<td><img src="web/public/screenshot/mobile-log.png" alt="移动端日志" width="140"></td>
<td><img src="web/public/screenshot/mobile-setting.png" alt="移动端设置" width="140"></td>
</tr>
</table>
</div>


## 📖 功能说明

### 📡 渠道管理

渠道是连接 LLM 供应商的基础配置单元。

代码中支持的渠道类型：`openai`、`openai_responses`、`anthropic`、`gemini`、`volcengine`。

**Base URL 说明：**

程序会根据渠道类型自动补全 API 路径，您只需填写基础 URL 即可：

| 渠道类型 | 自动补全路径 | 填写 URL | 完整请求地址示例 |
|----------|-------------|----------|-----------------|
| OpenAI Chat | `/chat/completions` | `https://api.openai.com/v1` | `https://api.openai.com/v1/chat/completions` |
| OpenAI Responses | `/responses` | `https://api.openai.com/v1` | `https://api.openai.com/v1/responses` |
| Anthropic | `/messages` | `https://api.anthropic.com/v1` | `https://api.anthropic.com/v1/messages` |
| Gemini | `/models/:model:generateContent` | `https://generativelanguage.googleapis.com/v1beta` | `https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent` |
| 火山引擎 | `/chat/completions` | `https://ark.cn-beijing.volces.com/api/v3` | `https://ark.cn-beijing.volces.com/api/v3/chat/completions` |

> 💡 **提示**：填写 Base URL 时无需包含具体的 API 端点路径，程序会自动处理。火山引擎使用豆包/方舟 Chat Completions 协议，并在需要时将 Base URL 规范化为带 `v3` 后缀的形式。

每个渠道还可单独配置代理地址；渠道代理会覆盖该渠道上的全局代理。

---

### 📁 分组管理

分组用于将多个渠道模型聚合为一个统一的对外模型名称。路由为**手动切换**，不是自动负载均衡。

**核心概念：**

- **分组名称** 即程序对外暴露的模型名称。客户端请求的 `model` 必须填该名称。
- **当前项**（`active_item_id`）是实际承接流量的唯一上游渠道+模型。在管理面板中手动切换；`0` 表示尚未选择上游。
- **重试间隔**（`retry_interval`）为上游失败后的等待秒数（最小 `1`，默认 `1`）。
- **优先级**（`priority`）仅用于界面展示顺序，不影响路由。
- **思考强度**（`thinking_level`）可按分组项单独设置。内置预设：`default`、`none`、`low`、`medium`、`high`、`xhigh`。`default` 表示保留客户端自身的推理/思考设置；其他值会覆盖该项。界面也支持自定义字符串。

> 💡 **示例**：创建分组名称为 `gpt-4o`，将多个供应商的 GPT-4o 模型加入为分组项，将其中一项设为当前项，然后以 `model: gpt-4o` 调用网关即可。

---

### 💰 价格管理

管理面板中的 **价格** 页用于管理计费相关的模型价格（截图文件名仍为 `*-price.png`）。

**数据来源：**

- 系统会定期从 [models.dev](https://github.com/sst/models.dev) 同步模型价格数据（间隔可在设置中配置，默认 24 小时）
- 当创建渠道时，若渠道包含的模型不在 models.dev 中，系统会自动在此页面创建价格记录，便于手动定价
- 也支持手动创建 models.dev 中已存在的模型，用于自定义价格

**价格优先级：**

| 优先级 | 来源 | 说明 |
|:------:|------|------|
| 🥇 高 | 本页面 | 用户在价格管理页面设置的价格 |
| 🥈 低 | models.dev | 自动同步的默认价格 |

> 💡 **提示**：如需覆盖某个模型的默认价格，只需在价格管理页面为其设置自定义价格即可。

---

### ⚙️ 设置

以下为存入数据库的全局运行时设置（不是 `data/config.json`）：

| 键 | 含义 | 默认值 |
|----|------|--------|
| `proxy_url` | 全局上游 HTTP/HTTPS/SOCKS5 代理 | 空（直连） |
| `stats_save_interval` | 内存统计写入数据库的周期（**分钟**） | `10` |
| `model_info_update_interval` | 从 models.dev 同步模型价格/信息的周期（**小时**） | `24` |
| `sync_llm_interval` | 同步渠道模型列表的周期（**小时**） | `24` |
| `cors_allow_origins` | CORS 白名单（逗号分隔源站）。为空禁止跨域；`*` 允许所有 | 空 |

**统计落库策略：**

- 请求统计与转发日志先缓存在 **内存** 中
- 按统计保存周期 **批量写入** 数据库

> ⚠️ **重要提示**：退出程序时请使用正常关闭方式（`Ctrl+C` 或 `SIGTERM`），以便刷出缓冲中的统计数据。**请勿使用 `kill -9`**，否则可能导致近期统计丢失。

管理面板设置页还包含：账号密码、API Key、外观、备份/恢复、LLM 价格同步、LLM 模型同步、日志相关偏好等模块。

---

### 🔌 公开 LLM API

所有公开转发路由均需 API Key（`Authorization: Bearer <key>`，或客户端对应的供应商鉴权头）。当前支持的路径：

| 方法 | 路径 | 典型客户端 |
|------|------|------------|
| `POST` | `/v1/chat/completions` | OpenAI Chat Completions SDK / 兼容客户端 |
| `POST` | `/v1/responses` | OpenAI Responses API（如 Codex `wire_api = "responses"`） |
| `POST` | `/v1/messages` | Anthropic Messages API（如 Claude Code） |

当前代码**没有**公开的 embeddings、images 或 `/v1/models` 列表路由。请求中的 `model` 请填写管理面板中配置的**分组名称**。

---

## 🔌 客户端接入

### OpenAI SDK

```python
from openai import OpenAI
import os

client = OpenAI(   
    base_url="http://127.0.0.1:8080/v1",   
    api_key="sk-octopus-P48ROljwJmWBYVARjwQM8Nkiezlg7WOrXXOWDYY8TI5p9Mzg", 
)
completion = client.chat.completions.create(
    model="octopus-openai",  // 填写正确的分组名称
    messages = [
        {"role": "user", "content": "Hello"},
    ],
)
print(completion.choices[0].message.content)
```

### Claude Code

编辑 `~/.claude/settings.json`

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "http://127.0.0.1:8080",
    "ANTHROPIC_AUTH_TOKEN": "sk-octopus-",
    "API_TIMEOUT_MS": "3000000",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
    "ANTHROPIC_MODEL": "octopus-sonnet-4-5",
    "ANTHROPIC_SMALL_FAST_MODEL": "octopus-haiku-4-5",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "octopus-sonnet-4-5",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "octopus-sonnet-4-5",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "octopus-haiku-4-5"
  }
}
```

### Codex

编辑 `~/.codex/config.toml`

```toml
model = "gpt-5.6-sol"
model_reasoning_effort = "xhigh"
model_provider = "octopus"
preferred_auth_method = "apikey"

[model_providers.octopus]
base_url = "http://127.0.0.1:8080/v1"
name = "octopus"
supports_websockets = false
requires_openai_auth = true
wire_api = "responses"
experimental_bearer_token = "sk-octopus-"
```
编辑 `~/.codex/auth.json`

```json
{
  "OPENAI_API_KEY": ""
}
```


---

## 🤝 致谢

- 🙏 [looplj/axonhub](https://github.com/looplj/axonhub) - 本项目的 LLM API 适配模块直接源自该仓库的实现
- 📊 [sst/models.dev](https://github.com/sst/models.dev) - AI 模型数据库，提供模型价格数据
- 🇨🇳 [AtomGit](https://atomgit.com/bestruirui/octopus) - 国内代码托管
- 💬 [Linux.do](https://linux.do/)
