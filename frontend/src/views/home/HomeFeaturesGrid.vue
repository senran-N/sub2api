<template>
  <div class="home-features mb-12 grid gap-6 md:grid-cols-3">
    <article
      v-for="(feature, index) in features"
      :key="feature.key"
      class="home-features__card group"
    >
      <!-- Index marker: factory shows `// 01`, claude shows `Chapter I`. -->
      <div class="home-features__index" aria-hidden="true">
        {{ formatIndex(index) }}
      </div>

      <!-- Corner rivets (factory only) - purely decorative. -->
      <span class="home-features__rivet home-features__rivet--tl" aria-hidden="true"></span>
      <span class="home-features__rivet home-features__rivet--tr" aria-hidden="true"></span>
      <span class="home-features__rivet home-features__rivet--bl" aria-hidden="true"></span>
      <span class="home-features__rivet home-features__rivet--br" aria-hidden="true"></span>

      <div
        class="home-features__icon-shell"
        :class="`home-features__icon-shell--${feature.accentTone}`"
      >
        <Icon :name="feature.icon" size="lg" class="home-features__icon" />
      </div>
      <h3 class="home-features__title">
        {{ feature.title }}
      </h3>
      <div class="home-features__underline" aria-hidden="true"></div>
      <p class="home-features__description">
        {{ feature.description }}
      </p>
    </article>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'
import type { HomeFeatureCard } from './homeView'

defineProps<{
  features: HomeFeatureCard[]
}>()

// Both themes surface an index — the visual treatment diverges in CSS. We
// store both formats once and let CSS hide whichever isn't in play.
function formatIndex(index: number): string {
  const zeroPadded = String(index + 1).padStart(2, '0')
  return zeroPadded
}
</script>

<style scoped>
.home-features__card {
  position: relative;
  padding: var(--theme-settings-card-body-padding);
  border-radius: var(--theme-home-features-card-radius);
  border: var(--theme-home-features-card-border);
  background: var(--theme-home-features-card-bg);
  box-shadow: var(--theme-home-features-card-shadow);
  backdrop-filter: var(--theme-home-features-card-backdrop-filter);
  transition:
    transform 0.25s ease,
    box-shadow 0.25s ease,
    border-color 0.25s ease;
}

.home-features__card:hover {
  transform: var(--theme-home-features-card-hover-transform);
  box-shadow: var(--theme-home-features-card-hover-shadow);
  border-color: var(--theme-home-features-card-hover-border);
}

.home-features__icon-shell {
  margin-bottom: 1rem;
  display: flex;
  height: 48px;
  width: 48px;
  align-items: center;
  justify-content: center;
  border-radius: var(--theme-home-feature-icon-radius);
  box-shadow: var(--theme-home-features-icon-shadow);
  transition: transform 0.25s ease;
}

.home-features__card:hover .home-features__icon-shell {
  transform: scale(1.08);
}

.home-features__icon-shell--info {
  background: linear-gradient(
    135deg,
    rgb(var(--theme-info-rgb)),
    color-mix(in srgb, rgb(var(--theme-info-rgb)) 72%, var(--theme-accent-strong))
  );
}

.home-features__icon-shell--accent {
  background: linear-gradient(
    135deg,
    var(--theme-accent),
    color-mix(in srgb, var(--theme-accent) 72%, var(--theme-accent-strong))
  );
}

.home-features__icon-shell--success {
  background: linear-gradient(
    135deg,
    rgb(var(--theme-success-rgb)),
    color-mix(in srgb, rgb(var(--theme-success-rgb)) 72%, rgb(var(--theme-info-rgb)))
  );
}

.home-features__icon {
  color: var(--theme-filled-text);
}

.home-features__title {
  margin-bottom: 0.5rem;
  font-family: var(--theme-home-features-title-font);
  font-size: var(--theme-home-features-title-size);
  font-weight: var(--theme-home-features-title-weight);
  letter-spacing: var(--theme-home-features-title-letter-spacing);
  text-transform: var(--theme-home-features-title-transform);
  color: var(--theme-page-text);
}

.home-features__description {
  font-size: 0.875rem;
  line-height: 1.6;
  color: var(--theme-page-muted);
}

.home-features__index {
  display: var(--theme-home-features-index-display);
  font-family: var(--theme-home-features-index-font);
  font-style: var(--theme-home-features-index-font-style);
  font-size: var(--theme-home-features-index-font-size);
  font-weight: var(--theme-home-features-index-font-weight);
  letter-spacing: var(--theme-home-features-index-letter-spacing);
  text-transform: var(--theme-home-features-index-text-transform);
  color: var(--theme-home-features-index-color);
  margin-bottom: var(--theme-home-features-index-margin-bottom);
}

.home-features__index::before {
  content: var(--theme-home-features-index-prefix);
  color: var(--theme-home-features-index-prefix-color);
}

.home-features__underline {
  display: var(--theme-home-features-underline-display);
  width: var(--theme-home-features-underline-width);
  height: var(--theme-home-features-underline-height);
  margin-bottom: 0.75rem;
  background: var(--theme-home-features-underline-bg);
  transition: width 0.3s ease;
}

.home-features__card:hover .home-features__underline {
  width: var(--theme-home-features-underline-hover-width);
}

.home-features__rivet {
  display: var(--theme-home-features-rivet-display);
  position: absolute;
  width: var(--theme-home-features-rivet-size);
  height: var(--theme-home-features-rivet-size);
  background: var(--theme-home-features-rivet-color);
  border-radius: var(--theme-home-features-rivet-radius);
}

.home-features__rivet--tl { top: 8px; left: 8px; }
.home-features__rivet--tr { top: 8px; right: 8px; }
.home-features__rivet--bl { bottom: 8px; left: 8px; }
.home-features__rivet--br { bottom: 8px; right: 8px; }
</style>
