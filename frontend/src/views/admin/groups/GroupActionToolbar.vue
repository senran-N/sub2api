<template>
  <div class="flex w-full flex-shrink-0 flex-wrap items-center justify-end gap-3 lg:w-auto">
    <button
      class="btn btn-secondary"
      :disabled="loading"
      :title="t('common.refresh')"
      @click="emit('refresh')"
    >
      <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
    </button>
    <button
      class="btn btn-secondary"
      :title="t('admin.groups.sortOrder')"
      @click="emit('sort-order')"
    >
      <Icon name="arrowsUpDown" size="md" class="mr-2" />
      {{ t('admin.groups.sortOrder') }}
    </button>
    <button
      class="btn btn-primary"
      data-tour="groups-create-btn"
      @click="emit('create')"
    >
      <Icon name="plus" size="md" class="mr-2" />
      {{ t('admin.groups.createGroup') }}
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
  'sort-order': []
  create: []
}>()

const { t } = useI18n()
</script>
