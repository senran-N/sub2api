<template>
  <div class="card">
    <div class="settings-custom-menu-card__header">
      <h2 class="settings-custom-menu-card__title text-lg font-semibold">
        {{ t('admin.settings.customMenu.title') }}
      </h2>
      <p class="settings-custom-menu-card__description mt-1 text-sm">
        {{ t('admin.settings.customMenu.description') }}
      </p>
    </div>
    <div class="settings-custom-menu-card__content space-y-4">
      <div
        v-for="(item, index) in form.custom_menu_items"
        :key="item.id || index"
        class="settings-custom-menu-card__item"
      >
        <div class="mb-3 flex items-center justify-between">
          <span class="settings-custom-menu-card__item-label text-sm font-medium">
            {{ t('admin.settings.customMenu.itemLabel', { n: index + 1 }) }}
          </span>
          <div class="flex items-center gap-2">
            <button
              v-if="index > 0"
              type="button"
              class="settings-custom-menu-card__icon-button"
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
              class="settings-custom-menu-card__icon-button"
              :title="t('admin.settings.customMenu.moveDown')"
              @click="$emit('move-item', index, 1)"
            >
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            <button
              type="button"
              class="settings-custom-menu-card__icon-button settings-custom-menu-card__icon-button--danger"
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
            <label class="settings-custom-menu-card__field-label mb-1 block text-xs font-medium">
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
            <label class="settings-custom-menu-card__field-label mb-1 block text-xs font-medium">
              {{ t('admin.settings.customMenu.visibility') }}
            </label>
            <select v-model="item.visibility" class="input text-sm">
              <option value="user">{{ t('admin.settings.customMenu.visibilityUser') }}</option>
              <option value="admin">{{ t('admin.settings.customMenu.visibilityAdmin') }}</option>
            </select>
          </div>

          <div class="sm:col-span-2">
            <label class="settings-custom-menu-card__field-label mb-1 block text-xs font-medium">
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
            <label class="settings-custom-menu-card__field-label mb-1 block text-xs font-medium">
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
        class="settings-custom-menu-card__add-button flex w-full items-center justify-center gap-2 text-sm transition-colors"
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
import type { SettingsCustomMenuFields } from './settingsForm'

defineProps<{
  form: SettingsCustomMenuFields
}>()

defineEmits<{
  'add-item': []
  'remove-item': [index: number]
  'move-item': [index: number, direction: -1 | 1]
}>()

const { t } = useI18n()
</script>

<style scoped>
.settings-custom-menu-card__header {
  padding: var(--theme-settings-custom-menu-header-padding-y)
    var(--theme-settings-custom-menu-header-padding-x);
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-custom-menu-card__content {
  padding: var(--theme-settings-custom-menu-content-padding);
}

.settings-custom-menu-card__title,
.settings-custom-menu-card__item-label,
.settings-custom-menu-card__field-label {
  color: var(--theme-page-text);
}

.settings-custom-menu-card__description {
  color: var(--theme-page-muted);
}

.settings-custom-menu-card__item {
  border-radius: var(--theme-settings-custom-menu-item-radius);
  padding: var(--theme-settings-custom-menu-item-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
}

.settings-custom-menu-card__icon-button {
  border-radius: var(--theme-settings-custom-menu-icon-button-radius);
  padding: var(--theme-settings-custom-menu-icon-button-padding);
  color: var(--theme-page-muted);
  transition:
    background-color 0.2s ease,
    color 0.2s ease;
}

.settings-custom-menu-card__icon-button:hover {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-text);
}

.settings-custom-menu-card__icon-button--danger:hover {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 9%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.settings-custom-menu-card__add-button {
  border-radius: var(--theme-settings-custom-menu-add-button-radius);
  padding-block: var(--theme-settings-custom-menu-add-button-padding-y);
  border: 2px dashed color-mix(in srgb, var(--theme-card-border) 78%, transparent);
  color: var(--theme-page-muted);
}

.settings-custom-menu-card__add-button:hover {
  border-color: color-mix(in srgb, var(--theme-accent) 46%, var(--theme-card-border));
  color: var(--theme-accent);
}
</style>
