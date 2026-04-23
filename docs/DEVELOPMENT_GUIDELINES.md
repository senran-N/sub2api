# Development Guidelines

本文档定义当前仓库的开发规范、验证基线和文档同步要求。它是对 [AGENT.md](/home/senran/Desktop/sub2api/AGENT.md) 的工程化展开，面向所有会修改本仓库代码或文档的人。

## 1. 工作流程

每次改动默认按下面顺序执行：

1. 先读 [AGENT.md](/home/senran/Desktop/sub2api/AGENT.md)、本文件，以及与你改动直接相关的专项文档。
2. 先看真实实现，再决定改法，不要只根据旧文档或文件名猜测结构。
3. 优先在已有模块和扩展 seam 上演进，避免再造并行实现。
4. 改完后做最小但足够的验证。
5. 在提交前同步更新 `docs/`，确保接口、命令、路径、架构说明仍然匹配代码。

## 2. 仓库级约定

### 2.1 根目录入口

- 常用统一入口在根目录 [Makefile](/home/senran/Desktop/sub2api/Makefile)。
- 适合作为仓库级基线的命令：
  - `make build`
  - `make test`
  - `make lint`
  - `make migrate-validate`
  - `make secret-scan`

### 2.2 Backend 约定

- 后端模块在 `backend/`，当前 `go.mod` 声明 Go 版本为 `1.26.2`。
- 优先通过 [backend/Makefile](/home/senran/Desktop/sub2api/backend/Makefile) 运行常用动作：
  - `make -C backend build`
  - `make -C backend test`
  - `make -C backend lint`
  - `make -C backend coverage`
  - `make -C backend migrate-validate`
- 以下文件属于生成产物，不要手改：
  - `backend/ent/**`
  - `backend/cmd/server/wire_gen.go`
- 当你修改 Ent schema 或 Wire provider 关系时，先改源文件，再重新生成：
  - `make -C backend generate`

### 2.3 Frontend 约定

- 前端模块在 `frontend/`，当前包管理器为 `pnpm@9.15.9`。
- 常用命令：
  - `pnpm --dir frontend run dev`
  - `pnpm --dir frontend run build`
  - `pnpm --dir frontend run lint:check`
  - `pnpm --dir frontend run typecheck`
  - `pnpm --dir frontend run test:run`
- 涉及主题、共享样式、视觉 Primitive 的改动，必须同时遵守 [docs/FRONTEND_TOKENIZATION_GUIDE.md](/home/senran/Desktop/sub2api/docs/FRONTEND_TOKENIZATION_GUIDE.md)。

## 3. 文档同步规则

文档同步不是可选项。以下改动必须同时更新 `docs/`：

| 改动类型 | 必须检查/更新的文档 |
| --- | --- |
| 开发流程、验证命令、代码生成方式、仓库协作规则 | [docs/DEVELOPMENT_GUIDELINES.md](/home/senran/Desktop/sub2api/docs/DEVELOPMENT_GUIDELINES.md)、[docs/README.md](/home/senran/Desktop/sub2api/docs/README.md) |
| 共享架构边界、兼容网关扩展 seam、后台可复用抽象 | [docs/ARCHITECTURE_EXTENSIBILITY.md](/home/senran/Desktop/sub2api/docs/ARCHITECTURE_EXTENSIBILITY.md) |
| 前端 Token、主题、共享样式入口、视觉复用规则 | [docs/FRONTEND_TOKENIZATION_GUIDE.md](/home/senran/Desktop/sub2api/docs/FRONTEND_TOKENIZATION_GUIDE.md) |
| Grok 控制面、账户状态、后台展示契约、运行时设置 | [docs/GROK_BACKEND_CONTROL_PLANE.md](/home/senran/Desktop/sub2api/docs/GROK_BACKEND_CONTROL_PLANE.md) |
| 外部支付接入、管理员充值接口、嵌入页参数 | [docs/ADMIN_PAYMENT_INTEGRATION_API.md](/home/senran/Desktop/sub2api/docs/ADMIN_PAYMENT_INTEGRATION_API.md) |
| OpenAI 调度压测脚本、指标口径、观测链路 | [docs/OPENAI_SCHEDULER_CAPACITY_TEST.md](/home/senran/Desktop/sub2api/docs/OPENAI_SCHEDULER_CAPACITY_TEST.md) |

如果现有文档都不合适，就新增文档，并在 [docs/README.md](/home/senran/Desktop/sub2api/docs/README.md) 登记入口。

## 4. 最小验证基线

按改动范围至少满足下面的一组或多组检查：

### 4.1 Backend 改动

- 优先运行受影响包的 `go test`。
- 如果是较完整的后端改动，运行：
  - `make -C backend test`

### 4.2 Frontend 改动

- 至少运行：
  - `pnpm --dir frontend run lint:check`
  - `pnpm --dir frontend run typecheck`
- 如果改动包含状态逻辑、表单逻辑或组件行为，补跑受影响 `vitest` 用例。

### 4.3 跨层改动

- 涉及前后端联动、设置项、接口契约时，优先运行：
  - `make test`
- 如果因为环境或耗时无法完整执行，至少运行受影响层的验证命令，并在变更说明中明确写出未执行项。

## 5. 提交前检查

提交前至少自查以下问题：

- 是否改动了实际行为，却没有同步更新对应文档？
- 文档里的命令、文件路径、接口路径、配置键是否仍然存在？
- 是否误改了生成文件，而没有更新它的源文件？
- 是否在已有扩展 seam 之外又新增了一条平行逻辑？
- 是否运行了与改动范围相匹配的验证？

如果以上任一项答案不成立，这次改动就还没有完成。
