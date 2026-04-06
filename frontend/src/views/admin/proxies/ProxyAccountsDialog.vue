<template>
  <BaseDialog
    :show="show"
    :title="t('admin.proxies.accountsTitle', { name: proxyName || '' })"
    width="normal"
    @close="emit('close')"
  >
    <div
      v-if="loading"
      class="proxy-accounts-dialog__status proxy-accounts-dialog__status--loading flex items-center justify-center text-sm"
    >
      <Icon name="refresh" size="md" class="mr-2 animate-spin" />
      {{ t('common.loading') }}
    </div>
    <div
      v-else-if="accounts.length === 0"
      class="proxy-accounts-dialog__status proxy-accounts-dialog__status--empty text-center text-sm"
    >
      {{ t('admin.proxies.accountsEmpty') }}
    </div>
    <div v-else class="proxy-accounts-dialog__table-shell table-container table-wrapper overflow-auto">
      <table class="table min-w-full text-sm">
        <thead>
          <tr>
            <th class="proxy-accounts-dialog__head-cell">{{ t('admin.proxies.accountName') }}</th>
            <th class="proxy-accounts-dialog__head-cell">{{ t('admin.accounts.columns.platformType') }}</th>
            <th class="proxy-accounts-dialog__head-cell">{{ t('admin.proxies.accountNotes') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="account in accounts" :key="account.id">
            <td class="proxy-accounts-dialog__name">{{ account.name }}</td>
            <td>
              <PlatformTypeBadge :platform="account.platform" :type="account.type" />
            </td>
            <td class="proxy-accounts-dialog__notes">
              {{ account.notes || '-' }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button @click="emit('close')" class="btn btn-secondary">
          {{ t('common.close') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformTypeBadge from '@/components/common/PlatformTypeBadge.vue'
import type { ProxyAccountSummary } from '@/types'

defineProps<{
  show: boolean
  proxyName?: string | null
  loading: boolean
  accounts: ProxyAccountSummary[]
}>()

const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()
</script>

<style scoped>
.proxy-accounts-dialog__status {
  color: var(--theme-page-muted);
}

.proxy-accounts-dialog__status--loading,
.proxy-accounts-dialog__status--empty {
  padding: var(--theme-auth-callback-card-padding);
}

.proxy-accounts-dialog__table-shell table th,
.proxy-accounts-dialog__table-shell table td {
  padding:
    var(--theme-settings-card-panel-padding)
    var(--theme-settings-card-header-padding-x);
}

.proxy-accounts-dialog__table-shell {
  max-height: var(--theme-proxy-accounts-table-max-height);
}

.proxy-accounts-dialog__head-cell {
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.proxy-accounts-dialog__name {
  font-weight: 600;
  color: var(--theme-page-text);
}

.proxy-accounts-dialog__notes {
  color: color-mix(in srgb, var(--theme-page-text) 78%, transparent);
}
</style>
