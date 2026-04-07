import { computed, onMounted, onUnmounted, ref, type ComputedRef } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { buildEmbeddedUrl, detectTheme } from '@/utils/embeddedUrl'
import { isEmbeddedPageUrl, loadEmbeddedPageSettings, type EmbeddedPageTheme } from './embeddedPageFrame'

export function createEmbeddedPageThemeObserver(
  onThemeChange: (theme: EmbeddedPageTheme) => void
): MutationObserver | null {
  if (typeof document === 'undefined') {
    return null
  }

  const observer = new MutationObserver(() => {
    onThemeChange(detectTheme())
  })

  observer.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class']
  })

  return observer
}

export function useEmbeddedPageFrame(baseUrl: ComputedRef<string>) {
  const { locale } = useI18n()
  const appStore = useAppStore()
  const authStore = useAuthStore()

  const loading = ref(false)
  const embeddedTheme = ref<EmbeddedPageTheme>('light')
  let themeObserver: MutationObserver | null = null

  const embeddedUrl = computed(() =>
    buildEmbeddedUrl(
      baseUrl.value.trim(),
      authStore.user?.id,
      authStore.token,
      embeddedTheme.value,
      locale.value
    )
  )

  const isValidUrl = computed(() => isEmbeddedPageUrl(embeddedUrl.value))

  onMounted(async () => {
    embeddedTheme.value = detectTheme()
    themeObserver = createEmbeddedPageThemeObserver((theme) => {
      embeddedTheme.value = theme
    })
    await loadEmbeddedPageSettings(appStore, (value) => {
      loading.value = value
    })
  })

  onUnmounted(() => {
    themeObserver?.disconnect()
    themeObserver = null
  })

  return {
    loading,
    embeddedUrl,
    isValidUrl
  }
}
