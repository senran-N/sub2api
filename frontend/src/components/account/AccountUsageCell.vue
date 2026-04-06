<template>
  <div v-if="showUsageWindows">
    <template
      v-if="
        account.platform === 'anthropic' &&
        (account.type === 'oauth' || account.type === 'setup-token')
      "
    >
      <div v-if="loading" class="space-y-1.5">
        <div v-for="index in account.type === 'oauth' ? 3 : 1" :key="index" class="account-usage-cell__skeleton-row">
          <div class="account-usage-cell__skeleton-block account-usage-cell__skeleton-block--label animate-pulse"></div>
          <div class="account-usage-cell__skeleton-bar animate-pulse"></div>
          <div class="account-usage-cell__skeleton-block account-usage-cell__skeleton-block--label animate-pulse"></div>
        </div>
      </div>

      <div v-else-if="error" class="account-usage-cell__error text-xs">
        {{ error }}
      </div>

      <div v-else-if="usageInfo" class="space-y-1">
        <div
          v-if="usageInfo.error"
          class="account-usage-cell__warning account-usage-cell__warning--truncated text-xs"
          :title="usageInfo.error"
        >
          {{ usageInfo.error }}
        </div>

        <UsageProgressBar
          v-if="usageInfo.five_hour"
          label="5h"
          :utilization="usageInfo.five_hour.utilization"
          :resets-at="usageInfo.five_hour.resets_at"
          :window-stats="usageInfo.five_hour.window_stats"
          color="indigo"
        />

        <UsageProgressBar
          v-if="usageInfo.seven_day"
          label="7d"
          :utilization="usageInfo.seven_day.utilization"
          :resets-at="usageInfo.seven_day.resets_at"
          color="emerald"
        />

        <UsageProgressBar
          v-if="usageInfo.seven_day_sonnet"
          label="7d S"
          :utilization="usageInfo.seven_day_sonnet.utilization"
          :resets-at="usageInfo.seven_day_sonnet.resets_at"
          color="purple"
        />

        <div class="account-usage-cell__inline-row account-usage-cell__inline-row--compact">
          <span v-if="usageInfo.source === 'passive'" class="account-usage-cell__muted-note account-usage-cell__muted-note--compact italic">
            {{ t('admin.accounts.usageWindow.passiveSampled') }}
          </span>
          <button
            type="button"
            class="account-usage-cell__action account-usage-cell__action--compact"
            :disabled="activeQueryLoading"
            @click="loadActiveUsage"
          >
            <svg
              class="h-2.5 w-2.5"
              :class="{ 'animate-spin': activeQueryLoading }"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
            {{ t('admin.accounts.usageWindow.activeQuery') }}
          </button>
        </div>
      </div>

      <div v-else class="account-usage-cell__empty text-xs">-</div>
    </template>

    <template v-else-if="account.platform === 'openai' && account.type === 'oauth'">
      <div v-if="hasOpenAIUsageFallback" class="space-y-1">
        <UsageProgressBar
          v-if="usageInfo?.five_hour"
          label="5h"
          :utilization="usageInfo.five_hour.utilization"
          :resets-at="usageInfo.five_hour.resets_at"
          :window-stats="usageInfo.five_hour.window_stats"
          :show-now-when-idle="true"
          color="indigo"
        />
        <UsageProgressBar
          v-if="usageInfo?.seven_day"
          label="7d"
          :utilization="usageInfo.seven_day.utilization"
          :resets-at="usageInfo.seven_day.resets_at"
          :window-stats="usageInfo.seven_day.window_stats"
          :show-now-when-idle="true"
          color="emerald"
        />
      </div>
      <div v-else-if="loading" class="space-y-1.5">
        <div v-for="index in 2" :key="index" class="account-usage-cell__skeleton-row">
          <div class="account-usage-cell__skeleton-block account-usage-cell__skeleton-block--label animate-pulse"></div>
          <div class="account-usage-cell__skeleton-bar animate-pulse"></div>
          <div class="account-usage-cell__skeleton-block account-usage-cell__skeleton-block--label animate-pulse"></div>
        </div>
      </div>
      <div v-else class="account-usage-cell__empty text-xs">-</div>
    </template>

    <template v-else-if="account.platform === 'antigravity' && account.type === 'oauth'">
      <div v-if="antigravityTierLabel" class="account-usage-cell__inline-row account-usage-cell__inline-row--with-margin">
        <span :class="antigravityTierClass">
          {{ antigravityTierLabel }}
        </span>
        <span v-if="hasIneligibleTiers" class="group relative cursor-help">
          <svg
            class="account-usage-cell__danger-icon h-3.5 w-3.5"
            fill="currentColor"
            viewBox="0 0 20 20"
          >
            <path
              fill-rule="evenodd"
              d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z"
              clip-rule="evenodd"
            />
          </svg>
          <span
            class="account-usage-cell__tooltip"
          >
            {{ t('admin.accounts.ineligibleWarning') }}
          </span>
        </span>
      </div>

      <div v-if="isForbidden" class="space-y-1">
        <span :class="forbiddenBadgeClass">
          {{ forbiddenLabel }}
        </span>
        <div v-if="validationURL" class="account-usage-cell__inline-row">
          <a
            :href="validationURL"
            target="_blank"
            rel="noopener noreferrer"
            class="account-usage-cell__link account-usage-cell__link--compact"
            :title="t('admin.accounts.openVerification')"
          >
            {{ t('admin.accounts.openVerification') }}
          </a>
          <button
            type="button"
            class="account-usage-cell__subtle-action account-usage-cell__subtle-action--compact"
            :title="t('admin.accounts.copyLink')"
            @click="copyValidationURL"
          >
            {{ linkCopied ? t('admin.accounts.linkCopied') : t('admin.accounts.copyLink') }}
          </button>
        </div>
      </div>

      <div v-else-if="needsReauth" class="space-y-1">
        <span class="theme-chip theme-chip--compact theme-chip--brand-orange">
          {{ t('admin.accounts.needsReauth') }}
        </span>
      </div>

      <div v-else-if="usageInfo?.error" class="space-y-1">
        <span class="theme-chip theme-chip--compact theme-chip--warning">
          {{ usageErrorLabel }}
        </span>
      </div>

      <div v-else-if="loading" class="space-y-1.5">
        <div class="account-usage-cell__skeleton-row">
          <div class="account-usage-cell__skeleton-block account-usage-cell__skeleton-block--label animate-pulse"></div>
          <div class="account-usage-cell__skeleton-bar animate-pulse"></div>
          <div class="account-usage-cell__skeleton-block account-usage-cell__skeleton-block--label animate-pulse"></div>
        </div>
      </div>

      <div v-else-if="error" class="account-usage-cell__error text-xs">
        {{ error }}
      </div>

      <div v-else-if="hasAntigravityQuotaFromAPI" class="space-y-1">
        <UsageProgressBar
          v-if="antigravity3ProUsageFromAPI !== null"
          :label="t('admin.accounts.usageWindow.gemini3Pro')"
          :utilization="antigravity3ProUsageFromAPI.utilization"
          :resets-at="antigravity3ProUsageFromAPI.resetTime"
          color="indigo"
        />

        <UsageProgressBar
          v-if="antigravity3FlashUsageFromAPI !== null"
          :label="t('admin.accounts.usageWindow.gemini3Flash')"
          :utilization="antigravity3FlashUsageFromAPI.utilization"
          :resets-at="antigravity3FlashUsageFromAPI.resetTime"
          color="emerald"
        />

        <UsageProgressBar
          v-if="antigravity3ImageUsageFromAPI !== null"
          :label="t('admin.accounts.usageWindow.gemini3Image')"
          :utilization="antigravity3ImageUsageFromAPI.utilization"
          :resets-at="antigravity3ImageUsageFromAPI.resetTime"
          color="purple"
        />

        <UsageProgressBar
          v-if="antigravityClaudeUsageFromAPI !== null"
          :label="t('admin.accounts.usageWindow.claude')"
          :utilization="antigravityClaudeUsageFromAPI.utilization"
          :resets-at="antigravityClaudeUsageFromAPI.resetTime"
          color="amber"
        />

        <div v-if="aiCreditsDisplay" class="account-usage-cell__credits account-usage-cell__credits--compact">
          💳 {{ t('admin.accounts.aiCreditsBalance') }}: {{ aiCreditsDisplay }}
        </div>
      </div>
      <div v-else-if="aiCreditsDisplay" class="account-usage-cell__credits account-usage-cell__credits--compact">
        💳 {{ t('admin.accounts.aiCreditsBalance') }}: {{ aiCreditsDisplay }}
      </div>
      <div v-else class="account-usage-cell__empty text-xs">-</div>
    </template>

    <template v-else-if="account.platform === 'gemini'">
      <div v-if="geminiAuthTypeLabel" class="account-usage-cell__inline-row account-usage-cell__inline-row--with-margin">
        <span :class="geminiTierClass">
          {{ geminiAuthTypeLabel }}
        </span>
        <span class="group relative cursor-help">
          <svg
            class="account-usage-cell__help-icon h-3.5 w-3.5"
            fill="currentColor"
            viewBox="0 0 20 20"
          >
            <path
              fill-rule="evenodd"
              d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-8-3a1 1 0 00-.867.5 1 1 0 11-1.731-1A3 3 0 0113 8a3.001 3.001 0 01-2 2.83V11a1 1 0 11-2 0v-1a1 1 0 011-1 1 1 0 100-2zm0 8a1 1 0 100-2 1 1 0 000 2z"
              clip-rule="evenodd"
            />
          </svg>
          <span
            class="account-usage-cell__tooltip"
          >
            <div class="account-usage-cell__tooltip-title font-semibold">
              {{ t('admin.accounts.gemini.quotaPolicy.title') }}
            </div>
            <div class="account-usage-cell__tooltip-note account-usage-cell__tooltip-note--spaced">
              {{ t('admin.accounts.gemini.quotaPolicy.note') }}
            </div>
            <div class="space-y-1">
              <div><strong>{{ geminiQuotaPolicyChannel }}:</strong></div>
              <div class="account-usage-cell__tooltip-detail">• {{ geminiQuotaPolicyLimits }}</div>
              <div class="account-usage-cell__tooltip-link-row">
                <a
                  :href="geminiQuotaPolicyDocsUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="account-usage-cell__tooltip-link underline"
                >
                  {{ t('admin.accounts.gemini.quotaPolicy.columns.docs') }} →
                </a>
              </div>
            </div>
          </span>
        </span>
      </div>

      <div class="space-y-1">
        <div v-if="loading" class="space-y-1">
          <div class="account-usage-cell__skeleton-row">
            <div class="account-usage-cell__skeleton-block account-usage-cell__skeleton-block--label animate-pulse"></div>
            <div class="account-usage-cell__skeleton-bar animate-pulse"></div>
            <div class="account-usage-cell__skeleton-block account-usage-cell__skeleton-block--label animate-pulse"></div>
          </div>
        </div>
        <div v-else-if="error" class="account-usage-cell__error text-xs">
          {{ error }}
        </div>
        <div v-else-if="geminiUsageAvailable" class="space-y-1">
          <UsageProgressBar
            v-for="bar in geminiUsageBars"
            :key="bar.key"
            :label="bar.label"
            :utilization="bar.utilization"
            :resets-at="bar.resetsAt"
            :window-stats="bar.windowStats"
            :color="bar.color"
          />
          <p class="account-usage-cell__muted-note account-usage-cell__muted-note--compact italic">
            * {{ t('admin.accounts.gemini.quotaPolicy.simulatedNote') || 'Simulated quota' }}
          </p>
        </div>
        <div v-else class="account-usage-cell__empty text-xs">
          {{ t('admin.accounts.gemini.rateLimit.unlimited') }}
        </div>
      </div>
    </template>

    <template v-else>
      <div class="account-usage-cell__empty text-xs">-</div>
    </template>
  </div>

  <div v-else>
    <AccountQuotaInfo v-if="account.platform === 'gemini'" :account="account" />
    <div v-else class="space-y-1">
      <div v-if="todayStats" class="mb-0.5 flex items-center">
        <div class="account-usage-cell__stats-row">
          <span class="account-usage-cell__stat-pill">
            {{ formatKeyRequests }} req
          </span>
          <span class="account-usage-cell__stat-pill">
            {{ formatKeyTokens }}
          </span>
          <span class="account-usage-cell__stat-pill" :title="t('usage.accountBilled')">
            A ${{ formatKeyCost }}
          </span>
          <span
            v-if="todayStats.user_cost != null"
            class="account-usage-cell__stat-pill"
            :title="t('usage.userBilled')"
          >
            U ${{ formatKeyUserCost }}
          </span>
        </div>
      </div>

      <div v-else-if="todayStatsLoading" class="mb-0.5 flex items-center gap-1">
        <div class="account-usage-cell__skeleton-block account-usage-cell__skeleton-block--wide animate-pulse"></div>
        <div class="account-usage-cell__skeleton-block account-usage-cell__skeleton-block--bar animate-pulse"></div>
        <div class="account-usage-cell__skeleton-block account-usage-cell__skeleton-block--cost animate-pulse"></div>
      </div>

      <UsageProgressBar
        v-if="quotaDailyBar"
        label="1d"
        :utilization="quotaDailyBar.utilization"
        :resets-at="quotaDailyBar.resetsAt"
        color="indigo"
      />
      <UsageProgressBar
        v-if="quotaWeeklyBar"
        label="7d"
        :utilization="quotaWeeklyBar.utilization"
        :resets-at="quotaWeeklyBar.resetsAt"
        color="emerald"
      />
      <UsageProgressBar
        v-if="quotaTotalBar"
        label="total"
        :utilization="quotaTotalBar.utilization"
        color="purple"
      />

      <div v-if="!todayStats && !todayStatsLoading && !hasApiKeyQuota" class="account-usage-cell__empty text-xs">
        -
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { Account, WindowStats } from '@/types'
import UsageProgressBar from './UsageProgressBar.vue'
import AccountQuotaInfo from './AccountQuotaInfo.vue'
import { useAccountUsageCellState } from './useAccountUsageCellState'

