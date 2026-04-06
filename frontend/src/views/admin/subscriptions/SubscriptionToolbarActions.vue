<template>
  <div class="ml-auto flex flex-wrap items-center justify-end gap-3">
    <button
      class="btn btn-secondary"
      :disabled="loading"
      :title="t('common.refresh')"
      @click="emit('refresh')"
    >
      <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
    </button>

    <div class="relative" ref="columnDropdownRef">
      <button
        class="subscription-toolbar-actions__trigger btn btn-secondary"
        :title="t('admin.users.columnSettings')"
        @click="toggleColumnDropdown"
      >
        <Icon name="grid" size="sm" class="md:mr-1.5" />
        <span class="hidden md:inline">{{ t('admin.users.columnSettings') }}</span>
      </button>
      <SubscriptionColumnSettingsMenu
        v-if="showColumnDropdown"
        :user-column-mode="userColumnMode"
        :toggleable-columns="toggleableColumns"
        :is-column-visible="isColumnVisible"
        @set-user-mode="emit('set-user-mode', $event)"
        @toggle-column="emit('toggle-column', $event)"
      />
    </div>

    <button
      class="btn btn-secondary"
      :title="t('admin.subscriptions.guide.showGuide')"
      @click="emit('guide')"
    >
      <Icon name="questionCircle" size="md" />
    </button>

    <button class="btn btn-primary" @click="emit('assign')">
      <Icon name="plus" size="md" class="mr-2" />
      {{ t('admin.subscriptions.assignSubscription') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Column } from '@/components/common/types'
import Icon from '@/components/icons/Icon.vue'
import SubscriptionColumnSettingsMenu from './SubscriptionColumnSettingsMenu.vue'

defineProps<{
  loading: boolean
  userColumnMode: 'email' | 'username'
  toggleableColumns: Column[]
  isColumnVisible: (key: string) => boolean
}>()

const emit = defineEmits<{
  refresh: []
  'set-user-mode': [mode: 'email' | 'username']
  'toggle-column': [key: string]
  guide: []
  assign: []
}>()

const { t } = useI18n()

const showColumnDropdown = ref(false)
const columnDropdownRef = ref<HTMLElement | null>(null)

const toggleColumnDropdown = () => {
  showColumnDropdown.value = !showColumnDropdown.value
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement | null
  if (!target) {
    return
  }

  if (columnDropdownRef.value && !columnDropdownRef.value.contains(target)) {
    showColumnDropdown.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.subscription-toolbar-actions__trigger {
  padding-inline: var(--theme-settings-code-padding-x);
}

@media (min-width: 768px) {
  .subscription-toolbar-actions__trigger {
    padding-inline: var(--theme-settings-action-padding-x);
  }
}
</style>
