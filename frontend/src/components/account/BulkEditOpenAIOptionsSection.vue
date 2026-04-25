<template>
  <BulkEditApplySection
    v-if="showPassthrough"
    id="bulk-edit-openai-passthrough"
    :enabled="passthroughApplyEnabled"
    label-key="admin.accounts.openai.oauthPassthrough"
    hint-key="admin.accounts.openai.oauthPassthroughDesc"
    @update:enabled="emit('update:passthroughApplyEnabled', $event)"
  >
    <AccountModalSwitch
      id="bulk-edit-openai-passthrough-toggle"
      :model-value="passthroughEnabled"
      :aria-label="t('admin.accounts.openai.oauthPassthrough')"
      @update:model-value="emit('update:passthroughEnabled', $event)"
    />
  </BulkEditApplySection>

  <BulkEditApplySection
    v-if="showWsMode"
    id="bulk-edit-openai-ws-mode"
    :enabled="wsModeApplyEnabled"
    label-key="admin.accounts.openai.wsMode"
    @update:enabled="emit('update:wsModeApplyEnabled', $event)"
  >
    <p class="bulk-edit-openai-options-section__description mb-3 text-xs">
      {{ t("admin.accounts.openai.wsModeDesc") }}
    </p>
    <p class="bulk-edit-openai-options-section__description mb-3 text-xs">
      {{ t(wsModeConcurrencyHintKey) }}
    </p>
    <Select
      :model-value="wsMode"
      data-testid="bulk-edit-openai-ws-mode-select"
      :options="wsModeOptions"
      aria-labelledby="bulk-edit-openai-ws-mode-label"
      @update:model-value="updateWsMode"
    />
  </BulkEditApplySection>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import AccountModalSwitch from "@/components/account/AccountModalSwitch.vue";
import BulkEditApplySection from "@/components/account/BulkEditApplySection.vue";
import Select from "@/components/common/Select.vue";
import { normalizeOpenAIWSMode, type OpenAIWSMode } from "@/utils/openaiWsMode";

defineProps<{
  passthroughApplyEnabled: boolean;
  passthroughEnabled: boolean;
  showPassthrough: boolean;
  showWsMode: boolean;
  wsMode: OpenAIWSMode;
  wsModeApplyEnabled: boolean;
  wsModeConcurrencyHintKey: string;
  wsModeOptions: Array<{ value: OpenAIWSMode; label: string }>;
}>();

const emit = defineEmits<{
  "update:passthroughApplyEnabled": [value: boolean];
  "update:passthroughEnabled": [value: boolean];
  "update:wsModeApplyEnabled": [value: boolean];
  "update:wsMode": [value: OpenAIWSMode];
}>();

const { t } = useI18n();

const updateWsMode = (value: string | number | boolean | null) => {
  const mode = normalizeOpenAIWSMode(value);
  if (mode) {
    emit("update:wsMode", mode);
  }
};
</script>

<style scoped>
.bulk-edit-openai-options-section__description {
  color: var(--theme-page-muted);
}
</style>
