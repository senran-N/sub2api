<template>
  <div class="home-hero mb-12 flex flex-col items-center justify-between gap-12 lg:flex-row lg:gap-16">
    <div class="home-hero__text flex-1 text-center lg:text-left">
      <div class="home-hero__eyebrow">
        <span class="home-hero__eyebrow-dot"></span>
        <span class="home-hero__eyebrow-text">{{ eyebrowLabel }}</span>
      </div>
      <h1 class="home-hero__title theme-text-strong mb-4 text-4xl md:text-5xl lg:text-6xl">
        {{ siteName }}
      </h1>
      <div class="home-hero__ornament" aria-hidden="true"></div>
      <p class="home-hero__subtitle theme-text-muted mb-8 text-lg md:text-xl">
        {{ siteSubtitle }}
      </p>

      <div>
        <router-link
          :to="ctaPath"
          class="home-hero__cta btn btn-primary text-base"
        >
          {{ ctaLabel }}
          <Icon name="arrowRight" size="md" class="ml-2" :stroke-width="2" />
        </router-link>
      </div>
    </div>

    <div class="home-hero__demo flex flex-1 justify-center lg:justify-end">
      <HomeTerminalDemo />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import HomeTerminalDemo from './HomeTerminalDemo.vue'

defineProps<{
  ctaLabel: string
  ctaPath: string
  siteName: string
  siteSubtitle: string
}>()

// The hero eyebrow surfaces a "live" badge that each theme dresses up via CSS:
// factory renders it as a blueprint status pill, claude renders it as an
// editorial dateline. The copy stays neutral so translators only maintain one
// string.
const { t } = useI18n()
const eyebrowLabel = computed(() => t('home.status.online'))
</script>

<style scoped>
.home-hero__eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
  font-family: var(--theme-font-mono);
  font-size: 0.72rem;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.home-hero__eyebrow-dot {
  width: 8px;
  height: 8px;
  background: rgb(var(--theme-success-rgb));
  box-shadow: 0 0 12px rgb(var(--theme-success-rgb) / 0.8);
}

.home-hero__eyebrow-text {
  color: var(--theme-page-muted);
}

.home-hero__title {
  font-family: var(--theme-home-title-font);
  font-weight: var(--theme-home-title-weight);
  letter-spacing: var(--theme-home-title-letter-spacing);
  line-height: 1.05;
}

.home-hero__subtitle {
  font-family: var(--theme-home-subtitle-font);
}

.home-hero__ornament {
  margin: 0 auto 1.25rem;
  width: 180px;
  height: 2px;
}

.home-hero__cta {
  padding:
    calc(var(--theme-markdown-block-padding) - 0.25rem)
    calc(var(--theme-auth-card-padding) - 0.5rem);
}

@media (min-width: 1024px) {
  .home-hero__ornament {
    margin-left: 0;
  }
}

/* ============== Factory: terminal-status eyebrow, blueprint ornament ============== */
:root[data-brand-theme='factory'] .home-hero__eyebrow {
  padding: 6px 12px;
  border: 2px solid var(--theme-page-text);
  background: var(--theme-surface);
  box-shadow: 3px 3px 0 var(--theme-page-text);
  border-radius: 0;
}

:root[data-brand-theme='factory'] .home-hero__eyebrow-dot {
  border-radius: 0;
  width: 10px;
  height: 10px;
  animation: home-hero-pulse 1.8s ease-in-out infinite;
}

:root[data-brand-theme='factory'] .home-hero__eyebrow-text {
  color: var(--theme-page-text);
  font-weight: 700;
}

:root[data-brand-theme='factory'] .home-hero__title {
  text-transform: uppercase;
}

:root[data-brand-theme='factory'] .home-hero__ornament {
  width: 100%;
  max-width: 280px;
  height: 8px;
  background-image:
    linear-gradient(90deg, var(--theme-page-text) 0 24px, transparent 24px 36px),
    repeating-linear-gradient(
      90deg,
      var(--theme-page-text) 0 2px,
      transparent 2px 10px
    );
  background-size: 36px 2px, 100% 2px;
  background-position: 0 0, 0 100%;
  background-repeat: repeat-x;
}

:root[data-brand-theme='factory'] .home-hero__subtitle {
  font-family: var(--theme-font-mono);
  position: relative;
  padding-left: 1.25rem;
}

