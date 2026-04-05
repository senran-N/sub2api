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
      <ProfilePasswordForm />
      <ProfileTotpCard />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import StatCard from '@/components/common/StatCard.vue'
import ProfileInfoCard from '@/components/user/profile/ProfileInfoCard.vue'
import ProfileEditForm from '@/components/user/profile/ProfileEditForm.vue'
import ProfilePasswordForm from '@/components/user/profile/ProfilePasswordForm.vue'
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
const { user, contactInfo, loadContactInfo } = useProfileViewModel()

onMounted(() => {
  void loadContactInfo()
})
</script>
