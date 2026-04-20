<template>
  <div class="flex flex-wrap items-center gap-1">
    <PlatformTypeBadge
      :platform="account.platform"
      :type="account.type"
      :plan-type="planType"
      :privacy-mode="privacyMode"
    />
    <span
      v-if="tierLabel"
      :class="[
        'theme-chip theme-chip--compact inline-block',
        tierClass
      ]"
    >
      {{ tierLabel }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import PlatformTypeBadge from '@/components/common/PlatformTypeBadge.vue'
import type { Account } from '@/types'
import { getGrokAccountRuntime } from '@/utils/grokAccountRuntime'
import {
  getAccountAntigravityTierClass,
  getAccountAntigravityTierLabel
} from './accountsView'

const props = defineProps<{
  account: Account
}>()

const { t } = useI18n()

const planType = computed(() => {
  const value = props.account.credentials?.plan_type
  return typeof value === 'string' ? value : undefined
})

const privacyMode = computed(() => {
  const value = props.account.extra?.privacy_mode
  return typeof value === 'string' ? value : undefined
})

const grokRuntime = computed(() => getGrokAccountRuntime(props.account))

const tierLabel = computed(() => {
  if (props.account.platform === 'grok') {
    const tier = grokRuntime.value?.tier.normalized ?? 'unknown'
    if (!grokRuntime.value?.hasState && tier === 'unknown') {
      return null
    }
    return t(`admin.accounts.grok.runtime.tiers.${tier}`)
  }

  return getAccountAntigravityTierLabel(props.account, t)
})

const tierClass = computed(() => {
  if (props.account.platform !== 'grok') {
    return getAccountAntigravityTierClass(props.account)
  }

  switch (grokRuntime.value?.tier.normalized) {
    case 'basic':
      return 'theme-chip--info'
    case 'heavy':
      return 'theme-chip--brand-orange'
    case 'super':
      return 'theme-chip--brand-purple'
    default:
      return 'theme-chip--neutral'
  }
})
</script>
