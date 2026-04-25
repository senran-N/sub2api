<template>
  <div class="space-y-4">
    <div>
      <label class="input-label">
        {{ t("admin.accounts.upstream.baseUrl") }}
      </label>
      <input
        :value="baseUrl"
        type="text"
        required
        class="input"
        placeholder="https://cloudcode-pa.googleapis.com"
        @input="updateBaseUrl"
      />
      <p class="input-hint">
        {{ t("admin.accounts.upstream.baseUrlHint") }}
      </p>
    </div>
    <div>
      <label class="input-label">
        {{ t("admin.accounts.upstream.apiKey") }}
      </label>
      <input
        :value="apiKey"
        type="password"
        required
        class="input font-mono"
        placeholder="sk-..."
        @input="updateAPIKey"
      />
      <p class="input-hint">
        {{ t("admin.accounts.upstream.apiKeyHint") }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";

defineProps<{
  baseUrl: string;
  apiKey: string;
}>();

const emit = defineEmits<{
  "update:baseUrl": [value: string];
  "update:apiKey": [value: string];
}>();

const { t } = useI18n();

const readInputValue = (event: Event) => {
  const target = event.target;
  if (!(target instanceof HTMLInputElement)) {
    return null;
  }
  return target.value;
};

const updateBaseUrl = (event: Event) => {
  const value = readInputValue(event);
  if (value === null) {
    return;
  }
  emit("update:baseUrl", value);
};

const updateAPIKey = (event: Event) => {
  const value = readInputValue(event);
  if (value === null) {
    return;
  }
  emit("update:apiKey", value);
};
</script>
