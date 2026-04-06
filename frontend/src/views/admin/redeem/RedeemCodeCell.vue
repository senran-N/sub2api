<template>
  <div class="flex items-center space-x-2">
    <code class="code text-sm">{{ code }}</code>
    <button
      :class="[
        'redeem-code-cell__copy-button theme-icon-button flex items-center',
        copied
          ? 'theme-icon-button--success'
          : 'theme-icon-button--neutral'
      ]"
      :title="copied ? t('admin.redeem.copied') : t('keys.copyToClipboard')"
      @click="emit('copy')"
    >
      <Icon v-if="!copied" name="copy" size="sm" :stroke-width="2" />
      <svg v-else class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M5 13l4 4L19 7"
        />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  code: string
  copied: boolean
}>()

const emit = defineEmits<{
  copy: []
}>()

const { t } = useI18n()
</script>

<style scoped>
.redeem-code-cell__copy-button {
  border-radius: var(--theme-button-radius);
  padding: calc(var(--theme-button-padding-y) / 2);
}
</style>