:root[data-brand-theme='factory'] .home-hero__subtitle::before {
  content: '>';
  position: absolute;
  left: 0;
  color: var(--theme-accent);
  font-weight: 700;
}

:root[data-brand-theme='factory'] .home-hero__cta {
  box-shadow: 6px 6px 0 var(--theme-page-text);
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

:root[data-brand-theme='factory'] .home-hero__cta:hover {
  transform: translate(-2px, -2px);
  box-shadow: 8px 8px 0 var(--theme-page-text);
}

:root[data-brand-theme='factory'] .home-hero__cta:active {
  transform: translate(2px, 2px);
  box-shadow: 2px 2px 0 var(--theme-page-text);
}

.dark[data-brand-theme='factory'] .home-hero__eyebrow {
  background: var(--theme-surface-muted);
  border-color: rgba(255, 255, 255, 0.4);
  box-shadow: 3px 3px 0 rgba(255, 255, 255, 0.3);
}

.dark[data-brand-theme='factory'] .home-hero__cta {
  box-shadow: 6px 6px 0 rgba(255, 255, 255, 0.3);
}

.dark[data-brand-theme='factory'] .home-hero__cta:hover {
  box-shadow: 8px 8px 0 rgba(255, 255, 255, 0.3);
}

.dark[data-brand-theme='factory'] .home-hero__ornament {
  background-image:
    linear-gradient(90deg, var(--theme-page-text) 0 24px, transparent 24px 36px),
    repeating-linear-gradient(
      90deg,
      var(--theme-page-text) 0 2px,
      transparent 2px 10px
    );
}

/* ============== Claude: editorial issue badge, ornament divider ============== */
:root[data-brand-theme='claude'] .home-hero__eyebrow {
  font-family: var(--theme-font-display);
  font-style: italic;
  font-size: 0.875rem;
  letter-spacing: 0.02em;
  text-transform: none;
  color: color-mix(in srgb, var(--theme-accent) 82%, var(--theme-page-text));
}

:root[data-brand-theme='claude'] .home-hero__eyebrow-dot {
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: var(--theme-accent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--theme-accent) 20%, transparent);
}

:root[data-brand-theme='claude'] .home-hero__eyebrow-text::before {
  content: '— ';
  color: var(--theme-accent);
}

:root[data-brand-theme='claude'] .home-hero__title {
  font-family: var(--theme-font-display);
}

:root[data-brand-theme='claude'] .home-hero__ornament {
  width: auto;
  max-width: 220px;
  height: 1rem;
  background: none;
  display: flex;
  align-items: center;
  justify-content: center;
  color: color-mix(in srgb, var(--theme-accent) 70%, var(--theme-page-muted));
  position: relative;
}

:root[data-brand-theme='claude'] .home-hero__ornament::before,
:root[data-brand-theme='claude'] .home-hero__ornament::after {
  content: '';
  flex: 1;
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent,
    color-mix(in srgb, var(--theme-accent) 40%, var(--theme-page-border)) 50%,
    transparent
  );
}

:root[data-brand-theme='claude'] .home-hero__ornament::after {
  margin-left: 8px;
}

:root[data-brand-theme='claude'] .home-hero__ornament::before {
  margin-right: 8px;
}

@media (min-width: 1024px) {
  :root[data-brand-theme='claude'] .home-hero__ornament {
    margin-left: 0;
    justify-content: flex-start;
  }

  :root[data-brand-theme='claude'] .home-hero__ornament::before {
    display: none;
  }
}

:root[data-brand-theme='claude'] .home-hero__subtitle {
  font-style: italic;
  font-family: var(--theme-font-display);
  letter-spacing: -0.01em;
}

:root[data-brand-theme='claude'] .home-hero__cta {
  box-shadow: 0 14px 28px color-mix(in srgb, var(--theme-accent) 22%, transparent);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

:root[data-brand-theme='claude'] .home-hero__cta:hover {
  transform: translateY(-2px);
  box-shadow: 0 20px 38px color-mix(in srgb, var(--theme-accent) 32%, transparent);
}

@keyframes home-hero-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.35; }
}

@media (prefers-reduced-motion: reduce) {
  :root[data-brand-theme='factory'] .home-hero__eyebrow-dot {
    animation: none;
  }
}
</style>
