import { describe, expect, it, vi } from 'vitest'
import type { CustomMenuItem } from '@/types'
import {
  isEmbeddedPageUrl,
  loadEmbeddedPageSettings,
  resolveCustomPageMenuItem
} from '../embedded/embeddedPageFrame'

const createMenuItem = (overrides: Partial<CustomMenuItem>): CustomMenuItem => ({
  id: 'item-id',
  label: 'Item',
  icon_svg: '',
  url: 'https://example.com',
  visibility: 'user',
  sort_order: 1,
  ...overrides
})

describe('embeddedPageFrame helpers', () => {
  it('accepts only http and https embedded urls', () => {
    expect(isEmbeddedPageUrl('https://example.com')).toBe(true)
    expect(isEmbeddedPageUrl('http://example.com')).toBe(true)
    expect(isEmbeddedPageUrl('javascript:alert(1)')).toBe(false)
    expect(isEmbeddedPageUrl('/relative/path')).toBe(false)
  })

  it('resolves custom menu items from public settings before admin fallbacks', () => {
    const publicItem = createMenuItem({ id: 'public-item', label: 'Public' })
    const adminItem = createMenuItem({
      id: 'admin-item',
      label: 'Admin',
      visibility: 'admin'
    })

    expect(resolveCustomPageMenuItem('public-item', [publicItem], [adminItem], false)).toBe(
      publicItem
    )
    expect(resolveCustomPageMenuItem('admin-item', [publicItem], [adminItem], false)).toBeNull()
    expect(resolveCustomPageMenuItem('admin-item', [publicItem], [adminItem], true)).toBe(adminItem)
  })

  it('loads public settings only when they are missing', async () => {
    const fetchPublicSettings = vi.fn().mockResolvedValue(null)
    const loadingStates: boolean[] = []

    await loadEmbeddedPageSettings(
      {
        publicSettingsLoaded: false,
        fetchPublicSettings
      },
      (loading) => {
        loadingStates.push(loading)
      }
    )

    expect(fetchPublicSettings).toHaveBeenCalledOnce()
    expect(loadingStates).toEqual([true, false])

    const skippedFetch = vi.fn()
    await loadEmbeddedPageSettings(
      {
        publicSettingsLoaded: true,
        fetchPublicSettings: skippedFetch
      },
      (loading) => {
        loadingStates.push(loading)
      }
    )

    expect(skippedFetch).not.toHaveBeenCalled()
  })
})
