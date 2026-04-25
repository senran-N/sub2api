<template>
  <BulkEditApplySection
    id="bulk-edit-status"
    :enabled="enabled"
    label-key="common.status"
    @update:enabled="emit('update:enabled', $event)"
  >
    <Select
      :model-value="status"
      :options="statusOptions"
      aria-labelledby="bulk-edit-status-label"
      @update:model-value="updateStatus"
    />
  </BulkEditApplySection>
</template>

<script setup lang="ts">
import BulkEditApplySection from "@/components/account/BulkEditApplySection.vue";
import Select from "@/components/common/Select.vue";

type BulkEditStatus = "active" | "inactive";

defineProps<{
  enabled: boolean;
  status: BulkEditStatus;
  statusOptions: Array<{ value: BulkEditStatus; label: string }>;
}>();

const emit = defineEmits<{
  "update:enabled": [value: boolean];
  "update:status": [value: BulkEditStatus];
}>();

const updateStatus = (value: string | number | boolean | null) => {
  if (value === "active" || value === "inactive") {
    emit("update:status", value);
  }
};
</script>
