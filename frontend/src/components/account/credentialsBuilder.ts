import { buildModelMappingObject } from '@/composables/useModelWhitelist'
import type { AccountPlatform } from '@/types'

export interface ModelMapping {
  from: string
  to: string
}

export interface TempUnschedRuleForm {
  error_code: number | null
  keywords: string
  duration_minutes: number | null
  description: string
}

export interface TempUnschedRulePayload {
  error_code: number
  keywords: string[]
  duration_minutes: number
  description: string
}

export const DEFAULT_POOL_MODE_RETRY_COUNT = 3
export const MAX_POOL_MODE_RETRY_COUNT = 10
export const DEFAULT_ANTHROPIC_BASE_URL = 'https://api.anthropic.com'
export const DEFAULT_OPENAI_BASE_URL = 'https://api.openai.com'
export const DEFAULT_GEMINI_BASE_URL = 'https://generativelanguage.googleapis.com'
export const DEFAULT_ANTIGRAVITY_BASE_URL = 'https://cloudcode-pa.googleapis.com'
export const OPENAI_COMPATIBLE_XAI_BASE_URL = 'https://api.x.ai'

export function applyInterceptWarmup(
  credentials: Record<string, unknown>,
  enabled: boolean,
  mode: 'create' | 'edit'
): void {
  if (enabled) {
    credentials.intercept_warmup_requests = true
  } else if (mode === 'edit') {
    delete credentials.intercept_warmup_requests
  }
}

export function getDefaultBaseURL(platform: AccountPlatform): string {
  if (platform === 'openai') {
    return DEFAULT_OPENAI_BASE_URL
  }
  if (platform === 'gemini') {
    return DEFAULT_GEMINI_BASE_URL
  }
  return DEFAULT_ANTHROPIC_BASE_URL
}

export function normalizePoolModeRetryCount(value: number): number {
  if (!Number.isFinite(value)) {
    return DEFAULT_POOL_MODE_RETRY_COUNT
  }
  const normalized = Math.trunc(value)
  if (normalized < 0) {
    return 0
  }
  if (normalized > MAX_POOL_MODE_RETRY_COUNT) {
    return MAX_POOL_MODE_RETRY_COUNT
  }
  return normalized
}

export function assignBuiltModelMapping(
  credentials: Record<string, unknown>,
  mode: 'whitelist' | 'mapping',
  allowedModels: string[],
  modelMappings: ModelMapping[]
): Record<string, string> | null {
  const modelMapping = buildModelMappingObject(mode, allowedModels, modelMappings)
  if (modelMapping) {
    credentials.model_mapping = modelMapping
  }
  return modelMapping
}

export function replaceBuiltModelMapping(
  credentials: Record<string, unknown>,
  mode: 'whitelist' | 'mapping',
  allowedModels: string[],
  modelMappings: ModelMapping[]
): Record<string, string> | null {
  const modelMapping = buildModelMappingObject(mode, allowedModels, modelMappings)
  if (modelMapping) {
    credentials.model_mapping = modelMapping
  } else {
    delete credentials.model_mapping
  }
  return modelMapping
}

export function replaceAntigravityModelMapping(
  credentials: Record<string, unknown>,
  modelMappings: ModelMapping[]
): Record<string, string> | null {
  delete credentials.model_whitelist
  return replaceBuiltModelMapping(credentials, 'mapping', [], modelMappings)
}

export function createTempUnschedRule(preset?: TempUnschedRuleForm): TempUnschedRuleForm {
  if (preset) {
    return { ...preset }
  }
  return {
    error_code: null,
    keywords: '',
    duration_minutes: 30,
    description: ''
  }
}

export function moveItemInPlace<T>(items: T[], index: number, direction: number): void {
  const target = index + direction
  if (target < 0 || target >= items.length) {
    return
  }
  const current = items[index]
  items[index] = items[target]
  items[target] = current
}

export function buildTempUnschedRules(rules: TempUnschedRuleForm[]): TempUnschedRulePayload[] {
  const out: TempUnschedRulePayload[] = []

  for (const rule of rules) {
    const errorCode = Number(rule.error_code)
    const duration = Number(rule.duration_minutes)
    const keywords = splitTempUnschedKeywords(rule.keywords)
    if (!Number.isFinite(errorCode) || errorCode < 100 || errorCode > 599) {
      continue
    }
    if (!Number.isFinite(duration) || duration <= 0) {
      continue
    }
    if (keywords.length === 0) {
      continue
    }
    out.push({
      error_code: Math.trunc(errorCode),
      keywords,
      duration_minutes: Math.trunc(duration),
      description: rule.description.trim()
    })
  }

  return out
}

export function applyTempUnschedConfig(
  credentials: Record<string, unknown>,
  enabled: boolean,
  rules: TempUnschedRuleForm[]
): boolean {
  if (!enabled) {
    delete credentials.temp_unschedulable_enabled
    delete credentials.temp_unschedulable_rules
    return true
  }

  const payload = buildTempUnschedRules(rules)
  if (payload.length === 0) {
    return false
  }

  credentials.temp_unschedulable_enabled = true
  credentials.temp_unschedulable_rules = payload
  return true
}

export function loadTempUnschedRuleState(
  credentials?: Record<string, unknown>
): { enabled: boolean; rules: TempUnschedRuleForm[] } {
  const enabled = credentials?.temp_unschedulable_enabled === true
  const rawRules = credentials?.temp_unschedulable_rules
  if (!Array.isArray(rawRules)) {
    return { enabled, rules: [] }
  }

  return {
    enabled,
    rules: rawRules.map((rule) => {
      const entry = rule as Record<string, unknown>
      return {
        error_code: toPositiveNumber(entry.error_code),
        keywords: formatTempUnschedKeywords(entry.keywords),
        duration_minutes: toPositiveNumber(entry.duration_minutes),
        description: typeof entry.description === 'string' ? entry.description : ''
      }
    })
  }
}

function splitTempUnschedKeywords(value: string): string[] {
  return value
    .split(/[,;]/)
    .map((item) => item.trim())
    .filter((item) => item.length > 0)
}

function formatTempUnschedKeywords(value: unknown): string {
  if (Array.isArray(value)) {
    return value
      .filter((item): item is string => typeof item === 'string')
      .map((item) => item.trim())
      .filter((item) => item.length > 0)
      .join(', ')
  }
  if (typeof value === 'string') {
    return value
  }
  return ''
}

function toPositiveNumber(value: unknown): number | null {
  const num = Number(value)
  if (!Number.isFinite(num) || num <= 0) {
    return null
  }
  return Math.trunc(num)
}
