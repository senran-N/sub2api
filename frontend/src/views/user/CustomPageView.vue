<template>
  <AppLayout>
    <EmbeddedPageShell
      :loading="loading"
      :available="Boolean(menuItem)"
      :available-icon-name="'link'"
      :available-title="t('customPage.notFoundTitle')"
      :available-description="t('customPage.notFoundDesc')"
      :is-valid-url="isValidUrl"
      :invalid-title="t('customPage.notConfiguredTitle')"
      :invalid-description="t('customPage.notConfiguredDesc')"
      :embedded-url="embeddedUrl"
      :open-in-new-tab-label="t('customPage.openInNewTab')"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useAdminSettingsStore } from '@/stores/adminSettings'
import AppLayout from '@/components/layout/AppLayout.vue'
import EmbeddedPageShell from './embedded/EmbeddedPageShell.vue'
import { resolveCustomPageMenuItem } from './embedded/embeddedPageFrame'
import { useEmbeddedPageFrame } from './embedded/useEmbeddedPageFrame'

const { t } = useI18n()
const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()
const adminSettingsStore = useAdminSettingsStore()

const menuItemId = computed(() => route.params.id as string)

const menuItem = computed(() => {
  return resolveCustomPageMenuItem(
    menuItemId.value,
    appStore.cachedPublicSettings?.custom_menu_items ?? [],
    adminSettingsStore.customMenuItems,
    authStore.isAdmin
  )
})

const embeddedBaseUrl = computed(() => menuItem.value?.url || '')

const { loading, embeddedUrl, isValidUrl } = useEmbeddedPageFrame(embeddedBaseUrl)
</script>
