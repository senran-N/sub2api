<template>
  <div v-if="account?.platform === 'grok'" class="form-section space-y-4">
    <div>
      <label class="input-label mb-0">{{
        t("admin.accounts.grok.runtime.title")
      }}</label>
      <p class="edit-account-modal__muted mt-1 text-xs">
        {{ t("admin.accounts.grok.runtime.hint") }}
      </p>
    </div>

    <div class="grid grid-cols-1 gap-3 lg:grid-cols-2">
      <div
        class="edit-account-modal__config-card edit-account-modal__config-card--compact space-y-2"
      >
        <div class="flex flex-wrap items-center gap-2">
          <span class="edit-account-modal__muted text-xs">{{
            t("admin.accounts.grok.runtime.tier")
          }}</span>
          <span :class="grokTierChipClass">
            {{ grokTierLabel }}
          </span>
        </div>
        <div class="grid grid-cols-1 gap-2 text-xs sm:grid-cols-2">
          <div>
            <span class="edit-account-modal__muted">{{
              t("admin.accounts.grok.runtime.authMode")
            }}</span>
            <div class="font-medium">{{ grokAuthModeLabel }}</div>
          </div>
          <div>
            <span class="edit-account-modal__muted">{{
              t("admin.accounts.grok.runtime.source")
            }}</span>
            <div>{{ grokTierSourceDisplay }}</div>
          </div>
          <div>
            <span class="edit-account-modal__muted">{{
              t("admin.accounts.grok.runtime.confidence")
            }}</span>
            <div>{{ grokTierConfidenceDisplay }}</div>
          </div>
        </div>
      </div>

      <div
        class="edit-account-modal__config-card edit-account-modal__config-card--compact space-y-2"
      >
        <div>
          <span class="edit-account-modal__muted text-xs">{{
            t("admin.accounts.grok.runtime.capabilityTitle")
          }}</span>
          <div class="mt-1 flex flex-wrap gap-1.5">
            <span
              v-for="capability in grokCapabilities"
              :key="capability"
              class="theme-chip theme-chip--compact theme-chip--info"
            >
              {{ t(`admin.accounts.grok.runtime.capabilities.${capability}`) }}
            </span>
            <span
              v-if="grokCapabilities.length === 0"
              class="edit-account-modal__muted text-xs"
            >
              {{ t("admin.accounts.grok.runtime.empty") }}
            </span>
          </div>
        </div>
        <div>
          <span class="edit-account-modal__muted text-xs">{{
            t("admin.accounts.grok.runtime.models")
          }}</span>
          <div class="mt-1 flex flex-wrap gap-1.5">
            <span
              v-for="model in grokModels"
              :key="model"
              class="theme-chip theme-chip--compact theme-chip--neutral font-mono"
            >
              {{ model }}
            </span>
            <span
              v-if="grokModels.length === 0"
              class="edit-account-modal__muted text-xs"
            >
              {{ t("admin.accounts.grok.runtime.empty") }}
            </span>
          </div>
        </div>
      </div>

      <div
        class="edit-account-modal__config-card edit-account-modal__config-card--compact space-y-2"
      >
        <span class="edit-account-modal__muted text-xs">{{
          t("admin.accounts.grok.runtime.quotaTitle")
        }}</span>
        <div v-if="grokQuotaWindows.length > 0" class="space-y-2">
          <div
            v-for="window in grokQuotaWindows"
            :key="window.name"
            class="rounded-lg border px-3 py-2 text-xs"
          >
            <div class="flex items-center justify-between gap-3">
              <span class="font-medium">{{
                t(`admin.accounts.grok.runtime.windows.${window.name}`)
              }}</span>
              <span class="font-mono"
                >{{ window.remaining }}/{{ window.total }}</span
              >
            </div>
            <div
              class="edit-account-modal__muted mt-1 flex flex-wrap gap-x-3 gap-y-1"
            >
              <span
                >{{ t("admin.accounts.grok.runtime.source") }}:
                {{ window.source || emptyRuntimeValue }}</span
              >
              <span
                >{{ t("admin.accounts.grok.runtime.resetAt") }}:
                {{ formatRuntimeValue(window.resetAt) }}</span
              >
            </div>
          </div>
        </div>
        <div v-else class="edit-account-modal__muted text-xs">
          {{ t("admin.accounts.grok.runtime.empty") }}
        </div>
      </div>

      <div
        class="edit-account-modal__config-card edit-account-modal__config-card--compact space-y-2"
      >
        <span class="edit-account-modal__muted text-xs">{{
          t("admin.accounts.grok.runtime.syncTitle")
        }}</span>
        <div class="grid grid-cols-1 gap-2 text-xs">
          <div class="flex items-center justify-between gap-3">
            <span class="edit-account-modal__muted">{{
              t("admin.accounts.grok.runtime.lastSyncAt")
            }}</span>
            <span>{{ formatRuntimeValue(grokRuntimeState?.sync.lastSyncAt) }}</span>
          </div>
          <div class="flex items-center justify-between gap-3">
            <span class="edit-account-modal__muted">{{
              t("admin.accounts.grok.runtime.lastProbeAt")
            }}</span>
            <span>{{ formatRuntimeValue(grokRuntimeState?.sync.lastProbeAt) }}</span>
          </div>
          <div class="flex items-center justify-between gap-3">
            <span class="edit-account-modal__muted">{{
              t("admin.accounts.grok.runtime.lastProbeOkAt")
            }}</span>
            <span>{{
              formatRuntimeValue(grokRuntimeState?.sync.lastProbeOkAt)
            }}</span>
          </div>
          <div class="flex items-center justify-between gap-3">
            <span class="edit-account-modal__muted">{{
              t("admin.accounts.grok.runtime.lastProbeErrorAt")
            }}</span>
            <span>{{
              formatRuntimeValue(grokRuntimeState?.sync.lastProbeErrorAt)
            }}</span>
          </div>
          <div class="flex items-center justify-between gap-3">
            <span class="edit-account-modal__muted">{{
              t("admin.accounts.grok.runtime.probeStatus")
            }}</span>
            <span>{{ grokProbeStatusDisplay }}</span>
          </div>
          <div class="space-y-1 rounded-lg border border-dashed px-3 py-2">
            <span class="edit-account-modal__muted block">{{
              t("admin.accounts.grok.runtime.lastProbeError")
            }}</span>
            <div>{{ grokProbeErrorDisplay }}</div>
          </div>
          <div class="space-y-1 rounded-lg border border-dashed px-3 py-2">
            <span class="edit-account-modal__muted block">{{
              t("admin.accounts.grok.runtime.lastRuntimeError")
            }}</span>
            <div>{{ grokRuntimeErrorDisplay }}</div>
          </div>
          <div class="flex items-center justify-between gap-3">
            <span class="edit-account-modal__muted">{{
              t("admin.accounts.grok.runtime.lastFailAt")
            }}</span>
            <span>{{
              formatRuntimeValue(grokRuntimeState?.runtime.lastFailAt)
            }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import type { Account } from "@/types";
import { formatDateTime } from "@/utils/format";
import {
  getGrokAccountRuntime,
  getGrokProbeOutcome,
} from "@/utils/grokAccountRuntime";

const props = defineProps<{
  account: Account | null;
}>();

const { t } = useI18n();
const emptyRuntimeValue = "-";
const grokRuntimeState = computed(() => getGrokAccountRuntime(props.account));

const grokTierLabel = computed(() => {
  const tier = grokRuntimeState.value?.tier.normalized ?? "unknown";
  return t(`admin.accounts.grok.runtime.tiers.${tier}`);
});

const grokTierChipClass = computed(() => {
  switch (grokRuntimeState.value?.tier.normalized) {
    case "basic":
      return "theme-chip theme-chip--compact theme-chip--info";
    case "heavy":
      return "theme-chip theme-chip--compact theme-chip--brand-orange";
    case "super":
      return "theme-chip theme-chip--compact theme-chip--brand-purple";
    default:
      return "theme-chip theme-chip--compact theme-chip--neutral";
  }
});

const grokAuthModeLabel = computed(() => {
  const mode = grokRuntimeState.value?.authMode;
  return mode
    ? t(`admin.accounts.grok.runtime.authModes.${mode}`)
    : emptyRuntimeValue;
});

const grokTierSourceDisplay = computed(
  () => grokRuntimeState.value?.tier.source ?? emptyRuntimeValue,
);
const grokTierConfidenceDisplay = computed(() => {
  const confidence = grokRuntimeState.value?.tier.confidence;
  return confidence === null || confidence === undefined
    ? emptyRuntimeValue
    : confidence.toFixed(2);
});
const grokCapabilities = computed(
  () => grokRuntimeState.value?.capabilities.operations ?? [],
);
const grokModels = computed(
  () => grokRuntimeState.value?.capabilities.models ?? [],
);
const grokQuotaWindows = computed(
  () =>
    grokRuntimeState.value?.quotaWindows.filter((window) => window.hasSignal) ??
    [],
);
const grokProbeOutcome = computed(() =>
  getGrokProbeOutcome(grokRuntimeState.value?.sync),
);

const grokProbeStatusDisplay = computed(() => {
  const sync = grokRuntimeState.value?.sync;
  if (!sync) {
    return emptyRuntimeValue;
  }

  if (grokProbeOutcome.value === "healthy") {
    return t("admin.accounts.grok.runtime.probeHealthy");
  }
  if (grokProbeOutcome.value === "failed") {
    return t("admin.accounts.grok.runtime.probeFailed");
  }
  return emptyRuntimeValue;
});
const grokProbeErrorDisplay = computed(() => {
  const sync = grokRuntimeState.value?.sync;
  if (!sync) {
    return emptyRuntimeValue;
  }

  const code = sync.lastProbeStatusCode;
  if (grokProbeOutcome.value === "healthy") {
    return t("common.success");
  }
  if (grokProbeOutcome.value === "failed") {
    return code !== null
      ? t("admin.accounts.grok.runtime.probeFailedWithCode", { code })
      : t("admin.accounts.grok.runtime.probeFailedShort");
  }

  return emptyRuntimeValue;
});
const grokRuntimeErrorDisplay = computed(
  () => grokRuntimeState.value?.runtime.lastFailReason ?? emptyRuntimeValue,
);

function formatRuntimeValue(value: string | null | undefined) {
  if (!value) {
    return emptyRuntimeValue;
  }
  const formatted = formatDateTime(value);
  return formatted || value;
}
</script>

<style scoped>
.form-section {
  border-top: 1px solid
    color-mix(in srgb, var(--theme-page-border) 76%, transparent);
  padding-top: 1rem;
}

.edit-account-modal__muted {
  color: var(--theme-page-muted);
}

.edit-account-modal__config-card {
  border-radius: var(--theme-surface-radius);
  padding: var(--theme-markdown-block-padding);
  border: 1px solid
    color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  background: color-mix(
    in srgb,
    var(--theme-surface-soft) 90%,
    var(--theme-surface)
  );
}

.edit-account-modal__config-card--compact {
  padding: 0.75rem;
}
</style>
