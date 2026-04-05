<template>
  <div class="relative flex min-h-screen flex-col bg-gray-50 dark:bg-dark-950">
    <header class="relative z-20 px-6 py-4">
      <nav class="mx-auto flex max-w-6xl items-center justify-between">
        <router-link to="/home" class="flex items-center gap-3">
          <div class="h-10 w-10 overflow-hidden rounded-xl shadow-md">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <span class="text-lg font-semibold tracking-tight text-gray-900 dark:text-white">
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

    <main class="mx-auto w-full max-w-5xl flex-1 px-6 py-12">
      <div class="mb-12 text-center">
        <h1 class="mb-3 text-3xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-4xl">
          {{ t('keyUsage.title') }}
        </h1>
        <p class="mx-auto max-w-md text-base text-gray-500 dark:text-dark-400">
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
            :gradients="KEY_USAGE_RING_GRADIENTS"
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
  KEY_USAGE_RING_CIRCUMFERENCE,
  KEY_USAGE_RING_GRADIENTS
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
  isDark,
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
