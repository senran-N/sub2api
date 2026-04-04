<template>
  <div class="contents">
    <div class="relative" ref="autoRefreshDropdownRef">
      <button
        class="btn btn-secondary px-2 md:px-3"
        :title="t('admin.accounts.autoRefresh')"
        @click="toggleAutoRefreshDropdown"
      >
        <Icon name="refresh" size="sm" :class="[autoRefreshEnabled ? 'animate-spin' : '']" />
        <span class="hidden md:inline">
          {{
            autoRefreshEnabled
              ? t('admin.accounts.autoRefreshCountdown', { seconds: autoRefreshCountdown })
              : t('admin.accounts.autoRefresh')
          }}
        </span>
      </button>
      <div
        v-if="showAutoRefreshDropdown"
        class="absolute right-0 z-50 mt-2 w-56 origin-top-right rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
      >
        <AccountAutoRefreshMenu
          :enabled="autoRefreshEnabled"
          :intervals="autoRefreshIntervals"
          :selected-interval-seconds="autoRefreshIntervalSeconds"
          :label-for-interval="autoRefreshIntervalLabel"
          @toggle-enabled="emit('toggle-auto-refresh-enabled')"
          @set-interval="emit('set-auto-refresh-interval', $event)"
        />
      </div>
    </div>

    <AccountAdminToolsButtons
      @error-passthrough="emit('error-passthrough')"
      @tls-profiles="emit('tls-profiles')"
    />

    <div class="relative" ref="columnDropdownRef">
      <button
        class="btn btn-secondary px-2 md:px-3"
        :title="t('admin.users.columnSettings')"
        @click="toggleColumnDropdown"
      >
        <Icon name="grid" size="sm" class="md:mr-1.5" />
        <span class="hidden md:inline">{{ t('admin.users.columnSettings') }}</span>
      </button>
      <div
        v-if="showColumnDropdown"
        class="absolute right-0 z-50 mt-2 w-48 origin-top-right rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
      >
        <AccountColumnSettingsMenu
          :toggleable-columns="toggleableColumns"
          :is-column-visible="isColumnVisible"
          @toggle-column="emit('toggle-column', $event)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Column } from '@/components/common/types'
import Icon from '@/components/icons/Icon.vue'
import { ACCOUNT_AUTO_REFRESH_INTERVALS } from '../accountsList'
import AccountAdminToolsButtons from './AccountAdminToolsButtons.vue'
import AccountAutoRefreshMenu from './AccountAutoRefreshMenu.vue'
import AccountColumnSettingsMenu from './AccountColumnSettingsMenu.vue'

type AccountAutoRefreshInterval = (typeof ACCOUNT_AUTO_REFRESH_INTERVALS)[number]

defineProps<{
  autoRefreshEnabled: boolean
  autoRefreshCountdown: number
  autoRefreshIntervals: readonly AccountAutoRefreshInterval[]
  autoRefreshIntervalSeconds: AccountAutoRefreshInterval
  autoRefreshIntervalLabel: (seconds: AccountAutoRefreshInterval) => string
  toggleableColumns: Column[]
  isColumnVisible: (key: string) => boolean
}>()

const emit = defineEmits<{
  'toggle-auto-refresh-enabled': []
  'set-auto-refresh-interval': [seconds: AccountAutoRefreshInterval]
  'toggle-column': [key: string]
  'error-passthrough': []
  'tls-profiles': []
}>()

const { t } = useI18n()

const showColumnDropdown = ref(false)
const showAutoRefreshDropdown = ref(false)
const columnDropdownRef = ref<HTMLElement | null>(null)
const autoRefreshDropdownRef = ref<HTMLElement | null>(null)

const toggleAutoRefreshDropdown = () => {
  showAutoRefreshDropdown.value = !showAutoRefreshDropdown.value
  showColumnDropdown.value = false
}

const toggleColumnDropdown = () => {
  showColumnDropdown.value = !showColumnDropdown.value
  showAutoRefreshDropdown.value = false
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement | null
  if (!target) {
    return
  }

  if (columnDropdownRef.value && !columnDropdownRef.value.contains(target)) {
    showColumnDropdown.value = false
  }
  if (autoRefreshDropdownRef.value && !autoRefreshDropdownRef.value.contains(target)) {
    showAutoRefreshDropdown.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>
