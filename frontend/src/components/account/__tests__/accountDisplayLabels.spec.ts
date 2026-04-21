import { describe, expect, it } from 'vitest'
import { getAccountStatusLabel, getAccountTypeLabel } from '../accountDisplayLabels'

const translations: Record<string, string> = {
  'admin.accounts.types.oauth': 'OAuth',
  'admin.accounts.types.session': 'Session',
  'admin.accounts.types.upstream': '对接上游',
  'admin.accounts.apiKey': 'API Key',
  'admin.accounts.setupToken': 'Setup Token',
  'admin.accounts.bedrockLabel': 'AWS Bedrock',
  'admin.accounts.status.active': '正常',
  'admin.accounts.status.inactive': '停用',
  'admin.accounts.status.error': '错误',
  'admin.accounts.status.cooldown': '冷却中',
  'admin.accounts.status.paused': '暂停',
  'admin.accounts.status.limited': '限流',
  'admin.accounts.status.rateLimited': '限流中',
  'admin.accounts.status.overloaded': '过载中',
  'admin.accounts.status.tempUnschedulable': '临时不可调度'
}

const t = (key: string) => translations[key] ?? key

describe('accountDisplayLabels', () => {
  it('maps account types to localized labels', () => {
    expect(getAccountTypeLabel('session', t)).toBe('Session')
    expect(getAccountTypeLabel('apikey', t)).toBe('API Key')
    expect(getAccountTypeLabel('setup-token', t)).toBe('Setup Token')
    expect(getAccountTypeLabel('bedrock', t)).toBe('AWS Bedrock')
    expect(getAccountTypeLabel('unknown', t)).toBe('unknown')
  })

  it('maps account statuses to localized labels', () => {
    expect(getAccountStatusLabel('active', t)).toBe('正常')
    expect(getAccountStatusLabel('rate_limited', t)).toBe('限流中')
    expect(getAccountStatusLabel('temp_unschedulable', t)).toBe('临时不可调度')
    expect(getAccountStatusLabel('mystery', t)).toBe('mystery')
  })
})
