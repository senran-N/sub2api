<template>
  <div class="card overflow-hidden">
    <div class="bg-gradient-to-br from-primary-500 to-primary-600 px-6 py-8 text-center">
      <div
        class="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-white/20 backdrop-blur-sm"
      >
        <Icon name="creditCard" size="xl" class="text-white" />
      </div>
      <p class="text-sm font-medium text-primary-100">{{ currentBalanceLabel }}</p>
      <p class="mt-2 text-4xl font-bold text-white">{{ balanceText }}</p>
      <p class="mt-2 text-sm text-primary-100">
        {{ concurrencyLabel }}: {{ concurrency }} {{ requestsLabel }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  balanceText: string
  concurrency: number
  concurrencyLabel: string
  currentBalanceLabel: string
  requestsLabel: string
}>()
</script>
