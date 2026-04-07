<template>
  <div class="flex items-center gap-2">
    <img
      v-if="proxy.country_code"
      :src="buildProxyFlagUrl(proxy.country_code)"
      :alt="proxy.country || proxy.country_code"
      class="h-4 w-6 rounded-sm"
    />
    <span v-if="locationLabel" class="theme-text-default text-sm">
      {{ locationLabel }}
    </span>
    <span v-else class="theme-text-subtle text-sm">-</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Proxy } from '@/types'
import { buildProxyFlagUrl, formatProxyLocation } from './proxyPresentation'

const props = defineProps<{
  proxy: Proxy
}>()

const locationLabel = computed(() => formatProxyLocation(props.proxy))
</script>
