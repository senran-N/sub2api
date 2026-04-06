import { onMounted, onUnmounted, ref } from 'vue'

export function useDocumentThemeVersion() {
  const themeVersion = ref(0)
  let themeObserver: MutationObserver | null = null

  const refreshTheme = () => {
    themeVersion.value += 1
  }

  onMounted(() => {
    refreshTheme()

    if (typeof MutationObserver === 'undefined' || typeof document === 'undefined') {
      return
    }

    themeObserver = new MutationObserver(refreshTheme)
    themeObserver.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class', 'data-brand-theme', 'style']
    })
  })

  onUnmounted(() => {
    themeObserver?.disconnect()
    themeObserver = null
  })

  return themeVersion
}
