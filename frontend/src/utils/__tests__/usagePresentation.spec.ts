import { describe, expect, it } from 'vitest'
import {
  getUsageRequestTypeBadgeClass,
  getUsageRequestTypeExportText,
  getUsageRequestTypeLabel,
  getUsageRequestTypeLabelKey
} from '../usagePresentation'

describe('usagePresentation helpers', () => {
  it('maps request types to translation keys and labels', () => {
    const t = (key: string) => `label:${key}`

    expect(getUsageRequestTypeLabelKey({ request_type: 'ws_v2' })).toBe('usage.ws')
    expect(getUsageRequestTypeLabelKey({ request_type: 'stream' })).toBe('usage.stream')
    expect(getUsageRequestTypeLabelKey({ request_type: 'sync' })).toBe('usage.sync')
    expect(getUsageRequestTypeLabelKey({ request_type: 'unknown' })).toBe('usage.unknown')
    expect(getUsageRequestTypeLabel({ request_type: 'stream' }, t)).toBe('label:usage.stream')
    expect(getUsageRequestTypeExportText({ request_type: 'stream' })).toBe('Stream')
    expect(getUsageRequestTypeExportText({ request_type: 'ws_v2' })).toBe('WS')
  })

  it('derives legacy request types and status badge classes', () => {
    expect(getUsageRequestTypeLabelKey({ stream: true })).toBe('usage.stream')
    expect(getUsageRequestTypeLabelKey({ stream: false })).toBe('usage.sync')
    expect(getUsageRequestTypeLabelKey({ openai_ws_mode: true })).toBe('usage.ws')
    expect(getUsageRequestTypeBadgeClass({ request_type: 'stream' })).toContain('bg-blue-100')
    expect(getUsageRequestTypeBadgeClass({ request_type: 'sync' })).toContain('bg-gray-100')
    expect(getUsageRequestTypeBadgeClass({ request_type: 'ws_v2' })).toContain('bg-violet-100')
    expect(getUsageRequestTypeBadgeClass({ request_type: 'unknown' })).toContain('bg-amber-100')
  })
})
