import type { AccountPlatform, GroupPlatform } from '@/types'

type AdminPlatform = AccountPlatform | GroupPlatform
type Translate = (key: string) => string
type PlatformLabelPrefix = 'admin.accounts.platforms' | 'admin.groups.platforms'
type AdminStringOption = { value: string; label: string }

const ADMIN_PLATFORM_ORDER: AdminPlatform[] = [
  'anthropic',
  'openai',
  'gemini',
  'grok',
  'antigravity'
]

export function buildAdminPlatformOptions(
  t: Translate,
  options: {
    allLabel?: string
    labelPrefix?: PlatformLabelPrefix
  } = {}
): AdminStringOption[] {
  const {
    allLabel,
    labelPrefix = 'admin.accounts.platforms'
  } = options

  const platformOptions = ADMIN_PLATFORM_ORDER.map((platform) => ({
    value: platform,
    label: t(`${labelPrefix}.${platform}`)
  }))

  if (!allLabel) {
    return platformOptions
  }

  return [{ value: '', label: allLabel }, ...platformOptions]
}

export function buildAdminAccountTypeOptions(t: Translate): AdminStringOption[] {
  return [
    { value: '', label: t('admin.accounts.allTypes') },
    { value: 'oauth', label: t('admin.accounts.oauthType') },
    { value: 'setup-token', label: t('admin.accounts.setupToken') },
    { value: 'apikey', label: t('admin.accounts.apiKey') },
    { value: 'upstream', label: t('admin.accounts.types.upstream') },
    { value: 'session', label: t('admin.accounts.types.session') },
    { value: 'bedrock', label: t('admin.accounts.bedrockLabel') }
  ]
}
