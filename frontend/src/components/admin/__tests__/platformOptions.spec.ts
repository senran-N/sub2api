import { describe, expect, it } from 'vitest'
import { buildAdminAccountTypeOptions, buildAdminPlatformOptions } from '../platformOptions'

const t = (key: string) => key

describe('admin platform options helpers', () => {
  it('builds the full shared platform list with Grok in the canonical order', () => {
    expect(buildAdminPlatformOptions(t)).toEqual([
      { value: 'anthropic', label: 'admin.accounts.platforms.anthropic' },
      { value: 'openai', label: 'admin.accounts.platforms.openai' },
      { value: 'gemini', label: 'admin.accounts.platforms.gemini' },
      { value: 'grok', label: 'admin.accounts.platforms.grok' },
      { value: 'antigravity', label: 'admin.accounts.platforms.antigravity' }
    ])
  })

  it('supports all-label and group translation namespaces without dropping Grok', () => {
    expect(buildAdminPlatformOptions(t, {
      allLabel: 'admin.groups.allPlatforms',
      labelPrefix: 'admin.groups.platforms'
    })).toEqual([
      { value: '', label: 'admin.groups.allPlatforms' },
      { value: 'anthropic', label: 'admin.groups.platforms.anthropic' },
      { value: 'openai', label: 'admin.groups.platforms.openai' },
      { value: 'gemini', label: 'admin.groups.platforms.gemini' },
      { value: 'grok', label: 'admin.groups.platforms.grok' },
      { value: 'antigravity', label: 'admin.groups.platforms.antigravity' }
    ])
  })

  it('builds account type filters that include Grok upstream and session modes', () => {
    expect(buildAdminAccountTypeOptions(t)).toEqual([
      { value: '', label: 'admin.accounts.allTypes' },
      { value: 'oauth', label: 'admin.accounts.oauthType' },
      { value: 'setup-token', label: 'admin.accounts.setupToken' },
      { value: 'apikey', label: 'admin.accounts.apiKey' },
      { value: 'upstream', label: 'admin.accounts.types.upstream' },
      { value: 'session', label: 'admin.accounts.types.session' },
      { value: 'bedrock', label: 'admin.accounts.bedrockLabel' }
    ])
  })
})
