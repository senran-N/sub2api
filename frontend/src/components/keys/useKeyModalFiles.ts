import type { GroupPlatform } from '@/types'

export interface UseKeyModalFileConfig {
  path: string
  content: string
  hint?: string
  highlighted?: string
}

interface BuildUseKeyModalFilesOptions {
  activeClientTab: string
  activeTab: string
  apiKey: string
  baseUrl: string
  platform: GroupPlatform | null
  t: (key: string) => string
}

function escapeHtml(value: string) {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function wrapToken(className: string, value: string) {
  return `<span class="${className}">${escapeHtml(value)}</span>`
}

const keyword = (value: string) => wrapToken('use-key-modal__syntax-keyword', value)
const variable = (value: string) => wrapToken('use-key-modal__syntax-variable', value)
const operator = (value: string) => wrapToken('use-key-modal__syntax-operator', value)
const string = (value: string) => wrapToken('use-key-modal__syntax-string', value)
const comment = (value: string) => wrapToken('use-key-modal__syntax-comment', value)

function buildAnthropicFiles(
  activeTab: string,
  baseUrl: string,
  apiKey: string
): UseKeyModalFileConfig[] {
  let path = 'Terminal'
  let content = ''

  switch (activeTab) {
    case 'unix':
      content = `export ANTHROPIC_BASE_URL="${baseUrl}"
export ANTHROPIC_AUTH_TOKEN="${apiKey}"
export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`
      break
    case 'cmd':
      path = 'Command Prompt'
      content = `set ANTHROPIC_BASE_URL=${baseUrl}
set ANTHROPIC_AUTH_TOKEN=${apiKey}
set CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`
      break
    case 'powershell':
      path = 'PowerShell'
      content = `$env:ANTHROPIC_BASE_URL="${baseUrl}"
$env:ANTHROPIC_AUTH_TOKEN="${apiKey}"
$env:CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`
      break
  }

  const vscodeSettingsPath =
    activeTab === 'unix' ? '~/.claude/settings.json' : '%userprofile%\\.claude\\settings.json'

  const vscodeContent = `{
  "env": {
    "ANTHROPIC_BASE_URL": "${baseUrl}",
    "ANTHROPIC_AUTH_TOKEN": "${apiKey}",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
    "CLAUDE_CODE_ATTRIBUTION_HEADER": "0"
  }
}`

  return [
    { path, content },
    { path: vscodeSettingsPath, content: vscodeContent, hint: 'VSCode Claude Code' }
  ]
}

function buildGeminiCliFile(
  activeTab: string,
  baseUrl: string,
  apiKey: string,
  t: (key: string) => string
): UseKeyModalFileConfig {
  const model = 'gemini-2.0-flash'
  const modelComment = t('keys.useKeyModal.gemini.modelComment')
  let path = 'Terminal'
  let content = ''
  let highlighted = ''

  switch (activeTab) {
    case 'unix':
      content = `export GOOGLE_GEMINI_BASE_URL="${baseUrl}"
export GEMINI_API_KEY="${apiKey}"
export GEMINI_MODEL="${model}"  # ${modelComment}`
      highlighted = `${keyword('export')} ${variable('GOOGLE_GEMINI_BASE_URL')}${operator('=')}${string(`"${baseUrl}"`)}
${keyword('export')} ${variable('GEMINI_API_KEY')}${operator('=')}${string(`"${apiKey}"`)}
${keyword('export')} ${variable('GEMINI_MODEL')}${operator('=')}${string(`"${model}"`)}  ${comment(`# ${modelComment}`)}`
      break
    case 'cmd':
      path = 'Command Prompt'
      content = `set GOOGLE_GEMINI_BASE_URL=${baseUrl}
set GEMINI_API_KEY=${apiKey}
set GEMINI_MODEL=${model}`
      highlighted = `${keyword('set')} ${variable('GOOGLE_GEMINI_BASE_URL')}${operator('=')}${string(baseUrl)}
${keyword('set')} ${variable('GEMINI_API_KEY')}${operator('=')}${string(apiKey)}
${keyword('set')} ${variable('GEMINI_MODEL')}${operator('=')}${string(model)}
${comment(`REM ${modelComment}`)}`
      break
    case 'powershell':
      path = 'PowerShell'
      content = `$env:GOOGLE_GEMINI_BASE_URL="${baseUrl}"
$env:GEMINI_API_KEY="${apiKey}"
$env:GEMINI_MODEL="${model}"  # ${modelComment}`
      highlighted = `${keyword('$env:')}${variable('GOOGLE_GEMINI_BASE_URL')}${operator('=')}${string(`"${baseUrl}"`)}
${keyword('$env:')}${variable('GEMINI_API_KEY')}${operator('=')}${string(`"${apiKey}"`)}
${keyword('$env:')}${variable('GEMINI_MODEL')}${operator('=')}${string(`"${model}"`)}  ${comment(`# ${modelComment}`)}`
      break
  }

  return { path, content, highlighted }
}

function buildOpenAIFiles(activeTab: string, baseUrl: string, apiKey: string, webSocket: boolean) {
  const isWindows = activeTab === 'windows'
  const configDir = isWindows ? '%userprofile%\\.codex' : '~/.codex'
  const configContent = `model_provider = "OpenAI"
model = "gpt-5.4"
review_model = "gpt-5.4"
model_reasoning_effort = "xhigh"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true
model_context_window = 1000000
model_auto_compact_token_limit = 900000

[model_providers.OpenAI]
name = "OpenAI"
base_url = "${baseUrl}"
wire_api = "responses"${webSocket ? '\nsupports_websockets = true' : ''}
requires_openai_auth = true${webSocket ? '\n\n[features]\nresponses_websockets_v2 = true' : ''}`

  const authContent = `{
  "OPENAI_API_KEY": "${apiKey}"
}`

  return [
    {
      path: `${configDir}/config.toml`,
      content: configContent
    },
    {
      path: `${configDir}/auth.json`,
      content: authContent
    }
  ]
}

type OpenCodeModelMap = Record<string, Record<string, unknown>>

function buildOpenAIModels(): OpenCodeModelMap {
  return {
    'gpt-5.2': {
      name: 'GPT-5.2',
      limit: { context: 400000, output: 128000 },
      options: { store: false },
      variants: { low: {}, medium: {}, high: {}, xhigh: {} }
    },
    'gpt-5.4': {
      name: 'GPT-5.4',
      limit: { context: 1050000, output: 128000 },
      options: { store: false },
      variants: { low: {}, medium: {}, high: {}, xhigh: {} }
    },
    'gpt-5.4-mini': {
      name: 'GPT-5.4 Mini',
      limit: { context: 400000, output: 128000 },
      options: { store: false },
      variants: { low: {}, medium: {}, high: {}, xhigh: {} }
    },
    'gpt-5.3-codex-spark': {
      name: 'GPT-5.3 Codex Spark',
      limit: { context: 128000, output: 32000 },
      options: { store: false },
      variants: { low: {}, medium: {}, high: {}, xhigh: {} }
    },
    'gpt-5.3-codex': {
      name: 'GPT-5.3 Codex',
      limit: { context: 400000, output: 128000 },
      options: { store: false },
      variants: { low: {}, medium: {}, high: {}, xhigh: {} }
    },
    'codex-mini-latest': {
      name: 'Codex Mini',
      limit: { context: 200000, output: 100000 },
      options: { store: false },
      variants: { low: {}, medium: {}, high: {} }
    }
  }
}

function buildGeminiModels(): OpenCodeModelMap {
  return {
    'gemini-2.0-flash': {
      name: 'Gemini 2.0 Flash',
      limit: { context: 1048576, output: 65536 },
      modalities: { input: ['text', 'image', 'pdf'], output: ['text'] }
    },
    'gemini-2.5-flash': {
      name: 'Gemini 2.5 Flash',
      limit: { context: 1048576, output: 65536 },
      modalities: { input: ['text', 'image', 'pdf'], output: ['text'] }
    },
    'gemini-2.5-pro': {
      name: 'Gemini 2.5 Pro',
      limit: { context: 2097152, output: 65536 },
      modalities: { input: ['text', 'image', 'pdf'], output: ['text'] },
      options: { thinking: { budgetTokens: 24576, type: 'enabled' } }
    },
    'gemini-3-flash-preview': {
      name: 'Gemini 3 Flash Preview',
      limit: { context: 1048576, output: 65536 },
      modalities: { input: ['text', 'image', 'pdf'], output: ['text'] }
    },
    'gemini-3-pro-preview': {
      name: 'Gemini 3 Pro Preview',
      limit: { context: 1048576, output: 65536 },
      modalities: { input: ['text', 'image', 'pdf'], output: ['text'] },
      options: { thinking: { budgetTokens: 24576, type: 'enabled' } }
    },
    'gemini-3.1-pro-preview': {
      name: 'Gemini 3.1 Pro Preview',
      limit: { context: 1048576, output: 65536 },
      modalities: { input: ['text', 'image', 'pdf'], output: ['text'] },
      options: { thinking: { budgetTokens: 24576, type: 'enabled' } }
    }
  }
}

function buildAntigravityGeminiModels(): OpenCodeModelMap {
  return {
    'gemini-2.5-flash': {
      name: 'Gemini 2.5 Flash',
      limit: { context: 1048576, output: 65536 },
      modalities: { input: ['text', 'image', 'pdf'], output: ['text'] },
      options: { thinking: { budgetTokens: 24576, type: 'disable' } }
    },
    'gemini-2.5-flash-lite': {
      name: 'Gemini 2.5 Flash Lite',
      limit: { context: 1048576, output: 65536 },
      modalities: { input: ['text', 'image', 'pdf'], output: ['text'] },
      options: { thinking: { budgetTokens: 24576, type: 'enabled' } }
    },
    'gemini-2.5-flash-thinking': {
      name: 'Gemini 2.5 Flash (Thinking)',
      limit: { context: 1048576, output: 65536 },
      modalities: { input: ['text', 'image', 'pdf'], output: ['text'] },
      options: { thinking: { budgetTokens: 24576, type: 'enabled' } }
    },
    'gemini-3-flash': {
      name: 'Gemini 3 Flash',
      limit: { context: 1048576, output: 65536 },
      modalities: { input: ['text', 'image', 'pdf'], output: ['text'] },
      options: { thinking: { budgetTokens: 24576, type: 'enabled' } }
    },
    'gemini-3.1-pro-low': {
      name: 'Gemini 3.1 Pro Low',
      limit: { context: 1048576, output: 65536 },
      modalities: { input: ['text', 'image', 'pdf'], output: ['text'] },
      options: { thinking: { budgetTokens: 24576, type: 'enabled' } }
    },
    'gemini-3.1-pro-high': {
      name: 'Gemini 3.1 Pro High',
      limit: { context: 1048576, output: 65536 },
      modalities: { input: ['text', 'image', 'pdf'], output: ['text'] },
      options: { thinking: { budgetTokens: 24576, type: 'enabled' } }
    },
    'gemini-2.5-flash-image': {
      name: 'Gemini 2.5 Flash Image',
      limit: { context: 1048576, output: 65536 },
      modalities: { input: ['text', 'image'], output: ['image'] },
      options: { thinking: { budgetTokens: 24576, type: 'enabled' } }
    },
    'gemini-3.1-flash-image': {
      name: 'Gemini 3.1 Flash Image',
      limit: { context: 1048576, output: 65536 },
      modalities: { input: ['text', 'image'], output: ['image'] },
      options: { thinking: { budgetTokens: 24576, type: 'enabled' } }
    }
  }
}

function buildClaudeModels(): OpenCodeModelMap {
  return {
    'claude-opus-4-6-thinking': {
      name: 'Claude 4.6 Opus (Thinking)',
      limit: { context: 200000, output: 128000 },
      modalities: { input: ['text', 'image', 'pdf'], output: ['text'] },
      options: { thinking: { budgetTokens: 24576, type: 'enabled' } }
    },
    'claude-sonnet-4-6': {
      name: 'Claude 4.6 Sonnet',
      limit: { context: 200000, output: 64000 },
      modalities: { input: ['text', 'image', 'pdf'], output: ['text'] },
      options: { thinking: { budgetTokens: 24576, type: 'enabled' } }
    }
  }
}

function buildOpenCodeConfig(
  platform: string,
  baseUrl: string,
  apiKey: string,
  t: (key: string) => string,
  pathLabel?: string
): UseKeyModalFileConfig {
  const provider: Record<string, Record<string, unknown>> = {
    [platform]: {
      options: {
        baseURL: baseUrl,
        apiKey
      }
    }
  }

  if (platform === 'gemini') {
    provider[platform].npm = '@ai-sdk/google'
    provider[platform].models = buildGeminiModels()
  } else if (platform === 'anthropic') {
    provider[platform].npm = '@ai-sdk/anthropic'
  } else if (platform === 'antigravity-claude') {
    provider[platform].npm = '@ai-sdk/anthropic'
    provider[platform].name = 'Antigravity (Claude)'
    provider[platform].models = buildClaudeModels()
  } else if (platform === 'antigravity-gemini') {
    provider[platform].npm = '@ai-sdk/google'
    provider[platform].name = 'Antigravity (Gemini)'
    provider[platform].models = buildAntigravityGeminiModels()
  } else if (platform === 'openai') {
    provider[platform].models = buildOpenAIModels()
  }

  const agent =
    platform === 'openai'
      ? {
          build: { options: { store: false } },
          plan: { options: { store: false } }
        }
      : undefined

  const content = JSON.stringify(
    {
      provider,
      ...(agent ? { agent } : {}),
      $schema: 'https://opencode.ai/config.json'
    },
    null,
    2
  )

  return {
    path: pathLabel ?? 'opencode.json',
    content,
    hint: t('keys.useKeyModal.opencode.hint')
  }
}

export function buildUseKeyModalFiles(
  options: BuildUseKeyModalFilesOptions
): UseKeyModalFileConfig[] {
  const baseUrl = options.baseUrl || window.location.origin
  const apiKey = options.apiKey
  const baseRoot = baseUrl.replace(/\/v1\/?$/, '').replace(/\/+$/, '')
  const ensureV1 = (value: string) => {
    const trimmed = value.replace(/\/+$/, '')
    return trimmed.endsWith('/v1') ? trimmed : `${trimmed}/v1`
  }

  const apiBase = ensureV1(baseRoot)
  const grokBase = ensureV1(`${baseRoot}/grok`)
  const antigravityBase = ensureV1(`${baseRoot}/antigravity`)
  const antigravityGeminiBase = (() => {
    const trimmed = `${baseRoot}/antigravity`.replace(/\/+$/, '')
    return trimmed.endsWith('/v1beta') ? trimmed : `${trimmed}/v1beta`
  })()
  const geminiBase = (() => {
    const trimmed = baseRoot.replace(/\/+$/, '')
    return trimmed.endsWith('/v1beta') ? trimmed : `${trimmed}/v1beta`
  })()

  if (options.activeClientTab === 'opencode') {
    switch (options.platform) {
      case 'anthropic':
        return [buildOpenCodeConfig('anthropic', apiBase, apiKey, options.t)]
      case 'openai':
        return [buildOpenCodeConfig('openai', apiBase, apiKey, options.t)]
      case 'grok':
        return [buildOpenCodeConfig('openai', grokBase, apiKey, options.t)]
      case 'gemini':
        return [buildOpenCodeConfig('gemini', geminiBase, apiKey, options.t)]
      case 'antigravity':
        return [
          buildOpenCodeConfig(
            'antigravity-claude',
            antigravityBase,
            apiKey,
            options.t,
            'opencode.json (Claude)'
          ),
          buildOpenCodeConfig(
            'antigravity-gemini',
            antigravityGeminiBase,
            apiKey,
            options.t,
            'opencode.json (Gemini)'
          )
        ]
      default:
        return [buildOpenCodeConfig('openai', apiBase, apiKey, options.t)]
    }
  }

  switch (options.platform) {
    case 'openai':
    case 'grok':
      if (options.activeClientTab === 'claude') {
        if (options.platform === 'grok') {
          return buildAnthropicFiles(options.activeTab, `${baseUrl}/grok`, apiKey)
        }
        return buildAnthropicFiles(options.activeTab, baseUrl, apiKey)
      }
      if (options.activeClientTab === 'codex-ws') {
        if (options.platform === 'grok') {
          return buildOpenAIFiles(options.activeTab, `${baseUrl}/grok`, apiKey, true)
        }
        return buildOpenAIFiles(options.activeTab, baseUrl, apiKey, true)
      }
      if (options.platform === 'grok') {
        return buildOpenAIFiles(options.activeTab, `${baseUrl}/grok`, apiKey, false)
      }
      return buildOpenAIFiles(options.activeTab, baseUrl, apiKey, false)
    case 'gemini':
      return [buildGeminiCliFile(options.activeTab, baseUrl, apiKey, options.t)]
    case 'antigravity':
      if (options.activeClientTab === 'gemini') {
        return [buildGeminiCliFile(options.activeTab, `${baseUrl}/antigravity`, apiKey, options.t)]
      }
      return buildAnthropicFiles(options.activeTab, `${baseUrl}/antigravity`, apiKey)
    default:
      return buildAnthropicFiles(options.activeTab, baseUrl, apiKey)
  }
}
