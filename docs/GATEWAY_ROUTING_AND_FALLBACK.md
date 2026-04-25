# Gateway Routing and Model Fallback

本文档记录兼容网关的优先级调度和模型回退语义。相关代码改动必须同步检查本文档、`deploy/config.example.yaml` 和调度/回退测试。

## Priority Mode

`gateway.scheduling.priority_mode` 控制账号优先级如何参与调度：

- `strict`：默认模式。`priority` 是硬分层，数值越小优先级越高。调度只会在当前优先级没有可用、非冷却、模型匹配且可获取或可生成等待计划的候选时，才进入下一优先级。
- `weighted`：兼容旧行为。`priority` 只作为评分维度之一，与负载、排队数、错误率和 TTFT 混合排序。

OpenAI 调度在 `strict` 下仍会在同一优先级内使用既有负载、排队、错误率和 TTFT 评分做均衡。处于 OpenAI WS transport fallback cooling 的账号会先让位给非 cooling 候选，避免刚失败的账号继续被优先级硬选中。

共享 load-aware fallback 路径本身已经按优先级排序；新增配置主要修正 OpenAI scheduler 的混合评分行为。

## Model Fallback

`group.default_mapped_model` 不再是隐式模型改写来源。它只在以下场景生效：

- `enable_model_fallback=true`，且请求模型无法选中账号，调度才允许用 `group.default_mapped_model` 再尝试一次。
- 已经发生上述显式回退时，转发层会用该模型改写上游请求。

`channel model_mapping` 是显式渠道映射，不受 `enable_model_fallback` 限制。渠道映射仍可用于调度模型和上游请求模型改写。

当模型回退关闭且请求模型不可调度时，网关应保留原始请求模型并返回选路错误，不应静默改写成默认模型。

兼容文本路由和 OpenAI-compatible passthrough 的显式模型回退由 `backend/internal/service/compatible_text_execution_kernel.go` 执行。OpenAI Responses WebSocket ingress 的初始显式模型回退由 `backend/internal/service/openai_ws_ingress_selection_kernel.go` 执行。Handler 只提供 fallback 解析 hook、HTTP/WebSocket 错误渲染、usage/runtime feedback hook；账号选择、fallback 后转发模型、failover 排除集和调度反馈必须保持在 service kernel 内，避免 `/responses`、`/chat/completions`、`/messages`、passthrough 和 WS ingress 各自复制循环。

原生 Gemini/Anthropic/Antigravity 路由的 load-aware 选号入口也通过 `backend/internal/service/selection_kernel.go`。在迁移完成前，通用 runtime failover 状态由 `backend/internal/service/runtime_failover_state.go` 承载。Handler 只消费 `HandleSelectionError` / `HandleForwardError` 的 outcome 并渲染路由错误；同账号重试、失败账号排除、绑定会话 429 保留、强制缓存计费、首次无账号判断和单账号 503 退避不应在 handler 中重新实现。

原生路由的运行时 session 准备由 `backend/internal/service/runtime_session.go` 统一处理。`PrepareRuntimeSession` 负责基于已校验 body、已解析 `ParsedRequest` 或 provider 显式传入的 `SessionHash` 补充 `SessionContext`、生成 session hash/key，并通过 `PrefetchRuntimeStickySession` 把已绑定 sticky 账号写入 request metadata context；handler 只传入客户端 IP、User-Agent、API key ID 和必要的 session key prefix。`/v1/messages`、`count_tokens`、Gemini native `/v1beta/models/*`、以及 OpenAI-compatible native Anthropic text flow 都应复用该入口。

`count_tokens` 不占用用户或账号并发槽位，但账号选择仍应通过 `backend/internal/service/runtime_count_tokens_selection.go`。该 helper 负责维护 failed-account exclusion set，并在 RPM admission 明确拒绝时继续选择下一个账号；RPM 预留异常仍按 fail-open 允许请求继续，handler 只记录 admission event 并渲染最终选路失败。选中账号后的 count_tokens 转发也应通过 `backend/internal/service/native_gateway_runtime.go` 的 `ForwardCountTokens`，handler 不应直接调用 `GatewayService.ForwardCountTokens`。

