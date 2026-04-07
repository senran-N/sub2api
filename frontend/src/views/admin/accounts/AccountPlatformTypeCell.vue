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

const tierLabel = computed(() => getAccountAntigravityTierLabel(props.account, t))
const tierClass = computed(() => getAccountAntigravityTierClass(props.account))
</script>
