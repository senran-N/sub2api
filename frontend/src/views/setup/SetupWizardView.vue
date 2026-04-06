<template>
  <div class="setup-shell">
    <div class="setup-shell__inner">
      <div class="setup-shell__intro">
        <div class="setup-shell__brand">
          <Icon name="cog" size="xl" />
        </div>
        <h1 class="setup-shell__title">{{ t('setup.title') }}</h1>
        <p class="setup-shell__description">{{ t('setup.description') }}</p>
      </div>

      <SetupWizardStepper :current-step="currentStep" :steps="steps" />

      <div class="setup-card">
        <SetupDatabaseStep
          v-if="currentStep === 0"
          :connected="dbConnected"
          :database="formData.database"
          :testing="testingDb"
          @test-connection="testDatabaseConnection"
          @update:database="updateDatabase"
        />

        <SetupRedisStep
          v-else-if="currentStep === 1"
          :connected="redisConnected"
          :redis="formData.redis"
          :testing="testingRedis"
          @test-connection="testRedisConnection"
          @update:redis="updateRedis"
        />

        <SetupAdminStep
          v-else-if="currentStep === 2"
          :admin="formData.admin"
          :confirm-password="confirmPassword"
          @update:admin="updateAdmin"
          @update:confirm-password="confirmPassword = $event"
        />

        <SetupReadyStep
          v-else
          :admin-email="formData.admin.email"
          :database="formData.database"
          :redis="formData.redis"
        />

        <div
          v-if="errorMessage"
          class="setup-status-banner setup-status-banner--error"
        >
          <div class="flex items-start gap-3">
            <Icon
              name="exclamationCircle"
              size="md"
              class="setup-status-icon--error flex-shrink-0"
            />
            <p class="setup-status-text setup-status-text--error">{{ errorMessage }}</p>
          </div>
        </div>

        <div
          v-if="installSuccess"
          class="setup-status-banner setup-status-banner--success"
        >
          <div class="flex items-start gap-3">
            <svg
              v-if="!serviceReady"
              class="setup-status-icon--success h-5 w-5 flex-shrink-0 animate-spin"
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
            <Icon
              v-else
              name="checkCircle"
              size="md"
              class="setup-status-icon--success flex-shrink-0"
            />
            <div>
              <p class="setup-status-title setup-status-title--success">
                {{ t('setup.status.completed') }}
              </p>
              <p class="setup-status-text setup-status-text--success">
                {{
                  serviceReady
                    ? t('setup.status.redirecting')
                    : t('setup.status.restarting')
                }}
              </p>
            </div>
          </div>
        </div>

        <div class="mt-8 flex justify-between">
          <button
            v-if="currentStep > 0 && !installSuccess"
            type="button"
            class="btn btn-secondary"
            @click="currentStep -= 1"
          >
            <Icon name="chevronLeft" size="sm" class="mr-2" :stroke-width="2" />
            {{ t('common.back') }}
          </button>
          <div v-else></div>

          <button
            v-if="currentStep < 3"
            type="button"
            class="btn btn-primary"
            :disabled="!canProceed"
            @click="nextStep"
          >
            {{ t('common.next') }}
            <Icon name="chevronRight" size="sm" class="ml-2" :stroke-width="2" />
          </button>

          <button
            v-else-if="!installSuccess"
            type="button"
            class="btn btn-primary"
            :disabled="installing"
            @click="performInstall"
          >
            <svg
              v-if="installing"
              class="-ml-1 mr-2 h-4 w-4 animate-spin"
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
            {{ installing ? t('setup.status.installing') : t('setup.status.completeInstallation') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { install, testDatabase, testRedis, type AdminConfig, type DatabaseConfig, type RedisConfig } from '@/api/setup'
import Icon from '@/components/icons/Icon.vue'
import SetupAdminStep from './SetupAdminStep.vue'
import SetupDatabaseStep from './SetupDatabaseStep.vue'
import SetupReadyStep from './SetupReadyStep.vue'
import SetupRedisStep from './SetupRedisStep.vue'
import SetupWizardStepper from './SetupWizardStepper.vue'
import {
  buildSetupWizardSteps,
  createSetupInstallRequest,
  pollSetupServiceReady,
  resolveSetupWizardErrorMessage,
  SETUP_SERVICE_REDIRECT_DELAY_MS
} from './setupWizardView'

const { t } = useI18n()

const steps = computed(() => buildSetupWizardSteps(t))
const currentStep = ref(0)
const errorMessage = ref('')
const installSuccess = ref(false)
const testingDb = ref(false)
const testingRedis = ref(false)
const dbConnected = ref(false)
const redisConnected = ref(false)
const installing = ref(false)
const confirmPassword = ref('')
const serviceReady = ref(false)
const formData = reactive(createSetupInstallRequest(window.location))

let redirectTimer: number | null = null
let disposed = false

const canProceed = computed(() => {
  switch (currentStep.value) {
    case 0:
      return dbConnected.value
    case 1:
      return redisConnected.value
    case 2:
      return (
        !!formData.admin.email &&
        formData.admin.password.length >= 8 &&
        formData.admin.password === confirmPassword.value
      )
    default:
      return true
  }
})

const updateDatabase = (database: DatabaseConfig) => {
  Object.assign(formData.database, database)
}

const updateRedis = (redis: RedisConfig) => {
  Object.assign(formData.redis, redis)
}

const updateAdmin = (admin: AdminConfig) => {
  Object.assign(formData.admin, admin)
}

const fetchSetupStatus = async () => {
  const response = await fetch('/setup/status', {
    method: 'GET',
    cache: 'no-store'
  })

  if (!response.ok) {
    return null
  }

  return response.json()
}

async function testDatabaseConnection() {
  testingDb.value = true
  errorMessage.value = ''
  dbConnected.value = false

  try {
    await testDatabase(formData.database)
    dbConnected.value = true
  } catch (error: unknown) {
    errorMessage.value = resolveSetupWizardErrorMessage(error, 'Connection failed')
  } finally {
    testingDb.value = false
  }
}

async function testRedisConnection() {
  testingRedis.value = true
  errorMessage.value = ''
  redisConnected.value = false

  try {
    await testRedis(formData.redis)
    redisConnected.value = true
  } catch (error: unknown) {
    errorMessage.value = resolveSetupWizardErrorMessage(error, 'Connection failed')
  } finally {
    testingRedis.value = false
  }
}

function nextStep() {
  if (!canProceed.value) {
    return
  }

  errorMessage.value = ''
  currentStep.value += 1
}

async function waitForServiceRestart() {
  const ready = await pollSetupServiceReady({
    fetchStatus: fetchSetupStatus
  })

  if (disposed) {
    return
  }

  if (!ready) {
    errorMessage.value = t('setup.status.timeout')
    return
  }

  serviceReady.value = true
  redirectTimer = window.setTimeout(() => {
    window.location.href = '/login'
  }, SETUP_SERVICE_REDIRECT_DELAY_MS)
}

async function performInstall() {
  installing.value = true
  errorMessage.value = ''

  try {
    await install(formData)
    installSuccess.value = true
    void waitForServiceRestart()
  } catch (error: unknown) {
    errorMessage.value = resolveSetupWizardErrorMessage(error, 'Installation failed')
  } finally {
    installing.value = false
  }
}

onUnmounted(() => {
  disposed = true

  if (redirectTimer !== null) {
    window.clearTimeout(redirectTimer)
  }
})
</script>
