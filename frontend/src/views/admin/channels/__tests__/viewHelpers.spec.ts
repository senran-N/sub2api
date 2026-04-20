import { describe, expect, it } from 'vitest'
import {
  formatDate,
  getActionButtonClasses,
  getGroupChipClasses,
  getPlatformTextClass,
  getPlatformToggleClasses,
  getRateBadgeClass,
  platformOrder
} from '../viewHelpers'

describe('channels view helpers', () => {
  it('defines the supported platform order', () => {
    expect(platformOrder).toEqual(['anthropic', 'openai', 'gemini', 'grok', 'antigravity'])
  })

  it('returns tone and badge classes for supported platforms', () => {
    expect(getPlatformTextClass('openai')).toContain('channel-view__tone-text--success')
    expect(getPlatformTextClass('gemini')).toContain('channel-view__tone-text--info')
    expect(getPlatformTextClass('grok')).toContain('channel-view__tone-text--brand-rose')
    expect(getPlatformTextClass('unknown')).toBe('channel-view__text-muted')

    expect(getRateBadgeClass('anthropic')).toContain('theme-chip--brand-orange')
    expect(getRateBadgeClass('openai')).toContain('theme-chip--success')
    expect(getRateBadgeClass('grok')).toContain('theme-chip--brand-rose')
    expect(getRateBadgeClass('unknown')).toContain('theme-chip--neutral')
  })

  it('builds toggle, group, and action classes', () => {
    expect(getPlatformToggleClasses('openai', true)).toContain('channel-view__platform-toggle--active')
    expect(getGroupChipClasses('gemini', true, true)).toContain('channel-view__group-chip--selected')
    expect(getGroupChipClasses('gemini', true, true)).toContain('opacity-40')
    expect(getActionButtonClasses('info')).toContain('channel-view__action-button--info')
    expect(getActionButtonClasses('danger')).toContain('channel-view__action-button--danger')
  })

  it('formats empty and populated dates', () => {
    expect(formatDate('')).toBe('-')
    expect(formatDate('2026-04-01T00:00:00Z')).not.toBe('-')
  })
})
