<template>
  <div class="space-y-4">
    <BedrockCredentialsSection
      :auth-mode="authMode"
      v-model:access-key-id="accessKeyId"
      v-model:secret-access-key="secretAccessKey"
      v-model:session-token="sessionToken"
      v-model:api-key-value="apiKeyValue"
      v-model:region="region"
      v-model:force-global="forceGlobal"
      :allow-auth-mode-change="false"
      :credentials-required="false"
      region-control="input"
      secret-access-key-placeholder-key="admin.accounts.bedrockSecretKeyLeaveEmpty"
      secret-access-key-hint-key="admin.accounts.bedrockSecretKeyLeaveEmpty"
      session-token-placeholder-key="admin.accounts.bedrockSecretKeyLeaveEmpty"
      api-key-placeholder-key="admin.accounts.bedrockApiKeyLeaveEmpty"
      api-key-hint-key="admin.accounts.bedrockApiKeyLeaveEmpty"
    />

    <ModelRestrictionSection
      v-model:mode="modelRestrictionMode"
      v-model:allowed-models="allowedModels"
      platform="anthropic"
      :mappings="mappings"
      :preset-mappings="presetMappings"
      :mapping-key="mappingKey"
      from-placeholder-key="admin.accounts.fromModel"
      to-placeholder-key="admin.accounts.toModel"
      :show-mapping-notice="false"
      @add-mapping="emit('addMapping')"
      @remove-mapping="emit('removeMapping', $event)"
      @add-preset="forwardAddPreset"
      @update-mapping="forwardUpdateMapping"
    />

    <PoolModeSection
      v-model:enabled="poolModeEnabled"
      v-model:retry-count="poolModeRetryCount"
    />
  </div>
</template>

<script setup lang="ts">
import BedrockCredentialsSection from "@/components/account/BedrockCredentialsSection.vue";
import ModelRestrictionSection from "@/components/account/ModelRestrictionSection.vue";
import PoolModeSection from "@/components/account/PoolModeSection.vue";
import type { BedrockAuthMode } from "@/components/account/createAccountModalHelpers";
import type { ModelMapping } from "@/components/account/credentialsBuilder";
import type { PresetMapping } from "@/composables/useModelWhitelist";

type ModelRestrictionMode = "whitelist" | "mapping";

defineProps<{
  authMode: BedrockAuthMode;
  mappings: ModelMapping[];
  mappingKey: (mapping: ModelMapping) => string;
  presetMappings: PresetMapping[];
}>();

const accessKeyId = defineModel<string>("accessKeyId", { required: true });
const secretAccessKey = defineModel<string>("secretAccessKey", {
  required: true,
});
const sessionToken = defineModel<string>("sessionToken", { required: true });
const apiKeyValue = defineModel<string>("apiKeyValue", { required: true });
const region = defineModel<string>("region", { required: true });
const forceGlobal = defineModel<boolean>("forceGlobal", { required: true });
const modelRestrictionMode = defineModel<ModelRestrictionMode>(
  "modelRestrictionMode",
  { required: true },
);
const allowedModels = defineModel<string[]>("allowedModels", {
  required: true,
});
const poolModeEnabled = defineModel<boolean>("poolModeEnabled", {
  required: true,
});
const poolModeRetryCount = defineModel<number>("poolModeRetryCount", {
  required: true,
});

const emit = defineEmits<{
  addMapping: [];
  removeMapping: [index: number];
  addPreset: [from: string, to: string];
  updateMapping: [index: number, field: keyof ModelMapping, value: string];
}>();

const forwardAddPreset = (from: string, to: string) => {
  emit("addPreset", from, to);
};

const forwardUpdateMapping = (
  index: number,
  field: keyof ModelMapping,
  value: string,
) => {
  emit("updateMapping", index, field, value);
};
</script>
