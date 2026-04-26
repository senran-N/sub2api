# Gateway Runtime Core

Sub2API gateway runtime is split into thin HTTP adapters and service-owned kernels. Handlers own protocol parsing, auth context extraction, and protocol error rendering. Account choice, account slot handling, admission, provider forwarding, failover state, runtime feedback, usage hooks, and cleanup must stay in `backend/internal/service`.

## Runtime Pipeline

`backend/internal/service/runtime_pipeline.go` is the native provider request pipeline. It uses the shared primitives already owned by service:

1. session/sticky state from `runtime_session.go`
2. selection through `selection_kernel.go`
3. account-slot acquisition through `runtime_account_slot_acquisition.go`
4. admission via `runtime_admission_control.go`
5. denied-admission cleanup through `runtime_admission_cleanup.go`
6. provider forward dispatch through `native_gateway_runtime.go`
7. forward cleanup through `runtime_forward_attempt.go`
8. failover through `runtime_failover_state.go`

The pipeline exposes protocol-neutral outcomes. HTTP adapters translate those outcomes into OpenAI, Anthropic, or Google/Gemini error shapes.

## Route Matrix

| Inbound family | Runtime owner | Handler responsibility |
| --- | --- | --- |
| OpenAI-compatible `/v1/responses`, `/v1/chat/completions`, `/v1/messages` | `CompatibleTextExecutionKernel` plus `CompatibleGatewayTextRuntime` | Validate body, build route request, render OpenAI-compatible errors, submit usage hooks |
| OpenAI-compatible passthrough | `CompatiblePassthroughExecutionKernel` | Resolve passthrough metadata, render errors, submit usage hooks |
| OpenAI Responses WebSocket ingress | `openai_ws_ingress_selection_kernel.go` and WS forwarder runtime | WebSocket close-code mapping and long-lived proxy hook wiring |
| Anthropic native `/v1/messages` | `RuntimePipeline` and `NativeGatewayRuntime` | Claude body validation, Anthropic error rendering, Antigravity fallback policy |
| Anthropic-compatible native `/v1/chat/completions` and `/v1/responses` | `RuntimePipeline` and `NativeGatewayRuntime` | Route-specific request validation and error rendering |
| Gemini native `/v1beta/models/*:generateContent` | `RuntimePipeline` and `NativeGatewayRuntime` | Google error rendering and Gemini digest/session request adaptation |
| Native `count_tokens` | `runtime_count_tokens_selection.go` and `NativeGatewayRuntime.ForwardCountTokens` | Count-token request validation and final error rendering |
| Grok text/media/session | Grok-owned runtimes and account state services | Route dispatch and provider-owned response rendering only |

## Provider Boundaries

- OpenAI-compatible routing uses OpenAI scheduler state only inside `OpenAIGatewayService`, `CompatibleTextExecutionKernel`, and WS ingress selection.
- Native Anthropic/Gemini/Antigravity routing uses `SelectionKernel`, `RuntimePipeline`, and `NativeGatewayRuntime`; handlers must not call provider `Forward*` methods directly.
- Grok `extra.grok` writes must go through `GrokAccountStateService` or a Grok-owned patch builder. Shared compatible code must not shallow-merge Grok state.
- Gemini digest sticky fallback is request adaptation. Selection, slot handling, failover, and provider forward dispatch remain service-owned.

## Cleanup Rules

- A response-started failover exhausts the request; the runtime must not switch accounts after protocol bytes have been written.
- Account release, user-message queue release, `OnUpstreamAccepted` clearing, and window-cost release are centralized in service helpers.
- Admission denial must mark the current account failed before re-entering selection.
- `ForceCacheBilling` and account-switch counts are owned by failover state and returned to handlers for usage recording only.

## Settings And Lifecycle

Runtime code should depend on narrow setting readers from `setting_runtime_readers.go` instead of the full `SettingService` when it only needs gateway, Grok, websearch, auth, or ops settings.

Background services that start workers during dependency construction should be registered with `LifecycleRegistry` so application shutdown can stop them in one place.
