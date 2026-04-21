<template>
  <div class="card">
    <div class="card-header">
      <h2 class="settings-model-routing-card__title text-lg font-semibold">
        {{ t('admin.settings.modelRouting.title') }}
      </h2>
      <p class="settings-model-routing-card__description mt-1 text-sm">
        {{ t('admin.settings.modelRouting.description') }}
      </p>
    </div>

    <div class="settings-model-routing-card__body">
      <div class="flex items-center justify-between gap-4">
        <div>
          <label class="settings-model-routing-card__field-label text-sm font-medium">
            {{ t('admin.settings.modelRouting.enableFallback') }}
          </label>
          <p class="settings-model-routing-card__description mt-0.5 text-xs">
            {{ t('admin.settings.modelRouting.enableFallbackHint') }}
          </p>
        </div>
        <Toggle v-model="form.enable_model_fallback" />
      </div>

      <div class="grid grid-cols-1 gap-4 xl:grid-cols-2">
        <div
          v-for="platform in fallbackPlatforms"
          :key="platform.key"
        >
          <label class="settings-model-routing-card__field-label mb-2 block text-sm font-medium">
            {{ t(platform.labelKey) }}
          </label>
          <Select
            v-model="form[platform.modelKey]"
            :options="platform.options"
            :disabled="!form.enable_model_fallback"
            searchable
          />
          <p class="settings-model-routing-card__description mt-1.5 text-xs">
            {{ t(platform.hintKey) }}
          </p>
        </div>
      </div>

      <div class="settings-model-routing-card__section space-y-4">
        <div class="settings-model-routing-card__behavior-panel">
          <div>
            <h3 class="settings-model-routing-card__field-label text-sm font-medium">
              {{ t('admin.settings.modelRouting.grokMediaBehaviorTitle') }}
            </h3>
            <p class="settings-model-routing-card__description mt-1 text-xs">
              {{ t('admin.settings.modelRouting.grokMediaBehaviorDescription') }}
            </p>
          </div>

          <div class="grid grid-cols-1 gap-3 xl:grid-cols-2">
            <div class="settings-model-routing-card__behavior-item">
              <p class="settings-model-routing-card__field-label text-sm font-medium">
                {{ t('admin.settings.modelRouting.grokMediaImageBehavior') }}
              </p>
              <p class="settings-model-routing-card__description mt-1 text-sm">
                {{ imageDeliveryBehavior }}
              </p>
            </div>

            <div class="settings-model-routing-card__behavior-item">
              <p class="settings-model-routing-card__field-label text-sm font-medium">
                {{ t('admin.settings.modelRouting.grokMediaVideoBehavior') }}
              </p>
              <p class="settings-model-routing-card__description mt-1 text-sm">
                {{ videoDeliveryBehavior }}
              </p>
            </div>
          </div>

          <p class="settings-model-routing-card__description text-xs">
            {{ cacheRetentionBehavior }}
          </p>
        </div>

        <div class="grid grid-cols-1 gap-4 xl:grid-cols-2">
        <div>
          <label class="settings-model-routing-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.modelRouting.grokOfficialBaseUrl') }}
          </label>
          <input
            v-model.trim="form.grok_official_base_url"
            type="url"
            class="input"
            placeholder="https://api.x.ai"
          />
          <p class="settings-model-routing-card__description mt-1.5 text-xs">
            {{ t('admin.settings.modelRouting.grokOfficialBaseUrlHint') }}
          </p>
        </div>

        <div>
          <label class="settings-model-routing-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.modelRouting.grokSessionBaseUrl') }}
          </label>
          <input
            v-model.trim="form.grok_session_base_url"
            type="url"
            class="input"
            placeholder="https://grok.com"
          />
          <p class="settings-model-routing-card__description mt-1.5 text-xs">
            {{ t('admin.settings.modelRouting.grokSessionBaseUrlHint') }}
          </p>
        </div>

        <div class="flex items-center justify-between gap-4 rounded-lg border border-[color:var(--theme-card-border)] p-4">
          <div>
            <label class="settings-model-routing-card__field-label text-sm font-medium">
              {{ t('admin.settings.modelRouting.grokThinkingSummary') }}
            </label>
            <p class="settings-model-routing-card__description mt-0.5 text-xs">
              {{ t('admin.settings.modelRouting.grokThinkingSummaryHint') }}
            </p>
          </div>
          <Toggle v-model="form.grok_thinking_summary" />
        </div>

        <div class="flex items-center justify-between gap-4 rounded-lg border border-[color:var(--theme-card-border)] p-4">
          <div>
            <label class="settings-model-routing-card__field-label text-sm font-medium">
              {{ t('admin.settings.modelRouting.grokShowSearchSources') }}
            </label>
            <p class="settings-model-routing-card__description mt-0.5 text-xs">
              {{ t('admin.settings.modelRouting.grokShowSearchSourcesHint') }}
            </p>
          </div>
          <Toggle v-model="form.grok_show_search_sources" />
        </div>

        <div>
          <label class="settings-model-routing-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.modelRouting.grokImageOutputFormat') }}
          </label>
          <Select
            v-model="form.grok_image_output_format"
            :options="imageOutputOptions"
          />
          <p class="settings-model-routing-card__description mt-1.5 text-xs">
            {{ t('admin.settings.modelRouting.grokImageOutputFormatHint') }}
          </p>
        </div>

        <div>
          <label class="settings-model-routing-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.modelRouting.grokVideoOutputFormat') }}
          </label>
          <Select
            v-model="form.grok_video_output_format"
            :options="videoOutputOptions"
          />
          <p class="settings-model-routing-card__description mt-1.5 text-xs">
            {{ t('admin.settings.modelRouting.grokVideoOutputFormatHint') }}
          </p>
        </div>

        <div>
          <label class="settings-model-routing-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.modelRouting.grokMediaCacheRetentionHours') }}
          </label>
          <input
            v-model.number="form.grok_media_cache_retention_hours"
            type="number"
            min="1"
            class="input"
            placeholder="72"
          />
          <p class="settings-model-routing-card__description mt-1.5 text-xs">
            {{ t('admin.settings.modelRouting.grokMediaCacheRetentionHoursHint') }}
          </p>
        </div>

        <div>
          <label class="settings-model-routing-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.modelRouting.grokQuotaSyncIntervalSeconds') }}
          </label>
          <input
            v-model.number="form.grok_quota_sync_interval_seconds"
            type="number"
            min="60"
            class="input"
            placeholder="900"
          />
          <p class="settings-model-routing-card__description mt-1.5 text-xs">
            {{ t('admin.settings.modelRouting.grokQuotaSyncIntervalSecondsHint') }}
          </p>
        </div>

        <div>
          <label class="settings-model-routing-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.modelRouting.grokCapabilityProbeIntervalSeconds') }}
          </label>
          <input
            v-model.number="form.grok_capability_probe_interval_seconds"
            type="number"
            min="60"
            class="input"
            placeholder="21600"
          />
          <p class="settings-model-routing-card__description mt-1.5 text-xs">
            {{ t('admin.settings.modelRouting.grokCapabilityProbeIntervalSecondsHint') }}
          </p>
        </div>

        <div>
          <label class="settings-model-routing-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.modelRouting.grokUsageSyncConcurrency') }}
          </label>
          <input
            v-model.number="form.grok_usage_sync_concurrency"
            type="number"
            min="1"
            class="input"
            placeholder="50"
          />
          <p class="settings-model-routing-card__description mt-1.5 text-xs">
            {{ t('admin.settings.modelRouting.grokUsageSyncConcurrencyHint') }}
          </p>
        </div>

        <div>
          <label class="settings-model-routing-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.modelRouting.grokCapabilityProbeConcurrency') }}
          </label>
          <input
            v-model.number="form.grok_capability_probe_concurrency"
            type="number"
            min="1"
            class="input"
            placeholder="10"
          />
          <p class="settings-model-routing-card__description mt-1.5 text-xs">
            {{ t('admin.settings.modelRouting.grokCapabilityProbeConcurrencyHint') }}
          </p>
        </div>

        <div>
          <label class="settings-model-routing-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.modelRouting.grokSessionValidityCheckInterval') }}
          </label>
          <input
            v-model.number="form.grok_session_validity_check_interval"
            type="number"
            min="60"
            class="input"
            placeholder="1800"
          />
          <p class="settings-model-routing-card__description mt-1.5 text-xs">
            {{ t('admin.settings.modelRouting.grokSessionValidityCheckIntervalHint') }}
          </p>
        </div>

        <div>
          <label class="settings-model-routing-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.modelRouting.grokVideoTimeout') }}
          </label>
          <input
            v-model.number="form.grok_video_timeout"
            type="number"
            min="30"
            class="input"
            placeholder="600"
          />
          <p class="settings-model-routing-card__description mt-1.5 text-xs">
            {{ t('admin.settings.modelRouting.grokVideoTimeoutHint') }}
          </p>
        </div>

        <div class="flex items-center justify-between gap-4 rounded-lg border border-[color:var(--theme-card-border)] p-4">
          <div>
            <label class="settings-model-routing-card__field-label text-sm font-medium">
              {{ t('admin.settings.modelRouting.grokMediaProxyEnabled') }}
            </label>
            <p class="settings-model-routing-card__description mt-0.5 text-xs">
              {{ t('admin.settings.modelRouting.grokMediaProxyEnabledHint') }}
            </p>
          </div>
          <Toggle v-model="form.grok_media_proxy_enabled" />
        </div>
      </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, toRefs } from 'vue'
