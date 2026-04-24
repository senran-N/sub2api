import type { AccountPlatform, AccountType } from '@/types'

export type AccountMutationSection =
  | 'anthropic-runtime'
  | 'antigravity-extra'
  | 'bedrock-credentials'
  | 'compatible-credentials'
  | 'grok-session'
  | 'model-restriction'
  | 'openai-runtime'
  | 'quota-limits'
  | 'temp-unsched'
  | 'warmup'

export interface AccountMutationProfile {
  platform: AccountPlatform
  type: AccountType
  sections: AccountMutationSection[]
}

export function resolveAccountMutationProfile(
  platform: AccountPlatform,
  type: AccountType
): AccountMutationProfile {
  const sections = new Set<AccountMutationSection>()

  if (isCompatibleCredentialAccount(platform, type)) {
    sections.add('compatible-credentials')
    sections.add('model-restriction')
    sections.add('quota-limits')
  }
  if (platform === 'anthropic' && type === 'bedrock') {
    sections.add('bedrock-credentials')
    sections.add('model-restriction')
    sections.add('quota-limits')
  }
  if (platform === 'grok' && type === 'session') {
    sections.add('grok-session')
  }
  if (platform === 'anthropic' && (type === 'oauth' || type === 'setup-token')) {
    sections.add('anthropic-runtime')
  }
  if (platform === 'anthropic' || platform === 'antigravity') {
    sections.add('warmup')
  }
  if (platform === 'openai' && (type === 'oauth' || type === 'apikey')) {
    sections.add('openai-runtime')
  }
  if (platform === 'openai' && type === 'oauth') {
    sections.add('model-restriction')
  }
  if (platform === 'antigravity') {
    sections.add('antigravity-extra')
    sections.add('model-restriction')
  }
  sections.add('temp-unsched')

  return {
    platform,
    type,
    sections: Array.from(sections)
  }
}

export function accountMutationProfileHasSection(
  profile: AccountMutationProfile | null | undefined,
  section: AccountMutationSection
): boolean {
  return profile?.sections.includes(section) === true
}

function isCompatibleCredentialAccount(
  platform: AccountPlatform,
  type: AccountType
): boolean {
  if (type === 'apikey') {
    return (
      platform === 'anthropic' ||
      platform === 'openai' ||
      platform === 'gemini' ||
      platform === 'grok'
    )
  }
  return platform === 'grok' && type === 'upstream'
}
