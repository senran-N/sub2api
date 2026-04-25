<template>
  <div class="space-y-4">
    <div>
      <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
        <label class="input-label mb-0">{{
          t("admin.accounts.baseUrl")
        }}</label>
        <div v-if="baseUrlPresets.length > 0" class="flex flex-wrap gap-2">
          <button
            v-for="preset in baseUrlPresets"
            :key="preset.value"
            type="button"
            :class="getPresetMappingChipClasses('success')"
            @click="emit('update:baseUrl', preset.value)"
          >
            {{ preset.label }}
          </button>
        </div>
      </div>
      <input
        :value="baseUrl"
        type="text"
        class="input"
        :placeholder="baseUrlPlaceholder"
        @input="emit('update:baseUrl', ($event.target as HTMLInputElement).value)"
      />
      <p class="input-hint">{{ baseUrlHint }}</p>
    </div>

    <div>
      <label class="input-label">{{ resolvedApiKeyLabel }}</label>
      <input
        :value="apiKeyValue"
        type="password"
        :required="apiKeyRequired"
        class="input font-mono"
        :autocomplete="apiKeyAutocomplete || undefined"
        :data-1p-ignore="ignorePasswordManagers ? '' : undefined"
        :data-lpignore="ignorePasswordManagers ? 'true' : undefined"
        :data-bwignore="ignorePasswordManagers ? 'true' : undefined"
        :placeholder="apiKeyPlaceholder"
        @input="emit('update:apiKeyValue', ($event.target as HTMLInputElement).value)"
      />
      <p class="input-hint">{{ apiKeyHint }}</p>
    </div>

    <GeminiApiKeyTierSection
      v-if="showGeminiApiKeyTier && platform === 'gemini'"
      :tier-ai-studio="tierAiStudio"
      @update:tier-ai-studio="emit('update:tierAiStudio', $event)"
    />

    <ModelRestrictionSection
      v-if="showModelRestriction"
      :mode="modelRestrictionMode"
      :allowed-models="allowedModels"
      :platform="platform"
      :mappings="mappings"
      :preset-mappings="presetMappings"
      :mapping-key="mappingKey"
      :disabled="modelRestrictionDisabled"
      @update:mode="emit('update:modelRestrictionMode', $event)"
      @update:allowed-models="emit('update:allowedModels', $event)"
      @add-mapping="emit('addMapping')"
      @remove-mapping="emit('removeMapping', $event)"
      @add-preset="(from, to) => emit('addPreset', from, to)"
      @update-mapping="
        (index, field, value) => emit('updateMapping', index, field, value)
      "
    />

    <PoolModeSection
      :enabled="poolModeEnabled"
      :retry-count="poolModeRetryCount"
      @update:enabled="emit('update:poolModeEnabled', $event)"
      @update:retry-count="emit('update:poolModeRetryCount', $event)"
    />

    <CustomErrorCodesSection
      :enabled="customErrorCodesEnabled"
      :input-value="customErrorCodeInput"
      :selected-codes="selectedErrorCodes"
      @update:enabled="emit('update:customErrorCodesEnabled', $event)"
      @update:input-value="emit('update:customErrorCodeInput', $event)"
      @toggle-code="emit('toggleCode', $event)"
      @add-code="emit('addCode')"
      @remove-code="emit('removeCode', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import CustomErrorCodesSection from "@/components/account/CustomErrorCodesSection.vue";
import GeminiApiKeyTierSection from "@/components/account/GeminiApiKeyTierSection.vue";
import ModelRestrictionSection from "@/components/account/ModelRestrictionSection.vue";
import PoolModeSection from "@/components/account/PoolModeSection.vue";
import {
  getPresetMappingChipClasses,
  type PresetMapping,
} from "@/composables/useModelWhitelist";
import type { AccountPlatform } from "@/types";
import type { AccountBaseUrlPreset } from "@/components/account/accountModalShared";
import type { GeminiAIStudioTier } from "@/components/account/createAccountModalHelpers";
import type { ModelMapping } from "@/components/account/credentialsBuilder";

type ModelRestrictionMode = "whitelist" | "mapping";

const props = withDefaults(
  defineProps<{
    allowedModels: string[];
    apiKeyAutocomplete?: string;
    apiKeyHint: string;
    apiKeyLabel?: string;
    apiKeyPlaceholder: string;
    apiKeyRequired?: boolean;
    apiKeyValue: string;
    baseUrl: string;
    baseUrlHint: string;
    baseUrlPlaceholder: string;
    baseUrlPresets: AccountBaseUrlPreset[];
    customErrorCodeInput: number | null;
    customErrorCodesEnabled: boolean;
    ignorePasswordManagers?: boolean;
    mappingKey: (mapping: ModelMapping) => string;
    mappings: ModelMapping[];
    modelRestrictionDisabled: boolean;
    modelRestrictionMode: ModelRestrictionMode;
    platform: AccountPlatform;
    poolModeEnabled: boolean;
    poolModeRetryCount: number;
    presetMappings: PresetMapping[];
    selectedErrorCodes: number[];
    showGeminiApiKeyTier?: boolean;
    showModelRestriction?: boolean;
    tierAiStudio?: GeminiAIStudioTier;
  }>(),
  {
    apiKeyAutocomplete: "",
    apiKeyLabel: "",
    apiKeyRequired: false,
    ignorePasswordManagers: false,
    showGeminiApiKeyTier: true,
    showModelRestriction: true,
    tierAiStudio: "aistudio_free",
  },
);

const emit = defineEmits<{
  "update:allowedModels": [value: string[]];
  "update:apiKeyValue": [value: string];
  "update:baseUrl": [value: string];
  "update:customErrorCodeInput": [value: number | null];
  "update:customErrorCodesEnabled": [value: boolean];
  "update:modelRestrictionMode": [value: ModelRestrictionMode];
  "update:poolModeEnabled": [value: boolean];
  "update:poolModeRetryCount": [value: number];
  "update:tierAiStudio": [value: GeminiAIStudioTier];
  addCode: [];
  addMapping: [];
  addPreset: [from: string, to: string];
  removeCode: [code: number];
  removeMapping: [index: number];
  toggleCode: [code: number];
  updateMapping: [index: number, field: keyof ModelMapping, value: string];
}>();

const { t } = useI18n();

const resolvedApiKeyLabel = computed(
  () => props.apiKeyLabel || t("admin.accounts.apiKeyRequired"),
);
</script>
