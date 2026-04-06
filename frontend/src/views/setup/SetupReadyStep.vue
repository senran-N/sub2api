<template>
  <div class="space-y-6">
    <div class="setup-step-header">
      <h2 class="setup-step-title">
        {{ t('setup.ready.title') }}
      </h2>
      <p class="setup-step-description">
        {{ t('setup.ready.description') }}
      </p>
    </div>

    <div class="space-y-4">
      <div class="setup-summary-card">
        <h3 class="setup-summary-label">
          {{ t('setup.ready.database') }}
        </h3>
        <p class="setup-summary-value">
          {{ database.user }}@{{ database.host }}:{{ database.port }}/{{ database.dbname }}
        </p>
      </div>

      <div class="setup-summary-card">
        <h3 class="setup-summary-label">
          {{ t('setup.ready.redis') }}
        </h3>
        <p class="setup-summary-value">
          {{ redis.host }}:{{ redis.port }}
        </p>
      </div>

      <div class="setup-summary-card">
        <h3 class="setup-summary-label">
          {{ t('setup.ready.adminEmail') }}
        </h3>
        <p class="setup-summary-value">{{ adminEmail }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { DatabaseConfig, RedisConfig } from '@/api/setup'

defineProps<{
  adminEmail: string
  database: DatabaseConfig
  redis: RedisConfig
}>()

const { t } = useI18n()
</script>
