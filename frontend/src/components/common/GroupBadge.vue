<template>
  <span
    :class="[
      'theme-chip theme-chip--regular group-badge',
      badgeClass
    ]"
  >
    <!-- Platform logo -->
    <PlatformIcon v-if="platform" :platform="platform" size="sm" />
    <!-- Group name -->
    <span class="truncate">{{ name }}</span>
    <!-- Right side label -->
    <span v-if="showLabel" :class="['theme-chip theme-chip--compact group-badge__label', labelClass]">
      <template v-if="hasCustomRate">
        <!-- 原倍率删除线 + 专属倍率高亮 -->
        <span class="mr-0.5 line-through opacity-50">{{ rateMultiplier }}x</span>
        <span class="font-bold">{{ userRateMultiplier }}x</span>
      </template>
      <template v-else>
        {{ labelText }}
      </template>
    </span>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SubscriptionType, GroupPlatform } from '@/types'
import PlatformIcon from './PlatformIcon.vue'

interface Props {
  name: string
  platform?: GroupPlatform
  subscriptionType?: SubscriptionType
  rateMultiplier?: number
  userRateMultiplier?: number | null // 用户专属倍率
  showRate?: boolean
  daysRemaining?: number | null // 剩余天数（订阅类型时使用）
}

const props = withDefaults(defineProps<Props>(), {
  subscriptionType: 'standard',
  showRate: true,
  daysRemaining: null,
  userRateMultiplier: null
})

const { t } = useI18n()

const isSubscription = computed(() => props.subscriptionType === 'subscription')

// 是否有专属倍率（且与默认倍率不同）
const hasCustomRate = computed(() => {
  return (
    props.userRateMultiplier !== null &&
    props.userRateMultiplier !== undefined &&
    props.rateMultiplier !== undefined &&
    props.userRateMultiplier !== props.rateMultiplier
  )
})

// 是否显示右侧标签
const showLabel = computed(() => {
  if (!props.showRate) return false
  // 订阅类型：显示天数或"订阅"
  if (isSubscription.value) return true
  // 标准类型：显示倍率（包括专属倍率）
  return props.rateMultiplier !== undefined || hasCustomRate.value
})

// Label text
const labelText = computed(() => {
  if (isSubscription.value) {
    // 如果有剩余天数，显示天数
    if (props.daysRemaining !== null && props.daysRemaining !== undefined) {
      if (props.daysRemaining <= 0) {
        return t('admin.users.expired')
      }
      return t('admin.users.daysRemaining', { days: props.daysRemaining })
    }
    // 否则显示"订阅"
    return t('groups.subscription')
  }
  return props.rateMultiplier !== undefined ? `${props.rateMultiplier}x` : ''
})

// Label style based on type and days remaining
const labelClass = computed(() => {
  if (!isSubscription.value) {
    return 'theme-chip--neutral group-badge__label--default'
  }

  if (props.daysRemaining !== null && props.daysRemaining !== undefined) {
    if (props.daysRemaining <= 0 || props.daysRemaining <= 3) {
      return 'theme-chip--danger'
    }
    if (props.daysRemaining <= 7) {
      return 'theme-chip--warning'
    }
  }

  if (props.platform === 'anthropic') {
    return 'theme-chip--brand-orange'
  }
  if (props.platform === 'openai') {
    return 'theme-chip--success'
  }
  if (props.platform === 'grok') {
    return 'theme-chip--brand-rose'
  }
  if (props.platform === 'gemini') {
    return 'theme-chip--info'
  }
  return 'theme-chip--brand-purple'
})

// Badge color based on platform and subscription type
const badgeClass = computed(() => {
  if (props.platform === 'anthropic') {
    return isSubscription.value
      ? 'theme-chip--brand-orange'
      : 'theme-chip--warning'
  } else if (props.platform === 'openai') {
    return isSubscription.value
      ? 'theme-chip--success'
      : 'theme-chip--accent'
  }
  if (props.platform === 'grok') {
    return isSubscription.value
      ? 'theme-chip--brand-rose'
      : 'theme-chip--accent'
  }
  if (props.platform === 'gemini') {
    return isSubscription.value
      ? 'theme-chip--info'
      : 'theme-chip--accent'
  }
  return isSubscription.value
    ? 'theme-chip--brand-purple'
    : 'theme-chip--accent'
})
</script>

<style scoped>
.group-badge {
  max-width: 100%;
}

.group-badge__label {
  font-weight: 600;
}

.group-badge__label--default {
  border-style: dashed;
}
</style>
