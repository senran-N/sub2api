<template>
  <div class="space-y-6">
    <div class="setup-step-header">
      <h2 class="setup-step-title">
        {{ t('setup.redis.title') }}
      </h2>
      <p class="setup-step-description">
        {{ t('setup.redis.description') }}
      </p>
    </div>

    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4">
      <div>
        <label class="input-label">{{ t('setup.redis.host') }}</label>
        <input
          v-model="host"
          type="text"
          class="input"
          placeholder="localhost"
        />
      </div>
      <div>
        <label class="input-label">{{ t('setup.redis.port') }}</label>
        <input
          v-model.number="port"
          type="number"
          class="input"
          placeholder="6379"
        />
      </div>
    </div>

    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4">
      <div>
        <label class="input-label">{{ t('setup.redis.password') }}</label>
        <input
          v-model="password"
          type="password"
          class="input"
          :placeholder="t('setup.redis.passwordPlaceholder')"
        />
      </div>
      <div>
        <label class="input-label">{{ t('setup.redis.database') }}</label>
        <input
          v-model.number="databaseIndex"
          type="number"
          class="input"
          placeholder="0"
        />
      </div>
    </div>

    <div class="setup-toggle-card flex items-center justify-between">
      <div>
        <p class="setup-toggle-title">
          {{ t('setup.redis.enableTls') }}
        </p>
        <p class="setup-toggle-description">
          {{ t('setup.redis.enableTlsHint') }}
        </p>
      </div>
      <Toggle v-model="enableTls" />
    </div>

    <button
      type="button"
      class="btn btn-secondary w-full"
      :disabled="testing"
      @click="$emit('test-connection')"
    >
      <svg
        v-if="testing"
        class="-ml-1 mr-2 h-4 w-4 animate-spin"
        fill="none"
        viewBox="0 0 24 24"
      >
        <circle
          class="opacity-25"
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          stroke-width="4"
        ></circle>
        <path
          class="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        ></path>
      </svg>
      <Icon
        v-else-if="connected"
        name="check"
        size="md"
        class="mr-2"
        style="color: rgb(var(--theme-success-rgb))"
        :stroke-width="2"
      />
      {{
        testing
          ? t('setup.status.testing')
          : connected
            ? t('setup.status.success')
            : t('setup.status.testConnection')
      }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { RedisConfig } from '@/api/setup'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  connected: boolean
  redis: RedisConfig
  testing: boolean
}>()

const emit = defineEmits<{
  'test-connection': []
  'update:redis': [value: RedisConfig]
}>()

const { t } = useI18n()

const updateRedis = (patch: Partial<RedisConfig>) => {
  emit('update:redis', {
    ...props.redis,
    ...patch
  })
}

const host = computed({
  get: () => props.redis.host,
  set: (value: string) => updateRedis({ host: value })
})

const port = computed({
  get: () => props.redis.port,
  set: (value: number) => updateRedis({ port: value })
})

const password = computed({
  get: () => props.redis.password,
  set: (value: string) => updateRedis({ password: value })
})

const databaseIndex = computed({
  get: () => props.redis.db,
  set: (value: number) => updateRedis({ db: value })
})

const enableTls = computed({
  get: () => props.redis.enable_tls,
  set: (value: boolean) => updateRedis({ enable_tls: value })
})
</script>
