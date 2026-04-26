import { ref } from 'vue'
import { getAntigravityDefaultModelMapping } from '@/api/admin/accounts'
import {
  getModelCatalog,
  type AdminModelCatalogEntry
} from '@/api/admin/modelCatalog'

// =====================
// 模型列表（硬编码，与 new-api 一致）
// =====================

// OpenAI
const openaiModels = [
  'gpt-3.5-turbo', 'gpt-3.5-turbo-0125', 'gpt-3.5-turbo-1106', 'gpt-3.5-turbo-16k',
  'gpt-4', 'gpt-4-turbo', 'gpt-4-turbo-preview',
  'gpt-4o', 'gpt-4o-2024-08-06', 'gpt-4o-2024-11-20',
  'gpt-4o-mini', 'gpt-4o-mini-2024-07-18',
  'gpt-4.5-preview',
  'gpt-4.1', 'gpt-4.1-mini', 'gpt-4.1-nano',
  'o1', 'o1-preview', 'o1-mini', 'o1-pro',
  'o3', 'o3-mini', 'o3-pro',
  'o4-mini',
  // GPT-5.2 系列
  'gpt-5.2', 'gpt-5.2-2025-12-11', 'gpt-5.2-chat-latest',
  'gpt-5.2-pro', 'gpt-5.2-pro-2025-12-11',
  // GPT-5.5 系列
  'gpt-5.5', 'gpt-5.5-2026-04-23',
  // GPT-5.4 系列
  'gpt-5.4', 'gpt-5.4-mini', 'gpt-5.4-2026-03-05',
  // GPT-5.3 系列
  'gpt-5.3-codex', 'gpt-5.3-codex-spark',
  // GPT Image 系列
  'gpt-image-1', 'gpt-image-1.5', 'gpt-image-2', 'gpt-image-2-2026-04-21',
  'chatgpt-4o-latest',
  'gpt-4o-audio-preview', 'gpt-4o-realtime-preview'
]

// Anthropic Claude
export const claudeModels = [
  'claude-3-5-sonnet-20241022', 'claude-3-5-sonnet-20240620',
  'claude-3-5-haiku-20241022',
  'claude-3-opus-20240229', 'claude-3-sonnet-20240229', 'claude-3-haiku-20240307',
  'claude-3-7-sonnet-20250219',
  'claude-sonnet-4-20250514', 'claude-opus-4-20250514',
  'claude-opus-4-1-20250805',
  'claude-sonnet-4-5-20250929', 'claude-haiku-4-5-20251001',
  'claude-opus-4-5-20251101',
  'claude-opus-4-6',
  'claude-opus-4-7',
  'claude-sonnet-4-6',
  'claude-2.1', 'claude-2.0', 'claude-instant-1.2'
]

// Google Gemini
const geminiModels = [
  // Keep in sync with backend curated Gemini lists.
  // This list is intentionally conservative (models commonly available across OAuth/API key).
  'gemini-3.1-flash-image',
  'gemini-2.5-flash-image',
  'gemini-2.0-flash',
  'gemini-2.5-flash',
  'gemini-2.5-pro',
  'gemini-3-flash-preview',
  'gemini-3-pro-preview'
]

// Antigravity 官方支持的模型（精确匹配）
// 基于官方 API 返回的模型列表，只支持 Claude 4.5+ 和 Gemini 2.5+
const antigravityModels = [
  // Claude 4.5+ 系列
  'claude-opus-4-6',
  'claude-opus-4-6-thinking',
  'claude-opus-4-7',
  'claude-opus-4-5-thinking',
  'claude-sonnet-4-6',
  'claude-sonnet-4-5',
  'claude-sonnet-4-5-thinking',
  // Gemini 2.5 系列
  'gemini-3.1-flash-image',
  'gemini-2.5-flash-image',
  'gemini-2.5-flash',
  'gemini-2.5-flash-lite',
  'gemini-2.5-flash-thinking',
  'gemini-2.5-pro',
  // Gemini 3 系列
  'gemini-3-flash',
  'gemini-3-pro-high',
  'gemini-3-pro-low',
  // Gemini 3.1 系列
  'gemini-3.1-pro-high',
  'gemini-3.1-pro-low',
  'gemini-3-pro-image',
  // 其他
  'gpt-oss-120b-medium',
  'tab_flash_lite_preview'
]

