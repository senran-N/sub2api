<template>
  <div class="sora-not-enabled">
    <Icon name="sparkles" size="xl" class="sora-not-enabled-icon" />
    <h2 class="sora-not-enabled-title">{{ title }}</h2>
    <p class="sora-not-enabled-desc">{{ description }}</p>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  title: string
  description: string
}>()
</script>
