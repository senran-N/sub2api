<template>
  <div>
    <label class="input-label">{{ t("admin.accounts.addMethod") }}</label>
    <div class="mt-2 flex gap-4">
      <label :class="getRadioOptionClasses(modelValue === 'oauth')">
        <input
          :checked="modelValue === 'oauth'"
          type="radio"
          value="oauth"
          class="anthropic-add-method-section__radio-input"
          @change="emit('update:modelValue', 'oauth')"
        />
        <span class="anthropic-add-method-section__title text-sm">{{
          t("admin.accounts.types.oauth")
        }}</span>
      </label>
      <label :class="getRadioOptionClasses(modelValue === 'setup-token')">
        <input
          :checked="modelValue === 'setup-token'"
          type="radio"
          value="setup-token"
          class="anthropic-add-method-section__radio-input"
          @change="emit('update:modelValue', 'setup-token')"
        />
        <span class="anthropic-add-method-section__title text-sm">{{
          t("admin.accounts.setupTokenLongLived")
        }}</span>
      </label>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import type { AddMethod } from "@/composables/useAccountOAuth";

defineProps<{
  modelValue: AddMethod;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: AddMethod];
}>();

const { t } = useI18n();

const getRadioOptionClasses = (isSelected: boolean) => [
  "anthropic-add-method-section__radio-option",
  isSelected && "anthropic-add-method-section__radio-option--active",
];
</script>

<style scoped>
.anthropic-add-method-section__radio-option {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  border: 1px solid
    color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  border-radius: calc(var(--theme-button-radius) + 2px);
  background: color-mix(
    in srgb,
    var(--theme-surface-soft) 84%,
    var(--theme-surface)
  );
  padding: 0.55rem 0.8rem;
}

.anthropic-add-method-section__radio-option--active {
  border-color: color-mix(
    in srgb,
    var(--theme-accent) 64%,
    var(--theme-card-border)
  );
  background: color-mix(
    in srgb,
    var(--theme-accent-soft) 78%,
    var(--theme-surface)
  );
}

.anthropic-add-method-section__radio-input {
  accent-color: var(--theme-accent);
}

.anthropic-add-method-section__title {
  color: var(--theme-page-text);
}
</style>
