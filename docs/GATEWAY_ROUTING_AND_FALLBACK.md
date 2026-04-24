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
