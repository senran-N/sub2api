<template>
  <div class="profile-support-card card">
    <div class="flex items-center gap-4">
      <div class="profile-support-card__icon">
        <Icon name="chat" size="lg" />
      </div>
      <div>
        <h3 class="profile-support-card__title font-semibold">
          {{ title }}
        </h3>
        <p class="theme-text-strong text-sm font-medium">{{ contactInfo }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  title: string
  contactInfo: string
}>()
</script>

<style scoped>
.profile-support-card {
  padding: var(--theme-settings-card-body-padding);
  border-color: color-mix(in srgb, var(--theme-accent) 18%, var(--theme-card-border));
  background: color-mix(in srgb, var(--theme-accent-soft) 82%, var(--theme-surface));
}

.profile-support-card__icon {
  padding: var(--theme-settings-card-panel-padding);
  border-radius: var(--theme-settings-card-panel-radius);
  background: color-mix(in srgb, var(--theme-accent-soft) 88%, var(--theme-surface));
  color: var(--theme-accent);
}

.profile-support-card__title {
  color: color-mix(in srgb, var(--theme-accent) 72%, var(--theme-page-text));
}
</style>
