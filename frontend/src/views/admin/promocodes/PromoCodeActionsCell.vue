<template>
  <div class="flex items-center space-x-1">
    <button
      class="theme-action-button theme-action-button--success promo-code-actions-cell__button flex flex-col items-center gap-0.5"
      :title="t('admin.promo.copyRegisterLink')"
      @click="emit('copy-link')"
    >
      <Icon name="link" size="sm" />
    </button>
    <button
      class="theme-action-button theme-action-button--info promo-code-actions-cell__button flex flex-col items-center gap-0.5"
      :title="t('admin.promo.viewUsages')"
      @click="emit('view-usages')"
    >
      <Icon name="eye" size="sm" />
    </button>
    <button
      class="theme-action-button theme-action-button--accent promo-code-actions-cell__button flex flex-col items-center gap-0.5"
      :title="t('common.edit')"
      @click="emit('edit')"
    >
      <Icon name="edit" size="sm" />
    </button>
    <button
      class="theme-action-button theme-action-button--danger promo-code-actions-cell__button flex flex-col items-center gap-0.5"
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
  'copy-link': []
  'view-usages': []
  edit: []
  delete: []
}>()

const { t } = useI18n()
</script>

<style scoped>
.promo-code-actions-cell__button {
  border-radius: var(--theme-promo-actions-button-radius);
  padding: var(--theme-promo-actions-button-padding);
}
</style>