// 智谱 GLM
const zhipuModels = [
  'glm-4', 'glm-4v', 'glm-4-plus', 'glm-4-0520',
  'glm-4-air', 'glm-4-airx', 'glm-4-long', 'glm-4-flash',
  'glm-4v-plus', 'glm-4.5', 'glm-4.6',
  'glm-3-turbo', 'glm-4-alltools',
  'chatglm_turbo', 'chatglm_pro', 'chatglm_std', 'chatglm_lite',
  'cogview-3', 'cogvideo'
]

// 阿里 通义千问
const qwenModels = [
  'qwen-turbo', 'qwen-plus', 'qwen-max', 'qwen-max-longcontext', 'qwen-long',
  'qwen2-72b-instruct', 'qwen2-57b-a14b-instruct', 'qwen2-7b-instruct',
  'qwen2.5-72b-instruct', 'qwen2.5-32b-instruct', 'qwen2.5-14b-instruct',
  'qwen2.5-7b-instruct', 'qwen2.5-3b-instruct', 'qwen2.5-1.5b-instruct',
  'qwen2.5-coder-32b-instruct', 'qwen2.5-coder-14b-instruct', 'qwen2.5-coder-7b-instruct',
  'qwen3-235b-a22b',
  'qwq-32b', 'qwq-32b-preview'
]

// DeepSeek
const deepseekModels = [
  'deepseek-chat', 'deepseek-coder', 'deepseek-reasoner',
  'deepseek-v3', 'deepseek-v3-0324',
  'deepseek-r1', 'deepseek-r1-0528',
  'deepseek-r1-distill-qwen-32b', 'deepseek-r1-distill-qwen-14b', 'deepseek-r1-distill-qwen-7b',
  'deepseek-r1-distill-llama-70b', 'deepseek-r1-distill-llama-8b'
]

// Mistral
const mistralModels = [
  'mistral-small-latest', 'mistral-medium-latest', 'mistral-large-latest',
  'open-mistral-7b', 'open-mixtral-8x7b', 'open-mixtral-8x22b',
  'codestral-latest', 'codestral-mamba',
  'pixtral-12b-2409', 'pixtral-large-latest'
]

// Meta Llama
const metaModels = [
  'llama-3.3-70b-instruct',
  'llama-3.2-90b-vision-instruct', 'llama-3.2-11b-vision-instruct',
  'llama-3.2-3b-instruct', 'llama-3.2-1b-instruct',
  'llama-3.1-405b-instruct', 'llama-3.1-70b-instruct', 'llama-3.1-8b-instruct',
  'llama-3-70b-instruct', 'llama-3-8b-instruct',
  'codellama-70b-instruct', 'codellama-34b-instruct', 'codellama-13b-instruct'
]

type ModelOption = {
  value: string
  label: string
}

const grokCatalogEntries = ref<AdminModelCatalogEntry[]>([])
let grokCatalogLoaded = false
let grokCatalogPromise: Promise<AdminModelCatalogEntry[]> | null = null

function dedupeModelOptions(options: ModelOption[]): ModelOption[] {
  const seen = new Set<string>()
  const deduped: ModelOption[] = []
  for (const option of options) {
    if (seen.has(option.value)) {
      continue
    }
    seen.add(option.value)
    deduped.push(option)
  }
  return deduped
}

function buildStaticModelOptions(models: string[]): ModelOption[] {
  return models.map((model) => ({ value: model, label: model }))
}

function normalizePlatformKey(platform: string): string {
  const normalized = platform.trim().toLowerCase()
  if (normalized === 'xai') {
    return 'grok'
  }
  if (normalized === 'claude') {
    return 'anthropic'
  }
  return normalized
}

function grokCatalogToOptions(entries: AdminModelCatalogEntry[]): ModelOption[] {
  return entries.map((entry) => ({
    value: entry.id,
    label: entry.display_name ? `${entry.display_name} (${entry.id})` : entry.id
  }))
}

