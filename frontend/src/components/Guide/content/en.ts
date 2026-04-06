import type { OnboardingGuideContent } from './types'

const onboardingGuideContent: OnboardingGuideContent = {
  admin: {
    welcome: {
      title: '👋 Welcome to Sub2API',
      description: '<div style="line-height: 1.8;"><p style="margin-bottom: 16px;">Sub2API is a powerful AI service gateway platform that helps you easily manage and distribute AI services.</p><p style="margin-bottom: 12px;"><b>🎯 Core Features:</b></p><ul style="margin-left: 20px; margin-bottom: 16px;"><li>📦 <b>Group Management</b> - Create service tiers (VIP, Free Trial, etc.)</li><li>🔗 <b>Account Pool</b> - Connect multiple upstream AI service accounts</li><li>🔑 <b>Key Distribution</b> - Generate independent API Keys for users</li><li>💰 <b>Billing Control</b> - Flexible rate and quota management</li></ul><p style="color: #10b981; font-weight: 600;">Let\'s complete the initial setup in 3 minutes →</p></div>',
      nextBtn: 'Start Setup 🚀',
      prevBtn: 'Skip'
    },
    groupManage: {
      title: '📦 Step 1: Group Management',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;"><b>What is a Group?</b></p><p style="margin-bottom: 12px;">Groups are the core concept of Sub2API, like a "service package":</p><ul style="margin-left: 20px; margin-bottom: 12px; font-size: 13px;"><li>🎯 Each group can contain multiple upstream accounts</li><li>💰 Each group has independent billing multiplier</li><li>👥 Can be set as public or exclusive</li></ul><p style="margin-top: 12px; padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Example:</b> You can create "VIP Premium" (high rate) and "Free Trial" (low rate) groups</p><p style="margin-top: 16px; color: #10b981; font-weight: 600;">👉 Click "Group Management" on the left sidebar</p></div>'
    },
    createGroup: {
      title: '➕ Create New Group',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Let\'s create your first group.</p><p style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>📝 Tip:</b> Recommend creating a test group first to familiarize yourself with the process</p><p style="color: #10b981; font-weight: 600;">👉 Click the "Create Group" button</p></div>'
    },
    groupName: {
      title: '✏️ 1. Group Name',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Give your group an easy-to-identify name.</p><div style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>💡 Naming Suggestions:</b><ul style="margin: 8px 0 0 16px;"><li>"Test Group" - For testing</li><li>"VIP Premium" - High-quality service</li><li>"Free Trial" - Trial version</li></ul></div><p style="font-size: 13px; color: #6b7280;">Click "Next" when done</p></div>',
      nextBtn: 'Next'
    },
    groupPlatform: {
      title: '🤖 2. Select Platform',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Choose the AI platform this group supports.</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>📌 Platform Guide:</b><ul style="margin: 8px 0 0 16px;"><li><b>Anthropic</b> - Claude models</li><li><b>OpenAI</b> - GPT models</li><li><b>Google</b> - Gemini models</li></ul></div><p style="font-size: 13px; color: #6b7280;">One group can only have one platform</p></div>',
      nextBtn: 'Next'
    },
    groupMultiplier: {
      title: '💰 3. Rate Multiplier',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Set the billing multiplier to control user charges.</p><div style="padding: 8px 12px; background: #fef3c7; border-left: 3px solid #f59e0b; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>⚙️ Billing Rules:</b><ul style="margin: 8px 0 0 16px;"><li><b>1.0</b> - Original price (cost price)</li><li><b>1.5</b> - User consumes $1, charged $1.5</li><li><b>2.0</b> - User consumes $1, charged $2</li><li><b>0.8</b> - Subsidy mode (loss-making)</li></ul></div><p style="font-size: 13px; color: #6b7280;">Recommend setting test group to 1.0</p></div>',
      nextBtn: 'Next'
    },
    groupExclusive: {
      title: '🔒 4. Exclusive Group (Optional)',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Control group visibility and access permissions.</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>🔐 Permission Guide:</b><ul style="margin: 8px 0 0 16px;"><li><b>Off</b> - Public group, visible to all users</li><li><b>On</b> - Exclusive group, only for specified users</li></ul></div><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Use Cases:</b> VIP exclusive, internal testing, special customers</p></div>',
      nextBtn: 'Next'
    },
    groupSubmit: {
      title: '✅ Save Group',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Confirm the information and click create to save the group.</p><p style="padding: 8px 12px; background: #fef3c7; border-left: 3px solid #f59e0b; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>⚠️ Note:</b> Platform type cannot be changed after creation, but other settings can be edited anytime</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>📌 Next Step:</b> After creation, we\'ll add upstream accounts to this group</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 Click "Create" button</p></div>'
    },
    accountManage: {
      title: '🔗 Step 2: Add Account',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;"><b>Great! Group created successfully 🎉</b></p><p style="margin-bottom: 12px;">Now add upstream AI service accounts to enable actual service delivery.</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>🔑 Account Purpose:</b><ul style="margin: 8px 0 0 16px;"><li>Connect to upstream AI services (Claude, GPT, etc.)</li><li>One group can contain multiple accounts (load balancing)</li><li>Supports OAuth and Session Key methods</li></ul></div><p style="margin-top: 16px; color: #10b981; font-weight: 600;">👉 Click "Account Management" on the left sidebar</p></div>'
    },
    createAccount: {
      title: '➕ Add New Account',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Click the button to start adding your first upstream account.</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Tip:</b> Recommend using OAuth method - more secure and no manual key extraction needed</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 Click "Add Account" button</p></div>'
    },
    accountName: {
      title: '✏️ 1. Account Name',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Set an easy-to-identify name for the account.</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Naming Suggestions:</b> "Claude Main", "GPT Backup 1", "Test Account", etc.</p></div>',
      nextBtn: 'Next'
    },
    accountPlatform: {
      title: '🤖 2. Select Platform',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Choose the service provider platform for this account.</p><p style="padding: 8px 12px; background: #fef3c7; border-left: 3px solid #f59e0b; border-radius: 4px; font-size: 13px;"><b>⚠️ Important:</b> Platform must match the group you just created</p></div>',
      nextBtn: 'Next'
    },
    accountType: {
      title: '🔐 3. Authentication Method',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Choose how to authenticate this upstream account.</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px;"><b>📌 Authentication Options:</b><ul style="margin: 8px 0 0 16px;"><li><b>OAuth</b> - Recommended, secure and automatic</li><li><b>Session Key</b> - Manual token entry</li></ul></div></div>',
      nextBtn: 'Next'
    },
    accountPriority: {
      title: '⚖️ 4. Priority',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Set the account priority for routing order.</p><div style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Recommendation:</b> Lower numbers mean higher priority. Start with 1 for your main account.</div></div>',
      nextBtn: 'Next'
    },
    accountGroups: {
      title: '🎯 5. Bind Groups',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Select which groups can use this account.</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px;"><b>📌 Tip:</b> Usually bind it to the group you created earlier so the API keys in that group can use this account.</div></div>',
      nextBtn: 'Next'
    },
    accountSubmit: {
      title: '✅ Save Account',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Review the account information and click save.</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>📌 Next Step:</b> After the account is ready, you can create API keys for users.</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 Click "Save" button</p></div>'
    },
    keyManage: {
      title: '🔑 Step 3: Create API Key',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;"><b>Almost done!</b></p><p style="margin-bottom: 12px;">Now create an API key so requests can access the group and account you configured.</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px;"><b>📌 API Key Usage:</b><ul style="margin: 8px 0 0 16px;"><li>Used by clients to call the gateway</li><li>Can bind to a specific group</li><li>Supports quota and permission control</li></ul></div><p style="margin-top: 16px; color: #10b981; font-weight: 600;">👉 Click "API Keys" in My Account</p></div>'
    },
    createKey: {
      title: '➕ Create API Key',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Click the button to create a new API key.</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Tip:</b> This is the final step before you can start making requests.</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 Click "Create Key"</p></div>'
    },
    keyName: {
      title: '✏️ 1. Key Name',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Give this API key a clear name.</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Examples:</b> "Main Production Key", "Test Client", "Internal Tool"</p></div>',
      nextBtn: 'Next'
    },
    keyGroup: {
      title: '🎯 2. Select Group',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Select the group you just configured.</p><div style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>📌 Group Determines:</b><ul style="margin: 8px 0 0 16px;"><li>Which accounts this key can use</li><li>What billing multiplier applies</li><li>Whether it\'s an exclusive key</li></ul></div><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Tip:</b> Select the test group you just created</p></div>',
      nextBtn: 'Next'
    },
    keySubmit: {
      title: '🎉 Generate and Copy',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">System will generate a complete API Key after clicking create.</p><div style="padding: 8px 12px; background: #fee2e2; border-left: 3px solid #ef4444; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>⚠️ Important Reminder:</b><ul style="margin: 8px 0 0 16px;"><li>Key is only shown once, copy immediately</li><li>Need to regenerate if lost</li><li>Keep it safe, don\'t share with others</li></ul></div><div style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>🚀 Next Steps:</b><ul style="margin: 8px 0 0 16px;"><li>Copy the generated sk-xxx key</li><li>Use in any OpenAI-compatible client</li><li>Start experiencing AI services!</li></ul></div><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 Click "Create" button</p></div>'
    }
  },
  user: {
    welcome: {
      title: '👋 Welcome to Sub2API',
      description: '<div style="line-height: 1.8;"><p style="margin-bottom: 16px;">Hello! Welcome to the Sub2API AI service platform.</p><p style="margin-bottom: 12px;"><b>🎯 Quick Start:</b></p><ul style="margin-left: 20px; margin-bottom: 16px;"><li>🔑 Create API Key</li><li>📋 Copy key to your application</li><li>🚀 Start using AI services</li></ul><p style="color: #10b981; font-weight: 600;">Just 1 minute, let\'s get started →</p></div>',
      nextBtn: 'Start 🚀',
      prevBtn: 'Skip'
    },
    keyManage: {
      title: '🔑 API Key Management',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Manage all your API access keys here.</p><p style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px;"><b>📌 What is an API Key?</b><br/>An API key is your credential for accessing AI services, like a key that allows your application to call AI capabilities.</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 Click to enter key page</p></div>'
    },
    createKey: {
      title: '➕ Create New Key',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Click the button to create your first API key.</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Tip:</b> Key is only shown once after creation, make sure to copy and save</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 Click "Create Key"</p></div>'
    },
    keyName: {
      title: '✏️ Key Name',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Give your key an easy-to-identify name.</p><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>💡 Examples:</b> "My First Key", "For Testing", etc.</p></div>',
      nextBtn: 'Next'
    },
    keyGroup: {
      title: '🎯 Select Group',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Select the service group assigned by the administrator.</p><p style="padding: 8px 12px; background: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; font-size: 13px;"><b>📌 Group Info:</b><br/>Different groups may have different service quality and billing rates, choose according to your needs.</p></div>',
      nextBtn: 'Next'
    },
    keySubmit: {
      title: '🎉 Complete Creation',
      description: '<div style="line-height: 1.7;"><p style="margin-bottom: 12px;">Click to confirm and create your API key.</p><div style="padding: 8px 12px; background: #fee2e2; border-left: 3px solid #ef4444; border-radius: 4px; font-size: 13px; margin-bottom: 12px;"><b>⚠️ Important:</b><ul style="margin: 8px 0 0 16px;"><li>Copy the key (sk-xxx) immediately after creation</li><li>Key is only shown once, need to regenerate if lost</li></ul></div><p style="padding: 8px 12px; background: #f0fdf4; border-left: 3px solid #10b981; border-radius: 4px; font-size: 13px;"><b>🚀 How to Use:</b><br/>Configure the key in any OpenAI-compatible client (like ChatBox, OpenCat, etc.) and start using!</p><p style="margin-top: 12px; color: #10b981; font-weight: 600;">👉 Click "Create" button</p></div>'
    }
  }
}

export default onboardingGuideContent
