<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
        {{ t('admin.settings.customMenu.title') }}
      </h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('admin.settings.customMenu.description') }}
      </p>
    </div>
    <div class="space-y-4 p-6">
      <div
        v-for="(item, index) in form.custom_menu_items"
        :key="item.id || index"
        class="rounded-lg border border-gray-200 p-4 dark:border-dark-600"
      >
        <div class="mb-3 flex items-center justify-between">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t('admin.settings.customMenu.itemLabel', { n: index + 1 }) }}
          </span>
          <div class="flex items-center gap-2">
            <button
              v-if="index > 0"
              type="button"
              class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700"
              :title="t('admin.settings.customMenu.moveUp')"
              @click="$emit('move-item', index, -1)"
            >
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
              </svg>
            </button>
            <button
              v-if="index < form.custom_menu_items.length - 1"
              type="button"
              class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700"
              :title="t('admin.settings.customMenu.moveDown')"
              @click="$emit('move-item', index, 1)"
            >
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            <button
              type="button"
              class="rounded p-1 text-red-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20"
              :title="t('admin.settings.customMenu.remove')"
              @click="$emit('remove-item', index)"
            >
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                />
              </svg>
            </button>
          </div>
        </div>

        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">
              {{ t('admin.settings.customMenu.name') }}
            </label>
            <input
              v-model="item.label"
              type="text"
              class="input text-sm"
              :placeholder="t('admin.settings.customMenu.namePlaceholder')"
            />
          </div>

          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">
              {{ t('admin.settings.customMenu.visibility') }}
            </label>
            <select v-model="item.visibility" class="input text-sm">
              <option value="user">{{ t('admin.settings.customMenu.visibilityUser') }}</option>
              <option value="admin">{{ t('admin.settings.customMenu.visibilityAdmin') }}</option>
            </select>
          </div>

          <div class="sm:col-span-2">
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">
              {{ t('admin.settings.customMenu.url') }}
            </label>
            <input
              v-model="item.url"
              type="url"
              class="input font-mono text-sm"
              :placeholder="t('admin.settings.customMenu.urlPlaceholder')"
            />
          </div>

          <div class="sm:col-span-2">
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">
              {{ t('admin.settings.customMenu.iconSvg') }}
            </label>
            <ImageUpload
              :model-value="item.icon_svg"
              mode="svg"
              size="sm"
              :upload-label="t('admin.settings.customMenu.uploadSvg')"
              :remove-label="t('admin.settings.customMenu.removeSvg')"
              @update:model-value="(value: string) => item.icon_svg = value"
            />
          </div>
        </div>
      </div>

      <button
        type="button"
        class="flex w-full items-center justify-center gap-2 rounded-lg border-2 border-dashed border-gray-300 py-3 text-sm text-gray-500 transition-colors hover:border-primary-400 hover:text-primary-600 dark:border-dark-600 dark:text-gray-400 dark:hover:border-primary-500 dark:hover:text-primary-400"
        @click="$emit('add-item')"
      >
        <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
        </svg>
        {{ t('admin.settings.customMenu.add') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import ImageUpload from '@/components/common/ImageUpload.vue'
import type { SettingsForm } from '../settingsForm'

defineProps<{
  form: SettingsForm
}>()

defineEmits<{
  'add-item': []
  'remove-item': [index: number]
  'move-item': [index: number, direction: -1 | 1]
}>()

const { t } = useI18n()
</script>
