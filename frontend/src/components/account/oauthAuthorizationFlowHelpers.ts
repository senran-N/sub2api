import type { AccountPlatform } from '@/types'

export function resolveOAuthKeyPrefix(platform: AccountPlatform) {
  if (platform === 'openai') {
    return 'admin.accounts.oauth.openai'
  }
  if (platform === 'gemini') {
    return 'admin.accounts.oauth.gemini'
  }
  if (platform === 'antigravity') {
    return 'admin.accounts.oauth.antigravity'
  }
  return 'admin.accounts.oauth'
}

export function resolveOAuthKey(platform: AccountPlatform, key: string) {
  return `${resolveOAuthKeyPrefix(platform)}.${key}`
}

export function countMultilineEntries(value: string) {
  return value
    .split('\n')
    .map((item) => item.trim())
    .filter((item) => item).length
}

export function extractOAuthCallbackPayload(value: string): {
  authCode: string | null
  oauthState: string | null
} {
  const trimmed = value.trim()
  if (!trimmed.includes('?') || !trimmed.includes('code=')) {
    return {
      authCode: null,
      oauthState: null
    }
  }

  try {
    const url = new URL(trimmed)
    return {
      authCode: url.searchParams.get('code'),
      oauthState: url.searchParams.get('state')
    }
  } catch {
    const codeMatch = trimmed.match(/[?&]code=([^&]+)/)
    const stateMatch = trimmed.match(/[?&]state=([^&]+)/)
    return {
      authCode: codeMatch?.[1] || null,
      oauthState: stateMatch?.[1] || null
    }
  }
}

export function shouldExtractOAuthCallback(platform: AccountPlatform) {
  return platform === 'openai' || platform === 'gemini' || platform === 'antigravity'
}

export function resolveOAuthImportantNoticeKey(platform: AccountPlatform) {
  if (platform === 'openai') {
    return 'admin.accounts.oauth.openai.importantNotice'
  }
  if (platform === 'antigravity') {
    return 'admin.accounts.oauth.antigravity.importantNotice'
  }
  return null
}

export function extractNormalizedOAuthCallback(platform: AccountPlatform, value: string): {
  authCode: string | null
  nextInputValue: string | null
  oauthState: string | null
} {
  if (!shouldExtractOAuthCallback(platform)) {
    return {
      authCode: null,
      nextInputValue: null,
      oauthState: null
    }
  }

  const extracted = extractOAuthCallbackPayload(value)
  return {
    authCode: extracted.authCode,
    nextInputValue: extracted.authCode && extracted.authCode !== value.trim() ? extracted.authCode : null,
    oauthState: extracted.oauthState
  }
}

export function parseBatchTokenInput(value: string) {
  return value
    .split('\n')
    .map((item) => item.trim())
    .filter((item) => item)
}

export interface RunBatchCreateFlowOptions {
  rawInput: string
  emptyInputMessage: string
  loadingRef: { value: boolean }
  errorRef: { value: string }
  isActive?: () => boolean
  onComplete?: (result: {
    failedCount: number
    successCount: number
    errors: string[]
  }) => void
  processEntry: (entry: string, index: number, entries: string[]) => Promise<string | null>
  resolveUnexpectedError?: (error: unknown) => string
}

export async function runBatchCreateFlow(options: RunBatchCreateFlowOptions) {
  if (!options.rawInput.trim()) return

  const entries = parseBatchTokenInput(options.rawInput)
  if (entries.length === 0) {
    options.errorRef.value = options.emptyInputMessage
    return
  }

  const isActive = () => options.isActive?.() ?? true
  if (!isActive()) {
    return
  }

  options.loadingRef.value = true
  options.errorRef.value = ''

  let successCount = 0
  let failedCount = 0
  const errors: string[] = []

  try {
    for (let index = 0; index < entries.length; index++) {
      if (!isActive()) {
        return
      }
      try {
        const entryError = await options.processEntry(entries[index], index, entries)
        if (!isActive()) {
          return
        }
        if (entryError) {
          failedCount++
          errors.push(`#${index + 1}: ${entryError}`)
          continue
        }
        successCount++
      } catch (error) {
        failedCount++
        const message = options.resolveUnexpectedError
          ? options.resolveUnexpectedError(error)
          : error instanceof Error
            ? error.message
            : 'Unknown error'
        if (!isActive()) {
          return
        }
        errors.push(`#${index + 1}: ${message}`)
      }
    }

    if (!isActive()) {
      return
    }
    options.onComplete?.({ failedCount, successCount, errors })
  } finally {
    if (isActive()) {
      options.loadingRef.value = false
    }
  }
}

export function consumeValidationFailureMessage(
  errorRef: { value: string },
  fallbackMessage = 'Validation failed'
) {
  const message = errorRef.value || fallbackMessage
  errorRef.value = ''
  return message
}

export function resolveOAuthExchangeState(options: {
  fallbackState?: string
  inputState?: string
  onMissingState: (message: string) => void
  authFailedMessage: string
}) {
  const stateToUse = (options.inputState || options.fallbackState || '').trim()
  if (stateToUse) {
    return stateToUse
  }

  options.onMissingState(options.authFailedMessage)
  return null
}

export async function runOAuthExchangeFlow(
  stateRefs: { loading: { value: boolean }; error: { value: string } },
  action: () => Promise<void>,
  resolveErrorMessage: (error: unknown) => string,
  showError: (message: string) => void,
  options?: {
    isActive?: () => boolean
  }
) {
  const isActive = () => options?.isActive?.() ?? true
  if (!isActive()) {
    return
  }

  stateRefs.loading.value = true
  stateRefs.error.value = ''

  try {
    await action()
  } catch (error) {
    if (!isActive()) {
      return
    }
    stateRefs.error.value = resolveErrorMessage(error)
    showError(stateRefs.error.value)
  } finally {
    if (isActive()) {
      stateRefs.loading.value = false
    }
  }
}

export function resolveAnthropicExchangeEndpoint(
  addMethod: 'oauth' | 'setup-token',
  mode: 'code' | 'cookie'
) {
  if (mode === 'cookie') {
    return addMethod === 'oauth'
      ? '/admin/accounts/cookie-auth'
      : '/admin/accounts/setup-token-cookie-auth'
  }

  return addMethod === 'oauth'
    ? '/admin/accounts/exchange-code'
    : '/admin/accounts/exchange-setup-token-code'
}
