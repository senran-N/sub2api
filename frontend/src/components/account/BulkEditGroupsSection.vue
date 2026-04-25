<template>
  <BulkEditApplySection
    id="bulk-edit-groups"
    :enabled="enabled"
    label-key="nav.groups"
    @update:enabled="emit('update:enabled', $event)"
  >
    <GroupSelector
      :model-value="groupIds"
      :groups="groups"
      aria-labelledby="bulk-edit-groups-label"
      @update:model-value="emit('update:groupIds', $event)"
    />
  </BulkEditApplySection>
</template>

<script setup lang="ts">
import BulkEditApplySection from "@/components/account/BulkEditApplySection.vue";
import GroupSelector from "@/components/common/GroupSelector.vue";
import type { AdminGroup } from "@/types";

defineProps<{
  enabled: boolean;
  groupIds: number[];
  groups: AdminGroup[];
}>();

const emit = defineEmits<{
  "update:enabled": [value: boolean];
  "update:groupIds": [value: number[]];
}>();
</script>
