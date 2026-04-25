<template>
  <QuotaControlCard
    :enabled="enabled"
    title-key="admin.accounts.quotaControl.windowCost.label"
    hint-key="admin.accounts.quotaControl.windowCost.hint"
    @update:enabled="emit('update:enabled', $event)"
  >
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4">
      <div>
        <label class="input-label">{{
          t("admin.accounts.quotaControl.windowCost.limit")
        }}</label>
        <div class="relative">
          <span
            class="window-cost-control-section__affix absolute left-3 top-1/2 -translate-y-1/2"
            >$</span
          >
          <input
            :value="limit ?? ''"
            type="number"
            min="0"
            step="1"
            class="input pl-7"
            :placeholder="
              t('admin.accounts.quotaControl.windowCost.limitPlaceholder')
            "
            @input="emitNumber('update:limit', $event)"
          />
        </div>
        <p class="input-hint">
          {{ t("admin.accounts.quotaControl.windowCost.limitHint") }}
        </p>
      </div>
      <div>
        <label class="input-label">{{
          t("admin.accounts.quotaControl.windowCost.stickyReserve")
        }}</label>
        <div class="relative">
          <span
            class="window-cost-control-section__affix absolute left-3 top-1/2 -translate-y-1/2"
            >$</span
          >
          <input
            :value="stickyReserve ?? ''"
            type="number"
            min="0"
            step="1"
            class="input pl-7"
            :placeholder="
              t(
                'admin.accounts.quotaControl.windowCost.stickyReservePlaceholder',
              )
            "
            @input="emitNumber('update:stickyReserve', $event)"
          />
        </div>
        <p class="input-hint">
          {{ t("admin.accounts.quotaControl.windowCost.stickyReserveHint") }}
        </p>
      </div>
    </div>
  </QuotaControlCard>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import QuotaControlCard from "@/components/account/QuotaControlCard.vue";

type NumberUpdateEvent = "update:limit" | "update:stickyReserve";

defineProps<{
  enabled: boolean;
  limit: number | null;
  stickyReserve: number | null;
}>();

const emit = defineEmits<{
  "update:enabled": [value: boolean];
  "update:limit": [value: number | null];
  "update:stickyReserve": [value: number | null];
}>();

const { t } = useI18n();

const emitNumber = (eventName: NumberUpdateEvent, event: Event) => {
  const value = (event.target as HTMLInputElement).value.trim();
  const nextValue = value === "" ? null : Number(value);
  if (eventName === "update:limit") {
    emit("update:limit", nextValue);
    return;
  }
  emit("update:stickyReserve", nextValue);
};
</script>

<style scoped>
.window-cost-control-section__affix {
  color: var(--theme-page-muted);
}
</style>
