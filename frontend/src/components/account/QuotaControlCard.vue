<template>
  <div class="quota-control-card">
    <div class="mb-3 flex items-center justify-between gap-4">
      <div>
        <label class="input-label mb-0">{{ t(titleKey) }}</label>
        <p class="quota-control-card__description mt-1 text-xs">
          {{ t(hintKey) }}
        </p>
      </div>
      <AccountModalSwitch
        :model-value="enabled"
        :aria-label="t(titleKey)"
        @update:model-value="emit('update:enabled', $event)"
      />
    </div>
    <slot v-if="enabled" />
    <slot name="footer" />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import AccountModalSwitch from "@/components/account/AccountModalSwitch.vue";

defineProps<{
  enabled: boolean;
  titleKey: string;
  hintKey: string;
}>();

const emit = defineEmits<{
  "update:enabled": [value: boolean];
}>();

const { t } = useI18n();
</script>

<style scoped>
.quota-control-card {
  border: 1px solid
    color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  border-radius: var(--theme-surface-radius);
  background: color-mix(
    in srgb,
    var(--theme-surface-soft) 90%,
    var(--theme-surface)
  );
  padding: var(--theme-markdown-block-padding);
}

.quota-control-card__description {
  color: var(--theme-page-muted);
}
</style>
