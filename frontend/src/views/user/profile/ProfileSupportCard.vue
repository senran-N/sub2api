<template>
  <div
    class="card border-primary-200 bg-primary-50 p-6 dark:bg-primary-900/20"
  >
    <div class="flex items-center gap-4">
      <div class="rounded-xl bg-primary-100 p-3 text-primary-600">
        <Icon name="chat" size="lg" />
      </div>
      <div>
        <h3 class="font-semibold text-primary-800 dark:text-primary-200">
          {{ title }}
        </h3>
        <p class="text-sm font-medium">{{ contactInfo }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  title: string
  contactInfo: string
}>()
</script>
