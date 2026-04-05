import { computed, ref } from 'vue'
import { useAppStore } from '@/stores'

export const PUBLIC_GITHUB_URL = 'https://github.com/senran-N/sub2api'

export function usePublicSiteShell() {
  const appStore = useAppStore()

  const siteName = computed(
    () => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API'
  )
  const siteLogo = computed(
    () => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || ''
  )
  const siteSubtitle = computed(
    () =>
      appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform'
  )
  const docUrl = computed(
    () => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''
  )
  const homeContent = computed(
    () => appStore.cachedPublicSettings?.home_content || ''
  )
  const isDark = ref(document.documentElement.classList.contains('dark'))
  const currentYear = computed(() => new Date().getFullYear())

  const toggleTheme = () => {
    isDark.value = !isDark.value
    document.documentElement.classList.toggle('dark', isDark.value)
    localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
  }

  const initTheme = () => {
    const savedTheme = localStorage.getItem('theme')

    if (
      savedTheme === 'dark' ||
      (!savedTheme &&
        window.matchMedia('(prefers-color-scheme: dark)').matches)
    ) {
      isDark.value = true
      document.documentElement.classList.add('dark')
      return
    }

    isDark.value = false
    document.documentElement.classList.remove('dark')
  }

  const ensurePublicSettingsLoaded = () => {
    if (!appStore.publicSettingsLoaded) {
      void appStore.fetchPublicSettings()
    }
  }

  return {
    siteName,
    siteLogo,
    siteSubtitle,
    docUrl,
    homeContent,
    githubUrl: PUBLIC_GITHUB_URL,
    isDark,
    currentYear,
    toggleTheme,
    initTheme,
    ensurePublicSettingsLoaded
  }
}
