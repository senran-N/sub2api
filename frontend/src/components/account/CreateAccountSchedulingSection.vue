<template>
  <div class="space-y-5">
    <div>
      <label class="input-label">{{ t("admin.accounts.proxy") }}</label>
      <ProxySelector
        :model-value="proxyId"
        :proxies="proxies"
        @update:model-value="emit('update:proxyId', $event)"
      />
    </div>

    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4 lg:grid-cols-4">
      <div>
        <label class="input-label">{{ t("admin.accounts.concurrency") }}</label>
        <input
          :value="concurrency"
          type="number"
          min="1"
          class="input"
          @input="updateConcurrency"
        />
      </div>
      <div>
        <label class="input-label">{{ t("admin.accounts.loadFactor") }}</label>
        <input
          :value="loadFactor ?? ''"
          type="number"
          min="1"
          class="input"
          :placeholder="String(concurrency || 1)"
          @input="updateLoadFactor"
        />
        <p class="input-hint">{{ t("admin.accounts.loadFactorHint") }}</p>
      </div>
      <div>
        <label class="input-label">{{ t("admin.accounts.priority") }}</label>
        <input
          :value="priority"
          type="number"
          min="1"
          class="input"
          data-tour="account-form-priority"
          @input="emitNumber('update:priority', $event)"
        />
        <p class="input-hint">{{ t("admin.accounts.priorityHint") }}</p>
      </div>
      <div>
        <label class="input-label">{{
          t("admin.accounts.billingRateMultiplier")
        }}</label>
        <input
          :value="rateMultiplier"
          type="number"
          min="0"
          step="0.001"
          class="input"
          @input="emitNumber('update:rateMultiplier', $event)"
        />
        <p class="input-hint">
          {{ t("admin.accounts.billingRateMultiplierHint") }}
        </p>
      </div>
    </div>

    <div class="create-account-scheduling-section__expires">
      <label class="input-label">{{ t("admin.accounts.expiresAt") }}</label>
      <input
        :value="expiresAt"
        type="datetime-local"
        class="input"
        @input="emit('update:expiresAt', ($event.target as HTMLInputElement).value)"
      />
      <p class="input-hint">{{ t("admin.accounts.expiresAtHint") }}</p>
    </div>

    <div class="create-account-scheduling-section__extra-options">
      <div v-if="platform === 'antigravity'" class="flex items-center gap-2">
        <label class="create-account-scheduling-section__checkbox">
          <input
            :checked="mixedScheduling"
            type="checkbox"
            class="create-account-scheduling-section__checkbox-input"
            @change="
              emit(
                'update:mixedScheduling',
                ($event.target as HTMLInputElement).checked,
              )
            "
          />
          <span class="create-account-scheduling-section__title text-sm">
            {{ t("admin.accounts.mixedScheduling") }}
          </span>
        </label>
        <div class="group relative">
          <span class="create-account-scheduling-section__tooltip-trigger">
            ?
          </span>
          <div class="create-account-scheduling-section__tooltip-panel">
            {{ t("admin.accounts.mixedSchedulingTooltip") }}
            <div class="create-account-scheduling-section__tooltip-arrow"></div>
          </div>
        </div>
      </div>
      <div
        v-if="platform === 'antigravity'"
        class="mt-3 flex items-center gap-2"
      >
        <label class="create-account-scheduling-section__checkbox">
          <input
            :checked="allowOverages"
            type="checkbox"
            class="create-account-scheduling-section__checkbox-input"
            @change="
              emit(
                'update:allowOverages',
                ($event.target as HTMLInputElement).checked,
              )
            "
          />
          <span class="create-account-scheduling-section__title text-sm">
            {{ t("admin.accounts.allowOverages") }}
          </span>
        </label>
        <div class="group relative">
          <span class="create-account-scheduling-section__tooltip-trigger">
            ?
          </span>
          <div class="create-account-scheduling-section__tooltip-panel">
            {{ t("admin.accounts.allowOveragesTooltip") }}
            <div class="create-account-scheduling-section__tooltip-arrow"></div>
          </div>
        </div>
      </div>

      <GroupSelector
        v-if="!simpleMode"
        :model-value="groupIds"
        :groups="groups"
        :platform="platform"
        :mixed-scheduling="mixedScheduling"
        data-tour="account-form-groups"
        @update:model-value="emit('update:groupIds', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import GroupSelector from "@/components/common/GroupSelector.vue";
