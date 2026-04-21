<template>
  <section class="home-providers">
    <div class="home-providers__header mb-8 text-center">
      <h2 class="home-providers__title mb-3 font-bold">
        {{ title }}
      </h2>
      <p class="home-providers__description text-sm">
        {{ description }}
      </p>
    </div>

    <div class="home-providers__list mb-16 flex flex-wrap items-center justify-center gap-4">
      <div
        v-for="provider in providers"
        :key="provider.key"
        class="home-providers__chip"
        :class="
          provider.supported
            ? 'home-providers__chip--supported'
            : 'home-providers__chip--unsupported'
        "
      >
        <div
          class="home-providers__avatar"
          :class="`home-providers__avatar--${provider.accentTone}`"
        >
          <span class="home-providers__avatar-text">{{ provider.initial }}</span>
        </div>
        <span class="home-providers__label">{{ provider.label }}</span>
        <span
          class="home-providers__status"
          :class="
            provider.supported
              ? 'home-providers__status--supported'
              : 'home-providers__status--unsupported'
          "
        >
          <span class="home-providers__status-dot" aria-hidden="true"></span>
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

<style scoped>
.home-providers__title,
.home-providers__label {
  color: var(--theme-page-text);
}

.home-providers__title {
  font-family: var(--theme-home-section-title-font);
  font-size: var(--theme-home-section-title-size);
  letter-spacing: var(--theme-home-section-title-letter-spacing);
}

.home-providers__description {
  color: var(--theme-page-muted);
}

.home-providers__chip {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: calc(var(--theme-markdown-block-padding) - 0.25rem) calc(var(--theme-markdown-block-padding) + 0.25rem);
  border-radius: var(--theme-home-provider-radius);
  backdrop-filter: blur(6px);
  transition: transform 0.2s ease, border-color 0.2s ease;
}

.home-providers__chip--supported {
  border: 1px solid color-mix(in srgb, var(--theme-accent) 18%, var(--theme-card-border));
  background: color-mix(in srgb, var(--theme-surface) 72%, transparent);
}

.home-providers__chip--unsupported {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 56%, transparent);
  background: color-mix(in srgb, var(--theme-surface) 40%, transparent);
  opacity: 0.6;
}

.home-providers__avatar {
  display: flex;
  height: 32px;
  width: 32px;
  align-items: center;
  justify-content: center;
  border-radius: var(--theme-home-provider-avatar-radius);
}

.home-providers__avatar-text {
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--theme-filled-text);
}

.home-providers__label {
  font-size: 0.875rem;
  font-weight: 500;
}

.home-providers__status {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: var(--theme-account-usage-pill-padding-y) var(--theme-account-usage-pill-padding-x);
  border-radius: var(--theme-public-action-radius);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.home-providers__status-dot {
  width: 6px;
  height: 6px;
  border-radius: 999px;
}

.home-providers__status--supported {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 14%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.home-providers__status--supported .home-providers__status-dot {
  background: rgb(var(--theme-success-rgb));
  box-shadow: 0 0 8px rgb(var(--theme-success-rgb) / 0.7);
}

.home-providers__status--unsupported {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-muted);
}

.home-providers__status--unsupported .home-providers__status-dot {
  background: var(--theme-page-muted);
  opacity: 0.5;
}

.home-providers__avatar--brand-orange {
  background: linear-gradient(
    135deg,
    rgb(var(--theme-brand-orange-rgb)),
    color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 72%, var(--theme-accent-strong))
  );
}

.home-providers__avatar--success {
  background: linear-gradient(
    135deg,
    rgb(var(--theme-success-rgb)),
    color-mix(in srgb, rgb(var(--theme-success-rgb)) 72%, var(--theme-accent-strong))
  );
}

.home-providers__avatar--info {
  background: linear-gradient(
    135deg,
    rgb(var(--theme-info-rgb)),
    color-mix(in srgb, rgb(var(--theme-info-rgb)) 72%, var(--theme-accent-strong))
  );
}

.home-providers__avatar--brand-rose {
  background: linear-gradient(
    135deg,
    rgb(var(--theme-brand-rose-rgb)),
    color-mix(in srgb, rgb(var(--theme-brand-rose-rgb)) 72%, var(--theme-accent-strong))
  );
}

.home-providers__avatar--neutral {
  background: linear-gradient(
    135deg,
    color-mix(in srgb, var(--theme-page-muted) 84%, var(--theme-page-text)),
    color-mix(in srgb, var(--theme-page-muted) 64%, var(--theme-accent-strong))
  );
}

/* ============== Factory: spec-sheet chip ============== */
:root[data-brand-theme='factory'] .home-providers__title {
  text-transform: uppercase;
}

:root[data-brand-theme='factory'] .home-providers__chip {
  border-radius: 0;
  border-width: 2px;
  background: var(--theme-surface);
  box-shadow: 2px 2px 0 var(--theme-page-text);
  backdrop-filter: none;
}

:root[data-brand-theme='factory'] .home-providers__chip--supported {
  border-color: var(--theme-page-text);
}

:root[data-brand-theme='factory'] .home-providers__chip:hover {
  transform: translate(-1px, -1px);
  box-shadow: 3px 3px 0 var(--theme-page-text);
}

:root[data-brand-theme='factory'] .home-providers__avatar {
  border-radius: 0;
}

:root[data-brand-theme='factory'] .home-providers__label {
  font-family: var(--theme-font-mono);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  font-weight: 700;
}

:root[data-brand-theme='factory'] .home-providers__status {
  border-radius: 0;
  border: 1px solid currentColor;
  font-family: var(--theme-font-mono);
}

.dark[data-brand-theme='factory'] .home-providers__chip {
  border-color: rgba(255, 255, 255, 0.3);
  box-shadow: 2px 2px 0 rgba(255, 255, 255, 0.2);
}

/* ============== Claude: editorial pill ============== */
:root[data-brand-theme='claude'] .home-providers__title {
  font-family: var(--theme-font-display);
}

:root[data-brand-theme='claude'] .home-providers__chip--supported {
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}

:root[data-brand-theme='claude'] .home-providers__chip:hover {
  transform: translateY(-2px);
  border-color: color-mix(in srgb, var(--theme-accent) 40%, var(--theme-page-border));
}

:root[data-brand-theme='claude'] .home-providers__label {
  font-family: var(--theme-font-display);
  font-weight: 700;
}

:root[data-brand-theme='claude'] .home-providers__status {
  text-transform: none;
  letter-spacing: 0.01em;
  font-style: italic;
  font-family: var(--theme-font-display);
  font-weight: 600;
  font-size: 0.72rem;
}
</style>