const props = withDefaults(
  defineProps<{
    account: Account
    todayStats?: WindowStats | null
    todayStatsLoading?: boolean
    manualRefreshToken?: number
  }>(),
  {
    todayStats: null,
    todayStatsLoading: false,
    manualRefreshToken: 0
  }
)

const { t } = useI18n()
const {
  activeQueryLoading,
  aiCreditsDisplay,
  antigravity3FlashUsageFromAPI,
  antigravity3ImageUsageFromAPI,
  antigravity3ProUsageFromAPI,
  antigravityClaudeUsageFromAPI,
  antigravityTierClass,
  antigravityTierLabel,
  copyValidationURL,
  error,
  forbiddenBadgeClass,
  forbiddenLabel,
  formatKeyCost,
  formatKeyRequests,
  formatKeyTokens,
  formatKeyUserCost,
  geminiAuthTypeLabel,
  geminiQuotaPolicyChannel,
  geminiQuotaPolicyDocsUrl,
  geminiQuotaPolicyLimits,
  geminiTierClass,
  geminiUsageAvailable,
  geminiUsageBars,
  hasAntigravityQuotaFromAPI,
  hasApiKeyQuota,
  hasIneligibleTiers,
  hasOpenAIUsageFallback,
  isForbidden,
  linkCopied,
  loadActiveUsage,
  loading,
  needsReauth,
  quotaDailyBar,
  quotaTotalBar,
  quotaWeeklyBar,
  showUsageWindows,
  usageErrorLabel,
  usageInfo,
  validationURL
} = useAccountUsageCellState(props, t)
</script>

