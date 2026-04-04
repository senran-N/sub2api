import { ref } from 'vue'
import type { AdminGroup, Proxy as AccountProxy } from '@/types'

interface WindowTarget {
  addEventListener: Window['addEventListener']
  removeEventListener: Window['removeEventListener']
}

interface AccountsViewBootstrapOptions {
  t: (key: string) => string
  showError: (message: string) => void
  load: () => Promise<void>
  fetchProxies: () => Promise<AccountProxy[]>
  fetchGroups: () => Promise<AdminGroup[]>
  closeActionMenu: () => void
  initializeAutoRefresh: () => void
  disposeAutoRefresh: () => void
  windowTarget?: WindowTarget
}

function getBootstrapErrorMessage(
  error: unknown,
  fallback: string
): string {
  if (error instanceof Error && error.message) {
    return error.message
  }

  if (
    typeof error === 'object' &&
    error !== null &&
    'message' in error &&
    typeof error.message === 'string' &&
    error.message
  ) {
    return error.message
  }

  return fallback
}

export function useAccountsViewBootstrap(options: AccountsViewBootstrapOptions) {
  const proxies = ref<AccountProxy[]>([])
  const groups = ref<AdminGroup[]>([])

  const handleScroll = () => {
    options.closeActionMenu()
  }

  const loadReferenceData = async () => {
    try {
      const [nextProxies, nextGroups] = await Promise.all([
        options.fetchProxies(),
        options.fetchGroups()
      ])
      proxies.value = nextProxies
      groups.value = nextGroups
    } catch (error) {
      proxies.value = []
      groups.value = []
      options.showError(
        getBootstrapErrorMessage(error, options.t('admin.accounts.failedToLoad'))
      )
    }
  }

  const initialize = () => {
    const windowTarget = options.windowTarget ?? window

    void options.load()
    void loadReferenceData()
    windowTarget.addEventListener('scroll', handleScroll, true)
    options.initializeAutoRefresh()
  }

  const dispose = () => {
    const windowTarget = options.windowTarget ?? window

    windowTarget.removeEventListener('scroll', handleScroll, true)
    options.disposeAutoRefresh()
  }

  return {
    proxies,
    groups,
    initialize,
    dispose
  }
}
