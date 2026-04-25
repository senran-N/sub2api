<template>
  <div class="temp-unsched-rules-section">
    <div class="mb-3 flex items-center justify-between gap-4">
      <div>
        <label class="input-label mb-0">{{
          t("admin.accounts.tempUnschedulable.title")
        }}</label>
        <p class="temp-unsched-rules-section__description mt-1 text-xs">
          {{ t("admin.accounts.tempUnschedulable.hint") }}
        </p>
      </div>
      <AccountModalSwitch
        :model-value="enabled"
        :aria-label="t('admin.accounts.tempUnschedulable.title')"
        @update:model-value="emit('update:enabled', $event)"
      />
    </div>

    <div v-if="enabled" class="space-y-3">
      <div class="temp-unsched-rules-section__notice">
        <p class="text-xs">
          <Icon
            name="exclamationTriangle"
            size="sm"
            class="mr-1 inline"
            :stroke-width="2"
          />
          {{ t("admin.accounts.tempUnschedulable.notice") }}
        </p>
      </div>

      <div class="flex flex-wrap gap-2">
        <button
          v-for="preset in presets"
          :key="preset.label"
          type="button"
          class="temp-unsched-rules-section__tag-button"
          @click="emit('addRule', preset.rule)"
        >
          + {{ preset.label }}
        </button>
      </div>

      <div v-if="rules.length > 0" class="space-y-3">
        <div
          v-for="(rule, index) in rules"
          :key="ruleKey(rule)"
          class="temp-unsched-rules-section__rule-card"
        >
          <div class="mb-2 flex items-center justify-between">
            <span class="temp-unsched-rules-section__rule-index">
              {{
                t("admin.accounts.tempUnschedulable.ruleIndex", {
                  index: index + 1,
                })
              }}
            </span>
            <div class="flex items-center gap-2">
              <button
                type="button"
                :disabled="index === 0"
                class="temp-unsched-rules-section__icon-button"
                :aria-label="t('common.moveUp')"
                @click="emit('moveRule', index, -1)"
              >
                <Icon name="chevronUp" size="sm" :stroke-width="2" />
              </button>
              <button
                type="button"
                :disabled="index === rules.length - 1"
                class="temp-unsched-rules-section__icon-button"
                :aria-label="t('common.moveDown')"
                @click="emit('moveRule', index, 1)"
              >
                <Icon name="chevronDown" size="sm" :stroke-width="2" />
              </button>
              <button
                type="button"
                class="temp-unsched-rules-section__icon-button temp-unsched-rules-section__icon-button--danger"
                :aria-label="t('common.delete')"
                @click="emit('removeRule', index)"
              >
                <Icon name="x" size="sm" :stroke-width="2" />
              </button>
            </div>
          </div>

          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label class="input-label">{{
                t("admin.accounts.tempUnschedulable.errorCode")
              }}</label>
              <input
                :value="rule.error_code ?? ''"
                type="number"
                min="100"
                max="599"
                class="input"
                :placeholder="
                  t('admin.accounts.tempUnschedulable.errorCodePlaceholder')
                "
                @input="updateNumberRule(index, 'error_code', $event)"
              />
            </div>
            <div>
              <label class="input-label">{{
                t("admin.accounts.tempUnschedulable.durationMinutes")
              }}</label>
              <input
                :value="rule.duration_minutes ?? ''"
                type="number"
                min="1"
                class="input"
                :placeholder="
                  t('admin.accounts.tempUnschedulable.durationPlaceholder')
                "
                @input="updateNumberRule(index, 'duration_minutes', $event)"
              />
            </div>
            <div class="sm:col-span-2">
              <label class="input-label">{{
                t("admin.accounts.tempUnschedulable.keywords")
              }}</label>
              <input
                :value="rule.keywords"
                type="text"
                class="input"
                :placeholder="
                  t('admin.accounts.tempUnschedulable.keywordsPlaceholder')
                "
                @input="updateStringRule(index, 'keywords', $event)"
              />
              <p class="input-hint">
                {{ t("admin.accounts.tempUnschedulable.keywordsHint") }}
              </p>
            </div>
            <div class="sm:col-span-2">
              <label class="input-label">{{
                t("admin.accounts.tempUnschedulable.description")
              }}</label>
              <input
                :value="rule.description"
                type="text"
                class="input"
                :placeholder="
                  t('admin.accounts.tempUnschedulable.descriptionPlaceholder')
                "
                @input="updateStringRule(index, 'description', $event)"
              />
            </div>
          </div>
        </div>
      </div>

      <button
        type="button"
        class="temp-unsched-rules-section__dashed-action w-full"
        @click="emit('addRule')"
      >
        <Icon name="plus" size="sm" class="mr-1 inline" :stroke-width="2" />
        {{ t("admin.accounts.tempUnschedulable.addRule") }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import AccountModalSwitch from "@/components/account/AccountModalSwitch.vue";
import Icon from "@/components/icons/Icon.vue";
import type { TempUnschedRuleForm } from "@/components/account/credentialsBuilder";

type TempUnschedPreset = {
  label: string;
  rule: TempUnschedRuleForm;
};

defineProps<{
  enabled: boolean;
  presets: TempUnschedPreset[];
  rules: TempUnschedRuleForm[];
  ruleKey: (rule: TempUnschedRuleForm) => string;
}>();

const emit = defineEmits<{
  "update:enabled": [value: boolean];
  addRule: [preset?: TempUnschedRuleForm];
  removeRule: [index: number];
  moveRule: [index: number, direction: number];
  updateRule: [
    index: number,
    field: keyof TempUnschedRuleForm,
    value: TempUnschedRuleForm[keyof TempUnschedRuleForm],
  ];
}>();

const { t } = useI18n();

const readInputValue = (event: Event) =>
  (event.target as HTMLInputElement).value;

const updateNumberRule = (
  index: number,
  field: "error_code" | "duration_minutes",
  event: Event,
) => {
  const value = readInputValue(event).trim();
  emit("updateRule", index, field, value === "" ? null : Number(value));
};

const updateStringRule = (
  index: number,
  field: "keywords" | "description",
  event: Event,
) => {
  emit("updateRule", index, field, readInputValue(event));
};
</script>

<style scoped>
.temp-unsched-rules-section {
  border-top: 1px solid
    color-mix(in srgb, var(--theme-page-border) 76%, transparent);
  padding-top: 1rem;
}

.temp-unsched-rules-section__description {
  color: var(--theme-page-muted);
}

.temp-unsched-rules-section__notice {
  border-radius: var(--theme-auth-feedback-radius);
  padding: var(--theme-auth-callback-feedback-padding);
  border-color: color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  background: color-mix(
    in srgb,
    rgb(var(--theme-info-rgb)) 10%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--theme-info-rgb)) 84%,
    var(--theme-page-text)
  );
}