<style scoped>
.account-usage-cell__skeleton-block,
.account-usage-cell__skeleton-bar {
  background: color-mix(in srgb, var(--theme-page-border) 82%, var(--theme-surface));
}

.account-usage-cell__skeleton-block {
  height: 0.75rem;
  border-radius: var(--theme-button-radius);
}

.account-usage-cell__skeleton-block--label {
  width: var(--theme-account-usage-skeleton-label-width);
}

.account-usage-cell__skeleton-block--wide {
  width: 2.5rem;
}

.account-usage-cell__skeleton-block--bar {
  width: 2rem;
}

.account-usage-cell__skeleton-block--cost {
  width: 3rem;
}

.account-usage-cell__skeleton-bar {
  width: var(--theme-account-usage-skeleton-bar-width);
  height: 0.375rem;
  border-radius: 999px;
}

.account-usage-cell__skeleton-row,
.account-usage-cell__inline-row,
.account-usage-cell__stats-row {
  display: flex;
  align-items: center;
  gap: var(--theme-account-usage-inline-gap);
}

.account-usage-cell__inline-row--compact {
  margin-top: 0.125rem;
  gap: calc(var(--theme-account-usage-inline-gap) * 1.5);
}

.account-usage-cell__inline-row--with-margin {
  margin-bottom: 0.25rem;
}

