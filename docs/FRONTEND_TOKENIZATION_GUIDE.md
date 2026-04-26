# Frontend Tokenization Guide

本文档定义当前前端的 Token 化设计系统约束，以及后续页面开发时必须遵守的规则。

目标不是“把页面改好看”，而是把视觉决策收口到统一的 Token、共享 Primitive 和语义类中，避免重新回到各页面各写各的样式。

## 1. 目标

前端样式体系必须满足以下目标：

- 视觉决策集中在共享层，而不是散落在页面组件里。
- 同一主题下，admin、auth、payment、home、user 页面视觉语言一致。
- 新页面优先复用 Token 和 Primitive，而不是再写一套新的视觉实现。
- 主题切换、明暗模式、品牌差异，应当由 Token 驱动，而不是组件内部分支驱动。

## 2. Source Of Truth

当前前端设计系统的真实来源如下。

### 2.1 Token 定义

文件：

- `frontend/src/themes/theme.css`

职责：

- 定义所有 `--theme-*` 变量。
- 统一维护基础值、语义值、组件 primitive 值。

约束：

- 保持 `--theme-*` 前缀，不做随意重命名。
- 新增 Token 时，优先补到现有层级中，不要临时塞到页面局部样式里。

建议层次：

- `foundation`
  - 字体、字号、字重、圆角、阴影、间距、动效、z-index、原始品牌色
- `semantic`
  - surface、text、border、focus、overlay、disabled、status、interactive
- `primitive`
  - button、input、select、dropdown、dialog、table、shell、auth、shared-card、payment

### 2.2 共享视觉实现

文件：

- `frontend/src/style.css`

职责：

- 作为共享皮肤实现层。
- 承载跨页面复用的视觉规则，例如：
  - `.btn`
  - `.input`
  - `.select`
  - `.dropdown`
  - `.modal-*`
  - `.theme-chip`
  - table / card / empty / shell / header / sidebar primitives

约束：

- 共享视觉配方必须优先落在这里，而不是分散到多个业务组件中重复实现。
- 如果某个样式模式已经在多个页面重复出现，应优先抽到这里。

### 2.3 运行时 Token 读取

文件：

- `frontend/src/utils/themeStyles.ts`

职责：

- 提供 JS/TS 侧读取 Token 的统一接口。
- 主要用于图表、tooltip、alpha palette、line/doughnut config 等运行时样式场景。

约束：

- 不允许在图表或运行时配置里手写硬编码颜色字符串。
- 新增图表主题能力时，优先扩展这里，而不是在业务组件中直接拼接颜色。

## 3. 共享 Primitive 约束

以下共享组件必须作为结构层或统一视觉入口使用，不允许各自维护分裂的视觉配方：

- `frontend/src/components/common/Input.vue`
- `frontend/src/components/common/TextArea.vue`
- `frontend/src/components/common/Select.vue`
- `frontend/src/components/common/BaseDialog.vue`
- `frontend/src/components/common/DataTable.vue`
- `frontend/src/components/common/EmptyState.vue`

原则：

- 共享组件保留结构和状态表达。
- focus / error / disabled / hover / border / radius / shadow 等视觉细节优先由共享 token + shared css 决定。
- 单个业务组件不应重新定义一套自己的输入框、按钮、对话框视觉体系。

## 4. 页面层开发规则

### 4.1 必须做的事

- 优先复用现有 Token。
- 优先复用共享 Primitive。
- 业务页面中优先写语义 class，而不是直接堆模板 utility。
- 有复用价值的视觉模式，抽成局部语义类或共享 primitive。
- 品牌差异通过 Token 表达。
- 明暗模式通过 Token 表达。

### 4.2 明确禁止

- 硬编码视觉值：
  - 颜色
  - 阴影
  - 边框色
  - 圆角
  - 字体视觉参数
- 组件内 `data-brand-theme` / `.dark[data-brand-theme=...]` 视觉分支。
- 为了“修一下页面”直接写大量局部视觉样式。
- 用业务组件自己再实现一套 `.btn` / `.input` / `.card`。
- 在 JS 里直接写 `#xxxxxx` 或 `rgba(...)` 当主题色。

