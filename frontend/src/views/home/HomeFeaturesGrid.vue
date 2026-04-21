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
  border-radius: var(--theme-home-feature-card-radius);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 76%, transparent);
  background: color-mix(in srgb, var(--theme-surface) 82%, transparent);
  backdrop-filter: blur(6px);
  transition:
    transform 0.25s ease,
    box-shadow 0.25s ease,
    border-color 0.25s ease;
}

.home-features__icon-shell {
  margin-bottom: 1rem;
  display: flex;
  height: 48px;
  width: 48px;
  align-items: center;
  justify-content: center;
  border-radius: var(--theme-home-feature-icon-radius);
  box-shadow: 0 18px 34px color-mix(in srgb, var(--theme-accent) 16%, transparent);
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
  font-family: var(--theme-home-section-title-font);
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--theme-page-text);
}

.home-features__description {
  font-size: 0.875rem;
  line-height: 1.6;
  color: var(--theme-page-muted);
}

.home-features__index,
.home-features__underline,
.home-features__rivet {
  display: none;
}

/* ============== Factory: blueprint spec card ============== */
:root[data-brand-theme='factory'] .home-features__card {
  border: 2px solid var(--theme-page-text);
  background: var(--theme-surface);
  box-shadow: 4px 4px 0 var(--theme-page-text);
  border-radius: 0;
}

:root[data-brand-theme='factory'] .home-features__card:hover {
  transform: translate(-3px, -3px);
  box-shadow: 7px 7px 0 var(--theme-page-text);
}

:root[data-brand-theme='factory'] .home-features__index {
  display: block;
  font-family: var(--theme-font-mono);
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--theme-accent);
  margin-bottom: 1rem;
}

:root[data-brand-theme='factory'] .home-features__index::before {
  content: '// ';
  color: var(--theme-page-muted);
}

:root[data-brand-theme='factory'] .home-features__icon-shell {
  border-radius: 0;
  box-shadow:
    0 0 0 2px var(--theme-page-text),
    3px 3px 0 var(--theme-page-text);
}

:root[data-brand-theme='factory'] .home-features__title {
  text-transform: uppercase;
  letter-spacing: -0.01em;
}

:root[data-brand-theme='factory'] .home-features__rivet {
  display: block;
  position: absolute;
  width: 6px;
  height: 6px;
  background: var(--theme-page-text);
  border-radius: 0;
}

:root[data-brand-theme='factory'] .home-features__rivet--tl { top: 8px; left: 8px; }
:root[data-brand-theme='factory'] .home-features__rivet--tr { top: 8px; right: 8px; }
:root[data-brand-theme='factory'] .home-features__rivet--bl { bottom: 8px; left: 8px; }
:root[data-brand-theme='factory'] .home-features__rivet--br { bottom: 8px; right: 8px; }

.dark[data-brand-theme='factory'] .home-features__card {
  border-color: rgba(255, 255, 255, 0.3);
  box-shadow: 4px 4px 0 rgba(255, 255, 255, 0.2);
}

.dark[data-brand-theme='factory'] .home-features__card:hover {
  box-shadow: 7px 7px 0 rgba(255, 255, 255, 0.3);
}

.dark[data-brand-theme='factory'] .home-features__icon-shell {
  box-shadow:
    0 0 0 2px rgba(255, 255, 255, 0.3),
    3px 3px 0 rgba(255, 255, 255, 0.2);
}

.dark[data-brand-theme='factory'] .home-features__rivet {
  background: rgba(255, 255, 255, 0.45);
}

/* ============== Claude: magazine article card ============== */
:root[data-brand-theme='claude'] .home-features__card {
  border: 1px solid var(--theme-page-border);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}

:root[data-brand-theme='claude'] .home-features__card:hover {
  transform: translateY(-3px);
  border-color: color-mix(in srgb, var(--theme-accent) 35%, var(--theme-page-border));
  box-shadow: var(--theme-card-shadow-hover);
}

:root[data-brand-theme='claude'] .home-features__index {
  display: block;
  font-family: var(--theme-font-display);
  font-style: italic;
  font-size: 0.875rem;
  color: color-mix(in srgb, var(--theme-accent) 75%, var(--theme-page-muted));
  margin-bottom: 0.75rem;
}

:root[data-brand-theme='claude'] .home-features__index::before {
  content: 'Chapter ';
  color: var(--theme-page-muted);
  font-style: italic;
}

:root[data-brand-theme='claude'] .home-features__title {
  font-family: var(--theme-font-display);
  font-weight: 700;
  font-size: 1.375rem;
  letter-spacing: -0.02em;
}

:root[data-brand-theme='claude'] .home-features__underline {
  display: block;
  width: 40px;
  height: 2px;
  margin-bottom: 0.75rem;
  background: linear-gradient(
    90deg,
    var(--theme-accent),
    color-mix(in srgb, var(--theme-accent) 40%, transparent)
  );
  transition: width 0.3s ease;
}

:root[data-brand-theme='claude'] .home-features__card:hover .home-features__underline {
  width: 80px;
}
</style>
