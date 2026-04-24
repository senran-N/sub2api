# Docs Guide

`docs/` 保存需要跟随代码一起维护的工程文档，不是历史存档区。

如果代码行为、接口、配置、脚本、架构边界或开发流程已经变化，那么对应文档必须在同一次工作中同步更新。

## 阅读顺序

1. [AGENT.md](/home/senran/Desktop/sub2api/AGENT.md)
2. [DEVELOPMENT_GUIDELINES.md](/home/senran/Desktop/sub2api/docs/DEVELOPMENT_GUIDELINES.md)
3. 如果涉及分支管理、fork 同步或上游移植，阅读 [GIT_WORKFLOW.md](/home/senran/Desktop/sub2api/docs/GIT_WORKFLOW.md)
4. 再阅读与你当前改动直接相关的专项文档

## 文档地图

| 文档 | 作用 | 什么时候必须更新 |
| --- | --- | --- |
| [docs/DEVELOPMENT_GUIDELINES.md](/home/senran/Desktop/sub2api/docs/DEVELOPMENT_GUIDELINES.md) | 仓库级开发规范、验证方式、文档维护规则 | 变更开发流程、测试入口、代码生成方式、文档维护规则时 |
| [docs/GIT_WORKFLOW.md](/home/senran/Desktop/sub2api/docs/GIT_WORKFLOW.md) | fork 分支模型、上游同步、选择性移植和冲突处理流程 | 变更分支策略、远端角色、同步流程、移植规则时 |
| [docs/ARCHITECTURE_EXTENSIBILITY.md](/home/senran/Desktop/sub2api/docs/ARCHITECTURE_EXTENSIBILITY.md) | 复用型扩展点和当前推荐架构边界 | 新增/替换共享扩展 seam、抽象层归属变更时 |
| [docs/OPENAI_DEFAULT_MODEL_CATALOG.md](/home/senran/Desktop/sub2api/docs/OPENAI_DEFAULT_MODEL_CATALOG.md) | OpenAI 内建默认模型目录及其后台/兼容网关影响范围 | 修改 `backend/internal/pkg/openai/constants.go`、默认模型暴露范围或相关回归测试时 |
| [docs/FRONTEND_TOKENIZATION_GUIDE.md](/home/senran/Desktop/sub2api/docs/FRONTEND_TOKENIZATION_GUIDE.md) | 前端 Token / Primitive / 样式收口规范 | 改主题 Token、共享样式入口、视觉复用规则时 |
| [docs/GROK_BACKEND_CONTROL_PLANE.md](/home/senran/Desktop/sub2api/docs/GROK_BACKEND_CONTROL_PLANE.md) | Grok provider 专属控制面与后台契约 | Grok 账户、媒体、运行时设置、后台展示契约变化时 |
| [docs/ADMIN_PAYMENT_INTEGRATION_API.md](/home/senran/Desktop/sub2api/docs/ADMIN_PAYMENT_INTEGRATION_API.md) | 外部支付/充值系统对接 Sub2API Admin API 的约束 | Admin 支付接口、幂等行为、嵌入页 query 参数变化时 |
| [docs/OPENAI_SCHEDULER_CAPACITY_TEST.md](/home/senran/Desktop/sub2api/docs/OPENAI_SCHEDULER_CAPACITY_TEST.md) | OpenAI 调度容量压测脚本和指标解释 | 压测脚本参数、实时监控指标、调度计数口径变化时 |
| [docs/EXTREME_SIMULATION_TESTING.md](/home/senran/Desktop/sub2api/docs/EXTREME_SIMULATION_TESTING.md) | 本地/预发极端环境矩阵、HTTP 压测脚本和 DevTools 退化检查 | 新增或修改极端测试脚本、profile、浏览器退化验证流程时 |
| [docs/CHANNEL_MONITOR_AND_AVAILABLE_CHANNELS.md](/home/senran/Desktop/sub2api/docs/CHANNEL_MONITOR_AND_AVAILABLE_CHANNELS.md) | 渠道监控、可用渠道、分组 RPM 覆盖和表格偏好契约 | 相关接口、设置键、前端入口或聚合口径变化时 |

## 目录边界

- 面向用户的项目介绍、功能总览、部署入口在根目录 `README*.md`。
- 部署细节、Docker/systemd 配置和运维说明在 [deploy/README.md](/home/senran/Desktop/sub2api/deploy/README.md)。
- `docs/` 只放工程实践、架构、集成和维护约束。

## 维护规则

- 改代码前，先确认是否已有对应文档，避免重新发明旧规则。
- 改代码时，优先让文档描述“当前真实实现”，不要保留过期示例。
- 改完接口、脚本、配置键、文件路径、命令后，必须逐条核对文档中的示例是否仍能对应代码。
- 如果改动没有现成文档承接，就补新文档，或者先在本目录登记入口。
- 不要把“后面再补文档”留成隐性债务；文档同步是交付的一部分。
