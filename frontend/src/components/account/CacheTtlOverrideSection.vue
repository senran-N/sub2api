<template>
  <QuotaControlCard
    :enabled="enabled"
    title-key="admin.accounts.quotaControl.cacheTTLOverride.label"
    hint-key="admin.accounts.quotaControl.cacheTTLOverride.hint"
    @update:enabled="emit('update:enabled', $event)"
  >
    <div class="mt-3">
      <label class="input-label text-xs">{{
        t("admin.accounts.quotaControl.cacheTTLOverride.target")
      }}</label>
      <select
        :value="target"
        class="input mt-1"
        @change="
          emit('update:target', ($event.target as HTMLSelectElement).value)
        "
      >
        <option value="5m">5m</option>
        <option value="1h">1h</option>
      </select>
      <p class="input-hint mt-1">
        {{ t("admin.accounts.quotaControl.cacheTTLOverride.targetHint") }}
      </p>
    </div>
  </QuotaControlCard>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import QuotaControlCard from "@/components/account/QuotaControlCard.vue";

defineProps<{
  enabled: boolean;
  target: string;
}>();

const emit = defineEmits<{
  "update:enabled": [value: boolean];
  "update:target": [value: string];
}>();

const { t } = useI18n();
</script>
