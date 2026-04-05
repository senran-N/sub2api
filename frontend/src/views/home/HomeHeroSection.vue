<template>
  <div class="mb-12 flex flex-col items-center justify-between gap-12 lg:flex-row lg:gap-16">
    <div class="flex-1 text-center lg:text-left">
      <h1 class="mb-4 text-4xl font-bold text-gray-900 dark:text-white md:text-5xl lg:text-6xl">
        {{ siteName }}
      </h1>
      <p class="mb-8 text-lg text-gray-600 dark:text-dark-300 md:text-xl">
        {{ siteSubtitle }}
      </p>

      <div>
        <router-link
          :to="ctaPath"
          class="btn btn-primary px-8 py-3 text-base shadow-lg shadow-primary-500/30"
        >
          {{ ctaLabel }}
          <Icon name="arrowRight" size="md" class="ml-2" :stroke-width="2" />
        </router-link>
      </div>
    </div>

    <div class="flex flex-1 justify-center lg:justify-end">
      <HomeTerminalDemo />
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'
import HomeTerminalDemo from './HomeTerminalDemo.vue'

defineProps<{
  ctaLabel: string
  ctaPath: string
  siteName: string
  siteSubtitle: string
}>()
</script>
