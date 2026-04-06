<template>
  <div class="empty-state">
    <!-- Icon -->
    <div
      class="empty-state__icon-surface mb-5 flex items-center justify-center"
    >
      <slot name="icon">
        <component v-if="icon" :is="icon" class="empty-state-icon" aria-hidden="true" />
        <svg
          v-else
          class="empty-state-icon"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          stroke-width="1.5"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
          />
        </svg>
      </slot>
    </div>

    <!-- Title -->
    <h3 class="empty-state-title">
      {{ displayTitle }}
    </h3>

    <!-- Description -->
    <p class="empty-state-description">
      {{ description }}
    </p>

    <!-- Action -->
    <div v-if="actionText || $slots.action" class="mt-6">
      <slot name="action">
        <component
          :is="actionTo ? 'RouterLink' : 'button'"
          v-if="actionText"
          :to="actionTo"
          @click="!actionTo && $emit('action')"
          class="btn btn-primary"
        >
          <Icon v-if="actionIcon" name="plus" size="md" class="mr-2" />
          {{ actionText }}
        </component>
      </slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Component } from 'vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

interface Props {
  icon?: Component | string
  title?: string
  description?: string
  actionText?: string
  actionTo?: string | object
  actionIcon?: boolean
  message?: string
}

const props = withDefaults(defineProps<Props>(), {
  description: '',
  actionIcon: true
})

const displayTitle = computed(() => props.title || t('common.noData'))

defineEmits(['action'])
</script>

<style scoped>
.empty-state__icon-surface {
  height: var(--theme-empty-icon-surface-size);
  width: var(--theme-empty-icon-surface-size);
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface)),
      color-mix(in srgb, var(--theme-page-border) 56%, transparent)
    );
  border: var(--theme-card-border-width) solid
    color-mix(in srgb, var(--theme-card-border) 70%, transparent);
  box-shadow: var(--theme-card-shadow);
  border-radius: var(--theme-empty-surface-radius);
}
</style>
