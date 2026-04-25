<template>
  <div :class="framed && 'model-restriction-section'">
    <label class="input-label">{{ t("admin.accounts.modelRestriction") }}</label>

    <div v-if="disabled" class="model-restriction-section__notice model-restriction-section__notice--amber mb-3">
      <p class="text-xs">
        {{ t(disabledMessageKey || "admin.accounts.openai.modelRestrictionDisabledByPassthrough") }}
      </p>
    </div>

    <template v-else>
      <div class="mb-4 flex gap-2">
        <button
          type="button"
          :class="getModeButtonClasses(mode === 'whitelist', 'accent')"
          @click="emit('update:mode', 'whitelist')"
        >
          <Icon name="checkCircle" size="sm" class="mr-1.5 inline" :stroke-width="2" />
          {{ t("admin.accounts.modelWhitelist") }}
        </button>
        <button
          type="button"
          :class="getModeButtonClasses(mode === 'mapping', 'purple')"
          @click="emit('update:mode', 'mapping')"
        >
          <Icon name="arrowRight" size="sm" class="mr-1.5 inline" :stroke-width="2" />
          {{ t("admin.accounts.modelMapping") }}
        </button>
      </div>

      <div v-if="mode === 'whitelist'">
        <ModelWhitelistSelector
          :model-value="allowedModels"
          :platform="platform"
          :platforms="platforms"
          @update:model-value="emit('update:allowedModels', $event)"
        />
        <p class="model-restriction-section__description text-xs">
          {{
            t("admin.accounts.selectedModels", {
              count: allowedModels.length,
            })
          }}
          <span v-if="allowedModels.length === 0">
            {{ t("admin.accounts.supportsAllModels") }}
          </span>
        </p>
      </div>

      <div v-else>
        <div
          v-if="showMappingNotice"
          class="model-restriction-section__notice model-restriction-section__notice--purple mb-3"
        >
          <p class="text-xs">
            <Icon name="infoCircle" size="sm" class="mr-1 inline" :stroke-width="2" />
            {{ t("admin.accounts.mapRequestModels") }}
          </p>
        </div>

        <div v-if="mappings.length > 0" class="mb-3 space-y-2">
          <div
            v-for="(mapping, index) in mappings"
            :key="mappingKey(mapping)"
            class="flex items-center gap-2"
          >
            <input
              :value="mapping.from"
              type="text"
              class="input flex-1"
              :placeholder="fromPlaceholder"
              @input="updateMapping(index, 'from', $event)"
            />
            <Icon
              name="arrowRight"
              size="sm"
              class="model-restriction-section__arrow flex-shrink-0"
              :stroke-width="2"
            />
            <input
              :value="mapping.to"
              type="text"
              class="input flex-1"
              :placeholder="toPlaceholder"
              @input="updateMapping(index, 'to', $event)"
            />
            <button
              type="button"
              class="model-restriction-section__remove-button"
              :aria-label="t('common.delete')"
              @click="emit('removeMapping', index)"
            >
              <Icon name="trash" size="sm" :stroke-width="2" />
            </button>
          </div>
        </div>

        <button
          type="button"
          class="btn btn-secondary mb-3 w-full border-2 border-dashed"
          @click="emit('addMapping')"
        >
          <Icon name="plus" size="sm" class="mr-1 inline" :stroke-width="2" />
          {{ t("admin.accounts.addMapping") }}
        </button>

        <div v-if="presetMappings.length > 0" class="flex flex-wrap gap-2">
          <button
            v-for="preset in presetMappings"
            :key="preset.label"
            type="button"
            :class="getPresetMappingChipClasses(preset.tone)"
            @click="emit('addPreset', preset.from, preset.to)"
          >
            + {{ preset.label }}
          </button>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import Icon from "@/components/icons/Icon.vue";
import ModelWhitelistSelector from "@/components/account/ModelWhitelistSelector.vue";
import {
  getPresetMappingChipClasses,
  type PresetMapping,
} from "@/composables/useModelWhitelist";
import type { ModelMapping } from "@/components/account/credentialsBuilder";