.temp-unsched-rules-section__tag-button {
  border-radius: calc(var(--theme-button-radius) + 2px);
  background: color-mix(
    in srgb,
    var(--theme-surface-soft) 86%,
    var(--theme-surface)
  );
  color: var(--theme-page-muted);
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.4rem 0.75rem;
}

.temp-unsched-rules-section__tag-button:hover,
.temp-unsched-rules-section__tag-button:focus-visible {
  background: color-mix(
    in srgb,
    var(--theme-page-border) 72%,
    var(--theme-surface)
  );
  color: var(--theme-page-text);
  outline: none;
}

.temp-unsched-rules-section__rule-card {
  border: 1px solid
    color-mix(in srgb, var(--theme-card-border) 80%, transparent);
  border-radius: calc(var(--theme-button-radius) + 2px);
  background: var(--theme-surface);
  padding: 0.75rem;
}

.temp-unsched-rules-section__rule-index {
  color: var(--theme-page-muted);
  font-size: 0.75rem;
  font-weight: 600;
}

.temp-unsched-rules-section__icon-button {
  border-radius: calc(var(--theme-button-radius) - 4px);
  color: var(--theme-page-muted);
  padding: 0.25rem;
}

.temp-unsched-rules-section__icon-button:hover,
.temp-unsched-rules-section__icon-button:focus-visible {
  background: color-mix(
    in srgb,
    var(--theme-button-ghost-hover-bg) 90%,
    transparent
  );
  color: var(--theme-page-text);
  outline: none;
}

.temp-unsched-rules-section__icon-button:disabled {
  cursor: not-allowed;
  opacity: 0.4;
}

.temp-unsched-rules-section__icon-button--danger {
  color: rgb(var(--theme-danger-rgb));
}

.temp-unsched-rules-section__icon-button--danger:hover,
.temp-unsched-rules-section__icon-button--danger:focus-visible {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, transparent);
  color: color-mix(
    in srgb,
    rgb(var(--theme-danger-rgb)) 88%,
    var(--theme-page-text)
  );
}

.temp-unsched-rules-section__dashed-action {
  border: 2px dashed
    color-mix(in srgb, var(--theme-card-border) 90%, transparent);
  border-radius: calc(var(--theme-button-radius) + 2px);
  color: var(--theme-page-muted);
  font-size: 0.875rem;
  padding: 0.625rem 1rem;
}

.temp-unsched-rules-section__dashed-action:hover,
.temp-unsched-rules-section__dashed-action:focus-visible {
  border-color: color-mix(
    in srgb,
    var(--theme-accent) 32%,
    var(--theme-card-border)
  );
  background: color-mix(
    in srgb,
    var(--theme-accent-soft) 46%,
    var(--theme-surface)
  );
  color: var(--theme-page-text);
  outline: none;
}
</style>
