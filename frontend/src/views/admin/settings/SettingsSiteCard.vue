<template>
  <div class="card">
    <div class="settings-site-card__header settings-site-card__header-spacing">
      <h2 class="settings-site-card__title text-lg font-semibold">
        {{ t('admin.settings.site.title') }}
      </h2>
      <p class="settings-site-card__description mt-1 text-sm">
        {{ t('admin.settings.site.description') }}
      </p>
    </div>
    <div class="settings-site-card__body space-y-6">
      <div
        class="settings-site-card__mode-banner settings-site-card__panel flex items-center justify-between"
      >
        <div>
          <h3 class="settings-site-card__title text-sm font-medium">
            {{ t('admin.settings.site.backendMode') }}
          </h3>
          <p class="settings-site-card__description mt-1 text-xs">
            {{ t('admin.settings.site.backendModeDescription') }}
          </p>
        </div>
        <Toggle v-model="form.backend_mode_enabled" />
      </div>

      <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
        <div>
          <label class="settings-site-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.site.siteName') }}
          </label>
          <input
            v-model="form.site_name"
            type="text"
            class="input"
            :placeholder="t('admin.settings.site.siteNamePlaceholder')"
          />
          <p class="settings-site-card__description mt-1.5 text-xs">
            {{ t('admin.settings.site.siteNameHint') }}
          </p>
        </div>
        <div>
          <label class="settings-site-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.site.siteSubtitle') }}
          </label>
          <input
            v-model="form.site_subtitle"
            type="text"
            class="input"
            :placeholder="t('admin.settings.site.siteSubtitlePlaceholder')"
          />
          <p class="settings-site-card__description mt-1.5 text-xs">
            {{ t('admin.settings.site.siteSubtitleHint') }}
          </p>
        </div>
      </div>

      <div>
        <label class="settings-site-card__field-label mb-2 block text-sm font-medium">
          Frontend Theme
        </label>
        <Select
          v-model="form.frontend_theme"
          :options="themeOptions"
          placeholder="Select a frontend theme"
        />
        <p class="settings-site-card__description mt-1.5 text-xs">
          Applied globally after save. New themes can be added through the theme registry without rewriting business pages.
        </p>
      </div>

      <div>
        <label class="settings-site-card__field-label mb-2 block text-sm font-medium">
          {{ t('admin.settings.site.apiBaseUrl') }}
        </label>
        <input
          v-model="form.api_base_url"
          type="text"
          class="input font-mono text-sm"
          :placeholder="t('admin.settings.site.apiBaseUrlPlaceholder')"
        />
        <p class="settings-site-card__description mt-1.5 text-xs">
          {{ t('admin.settings.site.apiBaseUrlHint') }}
        </p>
      </div>

      <div>
        <label class="settings-site-card__field-label mb-2 block text-sm font-medium">
          {{ t('admin.settings.site.customEndpoints.title') }}
        </label>
        <p class="settings-site-card__description mb-3 text-xs">
          {{ t('admin.settings.site.customEndpoints.description') }}
        </p>

        <div class="space-y-3">
          <div
            v-for="(endpoint, index) in form.custom_endpoints"
            :key="index"
            class="settings-site-card__endpoint-card settings-site-card__panel"
          >
            <div class="mb-3 flex items-center justify-between">
              <span class="settings-site-card__field-label text-sm font-medium">
                {{ t('admin.settings.site.customEndpoints.itemLabel', { n: index + 1 }) }}
              </span>
              <button
                type="button"
                class="settings-site-card__remove-button settings-site-card__remove-button-layout"
                @click="$emit('remove-endpoint', index)"
              >
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <div>
                <label class="settings-site-card__mini-label mb-1 block text-xs font-medium">
                  {{ t('admin.settings.site.customEndpoints.name') }}
                </label>
                <input
                  v-model="endpoint.name"
                  type="text"
                  class="input text-sm"
                  :placeholder="t('admin.settings.site.customEndpoints.namePlaceholder')"
                />
              </div>
              <div>
                <label class="settings-site-card__mini-label mb-1 block text-xs font-medium">
                  {{ t('admin.settings.site.customEndpoints.endpointUrl') }}
                </label>
                <input
                  v-model="endpoint.endpoint"
                  type="url"
                  class="input font-mono text-sm"
                  :placeholder="t('admin.settings.site.customEndpoints.endpointUrlPlaceholder')"
                />
              </div>
              <div class="sm:col-span-2">
                <label class="settings-site-card__mini-label mb-1 block text-xs font-medium">
                  {{ t('admin.settings.site.customEndpoints.descriptionLabel') }}
                </label>
                <input
                  v-model="endpoint.description"
                  type="text"
                  class="input text-sm"
                  :placeholder="t('admin.settings.site.customEndpoints.descriptionPlaceholder')"
                />
              </div>
            </div>
          </div>
        </div>

        <button
          type="button"
          class="settings-site-card__add-button settings-site-card__add-button-layout mt-3 flex w-full items-center justify-center gap-2 text-sm transition-colors"
          @click="$emit('add-endpoint')"
        >
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
          </svg>
          {{ t('admin.settings.site.customEndpoints.add') }}
        </button>
      </div>

      <div>
        <label class="settings-site-card__field-label mb-2 block text-sm font-medium">
          {{ t('admin.settings.site.contactInfo') }}
        </label>
        <input
          v-model="form.contact_info"
          type="text"
          class="input"
          :placeholder="t('admin.settings.site.contactInfoPlaceholder')"
        />
        <p class="settings-site-card__description mt-1.5 text-xs">
          {{ t('admin.settings.site.contactInfoHint') }}
        </p>
      </div>

      <div>
        <label class="settings-site-card__field-label mb-2 block text-sm font-medium">
          {{ t('admin.settings.site.docUrl') }}
        </label>
        <input
          v-model="form.doc_url"
          type="url"
          class="input font-mono text-sm"
          :placeholder="t('admin.settings.site.docUrlPlaceholder')"
        />
        <p class="settings-site-card__description mt-1.5 text-xs">
          {{ t('admin.settings.site.docUrlHint') }}
        </p>
      </div>

      <div>
        <label class="settings-site-card__field-label mb-2 block text-sm font-medium">
          {{ t('admin.settings.site.siteLogo') }}
        </label>
        <ImageUpload
          v-model="form.site_logo"
          mode="image"
          :upload-label="t('admin.settings.site.uploadImage')"
          :remove-label="t('admin.settings.site.remove')"
          :hint="t('admin.settings.site.logoHint')"
          :max-size="300 * 1024"
        />
      </div>

      <div>
        <label class="settings-site-card__field-label mb-2 block text-sm font-medium">
          {{ t('admin.settings.site.homeContent') }}
        </label>
        <textarea
          v-model="form.home_content"
          rows="6"
          class="input font-mono text-sm"
          :placeholder="t('admin.settings.site.homeContentPlaceholder')"
        ></textarea>
        <p class="settings-site-card__description mt-1.5 text-xs">
          {{ t('admin.settings.site.homeContentHint') }}
        </p>
        <p class="settings-site-card__warning mt-2 text-xs">
          {{ t('admin.settings.site.homeContentIframeWarning') }}
        </p>
      </div>

      <div
        class="settings-site-card__section flex items-center justify-between pt-4"
      >
        <div>
          <label class="settings-site-card__label font-medium">
            {{ t('admin.settings.site.hideCcsImportButton') }}
          </label>
          <p class="settings-site-card__description text-sm">
            {{ t('admin.settings.site.hideCcsImportButtonHint') }}
          </p>
        </div>
        <Toggle v-model="form.hide_ccs_import_button" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import ImageUpload from '@/components/common/ImageUpload.vue'
