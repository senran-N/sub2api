import type { AccountPlatform } from '@/types'

export function resolveOAuthKeyPrefix(platform: AccountPlatform) {
  if (platform === 'openai' || platform === 'sora') {
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
  return platform === 'openai' || platform === 'gemini' || platform === 'antigravity' || platform === 'sora'
}

export function resolveOAuthImportantNoticeKey(platform: AccountPlatform) {
  if (platform === 'openai' || platform === 'sora') {
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
