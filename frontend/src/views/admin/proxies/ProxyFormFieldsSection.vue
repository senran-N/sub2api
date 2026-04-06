<template>
  <div>
    <label class="input-label">{{ t('admin.proxies.name') }}</label>
    <input
      v-model="form.name"
      type="text"
      required
      class="input"
      :placeholder="namePlaceholder"
    />
  </div>
  <div>
    <label class="input-label">{{ t('admin.proxies.protocol') }}</label>
    <Select v-model="form.protocol" :options="protocolOptions" />
  </div>
  <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4">
    <div>
      <label class="input-label">{{ t('admin.proxies.host') }}</label>
      <input
        v-model="form.host"
        type="text"
        required
        class="input"
        :placeholder="hostPlaceholder"
      />
    </div>
    <div>
      <label class="input-label">{{ t('admin.proxies.port') }}</label>
      <input
        v-model.number="form.port"
        type="number"
        required
        min="1"
        max="65535"
        class="input"
        :placeholder="portPlaceholder"
      />
    </div>
  </div>
  <div>
    <label class="input-label">{{ t('admin.proxies.username') }}</label>
    <input
      v-model="form.username"
      type="text"
      class="input"
      :placeholder="usernamePlaceholder"
    />
  </div>
  <div>
    <label class="input-label">{{ t('admin.proxies.password') }}</label>
    <div class="relative">
      <input
        v-model="form.password"
        :type="passwordVisible ? 'text' : 'password'"
        :placeholder="passwordPlaceholder"
        class="input pr-10"
        @input="emit('password-input')"
      />
      <button
        type="button"
        class="proxy-form-fields-section__password-toggle absolute right-3 top-1/2 -translate-y-1/2"
        @click="emit('toggle-password-visibility')"
      >
        <Icon :name="passwordVisible ? 'eyeOff' : 'eye'" size="md" />
      </button>
    </div>
  </div>
  <div v-if="showStatus">
    <label class="input-label">{{ t('admin.proxies.status') }}</label>
    <Select v-model="form.status" :options="statusOptions" />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import type { SelectOption } from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import type { ProxyProtocol } from '@/types'

interface ProxyFormFieldsModel {
  name: string
  protocol: ProxyProtocol
  host: string
  port: number
  username: string
  password: string
  status?: 'active' | 'inactive'
}

withDefaults(
  defineProps<{
    form: ProxyFormFieldsModel
    protocolOptions: SelectOption[]
    passwordVisible: boolean
    namePlaceholder?: string
    hostPlaceholder?: string
    portPlaceholder?: string
    usernamePlaceholder?: string
    passwordPlaceholder?: string
    showStatus?: boolean
    statusOptions?: SelectOption[]
  }>(),
  {
    namePlaceholder: '',
    hostPlaceholder: '',
    portPlaceholder: '',
    usernamePlaceholder: '',
    passwordPlaceholder: '',
    showStatus: false,
    statusOptions: () => []
  }
)

const emit = defineEmits<{
  'toggle-password-visibility': []
  'password-input': []
}>()

const { t } = useI18n()
</script>

<style scoped>
.proxy-form-fields-section__password-toggle {
  color: var(--theme-input-placeholder);
  transition: color 0.2s ease;
}

.proxy-form-fields-section__password-toggle:hover {
  color: var(--theme-page-text);
}
</style>