import { useI18n } from 'vue-i18n'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import Toggle from '@/components/common/Toggle.vue'
import {
  ensureModelCatalogLoaded,
  getModelOptionsByPlatform
} from '@/composables/useModelWhitelist'
import type { SettingsModelRoutingFields } from './settingsForm'

const props = defineProps<{
  form: SettingsModelRoutingFields
}>()
const { form } = toRefs(props)

const { t } = useI18n()

const fallbackPlatforms = computed(() => [
  {
    key: 'anthropic',
    modelKey: 'fallback_model_anthropic' as const,
    labelKey: 'admin.settings.modelRouting.fallbackAnthropic',
    hintKey: 'admin.settings.modelRouting.fallbackAnthropicHint',
    options: getModelOptionsByPlatform('anthropic')
  },
  {
    key: 'openai',
    modelKey: 'fallback_model_openai' as const,
    labelKey: 'admin.settings.modelRouting.fallbackOpenai',
    hintKey: 'admin.settings.modelRouting.fallbackOpenaiHint',
    options: getModelOptionsByPlatform('openai')
  },
  {
    key: 'grok',
    modelKey: 'fallback_model_grok' as const,
    labelKey: 'admin.settings.modelRouting.fallbackGrok',
    hintKey: 'admin.settings.modelRouting.fallbackGrokHint',
    options: getModelOptionsByPlatform('grok')
  },
  {
    key: 'gemini',
    modelKey: 'fallback_model_gemini' as const,
    labelKey: 'admin.settings.modelRouting.fallbackGemini',
    hintKey: 'admin.settings.modelRouting.fallbackGeminiHint',
    options: getModelOptionsByPlatform('gemini')
  },
  {
    key: 'antigravity',
    modelKey: 'fallback_model_antigravity' as const,
    labelKey: 'admin.settings.modelRouting.fallbackAntigravity',
    hintKey: 'admin.settings.modelRouting.fallbackAntigravityHint',
    options: getModelOptionsByPlatform('antigravity')
  }
])

