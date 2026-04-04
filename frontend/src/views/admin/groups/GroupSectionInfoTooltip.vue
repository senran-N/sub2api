<template>
  <div class="group relative inline-flex">
    <Icon
      name="questionCircle"
      size="sm"
      :stroke-width="2"
      class="cursor-help text-gray-400 transition-colors hover:text-primary-500 dark:text-gray-500 dark:hover:text-primary-400"
    />
    <div class="pointer-events-none absolute bottom-full left-0 z-50 mb-2 opacity-0 transition-all duration-200 group-hover:pointer-events-auto group-hover:opacity-100">
      <div :class="['rounded-lg bg-gray-900 p-3 text-white shadow-lg dark:bg-gray-800', widthClass]">
        <p class="text-xs leading-relaxed text-gray-300">
          {{ text }}
        </p>
        <div class="absolute -bottom-1.5 left-3 h-3 w-3 rotate-45 bg-gray-900 dark:bg-gray-800"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

withDefaults(
  defineProps<{
    text: string
    widthClass?: string
  }>(),
  {
    widthClass: 'w-72'
  }
)
</script>
