<template>
  <div class="flex items-center gap-3">
    <LocaleSwitcher />
    <a
      v-if="docUrl"
      :href="docUrl"
      target="_blank"
      rel="noopener noreferrer"
      class="rounded-lg p-2 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
      :title="docsTitle"
    >
      <Icon name="book" size="md" />
    </a>
    <button
      type="button"
      class="rounded-lg p-2 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
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