function getGrokPresetTone(entry: AdminModelCatalogEntry): PresetMappingTone {
  switch (entry.capability) {
    case 'video':
      return 'brand-rose'
    case 'image':
    case 'image_edit':
      return 'warning'
    case 'voice':
      return 'success'
    default:
      if (entry.required_tier === 'heavy') {
        return 'brand-purple'
      }
      if (entry.required_tier === 'super') {
        return 'accent'
      }
      return 'info'
  }
}

function buildGrokPresetMappings(entries: AdminModelCatalogEntry[]): PresetMapping[] {
  return entries.map((entry) => ({
    label: entry.display_name || entry.id,
    from: entry.id,
    to: entry.id,
    tone: getGrokPresetTone(entry)
  }))
}

function setGrokCatalogEntries(entries: AdminModelCatalogEntry[]): AdminModelCatalogEntry[] {
  const seen = new Set<string>()
  const deduped: AdminModelCatalogEntry[] = []
  for (const entry of entries) {
    if (!entry?.id || seen.has(entry.id)) {
      continue
    }
    seen.add(entry.id)
    deduped.push(entry)
  }
  grokCatalogEntries.value = deduped
  return deduped
}

export async function ensureModelCatalogLoaded(platform: string): Promise<AdminModelCatalogEntry[]> {
  if (normalizePlatformKey(platform) !== 'grok') {
    return []
  }
  if (grokCatalogLoaded) {
    return grokCatalogEntries.value
  }
  if (grokCatalogPromise) {
    return grokCatalogPromise
  }

  grokCatalogPromise = getModelCatalog('grok')
    .then((response) => {
      grokCatalogLoaded = true
      return setGrokCatalogEntries(response.models)
    })
    .catch((error) => {
      grokCatalogLoaded = false
      setGrokCatalogEntries([])
      console.warn('[useModelWhitelist] Failed to load Grok model catalog', error)
      return []
    })
    .finally(() => {
      grokCatalogPromise = null
    })

  return grokCatalogPromise
}

function getGrokCatalogEntries(): AdminModelCatalogEntry[] {
  return grokCatalogEntries.value
}

function getStaticModelsByPlatform(platform: string): string[] {
  switch (normalizePlatformKey(platform)) {
    case 'openai': return openaiModels
    case 'anthropic': return claudeModels
    case 'gemini': return geminiModels
    case 'antigravity': return antigravityModels
    case 'zhipu': return zhipuModels
    case 'qwen': return qwenModels
    case 'deepseek': return deepseekModels
    case 'mistral': return mistralModels
    case 'meta': return metaModels
    case 'cohere': return cohereModels
    case 'yi': return yiModels
    case 'moonshot': return moonshotModels
    case 'doubao': return doubaoModels
    case 'minimax': return minimaxModels
    case 'baidu': return baiduModels
    case 'spark': return sparkModels
    case 'hunyuan': return hunyuanModels
    case 'perplexity': return perplexityModels
    default: return claudeModels
  }
}

export function getModelOptionsByPlatform(platform: string): ModelOption[] {
  if (normalizePlatformKey(platform) === 'grok') {
    return grokCatalogToOptions(getGrokCatalogEntries())
  }
  return buildStaticModelOptions(getStaticModelsByPlatform(platform))
}

// Cohere
const cohereModels = [
  'command-a-03-2025',
  'command-r', 'command-r-plus',
  'command-r-08-2024', 'command-r-plus-08-2024',
  'c4ai-aya-23-35b', 'c4ai-aya-23-8b',
  'command', 'command-light'
]

// Yi (01.AI)
const yiModels = [
  'yi-large', 'yi-large-turbo', 'yi-large-rag',
  'yi-medium', 'yi-medium-200k',
  'yi-spark', 'yi-vision',
  'yi-1.5-34b-chat', 'yi-1.5-9b-chat', 'yi-1.5-6b-chat'
]

// Moonshot/Kimi
const moonshotModels = [
  'moonshot-v1-8k', 'moonshot-v1-32k', 'moonshot-v1-128k',
  'kimi-latest'
]

// 字节跳动 豆包
const doubaoModels = [
  'doubao-pro-256k', 'doubao-pro-128k', 'doubao-pro-32k', 'doubao-pro-4k',
  'doubao-lite-128k', 'doubao-lite-32k', 'doubao-lite-4k',
  'doubao-vision-pro-32k', 'doubao-vision-lite-32k',
  'doubao-1.5-pro-256k', 'doubao-1.5-pro-32k', 'doubao-1.5-lite-32k',
  'doubao-1.5-pro-vision-32k', 'doubao-1.5-thinking-pro'
]