### 4.3 允许保留的 inline style

只允许动态几何类样式：

- 宽度 / 高度
- 位置 / 偏移
- 百分比尺寸
- canvas / chart 必需的运行时尺寸

不允许保留的 inline style：

- 颜色
- 阴影
- 字体
- 背景
- 边框
- 顺序类视觉属性（例如可用 class 替代的 `order`）

## 5. 页面开发流程

开发新页面或重构旧页面时，按下面顺序执行。

1. 先检查现有 Token 是否已经覆盖需求。
2. 如果没有，先补 `theme.css` 中的 Token。
3. 如果是通用视觉模式，再补 `style.css` 或 shared primitive。
4. 最后业务页面只消费这些能力。

判断标准：

- 只在一个页面出现、且无复用价值的结构样式，可以留在页面内。
- 只要出现跨页面复用倾向，就应该提升为 Token / Primitive / 语义类。

## 6. 各文件职责边界

### 6.1 `theme.css`

负责：

- 定义值
- 定义主题差异
- 定义明暗模式差异

不负责：

- 写具体业务页面布局

### 6.2 `style.css`

负责：

- 落地共享 primitive 的视觉实现
- 定义全局语义类

不负责：

- 业务页面独有的复杂布局逻辑

### 6.3 feature 页面 / 组件

负责：

- 业务结构
- 业务状态
- 组合已有 primitive

不负责：

- 发明新的主题系统
- 重复实现共享视觉配方

### 6.4 支付域样式

文件：

- `frontend/src/components/payment/paymentTheme.css`

职责：

- 支付域内部的语义层样式组织。
- 仍然必须消费共享 Token，不允许脱离设计系统单独造色板。

适用场景：

- 支付卡片
- 支付状态面板
- QR shell
- Stripe 内嵌支付容器

## 7. 推荐做法

### 7.1 推荐

```vue
<button class="btn btn-primary payment-submit-button">
  {{ t('payment.createOrder') }}
</button>
```

```css
.payment-submit-button {
  width: 100%;
}
```

说明：

- 共享按钮视觉来自 `.btn` / token。
- 页面只保留必要结构语义。

### 7.2 不推荐

```vue
<button
  style="background:#e85d3a;border-radius:14px;color:white;box-shadow:0 8px 24px rgba(0,0,0,.16)"
>
  Pay
</button>
```

问题：

- 视觉值硬编码。
- 绕过 Token。
- 无法随主题统一切换。

## 8. 品牌与明暗模式规范

品牌切换和明暗模式切换必须满足：

- 品牌差异只来自 Token。
- 组件模板不直接判断品牌来切换颜色。
- 明暗模式不在业务组件里重复写一套视觉实现。
- 首页、认证页、支付页、后台页必须共享同一套 Token 语言。

如果确实需要品牌特有装饰：

- 仍应通过 `theme.css` 中的品牌变量暴露。
- 页面层只消费语义变量，例如：
  - `--theme-home-glow`
  - `--theme-auth-accent`
  - `--theme-payment-brand-surface`

## 9. Code Review 清单

提交前至少自查以下问题：

- 是否新增了硬编码颜色或阴影？
- 是否新增了视觉型 inline style？
- 是否在组件里写了品牌 / dark 视觉分支？
- 是否重复实现了共享按钮、输入框、对话框样式？
- 是否本可以抽成 Token，却直接写死在页面里？
- 图表 / tooltip / runtime 配色是否走了 `themeStyles.ts`？
- 同类页面之间视觉是否一致？

只要以上任一项答案是“是”，就应该继续收口，而不是直接提交。

## 10. 最小验收标准

前端改动至少满足以下标准才可以认为合格：

- `pnpm exec vue-tsc --noEmit --pretty false` 通过。
- 不引入新的控制台样式相关错误。
- 目标页面在桌面和移动端下都可用。
- 明暗模式切换后页面仍保持统一视觉。
- 品牌切换后页面不出现脱离 Token 的颜色断层。

## 11. 一句话原则

页面负责业务，Token 负责视觉，Primitive 负责复用。

任何绕开这三层边界的写法，长期都会重新把前端带回不可维护状态。
