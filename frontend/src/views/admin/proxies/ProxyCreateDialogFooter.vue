<template>
  <div class="flex justify-end gap-3">
    <button type="button" class="btn btn-secondary" @click="emit('close')">
      {{ t('common.cancel') }}
    </button>
    <button
      v-if="mode === 'standard'"
      type="submit"
      form="create-proxy-form"
      :disabled="submitting"
      class="btn btn-primary"
    >
      <ProxyLoadingSpinnerIcon v-if="submitting" />
      {{ submitting ? t('admin.proxies.creating') : t('common.create') }}
    </button>
    <button
      v-else
      type="button"
      :disabled="submitting || validCount === 0"
      class="btn btn-primary"
      @click="emit('batch-create')"
    >
      <ProxyLoadingSpinnerIcon v-if="submitting" />
      {{
        submitting
          ? t('admin.proxies.importing')
          : t('admin.proxies.importProxies', { count: validCount })
      }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import ProxyLoadingSpinnerIcon from './ProxyLoadingSpinnerIcon.vue'

defineProps<{
  mode: 'standard' | 'batch'
  submitting: boolean
  validCount: number
}>()

const emit = defineEmits<{
  close: []
  'batch-create': []
}>()

const { t } = useI18n()
</script>
