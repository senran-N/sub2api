<template>
  <QuotaControlCard
    :enabled="enabled"
    title-key="admin.accounts.quotaControl.rpmLimit.label"
    hint-key="admin.accounts.quotaControl.rpmLimit.hint"
    @update:enabled="emit('update:enabled', $event)"
  >
    <div class="space-y-4">
      <div>
        <label class="input-label">{{
          t("admin.accounts.quotaControl.rpmLimit.baseRpm")
        }}</label>
        <input
          :value="baseRpm ?? ''"
          type="number"
          min="1"
          max="1000"
          step="1"
          class="input"
          :placeholder="
            t('admin.accounts.quotaControl.rpmLimit.baseRpmPlaceholder')
          "
          @input="emitNumber('update:baseRpm', $event)"
        />
        <p class="input-hint">
          {{ t("admin.accounts.quotaControl.rpmLimit.baseRpmHint") }}
        </p>
      </div>

      <div>
        <label class="input-label">{{
          t("admin.accounts.quotaControl.rpmLimit.strategy")
        }}</label>
        <div class="flex gap-2">
          <button
            type="button"
            :class="getModeButtonClasses(strategy === 'tiered')"
            @click="emit('update:strategy', 'tiered')"
          >
            <div class="text-center">
              <div>
                {{ t("admin.accounts.quotaControl.rpmLimit.strategyTiered") }}
              </div>
              <div class="mt-0.5 text-[10px] opacity-70">
                {{
                  t("admin.accounts.quotaControl.rpmLimit.strategyTieredHint")
                }}
              </div>
            </div>
          </button>
          <button
            type="button"
            :class="getModeButtonClasses(strategy === 'sticky_exempt')"
            @click="emit('update:strategy', 'sticky_exempt')"
          >
            <div class="text-center">
              <div>
                {{
                  t(
                    "admin.accounts.quotaControl.rpmLimit.strategyStickyExempt",
                  )
                }}
              </div>
              <div class="mt-0.5 text-[10px] opacity-70">
                {{
                  t(
                    "admin.accounts.quotaControl.rpmLimit.strategyStickyExemptHint",
                  )
                }}
              </div>
            </div>
          </button>
        </div>
      </div>

      <div v-if="strategy === 'tiered'">
        <label class="input-label">{{
          t("admin.accounts.quotaControl.rpmLimit.stickyBuffer")
        }}</label>
        <input
          :value="stickyBuffer ?? ''"
          type="number"
          min="1"
          step="1"
          class="input"
          :placeholder="
            t('admin.accounts.quotaControl.rpmLimit.stickyBufferPlaceholder')
          "
          @input="emitNumber('update:stickyBuffer', $event)"
        />
        <p class="input-hint">
          {{ t("admin.accounts.quotaControl.rpmLimit.stickyBufferHint") }}
        </p>
      </div>
    </div>

    <template #footer>
      <div class="mt-4">
        <label class="input-label">{{
          t("admin.accounts.quotaControl.rpmLimit.userMsgQueue")
        }}</label>
        <p class="rpm-limit-control-section__description mt-1 mb-2 text-xs">
          {{ t("admin.accounts.quotaControl.rpmLimit.userMsgQueueHint") }}
        </p>
        <div class="flex space-x-2">
          <button
            v-for="option in userMsgQueueModeOptions"
            :key="option.value"
            type="button"
            :class="
              getSegmentButtonClasses(userMsgQueueMode === option.value)
            "
            @click="emit('update:userMsgQueueMode', option.value)"
          >
            {{ option.label }}
          </button>
        </div>
      </div>
    </template>
  </QuotaControlCard>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import QuotaControlCard from "@/components/account/QuotaControlCard.vue";

export type RpmLimitStrategy = "tiered" | "sticky_exempt";

type NumberUpdateEvent = "update:baseRpm" | "update:stickyBuffer";

defineProps<{
  enabled: boolean;
  baseRpm: number | null;
  strategy: RpmLimitStrategy;
  stickyBuffer: number | null;
  userMsgQueueMode: string;
  userMsgQueueModeOptions: Array<{ value: string; label: string }>;
}>();

const emit = defineEmits<{
  "update:enabled": [value: boolean];
  "update:baseRpm": [value: number | null];
  "update:strategy": [value: RpmLimitStrategy];
  "update:stickyBuffer": [value: number | null];
  "update:userMsgQueueMode": [value: string];
}>();

const { t } = useI18n();

const getModeButtonClasses = (isSelected: boolean) => [
  "rpm-limit-control-section__mode-button",
  isSelected && "rpm-limit-control-section__mode-button--selected",
];

const getSegmentButtonClasses = (isSelected: boolean) => [
  "rpm-limit-control-section__segment-button",
  isSelected && "rpm-limit-control-section__segment-button--selected",
];

const emitNumber = (eventName: NumberUpdateEvent, event: Event) => {
  const value = (event.target as HTMLInputElement).value.trim();
  const nextValue = value === "" ? null : Number(value);
  if (eventName === "update:baseRpm") {
    emit("update:baseRpm", nextValue);
    return;
  }
  emit("update:stickyBuffer", nextValue);
};
</script>

<style scoped>
.rpm-limit-control-section__description {
  color: var(--theme-page-muted);
}

.rpm-limit-control-section__mode-button,
.rpm-limit-control-section__segment-button {
  border-radius: var(--theme-button-radius);
  background: color-mix(
    in srgb,
    var(--theme-surface-soft) 86%,
    var(--theme-surface)
  );
  color: var(--theme-page-muted);
}

.rpm-limit-control-section__mode-button {
  flex: 1 1 0;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
}

.rpm-limit-control-section__segment-button {
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
}

.rpm-limit-control-section__mode-button:hover,
.rpm-limit-control-section__mode-button:focus-visible,
.rpm-limit-control-section__segment-button:hover,
.rpm-limit-control-section__segment-button:focus-visible {
  background: color-mix(
    in srgb,
    var(--theme-page-border) 66%,
    var(--theme-surface)
  );
  color: var(--theme-page-text);
  outline: none;
}

.rpm-limit-control-section__mode-button--selected,
.rpm-limit-control-section__segment-button--selected {
  background: color-mix(in srgb, var(--theme-accent) 14%, var(--theme-surface));
  color: color-mix(in srgb, var(--theme-accent) 90%, var(--theme-page-text));
}
</style>
