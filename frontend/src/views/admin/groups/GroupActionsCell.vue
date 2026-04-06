<template>
  <div class="flex items-center gap-1">
    <button
      class="theme-action-button theme-action-button--accent"
      @click="$emit('edit', group)"
    >
      <Icon name="edit" size="sm" />
      <span class="text-xs">{{ t('common.edit') }}</span>
    </button>
    <button
      class="theme-action-button theme-action-button--brand-purple"
      @click="$emit('rate-multipliers', group)"
    >
      <Icon name="dollar" size="sm" />
      <span class="text-xs">{{ t('admin.groups.rateMultipliers') }}</span>
    </button>
    <button
      class="theme-action-button theme-action-button--danger"
      @click="$emit('delete', group)"
    >
      <Icon name="trash" size="sm" />
      <span class="text-xs">{{ t('common.delete') }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { AdminGroup } from '@/types'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  group: AdminGroup
}>()

defineEmits<{
  edit: [group: AdminGroup]
  'rate-multipliers': [group: AdminGroup]
  delete: [group: AdminGroup]
}>()

const { t } = useI18n()
</script>
