<template>
  <div class="flex items-center gap-1.5">
    <code class="code text-xs">{{ proxy.host }}:{{ proxy.port }}</code>
    <div class="relative">
      <button
        type="button"
        class="rounded p-0.5 text-gray-400 hover:text-primary-600 dark:hover:text-primary-400"
        :title="t('admin.proxies.copyProxyUrl')"
        @click.stop="emit('copy-url', proxy)"
        @contextmenu.prevent="emit('toggle-copy-menu', proxy.id)"
      >
        <Icon name="copy" size="sm" />
      </button>
      <div
        v-if="copyMenuOpen"
        class="absolute left-0 top-full z-50 mt-1 w-auto min-w-[180px] rounded-lg border border-gray-200 bg-white py-1 shadow-lg dark:border-dark-500 dark:bg-dark-700"
      >
        <button
          v-for="format in copyFormats"
          :key="format.label"
          class="flex w-full items-center gap-2 px-3 py-1.5 text-left text-xs hover:bg-gray-100 dark:hover:bg-dark-600"
          @click.stop="emit('copy-format', format.value)"
        >
          <span class="truncate font-mono text-gray-600 dark:text-gray-300">{{ format.label }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { Proxy } from '@/types'
import type { ProxyCopyFormat } from '../proxyUtils'

defineProps<{
  proxy: Proxy
  copyMenuOpen: boolean
  copyFormats: ProxyCopyFormat[]
}>()

const emit = defineEmits<{
  'copy-url': [proxy: Proxy]
  'toggle-copy-menu': [proxyId: number]
  'copy-format': [value: string]
}>()

const { t } = useI18n()
</script>
