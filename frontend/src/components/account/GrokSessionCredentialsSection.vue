<template>
  <div class="space-y-4">
    <GrokSessionBatchImportPanel
      :mode="mode"
      :single-token="singleToken"
      :batch-input="batchInput"
      :dry-run="dryRun"
      :test-after-create="testAfterCreate"
      :result="result"
      :submitting="submitting"
      @update:mode="emit('update:mode', $event)"
      @update:single-token="emit('update:singleToken', $event)"
      @update:batch-input="emit('update:batchInput', $event)"
      @update:dry-run="emit('update:dryRun', $event)"
      @update:test-after-create="emit('update:testAfterCreate', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import type { GrokSessionBatchImportResult } from "@/api/admin/accounts";
import GrokSessionBatchImportPanel from "@/components/account/GrokSessionBatchImportPanel.vue";

type GrokSessionInputMode = "single" | "batch";

defineProps<{
  mode: GrokSessionInputMode;
  singleToken: string;
  batchInput: string;
  dryRun: boolean;
  testAfterCreate: boolean;
  result: GrokSessionBatchImportResult | null;
  submitting: boolean;
}>();

const emit = defineEmits<{
  "update:mode": [value: GrokSessionInputMode];
  "update:singleToken": [value: string];
  "update:batchInput": [value: string];
  "update:dryRun": [value: boolean];
  "update:testAfterCreate": [value: boolean];
}>();
</script>
