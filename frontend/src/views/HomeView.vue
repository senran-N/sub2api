<template>
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div
    v-else
    class="relative flex min-h-screen flex-col overflow-hidden bg-gradient-to-br from-gray-50 via-primary-50/30 to-gray-100 dark:from-dark-950 dark:via-dark-900 dark:to-dark-950"
  >
    <HomeBackgroundDecor />

    <header class="relative z-20 px-6 py-4">
      <nav class="mx-auto flex max-w-6xl items-center justify-between">
        <div class="flex items-center">
          <div class="h-10 w-10 overflow-hidden rounded-xl shadow-md">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
        </div>

        <PublicHeaderActions
          :doc-url="docUrl"
          :is-dark="isDark"
          :docs-title="t('home.viewDocs')"
          :theme-title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          @toggle-theme="toggleTheme"
        >
          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="inline-flex items-center gap-1.5 rounded-full bg-gray-900 py-1 pl-1 pr-2.5 transition-colors hover:bg-gray-800 dark:bg-gray-800 dark:hover:bg-gray-700"
          >
            <span
              class="flex h-5 w-5 items-center justify-center rounded-full bg-gradient-to-br from-primary-400 to-primary-600 text-[10px] font-semibold text-white"
            >
              {{ userInitial }}
            </span>
            <span class="text-xs font-medium text-white">{{ t('home.dashboard') }}</span>
            <svg
              class="h-3 w-3 text-gray-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M4.5 19.5l15-15m0 0H8.25m11.25 0v11.25"
              />
            </svg>
          </router-link>
          <router-link
            v-else
            to="/login"
            class="inline-flex items-center rounded-full bg-gray-900 px-3 py-1 text-xs font-medium text-white transition-colors hover:bg-gray-800 dark:bg-gray-800 dark:hover:bg-gray-700"
          >
            {{ t('home.login') }}
          </router-link>
        </PublicHeaderActions>
      </nav>
    </header>

    <main class="relative z-10 flex-1 px-6 py-16">
      <div class="mx-auto max-w-6xl">
        <HomeHeroSection
          :cta-label="isAuthenticated ? t('home.goToDashboard') : t('home.getStarted')"
          :cta-path="isAuthenticated ? dashboardPath : '/login'"
          :site-name="siteName"
          :site-subtitle="siteSubtitle"
        />

        <HomeFeatureTags :tags="featureTags" />
        <HomeFeaturesGrid :features="features" />
        <HomeProvidersSection
          :description="t('home.providers.description')"
          :providers="providers"
          :title="t('home.providers.title')"
        />
      </div>
    </main>

    <PublicPageFooter
      :current-year="currentYear"
      :site-name="siteName"
      :doc-url="docUrl"
      :github-url="githubUrl"
      :rights-label="t('home.footer.allRightsReserved')"
      :docs-label="t('home.docs')"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import PublicHeaderActions from '@/components/public/PublicHeaderActions.vue'
import PublicPageFooter from '@/components/public/PublicPageFooter.vue'
import { usePublicSiteShell } from '@/composables/usePublicSiteShell'
import { useAuthStore } from '@/stores'
import HomeBackgroundDecor from './home/HomeBackgroundDecor.vue'
import HomeFeatureTags from './home/HomeFeatureTags.vue'
import HomeFeaturesGrid from './home/HomeFeaturesGrid.vue'
import HomeHeroSection from './home/HomeHeroSection.vue'
import HomeProvidersSection from './home/HomeProvidersSection.vue'
import {
  buildHomeFeatureTags,
  buildHomeFeatures,
  buildHomeProviders,
  resolveHomeContentUrl,
  resolveHomeDashboardPath,
  resolveHomeUserInitial
} from './home/homeView'

const { t } = useI18n()

const authStore = useAuthStore()
const {
  siteName,
  siteLogo,
  siteSubtitle,
  docUrl,
  homeContent,
  githubUrl,
  isDark,
  currentYear,
  toggleTheme,
  initTheme,
  ensurePublicSettingsLoaded
} = usePublicSiteShell()

const isHomeContentUrl = computed(() => resolveHomeContentUrl(homeContent.value))
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => resolveHomeDashboardPath(isAdmin.value))
const userInitial = computed(() => resolveHomeUserInitial(authStore.user?.email))
const featureTags = computed(() => buildHomeFeatureTags(t))
const features = computed(() => buildHomeFeatures(t))
const providers = computed(() => buildHomeProviders(t))

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  ensurePublicSettingsLoaded()
})
</script>
