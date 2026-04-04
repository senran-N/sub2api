<template>
  <div v-if="proxy.username || proxy.password" class="flex items-center gap-1.5">
    <div class="flex flex-col text-xs">
      <span v-if="proxy.username" class="text-gray-700 dark:text-gray-200">{{ proxy.username }}</span>
      <span v-if="proxy.password" class="font-mono text-gray-500 dark:text-gray-400">
        {{ passwordVisible ? proxy.password : '••••••' }}
      </span>
    </div>
    <button
      v-if="proxy.password"
      type="button"
      class="ml-1 rounded p-0.5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
      @click.stop="emit('toggle-password', proxy.id)"
    >
      <Icon :name="passwordVisible ? 'eyeOff' : 'eye'" size="sm" />
    </button>
  </div>
  <span v-else class="text-sm text-gray-400">-</span>
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