import Select from '@/components/common/Select.vue'
import type { SettingsForm } from '../settingsForm'
import { FRONTEND_THEMES } from '@/themes'

defineProps<{
  form: SettingsForm
}>()

defineEmits<{
  'add-endpoint': []
  'remove-endpoint': [index: number]
}>()

const { t } = useI18n()

const themeOptions = FRONTEND_THEMES.map((theme) => ({
  value: theme.id,
  label: `${theme.label} · ${theme.description}`
}))
</script>

<style scoped>
.settings-site-card__header,
.settings-site-card__section {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-site-card__header {
  border-top: none;
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-site-card__title,
.settings-site-card__label,
.settings-site-card__field-label,
.settings-site-card__mini-label {
  color: var(--theme-page-text);
}

.settings-site-card__description {
  color: var(--theme-page-muted);
}

.settings-site-card__mode-banner {
  border: 1px solid color-mix(in srgb, rgb(var(--theme-warning-rgb)) 20%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 8%, var(--theme-surface));
}

.settings-site-card__header-spacing {
  padding: var(--theme-settings-card-header-padding-y) var(--theme-settings-card-header-padding-x);
}

.settings-site-card__body {
  padding: var(--theme-settings-card-body-padding);
}

.settings-site-card__panel {
  border-radius: var(--theme-settings-card-panel-radius);
  padding: var(--theme-settings-card-panel-padding);
}

.settings-site-card__endpoint-card {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 62%, var(--theme-surface));
}

.settings-site-card__remove-button {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 60%, var(--theme-page-muted));
  transition: background-color 0.2s ease, color 0.2s ease;
}

.settings-site-card__remove-button-layout {
  border-radius: var(--theme-settings-inline-button-radius);
  padding: var(--theme-settings-inline-button-padding);
}

.settings-site-card__remove-button:hover {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.settings-site-card__add-button {
  border: 2px dashed color-mix(in srgb, var(--theme-card-border) 84%, transparent);
  color: var(--theme-page-muted);
}

.settings-site-card__add-button-layout {
  border-radius: var(--theme-settings-card-panel-radius);
  padding: var(--theme-settings-action-padding-y) var(--theme-settings-action-padding-x);
}

.settings-site-card__add-button:hover {
  border-color: color-mix(in srgb, var(--theme-accent) 44%, var(--theme-card-border));
  color: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-page-text));
  background: color-mix(in srgb, var(--theme-accent-soft) 46%, transparent);
}

.settings-site-card__warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}
</style>
