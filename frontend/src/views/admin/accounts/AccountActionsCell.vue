<template>
  <div class="flex items-center gap-1">
    <button
      class="theme-action-button theme-action-button--accent"
      @click="emit('edit', account)"
    >
      <Icon name="edit" size="sm" />
      <span class="text-xs">{{ t('common.edit') }}</span>
    </button>
    <button
      class="theme-action-button theme-action-button--danger"
      @click="emit('delete', account)"
    >
      <Icon name="trash" size="sm" />
      <span class="text-xs">{{ t('common.delete') }}</span>
    </button>
    <button
      class="theme-action-button theme-action-button--neutral"
      @click="emit('open-menu', account, $event)"
    >
      <Icon name="more" size="sm" />
      <span class="text-xs">{{ t('common.more') }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { Account } from '@/types'

defineProps<{
  account: Account
}>()

const emit = defineEmits<{
  edit: [account: Account]
  delete: [account: Account]
  'open-menu': [account: Account, event: MouseEvent]
}>()

const { t } = useI18n()
</script>
