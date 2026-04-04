<template>
  <div class="flex justify-end gap-3">
    <button type="button" class="btn btn-secondary" @click="emit('close')">
      {{ t('common.cancel') }}
    </button>
    <button
      v-if="showSubmit"
      type="submit"
      form="edit-proxy-form"
      :disabled="submitting"
      class="btn btn-primary"
    >
      <ProxyLoadingSpinnerIcon v-if="submitting" />
      {{ submitting ? t('admin.proxies.updating') : t('common.update') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import ProxyLoadingSpinnerIcon from './ProxyLoadingSpinnerIcon.vue'

defineProps<{
  showSubmit: boolean
  submitting: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()
</script>
