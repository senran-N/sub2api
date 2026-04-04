<template>
  <span :class="['badge', value === 'admin' ? 'badge-purple' : 'badge-gray']">
    {{ t(`admin.users.roles.${value}`) }}
  </span>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineProps<{
  value: string
}>()

const { t } = useI18n()
</script>
