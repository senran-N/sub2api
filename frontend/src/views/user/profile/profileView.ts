import { computed, h, type Component } from 'vue'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { formatCurrency, formatDate } from '@/utils/format'

interface ProfilePublicSettingsStore {
  publicSettingsLoaded: boolean
  fetchPublicSettings: () => Promise<unknown>
  cachedPublicSettings?: {
    contact_info?: string
    balance_low_notify_enabled?: boolean
    balance_low_notify_threshold?: number
    linuxdo_oauth_enabled?: boolean
    oidc_oauth_enabled?: boolean
    wechat_oauth_enabled?: boolean
  } | null
}

function createProfileStatIcon(path: string): Component {
  return {
    render() {
      return h(
        'svg',
        {
          fill: 'none',
          viewBox: '0 0 24 24',
          stroke: 'currentColor',
          'stroke-width': '1.5'
        },
        [h('path', { d: path })]
      )
    }
  }
}

export const profileWalletIcon = createProfileStatIcon(
  'M21 12a2.25 2.25 0 00-2.25-2.25H15a3 3 0 11-6 0H5.25A2.25 2.25 0 003 12'
)

export const profileConcurrencyIcon = createProfileStatIcon(
  'm3.75 13.5 10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z'
)

export const profileMemberSinceIcon = createProfileStatIcon('M6.75 3v2.25M17.25 3v2.25')

export function formatProfileBalance(balance: number | undefined): string {
  return formatCurrency(balance ?? 0)
}

export function formatProfileMemberSince(createdAt: string | undefined): string {
  return formatDate(createdAt ?? '', {
    year: 'numeric',
    month: 'long'
  })
}

export async function loadProfilePublicSettings(appStore: ProfilePublicSettingsStore): Promise<void> {
  if (appStore.publicSettingsLoaded) {
    return
  }

  await appStore.fetchPublicSettings()
}

export function useProfileViewModel() {
  const authStore = useAuthStore()
  const appStore = useAppStore()
  const user = computed(() => authStore.user)
  const contactInfo = computed(() => appStore.cachedPublicSettings?.contact_info || '')
  const balanceLowNotifyEnabled = computed(() => !!appStore.cachedPublicSettings?.balance_low_notify_enabled)
  const systemDefaultThreshold = computed(() => appStore.cachedPublicSettings?.balance_low_notify_threshold || 0)
  const linuxdoEnabled = computed(() => !!appStore.cachedPublicSettings?.linuxdo_oauth_enabled)
  const oidcEnabled = computed(() => !!appStore.cachedPublicSettings?.oidc_oauth_enabled)
  const wechatEnabled = computed(() => !!appStore.cachedPublicSettings?.wechat_oauth_enabled)

  async function loadContactInfo() {
    try {
      await loadProfilePublicSettings(appStore)
    } catch (error) {
      console.error('Failed to load contact info:', error)
    }
  }

  async function refreshProfile() {
    try {
      await authStore.refreshUser()
    } catch (error) {
      console.error('Failed to refresh profile:', error)
    }
  }

  return {
    user,
    contactInfo,
    balanceLowNotifyEnabled,
    systemDefaultThreshold,
    linuxdoEnabled,
    oidcEnabled,
    wechatEnabled,
    loadContactInfo,
    refreshProfile
  }
}
