import type { CustomMenuItem } from '@/types'

export type EmbeddedPageTheme = 'light' | 'dark'

export interface EmbeddedPageSettingsStore {
  publicSettingsLoaded: boolean
  fetchPublicSettings: () => Promise<unknown>
}

export function isEmbeddedPageUrl(url: string): boolean {
  return url.startsWith('http://') || url.startsWith('https://')
}

export function resolveCustomPageMenuItem(
  menuItemId: string,
  publicItems: CustomMenuItem[],
  adminItems: CustomMenuItem[],
  isAdmin: boolean
): CustomMenuItem | null {
  const publicItem = publicItems.find((item) => item.id === menuItemId) ?? null
  if (publicItem) {
    return publicItem
  }

  if (!isAdmin) {
    return null
  }

  return adminItems.find((item) => item.id === menuItemId) ?? null
}

export async function loadEmbeddedPageSettings(
  appStore: EmbeddedPageSettingsStore,
  setLoading: (loading: boolean) => void
): Promise<void> {
  if (appStore.publicSettingsLoaded) {
    return
  }

  setLoading(true)
  try {
    await appStore.fetchPublicSettings()
  } finally {
    setLoading(false)
  }
}
