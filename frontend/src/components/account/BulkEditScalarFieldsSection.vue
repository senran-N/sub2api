<template>
  <div class="form-section grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4 lg:grid-cols-4">
    <BulkEditNumberField
      id="bulk-edit-concurrency"
      :enabled="enableConcurrency"
      :value="concurrency"
      label-key="admin.accounts.concurrency"
      :min="1"
      @update:enabled="emit('update:enableConcurrency', $event)"
      @update:value="emit('update:concurrency', normalizeConcurrency($event))"
    />

    <BulkEditNumberField
      id="bulk-edit-load-factor"
      :enabled="enableLoadFactor"
      :value="loadFactor"
      label-key="admin.accounts.loadFactor"
      hint-key="admin.accounts.loadFactorHint"
      :min="1"
      @update:enabled="emit('update:enableLoadFactor', $event)"
      @update:value="emit('update:loadFactor', normalizeLoadFactor($event))"
    />

    <BulkEditNumberField
      id="bulk-edit-priority"
      :enabled="enablePriority"
      :value="priority"
      label-key="admin.accounts.priority"
      :min="1"
      @update:enabled="emit('update:enablePriority', $event)"
      @update:value="emit('update:priority', $event ?? 1)"
    />

    <BulkEditNumberField
      id="bulk-edit-rate-multiplier"
      :enabled="enableRateMultiplier"
      :value="rateMultiplier"
      label-key="admin.accounts.billingRateMultiplier"
      hint-key="admin.accounts.billingRateMultiplierHint"
      :min="0"
      step="0.01"
      @update:enabled="emit('update:enableRateMultiplier', $event)"
      @update:value="emit('update:rateMultiplier', $event ?? 1)"
    />
  </div>
</template>

<script setup lang="ts">
import BulkEditNumberField from "@/components/account/BulkEditNumberField.vue";

defineProps<{
  concurrency: number;
  enableConcurrency: boolean;
  enableLoadFactor: boolean;
  enablePriority: boolean;
  enableRateMultiplier: boolean;
  loadFactor: number | null;
  priority: number;
  rateMultiplier: number;
}>();

const emit = defineEmits<{
  "update:enableConcurrency": [value: boolean];
  "update:enableLoadFactor": [value: boolean];
  "update:enablePriority": [value: boolean];
  "update:enableRateMultiplier": [value: boolean];
  "update:concurrency": [value: number];
  "update:loadFactor": [value: number | null];
  "update:priority": [value: number];
  "update:rateMultiplier": [value: number];
}>();

const normalizeConcurrency = (value: number | null) => Math.max(1, value || 1);
const normalizeLoadFactor = (value: number | null) =>
  value != null && value >= 1 ? value : null;
</script>

<style scoped>
.form-section {
  border-top: 1px solid
    color-mix(in srgb, var(--theme-page-border) 76%, transparent);
  padding-top: 1rem;
}
</style>
