<template>
  <div class="space-y-5">
    <div>
      <label class="input-label">{{ t("common.name") }}</label>
      <input
        :value="name"
        type="text"
        required
        class="input"
        data-tour="edit-account-form-name"
        @input="emitInputValue('update:name', $event)"
      />
    </div>

    <div>
      <label class="input-label">{{ t("admin.accounts.notes") }}</label>
      <textarea
        :value="notes"
        rows="3"
        class="input"
        :placeholder="t('admin.accounts.notesPlaceholder')"
        @input="emitTextareaValue"
      ></textarea>
      <p class="input-hint">{{ t("admin.accounts.notesHint") }}</p>
    </div>

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

    <div class="edit-account-core-fields-section__separated">
      <label class="input-label">{{ t("admin.accounts.expiresAt") }}</label>
      <input
        :value="expiresAt"
        type="datetime-local"
        class="input"
        @input="emitInputValue('update:expiresAt', $event)"
      />
      <p class="input-hint">{{ t("admin.accounts.expiresAtHint") }}</p>
    </div>

    <div class="edit-account-core-fields-section__separated">
      <div>
        <label class="input-label">{{ t("common.status") }}</label>
        <Select
          :model-value="status"
          :options="statusOptions"
          @update:model-value="updateStatus"
        />
      </div>

      <template v-if="platform === 'antigravity'">
        <div class="flex items-center gap-2">
          <label
            class="edit-account-core-fields-section__checkbox edit-account-core-fields-section__checkbox--readonly"
          >
            <input
              :checked="mixedScheduling"
              type="checkbox"
              disabled
              class="edit-account-core-fields-section__checkbox-input cursor-not-allowed"
            />
            <span class="edit-account-core-fields-section__title text-sm font-medium">
              {{ t("admin.accounts.mixedScheduling") }}
            </span>
          </label>
          <div class="group relative">
            <span class="edit-account-core-fields-section__tooltip-trigger">
              ?
            </span>
            <div class="edit-account-core-fields-section__tooltip-panel">
              {{ t("admin.accounts.mixedSchedulingTooltip") }}
              <div class="edit-account-core-fields-section__tooltip-arrow"></div>
            </div>
          </div>
        </div>

        <div class="mt-3 flex items-center gap-2">
          <label class="edit-account-core-fields-section__checkbox">
            <input
              :checked="allowOverages"
              type="checkbox"
              class="edit-account-core-fields-section__checkbox-input"
              @change="
                emit(
                  'update:allowOverages',
                  ($event.target as HTMLInputElement).checked,
                )
              "
            />
            <span class="edit-account-core-fields-section__title text-sm font-medium">
              {{ t("admin.accounts.allowOverages") }}
            </span>
          </label>
          <div class="group relative">
            <span class="edit-account-core-fields-section__tooltip-trigger">
              ?
            </span>
            <div class="edit-account-core-fields-section__tooltip-panel">
              {{ t("admin.accounts.allowOveragesTooltip") }}
              <div class="edit-account-core-fields-section__tooltip-arrow"></div>
            </div>
          </div>
        </div>
      </template>
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
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import GroupSelector from "@/components/common/GroupSelector.vue";
import ProxySelector from "@/components/common/ProxySelector.vue";
import Select from "@/components/common/Select.vue";
import type { AccountPlatform, AdminGroup, Proxy } from "@/types";

type EditAccountStatus = "active" | "inactive" | "error";
type StringUpdateEvent = "update:name" | "update:expiresAt";
type NumberUpdateEvent = "update:priority" | "update:rateMultiplier";

defineProps<{
  allowOverages: boolean;
  concurrency: number;
  expiresAt: string;
  groupIds: number[];
  groups: AdminGroup[];
  loadFactor: number | null;
  mixedScheduling: boolean;
  name: string;
  notes: string;
  platform: AccountPlatform;
  priority: number;
  proxies: Proxy[];
  proxyId: number | null;
  rateMultiplier: number;
  simpleMode: boolean;
  status: EditAccountStatus;
  statusOptions: Array<{ value: EditAccountStatus; label: string }>;
}>();

const emit = defineEmits<{
  "update:name": [value: string];
  "update:notes": [value: string];
  "update:proxyId": [value: number | null];
  "update:concurrency": [value: number];
  "update:loadFactor": [value: number | null];
  "update:priority": [value: number];
  "update:rateMultiplier": [value: number];
  "update:expiresAt": [value: string];
  "update:status": [value: EditAccountStatus];
  "update:allowOverages": [value: boolean];
  "update:groupIds": [value: number[]];
}>();

const { t } = useI18n();

const readNumber = (event: Event) =>
  Number((event.target as HTMLInputElement).value);

const emitInputValue = (eventName: StringUpdateEvent, event: Event) => {
  const value = (event.target as HTMLInputElement).value;
  if (eventName === "update:name") {
    emit("update:name", value);
    return;
  }
  emit("update:expiresAt", value);
};

const emitTextareaValue = (event: Event) => {
  emit("update:notes", (event.target as HTMLTextAreaElement).value);
};

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

const updateStatus = (value: string | number | boolean | null) => {
  if (value === "active" || value === "inactive" || value === "error") {
    emit("update:status", value);
  }
};
</script>

<style scoped>
.edit-account-core-fields-section__separated {
  border-top: 1px solid
    color-mix(in srgb, var(--theme-page-border) 76%, transparent);
  padding-top: 1rem;
}

.edit-account-core-fields-section__checkbox {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}

.edit-account-core-fields-section__checkbox--readonly {
  cursor: not-allowed;
  opacity: 0.6;
}

.edit-account-core-fields-section__checkbox-input {
  accent-color: var(--theme-accent);
  border: 1px solid var(--theme-input-border);
  border-radius: 0.375rem;
}

.edit-account-core-fields-section__title {
  color: var(--theme-page-text);
}

.edit-account-core-fields-section__tooltip-trigger {
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

.edit-account-core-fields-section__tooltip-panel {
  position: absolute;
  left: 0;
  top: 100%;
  z-index: 100;
  margin-top: 0.375rem;
  width: 18rem;
  border-radius: var(--theme-tooltip-radius);
  background: var(--theme-surface-contrast);
  padding: 0.5rem 0.75rem;
  color: var(--theme-filled-text);
  font-size: 0.75rem;
  line-height: 1.4;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.2s ease;
}

.group:hover .edit-account-core-fields-section__tooltip-panel {
  opacity: 1;
}

.edit-account-core-fields-section__tooltip-arrow {
  position: absolute;
  bottom: 100%;
  left: 0.75rem;
  border: 4px solid transparent;
  border-bottom-color: var(--theme-surface-contrast);
}
</style>
