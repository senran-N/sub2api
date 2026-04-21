<template>
  <div
    v-if="ACCOUNT_FILTER_PLATFORMS.includes(form.platform)"
    class="group-account-filter-section mt-4 space-y-4 pt-4"
  >
    <h3 class="group-account-filter-section__title mb-3 text-sm font-medium">
      {{ t('admin.groups.accountFilter.title') }}
    </h3>

    <div class="flex items-center justify-between">
      <div>
        <label class="group-account-filter-section__label text-sm">
          {{ t('admin.groups.accountFilter.oauthOnly') }}
        </label>
        <p class="group-account-filter-section__hint mt-0.5 text-xs">
          {{
            form.require_oauth_only
              ? t('admin.groups.accountFilter.oauthOnlyEnabled')
              : t('admin.groups.accountFilter.disabled')
          }}
        </p>
      </div>
      <Toggle
        v-model="form.require_oauth_only"
        :aria-label="t('admin.groups.accountFilter.oauthOnly')"
      />
    </div>

    <div class="flex items-center justify-between">
      <div>
        <label class="group-account-filter-section__label text-sm">
          {{ t('admin.groups.accountFilter.privacySetOnly') }}
        </label>
        <p class="group-account-filter-section__hint mt-0.5 text-xs">
          {{
            form.require_privacy_set
              ? t('admin.groups.accountFilter.privacySetOnlyEnabled')
              : t('admin.groups.accountFilter.disabled')
          }}
        </p>
      </div>
      <Toggle
        v-model="form.require_privacy_set"
        :aria-label="t('admin.groups.accountFilter.privacySetOnly')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import { ACCOUNT_FILTER_PLATFORMS } from './groupsForm'
import type { CreateGroupForm, EditGroupForm } from './groupsForm'

defineProps<{
  form: CreateGroupForm | EditGroupForm
}>()

const { t } = useI18n()
</script>

<style scoped>
.group-account-filter-section {
  border-top: 1px solid var(--theme-page-border);
}

.group-account-filter-section__title,
.group-account-filter-section__label {
  color: var(--theme-page-text);
}

.group-account-filter-section__hint {
  color: var(--theme-page-muted);
}
</style>