.account-usage-cell__empty {
  color: color-mix(in srgb, var(--theme-page-muted) 78%, transparent);
}

.account-usage-cell__error {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.account-usage-cell__warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.account-usage-cell__warning--truncated {
  max-width: var(--theme-account-usage-warning-max-width);
}

.account-usage-cell__muted-note,
.account-usage-cell__credits,
.account-usage-cell__stats-row {
  color: var(--theme-page-muted);
}

.account-usage-cell__muted-note--compact,
.account-usage-cell__credits--compact,
.account-usage-cell__link--compact,
.account-usage-cell__subtle-action--compact,
.account-usage-cell__stats-row {
  font-size: 0.625rem;
}

.account-usage-cell__credits--compact,
.account-usage-cell__muted-note--compact {
  margin-top: 0.25rem;
  line-height: 1.25;
}

.account-usage-cell__action {
  display: inline-flex;
  align-items: center;
  gap: 0.125rem;
  border-radius: var(--theme-button-radius);
  padding: var(--theme-account-usage-action-padding-y) var(--theme-account-usage-action-padding-x);
  font-size: 0.625rem;
  font-weight: 500;
  transition: background-color 0.2s ease, color 0.2s ease;
  color: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-page-text));
}

.account-usage-cell__action:hover {
  background: color-mix(in srgb, var(--theme-accent-soft) 88%, var(--theme-surface));
}

