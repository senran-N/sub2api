<template>
  <div v-if="shouldShowQuota">
    <div class="mb-1 flex items-center gap-1">
      <span :class="tierBadgeClass">
        {{ tierLabel }}
      </span>
    </div>

    <div class="account-quota-info__status text-xs">
      <span v-if="!isRateLimited">
        {{ t('admin.accounts.gemini.rateLimit.unlimited') }}
      </span>
      <span v-else :class="limitStatusClass">
        {{ t('admin.accounts.gemini.rateLimit.limited', { time: resetCountdown }) }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Account, GeminiCredentials } from '@/types'

const props = defineProps<{
  account: Account
}>()

const { t } = useI18n()

const now = ref(new Date())
let timer: ReturnType<typeof setInterval> | null = null

type QuotaTone = 'neutral' | 'info' | 'brand'

const buildTierBadgeClass = (tone: QuotaTone) => [
  'theme-chip',
  'theme-chip--regular',
  'account-quota-info__tier-badge',
  `account-quota-info__tier-badge--${tone}`
]

const isCodeAssist = computed(() => {
  const creds = props.account.credentials as GeminiCredentials | undefined
  return creds?.oauth_type === 'code_assist' || (!creds?.oauth_type && !!creds?.project_id)
})

const isGoogleOne = computed(() => {
  const creds = props.account.credentials as GeminiCredentials | undefined
  return creds?.oauth_type === 'google_one'
})

const shouldShowQuota = computed(() => {
  return props.account.platform === 'gemini'
})

const tierLabel = computed(() => {
  const creds = props.account.credentials as GeminiCredentials | undefined

  if (isCodeAssist.value) {
    const tier = (creds?.tier_id || '').toString().trim().toLowerCase()
    if (tier === 'gcp_enterprise') return 'GCP Enterprise'
    if (tier === 'gcp_standard') return 'GCP Standard'
    const upper = (creds?.tier_id || '').toString().trim().toUpperCase()
    if (upper.includes('ULTRA') || upper.includes('ENTERPRISE')) return 'GCP Enterprise'
    if (upper) return `GCP ${upper}`
    return 'GCP'
  }

  if (isGoogleOne.value) {
    const tier = (creds?.tier_id || '').toString().trim().toLowerCase()
    if (tier === 'google_ai_ultra') return 'Google AI Ultra'
    if (tier === 'google_ai_pro') return 'Google AI Pro'
    if (tier === 'google_one_free') return 'Google One Free'
    const upper = (creds?.tier_id || '').toString().trim().toUpperCase()
    if (upper === 'AI_PREMIUM') return 'Google AI Pro'
    if (upper === 'GOOGLE_ONE_UNLIMITED') return 'Google AI Ultra'
    if (upper) return `Google One ${upper}`
    return 'Google One'
  }

  const tier = (creds?.tier_id || '').toString().trim().toLowerCase()
  if (tier === 'aistudio_paid') return 'AI Studio Pay-as-you-go'
  if (tier === 'aistudio_free') return 'AI Studio Free Tier'
  return 'AI Studio'
})

const tierBadgeClass = computed(() => {
  const creds = props.account.credentials as GeminiCredentials | undefined

  if (isCodeAssist.value) {
    const tier = (creds?.tier_id || '').toString().trim().toLowerCase()
    if (tier === 'gcp_enterprise') return buildTierBadgeClass('brand')
    if (tier === 'gcp_standard') return buildTierBadgeClass('info')
    const upper = (creds?.tier_id || '').toString().trim().toUpperCase()
    if (upper.includes('ULTRA') || upper.includes('ENTERPRISE')) return buildTierBadgeClass('brand')
    return buildTierBadgeClass('info')
  }

  if (isGoogleOne.value) {
    const tier = (creds?.tier_id || '').toString().trim().toLowerCase()
    if (tier === 'google_ai_ultra') return buildTierBadgeClass('brand')
    if (tier === 'google_ai_pro') return buildTierBadgeClass('info')
    if (tier === 'google_one_free') return buildTierBadgeClass('neutral')
    const upper = (creds?.tier_id || '').toString().trim().toUpperCase()
    if (upper === 'GOOGLE_ONE_UNLIMITED') return buildTierBadgeClass('brand')
    if (upper === 'AI_PREMIUM') return buildTierBadgeClass('info')
    return buildTierBadgeClass('neutral')
  }

  const tier = (creds?.tier_id || '').toString().trim().toLowerCase()
  if (tier === 'aistudio_paid') return buildTierBadgeClass('info')
  if (tier === 'aistudio_free') return buildTierBadgeClass('neutral')
  return buildTierBadgeClass('info')
})

const isRateLimited = computed(() => {
  if (!props.account.rate_limit_reset_at) return false
  const resetTime = Date.parse(props.account.rate_limit_reset_at)
  if (Number.isNaN(resetTime)) return false
  return resetTime > now.value.getTime()
})

const resetCountdown = computed(() => {
  if (!props.account.rate_limit_reset_at) return ''
  const resetTime = Date.parse(props.account.rate_limit_reset_at)
  if (Number.isNaN(resetTime)) return '-'

  const diffMs = resetTime - now.value.getTime()
  if (diffMs <= 0) return t('admin.accounts.gemini.rateLimit.now')

  const diffSeconds = Math.floor(diffMs / 1000)
  const diffMinutes = Math.floor(diffSeconds / 60)
  const diffHours = Math.floor(diffMinutes / 60)

  if (diffMinutes < 1) return `${diffSeconds}s`
  if (diffHours < 1) {
    const secs = diffSeconds % 60
    return `${diffMinutes}m ${secs}s`
  }
  const mins = diffMinutes % 60
  return `${diffHours}h ${mins}m`
})

const isUrgent = computed(() => {
  if (!props.account.rate_limit_reset_at) return false
  const resetTime = Date.parse(props.account.rate_limit_reset_at)
  if (Number.isNaN(resetTime)) return false

  const diffMs = resetTime - now.value.getTime()
  return diffMs > 0 && diffMs < 60000
})

const limitStatusClass = computed(() => [
  'account-quota-info__limit-status',
  'font-medium',
  isUrgent.value
    ? 'account-quota-info__limit-status--danger animate-pulse'
    : 'account-quota-info__limit-status--warning'
])

watch(
  () => isRateLimited.value,
  (limited) => {
    if (limited && !timer) {
      timer = setInterval(() => {
        now.value = new Date()
      }, 1000)
    } else if (!limited && timer) {
      clearInterval(timer)
      timer = null
    }
  },
  { immediate: true }
)

onUnmounted(() => {
  if (timer !== null) {
    clearInterval(timer)
    timer = null
  }
})
</script>

<style scoped>
.account-quota-info__status {
  color: color-mix(in srgb, var(--theme-page-muted) 78%, var(--theme-surface));
}

.account-quota-info__tier-badge--neutral {
  --theme-chip-bg: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  --theme-chip-fg: var(--theme-page-muted);
}

.account-quota-info__tier-badge--info {
  --theme-chip-bg: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface));
  --theme-chip-fg: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
  --theme-chip-border: color-mix(in srgb, rgb(var(--theme-info-rgb)) 18%, var(--theme-card-border));
}

.account-quota-info__tier-badge--brand {
  --theme-chip-bg: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 10%, var(--theme-surface));
  --theme-chip-fg: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 84%, var(--theme-page-text));
  --theme-chip-border: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 18%, var(--theme-card-border));
}

.account-quota-info__limit-status--warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.account-quota-info__limit-status--danger {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}
</style>
