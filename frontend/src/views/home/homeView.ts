import type Icon from '@/components/icons/Icon.vue'

type Translate = (key: string) => string
type IconName = InstanceType<typeof Icon>['$props']['name']

export interface HomeFeatureTag {
  icon: IconName
  key: string
  label: string
}

export interface HomeFeatureCard {
  accentClass: string
  description: string
  icon: IconName
  key: string
  title: string
}

export interface HomeProviderBadge {
  accentClass: string
  initial: string
  key: string
  label: string
  statusLabel: string
  supported: boolean
}

export function resolveHomeContentUrl(content: string): boolean {
  const trimmedContent = content.trim()
  return trimmedContent.startsWith('http://') || trimmedContent.startsWith('https://')
}

export function resolveHomeDashboardPath(isAdmin: boolean): string {
  return isAdmin ? '/admin/dashboard' : '/dashboard'
}

export function resolveHomeUserInitial(email: string | null | undefined): string {
  if (!email) {
    return ''
  }

  return email.charAt(0).toUpperCase()
}

export function buildHomeFeatureTags(t: Translate): HomeFeatureTag[] {
  return [
    {
      icon: 'swap',
      key: 'subscription-to-api',
      label: t('home.tags.subscriptionToApi')
    },
    {
      icon: 'shield',
      key: 'sticky-session',
      label: t('home.tags.stickySession')
    },
    {
      icon: 'chart',
      key: 'realtime-billing',
      label: t('home.tags.realtimeBilling')
    }
  ]
}

export function buildHomeFeatures(t: Translate): HomeFeatureCard[] {
  return [
    {
      accentClass: 'from-blue-500 to-blue-600 shadow-blue-500/30',
      description: t('home.features.unifiedGatewayDesc'),
      icon: 'server',
      key: 'unified-gateway',
      title: t('home.features.unifiedGateway')
    },
    {
      accentClass: 'from-primary-500 to-primary-600 shadow-primary-500/30',
      description: t('home.features.multiAccountDesc'),
      icon: 'users',
      key: 'multi-account',
      title: t('home.features.multiAccount')
    },
    {
      accentClass: 'from-emerald-500 to-teal-600 shadow-emerald-500/30',
      description: t('home.features.balanceQuotaDesc'),
      icon: 'creditCard',
      key: 'balance-quota',
      title: t('home.features.balanceQuota')
    }
  ]
}

export function buildHomeProviders(t: Translate): HomeProviderBadge[] {
  return [
    {
      accentClass: 'from-orange-400 to-orange-500',
      initial: 'C',
      key: 'claude',
      label: t('home.providers.claude'),
      statusLabel: t('home.providers.supported'),
      supported: true
    },
    {
      accentClass: 'from-green-500 to-green-600',
      initial: 'G',
      key: 'gpt',
      label: 'GPT',
      statusLabel: t('home.providers.supported'),
      supported: true
    },
    {
      accentClass: 'from-blue-500 to-blue-600',
      initial: 'G',
      key: 'gemini',
      label: t('home.providers.gemini'),
      statusLabel: t('home.providers.supported'),
      supported: true
    },
    {
      accentClass: 'from-rose-500 to-pink-600',
      initial: 'A',
      key: 'antigravity',
      label: t('home.providers.antigravity'),
      statusLabel: t('home.providers.supported'),
      supported: true
    },
    {
      accentClass: 'from-gray-500 to-gray-600',
      initial: '+',
      key: 'more',
      label: t('home.providers.more'),
      statusLabel: t('home.providers.soon'),
      supported: false
    }
  ]
}
