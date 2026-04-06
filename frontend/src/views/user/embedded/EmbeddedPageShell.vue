<template>
  <div class="embedded-page-layout">
    <div class="card min-h-0 flex-1 overflow-hidden">
      <div v-if="loading" class="embedded-page-layout__loading-state">
        <div
          class="embedded-page-spinner h-8 w-8 animate-spin rounded-full border-2 border-t-transparent"
        ></div>
      </div>

      <div v-else-if="!available" class="embedded-page-state">
        <div class="max-w-md">
          <div class="embedded-page-state-icon">
            <Icon :name="availableIconName" size="lg" class="embedded-page-state-icon__symbol" />
          </div>
          <h3 class="embedded-page-state__title text-lg font-semibold">
            {{ availableTitle }}
          </h3>
          <p class="embedded-page-state__description mt-2 text-sm">
            {{ availableDescription }}
          </p>
        </div>
      </div>

      <div v-else-if="!isValidUrl" class="embedded-page-state">
        <div class="max-w-md">
          <div class="embedded-page-state-icon">
            <Icon :name="invalidIconName" size="lg" class="embedded-page-state-icon__symbol" />
          </div>
          <h3 class="embedded-page-state__title text-lg font-semibold">
            {{ invalidTitle }}
          </h3>
          <p class="embedded-page-state__description mt-2 text-sm">
            {{ invalidDescription }}
          </p>
        </div>
      </div>

      <div v-else class="embedded-page-shell">
        <a
          :href="embeddedUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="btn btn-secondary btn-sm embedded-page-open-fab"
        >
          <Icon name="externalLink" size="sm" class="mr-1.5" :stroke-width="2" />
          {{ openInNewTabLabel }}
        </a>
        <iframe :src="embeddedUrl" class="embedded-page-frame" allowfullscreen></iframe>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

type EmbeddedPageStateIconName = 'creditCard' | 'link'

withDefaults(
  defineProps<{
    loading: boolean
    available: boolean
    availableIconName: EmbeddedPageStateIconName
    availableTitle: string
    availableDescription: string
    isValidUrl: boolean
    invalidTitle: string
    invalidDescription: string
    embeddedUrl: string
    openInNewTabLabel: string
    invalidIconName?: EmbeddedPageStateIconName
  }>(),
  {
    invalidIconName: 'link'
  }
)
</script>

<style scoped>
.embedded-page-layout {
  @apply flex flex-col;
  height: calc(100vh - var(--theme-shell-header-height) - var(--theme-embedded-page-height-offset));
}

.embedded-page-layout__loading-state {
  @apply flex h-full items-center justify-center;
  padding: var(--theme-embedded-page-state-padding);
}

.embedded-page-state {
  @apply flex h-full items-center justify-center text-center;
  padding: var(--theme-embedded-page-state-padding);
}

.embedded-page-state-icon {
  @apply mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full;
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.embedded-page-state-icon__symbol {
  color: color-mix(in srgb, var(--theme-page-muted) 70%, transparent);
}

.embedded-page-state__title {
  color: var(--theme-page-text);
}

.embedded-page-state__description {
  color: var(--theme-page-muted);
}

.embedded-page-shell {
  @apply relative h-full w-full overflow-hidden;
  border-radius: var(--theme-embedded-page-shell-radius);
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--theme-surface-soft) 90%, var(--theme-surface)) 0%,
    var(--theme-surface) 100%
  );
}

.embedded-page-open-fab {
  @apply absolute z-10 shadow-sm backdrop-blur;
  top: var(--theme-embedded-page-fab-offset);
  right: var(--theme-embedded-page-fab-offset);
  background: color-mix(in srgb, var(--theme-page-backdrop) 88%, transparent);
}

.embedded-page-frame {
  display: block;
  margin: 0;
  width: 100%;
  height: 100%;
  border: 0;
  border-radius: 0;
  box-shadow: none;
  background: transparent;
}

.embedded-page-spinner {
  border-color: var(--theme-accent);
}
</style>
