<template>
  <AppLayout>
    <EmbeddedPageShell
      :loading="loading"
      :available="purchaseEnabled"
      :available-icon-name="'creditCard'"
      :available-title="t('purchase.notEnabledTitle')"
      :available-description="t('purchase.notEnabledDesc')"
      :is-valid-url="isValidUrl"
      :invalid-title="t('purchase.notConfiguredTitle')"
      :invalid-description="t('purchase.notConfiguredDesc')"
      :embedded-url="embeddedUrl"
      :open-in-new-tab-label="t('purchase.openInNewTab')"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import AppLayout from '@/components/layout/AppLayout.vue'
import EmbeddedPageShell from './embedded/EmbeddedPageShell.vue'
import { useEmbeddedPageFrame } from './embedded/useEmbeddedPageFrame'

const { t } = useI18n()
const appStore = useAppStore()

const purchaseEnabled = computed(() => {
  return appStore.cachedPublicSettings?.purchase_subscription_enabled ?? false
})

const purchaseBaseUrl = computed(
  () => appStore.cachedPublicSettings?.purchase_subscription_url || ''
)

const { loading, embeddedUrl, isValidUrl } = useEmbeddedPageFrame(purchaseBaseUrl)
</script>
