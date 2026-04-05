import { describe, expect, it, vi } from 'vitest'
import {
  formatProfileBalance,
  formatProfileMemberSince,
  loadProfilePublicSettings
} from '../profile/profileView'

describe('profileView helpers', () => {
  it('formats missing profile values safely', () => {
    expect(formatProfileBalance(undefined)).toBe('$0.00')
    expect(formatProfileMemberSince(undefined)).toBe('')
    expect(formatProfileMemberSince('invalid-date')).toBe('')
  })

  it('loads public settings only when profile settings are not cached', async () => {
    const fetchPublicSettings = vi.fn().mockResolvedValue(null)

    await loadProfilePublicSettings({
      publicSettingsLoaded: false,
      fetchPublicSettings
    })

    expect(fetchPublicSettings).toHaveBeenCalledOnce()

    const skippedFetch = vi.fn()
    await loadProfilePublicSettings({
      publicSettingsLoaded: true,
      fetchPublicSettings: skippedFetch
    })

    expect(skippedFetch).not.toHaveBeenCalled()
  })
})
