<template>
  <button class="btn btn-secondary" @click="emit('import')">
    {{ t('admin.accounts.dataImport') }}
  </button>
  <button class="btn btn-secondary" @click="emit('export')">
    {{ selectedCount ? t('admin.accounts.dataExportSelected') : t('admin.accounts.dataExport') }}
  </button>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineProps<{
  selectedCount: number
}>()

const emit = defineEmits<{
  import: []
  export: []
}>()

const { t } = useI18n()
</script>