// MiniMax
const minimaxModels = [
  'abab6.5-chat', 'abab6.5s-chat', 'abab6.5s-chat-pro',
  'abab6-chat',
  'abab5.5-chat', 'abab5.5s-chat'
]

// 百度 文心
const baiduModels = [
  'ernie-4.0-8k-latest', 'ernie-4.0-8k', 'ernie-4.0-turbo-8k',
  'ernie-3.5-8k', 'ernie-3.5-128k',
  'ernie-speed-8k', 'ernie-speed-128k', 'ernie-speed-pro-128k',
  'ernie-lite-8k', 'ernie-lite-pro-128k',
  'ernie-tiny-8k'
]

// 讯飞 星火
const sparkModels = [
  'spark-desk', 'spark-desk-v1.1', 'spark-desk-v2.1',
  'spark-desk-v3.1', 'spark-desk-v3.5', 'spark-desk-v4.0',
  'spark-lite', 'spark-pro', 'spark-max', 'spark-ultra'
]

// 腾讯 混元
const hunyuanModels = [
  'hunyuan-lite', 'hunyuan-standard', 'hunyuan-standard-256k',
  'hunyuan-pro', 'hunyuan-turbo', 'hunyuan-large',
  'hunyuan-vision', 'hunyuan-code'
]

// Perplexity
const perplexityModels = [
  'sonar', 'sonar-pro', 'sonar-reasoning',
  'llama-3-sonar-small-32k-online', 'llama-3-sonar-large-32k-online',
  'llama-3-sonar-small-32k-chat', 'llama-3-sonar-large-32k-chat'
]

function getStaticAllModelOptions(): ModelOption[] {
  return dedupeModelOptions([
    ...buildStaticModelOptions(openaiModels),
    ...buildStaticModelOptions(claudeModels),
    ...buildStaticModelOptions(geminiModels),
    ...buildStaticModelOptions(zhipuModels),
    ...buildStaticModelOptions(qwenModels),
    ...buildStaticModelOptions(deepseekModels),
    ...buildStaticModelOptions(mistralModels),
    ...buildStaticModelOptions(metaModels),
    ...buildStaticModelOptions(cohereModels),
    ...buildStaticModelOptions(yiModels),
    ...buildStaticModelOptions(moonshotModels),
    ...buildStaticModelOptions(doubaoModels),
    ...buildStaticModelOptions(minimaxModels),
    ...buildStaticModelOptions(baiduModels),
    ...buildStaticModelOptions(sparkModels),
    ...buildStaticModelOptions(hunyuanModels),
    ...buildStaticModelOptions(perplexityModels),
    ...buildStaticModelOptions(antigravityModels)
  ])
}

export function getAllModelOptions(): ModelOption[] {
  return dedupeModelOptions([
    ...getStaticAllModelOptions(),
    ...getModelOptionsByPlatform('grok')
  ])
}

// =====================
// 预设映射
// =====================

export type PresetMappingTone =
  | 'accent'
  | 'brand-orange'
  | 'brand-purple'
  | 'brand-rose'
  | 'danger'
  | 'info'
  | 'success'
  | 'warning'

export interface PresetMapping {
  label: string
  from: string
  to: string
  tone: PresetMappingTone
}

const anthropicPresetMappings: PresetMapping[] = [
  { label: 'Sonnet 4', from: 'claude-sonnet-4-20250514', to: 'claude-sonnet-4-20250514', tone: 'info' },
  { label: 'Sonnet 4.5', from: 'claude-sonnet-4-5-20250929', to: 'claude-sonnet-4-5-20250929', tone: 'accent' },
  { label: 'Sonnet 4.6', from: 'claude-sonnet-4-6', to: 'claude-sonnet-4-6', tone: 'accent' },
  { label: 'Opus 4.5', from: 'claude-opus-4-5-20251101', to: 'claude-opus-4-5-20251101', tone: 'brand-purple' },
  { label: 'Opus 4.6', from: 'claude-opus-4-6', to: 'claude-opus-4-6', tone: 'brand-purple' },
  { label: 'Opus 4.7', from: 'claude-opus-4-7', to: 'claude-opus-4-7', tone: 'brand-purple' },
  { label: 'Haiku 3.5', from: 'claude-3-5-haiku-20241022', to: 'claude-3-5-haiku-20241022', tone: 'success' },
  { label: 'Haiku 4.5', from: 'claude-haiku-4-5-20251001', to: 'claude-haiku-4-5-20251001', tone: 'success' },
  { label: 'Opus->Sonnet', from: 'claude-opus-4-6', to: 'claude-sonnet-4-5-20250929', tone: 'warning' }
]

