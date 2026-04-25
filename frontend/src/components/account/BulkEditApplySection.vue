<template>
  <div class="form-section">
    <div class="mb-3 flex items-center justify-between gap-4">
      <div class="flex-1 pr-4">
        <label :id="labelId" class="input-label mb-0" :for="checkboxId">
          {{ t(labelKey) }}
        </label>
        <p v-if="hintKey" class="bulk-edit-apply-section__description mt-1 text-xs">
          {{ t(hintKey) }}
        </p>
      </div>
      <input
        :id="checkboxId"
        :checked="enabled"
        type="checkbox"
        :aria-controls="bodyId"
        class="bulk-edit-apply-section__checkbox rounded"
        @change="
          emit(
            'update:enabled',
            ($event.target as HTMLInputElement).checked,
          )
        "
      />
    </div>
    <div
      :id="bodyId"
      :class="!enabled && 'pointer-events-none opacity-50'"
      role="group"
      :aria-labelledby="labelId"
    >
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

const props = defineProps<{
  enabled: boolean;
  hintKey?: string;
  id: string;
  labelKey: string;
}>();

const emit = defineEmits<{
  "update:enabled": [value: boolean];
}>();

const { t } = useI18n();

const checkboxId = computed(() => `${props.id}-enabled`);
const labelId = computed(() => `${props.id}-label`);
const bodyId = computed(() => `${props.id}-body`);
</script>

<style scoped>
.form-section {
  border-top: 1px solid
    color-mix(in srgb, var(--theme-page-border) 76%, transparent);
  padding-top: 1rem;
}

.bulk-edit-apply-section__description {
  color: var(--theme-page-muted);
}

.bulk-edit-apply-section__checkbox {
  border-color: color-mix(in srgb, var(--theme-input-border) 82%, transparent);
  color: var(--theme-accent);
}

.bulk-edit-apply-section__checkbox:focus {
  outline: none;
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--theme-accent) 18%, transparent);
}
</style>
