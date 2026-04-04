import { describe, expect, it } from 'vitest'
import {
  createDefaultBetaPolicyRules,
  createDefaultOverloadCooldownSettings,
  createDefaultRectifierSettings,
  createDefaultSettingsForm,
  createDefaultStreamTimeoutSettings,
  getSettingsBetaDisplayName,
  getSettingsLinuxdoRedirectUrlSuggestion,
  maskSettingsApiKey,
  sanitizeRectifierPatterns
} from '../settingsView'

describe('settingsView helpers', () => {
  it('creates default settings and section state', () => {
    const form = createDefaultSettingsForm()
    expect(form.site_name).toBe('Sub2API')
    expect(form.smtp_port).toBe(587)
    expect(form.enable_fingerprint_unification).toBe(true)

    expect(createDefaultOverloadCooldownSettings()).toEqual({
      enabled: true,
      cooldown_minutes: 10
    })
    expect(createDefaultStreamTimeoutSettings()).toEqual({
      enabled: true,
      action: 'temp_unsched',
      temp_unsched_minutes: 5,
      threshold_count: 3,
      threshold_window_minutes: 10
    })
    expect(createDefaultRectifierSettings()).toEqual({
      enabled: true,
      thinking_signature_enabled: true,
      thinking_budget_enabled: true,
      apikey_signature_enabled: false,
      apikey_signature_patterns: []
    })
    expect(createDefaultBetaPolicyRules()).toEqual([])
  })

  it('normalizes helper values used by settings sections', () => {
    expect(maskSettingsApiKey('abcdefghijklmnopqrstuvwxyz')).toBe('abcdefghij...wxyz')
    expect(sanitizeRectifierPatterns(['  one  ', '', '  ', 'two'])).toEqual(['one', 'two'])
    expect(sanitizeRectifierPatterns(null)).toEqual([])
    expect(
      getSettingsLinuxdoRedirectUrlSuggestion({
        origin: 'https://sub2api.example.com',
        protocol: 'https:',
        host: 'sub2api.example.com'
      })
    ).toBe('https://sub2api.example.com/api/v1/auth/oauth/linuxdo/callback')
    expect(getSettingsLinuxdoRedirectUrlSuggestion(null)).toBe('')
    expect(getSettingsBetaDisplayName('fast-mode-2026-02-01')).toBe('Fast Mode')
    expect(getSettingsBetaDisplayName('unknown-token')).toBe('unknown-token')
  })
})