const openaiPresetMappings: PresetMapping[] = [
  { label: 'GPT-4o', from: 'gpt-4o', to: 'gpt-4o', tone: 'success' },
  { label: 'GPT-4o Mini', from: 'gpt-4o-mini', to: 'gpt-4o-mini', tone: 'info' },
  { label: 'GPT-4.1', from: 'gpt-4.1', to: 'gpt-4.1', tone: 'accent' },
  { label: 'o1', from: 'o1', to: 'o1', tone: 'brand-purple' },
  { label: 'o3', from: 'o3', to: 'o3', tone: 'success' },
  { label: 'GPT-5.3 Codex Spark', from: 'gpt-5.3-codex-spark', to: 'gpt-5.3-codex-spark', tone: 'success' },
  { label: 'GPT-5.2', from: 'gpt-5.2', to: 'gpt-5.2', tone: 'danger' },
  { label: 'GPT-5.5', from: 'gpt-5.5', to: 'gpt-5.5', tone: 'warning' },
  { label: 'GPT-5.4', from: 'gpt-5.4', to: 'gpt-5.4', tone: 'brand-rose' },
  { label: 'GPT Image 2', from: 'gpt-image-2', to: 'gpt-image-2', tone: 'info' },
  { label: 'Haiku→5.4', from: 'claude-haiku-4-5-20251001', to: 'gpt-5.4', tone: 'success' },
  { label: 'Opus→5.4', from: 'claude-opus-4-6', to: 'gpt-5.4', tone: 'brand-purple' },
  { label: 'Sonnet→5.4', from: 'claude-sonnet-4-6', to: 'gpt-5.4', tone: 'info' }
]

const geminiPresetMappings: PresetMapping[] = [
  { label: 'Flash 2.0', from: 'gemini-2.0-flash', to: 'gemini-2.0-flash', tone: 'info' },
  { label: '2.5 Flash', from: 'gemini-2.5-flash', to: 'gemini-2.5-flash', tone: 'accent' },
  { label: '2.5 Image', from: 'gemini-2.5-flash-image', to: 'gemini-2.5-flash-image', tone: 'info' },
  { label: '2.5 Pro', from: 'gemini-2.5-pro', to: 'gemini-2.5-pro', tone: 'brand-purple' },
  { label: '3.1 Image', from: 'gemini-3.1-flash-image', to: 'gemini-3.1-flash-image', tone: 'info' }
]

