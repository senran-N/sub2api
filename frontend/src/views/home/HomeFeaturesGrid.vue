<template>
  <div class="mb-12 grid gap-6 md:grid-cols-3">
    <div
      v-for="feature in features"
      :key="feature.key"
      class="home-features-grid__card group backdrop-blur-sm transition-all duration-300"
    >
      <div
        class="home-features-grid__icon-shell mb-4 flex h-12 w-12 items-center justify-center transition-transform group-hover:scale-110"
        :class="`home-features-grid__icon-shell--${feature.accentTone}`"
      >
        <Icon :name="feature.icon" size="lg" class="home-features-grid__icon" />
      </div>
      <h3 class="home-features-grid__title mb-2 text-lg font-semibold">
        {{ feature.title }}
      </h3>
      <p class="home-features-grid__description text-sm leading-relaxed">
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

<style scoped>
.home-features-grid__card {
  padding: var(--theme-settings-card-body-padding);
  border-radius: var(--theme-home-feature-card-radius);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 76%, transparent);
  background: color-mix(in srgb, var(--theme-surface) 64%, transparent);
}

.home-features-grid__card:hover {
  box-shadow: var(--theme-card-shadow-hover);
}

.home-features-grid__icon-shell {
  border-radius: var(--theme-home-feature-icon-radius);
  box-shadow: 0 18px 34px color-mix(in srgb, var(--theme-accent) 16%, transparent);
}

.home-features-grid__icon-shell--info {
  background: linear-gradient(
    135deg,
    rgb(var(--theme-info-rgb)),
    color-mix(in srgb, rgb(var(--theme-info-rgb)) 72%, var(--theme-accent-strong))
  );
}

.home-features-grid__icon-shell--accent {
  background: linear-gradient(
    135deg,
    var(--theme-accent),
    color-mix(in srgb, var(--theme-accent) 72%, var(--theme-accent-strong))
  );
}

.home-features-grid__icon-shell--success {
  background: linear-gradient(
    135deg,
    rgb(var(--theme-success-rgb)),
    color-mix(in srgb, rgb(var(--theme-success-rgb)) 72%, rgb(var(--theme-info-rgb)))
  );
}

.home-features-grid__icon {
  color: var(--theme-filled-text);
}

.home-features-grid__title {
  font-family: var(--theme-home-section-title-font);
  color: var(--theme-page-text);
}

.home-features-grid__description {
  color: var(--theme-page-muted);
}
</style>
