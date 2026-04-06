<template>
  <div class="flex items-center gap-1">
    <button
      type="button"
      class="theme-action-button user-actions-cell__button user-actions-cell__button--edit"
      @click="emit('edit', user)"
    >
      <Icon name="edit" size="sm" />
      <span class="text-xs">{{ t('common.edit') }}</span>
    </button>

    <button
      v-if="user.role !== 'admin'"
      type="button"
      :class="[
        'theme-action-button user-actions-cell__button',
        user.status === 'active'
          ? 'user-actions-cell__button--warning'
          : 'user-actions-cell__button--success'
      ]"
      @click="emit('toggle-status', user)"
    >
      <Icon v-if="user.status === 'active'" name="ban" size="sm" />
      <Icon v-else name="checkCircle" size="sm" />
      <span class="text-xs">
        {{ user.status === 'active' ? t('admin.users.disable') : t('admin.users.enable') }}
      </span>
    </button>

    <button
      type="button"
      class="action-menu-trigger theme-action-button user-actions-cell__button user-actions-cell__button--menu"
      :class="{ 'user-actions-cell__button--active': menuOpen }"
      @click="emit('open-menu', user, $event)"
    >
      <Icon name="more" size="sm" />
      <span class="text-xs">{{ t('common.more') }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { AdminUser } from '@/types'

defineProps<{
  user: AdminUser
  menuOpen: boolean
}>()

const emit = defineEmits<{
  edit: [user: AdminUser]
  'toggle-status': [user: AdminUser]
  'open-menu': [user: AdminUser, event: MouseEvent]
}>()

const { t } = useI18n()
</script>

<style scoped>
.user-actions-cell__button {
  color: var(--theme-page-muted);
  transition: color 0.2s ease, background-color 0.2s ease;
}

.user-actions-cell__button--edit:hover {
  background: var(--theme-button-ghost-hover-bg);
  color: var(--theme-accent);
}

.user-actions-cell__button--warning:hover {
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.user-actions-cell__button--success:hover {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.user-actions-cell__button--menu:hover,
.user-actions-cell__button--active {
  background: var(--theme-button-ghost-hover-bg);
  color: var(--theme-page-text);
}
</style>
