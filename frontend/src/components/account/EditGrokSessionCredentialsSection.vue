<template>
  <div class="space-y-4">
    <div>
      <label class="input-label">{{
        t("admin.accounts.grok.sessionToken")
      }}</label>
      <input
        :value="sessionToken"
        type="password"
        class="input font-mono"
        :placeholder="t('admin.accounts.grok.sessionTokenPlaceholder')"
        @input="
          emit(
            'update:sessionToken',
            ($event.target as HTMLInputElement).value,
          )
        "
      />
      <p class="input-hint">{{ t("admin.accounts.leaveEmptyToKeep") }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";

defineProps<{
  sessionToken: string;
}>();

const emit = defineEmits<{
  "update:sessionToken": [value: string];
}>();

const { t } = useI18n();
</script>
