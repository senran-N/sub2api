# OpenAI Default Model Catalog

本文档说明 `backend/internal/pkg/openai/constants.go` 中内建 OpenAI 默认模型目录的作用范围，以及修改它时必须同步检查的链路。

## 作用范围

内建 OpenAI 默认模型目录不是装饰性常量，它会直接影响以下行为：

- 管理后台账户可用模型接口 `GET /api/v1/admin/accounts/:id/models` 在 OpenAI passthrough 或无显式 `model_mapping` 时返回的模型列表。
- 兼容网关 OpenAI 平台默认模型视图，例如 `DefaultCompatibleGatewayModels()` 和 `LookupCompatibleGatewayDefaultModel()`。
- `DefaultModelIDs()` 导出的 OpenAI 默认模型 ID 集。

因此，新增或移除默认模型时，必须按“真实对外契约”对待，而不是只改一处常量。

## 当前内建模型

当前 OpenAI 默认模型目录包含：

- `gpt-5.5`
- `gpt-5.4`
- `gpt-5.4-mini`
- `gpt-5.3-codex`
- `gpt-5.3-codex-spark`
- `gpt-5.2`

`DefaultTestModel` 仍保持为 `gpt-5.4`。这表示“后台未显式指定测试模型时的默认选择”仍然稳定，不等于系统不支持 `gpt-5.5`。

## 修改清单

当你修改内建 OpenAI 默认模型目录时，至少同步检查下面几项：

1. `backend/internal/pkg/openai/constants.go`
2. `backend/internal/pkg/openai/constants_test.go`
3. `backend/internal/handler/admin/account_handler.go`
4. `backend/internal/handler/admin/account_handler_available_models_test.go`
5. `backend/internal/handler/gateway_handler_models_test.go`
6. `backend/internal/service/compatible_gateway_model_view.go`

如果是新增 OpenAI 模型族，而不是单纯调整展示列表，还要继续检查：

- `backend/internal/service/openai_codex_transform.go` 的模型归一化
- `backend/internal/service/openai_model_mapping.go` 的映射链路
- `backend/internal/service/billing_service.go` 与 `backend/internal/service/pricing_service_lookup.go` 的计费兜底

## 提交前验证

涉及该目录的改动，至少运行受影响包测试，并确认以下两类行为没有回退：

- 默认模型列表确实暴露了目标模型。
- OpenAI passthrough / 无映射账户的后台模型接口仍返回该模型。
- 兼容 `/v1/models` 的 OpenAI 默认响应仍返回该模型。
- 显式 `model_mapping` 生成的后台模型列表顺序稳定，不依赖 Go map 遍历顺序。
