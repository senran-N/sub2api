<template>
  <button
    class="btn btn-secondary"
    :title="t('admin.errorPassthrough.title')"
    @click="emit('error-passthrough')"
  >
    <Icon name="shield" size="md" class="mr-1.5" />
    <span class="hidden md:inline">{{ t('admin.errorPassthrough.title') }}</span>
  </button>

  <button
    class="btn btn-secondary"
    :title="t('admin.tlsFingerprintProfiles.title')"
    @click="emit('tls-profiles')"
  >
    <Icon name="lock" size="md" class="mr-1.5" />
    <span class="hidden md:inline">{{ t('admin.tlsFingerprintProfiles.title') }}</span>
  </button>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const emit = defineEmits<{
  'error-passthrough': []
  'tls-profiles': []
}>()

const { t } = useI18n()
</script>
