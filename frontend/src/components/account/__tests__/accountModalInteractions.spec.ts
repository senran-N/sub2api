import { describe, expect, it, vi } from 'vitest'
import {
  appendEmptyModelMapping,
  appendPresetModelMapping,
  applySharedAccountCredentialsState,
  applyTempUnschedCredentialsState,
  confirmCustomErrorCodeSelection,
  removeModelMappingAt
} from '../accountModalInteractions'

const t = (key: string) => key

describe('accountModalInteractions', () => {
  it('applies shared edit credential state and validates temp unsched rules', () => {
    const credentials: Record<string, unknown> = { api_key: 'sk-test' }
    const showError = vi.fn()

    expect(
      applySharedAccountCredentialsState(credentials, {
        interceptWarmupRequests: true,
        tempUnschedEnabled: true,
        tempUnschedRules: [
          {
            error_code: 429,
            keywords: 'quota, exhausted',
            duration_minutes: 15,
            description: 'protect upstream'
          }
        ],
        showError,
        t
      })
    ).toBe(true)

    expect(showError).not.toHaveBeenCalled()
    expect(credentials.intercept_warmup_requests).toBe(true)
    expect(credentials.temp_unschedulable_enabled).toBe(true)
    expect(credentials.temp_unschedulable_rules).toEqual([
      {
        error_code: 429,
        keywords: ['quota', 'exhausted'],
        duration_minutes: 15,
        description: 'protect upstream'
      }
    ])
  })

  it('reports invalid temp unsched state without mutating payload further', () => {
    const credentials: Record<string, unknown> = { api_key: 'sk-test' }
    const showError = vi.fn()

    expect(
      applySharedAccountCredentialsState(credentials, {
        interceptWarmupRequests: false,
        tempUnschedEnabled: true,
        tempUnschedRules: [
          {
            error_code: null,
            keywords: '',
            duration_minutes: 30,
            description: ''
          }
        ],
        showError,
        t
      })
    ).toBe(false)

    expect(showError).toHaveBeenCalledWith('admin.accounts.tempUnschedulable.rulesInvalid')
    expect('temp_unschedulable_rules' in credentials).toBe(false)
    expect('intercept_warmup_requests' in credentials).toBe(false)
  })

  it('applies temp unsched state without touching intercept flags', () => {
    const credentials: Record<string, unknown> = { api_key: 'sk-test' }
    const showError = vi.fn()

    expect(
      applyTempUnschedCredentialsState(credentials, {
        tempUnschedEnabled: true,
        tempUnschedRules: [
          {
            error_code: 500,
            keywords: 'retry later',
            duration_minutes: 20,
            description: 'server instability'
          }
        ],
        showError,
        t
      })
    ).toBe(true)

    expect(showError).not.toHaveBeenCalled()
    expect('intercept_warmup_requests' in credentials).toBe(false)
    expect(credentials.temp_unschedulable_enabled).toBe(true)
  })

  it('adds and removes model mappings through shared helpers', () => {
    const mappings = [{ from: 'gpt-4.1', to: 'gpt-4.1' }]

    appendEmptyModelMapping(mappings)
    expect(mappings).toEqual([
      { from: 'gpt-4.1', to: 'gpt-4.1' },
      { from: '', to: '' }
    ])

    removeModelMappingAt(mappings, 1)
    expect(mappings).toEqual([{ from: 'gpt-4.1', to: 'gpt-4.1' }])
  })

  it('guards duplicate preset mappings before appending', () => {
    const mappings = [{ from: 'gpt-4.1', to: 'gpt-4.1' }]
    const onDuplicate = vi.fn()

    appendPresetModelMapping(mappings, 'gpt-4.1', 'gpt-4.1-mini', onDuplicate)
    expect(onDuplicate).toHaveBeenCalledWith('gpt-4.1')
    expect(mappings).toEqual([{ from: 'gpt-4.1', to: 'gpt-4.1' }])

    appendPresetModelMapping(mappings, 'gpt-5', 'gpt-5-mini', onDuplicate)
    expect(mappings).toEqual([
      { from: 'gpt-4.1', to: 'gpt-4.1' },
      { from: 'gpt-5', to: 'gpt-5-mini' }
    ])
  })

  it('confirms only warning error codes', () => {
    const confirmFn = vi.fn(() => true)

    expect(confirmCustomErrorCodeSelection(429, confirmFn, t)).toBe(true)
    expect(confirmCustomErrorCodeSelection(529, confirmFn, t)).toBe(true)
    expect(confirmCustomErrorCodeSelection(500, confirmFn, t)).toBe(true)

    expect(confirmFn).toHaveBeenNthCalledWith(1, 'admin.accounts.customErrorCodes429Warning')
    expect(confirmFn).toHaveBeenNthCalledWith(2, 'admin.accounts.customErrorCodes529Warning')
    expect(confirmFn).toHaveBeenCalledTimes(2)
  })
})
