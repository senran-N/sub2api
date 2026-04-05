<template>
  <div class="mb-12 grid gap-6 md:grid-cols-3">
    <div
      v-for="feature in features"
      :key="feature.key"
      class="group rounded-2xl border border-gray-200/50 bg-white/60 p-6 backdrop-blur-sm transition-all duration-300 hover:shadow-xl hover:shadow-primary-500/10 dark:border-dark-700/50 dark:bg-dark-800/60"
    >
      <div
        class="mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br shadow-lg transition-transform group-hover:scale-110"
        :class="feature.accentClass"
      >
        <Icon :name="feature.icon" size="lg" class="text-white" />
      </div>
      <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
        {{ feature.title }}
      </h3>
      <p class="text-sm leading-relaxed text-gray-600 dark:text-dark-400">
        {{ feature.description }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'
import type { HomeFeatureCard } from './homeView'

defineProps<{
  features: HomeFeatureCard[]
}>()
</script>
