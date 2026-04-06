<template>
  <div v-if="proxy.username || proxy.password" class="flex items-center gap-1.5">
    <div class="flex flex-col text-xs">
      <span v-if="proxy.username" class="proxy-auth-cell__username">{{ proxy.username }}</span>
      <span v-if="proxy.password" class="proxy-auth-cell__password font-mono">
        {{ passwordVisible ? proxy.password : '••••••' }}
      </span>
    </div>
    <button
      v-if="proxy.password"
      type="button"
      class="proxy-auth-cell__toggle"
      @click.stop="emit('toggle-password', proxy.id)"
    >
      <Icon :name="passwordVisible ? 'eyeOff' : 'eye'" size="sm" />
    </button>
  </div>
  <span v-else class="proxy-auth-cell__empty text-sm">-</span>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'
import type { Proxy } from '@/types'

defineProps<{
  proxy: Proxy
  passwordVisible: boolean
}>()

const emit = defineEmits<{
  'toggle-password': [proxyId: number]
}>()
</script>

<style scoped>
.proxy-auth-cell__username {
  color: color-mix(in srgb, var(--theme-page-text) 84%, transparent);
}

.proxy-auth-cell__password,
.proxy-auth-cell__empty {
  color: var(--theme-page-muted);
}

.proxy-auth-cell__toggle {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
  transition: color 0.2s ease;
  margin-left: var(--theme-endpoint-popover-icon-button-padding);
  border-radius: var(--theme-settings-inline-button-radius);
  padding: var(--theme-endpoint-popover-icon-button-padding);
}

.proxy-auth-cell__toggle:hover {
  color: var(--theme-page-text);
}
</style>
