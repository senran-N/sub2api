<template>
  <div class="flex justify-end gap-3">
    <button
      type="button"
      :disabled="loading"
      class="btn btn-secondary"
      :title="t('common.refresh')"
      @click="emit('refresh')"
    >
      <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
    </button>
    <button type="button" class="btn btn-primary" data-tour="keys-create-btn" @click="emit('create')">
      <Icon name="plus" size="md" class="mr-2" />
      {{ t('keys.createKey') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  loading: boolean
}>()

const emit = defineEmits<{
  refresh: []
  create: []
}>()

const { t } = useI18n()
</script>
