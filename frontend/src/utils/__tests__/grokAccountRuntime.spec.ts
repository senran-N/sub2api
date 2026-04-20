import { describe, expect, it } from 'vitest'
import type { Account } from '@/types'
import { getGrokAccountRuntime } from '../grokAccountRuntime'

function createAccount(overrides: Partial<Account> = {}): Account {
  return {
    id: 101,
    name: 'Grok Session',
    platform: 'grok',
    type: 'session',
    credentials: {},
    extra: {},
    proxy_id: null,
    concurrency: 1,
    priority: 1,
    status: 'active',
    error_message: null,
    last_used_at: null,
    expires_at: null,
    auto_pause_on_expired: false,
    created_at: '2026-04-20T00:00:00Z',
    updated_at: '2026-04-20T00:00:00Z',
    schedulable: true,
    rate_limited_at: null,
    rate_limit_reset_at: null,
    overload_until: null,
    temp_unschedulable_until: null,
    temp_unschedulable_reason: null,
    session_window_start: null,
    session_window_end: null,
    session_window_status: null,
    ...overrides
  }
}

describe('getGrokAccountRuntime', () => {
  it('returns null for non-grok accounts', () => {
    const runtime = getGrokAccountRuntime(createAccount({ platform: 'openai' }))
    expect(runtime).toBeNull()
  })

  it('normalizes Grok runtime state for admin surfaces', () => {
    const runtime = getGrokAccountRuntime(
      createAccount({
        extra: {
          grok: {
            auth_mode: 'session',
            auth_fingerprint: 'sha256:ab12...cd34',
            tier: {
              normalized: 'heavy',
              raw: 'Heavy',
              source: 'quota_sync',
              confidence: 0.92
            },
            capabilities: {
              operations: ['video', 'chat'],
              image: true,
              models: ['grok-4', 'grok-4-video']
            },
            quota_windows: {
              auto: {
                remaining: 17,
                total: 150,
                window_seconds: 7200,
                source: 'sync',
                reset_at: '2026-04-20T02:00:00Z'
              },
              heavy: {
                remaining: 3,
                total: 20,
                window_seconds: 7200,
                source: 'sync'
              }
            },
            sync_state: {
              last_sync_at: '2026-04-20T00:00:00Z',
              last_probe_at: '2026-04-20T01:00:00Z',
              last_probe_ok_at: '2026-04-20T00:45:00Z',
              last_probe_error_at: '2026-04-20T01:00:00Z',
              last_probe_error: 'API returned 401 Unauthorized',
              last_probe_status_code: 401
            },
            runtime_state: {
              last_request_at: '2026-04-20T01:05:00Z',
              last_request_capability: 'video',
              last_request_model: 'grok-4-video',
              last_request_upstream_model: 'grok-4-video',
              last_fail_at: '2026-04-20T01:05:00Z',
              last_fail_reason: 'video tier required',
              last_fail_status_code: 403,
              last_fail_class: 'model_unsupported',
              last_fail_scope: 'model',
              selection_cooldown_until: '2026-04-20T01:10:00Z',
              selection_cooldown_model: 'grok-4-video'
            }
          }
        } as any
      })
    )

    expect(runtime).not.toBeNull()
    expect(runtime?.hasState).toBe(true)
    expect(runtime?.authFingerprint).toBe('sha256:ab12...cd34')
    expect(runtime?.tier).toEqual({
      normalized: 'heavy',
      raw: 'Heavy',
      source: 'quota_sync',
      confidence: 0.92
    })
    expect(runtime?.capabilities.operations).toEqual(['chat', 'image', 'video'])
    expect(runtime?.capabilities.models).toEqual(['grok-4', 'grok-4-video'])
    expect(runtime?.quotaWindows.map((window) => window.name)).toEqual(['auto', 'heavy'])
    expect(runtime?.sync.lastProbeStatusCode).toBe(401)
    expect(runtime?.runtime.lastFailReason).toBe('video tier required')
    expect(runtime?.runtime.cooldownModel).toBe('grok-4-video')
  })

  it('infers the tier from quota windows when tier metadata is missing', () => {
    const runtime = getGrokAccountRuntime(
      createAccount({
        extra: {
          grok: {
            quota_windows: {
              auto: {
                remaining: 50,
                total: 150
              }
            }
          }
        } as any
      })
    )

    expect(runtime?.tier.normalized).toBe('heavy')
    expect(runtime?.hasState).toBe(true)
  })
})
