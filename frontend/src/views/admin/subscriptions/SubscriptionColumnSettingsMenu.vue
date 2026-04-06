<template>
  <div
    class="subscription-column-settings-menu absolute right-0 z-50 origin-top-right"
  >
    <div class="subscription-column-settings-menu__body">
      <div class="subscription-column-settings-menu__section border-b">
        <div class="subscription-column-settings-menu__label text-xs font-medium">
          {{ t('admin.subscriptions.columns.user') }}
        </div>
        <button
          class="subscription-column-settings-menu__button flex w-full items-center justify-between text-sm"
          @click="emit('set-user-mode', 'email')"
        >
          <span>{{ t('admin.users.columns.email') }}</span>
          <Icon v-if="userColumnMode === 'email'" name="check" size="sm" class="subscription-column-settings-menu__check" />
        </button>
        <button
          class="subscription-column-settings-menu__button flex w-full items-center justify-between text-sm"
          @click="emit('set-user-mode', 'username')"
        >
          <span>{{ t('admin.users.columns.username') }}</span>
          <Icon v-if="userColumnMode === 'username'" name="check" size="sm" class="subscription-column-settings-menu__check" />
        </button>
      </div>
      <button
        v-for="column in toggleableColumns"
        :key="column.key"
        class="subscription-column-settings-menu__button flex w-full items-center justify-between text-sm"
        @click="emit('toggle-column', column.key)"
      >
        <span>{{ column.label }}</span>
        <Icon v-if="isColumnVisible(column.key)" name="check" size="sm" class="subscription-column-settings-menu__check" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { Column } from '@/components/common/types'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  userColumnMode: 'email' | 'username'
  toggleableColumns: Column[]
  isColumnVisible: (key: string) => boolean
}>()

const emit = defineEmits<{
  'set-user-mode': [mode: 'email' | 'username']
  'toggle-column': [key: string]
}>()

const { t } = useI18n()
</script>

<style scoped>
.subscription-column-settings-menu {
  width: var(--theme-settings-menu-width-sm);
  max-height: var(--theme-settings-menu-max-height);
  overflow-y: auto;
  margin-top: calc(var(--theme-floating-panel-gap) * 0.5 + 0.375rem);
  border: 1px solid color-mix(in srgb, var(--theme-dropdown-border) 88%, transparent);
  border-radius: calc(var(--theme-surface-radius) + 2px);
  background: var(--theme-dropdown-bg);
  box-shadow: var(--theme-dropdown-shadow);
}

.subscription-column-settings-menu__body {
  padding: calc(var(--theme-floating-panel-gap) * 0.5 + 0.25rem);
}

.subscription-column-settings-menu__section {
  margin-bottom: calc(var(--theme-floating-panel-gap) * 0.5 + 0.375rem);
  padding-bottom: calc(var(--theme-floating-panel-gap) * 0.5 + 0.25rem);
  border-color: color-mix(in srgb, var(--theme-card-border) 76%, transparent);
}

.subscription-column-settings-menu__label {
  padding: calc(var(--theme-floating-panel-gap) * 0.5 + 0.125rem) calc(var(--theme-button-padding-x) * 0.65);
  color: var(--theme-page-muted);
}

.subscription-column-settings-menu__button {
  padding: calc(var(--theme-button-padding-y) * 0.8) calc(var(--theme-button-padding-x) * 0.65);
  border-radius: calc(var(--theme-button-radius) + 2px);
  color: var(--theme-page-text);
  transition: background-color 0.2s ease, color 0.2s ease;
}

.subscription-column-settings-menu__button:hover {
  background: var(--theme-dropdown-item-hover-bg);
}

.subscription-column-settings-menu__check {
  color: var(--theme-accent);
}
</style>
