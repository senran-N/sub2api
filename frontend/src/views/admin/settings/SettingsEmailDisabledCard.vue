<template>
  <div class="card">
    <div class="p-6">
      <div class="flex items-start gap-3">
        <Icon name="mail" size="md" class="mt-0.5 flex-shrink-0 text-gray-400 dark:text-gray-500" />
        <div>
          <h3 class="font-medium text-gray-900 dark:text-white">
            {{ t('admin.settings.emailTabDisabledTitle') }}
          </h3>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.settings.emailTabDisabledHint') }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
</script>
