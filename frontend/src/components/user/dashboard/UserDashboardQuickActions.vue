<template>
  <div class="card">
    <div class="user-dashboard-quick-actions__header">
      <h2 class="user-dashboard-quick-actions__title">{{ t('dashboard.quickActions') }}</h2>
    </div>
    <div class="user-dashboard-quick-actions__body space-y-3">
      <button @click="router.push('/keys')" :class="getActionCardClasses()">
        <div :class="getActionIconShellClasses('accent')">
          <Icon name="key" size="lg" :class="getActionIconClasses('accent')" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="user-dashboard-quick-actions__action-title">{{ t('dashboard.createApiKey') }}</p>
          <p class="user-dashboard-quick-actions__action-description">{{ t('dashboard.generateNewKey') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          :class="getActionChevronClasses('accent')"
        />
      </button>

      <button @click="router.push('/usage')" :class="getActionCardClasses()">
        <div :class="getActionIconShellClasses('success')">
          <Icon name="chart" size="lg" :class="getActionIconClasses('success')" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="user-dashboard-quick-actions__action-title">{{ t('dashboard.viewUsage') }}</p>
          <p class="user-dashboard-quick-actions__action-description">{{ t('dashboard.checkDetailedLogs') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          :class="getActionChevronClasses('success')"
        />
      </button>

      <button @click="router.push('/redeem')" :class="getActionCardClasses()">
        <div :class="getActionIconShellClasses('warning')">
          <Icon name="gift" size="lg" :class="getActionIconClasses('warning')" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="user-dashboard-quick-actions__action-title">{{ t('dashboard.redeemCode') }}</p>
          <p class="user-dashboard-quick-actions__action-description">{{ t('dashboard.addBalanceWithCode') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          :class="getActionChevronClasses('warning')"
        />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

type QuickActionTone = 'accent' | 'success' | 'warning'

const router = useRouter()
const { t } = useI18n()

function joinClassNames(...classNames: Array<string | false | null | undefined>) {
  return classNames.filter(Boolean).join(' ')
}

function getActionCardClasses() {
  return joinClassNames(
    'user-dashboard-quick-actions__action-card group'
  )
}

function getActionIconShellClasses(tone: QuickActionTone) {
  return joinClassNames(
    'user-dashboard-quick-actions__icon-shell',
    `user-dashboard-quick-actions__icon-shell--${tone}`
  )
}

function getActionIconClasses(tone: QuickActionTone) {
  return joinClassNames(
    'user-dashboard-quick-actions__icon',
    `user-dashboard-quick-actions__icon--${tone}`
  )
}

function getActionChevronClasses(tone: QuickActionTone) {
  return joinClassNames(
    'user-dashboard-quick-actions__chevron',
    `user-dashboard-quick-actions__chevron--${tone}`
  )
}
</script>

<style scoped>
.user-dashboard-quick-actions__header {
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  padding: var(--theme-settings-card-panel-padding) calc(var(--theme-settings-card-panel-padding) * 1.5);
}

.user-dashboard-quick-actions__body {
  padding: var(--theme-settings-card-panel-padding);
}

.user-dashboard-quick-actions__title,
.user-dashboard-quick-actions__action-title {
  color: var(--theme-page-text);
}

.user-dashboard-quick-actions__title {
  font-size: 1.125rem;
  font-weight: 600;
}

.user-dashboard-quick-actions__action-card {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 1rem;
  border-radius: calc(var(--theme-surface-radius) + 2px);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  padding: var(--theme-settings-card-panel-padding);
  text-align: left;
  transition:
    background-color 0.2s ease,
    transform 0.2s ease,
    box-shadow 0.2s ease;
}

.user-dashboard-quick-actions__action-card:hover,
.user-dashboard-quick-actions__action-card:focus-visible {
  background: color-mix(in srgb, var(--theme-button-ghost-hover-bg) 90%, var(--theme-surface));
  box-shadow: var(--theme-card-shadow);
  transform: translateY(-1px);
  outline: none;
}

.user-dashboard-quick-actions__icon-shell {
  display: flex;
  height: 3rem;
  width: 3rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: calc(var(--theme-surface-radius) + 2px);
  transition: transform 0.2s ease;
}

.group:hover .user-dashboard-quick-actions__icon-shell {
  transform: scale(1.05);
}

.user-dashboard-quick-actions__icon-shell--accent {
  background: color-mix(in srgb, var(--theme-accent-soft) 90%, var(--theme-surface));
}

.user-dashboard-quick-actions__icon-shell--success {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 12%, var(--theme-surface));
}

.user-dashboard-quick-actions__icon-shell--warning {
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 12%, var(--theme-surface));
}

.user-dashboard-quick-actions__icon--accent,
.user-dashboard-quick-actions__chevron--accent {
  color: var(--theme-accent);
}

.user-dashboard-quick-actions__icon--success,
.user-dashboard-quick-actions__chevron--success {
  color: rgb(var(--theme-success-rgb));
}

.user-dashboard-quick-actions__icon--warning,
.user-dashboard-quick-actions__chevron--warning {
  color: rgb(var(--theme-warning-rgb));
}

.user-dashboard-quick-actions__action-title {
  font-size: 0.875rem;
  font-weight: 600;
}

.user-dashboard-quick-actions__action-description,
.user-dashboard-quick-actions__chevron {
  color: var(--theme-page-muted);
}

.user-dashboard-quick-actions__action-description {
  font-size: 0.75rem;
}

.group:hover .user-dashboard-quick-actions__chevron {
  color: inherit;
}
</style>
