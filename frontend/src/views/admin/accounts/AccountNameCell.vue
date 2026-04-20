<template>
  <div class="flex flex-col">
    <span class="theme-text-strong font-medium">{{ account.name }}</span>
    <span
      v-if="emailAddress"
      class="theme-text-muted account-name-cell__email truncate text-xs"
      :title="emailAddress"
    >
      {{ emailAddress }}
    </span>
    <span
      v-if="authFingerprint"
      class="theme-text-muted truncate font-mono text-[11px]"
      :title="authFingerprint"
    >
      {{ authFingerprint }}
    </span>
    <span
      v-if="grokSyncSummary"
      class="theme-text-muted truncate text-[11px]"
      :title="grokSyncSummary"
    >
      {{ grokSyncSummary }}
    </span>
    <span
      v-if="grokRecentErrorSummary"
      class="account-name-cell__runtime-error truncate text-[11px]"
      :title="grokRecentErrorSummary"
    >
      {{ grokRecentErrorSummary }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Account } from '@/types'
import { getGrokAccountRuntime } from '@/utils/grokAccountRuntime'
import { formatRelativeTime } from '@/utils/format'

const props = defineProps<{
  account: Account
}>()

const { t } = useI18n()

const grokRuntime = computed(() => getGrokAccountRuntime(props.account))

const emailAddress = computed(() => {
  const value = props.account.extra?.email_address
  return typeof value === 'string' ? value : null
})

const authFingerprint = computed(() => grokRuntime.value?.authFingerprint ?? null)

const grokSyncSummary = computed(() => {
  const runtime = grokRuntime.value
  if (!runtime?.hasState) {
    return null
  }

  const segments: string[] = []
  if (runtime.sync.lastSyncAt) {
    segments.push(`${t('admin.accounts.grok.runtime.lastSyncAt')}: ${formatRelativeTime(runtime.sync.lastSyncAt)}`)
  }
  if (runtime.sync.lastProbeAt) {
    segments.push(`${t('admin.accounts.grok.runtime.lastProbeAt')}: ${formatRelativeTime(runtime.sync.lastProbeAt)}`)
  }

  return segments.length > 0 ? segments.join(' · ') : null
})

const grokRecentErrorSummary = computed(() => {
  const runtime = grokRuntime.value
  if (!runtime?.hasState) {
    return null
  }

  if (runtime.sync.lastProbeError) {
    return `${t('admin.accounts.grok.runtime.lastProbeError')}: ${runtime.sync.lastProbeError}`
  }
  if (runtime.runtime.lastFailReason) {
    return `${t('admin.accounts.grok.runtime.lastRuntimeError')}: ${runtime.runtime.lastFailReason}`
  }
  return null
})
</script>

<style scoped>
.account-name-cell__email {
  max-width: var(--theme-account-name-secondary-max-width);
}

.account-name-cell__runtime-error {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}
</style>
