<template>
  <div class="embedded-page-layout">
    <div class="card min-h-0 flex-1 overflow-hidden">
      <div v-if="loading" class="flex h-full items-center justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
        ></div>
      </div>

      <div v-else-if="!available" class="embedded-page-state">
        <div class="max-w-md">
          <div class="embedded-page-state-icon">
            <Icon :name="availableIconName" size="lg" class="text-gray-400" />
          </div>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ availableTitle }}
          </h3>
          <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
            {{ availableDescription }}
          </p>
        </div>
      </div>

      <div v-else-if="!isValidUrl" class="embedded-page-state">
        <div class="max-w-md">
          <div class="embedded-page-state-icon">
            <Icon :name="invalidIconName" size="lg" class="text-gray-400" />
          </div>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ invalidTitle }}
          </h3>
          <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
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
  height: calc(100vh - 64px - 4rem);
}

.embedded-page-state {
  @apply flex h-full items-center justify-center p-10 text-center;
}

.embedded-page-state-icon {
  @apply mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700;
}

.embedded-page-shell {
  @apply relative h-full w-full overflow-hidden rounded-2xl bg-gradient-to-b from-gray-50 to-white p-0 dark:from-dark-900 dark:to-dark-950;
}

.embedded-page-open-fab {
  @apply absolute right-3 top-3 z-10 shadow-sm backdrop-blur supports-[backdrop-filter]:bg-white/80 dark:supports-[backdrop-filter]:bg-dark-800/80;
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
</style>
