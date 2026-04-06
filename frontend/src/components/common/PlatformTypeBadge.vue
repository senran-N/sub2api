<template>
  <div class="inline-flex flex-col gap-0.5 text-xs font-medium">
    <!-- Row 1: Platform + Type -->
    <div class="inline-flex items-center overflow-hidden">
      <span :class="['theme-chip theme-chip--segment platform-type-badge__segment platform-type-badge__segment--primary inline-flex items-center gap-1', platformClass]">
        <PlatformIcon :platform="platform" size="xs" />
        <span>{{ platformLabel }}</span>
      </span>
      <span :class="['theme-chip theme-chip--segment platform-type-badge__segment platform-type-badge__segment--secondary inline-flex items-center gap-1', typeClass]">
        <!-- OAuth icon -->
        <svg
          v-if="type === 'oauth'"
          class="h-3 w-3"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"
          />
        </svg>
        <!-- Setup Token icon -->
        <Icon v-else-if="type === 'setup-token'" name="shield" size="xs" />
        <!-- API Key icon -->
        <Icon v-else name="key" size="xs" />
        <span>{{ typeLabel }}</span>
      </span>
    </div>
    <!-- Row 2: Plan type + Privacy mode (only if either exists) -->
    <div v-if="planLabel || privacyBadge" class="inline-flex items-center overflow-hidden">
      <span v-if="planLabel" :class="['theme-chip theme-chip--segment platform-type-badge__segment platform-type-badge__segment--secondary inline-flex items-center gap-1', planBadgeClass]">
        <span>{{ planLabel }}</span>
      </span>
      <span
        v-if="privacyBadge"
        :class="['theme-chip theme-chip--segment platform-type-badge__segment platform-type-badge__segment--secondary inline-flex items-center gap-1', privacyBadge.class]"
        :title="privacyBadge.title"
      >
        <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" :d="privacyBadge.icon" />
        </svg>
        <span>{{ privacyBadge.label }}</span>
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AccountPlatform, AccountType } from '@/types'
import PlatformIcon from './PlatformIcon.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

interface Props {
  platform: AccountPlatform
  type: AccountType
  planType?: string
  privacyMode?: string
}

const props = defineProps<Props>()

const platformLabel = computed(() => {
  if (props.platform === 'anthropic') return 'Anthropic'
  if (props.platform === 'openai') return 'OpenAI'
  if (props.platform === 'antigravity') return 'Antigravity'
  if (props.platform === 'sora') return 'Sora'
  return 'Gemini'
})

const typeLabel = computed(() => {
  switch (props.type) {
    case 'oauth':
      return 'OAuth'
    case 'setup-token':
      return 'Token'
    case 'apikey':
      return 'Key'
    case 'bedrock':
      return 'AWS'
    default:
      return props.type
  }
})

const planLabel = computed(() => {
  if (!props.planType) return ''
  const lower = props.planType.toLowerCase()
  switch (lower) {
    case 'plus':
      return 'Plus'
    case 'team':
      return 'Team'
    case 'chatgptpro':
    case 'pro':
      return 'Pro'
    case 'free':
      return 'Free'
    case 'abnormal':
      return t('admin.accounts.subscriptionAbnormal')
    default:
      return props.planType
  }
})

const platformClass = computed(() => {
  if (props.platform === 'anthropic') {
    return 'theme-chip--brand-orange'
  }
  if (props.platform === 'openai') {
    return 'theme-chip--success'
  }
  if (props.platform === 'antigravity') {
    return 'theme-chip--brand-purple'
  }
  if (props.platform === 'sora') {
    return 'theme-chip--brand-rose'
  }
  return 'theme-chip--info'
})

const typeClass = computed(() => {
  if (props.platform === 'anthropic') {
    return 'theme-chip--warning'
  }
  if (props.platform === 'openai') {
    return 'theme-chip--accent'
  }
  if (props.platform === 'antigravity') {
    return 'theme-chip--brand-purple'
  }
  if (props.platform === 'sora') {
    return 'theme-chip--brand-rose'
  }
  return 'theme-chip--info'
})

const planBadgeClass = computed(() => {
  if (props.planType && props.planType.toLowerCase() === 'abnormal') {
    return 'theme-chip--danger'
  }
  return typeClass.value
})

// Privacy badge — shows different states for OpenAI/Antigravity OAuth privacy setting
const privacyBadge = computed(() => {
  if (props.type !== 'oauth' || !props.privacyMode) return null
  // 支持 OpenAI 和 Antigravity 平台
  if (props.platform !== 'openai' && props.platform !== 'antigravity') return null

  const shieldCheck = 'M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z'
  const shieldX = 'M12 9v3.75m0-10.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285zM12 18h.008v.008H12V18z'
  switch (props.privacyMode) {
    case 'training_off':
      return { label: 'Private', icon: shieldCheck, title: t('admin.accounts.privacyTrainingOff'), class: 'theme-chip--success' }
    case 'training_set_cf_blocked':
      return { label: 'CF', icon: shieldX, title: t('admin.accounts.privacyCfBlocked'), class: 'theme-chip--warning' }
    case 'training_set_failed':
      return { label: 'Fail', icon: shieldX, title: t('admin.accounts.privacyFailed'), class: 'theme-chip--danger' }
    case 'privacy_set':
      return { label: 'Private', icon: shieldCheck, title: t('admin.accounts.privacyAntigravitySet'), class: 'theme-chip--success' }
    case 'privacy_set_failed':
      return { label: 'Fail', icon: shieldX, title: t('admin.accounts.privacyAntigravityFailed'), class: 'theme-chip--danger' }
    default:
      return null
  }
})
</script>

<style scoped>
.platform-type-badge__segment {
  min-height: var(--theme-platform-type-badge-min-height);
  padding-top: var(--theme-platform-type-badge-padding-y);
  padding-bottom: var(--theme-platform-type-badge-padding-y);
}

.platform-type-badge__segment--primary {
  padding-left: var(--theme-platform-type-badge-primary-padding-x);
  padding-right: var(--theme-platform-type-badge-primary-padding-x);
}

.platform-type-badge__segment--secondary {
  padding-left: var(--theme-platform-type-badge-secondary-padding-x);
  padding-right: var(--theme-platform-type-badge-secondary-padding-x);
}
</style>
