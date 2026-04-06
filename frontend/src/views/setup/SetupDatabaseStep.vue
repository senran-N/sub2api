<template>
  <div class="space-y-6">
    <div class="setup-step-header">
      <h2 class="setup-step-title">
        {{ t('setup.database.title') }}
      </h2>
      <p class="setup-step-description">
        {{ t('setup.database.description') }}
      </p>
    </div>

    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4">
      <div>
        <label class="input-label">{{ t('setup.database.host') }}</label>
        <input
          v-model="host"
          type="text"
          class="input"
          placeholder="localhost"
        />
      </div>
      <div>
        <label class="input-label">{{ t('setup.database.port') }}</label>
        <input
          v-model.number="port"
          type="number"
          class="input"
          placeholder="5432"
        />
      </div>
    </div>

    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4">
      <div>
        <label class="input-label">{{ t('setup.database.username') }}</label>
        <input
          v-model="user"
          type="text"
          class="input"
          placeholder="postgres"
        />
      </div>
      <div>
        <label class="input-label">{{ t('setup.database.password') }}</label>
        <input
          v-model="password"
          type="password"
          class="input"
          :placeholder="t('setup.database.passwordPlaceholder')"
        />
      </div>
    </div>

    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4">
      <div>
        <label class="input-label">{{ t('setup.database.databaseName') }}</label>
        <input
          v-model="databaseName"
          type="text"
          class="input"
          placeholder="sub2api"
        />
      </div>
      <div>
        <label class="input-label">{{ t('setup.database.sslMode') }}</label>
        <Select
          v-model="sslMode"
          :options="sslModeOptions"
        />
      </div>
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
import type { DatabaseConfig } from '@/api/setup'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  connected: boolean
  database: DatabaseConfig
  testing: boolean
}>()

const emit = defineEmits<{
  'test-connection': []
  'update:database': [value: DatabaseConfig]
}>()

const { t } = useI18n()

const updateDatabase = (patch: Partial<DatabaseConfig>) => {
  emit('update:database', {
    ...props.database,
    ...patch
  })
}

const host = computed({
  get: () => props.database.host,
  set: (value: string) => updateDatabase({ host: value })
})

const port = computed({
  get: () => props.database.port,
  set: (value: number) => updateDatabase({ port: value })
})

const user = computed({
  get: () => props.database.user,
  set: (value: string) => updateDatabase({ user: value })
})

const password = computed({
  get: () => props.database.password,
  set: (value: string) => updateDatabase({ password: value })
})

const databaseName = computed({
  get: () => props.database.dbname,
  set: (value: string) => updateDatabase({ dbname: value })
})

const sslMode = computed({
  get: () => props.database.sslmode,
  set: (value: string | number | boolean | null) => updateDatabase({ sslmode: String(value ?? '') })
})

const sslModeOptions = computed(() => [
  { value: 'disable', label: t('setup.database.ssl.disable') },
  { value: 'require', label: t('setup.database.ssl.require') },
  { value: 'verify-ca', label: t('setup.database.ssl.verifyCa') },
  { value: 'verify-full', label: t('setup.database.ssl.verifyFull') }
])
</script>
