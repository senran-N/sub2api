<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-6">
      <div class="grid grid-cols-1 gap-6 sm:grid-cols-3">
        <StatCard
          :title="t('profile.accountBalance')"
          :value="formatProfileBalance(user?.balance)"
          :icon="profileWalletIcon"
          icon-variant="success"
        />
        <StatCard
          :title="t('profile.concurrencyLimit')"
          :value="user?.concurrency || 0"
          :icon="profileConcurrencyIcon"
          icon-variant="warning"
        />
        <StatCard
          :title="t('profile.memberSince')"
          :value="formatProfileMemberSince(user?.created_at)"
          :icon="profileMemberSinceIcon"
          icon-variant="primary"
        />
      </div>

      <ProfileInfoCard :user="user" />

      <ProfileSupportCard
        v-if="contactInfo"
        :contact-info="contactInfo"
        :title="t('common.contactSupport')"
      />

      <ProfileEditForm :initial-username="user?.username || ''" />
      <ProfileBalanceNotifyCard
        v-if="user && balanceLowNotifyEnabled"
        :enabled="user.balance_notify_enabled"
        :threshold="user.balance_notify_threshold"
        :threshold-type="user.balance_notify_threshold_type"
        :extra-emails="user.balance_notify_extra_emails || []"
        :system-default-threshold="systemDefaultThreshold"
      />
      <ProfileIdentityBindingsSection
        v-if="user"
        :user="user"
        :linuxdo-enabled="linuxdoEnabled"
        :oidc-enabled="oidcEnabled"
        :wechat-enabled="wechatEnabled"
      />
      <ProfilePasswordForm />
      <ProfileTotpCard />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import AppLayout from '@/components/layout/AppLayout.vue'
import StatCard from '@/components/common/StatCard.vue'
import ProfileInfoCard from '@/components/user/profile/ProfileInfoCard.vue'
import ProfileEditForm from '@/components/user/profile/ProfileEditForm.vue'
import ProfilePasswordForm from '@/components/user/profile/ProfilePasswordForm.vue'
import ProfileBalanceNotifyCard from '@/components/user/profile/ProfileBalanceNotifyCard.vue'
import ProfileIdentityBindingsSection from '@/components/user/profile/ProfileIdentityBindingsSection.vue'
import ProfileTotpCard from '@/components/user/profile/ProfileTotpCard.vue'
import ProfileSupportCard from './profile/ProfileSupportCard.vue'
import {
  formatProfileBalance,
  formatProfileMemberSince,
  profileConcurrencyIcon,
  profileMemberSinceIcon,
  profileWalletIcon,
  useProfileViewModel
} from './profile/profileView'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const {
  user,
  contactInfo,
  balanceLowNotifyEnabled,
  systemDefaultThreshold,
  linuxdoEnabled,
  oidcEnabled,
  wechatEnabled,
  loadContactInfo,
  refreshProfile
} = useProfileViewModel()

onMounted(() => {
  void Promise.all([refreshProfile(), loadContactInfo()])

  const oauthBind = typeof route.query.oauth_bind === 'string' ? route.query.oauth_bind : ''
  if (oauthBind.endsWith('_success')) {
    appStore.showSuccess(t('common.saved'))
    const nextQuery = { ...route.query }
    delete nextQuery.oauth_bind
    void router.replace({ path: route.path, query: nextQuery })
  }
})
</script>