type ModelRestrictionMode = "whitelist" | "mapping";
type ModeTone = "accent" | "purple";

const props = withDefaults(
  defineProps<{
    mode: ModelRestrictionMode;
    allowedModels: string[];
    platform: string;
    platforms?: string[];
    mappings: ModelMapping[];
    presetMappings: PresetMapping[];
    mappingKey: (mapping: ModelMapping) => string;
    disabled?: boolean;
    disabledMessageKey?: string;
    framed?: boolean;
    fromPlaceholderKey?: string;
    toPlaceholderKey?: string;
    showMappingNotice?: boolean;
  }>(),
  {
    disabled: false,
    disabledMessageKey: "",
    framed: true,
    fromPlaceholderKey: "admin.accounts.requestModel",
    toPlaceholderKey: "admin.accounts.actualModel",
    showMappingNotice: true,
  },
);

const emit = defineEmits<{
  "update:mode": [value: ModelRestrictionMode];
  "update:allowedModels": [value: string[]];
  addMapping: [];
  removeMapping: [index: number];
  addPreset: [from: string, to: string];
  updateMapping: [index: number, field: keyof ModelMapping, value: string];
}>();

const { t } = useI18n();

const fromPlaceholder = computed(() => t(props.fromPlaceholderKey));
const toPlaceholder = computed(() => t(props.toPlaceholderKey));

const getModeButtonClasses = (isSelected: boolean, tone: ModeTone) => [
  "model-restriction-section__mode-button",
  isSelected && `model-restriction-section__mode-button--${tone}`,
];

const updateMapping = (
  index: number,
  field: keyof ModelMapping,
  event: Event,
) => {
  emit(
    "updateMapping",
    index,
    field,
    (event.target as HTMLInputElement).value,
  );
};
</script>

<style scoped>
.model-restriction-section {
  border-top: 1px solid
    color-mix(in srgb, var(--theme-page-border) 76%, transparent);
  padding-top: 1rem;
}

.model-restriction-section__description,
.model-restriction-section__arrow {
  color: var(--theme-page-muted);
}

.model-restriction-section__notice {
  border-radius: var(--theme-auth-feedback-radius);
  padding: var(--theme-auth-callback-feedback-padding);
  border-color: color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.model-restriction-section__notice--amber {
  background: color-mix(
    in srgb,
    rgb(var(--theme-warning-rgb)) 10%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--theme-warning-rgb)) 84%,
    var(--theme-page-text)
  );
}

.model-restriction-section__notice--purple {
  background: color-mix(
    in srgb,
    rgb(var(--theme-brand-purple-rgb)) 10%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--theme-brand-purple-rgb)) 84%,
    var(--theme-page-text)
  );
}

.model-restriction-section__mode-button {
  flex: 1 1 0;
  border-radius: var(--theme-button-radius);
  background: color-mix(
    in srgb,
    var(--theme-surface-soft) 86%,
    var(--theme-surface)
  );
  color: var(--theme-page-muted);
  font-size: 0.875rem;
  font-weight: 500;
  padding: 0.5rem 1rem;
}

.model-restriction-section__mode-button:hover,
.model-restriction-section__mode-button:focus-visible {
  background: color-mix(
    in srgb,
    var(--theme-page-border) 66%,
    var(--theme-surface)
  );
  color: var(--theme-page-text);
  outline: none;
}

.model-restriction-section__mode-button--accent {
  background: color-mix(in srgb, var(--theme-accent) 14%, var(--theme-surface));
  color: color-mix(in srgb, var(--theme-accent) 90%, var(--theme-page-text));
}

.model-restriction-section__mode-button--purple {
  background: color-mix(
    in srgb,
    rgb(var(--theme-brand-purple-rgb)) 14%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--theme-brand-purple-rgb)) 88%,
    var(--theme-page-text)
  );
}

.model-restriction-section__remove-button {
  border-radius: var(--theme-button-radius);
  background: color-mix(
    in srgb,
    rgb(var(--theme-danger-rgb)) 12%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--theme-danger-rgb)) 88%,
    var(--theme-page-text)
  );
  padding: 0.5rem;
}
</style>