账号被选中后的槽位获取由 `backend/internal/service/runtime_account_slot_acquisition.go` 统一处理。Handler 只传入等待槽位时需要维持连接的 HTTP hook，并按 `queue_full`、`acquire_error`、`wait_acquire_error` 等 outcome 渲染协议错误；等待队列计数释放和等待成功后的粘性绑定不能再复制在各路由循环里。

转发前 admission control 由 `backend/internal/service/runtime_admission_control.go` 统一处理。RPM 与 window-cost reserve 失败仍按 fail-open 继续请求，但 reserve 明确拒绝时，service 会返回 `rpm_denied` 或 `window_cost_denied` 并清理对应粘性绑定；handler 只记录日志并进入下一轮选号。拒绝后的账号 slot、用户消息队列、`OnUpstreamAccepted` 回调和失败账号标记由 `backend/internal/service/runtime_admission_cleanup.go` 收口，避免各路由循环复制清理顺序。

实际转发尝试的资源收口由 `backend/internal/service/runtime_forward_attempt.go` 统一处理。Handler 仍负责协议错误映射和 usage 记录，但账号 slot 释放、用户消息队列兜底释放、`OnUpstreamAccepted` 清理、以及“上游错误且响应尚未开始写入”时释放 window-cost reservation 都由 service kernel 决定。Provider `Forward*` 分发由 `backend/internal/service/native_gateway_runtime.go` 统一处理；handler 只传入 provider/protocol/account request envelope，不应再内联选择 `GatewayService`、`GeminiMessagesCompatService` 或 `AntigravityGatewayService` 的具体 forward 方法。Forward error 的 failover 判定由 `RuntimeFailoverState.HandleForwardError` 统一处理；对于已经写入响应的 failover 错误，handler 应消费 `ResponseStarted` 决策结果并直接耗尽当前请求，避免跨账号重试拼接流式响应。

OpenAI-compatible native Anthropic `/v1/chat/completions` 与 `/v1/responses` 共用 `backend/internal/handler/gateway_handler_openai_compatible_text_flow.go`。两个路由仍保留各自的请求校验、Claude Code 限制、用户并发/计费检查和错误 JSON schema，但会把 session hash、账号选择、admission retry、native runtime 转发、failover 回调和 usage 记录交给同一个 flow，避免两条入口复制调度与资源清理循环。

## OpenAI Capability Index

OpenAI 调度快照的模型能力索引只把非空 `extra.supported_models` 视为显式能力声明：

- `supported_models` 缺失、空数组或空字符串：能力未知，账号进入 `model_any`，不能因为没有声明某个新模型就被索引路径过滤掉。
- `supported_models` 非空：按条目写入 `model_exact` / `model_pattern`，只匹配声明过的模型或通配符。
- `credentials.model_mapping` 是模型映射/别名规则，不再作为 OpenAI capability index 的模型能力声明来源。

这样可以兼容历史 OpenAI OAuth 账号：旧账号没有写 `supported_models` 时，更新后仍能参与新模型选路；只有明确声明了能力列表的账号才会按列表收窄。

## Test Checklist

涉及路由、回退或模型改写时，至少覆盖：

- `strict` 下高优先级健康账号不会被低优先级低负载账号抢走。
- 高优先级被排除、限流、临时不可调度或处于当前 transport cooling 时，才进入下一优先级。
- `weighted` 仍保留旧混合评分行为。
- `enable_model_fallback=false` 时不使用 `group.default_mapped_model`。
- `enable_model_fallback=true` 时才允许 default mapped model 作为选择回退。
- channel mapping 在模型回退关闭时仍正常生效。
- OpenAI `supported_models` 为空时仍进入 `model_any` 索引，不应触发 `no available OpenAI accounts supporting model`。
- OpenAI `supported_models` 非空时才按 exact/pattern capability 收窄候选。
