<template>
  <div>
    <label class="input-label">{{
      t("admin.accounts.gemini.tier.label")
    }}</label>
    <select
      :value="tierAiStudio"
      class="input"
      @change="updateTierAiStudio"
    >
      <option value="aistudio_free">
        {{ t("admin.accounts.gemini.tier.aiStudio.free") }}
      </option>
      <option value="aistudio_paid">
        {{ t("admin.accounts.gemini.tier.aiStudio.paid") }}
      </option>
    </select>
    <p class="input-hint">
      {{ t("admin.accounts.gemini.tier.aiStudioHint") }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import type { GeminiAIStudioTier } from "@/components/account/createAccountModalHelpers";

defineProps<{
  tierAiStudio: GeminiAIStudioTier;
}>();

const emit = defineEmits<{
  "update:tierAiStudio": [value: GeminiAIStudioTier];
}>();

const { t } = useI18n();

const updateTierAiStudio = (event: Event) => {
  emit(
    "update:tierAiStudio",
    (event.target as HTMLSelectElement).value as GeminiAIStudioTier,
  );
};
</script>