const imageOutputOptions = computed<SelectOption[]>(() => [
  { value: 'local_url', label: t('admin.settings.modelRouting.outputFormat.localUrl') },
  { value: 'upstream_url', label: t('admin.settings.modelRouting.outputFormat.upstreamUrl') },
  { value: 'markdown', label: t('admin.settings.modelRouting.outputFormat.markdown') },
  { value: 'base64', label: t('admin.settings.modelRouting.outputFormat.base64') }
])

const videoOutputOptions = computed<SelectOption[]>(() => [
  { value: 'local_url', label: t('admin.settings.modelRouting.outputFormat.localUrl') },
  { value: 'upstream_url', label: t('admin.settings.modelRouting.outputFormat.upstreamUrl') },
  { value: 'html', label: t('admin.settings.modelRouting.outputFormat.html') }
])

const imageDeliveryBehavior = computed(() => {
  if (form.value.grok_image_output_format === 'base64') {
    return t('admin.settings.modelRouting.grokMediaBehaviorImageBase64')
  }

  if (form.value.grok_image_output_format === 'upstream_url') {
    return t('admin.settings.modelRouting.grokMediaBehaviorDirect')
  }

  if (!form.value.grok_media_proxy_enabled) {
    if (form.value.grok_image_output_format === 'markdown') {
      return t('admin.settings.modelRouting.grokMediaBehaviorImageMarkdownDirect')
    }
    return t('admin.settings.modelRouting.grokMediaBehaviorDirectFallback')
  }

  if (form.value.grok_image_output_format === 'markdown') {
    return t('admin.settings.modelRouting.grokMediaBehaviorImageMarkdownProxy')
  }

  return t('admin.settings.modelRouting.grokMediaBehaviorProxyLazy')
})

