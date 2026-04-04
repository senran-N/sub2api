import type {
  CreateSoraS3ProfileRequest,
  SoraS3Profile,
  TestSoraS3ConnectionRequest,
  UpdateSoraS3ProfileRequest
} from '@/api/admin/settings'

const GIGABYTE = 1024 * 1024 * 1024

export interface SoraS3ProfileForm {
  profile_id: string
  name: string
  set_active: boolean
  enabled: boolean
  endpoint: string
  region: string
  bucket: string
  access_key_id: string
  secret_access_key: string
  secret_access_key_configured: boolean
  prefix: string
  force_path_style: boolean
  cdn_url: string
  default_storage_quota_gb: number
}

export function formatStorageQuotaGB(bytes: number): string {
  if (!bytes || bytes <= 0) {
    return '0 GB'
  }

  const gb = bytes / GIGABYTE
  return `${gb.toFixed(gb >= 10 ? 0 : 1)} GB`
}

export function formatDataManagementDate(value?: string): string {
  if (!value) {
    return '-'
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  return date.toLocaleString()
}

export function createDefaultSoraS3ProfileForm(
  profile?: SoraS3Profile
): SoraS3ProfileForm {
  if (!profile) {
    return {
      profile_id: '',
      name: '',
      set_active: false,
      enabled: false,
      endpoint: '',
      region: '',
      bucket: '',
      access_key_id: '',
      secret_access_key: '',
      secret_access_key_configured: false,
      prefix: 'sora/',
      force_path_style: false,
      cdn_url: '',
      default_storage_quota_gb: 0
    }
  }

  const quotaBytes = profile.default_storage_quota_bytes || 0
  return {
    profile_id: profile.profile_id,
    name: profile.name,
    set_active: false,
    enabled: profile.enabled,
    endpoint: profile.endpoint || '',
    region: profile.region || '',
    bucket: profile.bucket || '',
    access_key_id: profile.access_key_id || '',
    secret_access_key: '',
    secret_access_key_configured: Boolean(profile.secret_access_key_configured),
    prefix: profile.prefix || '',
    force_path_style: Boolean(profile.force_path_style),
    cdn_url: profile.cdn_url || '',
    default_storage_quota_gb: Number((quotaBytes / GIGABYTE).toFixed(2))
  }
}

export function getPreferredSoraProfileID(profiles: SoraS3Profile[]): string {
  const active = profiles.find((profile) => profile.is_active)
  if (active) {
    return active.profile_id
  }
  return profiles[0]?.profile_id || ''
}

export function buildSoraS3ProfileBasePayload(form: SoraS3ProfileForm): Omit<
  CreateSoraS3ProfileRequest,
  'profile_id' | 'name' | 'set_active'
> & { name: string } {
  return {
    name: form.name.trim(),
    enabled: form.enabled,
    endpoint: form.endpoint,
    region: form.region,
    bucket: form.bucket,
    access_key_id: form.access_key_id,
    secret_access_key: form.secret_access_key || undefined,
    prefix: form.prefix,
    force_path_style: form.force_path_style,
    cdn_url: form.cdn_url,
    default_storage_quota_bytes: Math.round((form.default_storage_quota_gb || 0) * GIGABYTE)
  }
}

export function buildCreateSoraS3ProfileRequest(
  form: SoraS3ProfileForm
): CreateSoraS3ProfileRequest {
  return {
    profile_id: form.profile_id.trim(),
    set_active: form.set_active,
    ...buildSoraS3ProfileBasePayload(form)
  }
}

export function buildUpdateSoraS3ProfileRequest(
  form: SoraS3ProfileForm
): UpdateSoraS3ProfileRequest {
  return buildSoraS3ProfileBasePayload(form)
}

export function buildTestSoraS3ConnectionRequest(
  form: SoraS3ProfileForm,
  profileID?: string
): TestSoraS3ConnectionRequest {
  return {
    profile_id: profileID,
    enabled: form.enabled,
    endpoint: form.endpoint,
    region: form.region,
    bucket: form.bucket,
    access_key_id: form.access_key_id,
    secret_access_key: form.secret_access_key || undefined,
    prefix: form.prefix,
    force_path_style: form.force_path_style,
    cdn_url: form.cdn_url,
    default_storage_quota_bytes: Math.round((form.default_storage_quota_gb || 0) * GIGABYTE)
  }
}

export function validateSoraS3ProfileForm(
  form: SoraS3ProfileForm,
  options: {
    creating: boolean
    selectedProfileID: string
  }
): string | null {
  if (!form.name.trim()) {
    return 'admin.settings.soraS3.profileNameRequired'
  }
  if (options.creating && !form.profile_id.trim()) {
    return 'admin.settings.soraS3.profileIDRequired'
  }
  if (!options.creating && !options.selectedProfileID) {
    return 'admin.settings.soraS3.profileSelectRequired'
  }
  if (form.enabled) {
    if (!form.endpoint.trim()) {
      return 'admin.settings.soraS3.endpointRequired'
    }
    if (!form.bucket.trim()) {
      return 'admin.settings.soraS3.bucketRequired'
    }
    if (!form.access_key_id.trim()) {
      return 'admin.settings.soraS3.accessKeyRequired'
    }
  }
  return null
}
