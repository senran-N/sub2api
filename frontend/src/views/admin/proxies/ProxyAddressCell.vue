<template>
  <div class="flex items-center gap-1.5">
    <code class="code text-xs">{{ proxy.host }}:{{ proxy.port }}</code>
    <div class="relative">
      <button
        type="button"
        class="proxy-address-cell__copy"
        :title="t('admin.proxies.copyProxyUrl')"
        @click.stop="emit('copy-url', proxy)"
        @contextmenu.prevent="emit('toggle-copy-menu', proxy.id)"
      >
        <Icon name="copy" size="sm" />
      </button>
      <div
        v-if="copyMenuOpen"
        class="proxy-address-cell__menu absolute left-0 top-full z-50 w-auto"
      >
        <button
          v-for="format in copyFormats"
          :key="format.label"
          class="proxy-address-cell__menu-item flex w-full items-center gap-2 text-left text-xs"
          @click.stop="emit('copy-format', format.value)"
        >
          <span class="proxy-address-cell__menu-label truncate font-mono">{{ format.label }}</span>
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

<style scoped>
.proxy-address-cell__copy {
  border-radius: var(--theme-proxy-address-copy-radius);
  padding: var(--theme-proxy-address-copy-padding);
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
  transition: color 0.2s ease;
}

.proxy-address-cell__copy:hover {
  color: var(--theme-accent);
}

.proxy-address-cell__menu {
  min-width: var(--theme-proxy-copy-menu-min-width);
  margin-top: var(--theme-proxy-address-menu-offset);
  padding-block: var(--theme-proxy-address-menu-padding-y);
  border: 1px solid color-mix(in srgb, var(--theme-dropdown-border) 88%, transparent);
  border-radius: var(--theme-proxy-address-menu-radius);
  background: var(--theme-dropdown-bg);
  box-shadow: var(--theme-dropdown-shadow);
}

.proxy-address-cell__menu-item {
  padding: var(--theme-proxy-address-menu-item-padding-y)
    var(--theme-proxy-address-menu-item-padding-x);
  transition: background-color 0.2s ease;
}

.proxy-address-cell__menu-item:hover {
  background: var(--theme-dropdown-item-hover-bg);
}

.proxy-address-cell__menu-label {
  color: var(--theme-dropdown-text);
}
</style>
