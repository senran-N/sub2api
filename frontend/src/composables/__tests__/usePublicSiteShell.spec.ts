import { beforeEach, describe, expect, it, vi } from 'vitest'
import { usePublicSiteShell } from '../usePublicSiteShell'

const fetchPublicSettings = vi.fn()

const storeState = vi.hoisted(() => ({
  publicSettingsLoaded: false,
  siteName: 'Fallback Site',
  siteLogo: '/fallback-logo.png',
  docUrl: 'https://fallback.example.com',
  cachedPublicSettings: {
    site_name: 'Cached Site',
    site_logo: '/cached-logo.png',
    site_subtitle: 'Cached Subtitle',
    doc_url: 'https://docs.example.com',
    home_content: '<h1>cached</h1>'
  }
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    ...storeState,
    fetchPublicSettings
  })
}))

describe('usePublicSiteShell', () => {
  beforeEach(() => {
    fetchPublicSettings.mockReset()
    storeState.publicSettingsLoaded = false
    document.documentElement.classList.remove('dark')
    localStorage.clear()
    vi.stubGlobal(
      'matchMedia',
      vi.fn().mockReturnValue({
        matches: false,
        addEventListener: vi.fn(),
        removeEventListener: vi.fn()
      })
    )
  })

  it('returns cached public settings and toggles theme', () => {
    const shell = usePublicSiteShell()

    expect(shell.siteName.value).toBe('Cached Site')
    expect(shell.siteLogo.value).toBe('/cached-logo.png')
    expect(shell.siteSubtitle.value).toBe('Cached Subtitle')
    expect(shell.docUrl.value).toBe('https://docs.example.com')
    expect(shell.homeContent.value).toBe('<h1>cached</h1>')

    shell.toggleTheme()

    expect(shell.isDark.value).toBe(true)
    expect(document.documentElement.classList.contains('dark')).toBe(true)
    expect(localStorage.getItem('theme')).toBe('dark')
  })

  it('initializes theme and only fetches public settings when needed', () => {
    localStorage.setItem('theme', 'dark')
    const shell = usePublicSiteShell()

    shell.initTheme()
    shell.ensurePublicSettingsLoaded()

    expect(shell.isDark.value).toBe(true)
    expect(fetchPublicSettings).toHaveBeenCalledTimes(1)

    storeState.publicSettingsLoaded = true
    const loadedShell = usePublicSiteShell()
    loadedShell.ensurePublicSettingsLoaded()
    expect(fetchPublicSettings).toHaveBeenCalledTimes(1)
  })
})
