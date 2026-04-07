import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { SoraS3Profile } from '@/api/admin/settings'
import { useDataManagementSoraProfiles } from '../useSoraProfiles'

const {
  listSoraS3Profiles,
  createSoraS3Profile,
  updateSoraS3Profile,
  testSoraS3Connection,
  setActiveSoraS3Profile,
  deleteSoraS3Profile
} = vi.hoisted(() => ({
  listSoraS3Profiles: vi.fn(),
  createSoraS3Profile: vi.fn(),
  updateSoraS3Profile: vi.fn(),
  testSoraS3Connection: vi.fn(),
  setActiveSoraS3Profile: vi.fn(),
  deleteSoraS3Profile: vi.fn()
}))

vi.mock('@/api', () => ({
  adminAPI: {
    settings: {
      listSoraS3Profiles,
      createSoraS3Profile,
      updateSoraS3Profile,
      testSoraS3Connection,
      setActiveSoraS3Profile,
      deleteSoraS3Profile
    }
  }
}))

function createProfile(overrides: Partial<SoraS3Profile> = {}): SoraS3Profile {
  return {
    profile_id: 'main',
    name: 'Main',
    is_active: true,
    enabled: true,
    endpoint: 'https://example.com',
    region: 'auto',
    bucket: 'bucket',
    access_key_id: 'AK',
    secret_access_key_configured: true,
    prefix: 'sora/',
    force_path_style: false,
    cdn_url: '',
    default_storage_quota_bytes: 5 * 1024 * 1024 * 1024,
    updated_at: '2026-04-04T00:00:00Z',
    ...overrides
  }
}

describe('useDataManagementSoraProfiles', () => {
  beforeEach(() => {
    listSoraS3Profiles.mockReset()
    createSoraS3Profile.mockReset()
    updateSoraS3Profile.mockReset()
    testSoraS3Connection.mockReset()
    setActiveSoraS3Profile.mockReset()
    deleteSoraS3Profile.mockReset()

    listSoraS3Profiles.mockResolvedValue({
      active_profile_id: 'main',
      items: [createProfile(), createProfile({ profile_id: 'secondary', is_active: false })]
    })
    createSoraS3Profile.mockResolvedValue(createProfile({ profile_id: 'new-profile' }))
    updateSoraS3Profile.mockResolvedValue(createProfile({ name: 'Updated' }))
    testSoraS3Connection.mockResolvedValue({ message: 'ok' })
    setActiveSoraS3Profile.mockResolvedValue(createProfile({ profile_id: 'secondary', is_active: true }))
    deleteSoraS3Profile.mockResolvedValue(undefined)
  })

  it('loads profiles and syncs edit state', async () => {
    const composable = useDataManagementSoraProfiles({
      t: (key: string) => key,
      showError: vi.fn(),
      showSuccess: vi.fn(),
      confirm: vi.fn(() => true)
    })

    await composable.loadSoraS3Profiles()
    expect(composable.soraS3Profiles.value).toHaveLength(2)

    composable.editSoraProfile('secondary')
    expect(composable.creatingSoraProfile.value).toBe(false)
    expect(composable.soraProfileDrawerOpen.value).toBe(true)
    expect(composable.soraProfileForm.value.profile_id).toBe('secondary')
  })

  it('creates, updates, and tests profiles', async () => {
    const showSuccess = vi.fn()
    const showError = vi.fn()
    const composable = useDataManagementSoraProfiles({
      t: (key: string) => key,
      showError,
      showSuccess,
      confirm: vi.fn(() => true)
    })

    await composable.loadSoraS3Profiles()
    composable.startCreateSoraProfile()
    composable.soraProfileForm.value.profile_id = 'new-profile'
    composable.soraProfileForm.value.name = 'New Profile'
    await composable.saveSoraProfile()
    expect(createSoraS3Profile).toHaveBeenCalledWith(
      expect.objectContaining({
        profile_id: 'new-profile',
        name: 'New Profile'
      })
    )
    expect(showSuccess).toHaveBeenCalledWith('admin.settings.soraS3.profileCreated')

    composable.editSoraProfile('main')
    composable.soraProfileForm.value.name = 'Updated'
    await composable.saveSoraProfile()
    expect(updateSoraS3Profile).toHaveBeenCalledWith(
      'main',
      expect.objectContaining({
        name: 'Updated'
      })
    )

    await composable.testSoraProfileConnection()
    expect(testSoraS3Connection).toHaveBeenCalledTimes(1)
    expect(showSuccess).toHaveBeenCalledWith('ok')
    expect(showError).not.toHaveBeenCalled()
  })

  it('activates, deletes, and resets create drawer state on close', async () => {
    const showSuccess = vi.fn()
    const confirm = vi.fn(() => true)
    const composable = useDataManagementSoraProfiles({
      t: (key: string) => key,
      showError: vi.fn(),
      showSuccess,
      confirm
    })

    await composable.loadSoraS3Profiles()
    await composable.activateSoraProfile('secondary')
    expect(setActiveSoraS3Profile).toHaveBeenCalledWith('secondary')
    expect(showSuccess).toHaveBeenCalledWith('admin.settings.soraS3.profileActivated')

    await composable.removeSoraProfile('secondary')
    expect(confirm).toHaveBeenCalledWith('admin.settings.soraS3.deleteConfirm')
    expect(deleteSoraS3Profile).toHaveBeenCalledWith('secondary')
    expect(showSuccess).toHaveBeenCalledWith('admin.settings.soraS3.profileDeleted')

    composable.startCreateSoraProfile()
    composable.soraProfileForm.value.profile_id = 'dirty'
    composable.closeSoraProfileDrawer()
    expect(composable.creatingSoraProfile.value).toBe(false)
    expect(composable.soraProfileDrawerOpen.value).toBe(false)
    expect(composable.soraProfileForm.value.profile_id).toBe('main')
  })

  it('uses shared request error details for failed profile loads', async () => {
    const showError = vi.fn()
    const composable = useDataManagementSoraProfiles({
      t: (key: string) => key,
      showError,
      showSuccess: vi.fn(),
      confirm: vi.fn(() => true)
    })
    listSoraS3Profiles.mockRejectedValueOnce({
      response: { data: { detail: 'profile-load-failed' } }
    })

    await composable.loadSoraS3Profiles()

    expect(showError).toHaveBeenCalledWith('profile-load-failed')
  })
})
