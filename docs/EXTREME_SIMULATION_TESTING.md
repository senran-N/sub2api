# Extreme Simulation Testing

## 目标

- 给本地或预发环境提供一套可重复执行的“资源受限 + 突发请求 + 浏览器退化”测试入口。
- 把环境矩阵和 HTTP 极端场景脚本化，避免每次手工改 compose 参数。
- 把浏览器端的慢网、离线、移动端、CPU 降速验证固定成可复现步骤。

## 脚本入口

### 环境矩阵

- 路径：`tools/extreme_env_matrix.py`
- 依赖：仅 Python 3 标准库
- 能力：
  - 列出内置 profile
  - 输出 compose 可用的环境覆盖变量
  - 支持 `dotenv` / `shell` / `json` 三种格式

查看所有 profile：

```bash
python3 tools/extreme_env_matrix.py profiles
```

导出 `cpu-starved` profile 的 shell 环境变量：

```bash
eval "$(python3 tools/extreme_env_matrix.py env --profile cpu-starved --format shell)"
docker compose -f deploy/docker-compose.dev.yml up -d --build
```

导出 `gateway-buffer-pressure` profile 到临时 dotenv 文件：

```bash
python3 tools/extreme_env_matrix.py env \
  --profile gateway-buffer-pressure \
  --format dotenv \
  --output /tmp/sub2api-extreme.env

set -a
. /tmp/sub2api-extreme.env
set +a
docker compose -f deploy/docker-compose.dev.yml up -d --build
```

### HTTP 极端探针

- 路径：`tools/http_extreme_probe.py`
- 依赖：仅 Python 3 标准库
- 能力：
  - 并发压测公开接口、登录接口、管理端 realtime 接口、OpenAI 兼容网关
  - 输出每个采样周期的吞吐、状态码速率、平均延迟
  - 输出最终 `avg/p50/p95/max latency` 与 `effective_rps`
  - 支持默认场景套件和单场景重放

默认场景套件：

- `public-settings`
- `auth-invalid-password`
- `admin-realtime`
- `gateway-auth-reject`

默认套件示例：

```bash
python3 tools/http_extreme_probe.py \
  --base-url http://127.0.0.1:8080 \
  --admin-email admin@sub2api.local \
  --admin-password sub2api-admin-pass \
  --scenario default-suite \
  --qps 40 \
  --concurrency 16 \
  --warmup-seconds 5 \
  --measure-seconds 30 \
  --output-json /tmp/sub2api-http-extreme.json
```

只压管理端 realtime：

```bash
python3 tools/http_extreme_probe.py \
  --base-url http://127.0.0.1:8080 \
  --admin-email admin@sub2api.local \
  --admin-password sub2api-admin-pass \
  --scenario admin-realtime \
  --qps 25 \
  --concurrency 8 \
  --measure-seconds 20
```

大请求体拒绝场景（建议搭配 `gateway-buffer-pressure`）：

```bash
python3 tools/http_extreme_probe.py \
  --base-url http://127.0.0.1:8080 \
  --scenario gateway-large-body-reject \
  --qps 2 \
  --concurrency 2 \
  --measure-seconds 15 \
  --large-body-bytes 9437184
```

## 内置环境 Profile

### `baseline`

- 对照组，不额外覆盖 compose 变量。

### `cpu-starved`

- 压低 `SUB2API_CPUS` / `SUB2API_MEMORY_LIMIT`
- 同时收紧数据库与 Redis 连接池
- 适合观察后台表格、登录、公共配置读取在资源紧张时的抖动

### `connection-pressure`

- 降低 PostgreSQL `max_connections`
- 缩小应用连接池和 Redis client 容量
- 适合观察高并发下连接等待、重试堆积、管理端读接口抖动

### `redis-pressure`

- 压缩 Redis 内存与 client 上限
- 适合观察限流、会话、缓存抖动时的级联退化

### `gateway-buffer-pressure`

