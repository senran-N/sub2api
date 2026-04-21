type Translate = (key: string, params?: Record<string, unknown>) => string

export function getAccountTypeLabel(accountType: string | null | undefined, t: Translate): string {
  switch (accountType) {
    case 'oauth':
      return t('admin.accounts.types.oauth')
    case 'apikey':
      return t('admin.accounts.apiKey')
    case 'upstream':
      return t('admin.accounts.types.upstream')
    case 'session':
      return t('admin.accounts.types.session')
    case 'setup-token':
      return t('admin.accounts.setupToken')
    case 'bedrock':
      return t('admin.accounts.bedrockLabel')
    default:
      return accountType ?? ''
  }
}

export function getAccountStatusLabel(accountStatus: string | null | undefined, t: Translate): string {
  switch (accountStatus) {
    case 'active':
      return t('admin.accounts.status.active')
    case 'inactive':
      return t('admin.accounts.status.inactive')
    case 'error':
      return t('admin.accounts.status.error')
    case 'cooldown':
      return t('admin.accounts.status.cooldown')
    case 'paused':
      return t('admin.accounts.status.paused')
    case 'limited':
      return t('admin.accounts.status.limited')
    case 'rate_limited':
      return t('admin.accounts.status.rateLimited')
    case 'overloaded':
      return t('admin.accounts.status.overloaded')
    case 'temp_unschedulable':
      return t('admin.accounts.status.tempUnschedulable')
    default:
      return accountStatus ?? ''
  }
}
