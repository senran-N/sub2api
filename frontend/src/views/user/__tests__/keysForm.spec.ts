import { describe, expect, it } from 'vitest'
import {
  applyUserKeyExpirationPreset,
  buildDefaultUserKeyFormData,
  buildEditUserKeyFormData,
  buildUserKeyExpirationPayload,
  buildUserKeyGroupOptions,
  buildUserKeyRateLimitPayload,
  filterUserKeyGroupOptions,
  formatDateTimeLocal,
  parseUserKeyIpList,
  resolveUserKeyQuotaValue
} from '../keys/keysForm'

describe('keysForm helpers', () => {
  it('builds default and edit form data', () => {
    expect(buildDefaultUserKeyFormData()).toEqual(
      expect.objectContaining({
        name: '',
        group_id: null,
        status: 'active',
        expiration_preset: '30'
      })
    )

    expect(
      buildEditUserKeyFormData({
        name: 'Demo',
        group_id: 3,
        status: 'quota_exhausted',
        ip_whitelist: ['1.1.1.1'],
        ip_blacklist: [],
        quota: 12,
        rate_limit_5h: 1,
        rate_limit_1d: 0,
        rate_limit_7d: 0,
        expires_at: '2026-04-05T10:30:00Z'
      } as any)
    ).toEqual(
      expect.objectContaining({
        name: 'Demo',
        group_id: 3,
        status: 'inactive',
        enable_ip_restriction: true,
        enable_quota: true,
        enable_rate_limit: true,
        enable_expiration: true
      })
    )
  })

  it('builds group options and filters them', () => {
    const options = buildUserKeyGroupOptions(
      [
        {
          id: 1,
          name: 'Claude',
          description: 'Anthropic',
          rate_multiplier: 1,
          subscription_type: 'quota',
          platform: 'anthropic'
        }
      ] as any,
      { 1: 1.5 }
    )

    expect(options[0]).toEqual({
      value: 1,
      label: 'Claude',
      description: 'Anthropic',
      rate: 1,
      userRate: 1.5,
      subscriptionType: 'quota',
      platform: 'anthropic'
    })
    expect(filterUserKeyGroupOptions(options, 'clau')).toHaveLength(1)
    expect(filterUserKeyGroupOptions(options, 'openai')).toHaveLength(0)
  })

  it('builds quota, ip, expiration and rate limit payloads', () => {
    expect(parseUserKeyIpList('1.1.1.1\n\n2.2.2.2')).toEqual(['1.1.1.1', '2.2.2.2'])
    expect(resolveUserKeyQuotaValue(null)).toBe(0)
    expect(
      buildUserKeyRateLimitPayload({
        ...buildDefaultUserKeyFormData(),
        enable_rate_limit: true,
        rate_limit_5h: 1,
        rate_limit_1d: null,
        rate_limit_7d: 3
      })
    ).toEqual({
      rate_limit_5h: 1,
      rate_limit_1d: 0,
      rate_limit_7d: 3
    })
    expect(
      buildUserKeyExpirationPayload(
        {
          ...buildDefaultUserKeyFormData(),
          enable_expiration: true,
          expiration_date: '2026-04-07T10:30'
        },
        false,
        new Date('2026-04-05T10:30:00Z')
      )
    ).toEqual({ expiresInDays: 2 })
    expect(
      buildUserKeyExpirationPayload(
        {
          ...buildDefaultUserKeyFormData(),
          enable_expiration: false,
          expiration_date: ''
        },
        true
      )
    ).toEqual({ expiresAt: '' })
  })

  it('applies expiration presets and formats local datetime', () => {
    expect(formatDateTimeLocal('2026-04-05T10:30:00Z')).toMatch(/^2026-04-05T\d{2}:30$/)
    const nextForm = applyUserKeyExpirationPreset(
      buildDefaultUserKeyFormData(),
      7,
      new Date('2026-04-05T10:30:00Z')
    )

    expect(nextForm.expiration_preset).toBe('7')
    expect(nextForm.expiration_date).toMatch(/^2026-04-12T\d{2}:30$/)
  })
})
