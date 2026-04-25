<template>
  <div :class="framed && 'form-section'">
    <div class="flex items-center justify-between gap-4">
      <div>
        <label class="input-label mb-0">{{
          t("admin.accounts.interceptWarmupRequests")
        }}</label>
        <p class="warmup-section__description mt-1 text-xs">
          {{ t("admin.accounts.interceptWarmupRequestsDesc") }}
        </p>
      </div>
      <AccountModalSwitch
        :model-value="enabled"
        :aria-label="t('admin.accounts.interceptWarmupRequests')"
        @update:model-value="emit('update:enabled', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import AccountModalSwitch from "@/components/account/AccountModalSwitch.vue";

withDefaults(defineProps<{
  enabled: boolean;
  framed?: boolean;
}>(), {
  framed: true,
});

const emit = defineEmits<{
  "update:enabled": [value: boolean];
}>();

const { t } = useI18n();
</script>

<style scoped>
.warmup-section__description {
  color: var(--theme-page-muted);
}
</style>
