<template>
  <div class="space-y-6">
    <div class="mb-6 text-center">
      <h2 class="text-xl font-semibold text-gray-900 dark:text-white">
        {{ t('setup.ready.title') }}
      </h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
        {{ t('setup.ready.description') }}
      </p>
    </div>

    <div class="space-y-4">
      <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-700">
        <h3 class="mb-2 text-sm font-medium text-gray-500 dark:text-dark-400">
          {{ t('setup.ready.database') }}
        </h3>
        <p class="text-gray-900 dark:text-white">
          {{ database.user }}@{{ database.host }}:{{ database.port }}/{{ database.dbname }}
        </p>
      </div>

      <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-700">
        <h3 class="mb-2 text-sm font-medium text-gray-500 dark:text-dark-400">
          {{ t('setup.ready.redis') }}
        </h3>
        <p class="text-gray-900 dark:text-white">
          {{ redis.host }}:{{ redis.port }}
        </p>
      </div>

      <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-700">
        <h3 class="mb-2 text-sm font-medium text-gray-500 dark:text-dark-400">
          {{ t('setup.ready.adminEmail') }}
        </h3>
        <p class="text-gray-900 dark:text-white">{{ adminEmail }}</p>
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
