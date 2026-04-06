<template>
  <section>
    <div class="mb-8 text-center">
      <h2 class="home-providers-section__title mb-3 font-bold">
        {{ title }}
      </h2>
      <p class="home-providers-section__description text-sm">
        {{ description }}
      </p>
    </div>

    <div class="mb-16 flex flex-wrap items-center justify-center gap-4">
      <div
        v-for="provider in providers"
        :key="provider.key"
        class="home-providers-section__provider flex items-center gap-2 backdrop-blur-sm"
        :class="
          provider.supported
            ? 'home-providers-section__provider--supported'
            : 'home-providers-section__provider--unsupported'
        "
        >
          <div
          class="home-providers-section__avatar flex h-8 w-8 items-center justify-center"
          :class="`home-providers-section__avatar--${provider.accentTone}`"
        >
          <span class="home-providers-section__avatar-text text-xs font-bold">{{ provider.initial }}</span>
        </div>
        <span class="home-providers-section__provider-label text-sm font-medium">{{ provider.label }}</span>
        <span
          class="home-providers-section__status text-[10px] font-medium"
          :class="
            provider.supported
              ? 'home-providers-section__status--supported'
              : 'home-providers-section__status--unsupported'
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

<style scoped>
.home-providers-section__title,
.home-providers-section__provider-label {
  color: var(--theme-page-text);
}

.home-providers-section__title {
  font-family: var(--theme-home-section-title-font);
  font-size: var(--theme-home-section-title-size);
  letter-spacing: var(--theme-home-section-title-letter-spacing);
}

.home-providers-section__description {
  color: var(--theme-page-muted);
}

.home-providers-section__provider--supported {
  padding: calc(var(--theme-markdown-block-padding) - 0.25rem) calc(var(--theme-markdown-block-padding) + 0.25rem);
  border-radius: var(--theme-home-provider-radius);
  border: 1px solid color-mix(in srgb, var(--theme-accent) 18%, var(--theme-card-border));
  background: color-mix(in srgb, var(--theme-surface) 60%, transparent);
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--theme-accent) 18%, transparent);
}

.home-providers-section__provider--unsupported {
  padding: calc(var(--theme-markdown-block-padding) - 0.25rem) calc(var(--theme-markdown-block-padding) + 0.25rem);
  border-radius: var(--theme-home-provider-radius);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 56%, transparent);
  background: color-mix(in srgb, var(--theme-surface) 40%, transparent);
  opacity: 0.6;
}

.home-providers-section__avatar {
  border-radius: var(--theme-home-provider-avatar-radius);
}

.home-providers-section__status--supported {
  border-radius: var(--theme-public-action-radius);
  padding: var(--theme-account-usage-pill-padding-y) var(--theme-account-usage-pill-padding-x);
  background: color-mix(in srgb, var(--theme-accent-soft) 82%, var(--theme-surface));
  color: var(--theme-accent);
}

.home-providers-section__status--unsupported {
  border-radius: var(--theme-public-action-radius);
  padding: var(--theme-account-usage-pill-padding-y) var(--theme-account-usage-pill-padding-x);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-muted);
}

.home-providers-section__avatar-text {
  color: var(--theme-filled-text);
}

.home-providers-section__avatar--brand-orange {
  background: linear-gradient(
    135deg,
    rgb(var(--theme-brand-orange-rgb)),
    color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 72%, var(--theme-accent-strong))
  );
}

.home-providers-section__avatar--success {
  background: linear-gradient(
    135deg,
    rgb(var(--theme-success-rgb)),
    color-mix(in srgb, rgb(var(--theme-success-rgb)) 72%, var(--theme-accent-strong))
  );
}

.home-providers-section__avatar--info {
  background: linear-gradient(
    135deg,
    rgb(var(--theme-info-rgb)),
    color-mix(in srgb, rgb(var(--theme-info-rgb)) 72%, var(--theme-accent-strong))
  );
}

.home-providers-section__avatar--brand-rose {
  background: linear-gradient(
    135deg,
    rgb(var(--theme-brand-rose-rgb)),
    color-mix(in srgb, rgb(var(--theme-brand-rose-rgb)) 72%, var(--theme-accent-strong))
  );
}

.home-providers-section__avatar--neutral {
  background: linear-gradient(
    135deg,
    color-mix(in srgb, var(--theme-page-muted) 84%, var(--theme-page-text)),
    color-mix(in srgb, var(--theme-page-muted) 64%, var(--theme-accent-strong))
  );
}
</style>
