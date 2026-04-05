<template>
  <AuthLayout>
    <div class="space-y-6">
      <div class="text-center">
        <h2 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ t('auth.resetPasswordTitle') }}
        </h2>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
          {{ t('auth.resetPasswordHint') }}
        </p>
      </div>

      <div v-if="isInvalidLink" class="space-y-6">
        <div class="rounded-xl border border-red-200 bg-red-50 p-6 dark:border-red-800/50 dark:bg-red-900/20">
          <div class="flex flex-col items-center gap-4 text-center">
            <div class="flex h-12 w-12 items-center justify-center rounded-full bg-red-100 dark:bg-red-800/50">
              <Icon name="exclamationCircle" size="lg" class="text-red-600 dark:text-red-400" />
            </div>
            <div>
              <h3 class="text-lg font-semibold text-red-800 dark:text-red-200">
                {{ t('auth.invalidResetLink') }}
              </h3>
              <p class="mt-2 text-sm text-red-700 dark:text-red-300">
                {{ t('auth.invalidResetLinkHint') }}
              </p>
            </div>
          </div>
        </div>

        <div class="text-center">
          <router-link
            to="/forgot-password"
            class="inline-flex items-center gap-2 font-medium text-primary-600 transition-colors hover:text-primary-500 dark:text-primary-400 dark:hover:text-primary-300"
          >
            {{ t('auth.requestNewResetLink') }}
          </router-link>
        </div>
      </div>

      <div v-else-if="isSuccess" class="space-y-6">
        <div class="rounded-xl border border-green-200 bg-green-50 p-6 dark:border-green-800/50 dark:bg-green-900/20">
          <div class="flex flex-col items-center gap-4 text-center">
            <div class="flex h-12 w-12 items-center justify-center rounded-full bg-green-100 dark:bg-green-800/50">
              <Icon name="checkCircle" size="lg" class="text-green-600 dark:text-green-400" />
            </div>
            <div>
              <h3 class="text-lg font-semibold text-green-800 dark:text-green-200">
                {{ t('auth.passwordResetSuccess') }}
              </h3>
              <p class="mt-2 text-sm text-green-700 dark:text-green-300">
                {{ t('auth.passwordResetSuccessHint') }}
              </p>
            </div>
          </div>
        </div>

        <div class="text-center">
          <router-link to="/login" class="btn btn-primary inline-flex items-center gap-2">
            <Icon name="login" size="md" />
            {{ t('auth.signIn') }}
          </router-link>
        </div>
      </div>

      <form v-else class="space-y-5" @submit.prevent="handleSubmit">
        <div>
          <label for="email" class="input-label">
            {{ t('auth.emailLabel') }}
          </label>
          <div class="relative">
            <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
              <Icon name="mail" size="md" class="text-gray-400 dark:text-dark-500" />
            </div>
            <input
              id="email"
              :value="routeState.email"
              type="email"
              readonly
              disabled
              class="input pl-11 bg-gray-50 dark:bg-dark-700"
            />
          </div>
        </div>

        <ResetPasswordField
          id="password"
          v-model="formData.password"
          :disabled="isLoading"
          :error="errors.password"
          :label="t('auth.newPassword')"
          :placeholder="t('auth.newPasswordPlaceholder')"
        />

        <ResetPasswordField
          id="confirmPassword"
          v-model="formData.confirmPassword"
          :disabled="isLoading"
          :error="errors.confirmPassword"
          :label="t('auth.confirmPassword')"
          :placeholder="t('auth.confirmPasswordPlaceholder')"
        />

        <transition name="fade">
          <div
            v-if="errorMessage"
            class="rounded-xl border border-red-200 bg-red-50 p-4 dark:border-red-800/50 dark:bg-red-900/20"
          >
            <div class="flex items-start gap-3">
              <div class="flex-shrink-0">
                <Icon name="exclamationCircle" size="md" class="text-red-500" />
              </div>
              <p class="text-sm text-red-700 dark:text-red-400">
                {{ errorMessage }}
              </p>
            </div>
          </div>
        </transition>

        <button type="submit" :disabled="isLoading" class="btn btn-primary w-full">
          <svg
            v-if="isLoading"
            class="-ml-1 mr-2 h-4 w-4 animate-spin text-white"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          <Icon v-else name="checkCircle" size="md" class="mr-2" />
          {{ isLoading ? t('auth.resettingPassword') : t('auth.resetPassword') }}
        </button>
      </form>
    </div>

    <template #footer>
      <p class="text-gray-500 dark:text-dark-400">
        {{ t('auth.rememberedPassword') }}
        <router-link
          to="/login"
          class="font-medium text-primary-600 transition-colors hover:text-primary-500 dark:text-primary-400 dark:hover:text-primary-300"
        >
          {{ t('auth.signIn') }}
        </router-link>
      </p>
    </template>
  </AuthLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { resetPassword } from '@/api/auth'
import Icon from '@/components/icons/Icon.vue'
import { AuthLayout } from '@/components/layout'
import { useAppStore } from '@/stores'
import ResetPasswordField from './reset-password/ResetPasswordField.vue'
import {
  buildResetPasswordSubmitPayload,
  createResetPasswordFormData,
  createResetPasswordFormErrors,
  hasResetPasswordFormErrors,
  isResetPasswordLinkInvalid,
  resolveResetPasswordErrorMessage,
  resolveResetPasswordRouteState,
  validateResetPasswordForm
} from './reset-password/resetPasswordView'

const { t } = useI18n()

const route = useRoute()
const appStore = useAppStore()

const isLoading = ref(false)
const isSuccess = ref(false)
const errorMessage = ref('')

const routeState = reactive(resolveResetPasswordRouteState({}))
const formData = reactive(createResetPasswordFormData())
const errors = reactive(createResetPasswordFormErrors())

const isInvalidLink = computed(() => isResetPasswordLinkInvalid(routeState))

function validateForm(): boolean {
  Object.assign(errors, validateResetPasswordForm(formData, t))
  return !hasResetPasswordFormErrors(errors)
}

async function handleSubmit(): Promise<void> {
  errorMessage.value = ''

  if (!validateForm()) {
    return
  }

  isLoading.value = true

  try {
    await resetPassword(buildResetPasswordSubmitPayload(routeState, formData))
    isSuccess.value = true
    appStore.showSuccess(t('auth.passwordResetSuccess'))
  } catch (error: unknown) {
    errorMessage.value = resolveResetPasswordErrorMessage(error, t)
    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  Object.assign(routeState, resolveResetPasswordRouteState(route.query))
})
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
