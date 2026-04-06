<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useClipboard } from '@/composables/useClipboard'
import type { CustomEndpoint } from '@/types'

const props = defineProps<{
  apiBaseUrl: string
  customEndpoints: CustomEndpoint[]
}>()

const { t } = useI18n()
const { copyToClipboard } = useClipboard()
const copiedEndpoint = ref<string | null>(null)

let copiedResetTimer: number | undefined

const allEndpoints = computed(() => {
  const items: Array<{ name: string; endpoint: string; description: string; isDefault: boolean }> = []
  if (props.apiBaseUrl) {
    items.push({
      name: t('keys.endpoints.title'),
      endpoint: props.apiBaseUrl,
      description: '',
      isDefault: true,
    })
  }
  for (const ep of props.customEndpoints) {
    items.push({ ...ep, isDefault: false })
  }
  return items
})

async function copy(url: string) {
  const success = await copyToClipboard(url, t('keys.endpoints.copied'))
  if (!success) return

  copiedEndpoint.value = url
  if (copiedResetTimer !== undefined) {
    window.clearTimeout(copiedResetTimer)
  }
  copiedResetTimer = window.setTimeout(() => {
    if (copiedEndpoint.value === url) {
      copiedEndpoint.value = null
    }
  }, 1800)
}

function tooltipHint(endpoint: string): string {
  return copiedEndpoint.value === endpoint
    ? t('keys.endpoints.copiedHint')
    : t('keys.endpoints.clickToCopy')
}

function speedTestUrl(endpoint: string): string {
  return `https://www.tcptest.cn/http/${encodeURIComponent(endpoint)}`
}

onBeforeUnmount(() => {
  if (copiedResetTimer !== undefined) {
    window.clearTimeout(copiedResetTimer)
  }
})
</script>

<template>
  <div v-if="allEndpoints.length > 0" class="flex flex-wrap gap-2">
    <div
      v-for="(item, index) in allEndpoints"
      :key="index"
      class="endpoint-popover__chip flex items-center gap-1.5 text-xs transition-colors"
    >
      <span class="endpoint-popover__name font-medium">{{ item.name }}</span>
      <span
        v-if="item.isDefault"
        class="endpoint-popover__default-tag rounded text-[10px] font-medium leading-tight"
      >{{ t('keys.endpoints.default') }}</span>

      <span class="endpoint-popover__divider">|</span>

      <div class="group/endpoint relative flex items-center gap-1.5">
        <div
          class="endpoint-popover__tooltip pointer-events-none absolute bottom-full left-1/2 z-20 mb-2 w-max -translate-x-1/2 translate-y-1 text-left opacity-0 transition-all duration-150 group-hover/endpoint:translate-y-0 group-hover/endpoint:opacity-100 group-focus-within/endpoint:translate-y-0 group-focus-within/endpoint:opacity-100"
        >
          <p
            v-if="item.description"
            class="endpoint-popover__tooltip-description break-words text-xs leading-5"
          >
            {{ item.description }}
          </p>
          <p
            class="endpoint-popover__tooltip-hint flex items-center gap-1.5 text-[11px] leading-4"
            :class="item.description ? 'mt-1.5' : ''"
          >
            <span class="endpoint-popover__tooltip-dot h-1.5 w-1.5 rounded-full"></span>
            {{ tooltipHint(item.endpoint) }}
          </p>
          <div class="endpoint-popover__tooltip-arrow absolute left-1/2 top-full h-3 w-3 -translate-x-1/2 -translate-y-1/2 rotate-45 border-b border-r"></div>
        </div>

        <code
          class="endpoint-popover__endpoint cursor-pointer font-mono decoration-dashed underline-offset-2 focus:outline-none"
          role="button"
          tabindex="0"
          @click="copy(item.endpoint)"
          @keydown.enter.prevent="copy(item.endpoint)"
          @keydown.space.prevent="copy(item.endpoint)"
        >{{ item.endpoint }}</code>

        <button
          type="button"
          class="endpoint-popover__icon-button transition-colors"
          :class="copiedEndpoint === item.endpoint
            ? 'endpoint-popover__icon-button--copied'
            : 'endpoint-popover__icon-button--idle'"
          :aria-label="tooltipHint(item.endpoint)"
          @click="copy(item.endpoint)"
        >
          <svg v-if="copiedEndpoint === item.endpoint" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
          </svg>
          <svg v-else class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
        </button>

        <a
          :href="speedTestUrl(item.endpoint)"
          target="_blank"
          rel="noopener noreferrer"
          class="endpoint-popover__icon-button endpoint-popover__icon-button--speed transition-colors"
          :title="t('keys.endpoints.speedTest')"
        >
          <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
        </a>
      </div>
    </div>
  </div>
</template>

<style scoped>
.endpoint-popover__chip {
  padding: var(--theme-endpoint-popover-chip-padding-y) var(--theme-endpoint-popover-chip-padding-x);
  border-radius: var(--theme-button-radius);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 74%, transparent);
  background: var(--theme-surface);
}

.endpoint-popover__chip:hover {
  border-color: color-mix(in srgb, var(--theme-accent) 24%, var(--theme-card-border));
}

.endpoint-popover__name {
  color: var(--theme-page-text);
}

.endpoint-popover__default-tag {
  padding: var(--theme-endpoint-popover-default-tag-padding-y)
    var(--theme-endpoint-popover-default-tag-padding-x);
  background: color-mix(in srgb, var(--theme-accent-soft) 86%, var(--theme-surface));
  color: var(--theme-accent);
}

.endpoint-popover__divider {
  color: color-mix(in srgb, var(--theme-page-muted) 42%, var(--theme-surface));
}

.endpoint-popover__tooltip {
  padding: var(--theme-endpoint-popover-tooltip-padding-y)
    var(--theme-endpoint-popover-tooltip-padding-x);
  max-width: max(24rem, calc(var(--theme-tooltip-width) * 1.5));
  border-radius: var(--theme-markdown-block-radius);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 70%, transparent);
  background: var(--theme-surface);
  box-shadow: var(--theme-dropdown-shadow);
}

.endpoint-popover__tooltip-description {
  max-width: max(24rem, calc(var(--theme-tooltip-width) * 1.5));
  color: var(--theme-page-text);
}

.endpoint-popover__tooltip-hint {
  color: var(--theme-accent);
}

.endpoint-popover__tooltip-dot {
  background: var(--theme-accent);
}

.endpoint-popover__tooltip-arrow {
  border-color: color-mix(in srgb, var(--theme-card-border) 70%, transparent);
  background: var(--theme-surface);
}

.endpoint-popover__endpoint {
  color: var(--theme-page-muted);
  text-decoration-color: color-mix(in srgb, var(--theme-page-muted) 62%, transparent);
}

.endpoint-popover__endpoint:hover,
.endpoint-popover__endpoint:focus {
  color: var(--theme-accent);
  text-decoration-line: underline;
}

.endpoint-popover__icon-button {
  padding: var(--theme-endpoint-popover-icon-button-padding);
  border-radius: calc(var(--theme-button-radius) - 2px);
}

.endpoint-popover__icon-button--idle {
  color: var(--theme-page-muted);
}

.endpoint-popover__icon-button--idle:hover {
  color: var(--theme-accent);
}

.endpoint-popover__icon-button--copied {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.endpoint-popover__icon-button--speed {
  color: var(--theme-page-muted);
}

.endpoint-popover__icon-button--speed:hover {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}
</style>