- 降低 `SERVER_MAX_REQUEST_BODY_SIZE`、`GATEWAY_MAX_BODY_SIZE`
- 收紧 h2c stream / upload buffer
- 适合验证超大请求体和流式入口是否能尽早失败，而不是拖垮进程
- 在极端大 body 场景下，客户端可能看到 `413`，也可能看到连接被服务端提前切断的 transport error；两者都说明早期拒绝生效

## 推荐组合

### 1. 登录与后台读取回归

```bash
eval "$(python3 tools/extreme_env_matrix.py env --profile cpu-starved --format shell)"
docker compose -f deploy/docker-compose.dev.yml up -d --build

python3 tools/http_extreme_probe.py \
  --base-url http://127.0.0.1:8080 \
  --admin-email admin@sub2api.local \
  --admin-password sub2api-admin-pass \
  --scenario public-settings \
  --scenario auth-invalid-password \
  --scenario admin-realtime \
  --qps 40 \
  --concurrency 16
```

### 2. 网关请求体边界

```bash
eval "$(python3 tools/extreme_env_matrix.py env --profile gateway-buffer-pressure --format shell)"
docker compose -f deploy/docker-compose.dev.yml up -d --build

python3 tools/http_extreme_probe.py \
  --base-url http://127.0.0.1:8080 \
  --scenario gateway-auth-reject \
  --scenario gateway-large-body-reject \
  --qps 4 \
  --concurrency 2 \
  --large-body-bytes 9437184
```

### 3. OpenAI scheduler 专项容量

如果要看 sticky / non-sticky miss-path 压力，不要重复造轮子，直接复用：

- [docs/OPENAI_SCHEDULER_CAPACITY_TEST.md](/home/senran/Desktop/sub2api/docs/OPENAI_SCHEDULER_CAPACITY_TEST.md)
- `tools/openai_scheduler_capacity_probe.py`

推荐把 scheduler 探针和本页的 profile 组合起来用，例如：

```bash
eval "$(python3 tools/extreme_env_matrix.py env --profile connection-pressure --format shell)"
docker compose -f deploy/docker-compose.dev.yml up -d --build
```

然后再运行 `openai_scheduler_capacity_probe.py`。

## DevTools MCP 浏览器退化检查

HTTP 探针覆盖不到浏览器线程、布局、脚本执行和离线状态，所以每次核心改动后至少补下面 4 组 DevTools 场景：

### 1. 移动端窄屏

- 视口：`390x844x3,mobile`
- 页面：`/login`、`/admin/dashboard`
- 关注点：
  - 首屏是否出现横向滚动
  - 登录按钮和错误提示是否被软遮挡
  - 后台侧栏是否还能到达主要入口

### 2. 慢网

- 网络：`Slow 4G`
- CPU：`4x`
- 页面：`/login`
- 动作：
  - 刷新页面
  - 观察公开配置注入是否失败
  - 提交一次登录
- 关注点：
  - 控制台不能出现未捕获异常
  - loading 状态必须可见且能恢复

### 3. 离线

- 网络：`Offline`
- 页面：`/login`、已登录后的 `/admin/dashboard`
- 动作：
  - 刷新
  - 已登录页切换一次路由
- 关注点：
  - 页面不能白屏
  - 需要保留明确错误态，而不是无限 loading

### 4. 高频刷新

- 网络：`Fast 4G` 或默认
- CPU：`2x` 或 `4x`
- 页面：`/admin/dashboard`
- 动作：
  - 连续 reload 5 次
  - 打开 Network 面板确认是否出现异常 5xx / 重复爆发请求

## 建议输出

做完极端测试后，建议至少保留：

- 使用的 compose profile
- `http_extreme_probe.py` 的 JSON 报告
- DevTools MCP 的网络/控制台异常结论
- 如果涉及 scheduler，再附上 `openai_scheduler_capacity_probe.py` 的结果

这样才能把“后端吞吐退化”和“前端交互退化”对齐到同一轮验证。
