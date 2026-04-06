import type Icon from '@/components/icons/Icon.vue'

type Translate = (key: string) => string
type IconName = InstanceType<typeof Icon>['$props']['name']

export interface HomeFeatureTag {
  icon: IconName
  key: string
  label: string
}

export interface HomeFeatureCard {
  accentTone: 'info' | 'accent' | 'success'
  description: string
  icon: IconName
  key: string
  title: string
}

export interface HomeProviderBadge {
  accentTone: 'brand-orange' | 'success' | 'info' | 'brand-rose' | 'neutral'
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
      accentTone: 'info',
      description: t('home.features.unifiedGatewayDesc'),
      icon: 'server',
      key: 'unified-gateway',
      title: t('home.features.unifiedGateway')
    },
    {
      accentTone: 'accent',
      description: t('home.features.multiAccountDesc'),
      icon: 'users',
      key: 'multi-account',
      title: t('home.features.multiAccount')
    },
    {
      accentTone: 'success',
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
      accentTone: 'brand-orange',
      initial: 'C',
      key: 'claude',
      label: t('home.providers.claude'),
      statusLabel: t('home.providers.supported'),
      supported: true
    },
    {
      accentTone: 'success',
      initial: 'G',
      key: 'gpt',
      label: 'GPT',
      statusLabel: t('home.providers.supported'),
      supported: true
    },
    {
      accentTone: 'info',
      initial: 'G',
      key: 'gemini',
      label: t('home.providers.gemini'),
      statusLabel: t('home.providers.supported'),
      supported: true
    },
    {
      accentTone: 'brand-rose',
      initial: 'A',
      key: 'antigravity',
      label: t('home.providers.antigravity'),
      statusLabel: t('home.providers.supported'),
      supported: true
    },
    {
      accentTone: 'neutral',
      initial: '+',
      key: 'more',
      label: t('home.providers.more'),
      statusLabel: t('home.providers.soon'),
      supported: false
    }
  ]
}
