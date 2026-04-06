import type { OnboardingGuideContent } from './types'

const onboardingGuideContent: OnboardingGuideContent = {
  admin: {
    welcome: {
      title: '👋 欢迎使用 Sub2API',
      description:
        '<div style="line-height: 1.8;"><p style="margin-bottom: 16px;">Sub2API 是一个强大的 AI 服务中转平台，让您轻松管理和分发 AI 服务。</p><p style="margin-bottom: 12px;"><b>🎯 核心功能：</b></p><ul style="margin-left: 20px; margin-bottom: 16px;"><li>📦 <b>分组管理</b> - 创建不同的服务套餐（VIP、免费试用等）</li><li>🔗 <b>账号池</b> - 连接多个上游 AI 服务商账号</li><li>🔑 <b>密钥分发</b> - 为用户生成独立的 API Key</li><li>💰 <b>计费管理</b> - 灵活的费率和配额控制</li></ul><p style="color: #10b981; font-weight: 600;">接下来，我们将用 3 分钟带您完成首次配置 →</p></div>',
      nextBtn: '开始配置 🚀',
      prevBtn: '跳过'
    },
    groupManage: {
      title: '📦 第一步：分组管理',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;"><b>什么是分组？</b></p><p style="margin-bottom: 12px;">分组是 Sub2API 的核心概念，它就像一个"服务套餐"：</p><ul style="margin-left: 20px; margin-bottom: 12px; font-size: 13px;"><li>🎯 每个分组可以包含多个上游账号</li><li>💰 每个分组有独立的计费倍率</li><li>👥 可以设置为公开或专属分组</li></ul><p style="margin-top: 12px; padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 示例：</b>您可以创建"VIP专线"（高倍率）和"免费试用"（低倍率）两个分组</p><p style="margin-top: 16px; color: #10b981; font-weight: 600;">👉 点击左侧的"分组管理"开始</p></div>'
    },
    createGroup: {
      title: '➕ 创建新分组',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">现在让我们创建第一个分组。</p><p style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>📝 提示：</b>建议先创建一个测试分组，熟悉流程后再创建正式分组</p><p style="color: #10b981; font-weight: 600;">👉 点击"创建分组"按钮</p></div>'
    },
    groupName: {
      title: '✏️ 1. 分组名称',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">为您的分组起一个易于识别的名称。</p><div style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>💡 命名建议：</b><ul style="margin: 8px 0 0 16px;"><li>"测试分组" - 用于测试</li><li>"VIP专线" - 高质量服务</li><li>"免费试用" - 体验版</li></ul></div><p style="font-size: 13px; color: #6b7280;">填写完成后点击"下一步"继续</p></div>',
      nextBtn: '下一步'
    },
    groupPlatform: {
      title: '🤖 2. 选择平台',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">选择该分组支持的 AI 平台。</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>📌 平台说明：</b><ul style="margin: 8px 0 0 16px;"><li><b>Anthropic</b> - Claude 系列模型</li><li><b>OpenAI</b> - GPT 系列模型</li><li><b>Google</b> - Gemini 系列模型</li></ul></div><p style="font-size: 13px; color: #6b7280;">一个分组只能选择一个平台</p></div>',
      nextBtn: '下一步'
    },
    groupMultiplier: {
      title: '💰 3. 费率倍数',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">设置该分组的计费倍率，控制用户的实际扣费。</p><div style="padding: 8px 12px; background: #fef3c7; border-left: 3px solid #f59e0b; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>⚙️ 计费规则：</b><ul style="margin: 8px 0 0 16px;"><li><b>1.0</b> - 原价计费（成本价）</li><li><b>1.5</b> - 用户消耗 $1，扣除 $1.5</li><li><b>2.0</b> - 用户消耗 $1，扣除 $2</li><li><b>0.8</b> - 补贴模式（亏本运营）</li></ul></div><p style="font-size: 13px; color: #6b7280;">建议测试分组设置为 1.0</p></div>',
      nextBtn: '下一步'
    },
    groupExclusive: {
      title: '🔒 4. 专属分组（可选）',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">控制分组的可见性和访问权限。</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>🔐 权限说明：</b><ul style="margin: 8px 0 0 16px;"><li><b>关闭</b> - 公开分组，所有用户可见</li><li><b>开启</b> - 专属分组，仅指定用户可见</li></ul></div><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 使用场景：</b>VIP 用户专属、内部测试、特殊客户等</p></div>',
      nextBtn: '下一步'
    },
    groupSubmit: {
      title: '✅ 保存分组',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">确认信息无误后，点击创建按钮保存分组。</p><p style="padding: 8px 12px; background: #fef3c7; border-left: 3px solid #f59e0b; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>⚠️ 注意：</b>分组创建后，平台类型不可修改，其他信息可以随时编辑</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>📌 下一步：</b>创建成功后，我们将添加上游账号到这个分组</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 点击"创建"按钮</p></div>'
    },
    accountManage: {
      title: '🔗 第二步：添加账号',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;"><b>太棒了！分组已创建成功 🎉</b></p><p style="margin-bottom: 12px;">现在需要添加上游 AI 服务商的账号，让分组能够实际提供服务。</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>🔑 账号的作用：</b><ul style="margin: 8px 0 0 16px;"><li>连接到上游 AI 服务（Claude、GPT 等）</li><li>一个分组可以包含多个账号（负载均衡）</li><li>支持 OAuth 和 Session Key 两种方式</li></ul></div><p style="margin-top: 16px; color: #10b981; font-weight: 600;">👉 点击左侧的"账号管理"</p></div>'
    },
    createAccount: {
      title: '➕ 添加新账号',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">点击按钮开始添加您的第一个上游账号。</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 提示：</b>建议使用 OAuth 方式，更安全且无需手动提取密钥</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 点击"添加账号"按钮</p></div>'
    },
    accountName: {
      title: '✏️ 1. 账号名称',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">为账号设置一个便于识别的名称。</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 命名建议：</b>"Claude主账号"、"GPT备用1"、"测试账号" 等</p></div>',
      nextBtn: '下一步'
    },
    accountPlatform: {
      title: '🤖 2. 选择平台',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">选择该账号对应的服务提供方平台。</p><p style="padding: 8px 12px; background: #fef3c7; border-left: 3px solid #f59e0b; border-radius: 4px; font-size: 13px;"><b>⚠️ 重要：</b>平台必须与刚刚创建的分组一致</p></div>',
      nextBtn: '下一步'
    },
    accountType: {
      title: '🔐 3. 认证方式',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">选择这个上游账号的认证方式。</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px;"><b>📌 认证选项：</b><ul style="margin: 8px 0 0 16px;"><li><b>OAuth</b> - 推荐，安全且自动化</li><li><b>Session Key</b> - 手动输入令牌</li></ul></div></div>',
      nextBtn: '下一步'
    },
    accountPriority: {
      title: '⚖️ 4. 优先级',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">设置账号优先级，控制路由顺序。</p><div style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 建议：</b>数字越小优先级越高，主账号建议先从 1 开始。</div></div>',
      nextBtn: '下一步'
    },
    accountGroups: {
      title: '🎯 5. 绑定分组',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">选择哪些分组可以使用这个账号。</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px;"><b>📌 提示：</b>通常绑定到你前面创建的分组，这样该分组下的 API Key 才能使用这个账号。</div></div>',
      nextBtn: '下一步'
    },
    accountSubmit: {
      title: '✅ 保存账号',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">检查账号信息后点击保存。</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>📌 下一步：</b>账号可用后，就可以给用户创建 API Key 了。</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 点击"保存"按钮</p></div>'
    },
    keyManage: {
      title: '🔑 第三步：创建 API Key',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;"><b>马上就完成了！</b></p><p style="margin-bottom: 12px;">现在创建一个 API Key，让请求可以访问你配置好的分组和账号。</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px;"><b>📌 API Key 的用途：</b><ul style="margin: 8px 0 0 16px;"><li>客户端通过它调用网关</li><li>可以绑定到特定分组</li><li>支持配额和权限控制</li></ul></div><p style="margin-top: 16px; color: #10b981; font-weight: 600;">👉 点击「我的账户」中的「API 密钥」</p></div>'
    },
    createKey: {
      title: '➕ 创建 API Key',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">点击按钮创建新的 API Key。</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 提示：</b>这是开始发起请求前的最后一步。</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 点击"创建密钥"</p></div>'
    },
    keyName: {
      title: '✏️ 1. 密钥名称',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">给这个 API Key 起一个清晰的名字。</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 示例：</b>"正式环境主密钥"、"测试客户端"、"内部工具"</p></div>',
      nextBtn: '下一步'
    },
    keyGroup: {
      title: '🎯 2. 选择分组',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">选择您刚刚配置好的分组。</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>📌 分组决定：</b><ul style="margin: 8px 0 0 16px;"><li>这个密钥可以使用哪些账号</li><li>适用哪个计费倍率</li><li>是否是专属密钥</li></ul></div><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 提示：</b>选择你刚刚创建的测试分组</p></div>',
      nextBtn: '下一步'
    },
    keySubmit: {
      title: '🎉 生成并复制',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">点击创建后，系统会生成完整的 API Key。</p><div style="padding: 8px 12px; background: #fee2e2; border-left: 3px solid #ef4444; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>⚠️ 重要提醒：</b><ul style="margin: 8px 0 0 16px;"><li>密钥只显示一次，请立即复制</li><li>丢失后需要重新生成</li><li>妥善保管，不要泄露给他人</li></ul></div><div style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>🚀 下一步：</b><ul style="margin: 8px 0 0 16px;"><li>复制生成的 sk-xxx 密钥</li><li>在支持 OpenAI 接口的客户端中使用</li><li>开始体验 AI 服务！</li></ul></div><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 点击"创建"按钮</p></div>'
    }
  },
  user: {
    welcome: {
      title: '👋 欢迎使用 Sub2API',
      description:
        '<div style="line-height: 1.8;"><p style="margin-bottom: 16px;">您好！欢迎来到 Sub2API AI 服务平台。</p><p style="margin-bottom: 12px;"><b>🎯 快速开始：</b></p><ul style="margin-left: 20px; margin-bottom: 16px;"><li>🔑 创建 API 密钥</li><li>📋 复制密钥到您的应用</li><li>🚀 开始使用 AI 服务</li></ul><p style="color: #10b981; font-weight: 600;">只需 1 分钟，让我们开始吧 →</p></div>',
      nextBtn: '开始 🚀',
      prevBtn: '跳过'
    },
    keyManage: {
      title: '🔑 API 密钥管理',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">在这里管理您的所有 API 访问密钥。</p><p style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px;"><b>📌 什么是 API 密钥？</b><br/>API 密钥是您访问 AI 服务的凭证，就像一把钥匙，让您的应用能够调用 AI 能力。</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 点击进入密钥页面</p></div>'
    },
    createKey: {
      title: '➕ 创建新密钥',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">点击按钮创建您的第一个 API 密钥。</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 提示：</b>创建后密钥只显示一次，请务必复制保存</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 点击"创建密钥"</p></div>'
    },
    keyName: {
      title: '✏️ 密钥名称',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">为密钥起一个便于识别的名称。</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 示例：</b>"我的第一个密钥"、"测试用" 等</p></div>',
      nextBtn: '下一步'
    },
    keyGroup: {
      title: '🎯 选择分组',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">选择管理员为您分配的服务分组。</p><p style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px;"><b>📌 分组说明：</b><br/>不同分组可能有不同的服务质量和计费标准，请根据需要选择。</p></div>',
      nextBtn: '下一步'
    },
    keySubmit: {
      title: '🎉 完成创建',
      description:
        '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">点击确认创建您的 API 密钥。</p><div style="padding: 8px 12px; background: #fee2e2; border-left: 3px solid #ef4444; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>⚠️ 重要：</b><ul style="margin: 8px 0 0 16px;"><li>创建后请立即复制密钥（sk-xxx）</li><li>密钥只显示一次，丢失需重新生成</li></ul></div><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>🚀 如何使用：</b><br/>将密钥配置到支持 OpenAI 接口的任何客户端（如 ChatBox、OpenCat 等），即可开始使用！</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 点击"创建"按钮</p></div>'
    }
  }
}

export default onboardingGuideContent
