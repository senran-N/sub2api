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
  padding: var(--theme-home-hero-eyebrow-padding);
  border: var(--theme-home-hero-eyebrow-border);
  background: var(--theme-home-hero-eyebrow-bg);
  box-shadow: var(--theme-home-hero-eyebrow-shadow);
  border-radius: var(--theme-home-hero-eyebrow-radius);
  font-family: var(--theme-home-hero-eyebrow-font);
  font-style: var(--theme-home-hero-eyebrow-font-style);
  font-size: var(--theme-home-hero-eyebrow-font-size);
  font-weight: var(--theme-home-hero-eyebrow-font-weight);
  letter-spacing: var(--theme-home-hero-eyebrow-letter-spacing);
  text-transform: var(--theme-home-hero-eyebrow-transform);
  color: var(--theme-home-hero-eyebrow-text-color);
}

.home-hero__eyebrow-dot {
  width: var(--theme-home-hero-dot-size);
  height: var(--theme-home-hero-dot-size);
  border-radius: var(--theme-home-hero-dot-radius);
  background: var(--theme-home-hero-dot-bg);
  box-shadow: var(--theme-home-hero-dot-shadow);
  animation: var(--theme-home-hero-dot-animation);
}

.home-hero__eyebrow-text {
  color: inherit;
}

.home-hero__eyebrow-text::before {
  content: var(--theme-home-hero-eyebrow-prefix);
  color: var(--theme-home-hero-eyebrow-prefix-color);
}

.home-hero__title {
  font-family: var(--theme-home-title-font);
  font-weight: var(--theme-home-title-weight);
  letter-spacing: var(--theme-home-title-letter-spacing);
  line-height: 1.05;
  text-transform: var(--theme-home-title-transform);
}

.home-hero__subtitle {
  font-family: var(--theme-home-subtitle-font);
  font-style: var(--theme-home-hero-subtitle-font-style);
  letter-spacing: var(--theme-home-hero-subtitle-letter-spacing);
  position: relative;
  padding-left: var(--theme-home-hero-subtitle-padding-left);
}

.home-hero__subtitle::before {
  content: var(--theme-home-hero-subtitle-prefix);
  position: absolute;
  left: 0;
  color: var(--theme-home-hero-subtitle-prefix-color);
  font-weight: 700;
}

.home-hero__ornament {
  margin: 0 auto 1.25rem;
  width: var(--theme-home-hero-ornament-width);
  max-width: var(--theme-home-hero-ornament-max-width);
  height: var(--theme-home-hero-ornament-height);
  background: var(--theme-home-hero-ornament-bg);
  display: var(--theme-home-hero-ornament-display);
  align-items: center;
  justify-content: var(--theme-home-hero-ornament-justify);
  color: var(--theme-home-hero-ornament-color);
  position: relative;
}

.home-hero__ornament::before,
.home-hero__ornament::after {
  content: '';
  flex: 1;
  height: 1px;
}

.home-hero__ornament::before {
  display: var(--theme-home-hero-ornament-before-display);
  background: var(--theme-home-hero-ornament-before-bg);
  margin: var(--theme-home-hero-ornament-before-margin);
}

.home-hero__ornament::after {
  display: var(--theme-home-hero-ornament-after-display);
  background: var(--theme-home-hero-ornament-after-bg);
  margin: var(--theme-home-hero-ornament-after-margin);
}

.home-hero__cta {
  padding:
    calc(var(--theme-markdown-block-padding) - 0.25rem)
    calc(var(--theme-auth-card-padding) - 0.5rem);
  box-shadow: var(--theme-home-hero-cta-shadow);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.home-hero__cta:hover {
  transform: var(--theme-home-hero-cta-hover-transform);
  box-shadow: var(--theme-home-hero-cta-hover-shadow);
}

.home-hero__cta:active {
  transform: var(--theme-home-hero-cta-active-transform);
  box-shadow: var(--theme-home-hero-cta-active-shadow);
}

@media (min-width: 1024px) {
  .home-hero__ornament {
    margin-left: 0;
    justify-content: var(--theme-home-hero-ornament-desktop-justify);
  }

  .home-hero__ornament::before {
    display: var(--theme-home-hero-ornament-before-desktop-display);
  }
}

@keyframes home-hero-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.35; }
}

@media (prefers-reduced-motion: reduce) {
  .home-hero__eyebrow-dot {
    animation: none;
  }
}
</style>
