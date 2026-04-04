<template>
  <span :class="badgeClass">
    {{ exclusive ? t('admin.groups.exclusive') : t('admin.groups.public') }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  exclusive: boolean
}>()

const { t } = useI18n()

const badgeClass = computed(() => [
  'badge',
  props.exclusive ? 'badge-primary' : 'badge-gray'
])
</script>
