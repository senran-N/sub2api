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
    class="home-view relative flex min-h-screen flex-col overflow-hidden"
  >
    <HomeBackgroundDecor />

    <header class="home-view__header">
      <nav class="home-view__nav">
        <div class="flex items-center">
          <div class="home-view__brand-mark">
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
            class="home-view__dashboard-link"
          >
            <span class="home-view__dashboard-avatar">
              {{ userInitial }}
            </span>
            <span class="home-view__dashboard-label">{{ t('home.dashboard') }}</span>
            <svg
              class="home-view__dashboard-arrow h-3 w-3"
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
            class="home-view__login-link"
          >
            {{ t('home.login') }}
          </router-link>
        </PublicHeaderActions>
      </nav>
    </header>

    <main class="home-view__main">
      <div class="home-view__main-inner">
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
  ensurePublicSettingsLoaded()
})
</script>

<style scoped>
.home-view {
  background:
    linear-gradient(135deg, color-mix(in srgb, var(--theme-page-bg) 88%, var(--theme-accent-soft)) 0%, color-mix(in srgb, var(--theme-page-bg) 94%, var(--theme-surface-soft)) 45%, var(--theme-page-bg) 100%);
}

.home-view__header {
  position: relative;
  z-index: 20;
  padding: var(--theme-public-header-padding-y) 1.5rem;
}

.home-view__nav,
.home-view__main-inner {
  width: 100%;
  max-width: var(--theme-public-shell-max-width);
  margin: 0 auto;
}

.home-view__nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.home-view__brand-mark {
  display: flex;
  height: var(--theme-public-logo-size);
  width: var(--theme-public-logo-size);
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: var(--theme-public-logo-radius);
  border: var(--theme-public-logo-border);
  box-shadow: var(--theme-public-logo-shadow);
  background: var(--theme-surface);
}

.home-view__main {
  position: relative;
  z-index: 10;
  flex: 1;
  padding: var(--theme-public-main-padding-y) 1.5rem;
}

.home-view__dashboard-link,
.home-view__login-link {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  border-radius: var(--theme-public-action-radius);
  background: var(--theme-surface-contrast);
  border: var(--theme-button-border-width) solid var(--theme-button-primary-border-color);
  box-shadow: var(--theme-button-primary-shadow);
  transition: background-color 0.18s ease, transform 0.18s ease;
}

.home-view__dashboard-link {
  padding: var(--theme-public-action-padding);
}

.home-view__login-link {
  padding: calc(var(--theme-button-padding-y) * 0.6) calc(var(--theme-button-padding-x) * 0.75);
  font-size: 0.75rem;
  font-weight: 600;
}

.home-view__dashboard-link:hover,
.home-view__login-link:hover {
  background: color-mix(in srgb, var(--theme-surface-contrast) 88%, var(--theme-page-text));
  box-shadow: var(--theme-button-primary-hover-shadow);
  transform: var(--theme-button-primary-hover-transform);
}

.home-view__dashboard-avatar {
  display: flex;
  height: calc(var(--theme-header-avatar-size) - 12px);
  width: calc(var(--theme-header-avatar-size) - 12px);
  align-items: center;
  justify-content: center;
  border-radius: var(--theme-header-avatar-radius);
  background: linear-gradient(
    135deg,
    color-mix(in srgb, var(--theme-accent) 84%, var(--theme-surface)),
    var(--theme-accent)
  );
  color: var(--theme-filled-text);
  font-size: 0.625rem;
  font-weight: 700;
}

.home-view__dashboard-label,
.home-view__login-link {
  color: var(--theme-surface-contrast-text);
}

.home-view__dashboard-label {
  font-size: 0.75rem;
  font-weight: 500;
}

.home-view__dashboard-arrow {
  color: color-mix(in srgb, var(--theme-surface-contrast-text) 56%, transparent);
}

@media (max-width: 767px) {
  .home-view__header,
  .home-view__main {
    padding-left: 1rem;
    padding-right: 1rem;
  }
}
</style>