.account-usage-cell__danger-icon {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.account-usage-cell__help-icon {
  color: color-mix(in srgb, var(--theme-page-muted) 74%, transparent);
}

.account-usage-cell__help-icon:hover {
  color: var(--theme-page-text);
}

.account-usage-cell__tooltip {
  pointer-events: none;
  position: absolute;
  left: 0;
  top: 100%;
  z-index: 50;
  width: var(--theme-account-usage-tooltip-width);
  margin-top: 0.25rem;
  border-radius: var(--theme-tooltip-radius);
  padding: var(--theme-tooltip-padding);
  white-space: normal;
  word-break: break-word;
  font-size: 0.75rem;
  line-height: 1.5;
  opacity: 0;
  box-shadow: var(--theme-dropdown-shadow);
  transition: opacity 0.2s ease;
  border: 1px solid color-mix(in srgb, var(--theme-surface-contrast) 16%, transparent);
  background: color-mix(in srgb, var(--theme-surface-contrast) 94%, var(--theme-surface));
  color: var(--theme-surface-contrast-text);
}

.group:hover .account-usage-cell__tooltip {
  opacity: 1;
}

.account-usage-cell__tooltip-title {
  margin-bottom: 0.25rem;
  color: var(--theme-surface-contrast-text);
}

.account-usage-cell__tooltip-note {
  color: color-mix(in srgb, var(--theme-surface-contrast-text) 68%, transparent);
}

.account-usage-cell__tooltip-note--spaced {
  margin-bottom: 0.5rem;
}

.account-usage-cell__tooltip-detail {
  padding-left: 0.5rem;
}

.account-usage-cell__tooltip-link-row {
  margin-top: 0.5rem;
}

.account-usage-cell__tooltip-link {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 74%, var(--theme-surface-contrast-text));
}

.account-usage-cell__tooltip-link:hover {
  color: var(--theme-surface-contrast-text);
}

.account-usage-cell__link {
  color: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-page-text));
}

.account-usage-cell__link:hover {
  color: color-mix(in srgb, var(--theme-accent-strong) 22%, var(--theme-accent) 78%);
}

.account-usage-cell__subtle-action {
  color: var(--theme-page-muted);
}

.account-usage-cell__subtle-action:hover {
  color: var(--theme-page-text);
}

.account-usage-cell__stat-pill {
  border-radius: var(--theme-button-radius);
  padding: var(--theme-account-usage-pill-padding-y) var(--theme-account-usage-pill-padding-x);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}
</style>
