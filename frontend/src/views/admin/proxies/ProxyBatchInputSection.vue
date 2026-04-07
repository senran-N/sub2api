<template>
  <div class="space-y-5">
    <div>
      <label class="input-label">{{ t('admin.proxies.batchInput') }}</label>
      <textarea
        :value="modelValue"
        rows="10"
        class="input font-mono text-sm"
        :placeholder="t('admin.proxies.batchInputPlaceholder')"
        @input="handleInput"
      ></textarea>
      <p class="input-hint mt-2">
        {{ t('admin.proxies.batchInputHint') }}
      </p>
    </div>

    <ProxyBatchParseSummary :summary="summary" />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { ProxyBatchParseState } from './proxyForm'
import ProxyBatchParseSummary from './ProxyBatchParseSummary.vue'

defineProps<{
  modelValue: string
  summary: ProxyBatchParseState
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  input: [event: Event]
}>()

const { t } = useI18n()

const handleInput = (event: Event) => {
  const target = event.target as HTMLTextAreaElement
  emit('update:modelValue', target.value)
  emit('input', event)
}
</script>
