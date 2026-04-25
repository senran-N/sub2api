<template>
  <div class="form-section space-y-4">
    <div class="mb-3">
      <h3 class="input-label mb-0 text-base font-semibold">
        {{ t("admin.accounts.quotaControl.title") }}
      </h3>
      <p class="anthropic-quota-controls-section__description mt-1 text-xs">
        {{ t("admin.accounts.quotaControl.hint") }}
      </p>
    </div>

    <WindowCostControlSection
      v-model:enabled="windowCostEnabled"
      v-model:limit="windowCostLimit"
      v-model:sticky-reserve="windowCostStickyReserve"
    />

    <SessionLimitControlSection
      v-model:enabled="sessionLimitEnabled"
      v-model:max-sessions="maxSessions"
      v-model:idle-timeout="sessionIdleTimeout"
    />

    <RpmLimitControlSection
      v-model:enabled="rpmLimitEnabled"
      v-model:base-rpm="baseRpm"
      v-model:strategy="rpmStrategy"
      v-model:sticky-buffer="rpmStickyBuffer"
      v-model:user-msg-queue-mode="userMsgQueueMode"
      :user-msg-queue-mode-options="userMsgQueueModeOptions"
    />

    <TlsFingerprintControlSection
      v-model:enabled="tlsFingerprintEnabled"
      v-model:profile-id="tlsFingerprintProfileId"
      :profiles="tlsFingerprintProfiles"
    />

    <SessionIdMaskingControlSection
      v-model:enabled="sessionIdMaskingEnabled"
    />

    <CacheTtlOverrideSection
      v-model:enabled="cacheTTLOverrideEnabled"
      v-model:target="cacheTTLOverrideTarget"
    />

    <CustomBaseUrlControlSection
      v-model:enabled="customBaseUrlEnabled"
      v-model:base-url="customBaseUrl"
    />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import CacheTtlOverrideSection from "@/components/account/CacheTtlOverrideSection.vue";
import CustomBaseUrlControlSection from "@/components/account/CustomBaseUrlControlSection.vue";
import RpmLimitControlSection from "@/components/account/RpmLimitControlSection.vue";
import SessionIdMaskingControlSection from "@/components/account/SessionIdMaskingControlSection.vue";
import SessionLimitControlSection from "@/components/account/SessionLimitControlSection.vue";
import TlsFingerprintControlSection from "@/components/account/TlsFingerprintControlSection.vue";
import WindowCostControlSection from "@/components/account/WindowCostControlSection.vue";

type RpmStrategy = "tiered" | "sticky_exempt";

defineProps<{
  tlsFingerprintProfiles: Array<{ id: number; name: string }>;
  userMsgQueueModeOptions: Array<{ value: string; label: string }>;
}>();

const windowCostEnabled = defineModel<boolean>("windowCostEnabled", {
  required: true,
});
const windowCostLimit = defineModel<number | null>("windowCostLimit", {
  required: true,
});
const windowCostStickyReserve = defineModel<number | null>(
  "windowCostStickyReserve",
  { required: true },
);
const sessionLimitEnabled = defineModel<boolean>("sessionLimitEnabled", {
  required: true,
});
const maxSessions = defineModel<number | null>("maxSessions", {
  required: true,
});
const sessionIdleTimeout = defineModel<number | null>("sessionIdleTimeout", {
  required: true,
});
const rpmLimitEnabled = defineModel<boolean>("rpmLimitEnabled", {
  required: true,
});
const baseRpm = defineModel<number | null>("baseRpm", { required: true });
const rpmStrategy = defineModel<RpmStrategy>("rpmStrategy", { required: true });
const rpmStickyBuffer = defineModel<number | null>("rpmStickyBuffer", {
  required: true,
});
const userMsgQueueMode = defineModel<string>("userMsgQueueMode", {
  required: true,
});
const tlsFingerprintEnabled = defineModel<boolean>("tlsFingerprintEnabled", {
  required: true,
});
const tlsFingerprintProfileId = defineModel<number | null>(
  "tlsFingerprintProfileId",
  { required: true },
);
const sessionIdMaskingEnabled = defineModel<boolean>("sessionIdMaskingEnabled", {
  required: true,
});
const cacheTTLOverrideEnabled = defineModel<boolean>("cacheTtlOverrideEnabled", {
  required: true,
});
const cacheTTLOverrideTarget = defineModel<string>("cacheTtlOverrideTarget", {
  required: true,
});
const customBaseUrlEnabled = defineModel<boolean>("customBaseUrlEnabled", {
  required: true,
});
const customBaseUrl = defineModel<string>("customBaseUrl", { required: true });

const { t } = useI18n();
</script>

<style scoped>
.form-section {
  border-top: 1px solid
    color-mix(in srgb, var(--theme-page-border) 76%, transparent);
  padding-top: 1rem;
}

.anthropic-quota-controls-section__description {
  color: var(--theme-page-muted);
}
</style>
