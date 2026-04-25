<template>
  <div>
    <div class="mb-3 flex items-center justify-between">
      <label :id="labelId" class="input-label mb-0" :for="enabledId">
        {{ t(labelKey) }}
      </label>
      <input
        :id="enabledId"
        :checked="enabled"
        type="checkbox"
        :aria-controls="id"
        class="bulk-edit-number-field__checkbox rounded"
        @change="
          emit(
            'update:enabled',
            ($event.target as HTMLInputElement).checked,
          )
        "
      />
    </div>
    <input
      :id="id"
      :value="value ?? ''"
      type="number"
      :min="min"
      :step="step"
      :disabled="!enabled"
      class="input"
      :class="!enabled && 'cursor-not-allowed opacity-50'"
      :aria-labelledby="labelId"
      @input="updateValue"
    />
    <p v-if="hintKey" class="input-hint">{{ t(hintKey) }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

const props = withDefaults(
  defineProps<{
    enabled: boolean;
    hintKey?: string;
    id: string;
    labelKey: string;
    min?: number;
    step?: number | string;
    value: number | null;
  }>(),
  {
    hintKey: "",
    min: undefined,
    step: undefined,
  },
);

const emit = defineEmits<{
  "update:enabled": [value: boolean];
  "update:value": [value: number | null];
}>();

const { t } = useI18n();

const enabledId = computed(() => `${props.id}-enabled`);
const labelId = computed(() => `${props.id}-label`);

const updateValue = (event: Event) => {
  const value = (event.target as HTMLInputElement).value.trim();
  emit("update:value", value === "" ? null : Number(value));
};
</script>

<style scoped>
.bulk-edit-number-field__checkbox {
  border-color: color-mix(in srgb, var(--theme-input-border) 82%, transparent);
  color: var(--theme-accent);
}

.bulk-edit-number-field__checkbox:focus {
  outline: none;
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--theme-accent) 18%, transparent);
}
</style>
