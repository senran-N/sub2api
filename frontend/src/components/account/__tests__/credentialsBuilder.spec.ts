import { describe, it, expect } from 'vitest'
import {
  applyInterceptWarmup,
  applyTempUnschedConfig,
  assignBuiltModelMapping,
  createTempUnschedRule,
  getDefaultBaseURL,
  loadTempUnschedRuleState,
  normalizePoolModeRetryCount,
  replaceAntigravityModelMapping
} from '../credentialsBuilder'

describe('applyInterceptWarmup', () => {
  it('create + enabled=true: should set intercept_warmup_requests to true', () => {
    const creds: Record<string, unknown> = { access_token: 'tok' }
    applyInterceptWarmup(creds, true, 'create')
    expect(creds.intercept_warmup_requests).toBe(true)
  })

  it('create + enabled=false: should not add the field', () => {
    const creds: Record<string, unknown> = { access_token: 'tok' }
    applyInterceptWarmup(creds, false, 'create')
    expect('intercept_warmup_requests' in creds).toBe(false)
  })

  it('edit + enabled=true: should set intercept_warmup_requests to true', () => {
    const creds: Record<string, unknown> = { api_key: 'sk' }
    applyInterceptWarmup(creds, true, 'edit')
    expect(creds.intercept_warmup_requests).toBe(true)
  })

  it('edit + enabled=false + field exists: should delete the field', () => {
    const creds: Record<string, unknown> = { api_key: 'sk', intercept_warmup_requests: true }
    applyInterceptWarmup(creds, false, 'edit')
    expect('intercept_warmup_requests' in creds).toBe(false)
  })

  it('edit + enabled=false + field absent: should not throw', () => {
    const creds: Record<string, unknown> = { api_key: 'sk' }
    applyInterceptWarmup(creds, false, 'edit')
    expect('intercept_warmup_requests' in creds).toBe(false)
  })

  it('should not affect other fields', () => {
    const creds: Record<string, unknown> = {
      api_key: 'sk',
      base_url: 'url',
      intercept_warmup_requests: true
    }
    applyInterceptWarmup(creds, false, 'edit')
    expect(creds.api_key).toBe('sk')
    expect(creds.base_url).toBe('url')
    expect('intercept_warmup_requests' in creds).toBe(false)
  })
})

describe('getDefaultBaseURL', () => {
  it('returns the correct upstream base URL per platform', () => {
    expect(getDefaultBaseURL('anthropic')).toBe('https://api.anthropic.com')
    expect(getDefaultBaseURL('openai')).toBe('https://api.openai.com')
    expect(getDefaultBaseURL('gemini')).toBe('https://generativelanguage.googleapis.com')
    expect(getDefaultBaseURL('antigravity')).toBe('https://api.anthropic.com')
  })
})

describe('normalizePoolModeRetryCount', () => {
  it('clamps invalid and out-of-range values', () => {
    expect(normalizePoolModeRetryCount(Number.NaN)).toBe(3)
    expect(normalizePoolModeRetryCount(-2)).toBe(0)
    expect(normalizePoolModeRetryCount(99)).toBe(10)
    expect(normalizePoolModeRetryCount(4.8)).toBe(4)
  })
})

describe('model mapping helpers', () => {
  it('assignBuiltModelMapping writes whitelist mappings as identity pairs', () => {
    const creds: Record<string, unknown> = {}
    assignBuiltModelMapping(creds, 'whitelist', ['gpt-5.2', 'gpt-5.4'], [])
    expect(creds.model_mapping).toEqual({
      'gpt-5.2': 'gpt-5.2',
      'gpt-5.4': 'gpt-5.4'
    })
  })

  it('replaceAntigravityModelMapping clears legacy whitelist before writing mapping', () => {
    const creds: Record<string, unknown> = {
      model_whitelist: ['legacy'],
      model_mapping: {
        legacy: 'legacy'
      }
    }
    replaceAntigravityModelMapping(creds, [{ from: 'claude-*', to: 'claude-sonnet-4-5' }])
    expect('model_whitelist' in creds).toBe(false)
    expect(creds.model_mapping).toEqual({
      'claude-*': 'claude-sonnet-4-5'
    })
  })
})

describe('temp unsched helpers', () => {
  it('createTempUnschedRule returns defaults and clones presets', () => {
    expect(createTempUnschedRule()).toEqual({
      error_code: null,
      keywords: '',
      duration_minutes: 30,
      description: ''
    })

    const preset = {
      error_code: 429,
      keywords: 'quota',
      duration_minutes: 15,
      description: 'rate limit'
    }
    const next = createTempUnschedRule(preset)
    expect(next).toEqual(preset)
    expect(next).not.toBe(preset)
  })

  it('applyTempUnschedConfig writes normalized rules and rejects invalid payloads', () => {
    const validCreds: Record<string, unknown> = {}
    const ok = applyTempUnschedConfig(validCreds, true, [
      {
        error_code: 429,
        keywords: 'quota, exhausted',
        duration_minutes: 15,
        description: 'rate limit'
      }
    ])
    expect(ok).toBe(true)
    expect(validCreds.temp_unschedulable_enabled).toBe(true)
    expect(validCreds.temp_unschedulable_rules).toEqual([
      {
        error_code: 429,
        keywords: ['quota', 'exhausted'],
        duration_minutes: 15,
        description: 'rate limit'
      }
    ])

    const invalidCreds: Record<string, unknown> = {}
    expect(applyTempUnschedConfig(invalidCreds, true, [createTempUnschedRule()])).toBe(false)
    expect('temp_unschedulable_rules' in invalidCreds).toBe(false)
  })

  it('loadTempUnschedRuleState restores UI state from credential payload', () => {
    const state = loadTempUnschedRuleState({
      temp_unschedulable_enabled: true,
      temp_unschedulable_rules: [
        {
          error_code: 429,
          keywords: ['quota', 'burst'],
          duration_minutes: 30,
          description: 'protect upstream'
        }
      ]
    })

    expect(state).toEqual({
      enabled: true,
      rules: [
        {
          error_code: 429,
          keywords: 'quota, burst',
          duration_minutes: 30,
          description: 'protect upstream'
        }
      ]
    })
  })
})
