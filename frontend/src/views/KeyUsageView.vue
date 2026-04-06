<template>
  <div class="key-usage-view relative flex min-h-screen flex-col">
    <header class="key-usage-view__header relative z-20">
      <nav class="key-usage-view__nav mx-auto flex items-center justify-between">
        <router-link to="/home" class="flex items-center gap-3">
          <div class="key-usage-view__logo overflow-hidden">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <span class="key-usage-view__brand text-lg font-semibold tracking-tight">
            {{ siteName }}
          </span>
        </router-link>
        <PublicHeaderActions
          :doc-url="docUrl"
          :is-dark="isDark"
          :docs-title="t('home.viewDocs')"
          :theme-title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          @toggle-theme="toggleTheme"
        />
      </nav>
    </header>

    <main class="key-usage-view__main mx-auto w-full flex-1">
      <div class="mb-12 text-center">
        <h1 class="key-usage-view__title mb-3 text-3xl font-bold tracking-tight sm:text-4xl">
          {{ t('keyUsage.title') }}
        </h1>
        <p class="key-usage-view__subtitle mx-auto max-w-md text-base">
          {{ t('keyUsage.subtitle') }}
        </p>
      </div>

      <KeyUsageQueryForm
        :api-key="apiKey"
        :key-visible="keyVisible"
        :is-querying="isQuerying"
        :show-date-picker="showDatePicker"
        :current-range="currentRange"
        :custom-start-date="customStartDate"
        :custom-end-date="customEndDate"
        :date-ranges="dateRanges"
        :placeholder="t('keyUsage.placeholder')"
        :query-label="t('keyUsage.query')"
        :querying-label="t('keyUsage.querying')"
        :privacy-note="t('keyUsage.privacyNote')"
        :date-range-label="t('keyUsage.dateRange')"
        :apply-label="t('keyUsage.apply')"
        @update:api-key="apiKey = $event"
        @update:custom-start-date="customStartDate = $event"
        @update:custom-end-date="customEndDate = $event"
        @toggle-visible="keyVisible = !keyVisible"
        @set-range="setDateRange"
        @query="queryKey"
      />

      <div v-if="showResults">
        <KeyUsageLoadingState v-if="showLoading" />

        <div v-else-if="resultData" class="space-y-6">
          <KeyUsageStatusBadge :status="statusInfo" />
          <KeyUsageRingCards
            :items="ringItems"
            :grid-class="ringGridClass"
            :display-pcts="displayPcts"
            :used-label="t('keyUsage.used')"
            :track-color="ringTrackColor"
            :circumference="KEY_USAGE_RING_CIRCUMFERENCE"
            :gradients="ringGradients"
            :animated="ringAnimated"
            :format-reset-time="formatResetTime"
          />
          <KeyUsageDetailCard :rows="detailRows" :title="t('keyUsage.detailInfo')" />
          <KeyUsageUsageStatsCard :cells="usageStatCells" :title="t('keyUsage.tokenStats')" />
          <KeyUsageModelStatsTable
            :items="modelStats"
            :title="t('keyUsage.modelStats')"
            :labels="modelStatsLabels"
            :fmt-num="fmtNum"
            :usd="usd"
          />
        </div>
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
import { useAppStore } from '@/stores'
import KeyUsageDetailCard from './key-usage/KeyUsageDetailCard.vue'
import KeyUsageLoadingState from './key-usage/KeyUsageLoadingState.vue'
import KeyUsageModelStatsTable from './key-usage/KeyUsageModelStatsTable.vue'
import KeyUsageQueryForm from './key-usage/KeyUsageQueryForm.vue'
import KeyUsageRingCards from './key-usage/KeyUsageRingCards.vue'
import KeyUsageStatusBadge from './key-usage/KeyUsageStatusBadge.vue'
import KeyUsageUsageStatsCard from './key-usage/KeyUsageUsageStatsCard.vue'
import {
  KEY_USAGE_RING_CIRCUMFERENCE
} from './key-usage/keyUsageView'
import { useKeyUsageViewModel } from './key-usage/useKeyUsageViewModel'

const { t, locale } = useI18n()
const appStore = useAppStore()

const {
  siteName,
  siteLogo,
  docUrl,
  githubUrl,
  isDark,
  currentYear,
  toggleTheme,
  initTheme,
  ensurePublicSettingsLoaded
} = usePublicSiteShell()

const {
  apiKey,
  currentRange,
  customEndDate,
  customStartDate,
  dateRanges,
  detailRows,
  displayPcts,
  fmtNum,
  formatResetTime,
  isQuerying,
  keyVisible,
  modelStats,
  modelStatsLabels,
  queryKey,
  resultData,
  ringAnimated,
  ringGradients,
  ringGridClass,
  ringItems,
  ringTrackColor,
  setDateRange,
  showDatePicker,
  showLoading,
  showResults,
  statusInfo,
  usageStatCells,
  usd
} = useKeyUsageViewModel({
  locale: computed(() => locale.value),
  showError: appStore.showError,
  showInfo: appStore.showInfo,
  showSuccess: appStore.showSuccess,
  t
})

onMounted(() => {
  initTheme()
  ensurePublicSettingsLoaded()
})
</script>

<style scoped>
.key-usage-view {
  background:
    radial-gradient(circle at top center, color-mix(in srgb, var(--theme-accent-soft) 36%, transparent), transparent 48%),
    linear-gradient(180deg, color-mix(in srgb, var(--theme-page-bg) 94%, var(--theme-surface-soft)) 0%, var(--theme-page-bg) 100%);
}

.key-usage-view__header {
  padding: var(--theme-public-header-padding-y) 1.5rem;
}

.key-usage-view__nav,
.key-usage-view__main {
  max-width: var(--theme-public-shell-max-width);
}

.key-usage-view__logo {
  height: var(--theme-public-logo-size);
  width: var(--theme-public-logo-size);
  border-radius: var(--theme-public-logo-radius);
  border: var(--theme-public-logo-border);
  box-shadow: var(--theme-public-logo-shadow);
}

.key-usage-view__main {
  padding: var(--theme-public-main-padding-y) 1.5rem;
}

.key-usage-view__brand,
.key-usage-view__title {
  color: var(--theme-page-text);
}

.key-usage-view__subtitle {
  color: var(--theme-page-muted);
}
</style>
