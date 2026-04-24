# Channel Monitor 与 Available Channels

本文记录本次从上游引入并保留到 fork 的渠道监控、可用渠道、RPM 覆盖和表格偏好相关契约。

## 后端接口

- 管理端渠道监控：`/api/v1/admin/channel-monitors`，支持列表、创建、更新、删除、手动检测和历史记录。
- 管理端监控模板：`/api/v1/admin/channel-monitor-templates`，用于保存并套用高级请求配置。
- 用户端渠道状态：`/api/v1/channel-monitors`，提供用户可见的监控状态与详情。
- 用户端可用渠道：`/api/v1/channels/available`，按渠道、平台、分组和模型定价聚合当前用户可访问能力。
- 分组 RPM 覆盖：`/api/v1/admin/groups/:id/rpm-overrides`，仅维护用户在指定分组下的 RPM 覆盖，不覆盖已有倍率设置。

## 配置与公开设置

- `channel_monitor_enabled`：控制渠道状态页和监控能力展示，默认开启。
- `channel_monitor_default_interval_seconds`：新建监控默认检测间隔，后端限制为 15–3600 秒。
- `available_channels_enabled`：控制用户端可用渠道入口，默认按设置返回。
- `table_default_page_size` / `table_page_size_options`：前端表格分页偏好，由系统设置统一下发，不再依赖各页面本地默认值。

## 前端入口

- 管理端页面：`frontend/src/views/admin/ChannelMonitorView.vue`。
- 用户端页面：`frontend/src/views/user/ChannelStatusView.vue` 与 `frontend/src/views/user/AvailableChannelsView.vue`。
- 共享常量：`frontend/src/constants/channel.ts` 与 `frontend/src/constants/channelMonitor.ts`。
- 表格偏好：`frontend/src/utils/tablePreferences.ts` 与 `frontend/src/composables/usePersistedPageSize.ts`。

## Fork 保留点

- 保留 fork 的 Grok、Antigravity、兼容网关、TLS 指纹、支付和 WeChat 双模式配置字段。
- 后端 service/handler 签名已按上游功能补齐排序与筛选参数，同时保持旧测试构造器兼容。
- Grok 账户模型列表继续使用 fork 的 Grok registry，不回退到 Anthropic/OpenAI 默认模型。
