<script setup lang="ts">
import { RouterView, useRouter, useRoute } from 'vue-router'
import { computed, defineAsyncComponent, onMounted, onBeforeUnmount, watch, watchEffect } from 'vue'
import Toast from '@/components/common/Toast.vue'
import NavigationProgress from '@/components/common/NavigationProgress.vue'
import i18n from '@/i18n'
import { resolveRouteDocumentTitle } from '@/router/title'
import { scheduleDeferredTask } from '@/utils/deferredTask'
import {
  useAdminSettingsStore,
  useAppStore,
  useAuthStore,
  useSubscriptionStore,
  useAnnouncementStore
} from '@/stores'
import { fetchSetupStatus } from '@/api/bootstrap'

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()
const adminSettingsStore = useAdminSettingsStore()
const authStore = useAuthStore()
const subscriptionStore = useSubscriptionStore()
const announcementStore = useAnnouncementStore()
const AnnouncementPopup = defineAsyncComponent(() => import('@/components/common/AnnouncementPopup.vue'))
const shouldRenderAnnouncementPopup = computed(() => authStore.isAuthenticated)
let cancelAuthenticatedWarmup: (() => void) | null = null
let delayedAnnouncementTimer: ReturnType<typeof setTimeout> | null = null

/**
 * Update favicon dynamically
 * @param logoUrl - URL of the logo to use as favicon
 */
function updateFavicon(logoUrl: string) {
  // Find existing favicon link or create new one
  let link = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
  if (!link) {
    link = document.createElement('link')
    link.rel = 'icon'
    document.head.appendChild(link)
  }
  link.type = logoUrl.endsWith('.svg') ? 'image/svg+xml' : 'image/x-icon'
  link.href = logoUrl
}

// Watch for site settings changes and update favicon/title
watch(
  () => appStore.siteLogo,
  (newLogo) => {
    if (newLogo) {
      updateFavicon(newLogo)
    }
  },
  { immediate: true }
)

watchEffect(() => {
  void i18n.global.locale.value

  document.title = resolveRouteDocumentTitle(route, {
    siteName: appStore.siteName,
    publicCustomMenuItems: appStore.cachedPublicSettings?.custom_menu_items ?? [],
    adminCustomMenuItems: adminSettingsStore.customMenuItems,
    isAdmin: authStore.isAdmin
  })
})

// Watch for authentication state and manage subscription data + announcements
function onVisibilityChange() {
  if (document.visibilityState === 'visible' && authStore.isAuthenticated) {
    announcementStore.fetchAnnouncements()
  }
}

function clearAuthenticatedWarmup(): void {
  if (cancelAuthenticatedWarmup) {
    cancelAuthenticatedWarmup()
    cancelAuthenticatedWarmup = null
  }

  if (delayedAnnouncementTimer) {
    clearTimeout(delayedAnnouncementTimer)
    delayedAnnouncementTimer = null
  }
}

function scheduleAuthenticatedWarmup(isNewLogin: boolean): void {
  clearAuthenticatedWarmup()
  subscriptionStore.startPolling()

  cancelAuthenticatedWarmup = scheduleDeferredTask(() => {
    cancelAuthenticatedWarmup = null

    subscriptionStore.fetchActiveSubscriptions().catch((error) => {
      console.error('Failed to preload subscriptions:', error)
    })

    if (isNewLogin) {
      delayedAnnouncementTimer = setTimeout(() => {
        delayedAnnouncementTimer = null
        void announcementStore.fetchAnnouncements(true)
      }, 3000)
      return
    }

    void announcementStore.fetchAnnouncements()
  }, { timeout: 2000 })
}

watch(
  () => authStore.isAuthenticated,
  (isAuthenticated, oldValue) => {
    if (isAuthenticated) {
      // Warm up authenticated-only data after the first paint to keep
      // initial rendering responsive.
      scheduleAuthenticatedWarmup(oldValue === false)

      // Register visibility change listener
      document.addEventListener('visibilitychange', onVisibilityChange)
    } else {
      clearAuthenticatedWarmup()

      // User logged out: clear data and stop polling
      subscriptionStore.clear()
      announcementStore.reset()
      document.removeEventListener('visibilitychange', onVisibilityChange)
    }
  },
  { immediate: true }
)

// Route change trigger (throttled by store)
router.afterEach(() => {
  if (authStore.isAuthenticated) {
    announcementStore.fetchAnnouncements()
  }
})

onBeforeUnmount(() => {
  clearAuthenticatedWarmup()
  document.removeEventListener('visibilitychange', onVisibilityChange)
})

onMounted(async () => {
  // Check if setup is needed
  try {
    const status = await fetchSetupStatus()
    if (status.needs_setup && route.path !== '/setup') {
      router.replace('/setup')
      return
    }
  } catch {
    // If setup endpoint fails, assume normal mode and continue
  }

  // Load public settings into appStore (will be cached for other components)
  await appStore.fetchPublicSettings()
})
</script>

<template>
  <div class="scanlines"></div>
  <NavigationProgress />
  <RouterView />
  <Toast />
  <AnnouncementPopup v-if="shouldRenderAnnouncementPopup" />
</template>