// Antigravity 预设映射（支持通配符）
const antigravityPresetMappings: PresetMapping[] = [
  // Claude 通配符映射
  { label: 'Claude→Sonnet', from: 'claude-*', to: 'claude-sonnet-4-5', tone: 'info' },
  { label: 'Sonnet→Sonnet', from: 'claude-sonnet-*', to: 'claude-sonnet-4-5', tone: 'accent' },
  { label: 'Opus→Opus', from: 'claude-opus-*', to: 'claude-opus-4-6-thinking', tone: 'brand-purple' },
  { label: 'Haiku→Sonnet', from: 'claude-haiku-*', to: 'claude-sonnet-4-5', tone: 'success' },
  { label: 'Sonnet4→4.6', from: 'claude-sonnet-4-20250514', to: 'claude-sonnet-4-6', tone: 'info' },
  { label: 'Sonnet4.5→4.6', from: 'claude-sonnet-4-5-20250929', to: 'claude-sonnet-4-6', tone: 'accent' },
  { label: 'Sonnet3.5→4.6', from: 'claude-3-5-sonnet-20241022', to: 'claude-sonnet-4-6', tone: 'success' },
  { label: 'Opus4.5→4.6', from: 'claude-opus-4-5-20251101', to: 'claude-opus-4-6-thinking', tone: 'brand-purple' },
  // Gemini 3→3.1 映射
  { label: '3-Pro-Preview→3.1-Pro-High', from: 'gemini-3-pro-preview', to: 'gemini-3.1-pro-high', tone: 'warning' },
  { label: '3-Pro-High→3.1-Pro-High', from: 'gemini-3-pro-high', to: 'gemini-3.1-pro-high', tone: 'brand-orange' },
  { label: '3-Pro-Low→3.1-Pro-Low', from: 'gemini-3-pro-low', to: 'gemini-3.1-pro-low', tone: 'warning' },
  { label: '3.1-Pro-High透传', from: 'gemini-3.1-pro-high', to: 'gemini-3.1-pro-high', tone: 'brand-orange' },
  { label: '3.1-Pro-Low透传', from: 'gemini-3.1-pro-low', to: 'gemini-3.1-pro-low', tone: 'warning' },
  // Gemini 通配符映射
  { label: 'Gemini 3→Flash', from: 'gemini-3*', to: 'gemini-3-flash', tone: 'warning' },
  { label: 'Gemini 2.5→Flash', from: 'gemini-2.5*', to: 'gemini-2.5-flash', tone: 'brand-orange' },
  { label: '2.5-Flash-Image透传', from: 'gemini-2.5-flash-image', to: 'gemini-2.5-flash-image', tone: 'info' },
  { label: '3.1-Flash-Image透传', from: 'gemini-3.1-flash-image', to: 'gemini-3.1-flash-image', tone: 'info' },
  { label: '3-Pro-Image→3.1', from: 'gemini-3-pro-image', to: 'gemini-3.1-flash-image', tone: 'info' },
  { label: '3-Flash透传', from: 'gemini-3-flash', to: 'gemini-3-flash', tone: 'success' },
  { label: '2.5-Flash-Lite透传', from: 'gemini-2.5-flash-lite', to: 'gemini-2.5-flash-lite', tone: 'success' },
  // 精确映射
  { label: 'Sonnet 4.6', from: 'claude-sonnet-4-6', to: 'claude-sonnet-4-6', tone: 'accent' },
  { label: 'Sonnet 4.5', from: 'claude-sonnet-4-5', to: 'claude-sonnet-4-5', tone: 'accent' },
  { label: 'Opus 4.6', from: 'claude-opus-4-6', to: 'claude-opus-4-6-thinking', tone: 'brand-rose' },
  { label: 'Opus 4.6-thinking', from: 'claude-opus-4-6-thinking', to: 'claude-opus-4-6-thinking', tone: 'brand-rose' },
  { label: 'Opus 4.7', from: 'claude-opus-4-7', to: 'claude-opus-4-7', tone: 'brand-rose' }
]

// Bedrock 预设映射（与后端 DefaultBedrockModelMapping 保持一致）
const bedrockPresetMappings: PresetMapping[] = [
  { label: 'Opus 4.7', from: 'claude-opus-4-7', to: 'us.anthropic.claude-opus-4-7-v1', tone: 'brand-rose' },
  { label: 'Opus 4.6', from: 'claude-opus-4-6', to: 'us.anthropic.claude-opus-4-6-v1', tone: 'brand-rose' },
  { label: 'Sonnet 4.6', from: 'claude-sonnet-4-6', to: 'us.anthropic.claude-sonnet-4-6', tone: 'accent' },
  { label: 'Opus 4.5', from: 'claude-opus-4-5-thinking', to: 'us.anthropic.claude-opus-4-5-20251101-v1:0', tone: 'brand-rose' },
  { label: 'Sonnet 4.5', from: 'claude-sonnet-4-5', to: 'us.anthropic.claude-sonnet-4-5-20250929-v1:0', tone: 'accent' },
  { label: 'Haiku 4.5', from: 'claude-haiku-4-5', to: 'us.anthropic.claude-haiku-4-5-20251001-v1:0', tone: 'success' },
]

let _antigravityDefaultMappingsCache: { from: string; to: string }[] | null = null

