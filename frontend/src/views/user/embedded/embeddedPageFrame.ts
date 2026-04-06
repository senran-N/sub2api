import { resolveCustomPageMenuItem } from '@/utils/customMenu'

export type EmbeddedPageTheme = 'light' | 'dark'

export interface EmbeddedPageSettingsStore {
  publicSettingsLoaded: boolean
  fetchPublicSettings: () => Promise<unknown>
}

export function isEmbeddedPageUrl(url: string): boolean {
  return url.startsWith('http://') || url.startsWith('https://')
}

export { resolveCustomPageMenuItem }

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
