<template>
  <div class="proxy-create-mode-tabs mb-6 flex border-b">
    <button
      type="button"
      :class="tabClass('standard')"
      @click="emit('update:modelValue', 'standard')"
    >
      <Icon name="plus" size="sm" class="mr-1.5 inline" />
      {{ t('admin.proxies.standardAdd') }}
    </button>
    <button
      type="button"
      :class="tabClass('batch')"
      @click="emit('update:modelValue', 'batch')"
    >
      <Icon name="menu" size="sm" class="mr-1.5 inline" />
      {{ t('admin.proxies.batchAdd') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  modelValue: 'standard' | 'batch'
}>()

const emit = defineEmits<{
  'update:modelValue': [value: 'standard' | 'batch']
}>()

const { t } = useI18n()

const tabClass = (mode: 'standard' | 'batch') => [
  'proxy-create-mode-tabs__tab',
  props.modelValue === mode
    ? 'proxy-create-mode-tabs__tab--active'
    : 'proxy-create-mode-tabs__tab--inactive'
]
</script>

<style scoped>
.proxy-create-mode-tabs {
  border-color: color-mix(in srgb, var(--theme-card-border) 82%, transparent);
}

.proxy-create-mode-tabs__tab {
  transition: color 0.2s ease, border-color 0.2s ease;
  padding: var(--theme-proxy-selector-trigger-padding-y)
    var(--theme-proxy-selector-trigger-padding-x);
  font-size: 0.875rem;
  font-weight: 500;
  border-bottom-width: 2px;
  border-bottom-style: solid;
  border-bottom-color: transparent;
  margin-bottom: -1px;
}

.proxy-create-mode-tabs__tab--active {
  border-color: var(--theme-accent);
  color: var(--theme-accent);
}

.proxy-create-mode-tabs__tab--inactive {
  border-color: transparent;
  color: var(--theme-page-muted);
}

.proxy-create-mode-tabs__tab--inactive:hover {
  color: var(--theme-page-text);
}
</style>
