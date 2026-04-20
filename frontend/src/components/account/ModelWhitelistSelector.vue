<template>
  <div>
    <div class="relative mb-3">
      <div
        @click="toggleDropdown"
        class="model-whitelist-selector__trigger cursor-pointer"
      >
        <div class="grid grid-cols-2 gap-1.5">
          <span
            v-for="model in modelValue"
            :key="model"
            class="model-whitelist-selector__chip inline-flex items-center justify-between gap-1 text-xs"
          >
            <span class="flex items-center gap-1 truncate">
              <ModelIcon :model="model" size="14px" />
              <span class="truncate">{{ model }}</span>
            </span>
            <button
              type="button"
              @click.stop="removeModel(model)"
              class="model-whitelist-selector__chip-remove shrink-0"
            >
              <Icon name="x" size="xs" class="h-3.5 w-3.5" :stroke-width="2" />
            </button>
          </span>
        </div>
        <div class="model-whitelist-selector__summary flex items-center justify-between">
          <span class="model-whitelist-selector__summary-text text-xs">{{ t('admin.accounts.modelCount', { count: modelValue.length }) }}</span>
          <svg class="model-whitelist-selector__summary-icon h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </div>
      </div>
      <div
        v-if="showDropdown"
        class="model-whitelist-selector__dropdown absolute left-0 right-0 top-full z-50"
      >
        <div class="model-whitelist-selector__dropdown-header sticky top-0">
          <input
            v-model="searchQuery"
            type="text"
            class="input w-full text-sm"
            :placeholder="t('admin.accounts.searchModels')"
            @click.stop
          />
        </div>
        <div class="model-whitelist-selector__options-list overflow-auto">
          <button
            v-for="model in filteredModels"
            :key="model.value"
            type="button"
            @click="toggleModel(model.value)"
            class="model-whitelist-selector__option flex w-full items-center gap-2 text-left text-sm"
          >
            <span :class="getOptionCheckboxClass(modelValue.includes(model.value))">
              <svg v-if="modelValue.includes(model.value)" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
              </svg>
            </span>
            <ModelIcon :model="model.value" size="18px" />
            <span class="model-whitelist-selector__option-label truncate">{{ model.value }}</span>
          </button>
          <div v-if="filteredModels.length === 0" class="model-whitelist-selector__empty text-center text-sm">
            {{ t('admin.accounts.noMatchingModels') }}
          </div>
        </div>
      </div>
    </div>

    <div class="mb-4 flex flex-wrap gap-2">
      <button
        type="button"
        @click="fillRelated"
        class="model-whitelist-selector__action model-whitelist-selector__action--info text-sm"
      >
        {{ t('admin.accounts.fillRelatedModels') }}
      </button>
      <button
        type="button"
        @click="clearAll"
        class="model-whitelist-selector__action model-whitelist-selector__action--danger text-sm"
      >
        {{ t('admin.accounts.clearAllModels') }}
      </button>
    </div>

    <div class="mb-3">
      <label class="model-whitelist-selector__label mb-1.5 block text-sm font-medium">{{ t('admin.accounts.customModelName') }}</label>
      <div class="flex gap-2">
        <input
          v-model="customModel"
          type="text"
          class="input flex-1"
          :placeholder="t('admin.accounts.enterCustomModelName')"
          @keydown.enter.prevent="handleEnter"
          @compositionstart="isComposing = true"
          @compositionend="isComposing = false"
        />
        <button
          type="button"
          @click="addCustom"
          class="model-whitelist-selector__add-button text-sm font-medium"
        >
          {{ t('admin.accounts.addModel') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import ModelIcon from '@/components/common/ModelIcon.vue'
import Icon from '@/components/icons/Icon.vue'
import {
  ensureModelCatalogLoaded,
  getAllModelOptions,
  getModelsByPlatform
} from '@/composables/useModelWhitelist'

const { t } = useI18n()

const props = defineProps<{
  modelValue: string[]
  platform?: string
  platforms?: string[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
}>()

const appStore = useAppStore()

const showDropdown = ref(false)
const searchQuery = ref('')
const customModel = ref('')
const isComposing = ref(false)
const normalizedPlatforms = computed(() => {
  const rawPlatforms =
    props.platforms && props.platforms.length > 0
      ? props.platforms
      : props.platform
        ? [props.platform]
        : []

  return Array.from(
    new Set(
      rawPlatforms
        .map(platform => platform?.trim())
        .filter((platform): platform is string => Boolean(platform))
    )
  )
})

watch(
  normalizedPlatforms,
  (platforms) => {
    const catalogPlatforms = platforms.length > 0 ? platforms : ['grok']
    void Promise.all(catalogPlatforms.map((platform) => ensureModelCatalogLoaded(platform)))
  },
  { immediate: true }
)

const availableOptions = computed(() => {
  if (normalizedPlatforms.value.length === 0) {
    return getAllModelOptions()
  }

  const allowedModels = new Set<string>()
  for (const platform of normalizedPlatforms.value) {
    for (const model of getModelsByPlatform(platform)) {
      allowedModels.add(model)
    }
  }

  return getAllModelOptions().filter(model => allowedModels.has(model.value))
})

const filteredModels = computed(() => {
  const query = searchQuery.value.toLowerCase().trim()
  if (!query) return availableOptions.value
  return availableOptions.value.filter(
    m => m.value.toLowerCase().includes(query) || m.label.toLowerCase().includes(query)
  )
})

const toggleDropdown = () => {
  showDropdown.value = !showDropdown.value
  if (!showDropdown.value) searchQuery.value = ''
}

const getOptionCheckboxClass = (isSelected: boolean) => [
  'model-whitelist-selector__checkbox',
  'flex h-4 w-4 shrink-0 items-center justify-center',
  isSelected
    ? 'model-whitelist-selector__checkbox--selected'
    : 'model-whitelist-selector__checkbox--idle'
]

const removeModel = (model: string) => {
  emit('update:modelValue', props.modelValue.filter(m => m !== model))
}

const toggleModel = (model: string) => {
  if (props.modelValue.includes(model)) {
    removeModel(model)
  } else {
    emit('update:modelValue', [...props.modelValue, model])
  }
}

const addCustom = () => {
  const model = customModel.value.trim()
  if (!model) return
  if (props.modelValue.includes(model)) {
    appStore.showInfo(t('admin.accounts.modelExists'))
    return
  }
  emit('update:modelValue', [...props.modelValue, model])
  customModel.value = ''
}

const handleEnter = () => {
  if (!isComposing.value) addCustom()
}

const fillRelated = async () => {
  await Promise.all(
    normalizedPlatforms.value.map((platform) => ensureModelCatalogLoaded(platform))
  )
  const newModels = [...props.modelValue]
  for (const platform of normalizedPlatforms.value) {
    for (const model of getModelsByPlatform(platform)) {
      if (!newModels.includes(model)) {
        newModels.push(model)
      }
    }
  }
  emit('update:modelValue', newModels)
}

const clearAll = () => {
  emit('update:modelValue', [])
}

</script>

<style scoped>
.model-whitelist-selector__trigger,
.model-whitelist-selector__dropdown,
.model-whitelist-selector__dropdown-header {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 74%, transparent);
  background: var(--theme-surface);
}

.model-whitelist-selector__trigger {
  border-radius: var(--theme-model-whitelist-radius);
  padding: var(--theme-model-whitelist-trigger-padding-y)
    var(--theme-model-whitelist-trigger-padding-x);
}

.model-whitelist-selector__chip {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-text);
  border-radius: var(--theme-model-whitelist-chip-radius);
  padding: var(--theme-model-whitelist-chip-padding-y)
    var(--theme-model-whitelist-chip-padding-x);
}

.model-whitelist-selector__chip-remove {
  border-radius: 999px;
  transition:
    background-color 0.2s ease,
    color 0.2s ease;
}

.model-whitelist-selector__chip-remove:hover {
  background: color-mix(in srgb, var(--theme-card-border) 58%, transparent);
}

.model-whitelist-selector__summary {
  margin-top: var(--theme-model-whitelist-summary-margin-top);
  padding-top: var(--theme-model-whitelist-summary-padding-top);
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.model-whitelist-selector__summary-text,
.model-whitelist-selector__summary-icon,
.model-whitelist-selector__empty {
  color: var(--theme-page-muted);
}

.model-whitelist-selector__dropdown {
  margin-top: var(--theme-model-whitelist-dropdown-offset);
  box-shadow: var(--theme-dropdown-shadow);
  border-radius: var(--theme-model-whitelist-dropdown-radius);
}

.model-whitelist-selector__dropdown-header {
  border-right: none;
  border-left: none;
  border-top: none;
  padding: var(--theme-model-whitelist-dropdown-header-padding);
}

.model-whitelist-selector__option {
  padding: var(--theme-model-whitelist-option-padding-y)
    var(--theme-model-whitelist-option-padding-x);
}

.model-whitelist-selector__options-list {
  max-height: var(--theme-model-whitelist-dropdown-max-height);
}

.model-whitelist-selector__option {
  transition: background-color 0.2s ease;
}

.model-whitelist-selector__option:hover {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.model-whitelist-selector__option-label,
.model-whitelist-selector__label {
  color: var(--theme-page-text);
}

.model-whitelist-selector__empty {
  padding: var(--theme-model-whitelist-empty-padding-y)
    var(--theme-model-whitelist-empty-padding-x);
}

.model-whitelist-selector__checkbox {
  border: 1px solid currentColor;
  border-radius: var(--theme-model-whitelist-chip-radius);
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease,
    color 0.2s ease;
}

.model-whitelist-selector__checkbox--selected {
  border-color: var(--theme-accent);
  background: var(--theme-accent);
  color: var(--theme-filled-text);
}

.model-whitelist-selector__checkbox--idle {
  border-color: var(--theme-input-border);
  color: transparent;
}

.model-whitelist-selector__action,
.model-whitelist-selector__add-button {
  border-radius: var(--theme-model-whitelist-radius);
  border: 1px solid transparent;
  padding: var(--theme-model-whitelist-action-padding-y)
    var(--theme-model-whitelist-action-padding-x);
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease,
    color 0.2s ease;
}

.model-whitelist-selector__add-button {
  padding: var(--theme-model-whitelist-add-padding-y)
    var(--theme-model-whitelist-add-padding-x);
}

.model-whitelist-selector__action--info {
  border-color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 24%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 8%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.model-whitelist-selector__action--info:hover {
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 14%, var(--theme-surface));
}

.model-whitelist-selector__action--danger {
  border-color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 24%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 8%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.model-whitelist-selector__action--danger:hover {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 14%, var(--theme-surface));
}

.model-whitelist-selector__add-button {
  background: color-mix(in srgb, var(--theme-accent-soft) 86%, var(--theme-surface));
  color: var(--theme-accent);
}

.model-whitelist-selector__add-button:hover {
  background: color-mix(in srgb, var(--theme-accent-soft) 96%, var(--theme-surface));
}
</style>
