<template>
  <section class="quota-limit-section form-section space-y-4">
    <div class="mb-3">
      <h3 class="input-label mb-0 text-base font-semibold">
        {{ t("admin.accounts.quotaLimit") }}
      </h3>
      <p class="quota-limit-section__description mt-1 text-xs">
        {{ t("admin.accounts.quotaLimitHint") }}
      </p>
    </div>

    <QuotaLimitCard
      :total-limit="totalLimit"
      :daily-limit="dailyLimit"
      :weekly-limit="weeklyLimit"
      :daily-reset-mode="dailyResetMode"
      :daily-reset-hour="dailyResetHour"
      :weekly-reset-mode="weeklyResetMode"
      :weekly-reset-day="weeklyResetDay"
      :weekly-reset-hour="weeklyResetHour"
      :reset-timezone="resetTimezone"
      @update:totalLimit="emit('update:totalLimit', $event)"
      @update:dailyLimit="emit('update:dailyLimit', $event)"
      @update:weeklyLimit="emit('update:weeklyLimit', $event)"
      @update:dailyResetMode="emit('update:dailyResetMode', $event)"
      @update:dailyResetHour="emit('update:dailyResetHour', $event)"
      @update:weeklyResetMode="emit('update:weeklyResetMode', $event)"
      @update:weeklyResetDay="emit('update:weeklyResetDay', $event)"
      @update:weeklyResetHour="emit('update:weeklyResetHour', $event)"
      @update:resetTimezone="emit('update:resetTimezone', $event)"
    />

    <div class="quota-limit-section__notify-card space-y-3">
      <div>
        <h4 class="input-label mb-1">
          {{ t("admin.accounts.quotaNotify.title") }}
        </h4>
        <p class="quota-limit-section__description text-xs">
          {{ t("admin.accounts.quotaNotify.hint") }}
        </p>
      </div>
      <div class="grid gap-3 md:grid-cols-3">
        <div>
          <label class="input-label">{{
            t("admin.accounts.quotaNotify.daily")
          }}</label>
          <QuotaNotifyToggle
            :enabled="notifyDailyEnabled"
            :threshold="notifyDailyThreshold"
            :threshold-type="notifyDailyThresholdType"
            @update:enabled="emit('update:notifyDailyEnabled', $event)"
            @update:threshold="emit('update:notifyDailyThreshold', $event)"
            @update:thresholdType="
              emit('update:notifyDailyThresholdType', $event)
            "
          />
        </div>
        <div>
          <label class="input-label">{{
            t("admin.accounts.quotaNotify.weekly")
          }}</label>
          <QuotaNotifyToggle
            :enabled="notifyWeeklyEnabled"
            :threshold="notifyWeeklyThreshold"
            :threshold-type="notifyWeeklyThresholdType"
            @update:enabled="emit('update:notifyWeeklyEnabled', $event)"
            @update:threshold="emit('update:notifyWeeklyThreshold', $event)"
            @update:thresholdType="
              emit('update:notifyWeeklyThresholdType', $event)
            "
          />
        </div>
        <div>
          <label class="input-label">{{
            t("admin.accounts.quotaNotify.total")
          }}</label>
          <QuotaNotifyToggle
            :enabled="notifyTotalEnabled"
            :threshold="notifyTotalThreshold"
            :threshold-type="notifyTotalThresholdType"
            @update:enabled="emit('update:notifyTotalEnabled', $event)"
            @update:threshold="emit('update:notifyTotalThreshold', $event)"
            @update:thresholdType="
              emit('update:notifyTotalThresholdType', $event)
            "
          />
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import QuotaLimitCard from "@/components/account/QuotaLimitCard.vue";
import QuotaNotifyToggle from "@/components/account/QuotaNotifyToggle.vue";

type QuotaResetMode = "rolling" | "fixed" | null;
type QuotaThresholdType = "fixed" | "percentage" | null;

defineProps<{
  dailyLimit: number | null;
  dailyResetHour: number | null;
  dailyResetMode: QuotaResetMode;
  notifyDailyEnabled: boolean | null;
  notifyDailyThreshold: number | null;
  notifyDailyThresholdType: QuotaThresholdType;
  notifyTotalEnabled: boolean | null;
  notifyTotalThreshold: number | null;
  notifyTotalThresholdType: QuotaThresholdType;
  notifyWeeklyEnabled: boolean | null;
  notifyWeeklyThreshold: number | null;
  notifyWeeklyThresholdType: QuotaThresholdType;
  resetTimezone: string | null;
  totalLimit: number | null;
  weeklyLimit: number | null;
  weeklyResetDay: number | null;
  weeklyResetHour: number | null;
  weeklyResetMode: QuotaResetMode;
}>();

const emit = defineEmits<{
  "update:dailyLimit": [value: number | null];
  "update:dailyResetHour": [value: number | null];
  "update:dailyResetMode": [value: QuotaResetMode];
  "update:notifyDailyEnabled": [value: boolean | null];
  "update:notifyDailyThreshold": [value: number | null];
  "update:notifyDailyThresholdType": [value: QuotaThresholdType];
  "update:notifyTotalEnabled": [value: boolean | null];
  "update:notifyTotalThreshold": [value: number | null];
  "update:notifyTotalThresholdType": [value: QuotaThresholdType];
  "update:notifyWeeklyEnabled": [value: boolean | null];
  "update:notifyWeeklyThreshold": [value: number | null];
  "update:notifyWeeklyThresholdType": [value: QuotaThresholdType];
  "update:resetTimezone": [value: string | null];
  "update:totalLimit": [value: number | null];
  "update:weeklyLimit": [value: number | null];
  "update:weeklyResetDay": [value: number | null];
  "update:weeklyResetHour": [value: number | null];
  "update:weeklyResetMode": [value: QuotaResetMode];
}>();

const { t } = useI18n();
</script>

<style scoped>
.quota-limit-section {
  border-top: 1px solid
    color-mix(in srgb, var(--theme-page-border) 76%, transparent);
  padding-top: 1rem;
}

.quota-limit-section__description {
  color: var(--theme-page-muted);
}

.quota-limit-section__notify-card {
  border: 1px solid
    color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  border-radius: var(--theme-surface-radius);
  padding: 1rem;
  background: var(--theme-surface);
}
</style>
