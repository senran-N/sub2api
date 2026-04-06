<template>
  <div class="flex items-center space-x-1">
    <button
      class="theme-action-button theme-action-button--info announcement-actions-cell__button flex flex-col items-center gap-0.5"
      :title="t('admin.announcements.readStatus')"
      @click="emit('read-status')"
    >
      <Icon name="eye" size="sm" />
    </button>
    <button
      class="theme-action-button theme-action-button--accent announcement-actions-cell__button flex flex-col items-center gap-0.5"
      :title="t('common.edit')"
      @click="emit('edit')"
    >
      <Icon name="edit" size="sm" />
    </button>
    <button
      class="theme-action-button theme-action-button--danger announcement-actions-cell__button flex flex-col items-center gap-0.5"
      :title="t('common.delete')"
      @click="emit('delete')"
    >
      <Icon name="trash" size="sm" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const emit = defineEmits<{
  'read-status': []
  edit: []
  delete: []
}>()

const { t } = useI18n()
</script>

<style scoped>
.announcement-actions-cell__button {
  border-radius: var(--theme-promo-actions-button-radius);
  padding: var(--theme-promo-actions-button-padding);
}
</style>
