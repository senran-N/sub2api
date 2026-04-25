<template>
  <QuotaControlCard
    :enabled="enabled"
    title-key="admin.accounts.quotaControl.sessionLimit.label"
    hint-key="admin.accounts.quotaControl.sessionLimit.hint"
    @update:enabled="emit('update:enabled', $event)"
  >
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4">
      <div>
        <label class="input-label">{{
          t("admin.accounts.quotaControl.sessionLimit.maxSessions")
        }}</label>
        <input
          :value="maxSessions ?? ''"
          type="number"
          min="1"
          step="1"
          class="input"
          :placeholder="
            t(
              'admin.accounts.quotaControl.sessionLimit.maxSessionsPlaceholder',
            )
          "
          @input="emitNumber('update:maxSessions', $event)"
        />
        <p class="input-hint">
          {{ t("admin.accounts.quotaControl.sessionLimit.maxSessionsHint") }}
        </p>
      </div>
      <div>
        <label class="input-label">{{
          t("admin.accounts.quotaControl.sessionLimit.idleTimeout")
        }}</label>
        <div class="relative">
          <input
            :value="idleTimeout ?? ''"
            type="number"
            min="1"
            step="1"
            class="input pr-12"
            :placeholder="
              t(
                'admin.accounts.quotaControl.sessionLimit.idleTimeoutPlaceholder',
              )
            "
            @input="emitNumber('update:idleTimeout', $event)"
          />
          <span
            class="session-limit-control-section__affix absolute right-3 top-1/2 -translate-y-1/2"
            >{{ t("common.minutes") }}</span
          >
        </div>
        <p class="input-hint">
          {{ t("admin.accounts.quotaControl.sessionLimit.idleTimeoutHint") }}
        </p>
      </div>
    </div>
  </QuotaControlCard>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import QuotaControlCard from "@/components/account/QuotaControlCard.vue";

type NumberUpdateEvent = "update:maxSessions" | "update:idleTimeout";

defineProps<{
  enabled: boolean;
  maxSessions: number | null;
  idleTimeout: number | null;
}>();

const emit = defineEmits<{
  "update:enabled": [value: boolean];
  "update:maxSessions": [value: number | null];
  "update:idleTimeout": [value: number | null];
}>();

const { t } = useI18n();

const emitNumber = (eventName: NumberUpdateEvent, event: Event) => {
  const value = (event.target as HTMLInputElement).value.trim();
  const nextValue = value === "" ? null : Number(value);
  if (eventName === "update:maxSessions") {
    emit("update:maxSessions", nextValue);
    return;
  }
  emit("update:idleTimeout", nextValue);
};
</script>

<style scoped>
.session-limit-control-section__affix {
  color: var(--theme-page-muted);
}
</style>
