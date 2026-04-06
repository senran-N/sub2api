<template>
  <AppLayout>
    <div class="subscriptions-view">
      <div v-if="loading" class="subscriptions-view__loading">
        <div class="subscriptions-view__spinner h-8 w-8 animate-spin rounded-full border-2"></div>
      </div>

      <div v-else-if="subscriptions.length === 0" class="card subscriptions-view__empty-card text-center">
        <div class="subscriptions-view__empty-icon mx-auto mb-4 flex items-center justify-center">
          <Icon name="creditCard" size="xl" class="subscriptions-view__empty-icon-symbol" />
        </div>
        <h3 class="subscriptions-view__empty-title mb-2 text-lg font-semibold">
          {{ t('userSubscriptions.noActiveSubscriptions') }}
        </h3>
        <p class="subscriptions-view__empty-description">
          {{ t('userSubscriptions.noActiveSubscriptionsDesc') }}
        </p>
      </div>

      <div v-else class="subscriptions-view__grid">
        <SubscriptionUsageCard
          v-for="subscription in subscriptions"
          :key="subscription.id"
          :now="now"
          :subscription="subscription"
        />
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import subscriptionsAPI from '@/api/subscriptions'
import type { UserSubscription } from '@/types'
import Icon from '@/components/icons/Icon.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAppStore } from '@/stores/app'
import SubscriptionUsageCard from './subscriptions/SubscriptionUsageCard.vue'

const { t } = useI18n()
const appStore = useAppStore()

const subscriptions = ref<UserSubscription[]>([])
const loading = ref(true)
const now = ref(new Date())

async function loadSubscriptions() {
  try {
    loading.value = true
    subscriptions.value = await subscriptionsAPI.getMySubscriptions()
    now.value = new Date()
  } catch (error) {
    console.error('Failed to load subscriptions:', error)
    appStore.showError(t('userSubscriptions.failedToLoad'))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadSubscriptions()
})
</script>

<style scoped>
.subscriptions-view {
  display: flex;
  flex-direction: column;
  gap: var(--theme-table-layout-gap-lg);
}

.subscriptions-view__loading {
  display: flex;
  justify-content: center;
  padding: var(--theme-table-mobile-empty-padding) 0;
}

.subscriptions-view__empty-card {
  padding: var(--theme-table-mobile-empty-padding);
}

.subscriptions-view__grid {
  display: grid;
  gap: var(--theme-table-layout-gap-lg);
}

.subscriptions-view__empty-icon {
  width: var(--theme-empty-icon-surface-size);
  height: var(--theme-empty-icon-surface-size);
  border-radius: var(--theme-empty-surface-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.subscriptions-view__spinner {
  border-color: color-mix(in srgb, var(--theme-accent) 28%, transparent);
  border-top-color: transparent;
}

.subscriptions-view__empty-icon-symbol {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.subscriptions-view__empty-title {
  color: var(--theme-page-text);
}

.subscriptions-view__empty-description {
  color: var(--theme-page-muted);
}

@media (min-width: 1024px) {
  .subscriptions-view__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
