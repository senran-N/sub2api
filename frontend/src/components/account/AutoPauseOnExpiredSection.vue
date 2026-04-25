<template>
  <div>
    <div class="flex items-center justify-between gap-4">
      <div>
        <label class="input-label mb-0">{{
          t("admin.accounts.autoPauseOnExpired")
        }}</label>
        <p class="auto-pause-on-expired-section__description mt-1 text-xs">
          {{ t("admin.accounts.autoPauseOnExpiredDesc") }}
        </p>
      </div>
      <AccountModalSwitch
        :model-value="enabled"
        :aria-label="t('admin.accounts.autoPauseOnExpired')"
        @update:model-value="emit('update:enabled', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import AccountModalSwitch from "@/components/account/AccountModalSwitch.vue";

defineProps<{
  enabled: boolean;
}>();

const emit = defineEmits<{
  "update:enabled": [value: boolean];
}>();

const { t } = useI18n();
</script>

<style scoped>
.auto-pause-on-expired-section__description {
  color: var(--theme-page-muted);
}
</style>
