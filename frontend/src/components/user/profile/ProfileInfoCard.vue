<template>
  <div class="card overflow-hidden">
    <div class="profile-info-card__hero">
      <div class="flex items-center gap-4">
        <div class="profile-info-card__avatar flex items-center justify-center text-2xl font-bold">
          {{ user?.email?.charAt(0).toUpperCase() || 'U' }}
        </div>
        <div class="min-w-0 flex-1">
          <h2 class="profile-info-card__title truncate text-lg font-semibold">
            {{ user?.email }}
          </h2>
          <div class="mt-1 flex items-center gap-2">
            <span :class="['badge', user?.role === 'admin' ? 'badge-primary' : 'badge-gray']">
              {{ user?.role === 'admin' ? t('profile.administrator') : t('profile.user') }}
            </span>
            <span
              :class="['badge', user?.status === 'active' ? 'badge-success' : 'badge-danger']"
            >
              {{ user?.status }}
            </span>
          </div>
        </div>
      </div>
    </div>
    <div class="profile-info-card__body">
      <div class="space-y-3">
        <div class="profile-info-card__detail flex items-center gap-3 text-sm">
          <Icon name="mail" size="sm" class="profile-info-card__detail-icon" />
          <span class="truncate">{{ user?.email }}</span>
        </div>
        <div
          v-if="user?.username"
          class="profile-info-card__detail flex items-center gap-3 text-sm"
        >
          <Icon name="user" size="sm" class="profile-info-card__detail-icon" />
          <span class="truncate">{{ user.username }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { User } from '@/types'

defineProps<{
  user: User | null
}>()

const { t } = useI18n()
</script>

<style scoped>
.profile-info-card__hero {
  padding:
    calc(var(--theme-settings-card-header-padding-y) + 0.25rem)
    var(--theme-settings-card-header-padding-x);
  border-bottom: 1px solid var(--theme-page-border);
  background: linear-gradient(
    135deg,
    color-mix(in srgb, var(--theme-accent-soft) 92%, var(--theme-surface)) 0%,
    color-mix(in srgb, var(--theme-accent) 10%, var(--theme-surface)) 100%
  );
}

.profile-info-card__avatar {
  height: calc(var(--theme-balance-history-avatar-size) + 1.5rem);
  width: calc(var(--theme-balance-history-avatar-size) + 1.5rem);
  border-radius: calc(var(--theme-surface-radius) + 4px);
  color: var(--theme-filled-text);
  background: linear-gradient(
    135deg,
    var(--theme-accent),
    color-mix(in srgb, var(--theme-accent-strong) 42%, var(--theme-accent))
  );
  box-shadow: 0 18px 36px color-mix(in srgb, var(--theme-accent) 24%, transparent);
}

.profile-info-card__body {
  padding:
    var(--theme-settings-card-header-padding-y)
    var(--theme-settings-card-header-padding-x);
}

.profile-info-card__title {
  color: var(--theme-page-text);
}

.profile-info-card__detail {
  color: color-mix(in srgb, var(--theme-page-text) 74%, transparent);
}

.profile-info-card__detail-icon {
  color: color-mix(in srgb, var(--theme-page-muted) 76%, transparent);
}
</style>
