import { describe, expect, it } from 'vitest'
import type { SoraS3Profile } from '@/api/admin/settings'
import {
  buildCreateSoraS3ProfileRequest,
  buildTestSoraS3ConnectionRequest,
  buildUpdateSoraS3ProfileRequest,
  createDefaultSoraS3ProfileForm,
  formatDataManagementDate,
  formatStorageQuotaGB,
  getPreferredSoraProfileID,
  validateSoraS3ProfileForm
} from '../dataManagementView'

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
    force_path_style: true,
    cdn_url: 'https://cdn.example.com',
    default_storage_quota_bytes: 5 * 1024 * 1024 * 1024,
    updated_at: '2026-04-04T00:00:00Z',
    ...overrides
  }
}

describe('dataManagementView helpers', () => {
  it('formats quota/date and selects preferred profile', () => {
    expect(formatStorageQuotaGB(0)).toBe('0 GB')
    expect(formatStorageQuotaGB(5 * 1024 * 1024 * 1024)).toBe('5.0 GB')
    expect(formatStorageQuotaGB(12 * 1024 * 1024 * 1024)).toBe('12 GB')
    expect(formatDataManagementDate()).toBe('-')
    expect(formatDataManagementDate('invalid')).toBe('invalid')

    expect(
      getPreferredSoraProfileID([
        createProfile({ profile_id: 'a', is_active: false }),
        createProfile({ profile_id: 'b', is_active: true })
      ])
    ).toBe('b')
  })

  it('hydrates default form state and builds payloads', () => {
    const form = createDefaultSoraS3ProfileForm(createProfile())
    expect(form).toEqual({
      profile_id: 'main',
      name: 'Main',
      set_active: false,
      enabled: true,
      endpoint: 'https://example.com',
      region: 'auto',
      bucket: 'bucket',
      access_key_id: 'AK',
      secret_access_key: '',
      secret_access_key_configured: true,
      prefix: 'sora/',
      force_path_style: true,
      cdn_url: 'https://cdn.example.com',
      default_storage_quota_gb: 5
    })

    form.secret_access_key = 'secret'
    form.set_active = true
    expect(buildCreateSoraS3ProfileRequest(form)).toEqual({
      profile_id: 'main',
      name: 'Main',
      set_active: true,
      enabled: true,
      endpoint: 'https://example.com',
      region: 'auto',
      bucket: 'bucket',
      access_key_id: 'AK',
      secret_access_key: 'secret',
      prefix: 'sora/',
      force_path_style: true,
      cdn_url: 'https://cdn.example.com',
      default_storage_quota_bytes: 5 * 1024 * 1024 * 1024
    })
    expect(buildUpdateSoraS3ProfileRequest(form)).toEqual({
      name: 'Main',
      enabled: true,
      endpoint: 'https://example.com',
      region: 'auto',
      bucket: 'bucket',
      access_key_id: 'AK',
      secret_access_key: 'secret',
      prefix: 'sora/',
      force_path_style: true,
      cdn_url: 'https://cdn.example.com',
      default_storage_quota_bytes: 5 * 1024 * 1024 * 1024
    })
    expect(buildTestSoraS3ConnectionRequest(form, 'main')).toEqual({
      profile_id: 'main',
      enabled: true,
      endpoint: 'https://example.com',
      region: 'auto',
      bucket: 'bucket',
      access_key_id: 'AK',
      secret_access_key: 'secret',
      prefix: 'sora/',
      force_path_style: true,
      cdn_url: 'https://cdn.example.com',
      default_storage_quota_bytes: 5 * 1024 * 1024 * 1024
    })
  })

  it('validates form requirements based on mode and enabled state', () => {
    const form = createDefaultSoraS3ProfileForm()
    expect(
      validateSoraS3ProfileForm(form, {
        creating: true,
        selectedProfileID: ''
      })
    ).toBe('admin.settings.soraS3.profileNameRequired')

    form.name = 'Profile'
    expect(
      validateSoraS3ProfileForm(form, {
        creating: true,
        selectedProfileID: ''
      })
    ).toBe('admin.settings.soraS3.profileIDRequired')

    form.profile_id = 'main'
    form.enabled = true
    expect(
      validateSoraS3ProfileForm(form, {
        creating: true,
        selectedProfileID: ''
      })
    ).toBe('admin.settings.soraS3.endpointRequired')

    form.endpoint = 'https://example.com'
    expect(
      validateSoraS3ProfileForm(form, {
        creating: true,
        selectedProfileID: ''
      })
    ).toBe('admin.settings.soraS3.bucketRequired')

    form.bucket = 'bucket'
    expect(
      validateSoraS3ProfileForm(form, {
        creating: true,
        selectedProfileID: ''
      })
    ).toBe('admin.settings.soraS3.accessKeyRequired')

    form.access_key_id = 'AK'
    expect(
      validateSoraS3ProfileForm(form, {
        creating: true,
        selectedProfileID: ''
      })
    ).toBeNull()
  })
})
