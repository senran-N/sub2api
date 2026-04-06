<template>
  <div class="mb-12 flex flex-col items-center justify-between gap-12 lg:flex-row lg:gap-16">
    <div class="flex-1 text-center lg:text-left">
      <h1 class="home-hero-section__title theme-text-strong mb-4 text-4xl md:text-5xl lg:text-6xl">
        {{ siteName }}
      </h1>
      <p class="home-hero-section__subtitle theme-text-muted mb-8 text-lg md:text-xl">
        {{ siteSubtitle }}
      </p>

      <div>
        <router-link
          :to="ctaPath"
          class="home-hero-section__cta btn btn-primary text-base shadow-lg"
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

<style scoped>
.home-hero-section__title {
  font-family: var(--theme-home-title-font);
  font-weight: var(--theme-home-title-weight);
  letter-spacing: var(--theme-home-title-letter-spacing);
}

.home-hero-section__subtitle {
  font-family: var(--theme-home-subtitle-font);
}

.home-hero-section__cta {
  padding:
    calc(var(--theme-markdown-block-padding) - 0.25rem)
    calc(var(--theme-auth-card-padding) - 0.5rem);
  box-shadow: 0 18px 34px color-mix(in srgb, var(--theme-accent) 28%, transparent);
}
</style>
