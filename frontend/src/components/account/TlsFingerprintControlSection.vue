<template>
  <QuotaControlCard
    :enabled="enabled"
    title-key="admin.accounts.quotaControl.tlsFingerprint.label"
    hint-key="admin.accounts.quotaControl.tlsFingerprint.hint"
    @update:enabled="emit('update:enabled', $event)"
  >
    <div class="mt-3">
      <select :value="profileId ?? ''" class="input" @change="updateProfileId">
        <option value="">
          {{ t("admin.accounts.quotaControl.tlsFingerprint.defaultProfile") }}
        </option>
        <option v-if="profiles.length > 0" value="-1">
          {{ t("admin.accounts.quotaControl.tlsFingerprint.randomProfile") }}
        </option>
        <option v-for="profile in profiles" :key="profile.id" :value="profile.id">
          {{ profile.name }}
        </option>
      </select>
    </div>
  </QuotaControlCard>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import QuotaControlCard from "@/components/account/QuotaControlCard.vue";

defineProps<{
  enabled: boolean;
  profileId: number | null;
  profiles: Array<{ id: number; name: string }>;
}>();

const emit = defineEmits<{
  "update:enabled": [value: boolean];
  "update:profileId": [value: number | null];
}>();

const { t } = useI18n();

const updateProfileId = (event: Event) => {
  const value = (event.target as HTMLSelectElement).value;
  emit("update:profileId", value === "" ? null : Number(value));
};
</script>
