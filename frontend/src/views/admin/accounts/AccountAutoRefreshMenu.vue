<template>
  <div class="account-auto-refresh-menu">
    <button
      class="account-auto-refresh-menu__button flex w-full items-center justify-between text-sm"
      @click="emit('toggle-enabled')"
    >
      <span>{{ t('admin.accounts.enableAutoRefresh') }}</span>
      <Icon
        v-if="enabled"
        name="check"
        size="sm"
        class="account-auto-refresh-menu__check"
      />
    </button>
    <div class="account-auto-refresh-menu__divider my-1 border-t"></div>
    <button
      v-for="seconds in intervals"
      :key="seconds"
      class="account-auto-refresh-menu__button flex w-full items-center justify-between text-sm"
      @click="emit('set-interval', seconds)"
    >
      <span>{{ labelForInterval(seconds) }}</span>
      <Icon
        v-if="selectedIntervalSeconds === seconds"
        name="check"
        size="sm"
        class="account-auto-refresh-menu__check"
      />
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { ACCOUNT_AUTO_REFRESH_INTERVALS } from './accountsList'

type AccountAutoRefreshInterval = (typeof ACCOUNT_AUTO_REFRESH_INTERVALS)[number]

defineProps<{
  enabled: boolean
  intervals: readonly AccountAutoRefreshInterval[]
  selectedIntervalSeconds: AccountAutoRefreshInterval
  labelForInterval: (seconds: AccountAutoRefreshInterval) => string
}>()

const emit = defineEmits<{
  'toggle-enabled': []
  'set-interval': [seconds: AccountAutoRefreshInterval]
}>()

const { t } = useI18n()
</script>

<style scoped>
.account-auto-refresh-menu {
  padding: var(--theme-account-auto-refresh-menu-padding);
}

.account-auto-refresh-menu__button {
  border-radius: var(--theme-account-auto-refresh-menu-button-radius);
  padding: var(--theme-account-auto-refresh-menu-button-padding-y)
    var(--theme-account-auto-refresh-menu-button-padding-x);
  color: var(--theme-page-text);
  transition: background-color 0.2s ease, color 0.2s ease;
}

.account-auto-refresh-menu__button:hover {
  background: var(--theme-dropdown-item-hover-bg);
}

.account-auto-refresh-menu__check {
  color: var(--theme-accent);
}

.account-auto-refresh-menu__divider {
  border-color: color-mix(in srgb, var(--theme-dropdown-border) 88%, transparent);
}
</style>
