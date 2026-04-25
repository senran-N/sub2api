<template>
  <BulkEditApplySection
    id="bulk-edit-proxy"
    :enabled="enabled"
    label-key="admin.accounts.proxy"
    @update:enabled="emit('update:enabled', $event)"
  >
    <ProxySelector
      :model-value="proxyId"
      :proxies="proxies"
      aria-labelledby="bulk-edit-proxy-label"
      @update:model-value="emit('update:proxyId', $event)"
    />
  </BulkEditApplySection>
</template>

<script setup lang="ts">
import BulkEditApplySection from "@/components/account/BulkEditApplySection.vue";
import ProxySelector from "@/components/common/ProxySelector.vue";
import type { Proxy as ProxyConfig } from "@/types";

defineProps<{
  enabled: boolean;
  proxies: ProxyConfig[];
  proxyId: number | null;
}>();

const emit = defineEmits<{
  "update:enabled": [value: boolean];
  "update:proxyId": [value: number | null];
}>();
</script>