import ProxySelector from "@/components/common/ProxySelector.vue";
import type {
  AccountPlatform,
  AdminGroup,
  Proxy,
} from "@/types";

type NumberUpdateEvent = "update:priority" | "update:rateMultiplier";

defineProps<{
  proxyId: number | null;
  proxies: Proxy[];
  concurrency: number;
  loadFactor: number | null;
  priority: number;
  rateMultiplier: number;
  expiresAt: string;
  platform: AccountPlatform;
  mixedScheduling: boolean;
  allowOverages: boolean;
  groupIds: number[];
  groups: AdminGroup[];
  simpleMode: boolean;
}>();

const emit = defineEmits<{
  "update:proxyId": [value: number | null];
  "update:concurrency": [value: number];
  "update:loadFactor": [value: number | null];
  "update:priority": [value: number];
  "update:rateMultiplier": [value: number];
  "update:expiresAt": [value: string];
  "update:mixedScheduling": [value: boolean];
  "update:allowOverages": [value: boolean];
  "update:groupIds": [value: number[]];
}>();

const { t } = useI18n();

const readNumber = (event: Event) =>
  Number((event.target as HTMLInputElement).value);

const updateConcurrency = (event: Event) => {
  emit("update:concurrency", Math.max(1, readNumber(event) || 1));
};

const updateLoadFactor = (event: Event) => {
  const value = readNumber(event);
  emit("update:loadFactor", value && value >= 1 ? value : null);
};

const emitNumber = (eventName: NumberUpdateEvent, event: Event) => {
  const value = readNumber(event);
  if (eventName === "update:priority") {
    emit("update:priority", value);
    return;
  }
  emit("update:rateMultiplier", value);
};
</script>

<style scoped>
.create-account-scheduling-section__expires,
.create-account-scheduling-section__extra-options {
  border-top: 1px solid
    color-mix(in srgb, var(--theme-page-border) 76%, transparent);
  padding-top: 1rem;
}

.create-account-scheduling-section__checkbox {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}

.create-account-scheduling-section__checkbox-input {
  accent-color: var(--theme-accent);
  border: 1px solid var(--theme-input-border);
  border-radius: 0.375rem;
}

.create-account-scheduling-section__title {
  color: var(--theme-page-text);
}

.create-account-scheduling-section__tooltip-trigger {
  display: inline-flex;
  height: 1rem;
  width: 1rem;
  cursor: help;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: color-mix(
    in srgb,
    var(--theme-page-border) 82%,
    var(--theme-surface)
  );
  color: var(--theme-page-muted);
  font-size: 0.75rem;
}

.create-account-scheduling-section__tooltip-panel {
  position: absolute;
  bottom: 100%;
  left: 50%;
  z-index: 10;
  margin-bottom: 0.5rem;
  width: 16rem;
  transform: translateX(-50%);
  border-radius: var(--theme-button-radius);
  background: var(--theme-surface-contrast);
  padding: 0.5rem 0.75rem;
  color: var(--theme-filled-text);
  font-size: 0.75rem;
  line-height: 1.4;
  opacity: 0;
  pointer-events: none;
}

.group:hover .create-account-scheduling-section__tooltip-panel {
  opacity: 1;
}

.create-account-scheduling-section__tooltip-arrow {
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  border: 4px solid transparent;
  border-top-color: var(--theme-surface-contrast);
}
</style>
