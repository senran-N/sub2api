import { describe, expect, it } from 'vitest'
import {
  buildHomeFeatureTags,
  buildHomeFeatures,
  buildHomeProviders,
  resolveHomeContentUrl,
  resolveHomeDashboardPath,
  resolveHomeUserInitial
} from '../home/homeView'

const t = (key: string) => key

describe('homeView', () => {
  it('detects when custom home content should render as an iframe', () => {
    expect(resolveHomeContentUrl('https://example.com/embed')).toBe(true)
    expect(resolveHomeContentUrl('http://example.com')).toBe(true)
    expect(resolveHomeContentUrl('<div>inline</div>')).toBe(false)
  })

  it('resolves dashboard path and user initials from auth state', () => {
    expect(resolveHomeDashboardPath(true)).toBe('/admin/dashboard')
    expect(resolveHomeDashboardPath(false)).toBe('/dashboard')
    expect(resolveHomeUserInitial('user@example.com')).toBe('U')
    expect(resolveHomeUserInitial('')).toBe('')
  })

  it('builds translated home feature tags, cards, and provider badges', () => {
    const tags = buildHomeFeatureTags(t)
    const features = buildHomeFeatures(t)
    const providers = buildHomeProviders(t)

    expect(tags.map((tag) => tag.key)).toEqual([
      'subscription-to-api',
      'sticky-session',
      'realtime-billing'
    ])

    expect(features.map((feature) => feature.key)).toEqual([
      'unified-gateway',
      'multi-account',
      'balance-quota'
    ])

    expect(providers.map((provider) => provider.key)).toEqual([
      'claude',
      'gpt',
      'gemini',
      'antigravity',
      'more'
    ])
    expect(providers.at(-1)?.supported).toBe(false)
  })
})
