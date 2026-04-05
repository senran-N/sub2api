<template>
  <section>
    <div class="mb-8 text-center">
      <h2 class="mb-3 text-2xl font-bold text-gray-900 dark:text-white">
        {{ title }}
      </h2>
      <p class="text-sm text-gray-600 dark:text-dark-400">
        {{ description }}
      </p>
    </div>

    <div class="mb-16 flex flex-wrap items-center justify-center gap-4">
      <div
        v-for="provider in providers"
        :key="provider.key"
        class="flex items-center gap-2 rounded-xl px-5 py-3 backdrop-blur-sm"
        :class="
          provider.supported
            ? 'border border-primary-200 bg-white/60 ring-1 ring-primary-500/20 dark:border-primary-800 dark:bg-dark-800/60'
            : 'border border-gray-200/50 bg-white/40 opacity-60 dark:border-dark-700/50 dark:bg-dark-800/40'
        "
      >
        <div
          class="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br"
          :class="provider.accentClass"
        >
          <span class="text-xs font-bold text-white">{{ provider.initial }}</span>
        </div>
        <span class="text-sm font-medium text-gray-700 dark:text-dark-200">{{ provider.label }}</span>
        <span
          class="rounded px-1.5 py-0.5 text-[10px] font-medium"
          :class="
            provider.supported
              ? 'bg-primary-100 text-primary-600 dark:bg-primary-900/30 dark:text-primary-400'
              : 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-dark-400'
          "
        >
          {{ provider.statusLabel }}
        </span>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { HomeProviderBadge } from './homeView'

defineProps<{
  description: string
  providers: HomeProviderBadge[]
  title: string
}>()
</script>
