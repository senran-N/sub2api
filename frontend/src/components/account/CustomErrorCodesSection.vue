<template>
  <div class="custom-error-codes-section">
    <div class="mb-3 flex items-center justify-between gap-4">
      <div>
        <label class="input-label mb-0">{{
          t("admin.accounts.customErrorCodes")
        }}</label>
        <p class="custom-error-codes-section__description mt-1 text-xs">
          {{ t("admin.accounts.customErrorCodesHint") }}
        </p>
      </div>
      <AccountModalSwitch
        :model-value="enabled"
        :aria-label="t('admin.accounts.customErrorCodes')"
        @update:model-value="emit('update:enabled', $event)"
      />
    </div>

    <div v-if="enabled" class="space-y-3">
      <div class="custom-error-codes-section__notice">
        <p class="text-xs">
          <Icon
            name="exclamationTriangle"
            size="sm"
            class="mr-1 inline"
            :stroke-width="2"
          />
          {{ t("admin.accounts.customErrorCodesWarning") }}
        </p>
      </div>

      <div class="flex flex-wrap gap-2">
        <button
          v-for="code in commonErrorCodes"
          :key="code.value"
          type="button"
          :class="getCodeChipClasses(selectedCodes.includes(code.value))"
          @click="emit('toggleCode', code.value)"
        >
          {{ code.value }} {{ code.label }}
        </button>
      </div>

      <div class="flex items-center gap-2">
        <input
          :value="inputValue ?? ''"
          type="number"
          min="100"
          max="599"
          class="input flex-1"
          :placeholder="t('admin.accounts.enterErrorCode')"
          @input="updateInputValue"
          @keyup.enter="emit('addCode')"
        />
        <button
          type="button"
          class="btn btn-secondary custom-error-codes-section__add-button"
          :aria-label="t('admin.accounts.enterErrorCode')"
          @click="emit('addCode')"
        >
          <Icon name="plus" size="sm" :stroke-width="2" />
        </button>
      </div>

      <div class="flex flex-wrap gap-1.5">
        <span
          v-for="code in sortedSelectedCodes"
          :key="code"
          class="custom-error-codes-section__selected-chip"
        >
          {{ code }}
          <button
            type="button"
            class="custom-error-codes-section__remove-button"
            :aria-label="t('common.delete')"
            @click="emit('removeCode', code)"
          >
            <Icon name="x" size="sm" :stroke-width="2" />
          </button>
        </span>
        <span
          v-if="selectedCodes.length === 0"
          class="custom-error-codes-section__description text-xs"
        >
          {{ t("admin.accounts.noneSelectedUsesDefault") }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import AccountModalSwitch from "@/components/account/AccountModalSwitch.vue";
import Icon from "@/components/icons/Icon.vue";
import { commonErrorCodes } from "@/composables/useModelWhitelist";

const props = defineProps<{
  enabled: boolean;
  selectedCodes: number[];
  inputValue: number | null;
}>();

const emit = defineEmits<{
  "update:enabled": [value: boolean];
  "update:inputValue": [value: number | null];
  toggleCode: [code: number];
  addCode: [];
  removeCode: [code: number];
}>();

const { t } = useI18n();

const sortedSelectedCodes = computed(() =>
  [...props.selectedCodes].sort((left, right) => left - right),
);

const getCodeChipClasses = (isSelected: boolean) => [
  "custom-error-codes-section__code-chip",
  isSelected && "custom-error-codes-section__code-chip--selected",
];

const updateInputValue = (event: Event) => {
  const value = (event.target as HTMLInputElement).value.trim();
  emit("update:inputValue", value === "" ? null : Number(value));
};
</script>

<style scoped>
.custom-error-codes-section {
  border-top: 1px solid
    color-mix(in srgb, var(--theme-page-border) 76%, transparent);
  padding-top: 1rem;
}

.custom-error-codes-section__description {
  color: var(--theme-page-muted);
}

.custom-error-codes-section__notice {
  border-radius: var(--theme-auth-feedback-radius);
  padding: var(--theme-auth-callback-feedback-padding);
  border-color: color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  background: color-mix(
    in srgb,
    rgb(var(--theme-warning-rgb)) 10%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--theme-warning-rgb)) 84%,
    var(--theme-page-text)
  );
}

.custom-error-codes-section__code-chip {
  border-radius: var(--theme-button-radius);
  padding: 0.375rem 0.75rem;
  background: color-mix(
    in srgb,
    var(--theme-surface-soft) 86%,
    var(--theme-surface)
  );
  color: var(--theme-page-muted);
  font-size: 0.875rem;
  font-weight: 500;
}

.custom-error-codes-section__code-chip:hover {
  background: color-mix(
    in srgb,
    var(--theme-page-border) 66%,
    var(--theme-surface)
  );
  color: var(--theme-page-text);
}

.custom-error-codes-section__code-chip--selected,
.custom-error-codes-section__selected-chip {
  background: color-mix(
    in srgb,
    rgb(var(--theme-danger-rgb)) 12%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--theme-danger-rgb)) 88%,
    var(--theme-page-text)
  );
}

.custom-error-codes-section__add-button {
  padding-inline: calc(var(--theme-button-padding-x) * 0.75);
}

.custom-error-codes-section__selected-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  border-radius: 999px;
  padding: 0.125rem 0.625rem;
}

.custom-error-codes-section__remove-button {
  display: inline-flex;
  color: inherit;
}
</style>
