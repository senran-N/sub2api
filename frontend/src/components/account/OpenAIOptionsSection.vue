<template>
  <div class="form-section">
    <div class="flex items-center justify-between gap-4">
      <div>
        <label class="input-label mb-0">{{
          t("admin.accounts.openai.oauthPassthrough")
        }}</label>
        <p class="openai-options-section__description mt-1 text-xs">
          {{ t("admin.accounts.openai.oauthPassthroughDesc") }}
        </p>
      </div>
      <AccountModalSwitch
        :model-value="passthroughEnabled"
        :aria-label="t('admin.accounts.openai.oauthPassthrough')"
        @update:model-value="emit('update:passthroughEnabled', $event)"
      />
    </div>
  </div>

  <div v-if="showTextRuntimeOptions" class="form-section">
    <div class="flex items-center justify-between gap-4">
      <div>
        <label class="input-label mb-0">{{
          t("admin.accounts.openai.wsMode")
        }}</label>
        <p class="openai-options-section__description mt-1 text-xs">
          {{ t("admin.accounts.openai.wsModeDesc") }}
        </p>
        <p class="openai-options-section__description mt-1 text-xs">
          {{ t(wsModeConcurrencyHintKey) }}
        </p>
      </div>
      <div class="w-52">
        <Select v-model="selectedWsMode" :options="wsModeOptions" />
      </div>
    </div>
  </div>

  <div v-if="accountCategory === 'oauth-based'" class="form-section">
    <div class="flex items-center justify-between gap-4">
      <div>
        <label class="input-label mb-0">{{
          t("admin.accounts.openai.codexCLIOnly")
        }}</label>
        <p class="openai-options-section__description mt-1 text-xs">
          {{ t("admin.accounts.openai.codexCLIOnlyDesc") }}
        </p>
      </div>
      <AccountModalSwitch
        :model-value="codexCliOnlyEnabled"
        :aria-label="t('admin.accounts.openai.codexCLIOnly')"
        @update:model-value="emit('update:codexCliOnlyEnabled', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import AccountModalSwitch from "@/components/account/AccountModalSwitch.vue";
import Select from "@/components/common/Select.vue";
import type { CreateAccountCategory } from "@/components/account/createAccountModalHelpers";
import type { OpenAIWSMode } from "@/utils/openaiWsMode";

const props = defineProps<{
  accountCategory: CreateAccountCategory;
  passthroughEnabled: boolean;
  wsMode: OpenAIWSMode;
  wsModeOptions: Array<{ value: OpenAIWSMode; label: string }>;
  wsModeConcurrencyHintKey: string;
  codexCliOnlyEnabled: boolean;
}>();

const emit = defineEmits<{
  "update:passthroughEnabled": [value: boolean];
  "update:wsMode": [value: OpenAIWSMode];
  "update:codexCliOnlyEnabled": [value: boolean];
}>();

const { t } = useI18n();

const showTextRuntimeOptions = computed(
  () =>
    props.accountCategory === "oauth-based" ||
    props.accountCategory === "apikey",
);

const selectedWsMode = computed({
  get: () => props.wsMode,
  set: (mode: OpenAIWSMode) => {
    emit("update:wsMode", mode);
  },
});

</script>

<style scoped>
.openai-options-section__description {
  color: var(--theme-page-muted);
}
</style>
