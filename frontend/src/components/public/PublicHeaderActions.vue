<template>
  <div class="flex items-center gap-3">
    <LocaleSwitcher />
    <a
      v-if="docUrl"
      :href="docUrl"
      target="_blank"
      rel="noopener noreferrer"
      class="public-header-actions__icon theme-icon-button theme-icon-button--neutral"
      :title="docsTitle"
    >
      <Icon name="book" size="md" />
    </a>
    <button
      type="button"
      class="public-header-actions__icon theme-icon-button theme-icon-button--neutral"
      :title="themeTitle"
      @click="$emit('toggle-theme')"
    >
      <Icon v-if="isDark" name="sun" size="md" />
      <Icon v-else name="moon" size="md" />
    </button>
    <slot />
  </div>
</template>

<script setup lang="ts">
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  docUrl?: string
  isDark: boolean
  docsTitle: string
  themeTitle: string
}>()

defineEmits<{
  'toggle-theme': []
}>()
</script>

<style scoped>
.public-header-actions__icon {
  border-radius: var(--theme-public-action-radius);
  padding: calc(var(--theme-button-padding-y) * 0.7);
}
</style>