const videoDeliveryBehavior = computed(() => {
  if (form.value.grok_video_output_format === 'upstream_url') {
    return t('admin.settings.modelRouting.grokMediaBehaviorDirect')
  }

  if (!form.value.grok_media_proxy_enabled) {
    if (form.value.grok_video_output_format === 'html') {
      return t('admin.settings.modelRouting.grokMediaBehaviorVideoHTMLDirect')
    }
    return t('admin.settings.modelRouting.grokMediaBehaviorDirectFallback')
  }

  if (form.value.grok_video_output_format === 'html') {
    return t('admin.settings.modelRouting.grokMediaBehaviorVideoHTMLProxy')
  }

  return t('admin.settings.modelRouting.grokMediaBehaviorProxyLazy')
})

const cacheRetentionBehavior = computed(() => {
  const imageUsesLocalCache =
    form.value.grok_image_output_format === 'base64' ||
    (form.value.grok_media_proxy_enabled && form.value.grok_image_output_format !== 'upstream_url')
  const videoUsesLocalCache =
    form.value.grok_media_proxy_enabled && form.value.grok_video_output_format !== 'upstream_url'

  if (imageUsesLocalCache || videoUsesLocalCache) {
    return t('admin.settings.modelRouting.grokMediaBehaviorRetentionActive', {
      hours: form.value.grok_media_cache_retention_hours || 72
    })
  }

  return t('admin.settings.modelRouting.grokMediaBehaviorRetentionInactive')
})

onMounted(() => {
  void ensureModelCatalogLoaded('grok')
})
</script>

<style scoped>
.settings-model-routing-card__title,
.settings-model-routing-card__field-label {
  color: var(--theme-page-text);
}

.settings-model-routing-card__description {
  color: var(--theme-page-muted);
}

.settings-model-routing-card__body {
  padding: var(--theme-settings-card-panel-padding);
  display: flex;
  flex-direction: column;
  gap: var(--theme-settings-card-body-padding);
}

.settings-model-routing-card__section {
  padding-top: 0.5rem;
}

.settings-model-routing-card__behavior-panel {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1rem;
  border-radius: 1rem;
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 88%, transparent);
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 9%, var(--theme-surface));
}

.settings-model-routing-card__behavior-item {
  padding: 0.875rem 1rem;
  border-radius: 0.875rem;
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 82%, transparent);
  background: color-mix(in srgb, var(--theme-surface) 84%, transparent);
}
</style>
