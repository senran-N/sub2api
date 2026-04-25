<template>
  <div class="form-section">
    <div class="mb-3 flex items-center justify-between gap-4">
      <div>
        <label class="input-label mb-0">{{
          t("admin.accounts.poolMode")
        }}</label>
        <p class="pool-mode-section__description mt-1 text-xs">
          {{ t("admin.accounts.poolModeHint") }}
        </p>
      </div>
      <AccountModalSwitch
        :model-value="enabled"
        :aria-label="t('admin.accounts.poolMode')"
        @update:model-value="emit('update:enabled', $event)"
      />
    </div>
    <div v-if="enabled" class="pool-mode-section__notice">
      <p class="text-xs">
        <Icon
          name="exclamationCircle"
          size="sm"
          class="mr-1 inline"
          :stroke-width="2"
        />
        {{ t("admin.accounts.poolModeInfo") }}
      </p>
    </div>
    <div v-if="enabled" class="mt-3">
      <label class="input-label">{{
        t("admin.accounts.poolModeRetryCount")
      }}</label>
      <input
        :value="retryCount"
        type="number"
        min="0"
        :max="MAX_POOL_MODE_RETRY_COUNT"
        step="1"
        class="input"
        @input="updateRetryCount"
      />
      <p class="pool-mode-section__description mt-1 text-xs">
        {{
          t("admin.accounts.poolModeRetryCountHint", {
            default: DEFAULT_POOL_MODE_RETRY_COUNT,
            max: MAX_POOL_MODE_RETRY_COUNT,
          })
        }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import AccountModalSwitch from "@/components/account/AccountModalSwitch.vue";
import Icon from "@/components/icons/Icon.vue";
import {
  DEFAULT_POOL_MODE_RETRY_COUNT,
  MAX_POOL_MODE_RETRY_COUNT,
} from "@/components/account/credentialsBuilder";

defineProps<{
  enabled: boolean;
  retryCount: number;
}>();

const emit = defineEmits<{
  "update:enabled": [value: boolean];
  "update:retryCount": [value: number];
}>();

const { t } = useI18n();

const updateRetryCount = (event: Event) => {
  emit(
    "update:retryCount",
    Number((event.target as HTMLInputElement).value),
  );
};
</script>

<style scoped>
.pool-mode-section__description {
  color: var(--theme-page-muted);
}

.pool-mode-section__notice {
  border-radius: var(--theme-auth-feedback-radius);
  padding: var(--theme-auth-callback-feedback-padding);
  border-color: color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  background: color-mix(
    in srgb,
    rgb(var(--theme-info-rgb)) 10%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--theme-info-rgb)) 84%,
    var(--theme-page-text)
  );
}
</style>
