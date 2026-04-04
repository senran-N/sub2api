<template>
  <div class="flex flex-1 flex-wrap items-center justify-end gap-2">
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
      :disabled="batchTesting || loading"
      :title="t('admin.proxies.testConnection')"
      @click="emit('batch-test')"
    >
      <Icon name="play" size="md" class="mr-2" />
      {{ t('admin.proxies.testConnection') }}
    </button>
    <button
      class="btn btn-secondary"
      :disabled="batchQualityChecking || loading"
      :title="t('admin.proxies.batchQualityCheck')"
      @click="emit('batch-quality-check')"
    >
      <Icon
        name="shield"
        size="md"
        class="mr-2"
        :class="batchQualityChecking ? 'animate-pulse' : ''"
      />
      {{ t('admin.proxies.batchQualityCheck') }}
    </button>
    <button
      class="btn btn-danger"
      :disabled="selectedCount === 0"
      :title="t('admin.proxies.batchDeleteAction')"
      @click="emit('batch-delete')"
    >
      <Icon name="trash" size="md" class="mr-2" />
      {{ t('admin.proxies.batchDeleteAction') }}
    </button>
    <button class="btn btn-secondary" @click="emit('import')">
      {{ t('admin.proxies.dataImport') }}
    </button>
    <button class="btn btn-secondary" @click="emit('export')">
      {{ selectedCount > 0 ? t('admin.proxies.dataExportSelected') : t('admin.proxies.dataExport') }}
    </button>
    <button class="btn btn-primary" @click="emit('create')">
      <Icon name="plus" size="md" class="mr-2" />
      {{ t('admin.proxies.createProxy') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  loading: boolean
  batchTesting: boolean
  batchQualityChecking: boolean
  selectedCount: number
}>()

const emit = defineEmits<{
  refresh: []
  'batch-test': []
  'batch-quality-check': []
  'batch-delete': []
  import: []
  export: []
  create: []
}>()

const { t } = useI18n()
</script>
