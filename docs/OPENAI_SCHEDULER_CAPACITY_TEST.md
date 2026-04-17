# OpenAI Scheduler Capacity Test

## 目标

- 容量压测要分开看高 sticky 命中和低 sticky 命中。
- 不要只盯账户总数，重点看每秒到底有多少请求掉进非 sticky path。
- 对 OpenAI 调度链路，最关键的实时量不是总 QPS，而是：
  - `non_sticky_intent_total` 的增速
  - `sticky_miss_fallback_total` 的增速
  - `indexed_load_balance_select_total` 的增速
  - `sticky_miss_indexed_select_total` 的增速

`indexed_load_balance_select_total` 的每秒速率最接近“当前真正压到昂贵 miss path 的请求量”。

## 脚本

- 路径：`tools/openai_scheduler_capacity_probe.py`
- 依赖：仅 Python 3 标准库
- 能力：
  - 发 OpenAI 兼容请求到网关
  - 自动区分 `high-sticky` / `low-sticky` 两种场景
  - 用 `session_id` + `conversation_id` 驱动 sticky / non-sticky 会话
  - 周期性拉取 `/api/v1/admin/ops/realtime-traffic?window=1min`
  - 把累积调度计数转成每秒速率，直接输出 non-sticky path 压力

## 推荐用法

### 1. 同一个 group 的高 sticky 命中场景

```bash
python3 tools/openai_scheduler_capacity_probe.py \
  --gateway-base-url http://127.0.0.1:8080 \
  --admin-base-url http://127.0.0.1:8080/api/v1 \
  --gateway-api-key sk_live_xxx \
  --admin-token admin_jwt_xxx \
  --platform openai \
  --group-id 12 \
  --model gpt-5.1 \
  --scenario high-sticky \
  --qps 60 \
  --concurrency 24 \
  --sticky-pool-size 128 \
  --warmup-seconds 15 \
  --measure-seconds 60
```

### 2. 低 sticky 命中 / 新会话场景

```bash
python3 tools/openai_scheduler_capacity_probe.py \
  --gateway-base-url http://127.0.0.1:8080 \
  --admin-base-url http://127.0.0.1:8080/api/v1 \
  --gateway-api-key sk_live_xxx \
  --admin-token admin_jwt_xxx \
  --platform openai \
  --group-id 12 \
  --model gpt-5.1 \
  --scenario low-sticky \
  --qps 60 \
  --concurrency 24 \
  --warmup-seconds 10 \
  --measure-seconds 60
```

### 3. 两个场景连续跑

```bash
python3 tools/openai_scheduler_capacity_probe.py \
  --gateway-base-url http://127.0.0.1:8080 \
  --admin-base-url http://127.0.0.1:8080/api/v1 \
  --gateway-api-key sk_live_xxx \
  --admin-token admin_jwt_xxx \
  --platform openai \
  --group-id 12 \
  --model gpt-5.1 \
  --scenario both \
  --qps 80 \
  --concurrency 32 \
  --sticky-pool-size 256 \
  --warmup-seconds 15 \
  --measure-seconds 90 \
  --output-json /tmp/openai-capacity.json
```

## 场景解释

### `high-sticky`

- 使用固定大小的 session 池反复复用 `session_id` / `conversation_id`
- 目标是看 steady-state 下 sticky 命中后，`indexed_load_balance_rps` 能否接近 0
- 重点观察：
  - `sticky_hit` 是否稳定偏高
  - `indexed_lb_rps` 是否接近 0
  - `sticky_miss_rps` 是否只在 warmup 初期短暂抬头

### `low-sticky`

- 每个请求都生成新的 `session_id` / `conversation_id`
- 目标是测“每秒有多少请求必须走非 sticky path”
- 重点观察：
  - `non_sticky_rps`
  - `indexed_lb_rps`
  - `qps_current`
  - 4xx/5xx/transport error 是否开始抬升

## 输出解释

脚本每个采样周期会打印类似：

```text
[low-sticky ] + 15.0s qps=  58.20 non_sticky_rps= 57.90 sticky_miss_rps=  0.20 indexed_lb_rps= 58.00 ...
```

关键字段：

- `non_sticky_rps`
  - `non_sticky_intent_total` 在这个采样窗口里的增速
  - 更接近“新会话直接走非 sticky path”的请求量
- `sticky_miss_rps`
  - `sticky_miss_fallback_total` 的增速
  - 表示本来有 sticky intent，但命中失败后掉入 fallback 的请求量
- `indexed_lb_rps`
  - `indexed_load_balance_select_total` 的增速
  - 这是最关键的 miss-path 压力指标
- `sticky_hit` / `sticky_miss`
  - 来自 runtime observability summary 的累计比例
  - 更适合看趋势，不适合直接当作瞬时速率

## 压测结论怎么读

- 如果 `high-sticky` 下：
  - `indexed_lb_rps` 很低
  - `sticky_hit` 高
  - 但系统仍然扛不住
  - 问题通常不在 miss path，而在上游/并发/限流/网络

- 如果 `low-sticky` 下：
  - `indexed_lb_rps` 随 QPS 线性抬升
  - 并且 4xx/5xx/transport error 或平均延迟明显恶化
  - 当前容量瓶颈就更接近非 sticky path

- 真正要汇报的容量数字建议至少给两组：
  - 高 sticky 命中时的稳定 QPS
  - 低 sticky 命中时的稳定 `indexed_lb_rps`

## 参数建议

- `--platform` / `--group-id`
  - 建议总是带上，避免其它业务流量污染 admin realtime 统计
- `--sticky-pool-size`
  - 太小会把高 sticky 场景压成“少量超热 session”
  - 通常用 `64`、`128`、`256` 比单 session 更接近真实情况
- `--qps`
  - 建议逐步爬坡，不要一步打满
  - 例如 `20 -> 40 -> 60 -> 80 -> 120`
- `--request-body-file`
  - 若线上主要走自定义 body 或 `/v1/chat/completions`，可以提供真实请求体文件复用

## 注意点

- 这个脚本的目标是量化 scheduler 路径压力，不是替代专门的 HTTP benchmark 工具。
- 默认 body 走 OpenAI 兼容 JSON；如果你的真实流量主要是 `/v1/chat/completions`，把 `--request-path` 改成对应路径即可。
- 管理端统计依赖 realtime monitoring；如果 `/admin/ops/realtime-traffic` 返回 disabled，先打开监控再测。
- 若想把脚本结果纳入报告，建议保留 `--output-json` 输出，便于和系统资源图、上游错误率、账户可用性一起对齐。
