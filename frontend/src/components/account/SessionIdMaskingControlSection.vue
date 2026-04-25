<template>
  <QuotaControlCard
    :enabled="enabled"
    title-key="admin.accounts.quotaControl.sessionIdMasking.label"
    hint-key="admin.accounts.quotaControl.sessionIdMasking.hint"
    @update:enabled="emit('update:enabled', $event)"
  />
</template>

<script setup lang="ts">
import QuotaControlCard from "@/components/account/QuotaControlCard.vue";

defineProps<{
  enabled: boolean;
}>();

const emit = defineEmits<{
  "update:enabled": [value: boolean];
}>();
</script>