export async function fetchAntigravityDefaultMappings(): Promise<{ from: string; to: string }[]> {
  if (_antigravityDefaultMappingsCache !== null) {
    return _antigravityDefaultMappingsCache
  }
  try {
    const mapping = await getAntigravityDefaultModelMapping()
    _antigravityDefaultMappingsCache = Object.entries(mapping).map(([from, to]) => ({ from, to }))
  } catch (e) {
    console.warn('[fetchAntigravityDefaultMappings] API failed, using empty fallback', e)
    _antigravityDefaultMappingsCache = []
  }
  return _antigravityDefaultMappingsCache
}

// =====================
// 常用错误码
// =====================

export const commonErrorCodes = [
  { value: 401, label: 'Unauthorized' },
  { value: 403, label: 'Forbidden' },
  { value: 429, label: 'Rate Limit' },
  { value: 500, label: 'Server Error' },
  { value: 502, label: 'Bad Gateway' },
  { value: 503, label: 'Unavailable' },
  { value: 529, label: 'Overloaded' }
]

// =====================
// 辅助函数
// =====================

// 按平台获取模型
export function getModelsByPlatform(platform: string): string[] {
  if (normalizePlatformKey(platform) === 'grok') {
    return getGrokCatalogEntries().map((entry) => entry.id)
  }
  return getStaticModelsByPlatform(platform)
}

// 按平台获取预设映射
export function getPresetMappingsByPlatform(platform: string) {
  const normalized = normalizePlatformKey(platform)
  if (normalized === 'openai') return openaiPresetMappings
  if (normalized === 'gemini') return geminiPresetMappings
  if (normalized === 'grok') return buildGrokPresetMappings(getGrokCatalogEntries())
  if (normalized === 'antigravity') return antigravityPresetMappings
  if (normalized === 'bedrock') return bedrockPresetMappings
  return anthropicPresetMappings
}

export function getPresetMappingChipClasses(tone: PresetMappingTone) {
  return [
    'theme-chip',
    'theme-chip--regular',
    'cursor-pointer',
    'hover:opacity-90',
    tone === 'accent' && 'theme-chip--accent',
    tone === 'brand-orange' && 'theme-chip--brand-orange',
    tone === 'brand-purple' && 'theme-chip--brand-purple',
    tone === 'brand-rose' && 'theme-chip--brand-rose',
    tone === 'danger' && 'theme-chip--danger',
    tone === 'info' && 'theme-chip--info',
    tone === 'success' && 'theme-chip--success',
    tone === 'warning' && 'theme-chip--warning'
  ].filter(Boolean).join(' ')
}

// =====================
// 构建模型映射对象（用于 API）
// =====================

// isValidWildcardPattern 校验通配符格式：* 只能放在末尾
// 导出供表单组件使用实时校验
export function isValidWildcardPattern(pattern: string): boolean {
  const starIndex = pattern.indexOf('*')
  if (starIndex === -1) return true // 无通配符，有效
  // * 必须在末尾，且只能有一个
  return starIndex === pattern.length - 1 && pattern.lastIndexOf('*') === starIndex
}

export function buildModelMappingObject(
  mode: 'whitelist' | 'mapping',
  allowedModels: string[],
  modelMappings: { from: string; to: string }[]
): Record<string, string> | null {
  const mapping: Record<string, string> = {}

  if (mode === 'whitelist') {
    for (const model of allowedModels) {
      // whitelist 模式的本意是"精确模型列表"，如果用户输入了通配符（如 claude-*），
      // 写入 model_mapping 会导致 GetMappedModel() 把真实模型映射成 "claude-*"，从而转发失败。
      // 因此这里跳过包含通配符的条目。
      if (!model.includes('*')) {
        mapping[model] = model
      }
    }
  } else {
    for (const m of modelMappings) {
      const from = m.from.trim()
      const to = m.to.trim()
      if (!from || !to) continue
      // 校验通配符格式：* 只能放在末尾
      if (!isValidWildcardPattern(from)) {
        console.warn(`[buildModelMappingObject] 无效的通配符格式，跳过: ${from}`)
        continue
      }
      // to 不允许包含通配符
      if (to.includes('*')) {
        console.warn(`[buildModelMappingObject] 目标模型不能包含通配符，跳过: ${from} -> ${to}`)
        continue
      }
      mapping[from] = to
    }
  }

  return Object.keys(mapping).length > 0 ? mapping : null
}
