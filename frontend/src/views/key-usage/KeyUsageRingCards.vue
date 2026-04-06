<template>
  <div v-if="items.length > 0" :class="gridClass">
    <div
      v-for="(ring, index) in items"
      :key="index"
      class="key-usage-ring-cards__card fade-up border backdrop-blur-sm transition-all duration-300 hover:shadow-lg"
      :class="`fade-up-delay-${Math.min(index + 1, 4)}`"
    >
      <div class="mb-6 flex items-center justify-between">
        <h3 class="key-usage-ring-cards__label text-sm font-semibold uppercase tracking-wider">
          {{ ring.title }}
        </h3>
        <svg
          v-if="ring.iconType === 'clock'"
          class="key-usage-ring-cards__icon h-5 w-5"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <circle cx="12" cy="12" r="10" />
          <polyline points="12 6 12 12 16 14" />
        </svg>
        <svg
          v-else-if="ring.iconType === 'calendar'"
          class="key-usage-ring-cards__icon h-5 w-5"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
          <line x1="16" y1="2" x2="16" y2="6" />
          <line x1="8" y1="2" x2="8" y2="6" />
          <line x1="3" y1="10" x2="21" y2="10" />
        </svg>
        <svg
          v-else
          class="key-usage-ring-cards__icon h-5 w-5"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <line x1="12" y1="1" x2="12" y2="23" />
          <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
        </svg>
      </div>

      <div class="flex justify-center">
        <div class="relative">
          <svg class="h-44 w-44" viewBox="0 0 160 160">
            <circle cx="80" cy="80" r="68" fill="none" :stroke="trackColor" stroke-width="10" />
            <circle
              class="progress-ring"
              cx="80"
              cy="80"
              r="68"
              fill="none"
              :stroke="`url(#ring-grad-${index})`"
              stroke-width="10"
              stroke-linecap="round"
              :stroke-dasharray="circumference.toFixed(2)"
              :stroke-dashoffset="getRingOffset(ring)"
            />
            <defs>
              <linearGradient :id="`ring-grad-${index}`" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" :stop-color="gradients[index % gradients.length].from" />
                <stop offset="100%" :stop-color="gradients[index % gradients.length].to" />
              </linearGradient>
            </defs>
          </svg>

          <div class="absolute inset-0 flex flex-col items-center justify-center">
            <template v-if="ring.isBalance">
              <span
                class="text-2xl font-bold tabular-nums"
                :style="{ color: gradients[index % gradients.length].from }"
              >
                {{ ring.amount }}
              </span>
            </template>
            <template v-else>
              <span class="key-usage-ring-cards__value text-3xl font-bold tabular-nums">
                {{ displayPcts[index] ?? 0 }}%
              </span>
              <span class="key-usage-ring-cards__label mt-0.5 text-xs">{{ usedLabel }}</span>
              <span
                class="mt-1 text-sm font-semibold tabular-nums"
                :style="{ color: gradients[index % gradients.length].from }"
              >
                {{ ring.amount }}
              </span>
              <p
                v-if="ring.resetAt && formatResetTime(ring.resetAt)"
                class="key-usage-ring-cards__hint mt-0.5 text-xs tabular-nums"
              >
                ⟳ {{ formatResetTime(ring.resetAt) }}
              </p>
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { KeyUsageRingItem } from './keyUsageView'

const props = defineProps<{
  items: KeyUsageRingItem[]
  gridClass: string
  displayPcts: number[]
  usedLabel: string
  trackColor: string
  circumference: number
  gradients: ReadonlyArray<{ from: string; to: string }>
  animated: boolean
  formatResetTime: (value: string | null | undefined) => string
}>()

function getRingOffset(ring: KeyUsageRingItem): number {
  if (!props.animated) {
    return props.circumference
  }
  if (ring.isBalance) {
    return 0
  }

  return props.circumference - (Math.min(ring.pct, 100) / 100) * props.circumference
}
</script>

<style scoped>
.key-usage-ring-cards__card {
  border-radius: var(--theme-key-usage-card-radius);
  padding: var(--theme-key-usage-loading-card-padding);
  border-color: color-mix(in srgb, var(--theme-card-border) 84%, transparent);
  background: color-mix(in srgb, var(--theme-surface) 90%, transparent);
}

.key-usage-ring-cards__label,
.key-usage-ring-cards__hint,
.key-usage-ring-cards__icon {
  color: var(--theme-page-muted);
}

.key-usage-ring-cards__value {
  color: var(--theme-page-text);
}
</style>
