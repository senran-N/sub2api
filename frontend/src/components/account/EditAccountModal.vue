<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.editAccount')"
    width="normal"
    @close="handleClose"
  >
    <form
      v-if="account"
      id="edit-account-form"
      @submit.prevent="handleSubmit"
      class="space-y-5"
    >
      <div>
        <label class="input-label">{{ t("common.name") }}</label>
        <input
          v-model="form.name"
          type="text"
          required
          class="input"
          data-tour="edit-account-form-name"
        />
      </div>
      <div>
        <label class="input-label">{{ t("admin.accounts.notes") }}</label>
        <textarea
          v-model="form.notes"
          rows="3"
          class="input"
          :placeholder="t('admin.accounts.notesPlaceholder')"
        ></textarea>
        <p class="input-hint">{{ t("admin.accounts.notesHint") }}</p>
      </div>

      <div v-if="account.platform === 'grok'" class="form-section space-y-4">
        <div>
          <label class="input-label mb-0">{{
            t("admin.accounts.grok.runtime.title")
          }}</label>
          <p class="edit-account-modal__muted mt-1 text-xs">
            {{ t("admin.accounts.grok.runtime.hint") }}
          </p>
        </div>

        <div class="grid grid-cols-1 gap-3 lg:grid-cols-2">
          <div
            class="edit-account-modal__config-card edit-account-modal__config-card--compact space-y-2"
          >
            <div class="flex flex-wrap items-center gap-2">
              <span class="edit-account-modal__muted text-xs">{{
                t("admin.accounts.grok.runtime.tier")
              }}</span>
              <span :class="grokTierChipClass">
                {{ grokTierLabel }}
              </span>
            </div>
            <div class="grid grid-cols-1 gap-2 text-xs sm:grid-cols-2">
              <div>
                <span class="edit-account-modal__muted">{{
                  t("admin.accounts.grok.runtime.authMode")
                }}</span>
                <div class="font-medium">{{ grokAuthModeLabel }}</div>
              </div>
              <div>
                <span class="edit-account-modal__muted">{{
                  t("admin.accounts.grok.runtime.source")
                }}</span>
                <div>{{ grokTierSourceDisplay }}</div>
              </div>
              <div>
                <span class="edit-account-modal__muted">{{
                  t("admin.accounts.grok.runtime.confidence")
                }}</span>
                <div>{{ grokTierConfidenceDisplay }}</div>
              </div>
            </div>
          </div>

          <div
            class="edit-account-modal__config-card edit-account-modal__config-card--compact space-y-2"
          >
            <div>
              <span class="edit-account-modal__muted text-xs">{{
                t("admin.accounts.grok.runtime.capabilityTitle")
              }}</span>
              <div class="mt-1 flex flex-wrap gap-1.5">
                <span
                  v-for="capability in grokCapabilities"
                  :key="capability"
                  class="theme-chip theme-chip--compact theme-chip--info"
                >
                  {{
                    t(`admin.accounts.grok.runtime.capabilities.${capability}`)
                  }}
                </span>
                <span
                  v-if="grokCapabilities.length === 0"
                  class="edit-account-modal__muted text-xs"
                >
                  {{ t("admin.accounts.grok.runtime.empty") }}
                </span>
              </div>
            </div>
            <div>
              <span class="edit-account-modal__muted text-xs">{{
                t("admin.accounts.grok.runtime.models")
              }}</span>
              <div class="mt-1 flex flex-wrap gap-1.5">
                <span
                  v-for="model in grokModels"
                  :key="model"
                  class="theme-chip theme-chip--compact theme-chip--neutral font-mono"
                >
                  {{ model }}
                </span>
                <span
                  v-if="grokModels.length === 0"
                  class="edit-account-modal__muted text-xs"
                >
                  {{ t("admin.accounts.grok.runtime.empty") }}
                </span>
              </div>
            </div>
          </div>

          <div
            class="edit-account-modal__config-card edit-account-modal__config-card--compact space-y-2"
          >
            <span class="edit-account-modal__muted text-xs">{{
              t("admin.accounts.grok.runtime.quotaTitle")
            }}</span>
            <div v-if="grokQuotaWindows.length > 0" class="space-y-2">
              <div
                v-for="window in grokQuotaWindows"
                :key="window.name"
                class="rounded-lg border px-3 py-2 text-xs"
              >
                <div class="flex items-center justify-between gap-3">
                  <span class="font-medium">{{
                    t(`admin.accounts.grok.runtime.windows.${window.name}`)
                  }}</span>
                  <span class="font-mono"
                    >{{ window.remaining }}/{{ window.total }}</span
                  >
                </div>
                <div
                  class="edit-account-modal__muted mt-1 flex flex-wrap gap-x-3 gap-y-1"
                >
                  <span
                    >{{ t("admin.accounts.grok.runtime.source") }}:
                    {{ window.source || emptyRuntimeValue }}</span
                  >
                  <span
                    >{{ t("admin.accounts.grok.runtime.resetAt") }}:
                    {{ formatRuntimeValue(window.resetAt) }}</span
                  >
                </div>
              </div>
            </div>
            <div v-else class="edit-account-modal__muted text-xs">
              {{ t("admin.accounts.grok.runtime.empty") }}
            </div>
          </div>

          <div
            class="edit-account-modal__config-card edit-account-modal__config-card--compact space-y-2"
          >
            <span class="edit-account-modal__muted text-xs">{{
              t("admin.accounts.grok.runtime.syncTitle")
            }}</span>
            <div class="grid grid-cols-1 gap-2 text-xs">
              <div class="flex items-center justify-between gap-3">
                <span class="edit-account-modal__muted">{{
                  t("admin.accounts.grok.runtime.lastSyncAt")
                }}</span>
                <span>{{
                  formatRuntimeValue(grokRuntimeState?.sync.lastSyncAt)
                }}</span>
              </div>
              <div class="flex items-center justify-between gap-3">
                <span class="edit-account-modal__muted">{{
                  t("admin.accounts.grok.runtime.lastProbeAt")
                }}</span>
                <span>{{
                  formatRuntimeValue(grokRuntimeState?.sync.lastProbeAt)
                }}</span>
              </div>
              <div class="flex items-center justify-between gap-3">
                <span class="edit-account-modal__muted">{{
                  t("admin.accounts.grok.runtime.lastProbeOkAt")
                }}</span>
                <span>{{
                  formatRuntimeValue(grokRuntimeState?.sync.lastProbeOkAt)
                }}</span>
              </div>
              <div class="flex items-center justify-between gap-3">
                <span class="edit-account-modal__muted">{{
                  t("admin.accounts.grok.runtime.lastProbeErrorAt")
                }}</span>
                <span>{{
                  formatRuntimeValue(grokRuntimeState?.sync.lastProbeErrorAt)
                }}</span>
              </div>
              <div class="flex items-center justify-between gap-3">
                <span class="edit-account-modal__muted">{{
                  t("admin.accounts.grok.runtime.probeStatus")
                }}</span>
                <span>{{ grokProbeStatusDisplay }}</span>
              </div>
              <div class="space-y-1 rounded-lg border border-dashed px-3 py-2">
                <span class="edit-account-modal__muted block">{{
                  t("admin.accounts.grok.runtime.lastProbeError")
                }}</span>
                <div>{{ grokProbeErrorDisplay }}</div>
              </div>
              <div class="space-y-1 rounded-lg border border-dashed px-3 py-2">
                <span class="edit-account-modal__muted block">{{
                  t("admin.accounts.grok.runtime.lastRuntimeError")
                }}</span>
                <div>{{ grokRuntimeErrorDisplay }}</div>
                <div class="edit-account-modal__muted">
                  {{ t("admin.accounts.grok.runtime.lastFailAt") }}:
                  {{ formatRuntimeValue(grokRuntimeState?.runtime.lastFailAt) }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Compatible API credentials (API Key / Upstream) -->
      <div v-if="showCompatibleCredentialsForm" class="space-y-4">
        <div>
          <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
            <label class="input-label mb-0">{{
              t("admin.accounts.baseUrl")
            }}</label>
            <div
              v-if="compatibleBaseUrlPresets.length > 0"
              class="flex flex-wrap gap-2"
            >
              <button
                v-for="preset in compatibleBaseUrlPresets"
                :key="preset.value"
                type="button"
                :class="getPresetMappingChipClasses('success')"
                @click="editBaseUrl = preset.value"
              >
                {{ preset.label }}
              </button>
            </div>
          </div>
          <input
            v-model="editBaseUrl"
            type="text"
            class="input"
            :placeholder="baseUrlPlaceholder"
          />
          <p class="input-hint">{{ baseUrlHint }}</p>
        </div>
        <div>
          <label class="input-label">{{ t("admin.accounts.apiKey") }}</label>
          <input
            v-model="editApiKey"
            type="password"
            class="input font-mono"
            autocomplete="new-password"
            data-1p-ignore
            data-lpignore="true"
            data-bwignore="true"
            :placeholder="apiKeyPlaceholder"
          />
          <p class="input-hint">{{ t("admin.accounts.leaveEmptyToKeep") }}</p>
        </div>

        <!-- Model Restriction Section (不适用于 Antigravity) -->
        <div v-if="account.platform !== 'antigravity'" class="form-section">
          <label class="input-label">{{
            t("admin.accounts.modelRestriction")
          }}</label>

          <div
            v-if="isOpenAIModelRestrictionDisabled"
            :class="['mb-3', getToneNoticeClasses('amber')]"
          >
            <p class="text-xs">
              {{
                t("admin.accounts.openai.modelRestrictionDisabledByPassthrough")
              }}
            </p>
          </div>

          <template v-else>
            <!-- Mode Toggle -->
            <div class="mb-4 flex gap-2">
              <button
                type="button"
                @click="modelRestrictionMode = 'whitelist'"
                :class="
                  getModeToggleClasses(
                    modelRestrictionMode === 'whitelist',
                    'accent',
                  )
                "
              >
                <svg
                  class="mr-1.5 inline h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
                {{ t("admin.accounts.modelWhitelist") }}
              </button>
              <button
                type="button"
                @click="modelRestrictionMode = 'mapping'"
                :class="
                  getModeToggleClasses(
                    modelRestrictionMode === 'mapping',
                    'purple',
                  )
                "
              >
                <svg
                  class="mr-1.5 inline h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"
                  />
                </svg>
                {{ t("admin.accounts.modelMapping") }}
              </button>
            </div>

            <!-- Whitelist Mode -->
            <div v-if="modelRestrictionMode === 'whitelist'">
              <ModelWhitelistSelector
                v-model="allowedModels"
                :platform="account?.platform || 'anthropic'"
              />
              <p class="edit-account-modal__muted text-xs">
                {{
                  t("admin.accounts.selectedModels", {
                    count: allowedModels.length,
                  })
                }}
                <span v-if="allowedModels.length === 0">{{
                  t("admin.accounts.supportsAllModels")
                }}</span>
              </p>
            </div>

            <!-- Mapping Mode -->
            <div v-else>
              <div :class="['mb-3', getToneNoticeClasses('purple')]">
                <p class="text-xs">
                  <svg
                    class="mr-1 inline h-4 w-4"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                  {{ t("admin.accounts.mapRequestModels") }}
                </p>
              </div>

              <!-- Model Mapping List -->
              <div v-if="modelMappings.length > 0" class="mb-3 space-y-2">
                <div
                  v-for="(mapping, index) in modelMappings"
                  :key="getModelMappingKey(mapping)"
                  class="flex items-center gap-2"
                >
                  <input
                    v-model="mapping.from"
                    type="text"
                    class="input flex-1"
                    :placeholder="t('admin.accounts.requestModel')"
                  />
                  <svg
                    class="edit-account-modal__muted h-4 w-4 flex-shrink-0"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M14 5l7 7m0 0l-7 7m7-7H3"
                    />
                  </svg>
                  <input
                    v-model="mapping.to"
                    type="text"
                    class="input flex-1"
                    :placeholder="t('admin.accounts.actualModel')"
                  />
                  <button
                    type="button"
                    @click="removeModelMapping(index)"
                    :class="getStatusChipClasses(true, 'danger')"
                  >
                    <svg
                      class="h-4 w-4"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                      />
                    </svg>
                  </button>
                </div>
              </div>

              <button
                type="button"
                @click="addModelMapping"
                class="btn btn-secondary mb-3 w-full border-2 border-dashed"
              >
                <svg
                  class="mr-1 inline h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 4v16m8-8H4"
                  />
                </svg>
                {{ t("admin.accounts.addMapping") }}
              </button>

              <!-- Quick Add Buttons -->
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="preset in presetMappings"
                  :key="preset.label"
                  type="button"
                  @click="addPresetMapping(preset.from, preset.to)"
                  :class="getPresetMappingChipClasses(preset.tone)"
                >
                  + {{ preset.label }}
                </button>
              </div>
            </div>
          </template>
        </div>

        <!-- Pool Mode Section -->
        <div class="form-section">
          <div class="mb-3 flex items-center justify-between">
            <div>
              <label class="input-label mb-0">{{
                t("admin.accounts.poolMode")
              }}</label>
              <p class="edit-account-modal__muted mt-1 text-xs">
                {{ t("admin.accounts.poolModeHint") }}
              </p>
            </div>
            <button
              type="button"
              @click="poolModeEnabled = !poolModeEnabled"
              :class="getSwitchTrackClasses(poolModeEnabled)"
            >
              <span :class="getSwitchThumbClasses(poolModeEnabled)" />
            </button>
          </div>
          <div v-if="poolModeEnabled" :class="getToneNoticeClasses('blue')">
            <p class="text-xs">
              <Icon
                name="exclamationCircle"
                size="sm"
                class="mr-1 inline"
                :stroke-width="2"
              />
              {{ t("admin.accounts.poolModeInfo") }}
            </p>
          </div>
          <div v-if="poolModeEnabled" class="mt-3">
            <label class="input-label">{{
              t("admin.accounts.poolModeRetryCount")
            }}</label>
            <input
              v-model.number="poolModeRetryCount"
              type="number"
              min="0"
              :max="MAX_POOL_MODE_RETRY_COUNT"
              step="1"
              class="input"
            />
            <p class="edit-account-modal__muted mt-1 text-xs">
              {{
                t("admin.accounts.poolModeRetryCountHint", {
                  default: DEFAULT_POOL_MODE_RETRY_COUNT,
                  max: MAX_POOL_MODE_RETRY_COUNT,
                })
              }}
            </p>
          </div>
        </div>

        <!-- Custom Error Codes Section -->
        <div class="form-section">
          <div class="mb-3 flex items-center justify-between">
            <div>
              <label class="input-label mb-0">{{
                t("admin.accounts.customErrorCodes")
              }}</label>
              <p class="edit-account-modal__muted mt-1 text-xs">
                {{ t("admin.accounts.customErrorCodesHint") }}
              </p>
            </div>
            <button
              type="button"
              @click="customErrorCodesEnabled = !customErrorCodesEnabled"
              :class="getSwitchTrackClasses(customErrorCodesEnabled)"
            >
              <span :class="getSwitchThumbClasses(customErrorCodesEnabled)" />
            </button>
          </div>

          <div v-if="customErrorCodesEnabled" class="space-y-3">
            <div :class="getToneNoticeClasses('amber')">
              <p class="text-xs">
                <Icon
                  name="exclamationTriangle"
                  size="sm"
                  class="mr-1 inline"
                  :stroke-width="2"
                />
                {{ t("admin.accounts.customErrorCodesWarning") }}
              </p>
            </div>

            <!-- Error Code Buttons -->
            <div class="flex flex-wrap gap-2">
              <button
                v-for="code in commonErrorCodes"
                :key="code.value"
                type="button"
                @click="toggleErrorCode(code.value)"
                :class="
                  getStatusChipClasses(
                    selectedErrorCodes.includes(code.value),
                    'danger',
                  )
                "
              >
                {{ code.value }} {{ code.label }}
              </button>
            </div>

            <!-- Manual input -->
            <div class="flex items-center gap-2">
              <input
                v-model.number="customErrorCodeInput"
                type="number"
                min="100"
                max="599"
                class="input flex-1"
                :placeholder="t('admin.accounts.enterErrorCode')"
                @keyup.enter="addCustomErrorCode"
              />
              <button
                type="button"
                @click="addCustomErrorCode"
                class="edit-account-modal__compact-action btn btn-secondary"
              >
                <svg
                  class="h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 4v16m8-8H4"
                  />
                </svg>
              </button>
            </div>

            <!-- Selected codes summary -->
            <div class="flex flex-wrap gap-1.5">
              <span
                v-for="code in selectedErrorCodes.sort((a, b) => a - b)"
                :key="code"
                :class="[
                  'edit-account-modal__summary-chip inline-flex items-center gap-1',
                  getStatusChipClasses(true, 'danger'),
                ]"
              >
                {{ code }}
                <button
                  type="button"
                  @click="removeErrorCode(code)"
                  class="edit-account-modal__choice-text"
                >
                  <Icon name="x" size="sm" :stroke-width="2" />
                </button>
              </span>
              <span
                v-if="selectedErrorCodes.length === 0"
                class="edit-account-modal__muted text-xs"
              >
                {{ t("admin.accounts.noneSelectedUsesDefault") }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- OpenAI OAuth Model Mapping (OAuth 类型没有 apikey 容器，需要独立的模型映射区域) -->
      <div
        v-if="account.platform === 'openai' && account.type === 'oauth'"
        class="form-section"
      >
        <label class="input-label">{{
          t("admin.accounts.modelRestriction")
        }}</label>

        <div
          v-if="isOpenAIModelRestrictionDisabled"
          :class="['mb-3', getToneNoticeClasses('amber')]"
        >
          <p class="text-xs">
            {{
              t("admin.accounts.openai.modelRestrictionDisabledByPassthrough")
            }}
          </p>
        </div>

        <template v-else>
          <!-- Mode Toggle -->
          <div class="mb-4 flex gap-2">
            <button
              type="button"
              @click="modelRestrictionMode = 'whitelist'"
              :class="
                getModeToggleClasses(
                  modelRestrictionMode === 'whitelist',
                  'accent',
                )
              "
            >
              {{ t("admin.accounts.modelWhitelist") }}
            </button>
            <button
              type="button"
              @click="modelRestrictionMode = 'mapping'"
              :class="
                getModeToggleClasses(
                  modelRestrictionMode === 'mapping',
                  'purple',
                )
              "
            >
              {{ t("admin.accounts.modelMapping") }}
            </button>
          </div>

          <!-- Whitelist Mode -->
          <div v-if="modelRestrictionMode === 'whitelist'">
            <ModelWhitelistSelector
              v-model="allowedModels"
              :platform="account?.platform || 'anthropic'"
            />
            <p class="edit-account-modal__muted text-xs">
              {{
                t("admin.accounts.selectedModels", {
                  count: allowedModels.length,
                })
              }}
              <span v-if="allowedModels.length === 0">{{
                t("admin.accounts.supportsAllModels")
              }}</span>
            </p>
          </div>

          <!-- Mapping Mode -->
          <div v-else>
            <div :class="['mb-3', getToneNoticeClasses('purple')]">
              <p class="text-xs">
                {{ t("admin.accounts.mapRequestModels") }}
              </p>
            </div>

            <div v-if="modelMappings.length > 0" class="mb-3 space-y-2">
              <div
                v-for="(mapping, index) in modelMappings"
                :key="'oauth-' + getModelMappingKey(mapping)"
                class="flex items-center gap-2"
              >
                <input
                  v-model="mapping.from"
                  type="text"
                  class="input flex-1"
                  :placeholder="t('admin.accounts.requestModel')"
                />
                <svg
                  class="edit-account-modal__muted h-4 w-4 flex-shrink-0"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M14 5l7 7m0 0l-7 7m7-7H3"
                  />
                </svg>
                <input
                  v-model="mapping.to"
                  type="text"
                  class="input flex-1"
                  :placeholder="t('admin.accounts.actualModel')"
                />
                <button
                  type="button"
                  @click="removeModelMapping(index)"
                  :class="getStatusChipClasses(true, 'danger')"
                >
                  <svg
                    class="h-4 w-4"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                    />
                  </svg>
                </button>
              </div>
            </div>

            <button
              type="button"
              @click="addModelMapping"
              class="btn btn-secondary mb-3 w-full border-2 border-dashed"
            >
              + {{ t("admin.accounts.addMapping") }}
            </button>

            <!-- Quick Add Buttons -->
            <div class="flex flex-wrap gap-2">
              <button
                v-for="preset in presetMappings"
                :key="'oauth-' + preset.label"
                type="button"
                @click="addPresetMapping(preset.from, preset.to)"
                :class="getPresetMappingChipClasses(preset.tone)"
              >
                + {{ preset.label }}
              </button>
            </div>
          </div>
        </template>
      </div>

      <!-- Grok session credentials -->
      <div v-if="account.type === 'session'" class="space-y-4">
        <div>
          <label class="input-label">{{
            t("admin.accounts.grok.sessionToken")
          }}</label>
          <input
            v-model="editSessionToken"
            type="password"
            class="input font-mono"
            :placeholder="t('admin.accounts.grok.sessionTokenPlaceholder')"
          />
          <p class="input-hint">{{ t("admin.accounts.leaveEmptyToKeep") }}</p>
        </div>
      </div>

      <!-- Bedrock fields (for bedrock type, both SigV4 and API Key modes) -->
      <div v-if="account.type === 'bedrock'" class="space-y-4">
        <!-- SigV4 fields -->
        <template v-if="!isBedrockAPIKeyMode">
          <div>
            <label class="input-label">{{
              t("admin.accounts.bedrockAccessKeyId")
            }}</label>
            <input
              v-model="editBedrockAccessKeyId"
              type="text"
              class="input font-mono"
              placeholder="AKIA..."
            />
          </div>
          <div>
            <label class="input-label">{{
              t("admin.accounts.bedrockSecretAccessKey")
            }}</label>
            <input
              v-model="editBedrockSecretAccessKey"
              type="password"
              class="input font-mono"
              :placeholder="t('admin.accounts.bedrockSecretKeyLeaveEmpty')"
            />
            <p class="input-hint">
              {{ t("admin.accounts.bedrockSecretKeyLeaveEmpty") }}
            </p>
          </div>
          <div>
            <label class="input-label">{{
              t("admin.accounts.bedrockSessionToken")
            }}</label>
            <input
              v-model="editBedrockSessionToken"
              type="password"
              class="input font-mono"
              :placeholder="t('admin.accounts.bedrockSecretKeyLeaveEmpty')"
            />
            <p class="input-hint">
              {{ t("admin.accounts.bedrockSessionTokenHint") }}
            </p>
          </div>
        </template>

        <!-- API Key field -->
        <div v-if="isBedrockAPIKeyMode">
          <label class="input-label">{{
            t("admin.accounts.bedrockApiKeyInput")
          }}</label>
          <input
            v-model="editBedrockApiKeyValue"
            type="password"
            class="input font-mono"
            :placeholder="t('admin.accounts.bedrockApiKeyLeaveEmpty')"
          />
          <p class="input-hint">
            {{ t("admin.accounts.bedrockApiKeyLeaveEmpty") }}
          </p>
        </div>

        <!-- Shared: Region -->
        <div>
          <label class="input-label">{{
            t("admin.accounts.bedrockRegion")
          }}</label>
          <input
            v-model="editBedrockRegion"
            type="text"
            class="input"
            placeholder="us-east-1"
          />
          <p class="input-hint">{{ t("admin.accounts.bedrockRegionHint") }}</p>
        </div>

        <!-- Shared: Force Global -->
        <div>
          <label class="flex items-center gap-2 cursor-pointer">
            <input
              v-model="editBedrockForceGlobal"
              type="checkbox"
              class="edit-account-modal__checkbox rounded"
            />
            <span class="edit-account-modal__choice-text text-sm">{{
              t("admin.accounts.bedrockForceGlobal")
            }}</span>
          </label>
          <p class="input-hint mt-1">
            {{ t("admin.accounts.bedrockForceGlobalHint") }}
          </p>
        </div>

        <!-- Model Restriction for Bedrock -->
        <div class="form-section">
          <label class="input-label">{{
            t("admin.accounts.modelRestriction")
          }}</label>

          <!-- Mode Toggle -->
          <div class="mb-4 flex gap-2">
            <button
              type="button"
              @click="modelRestrictionMode = 'whitelist'"
              :class="
                getModeToggleClasses(
                  modelRestrictionMode === 'whitelist',
                  'accent',
                )
              "
            >
              {{ t("admin.accounts.modelWhitelist") }}
            </button>
            <button
              type="button"
              @click="modelRestrictionMode = 'mapping'"
              :class="
                getModeToggleClasses(
                  modelRestrictionMode === 'mapping',
                  'purple',
                )
              "
            >
              {{ t("admin.accounts.modelMapping") }}
            </button>
          </div>

          <!-- Whitelist Mode -->
          <div v-if="modelRestrictionMode === 'whitelist'">
            <ModelWhitelistSelector
              v-model="allowedModels"
              platform="anthropic"
            />
            <p class="edit-account-modal__muted text-xs">
              {{
                t("admin.accounts.selectedModels", {
                  count: allowedModels.length,
                })
              }}
              <span v-if="allowedModels.length === 0">{{
                t("admin.accounts.supportsAllModels")
              }}</span>
            </p>
          </div>

          <!-- Mapping Mode -->
          <div v-else class="space-y-3">
            <div
              v-for="(mapping, index) in modelMappings"
              :key="getModelMappingKey(mapping)"
              class="flex items-center gap-2"
            >
              <input
                v-model="mapping.from"
                type="text"
                class="input flex-1"
                :placeholder="t('admin.accounts.fromModel')"
              />
              <span class="edit-account-modal__muted">→</span>
              <input
                v-model="mapping.to"
                type="text"
                class="input flex-1"
                :placeholder="t('admin.accounts.toModel')"
              />
              <button
                type="button"
                @click="modelMappings.splice(index, 1)"
                :class="getStatusChipClasses(true, 'danger')"
              >
                <Icon name="trash" size="sm" />
              </button>
            </div>
            <button
              type="button"
              @click="modelMappings.push({ from: '', to: '' })"
              class="btn btn-secondary text-sm"
            >
              + {{ t("admin.accounts.addMapping") }}
            </button>
            <!-- Bedrock Preset Mappings -->
            <div class="flex flex-wrap gap-2">
              <button
                v-for="preset in bedrockPresets"
                :key="preset.from"
                type="button"
                @click="
                  modelMappings.push({ from: preset.from, to: preset.to })
                "
                :class="getPresetMappingChipClasses(preset.tone)"
              >
                + {{ preset.label }}
              </button>
            </div>
          </div>
        </div>

        <!-- Pool Mode Section for Bedrock -->
        <div class="form-section">
          <div class="mb-3 flex items-center justify-between">
            <div>
              <label class="input-label mb-0">{{
                t("admin.accounts.poolMode")
              }}</label>
              <p class="edit-account-modal__muted mt-1 text-xs">
                {{ t("admin.accounts.poolModeHint") }}
              </p>
            </div>
            <button
              type="button"
              @click="poolModeEnabled = !poolModeEnabled"
              :class="getSwitchTrackClasses(poolModeEnabled)"
            >
              <span :class="getSwitchThumbClasses(poolModeEnabled)" />
            </button>
          </div>
          <div v-if="poolModeEnabled" :class="getToneNoticeClasses('blue')">
            <p class="text-xs">
              <Icon
                name="exclamationCircle"
                size="sm"
                class="mr-1 inline"
                :stroke-width="2"
              />
              {{ t("admin.accounts.poolModeInfo") }}
            </p>
          </div>
          <div v-if="poolModeEnabled" class="mt-3">
            <label class="input-label">{{
              t("admin.accounts.poolModeRetryCount")
            }}</label>
            <input
              v-model.number="poolModeRetryCount"
              type="number"
              min="0"
              :max="MAX_POOL_MODE_RETRY_COUNT"
              step="1"
              class="input"
            />
            <p class="edit-account-modal__muted mt-1 text-xs">
              {{
                t("admin.accounts.poolModeRetryCountHint", {
                  default: DEFAULT_POOL_MODE_RETRY_COUNT,
                  max: MAX_POOL_MODE_RETRY_COUNT,
                })
              }}
            </p>
          </div>
        </div>
      </div>

      <!-- Antigravity model restriction (applies to all antigravity types) -->
      <!-- Antigravity 只支持模型映射模式，不支持白名单模式 -->
      <div v-if="account.platform === 'antigravity'" class="form-section">
        <label class="input-label">{{
          t("admin.accounts.modelRestriction")
        }}</label>

        <!-- Mapping Mode Only (no toggle for Antigravity) -->
        <div>
          <div :class="['mb-3', getToneNoticeClasses('purple')]">
            <p class="text-xs">{{ t("admin.accounts.mapRequestModels") }}</p>
          </div>

          <div
            v-if="antigravityModelMappings.length > 0"
            class="mb-3 space-y-2"
          >
            <div
              v-for="(mapping, index) in antigravityModelMappings"
              :key="getAntigravityModelMappingKey(mapping)"
              class="space-y-1"
            >
              <div class="flex items-center gap-2">
                <input
                  v-model="mapping.from"
                  type="text"
                  :class="[
                    'input flex-1',
                    !isValidWildcardPattern(mapping.from)
                      ? 'edit-account-modal__input-error'
                      : '',
                    mapping.to.includes('*') ? '' : '',
                  ]"
                  :placeholder="t('admin.accounts.requestModel')"
                />
                <svg
                  class="edit-account-modal__muted h-4 w-4 flex-shrink-0"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M14 5l7 7m0 0l-7 7m7-7H3"
                  />
                </svg>
                <input
                  v-model="mapping.to"
                  type="text"
                  :class="[
                    'input flex-1',
                    mapping.to.includes('*')
                      ? 'edit-account-modal__input-error'
                      : '',
                  ]"
                  :placeholder="t('admin.accounts.actualModel')"
                />
                <button
                  type="button"
                  @click="removeAntigravityModelMapping(index)"
                  :class="getStatusChipClasses(true, 'danger')"
                >
                  <svg
                    class="h-4 w-4"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                    />
                  </svg>
                </button>
              </div>
              <!-- 校验错误提示 -->
              <p
                v-if="!isValidWildcardPattern(mapping.from)"
                class="edit-account-modal__error-text text-xs"
              >
                {{ t("admin.accounts.wildcardOnlyAtEnd") }}
              </p>
              <p
                v-if="mapping.to.includes('*')"
                class="edit-account-modal__error-text text-xs"
              >
                {{ t("admin.accounts.targetNoWildcard") }}
              </p>
            </div>
          </div>

          <button
            type="button"
            @click="addAntigravityModelMapping"
            class="btn btn-secondary mb-3 w-full border-2 border-dashed"
          >
            <svg
              class="mr-1 inline h-4 w-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 4v16m8-8H4"
              />
            </svg>
            {{ t("admin.accounts.addMapping") }}
          </button>

          <div class="flex flex-wrap gap-2">
            <button
              v-for="preset in antigravityPresetMappings"
              :key="preset.label"
              type="button"
              @click="addAntigravityPresetMapping(preset.from, preset.to)"
              :class="getPresetMappingChipClasses(preset.tone)"
            >
              + {{ preset.label }}
            </button>
          </div>
        </div>
      </div>

      <!-- Temp Unschedulable Rules -->
      <div class="form-section space-y-4">
        <div class="mb-3 flex items-center justify-between">
          <div>
            <label class="input-label mb-0">{{
              t("admin.accounts.tempUnschedulable.title")
            }}</label>
            <p class="edit-account-modal__muted mt-1 text-xs">
              {{ t("admin.accounts.tempUnschedulable.hint") }}
            </p>
          </div>
          <button
            type="button"
            @click="tempUnschedEnabled = !tempUnschedEnabled"
            :class="getSwitchTrackClasses(tempUnschedEnabled)"
          >
            <span :class="getSwitchThumbClasses(tempUnschedEnabled)" />
          </button>
        </div>

        <div v-if="tempUnschedEnabled" class="space-y-3">
          <div :class="getToneNoticeClasses('blue')">
            <p class="text-xs">
              <Icon
                name="exclamationTriangle"
                size="sm"
                class="mr-1 inline"
                :stroke-width="2"
              />
              {{ t("admin.accounts.tempUnschedulable.notice") }}
            </p>
          </div>

          <div class="flex flex-wrap gap-2">
            <button
              v-for="preset in tempUnschedPresets"
              :key="preset.label"
              type="button"
              @click="addTempUnschedRule(preset.rule)"
              :class="getStatusChipClasses(false, 'accent')"
            >
              + {{ preset.label }}
            </button>
          </div>

          <div v-if="tempUnschedRules.length > 0" class="space-y-3">
            <div
              v-for="(rule, index) in tempUnschedRules"
              :key="getTempUnschedRuleKey(rule)"
              class="edit-account-modal__config-card edit-account-modal__config-card--compact"
            >
              <div class="mb-2 flex items-center justify-between">
                <span class="edit-account-modal__muted text-xs font-medium">
                  {{
                    t("admin.accounts.tempUnschedulable.ruleIndex", {
                      index: index + 1,
                    })
                  }}
                </span>
                <div class="flex items-center gap-2">
                  <button
                    type="button"
                    :disabled="index === 0"
                    @click="moveTempUnschedRule(index, -1)"
                    class="edit-account-modal__icon-button edit-account-modal__muted transition-colors hover:text-inherit disabled:cursor-not-allowed disabled:opacity-40"
                  >
                    <Icon name="chevronUp" size="sm" :stroke-width="2" />
                  </button>
                  <button
                    type="button"
                    :disabled="index === tempUnschedRules.length - 1"
                    @click="moveTempUnschedRule(index, 1)"
                    class="edit-account-modal__icon-button edit-account-modal__muted transition-colors hover:text-inherit disabled:cursor-not-allowed disabled:opacity-40"
                  >
                    <svg
                      class="h-4 w-4"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M19 9l-7 7-7-7"
                      />
                    </svg>
                  </button>
                  <button
                    type="button"
                    @click="removeTempUnschedRule(index)"
                    :class="getStatusChipClasses(true, 'danger')"
                  >
                    <Icon name="x" size="sm" :stroke-width="2" />
                  </button>
                </div>
              </div>

              <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div>
                  <label class="input-label">{{
                    t("admin.accounts.tempUnschedulable.errorCode")
                  }}</label>
                  <input
                    v-model.number="rule.error_code"
                    type="number"
                    min="100"
                    max="599"
                    class="input"
                    :placeholder="
                      t('admin.accounts.tempUnschedulable.errorCodePlaceholder')
                    "
                  />
                </div>
                <div>
                  <label class="input-label">{{
                    t("admin.accounts.tempUnschedulable.durationMinutes")
                  }}</label>
                  <input
                    v-model.number="rule.duration_minutes"
                    type="number"
                    min="1"
                    class="input"
                    :placeholder="
                      t('admin.accounts.tempUnschedulable.durationPlaceholder')
                    "
                  />
                </div>
                <div class="sm:col-span-2">
                  <label class="input-label">{{
                    t("admin.accounts.tempUnschedulable.keywords")
                  }}</label>
                  <input
                    v-model="rule.keywords"
                    type="text"
                    class="input"
                    :placeholder="
                      t('admin.accounts.tempUnschedulable.keywordsPlaceholder')
                    "
                  />
                  <p class="input-hint">
                    {{ t("admin.accounts.tempUnschedulable.keywordsHint") }}
                  </p>
                </div>
                <div class="sm:col-span-2">
                  <label class="input-label">{{
                    t("admin.accounts.tempUnschedulable.description")
                  }}</label>
                  <input
                    v-model="rule.description"
                    type="text"
                    class="input"
                    :placeholder="
                      t(
                        'admin.accounts.tempUnschedulable.descriptionPlaceholder',
                      )
                    "
                  />
                </div>
              </div>
            </div>
          </div>

          <button
            type="button"
            @click="addTempUnschedRule()"
            class="btn btn-secondary w-full border-2 border-dashed"
          >
            <svg
              class="mr-1 inline h-4 w-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 4v16m8-8H4"
              />
            </svg>
            {{ t("admin.accounts.tempUnschedulable.addRule") }}
          </button>
        </div>
      </div>

      <!-- Intercept Warmup Requests (Anthropic/Antigravity) -->
      <div v-if="showWarmupSection" class="form-section">
        <div class="flex items-center justify-between">
          <div>
            <label class="input-label mb-0">{{
              t("admin.accounts.interceptWarmupRequests")
            }}</label>
            <p class="edit-account-modal__muted mt-1 text-xs">
              {{ t("admin.accounts.interceptWarmupRequestsDesc") }}
            </p>
          </div>
          <button
            type="button"
            @click="interceptWarmupRequests = !interceptWarmupRequests"
            :class="getSwitchTrackClasses(interceptWarmupRequests)"
          >
            <span :class="getSwitchThumbClasses(interceptWarmupRequests)" />
          </button>
        </div>
      </div>

      <div>
        <label class="input-label">{{ t("admin.accounts.proxy") }}</label>
        <ProxySelector v-model="form.proxy_id" :proxies="proxies" />
      </div>

      <div
        class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4 lg:grid-cols-4"
      >
        <div>
          <label class="input-label">{{
            t("admin.accounts.concurrency")
          }}</label>
          <input
            v-model.number="form.concurrency"
            type="number"
            min="1"
            class="input"
            @input="form.concurrency = Math.max(1, form.concurrency || 1)"
          />
        </div>
        <div>
          <label class="input-label">{{
            t("admin.accounts.loadFactor")
          }}</label>
          <input
            v-model.number="form.load_factor"
            type="number"
            min="1"
            class="input"
            :placeholder="String(form.concurrency || 1)"
            @input="form.load_factor = (form.load_factor &amp;&amp; form.load_factor >= 1) ? form.load_factor : null"
          />
          <p class="input-hint">{{ t("admin.accounts.loadFactorHint") }}</p>
        </div>
        <div>
          <label class="input-label">{{ t("admin.accounts.priority") }}</label>
          <input
            v-model.number="form.priority"
            type="number"
            min="1"
            class="input"
            data-tour="account-form-priority"
          />
          <p class="input-hint">{{ t("admin.accounts.priorityHint") }}</p>
        </div>
        <div>
          <label class="input-label">{{
            t("admin.accounts.billingRateMultiplier")
          }}</label>
          <input
            v-model.number="form.rate_multiplier"
            type="number"
            min="0"
            step="0.001"
            class="input"
          />
          <p class="input-hint">
            {{ t("admin.accounts.billingRateMultiplierHint") }}
          </p>
        </div>
      </div>
      <div class="form-section">
        <label class="input-label">{{ t("admin.accounts.expiresAt") }}</label>
        <input v-model="expiresAtInput" type="datetime-local" class="input" />
        <p class="input-hint">{{ t("admin.accounts.expiresAtHint") }}</p>
      </div>

      <!-- OpenAI 自动透传开关（OAuth/API Key） -->
      <div
        v-if="
          account?.platform === 'openai' &&
          (account?.type === 'oauth' || account?.type === 'apikey')
        "
        class="form-section"
      >
        <div class="flex items-center justify-between">
          <div>
            <label class="input-label mb-0">{{
              t("admin.accounts.openai.oauthPassthrough")
            }}</label>
            <p class="edit-account-modal__muted mt-1 text-xs">
              {{ t("admin.accounts.openai.oauthPassthroughDesc") }}
            </p>
          </div>
          <button
            type="button"
            @click="openaiPassthroughEnabled = !openaiPassthroughEnabled"
            :class="getSwitchTrackClasses(openaiPassthroughEnabled)"
          >
            <span :class="getSwitchThumbClasses(openaiPassthroughEnabled)" />
          </button>
        </div>
      </div>

      <!-- OpenAI WS Mode 三态（off/ctx_pool/passthrough） -->
      <div
        v-if="
          account?.platform === 'openai' &&
          (account?.type === 'oauth' || account?.type === 'apikey')
        "
        class="form-section"
      >
        <div class="flex items-center justify-between">
          <div>
            <label class="input-label mb-0">{{
              t("admin.accounts.openai.wsMode")
            }}</label>
            <p class="edit-account-modal__muted mt-1 text-xs">
              {{ t("admin.accounts.openai.wsModeDesc") }}
            </p>
            <p class="edit-account-modal__muted mt-1 text-xs">
              {{ t(openAIWSModeConcurrencyHintKey) }}
            </p>
          </div>
          <div class="w-52">
            <Select
              v-model="openaiResponsesWebSocketV2Mode"
              :options="openAIWSModeOptions"
            />
          </div>
        </div>
      </div>

      <!-- Anthropic API Key 自动透传开关 -->
      <div
        v-if="account?.platform === 'anthropic' && account?.type === 'apikey'"
        class="form-section"
      >
        <div class="flex items-center justify-between">
          <div>
            <label class="input-label mb-0">{{
              t("admin.accounts.anthropic.apiKeyPassthrough")
            }}</label>
            <p class="edit-account-modal__muted mt-1 text-xs">
              {{ t("admin.accounts.anthropic.apiKeyPassthroughDesc") }}
            </p>
          </div>
          <button
            type="button"
            @click="anthropicPassthroughEnabled = !anthropicPassthroughEnabled"
            :class="getSwitchTrackClasses(anthropicPassthroughEnabled)"
          >
            <span :class="getSwitchThumbClasses(anthropicPassthroughEnabled)" />
          </button>
        </div>
      </div>

      <!-- Compatible API / Bedrock 账号配额限制 -->
      <div v-if="showQuotaLimitSection" class="form-section space-y-4">
        <div class="mb-3">
          <h3 class="input-label mb-0 text-base font-semibold">
            {{ t("admin.accounts.quotaLimit") }}
          </h3>
          <p class="edit-account-modal__muted mt-1 text-xs">
            {{ t("admin.accounts.quotaLimitHint") }}
          </p>
        </div>
        <QuotaLimitCard
          :totalLimit="editQuotaLimit"
          :dailyLimit="editQuotaDailyLimit"
          :weeklyLimit="editQuotaWeeklyLimit"
          :dailyResetMode="editDailyResetMode"
          :dailyResetHour="editDailyResetHour"
          :weeklyResetMode="editWeeklyResetMode"
          :weeklyResetDay="editWeeklyResetDay"
          :weeklyResetHour="editWeeklyResetHour"
          :resetTimezone="editResetTimezone"
          @update:totalLimit="editQuotaLimit = $event"
          @update:dailyLimit="editQuotaDailyLimit = $event"
          @update:weeklyLimit="editQuotaWeeklyLimit = $event"
          @update:dailyResetMode="editDailyResetMode = $event"
          @update:dailyResetHour="editDailyResetHour = $event"
          @update:weeklyResetMode="editWeeklyResetMode = $event"
          @update:weeklyResetDay="editWeeklyResetDay = $event"
          @update:weeklyResetHour="editWeeklyResetHour = $event"
          @update:resetTimezone="editResetTimezone = $event"
        />

        <div
          class="space-y-3 rounded-lg border border-gray-200 p-4 dark:border-dark-600"
        >
          <div>
            <h4 class="input-label mb-1">
              {{ t("admin.accounts.quotaNotify.title") }}
            </h4>
            <p class="edit-account-modal__muted text-xs">
              {{ t("admin.accounts.quotaNotify.hint") }}
            </p>
          </div>
          <div class="grid gap-3 md:grid-cols-3">
            <div>
              <label class="input-label">{{
                t("admin.accounts.quotaNotify.daily")
              }}</label>
              <QuotaNotifyToggle
                :enabled="editQuotaNotifyDailyEnabled"
                :threshold="editQuotaNotifyDailyThreshold"
                :threshold-type="editQuotaNotifyDailyThresholdType"
                @update:enabled="editQuotaNotifyDailyEnabled = $event"
                @update:threshold="editQuotaNotifyDailyThreshold = $event"
                @update:thresholdType="
                  editQuotaNotifyDailyThresholdType = $event
                "
              />
            </div>
            <div>
              <label class="input-label">{{
                t("admin.accounts.quotaNotify.weekly")
              }}</label>
              <QuotaNotifyToggle
                :enabled="editQuotaNotifyWeeklyEnabled"
                :threshold="editQuotaNotifyWeeklyThreshold"
                :threshold-type="editQuotaNotifyWeeklyThresholdType"
                @update:enabled="editQuotaNotifyWeeklyEnabled = $event"
                @update:threshold="editQuotaNotifyWeeklyThreshold = $event"
                @update:thresholdType="
                  editQuotaNotifyWeeklyThresholdType = $event
                "
              />
            </div>
            <div>
              <label class="input-label">{{
                t("admin.accounts.quotaNotify.total")
              }}</label>
              <QuotaNotifyToggle
                :enabled="editQuotaNotifyTotalEnabled"
                :threshold="editQuotaNotifyTotalThreshold"
                :threshold-type="editQuotaNotifyTotalThresholdType"
                @update:enabled="editQuotaNotifyTotalEnabled = $event"
                @update:threshold="editQuotaNotifyTotalThreshold = $event"
                @update:thresholdType="
                  editQuotaNotifyTotalThresholdType = $event
                "
              />
            </div>
          </div>
        </div>
      </div>

      <!-- OpenAI OAuth Codex 官方客户端限制开关 -->
      <div
        v-if="account?.platform === 'openai' && account?.type === 'oauth'"
        class="form-section"
      >
        <div class="flex items-center justify-between">
          <div>
            <label class="input-label mb-0">{{
              t("admin.accounts.openai.codexCLIOnly")
            }}</label>
            <p class="edit-account-modal__muted mt-1 text-xs">
              {{ t("admin.accounts.openai.codexCLIOnlyDesc") }}
            </p>
          </div>
          <button
            type="button"
            @click="codexCLIOnlyEnabled = !codexCLIOnlyEnabled"
            :class="getSwitchTrackClasses(codexCLIOnlyEnabled)"
          >
            <span :class="getSwitchThumbClasses(codexCLIOnlyEnabled)" />
          </button>
        </div>
      </div>

      <div>
        <div class="flex items-center justify-between">
          <div>
            <label class="input-label mb-0">{{
              t("admin.accounts.autoPauseOnExpired")
            }}</label>
            <p class="edit-account-modal__muted mt-1 text-xs">
              {{ t("admin.accounts.autoPauseOnExpiredDesc") }}
            </p>
          </div>
          <button
            type="button"
            @click="autoPauseOnExpired = !autoPauseOnExpired"
            :class="getSwitchTrackClasses(autoPauseOnExpired)"
          >
            <span :class="getSwitchThumbClasses(autoPauseOnExpired)" />
          </button>
        </div>
      </div>

      <!-- Quota Control Section (Anthropic OAuth/SetupToken only) -->
      <div
        v-if="
          account?.platform === 'anthropic' &&
          (account?.type === 'oauth' || account?.type === 'setup-token')
        "
        class="form-section space-y-4"
      >
        <div class="mb-3">
          <h3 class="input-label mb-0 text-base font-semibold">
            {{ t("admin.accounts.quotaControl.title") }}
          </h3>
          <p class="edit-account-modal__muted mt-1 text-xs">
            {{ t("admin.accounts.quotaControl.hint") }}
          </p>
        </div>

        <!-- Window Cost Limit -->
        <div class="edit-account-modal__config-card">
          <div class="mb-3 flex items-center justify-between">
            <div>
              <label class="input-label mb-0">{{
                t("admin.accounts.quotaControl.windowCost.label")
              }}</label>
              <p class="edit-account-modal__muted mt-1 text-xs">
                {{ t("admin.accounts.quotaControl.windowCost.hint") }}
              </p>
            </div>
            <button
              type="button"
              @click="windowCostEnabled = !windowCostEnabled"
              :class="getSwitchTrackClasses(windowCostEnabled)"
            >
              <span :class="getSwitchThumbClasses(windowCostEnabled)" />
            </button>
          </div>

          <div
            v-if="windowCostEnabled"
            class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4"
          >
            <div>
              <label class="input-label">{{
                t("admin.accounts.quotaControl.windowCost.limit")
              }}</label>
              <div class="relative">
                <span
                  class="edit-account-modal__muted absolute left-3 top-1/2 -translate-y-1/2"
                  >$</span
                >
                <input
                  v-model.number="windowCostLimit"
                  type="number"
                  min="0"
                  step="1"
                  class="input pl-7"
                  :placeholder="
                    t('admin.accounts.quotaControl.windowCost.limitPlaceholder')
                  "
                />
              </div>
              <p class="input-hint">
                {{ t("admin.accounts.quotaControl.windowCost.limitHint") }}
              </p>
            </div>
            <div>
              <label class="input-label">{{
                t("admin.accounts.quotaControl.windowCost.stickyReserve")
              }}</label>
              <div class="relative">
                <span
                  class="edit-account-modal__muted absolute left-3 top-1/2 -translate-y-1/2"
                  >$</span
                >
                <input
                  v-model.number="windowCostStickyReserve"
                  type="number"
                  min="0"
                  step="1"
                  class="input pl-7"
                  :placeholder="
                    t(
                      'admin.accounts.quotaControl.windowCost.stickyReservePlaceholder',
                    )
                  "
                />
              </div>
              <p class="input-hint">
                {{
                  t("admin.accounts.quotaControl.windowCost.stickyReserveHint")
                }}
              </p>
            </div>
          </div>
        </div>

        <!-- Session Limit -->
        <div class="edit-account-modal__config-card">
          <div class="mb-3 flex items-center justify-between">
            <div>
              <label class="input-label mb-0">{{
                t("admin.accounts.quotaControl.sessionLimit.label")
              }}</label>
              <p class="edit-account-modal__muted mt-1 text-xs">
                {{ t("admin.accounts.quotaControl.sessionLimit.hint") }}
              </p>
            </div>
            <button
              type="button"
              @click="sessionLimitEnabled = !sessionLimitEnabled"
              :class="getSwitchTrackClasses(sessionLimitEnabled)"
            >
              <span :class="getSwitchThumbClasses(sessionLimitEnabled)" />
            </button>
          </div>

          <div
            v-if="sessionLimitEnabled"
            class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4"
          >
            <div>
              <label class="input-label">{{
                t("admin.accounts.quotaControl.sessionLimit.maxSessions")
              }}</label>
              <input
                v-model.number="maxSessions"
                type="number"
                min="1"
                step="1"
                class="input"
                :placeholder="
                  t(
                    'admin.accounts.quotaControl.sessionLimit.maxSessionsPlaceholder',
                  )
                "
              />
              <p class="input-hint">
                {{
                  t("admin.accounts.quotaControl.sessionLimit.maxSessionsHint")
                }}
              </p>
            </div>
            <div>
              <label class="input-label">{{
                t("admin.accounts.quotaControl.sessionLimit.idleTimeout")
              }}</label>
              <div class="relative">
                <input
                  v-model.number="sessionIdleTimeout"
                  type="number"
                  min="1"
                  step="1"
                  class="input pr-12"
                  :placeholder="
                    t(
                      'admin.accounts.quotaControl.sessionLimit.idleTimeoutPlaceholder',
                    )
                  "
                />
                <span
                  class="edit-account-modal__muted absolute right-3 top-1/2 -translate-y-1/2"
                  >{{ t("common.minutes") }}</span
                >
              </div>
              <p class="input-hint">
                {{
                  t("admin.accounts.quotaControl.sessionLimit.idleTimeoutHint")
                }}
              </p>
            </div>
          </div>
        </div>

        <!-- RPM Limit -->
        <div class="edit-account-modal__config-card">
          <div class="mb-3 flex items-center justify-between">
            <div>
              <label class="input-label mb-0">{{
                t("admin.accounts.quotaControl.rpmLimit.label")
              }}</label>
              <p class="edit-account-modal__muted mt-1 text-xs">
                {{ t("admin.accounts.quotaControl.rpmLimit.hint") }}
              </p>
            </div>
            <button
              type="button"
              @click="rpmLimitEnabled = !rpmLimitEnabled"
              :class="getSwitchTrackClasses(rpmLimitEnabled)"
            >
              <span :class="getSwitchThumbClasses(rpmLimitEnabled)" />
            </button>
          </div>

          <div v-if="rpmLimitEnabled" class="space-y-4">
            <div>
              <label class="input-label">{{
                t("admin.accounts.quotaControl.rpmLimit.baseRpm")
              }}</label>
              <input
                v-model.number="baseRpm"
                type="number"
                min="1"
                max="1000"
                step="1"
                class="input"
                :placeholder="
                  t('admin.accounts.quotaControl.rpmLimit.baseRpmPlaceholder')
                "
              />
              <p class="input-hint">
                {{ t("admin.accounts.quotaControl.rpmLimit.baseRpmHint") }}
              </p>
            </div>

            <div>
              <label class="input-label">{{
                t("admin.accounts.quotaControl.rpmLimit.strategy")
              }}</label>
              <div class="flex gap-2">
                <button
                  type="button"
                  @click="rpmStrategy = 'tiered'"
                  :class="
                    getModeToggleClasses(rpmStrategy === 'tiered', 'accent')
                  "
                >
                  <div class="text-center">
                    <div>
                      {{
                        t("admin.accounts.quotaControl.rpmLimit.strategyTiered")
                      }}
                    </div>
                    <div class="mt-0.5 text-[10px] opacity-70">
                      {{
                        t(
                          "admin.accounts.quotaControl.rpmLimit.strategyTieredHint",
                        )
                      }}
                    </div>
                  </div>
                </button>
                <button
                  type="button"
                  @click="rpmStrategy = 'sticky_exempt'"
                  :class="
                    getModeToggleClasses(
                      rpmStrategy === 'sticky_exempt',
                      'accent',
                    )
                  "
                >
                  <div class="text-center">
                    <div>
                      {{
                        t(
                          "admin.accounts.quotaControl.rpmLimit.strategyStickyExempt",
                        )
                      }}
                    </div>
                    <div class="mt-0.5 text-[10px] opacity-70">
                      {{
                        t(
                          "admin.accounts.quotaControl.rpmLimit.strategyStickyExemptHint",
                        )
                      }}
                    </div>
                  </div>
                </button>
              </div>
            </div>

            <div v-if="rpmStrategy === 'tiered'">
              <label class="input-label">{{
                t("admin.accounts.quotaControl.rpmLimit.stickyBuffer")
              }}</label>
              <input
                v-model.number="rpmStickyBuffer"
                type="number"
                min="1"
                step="1"
                class="input"
                :placeholder="
                  t(
                    'admin.accounts.quotaControl.rpmLimit.stickyBufferPlaceholder',
                  )
                "
              />
              <p class="input-hint">
                {{ t("admin.accounts.quotaControl.rpmLimit.stickyBufferHint") }}
              </p>
            </div>
          </div>

          <!-- 用户消息限速模式（独立于 RPM 开关，始终可见） -->
          <div class="mt-4">
            <label class="input-label">{{
              t("admin.accounts.quotaControl.rpmLimit.userMsgQueue")
            }}</label>
            <p class="edit-account-modal__muted mt-1 mb-2 text-xs">
              {{ t("admin.accounts.quotaControl.rpmLimit.userMsgQueueHint") }}
            </p>
            <div class="flex space-x-2">
              <button
                type="button"
                v-for="opt in umqModeOptions"
                :key="opt.value"
                @click="userMsgQueueMode = opt.value"
                :class="[
                  'edit-account-modal__umq-option edit-account-modal__umq-option-control text-sm transition-colors',
                  userMsgQueueMode === opt.value
                    ? 'edit-account-modal__umq-option--selected'
                    : 'edit-account-modal__umq-option--idle',
                ]"
              >
                {{ opt.label }}
              </button>
            </div>
          </div>
        </div>

        <!-- TLS Fingerprint -->
        <div class="edit-account-modal__config-card">
          <div class="flex items-center justify-between">
            <div>
              <label class="input-label mb-0">{{
                t("admin.accounts.quotaControl.tlsFingerprint.label")
              }}</label>
              <p class="edit-account-modal__muted mt-1 text-xs">
                {{ t("admin.accounts.quotaControl.tlsFingerprint.hint") }}
              </p>
            </div>
            <button
              type="button"
              @click="tlsFingerprintEnabled = !tlsFingerprintEnabled"
              :class="getSwitchTrackClasses(tlsFingerprintEnabled)"
            >
              <span :class="getSwitchThumbClasses(tlsFingerprintEnabled)" />
            </button>
          </div>
          <!-- Profile selector -->
          <div v-if="tlsFingerprintEnabled" class="mt-3">
            <select v-model="tlsFingerprintProfileId" class="input">
              <option :value="null">
                {{
                  t("admin.accounts.quotaControl.tlsFingerprint.defaultProfile")
                }}
              </option>
              <option v-if="tlsFingerprintProfiles.length > 0" :value="-1">
                {{
                  t("admin.accounts.quotaControl.tlsFingerprint.randomProfile")
                }}
              </option>
              <option
                v-for="p in tlsFingerprintProfiles"
                :key="p.id"
                :value="p.id"
              >
                {{ p.name }}
              </option>
            </select>
          </div>
        </div>

        <!-- Session ID Masking -->
        <div class="edit-account-modal__config-card">
          <div class="flex items-center justify-between">
            <div>
              <label class="input-label mb-0">{{
                t("admin.accounts.quotaControl.sessionIdMasking.label")
              }}</label>
              <p class="edit-account-modal__muted mt-1 text-xs">
                {{ t("admin.accounts.quotaControl.sessionIdMasking.hint") }}
              </p>
            </div>
            <button
              type="button"
              @click="sessionIdMaskingEnabled = !sessionIdMaskingEnabled"
              :class="getSwitchTrackClasses(sessionIdMaskingEnabled)"
            >
              <span :class="getSwitchThumbClasses(sessionIdMaskingEnabled)" />
            </button>
          </div>
        </div>

        <!-- Cache TTL Override -->
        <div class="edit-account-modal__config-card">
          <div class="flex items-center justify-between">
            <div>
              <label class="input-label mb-0">{{
                t("admin.accounts.quotaControl.cacheTTLOverride.label")
              }}</label>
              <p class="edit-account-modal__muted mt-1 text-xs">
                {{ t("admin.accounts.quotaControl.cacheTTLOverride.hint") }}
              </p>
            </div>
            <button
              type="button"
              @click="cacheTTLOverrideEnabled = !cacheTTLOverrideEnabled"
              :class="getSwitchTrackClasses(cacheTTLOverrideEnabled)"
            >
              <span :class="getSwitchThumbClasses(cacheTTLOverrideEnabled)" />
            </button>
          </div>
          <div v-if="cacheTTLOverrideEnabled" class="mt-3">
            <label class="input-label text-xs">{{
              t("admin.accounts.quotaControl.cacheTTLOverride.target")
            }}</label>
            <select
              v-model="cacheTTLOverrideTarget"
              class="input mt-1 block w-full text-sm"
            >
              <option value="5m">5m</option>
              <option value="1h">1h</option>
            </select>
            <p class="edit-account-modal__muted mt-1 text-xs">
              {{ t("admin.accounts.quotaControl.cacheTTLOverride.targetHint") }}
            </p>
          </div>
        </div>

        <!-- Custom Base URL Relay -->
        <div class="edit-account-modal__config-card">
          <div class="flex items-center justify-between">
            <div>
              <label class="input-label mb-0">{{
                t("admin.accounts.quotaControl.customBaseUrl.label")
              }}</label>
              <p class="edit-account-modal__muted mt-1 text-xs">
                {{ t("admin.accounts.quotaControl.customBaseUrl.hint") }}
              </p>
            </div>
            <button
              type="button"
              @click="customBaseUrlEnabled = !customBaseUrlEnabled"
              :class="getSwitchTrackClasses(customBaseUrlEnabled)"
            >
              <span :class="getSwitchThumbClasses(customBaseUrlEnabled)" />
            </button>
          </div>
          <div v-if="customBaseUrlEnabled" class="mt-3">
            <input
              v-model="customBaseUrl"
              type="text"
              class="input"
              :placeholder="
                t('admin.accounts.quotaControl.customBaseUrl.urlHint')
              "
            />
          </div>
        </div>
      </div>

      <div class="form-section">
        <div>
          <label class="input-label">{{ t("common.status") }}</label>
          <Select v-model="form.status" :options="statusOptions" />
        </div>

        <!-- Mixed Scheduling (only for antigravity accounts, read-only in edit mode) -->
        <div
          v-if="account?.platform === 'antigravity'"
          class="flex items-center gap-2"
        >
          <label class="flex cursor-not-allowed items-center gap-2 opacity-60">
            <input
              type="checkbox"
              v-model="mixedScheduling"
              disabled
              class="edit-account-modal__checkbox h-4 w-4 cursor-not-allowed"
            />
            <span class="edit-account-modal__choice-text text-sm font-medium">
              {{ t("admin.accounts.mixedScheduling") }}
            </span>
          </label>
          <div class="group relative">
            <span
              class="edit-account-modal__status-chip edit-account-modal__status-chip--idle inline-flex h-4 w-4 cursor-help items-center justify-center rounded-full text-xs"
            >
              ?
            </span>
            <!-- Tooltip（向下显示避免被弹窗裁剪） -->
            <div
              class="edit-account-modal__tooltip pointer-events-none absolute left-0 top-full z-[100] mt-1.5 text-xs opacity-0 transition-opacity group-hover:opacity-100"
            >
              {{ t("admin.accounts.mixedSchedulingTooltip") }}
              <div
                class="edit-account-modal__tooltip-arrow absolute bottom-full left-3 border-4 border-transparent"
              ></div>
            </div>
          </div>
        </div>
        <div
          v-if="account?.platform === 'antigravity'"
          class="mt-3 flex items-center gap-2"
        >
          <label class="flex cursor-pointer items-center gap-2">
            <input
              type="checkbox"
              v-model="allowOverages"
              class="edit-account-modal__checkbox h-4 w-4"
            />
            <span class="edit-account-modal__choice-text text-sm font-medium">
              {{ t("admin.accounts.allowOverages") }}
            </span>
          </label>
          <div class="group relative">
            <span
              class="edit-account-modal__status-chip edit-account-modal__status-chip--idle inline-flex h-4 w-4 cursor-help items-center justify-center rounded-full text-xs"
            >
              ?
            </span>
            <div
              class="edit-account-modal__tooltip pointer-events-none absolute left-0 top-full z-[100] mt-1.5 text-xs opacity-0 transition-opacity group-hover:opacity-100"
            >
              {{ t("admin.accounts.allowOveragesTooltip") }}
              <div
                class="edit-account-modal__tooltip-arrow absolute bottom-full left-3 border-4 border-transparent"
              ></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Group Selection - 仅标准模式显示 -->
      <GroupSelector
        v-if="!authStore.isSimpleMode"
        v-model="form.group_ids"
        :groups="groups"
        :platform="account?.platform"
        :mixed-scheduling="mixedScheduling"
        data-tour="account-form-groups"
      />
    </form>

    <template #footer>
      <div v-if="account" class="flex justify-end gap-3">
        <button @click="handleClose" type="button" class="btn btn-secondary">
          {{ t("common.cancel") }}
        </button>
        <button
          type="submit"
          form="edit-account-form"
          :disabled="submitting"
          class="btn btn-primary"
          data-tour="account-form-submit"
        >
          <svg
            v-if="submitting"
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
          {{ submitting ? t("admin.accounts.updating") : t("common.update") }}
        </button>
      </div>
    </template>
  </BaseDialog>

  <!-- Mixed Channel Warning Dialog -->
  <ConfirmDialog
    :show="showMixedChannelWarning"
    :title="t('admin.accounts.mixedChannelWarningTitle')"
    :message="mixedChannelWarningMessageText"
    :confirm-text="t('common.confirm')"
    :cancel-text="t('common.cancel')"
    :danger="true"
    @confirm="handleMixedChannelConfirm"
    @cancel="handleMixedChannelCancel"
  />
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useAppStore } from "@/stores/app";
import { useAuthStore } from "@/stores/auth";
import { adminAPI } from "@/api/admin";
import type {
  Account,
  AccountType,
  Proxy,
  AdminGroup,
  CheckMixedChannelResponse,
} from "@/types";
import BaseDialog from "@/components/common/BaseDialog.vue";
import ConfirmDialog from "@/components/common/ConfirmDialog.vue";
import Select from "@/components/common/Select.vue";
import Icon from "@/components/icons/Icon.vue";
import ProxySelector from "@/components/common/ProxySelector.vue";
import GroupSelector from "@/components/common/GroupSelector.vue";
import ModelWhitelistSelector from "@/components/account/ModelWhitelistSelector.vue";
import QuotaLimitCard from "@/components/account/QuotaLimitCard.vue";
import QuotaNotifyToggle from "@/components/account/QuotaNotifyToggle.vue";
import {
  buildCompatibleBaseUrlPresets,
  buildAccountOpenAIWSModeOptions,
  buildAccountQuotaExtra,
  buildAccountTempUnschedPresets,
  buildAccountUmqModeOptions,
  buildEditAccountBasePayload,
  buildMixedChannelDetails,
  createDefaultEditAccountForm,
  hydrateEditAccountForm,
  needsMixedChannelCheck,
  resolveAccountApiKeyPlaceholder,
  resolveAccountBaseUrlHint,
  resolveAccountBaseUrlPlaceholder,
  resolveMixedChannelWarningMessage,
  type EditAccountForm,
} from "@/components/account/accountModalShared";
import {
  getAccountModalModeToggleClasses,
  getAccountModalStatusChipClasses,
  getAccountModalSwitchThumbClasses,
  getAccountModalSwitchTrackClasses,
  getEditToneNoticeClasses,
} from "@/components/account/accountModalClasses";
import {
  buildEditableBedrockCredentials,
  buildEditableCompatibleCredentials,
  buildUpdatedAnthropicAPIKeyExtra,
  buildUpdatedAnthropicQuotaControlExtra,
  buildUpdatedAntigravityExtra,
  buildUpdatedOpenAIExtra,
  createEmptyModelRestrictionState,
  deriveAntigravityModelMappings,
  deriveModelRestrictionStateFromMapping,
  deriveOpenAIExtraState,
} from "@/components/account/editAccountModalHelpers";
import {
  accountMutationProfileHasSection,
  resolveAccountMutationProfile,
} from "@/components/account/accountMutationProfiles";
import {
  createTempUnschedRule,
  DEFAULT_POOL_MODE_RETRY_COUNT,
  getDefaultBaseURL,
  loadTempUnschedRuleState,
  MAX_POOL_MODE_RETRY_COUNT,
  moveItemInPlace,
  normalizePoolModeRetryCount,
  replaceAntigravityModelMapping,
  replaceBuiltModelMapping,
  type ModelMapping,
  type TempUnschedRuleForm,
} from "@/components/account/credentialsBuilder";
import {
  appendEmptyModelMapping,
  appendPresetModelMapping,
  applySharedAccountCredentialsState,
  confirmCustomErrorCodeSelection,
  removeModelMappingAt,
} from "@/components/account/accountModalInteractions";
import {
  formatDateTime,
  formatDateTimeLocalInput,
  parseDateTimeLocalInput,
} from "@/utils/format";
import { normalizeGrokSessionToken } from "@/utils/grokSessionToken";
import {
  getGrokAccountRuntime,
  getGrokProbeOutcome,
} from "@/utils/grokAccountRuntime";
import { createStableObjectKeyResolver } from "@/utils/stableObjectKey";
import {
  OPENAI_WS_MODE_OFF,
  resolveOpenAIWSModeConcurrencyHintKey,
  type OpenAIWSMode,
} from "@/utils/openaiWsMode";
import { resolveRequestErrorMessage } from "@/utils/requestError";
import {
  ensureModelCatalogLoaded,
  getPresetMappingChipClasses,
  getPresetMappingsByPlatform,
  commonErrorCodes,
  isValidWildcardPattern,
} from "@/composables/useModelWhitelist";

interface Props {
  show: boolean;
  account: Account | null;
  proxies: Proxy[];
  groups: AdminGroup[];
}

const props = defineProps<Props>();
const emit = defineEmits<{
  close: [];
  updated: [account: Account];
}>();

const { t } = useI18n();
const appStore = useAppStore();
const authStore = useAuthStore();

// Platform-specific hint for Base URL
const baseUrlHint = computed(() => {
  return resolveAccountBaseUrlHint(props.account?.platform, t);
});

const baseUrlPlaceholder = computed(() => {
  return resolveAccountBaseUrlPlaceholder(props.account?.platform, t);
});

const apiKeyPlaceholder = computed(() => {
  return resolveAccountApiKeyPlaceholder(props.account?.platform, t);
});

const compatibleBaseUrlPresets = computed(() => {
  return buildCompatibleBaseUrlPresets(props.account?.platform, t);
});

const grokRuntimeState = computed(() => getGrokAccountRuntime(props.account));
const emptyRuntimeValue = "-";

const antigravityPresetMappings = computed(() =>
  getPresetMappingsByPlatform("antigravity"),
);
const bedrockPresets = computed(() => getPresetMappingsByPlatform("bedrock"));

// State
const submitting = ref(false);
const editBaseUrl = ref(getDefaultBaseURL("anthropic"));
const editApiKey = ref("");
const editSessionToken = ref("");
// Bedrock credentials
const editBedrockAccessKeyId = ref("");
const editBedrockSecretAccessKey = ref("");
const editBedrockSessionToken = ref("");
const editBedrockRegion = ref("");
const editBedrockForceGlobal = ref(false);
const editBedrockApiKeyValue = ref("");
const isBedrockAPIKeyMode = computed(
  () =>
    props.account?.type === "bedrock" &&
    (props.account?.credentials as Record<string, unknown>)?.auth_mode ===
      "apikey",
);
const modelMappings = ref<ModelMapping[]>([]);
const modelRestrictionMode = ref<"whitelist" | "mapping">("whitelist");
const allowedModels = ref<string[]>([]);
const poolModeEnabled = ref(false);
const poolModeRetryCount = ref(DEFAULT_POOL_MODE_RETRY_COUNT);
const customErrorCodesEnabled = ref(false);
const selectedErrorCodes = ref<number[]>([]);
const customErrorCodeInput = ref<number | null>(null);
const interceptWarmupRequests = ref(false);
const autoPauseOnExpired = ref(false);
const mixedScheduling = ref(false); // For antigravity accounts: enable mixed scheduling
const allowOverages = ref(false); // For antigravity accounts: enable AI Credits overages
const antigravityModelRestrictionMode = ref<"whitelist" | "mapping">(
  "whitelist",
);
const antigravityWhitelistModels = ref<string[]>([]);
const antigravityModelMappings = ref<ModelMapping[]>([]);
const tempUnschedEnabled = ref(false);
const tempUnschedRules = ref<TempUnschedRuleForm[]>([]);
const getModelMappingKey =
  createStableObjectKeyResolver<ModelMapping>("edit-model-mapping");
const getAntigravityModelMappingKey =
  createStableObjectKeyResolver<ModelMapping>("edit-antigravity-model-mapping");
const getTempUnschedRuleKey =
  createStableObjectKeyResolver<TempUnschedRuleForm>("edit-temp-unsched-rule");

const showMixedChannelWarning = ref(false);
const mixedChannelWarningDetails = ref<{
  groupName: string;
  currentPlatform: string;
  otherPlatform: string;
} | null>(null);
const mixedChannelWarningRawMessage = ref("");
const mixedChannelWarningAction = ref<(() => Promise<void>) | null>(null);
const antigravityMixedChannelConfirmed = ref(false);

// Quota control state (Anthropic OAuth/SetupToken only)
const windowCostEnabled = ref(false);
const windowCostLimit = ref<number | null>(null);
const windowCostStickyReserve = ref<number | null>(null);
const sessionLimitEnabled = ref(false);
const maxSessions = ref<number | null>(null);
const sessionIdleTimeout = ref<number | null>(null);
const rpmLimitEnabled = ref(false);
const baseRpm = ref<number | null>(null);
const rpmStrategy = ref<"tiered" | "sticky_exempt">("tiered");
const rpmStickyBuffer = ref<number | null>(null);
const userMsgQueueMode = ref("");
const umqModeOptions = computed(() => buildAccountUmqModeOptions(t));
const tlsFingerprintEnabled = ref(false);
const tlsFingerprintProfileId = ref<number | null>(null);
const tlsFingerprintProfiles = ref<{ id: number; name: string }[]>([]);
const sessionIdMaskingEnabled = ref(false);
const cacheTTLOverrideEnabled = ref(false);
const cacheTTLOverrideTarget = ref<string>("5m");
const customBaseUrlEnabled = ref(false);
const customBaseUrl = ref("");

// OpenAI 自动透传开关（OAuth/API Key）
const openaiPassthroughEnabled = ref(false);
const openaiOAuthResponsesWebSocketV2Mode =
  ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF);
const openaiAPIKeyResponsesWebSocketV2Mode =
  ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF);
const codexCLIOnlyEnabled = ref(false);
const anthropicPassthroughEnabled = ref(false);
const editQuotaLimit = ref<number | null>(null);
const editQuotaDailyLimit = ref<number | null>(null);
const editQuotaWeeklyLimit = ref<number | null>(null);
const editDailyResetMode = ref<"rolling" | "fixed" | null>(null);
const editDailyResetHour = ref<number | null>(null);
const editWeeklyResetMode = ref<"rolling" | "fixed" | null>(null);
const editWeeklyResetDay = ref<number | null>(null);
const editWeeklyResetHour = ref<number | null>(null);
const editResetTimezone = ref<string | null>(null);
const editQuotaNotifyDailyEnabled = ref<boolean | null>(null);
const editQuotaNotifyDailyThreshold = ref<number | null>(null);
const editQuotaNotifyDailyThresholdType = ref<"fixed" | "percentage" | null>(
  null,
);
const editQuotaNotifyWeeklyEnabled = ref<boolean | null>(null);
const editQuotaNotifyWeeklyThreshold = ref<number | null>(null);
const editQuotaNotifyWeeklyThresholdType = ref<"fixed" | "percentage" | null>(
  null,
);
const editQuotaNotifyTotalEnabled = ref<boolean | null>(null);
const editQuotaNotifyTotalThreshold = ref<number | null>(null);
const editQuotaNotifyTotalThresholdType = ref<"fixed" | "percentage" | null>(
  null,
);
const openAIWSModeOptions = computed(() => buildAccountOpenAIWSModeOptions(t));
const openaiResponsesWebSocketV2Mode = computed({
  get: () => {
    if (props.account?.type === "apikey") {
      return openaiAPIKeyResponsesWebSocketV2Mode.value;
    }
    return openaiOAuthResponsesWebSocketV2Mode.value;
  },
  set: (mode: OpenAIWSMode) => {
    if (props.account?.type === "apikey") {
      openaiAPIKeyResponsesWebSocketV2Mode.value = mode;
      return;
    }
    openaiOAuthResponsesWebSocketV2Mode.value = mode;
  },
});
const openAIWSModeConcurrencyHintKey = computed(() =>
  resolveOpenAIWSModeConcurrencyHintKey(openaiResponsesWebSocketV2Mode.value),
);
const isOpenAIModelRestrictionDisabled = computed(
  () => props.account?.platform === "openai" && openaiPassthroughEnabled.value,
);

const mutationProfile = computed(() => {
  const account = props.account;
  return account
    ? resolveAccountMutationProfile(account.platform, account.type)
    : null;
});

const showCompatibleCredentialsForm = computed(() => {
  return accountMutationProfileHasSection(
    mutationProfile.value,
    "compatible-credentials",
  );
});

const grokTierLabel = computed(() => {
  const tier = grokRuntimeState.value?.tier.normalized ?? "unknown";
  return t(`admin.accounts.grok.runtime.tiers.${tier}`);
});

const grokTierChipClass = computed(() => {
  switch (grokRuntimeState.value?.tier.normalized) {
    case "basic":
      return "theme-chip theme-chip--compact theme-chip--info";
    case "heavy":
      return "theme-chip theme-chip--compact theme-chip--brand-orange";
    case "super":
      return "theme-chip theme-chip--compact theme-chip--brand-purple";
    default:
      return "theme-chip theme-chip--compact theme-chip--neutral";
  }
});

const grokAuthModeLabel = computed(() => {
  const mode = grokRuntimeState.value?.authMode;
  return mode
    ? t(`admin.accounts.grok.runtime.authModes.${mode}`)
    : emptyRuntimeValue;
});

const grokTierSourceDisplay = computed(
  () => grokRuntimeState.value?.tier.source ?? emptyRuntimeValue,
);
const grokTierConfidenceDisplay = computed(() => {
  const confidence = grokRuntimeState.value?.tier.confidence;
  return confidence === null || confidence === undefined
    ? emptyRuntimeValue
    : confidence.toFixed(2);
});
const grokCapabilities = computed(
  () => grokRuntimeState.value?.capabilities.operations ?? [],
);
const grokModels = computed(
  () => grokRuntimeState.value?.capabilities.models ?? [],
);
const grokQuotaWindows = computed(
  () =>
    grokRuntimeState.value?.quotaWindows.filter((window) => window.hasSignal) ??
    [],
);
const grokProbeOutcome = computed(() =>
  getGrokProbeOutcome(grokRuntimeState.value?.sync),
);

const grokProbeStatusDisplay = computed(() => {
  const sync = grokRuntimeState.value?.sync;
  if (!sync) {
    return emptyRuntimeValue;
  }

  if (grokProbeOutcome.value === "healthy") {
    return t("admin.accounts.grok.runtime.probeHealthy");
  }
  if (grokProbeOutcome.value === "failed") {
    return t("admin.accounts.grok.runtime.probeFailed");
  }
  return emptyRuntimeValue;
});
const grokProbeErrorDisplay = computed(() => {
  const sync = grokRuntimeState.value?.sync;
  if (!sync) {
    return emptyRuntimeValue;
  }

  const code = sync.lastProbeStatusCode;
  if (grokProbeOutcome.value === "healthy") {
    return t("common.success");
  }
  if (grokProbeOutcome.value === "failed") {
    return code !== null
      ? t("admin.accounts.grok.runtime.probeFailedWithCode", { code })
      : t("admin.accounts.grok.runtime.probeFailedShort");
  }

  return emptyRuntimeValue;
});
const grokRuntimeErrorDisplay = computed(
  () => grokRuntimeState.value?.runtime.lastFailReason ?? emptyRuntimeValue,
);

const showQuotaLimitSection = computed(() => {
  return accountMutationProfileHasSection(mutationProfile.value, "quota-limits");
});

const showWarmupSection = computed(() => {
  return accountMutationProfileHasSection(mutationProfile.value, "warmup");
});

watch(
  () => props.account?.platform,
  (platform) => {
    if (platform === "grok") {
      void ensureModelCatalogLoaded(platform);
    }
  },
  { immediate: true },
);

// Computed: current preset mappings based on platform
const presetMappings = computed(() =>
  getPresetMappingsByPlatform(props.account?.platform || "anthropic"),
);
const tempUnschedPresets = computed(() => buildAccountTempUnschedPresets(t));

// Computed: default base URL based on platform
const defaultBaseUrl = computed(() => {
  return getDefaultBaseURL(props.account?.platform || "anthropic");
});

const mixedChannelWarningMessageText = computed(() => {
  return resolveMixedChannelWarningMessage({
    details: mixedChannelWarningDetails.value,
    rawMessage: mixedChannelWarningRawMessage.value,
    t,
  });
});

const getModeToggleClasses = (isSelected: boolean, tone: "accent" | "purple" | "danger") =>
  getAccountModalModeToggleClasses("edit-account-modal", isSelected, tone);

function formatRuntimeValue(value: string | null | undefined) {
  if (!value) {
    return emptyRuntimeValue;
  }
  const formatted = formatDateTime(value);
  return formatted || value;
}

const getSwitchTrackClasses = (isEnabled: boolean) =>
  getAccountModalSwitchTrackClasses("edit-account-modal", isEnabled);
const getSwitchThumbClasses = (isEnabled: boolean) =>
  getAccountModalSwitchThumbClasses("edit-account-modal", isEnabled);
const getStatusChipClasses = (
  isSelected: boolean,
  tone: "accent" | "purple" | "danger" = "danger",
) => getAccountModalStatusChipClasses("edit-account-modal", isSelected, tone);
const getToneNoticeClasses = getEditToneNoticeClasses;

const resetMixedChannelDialogState = () => {
  showMixedChannelWarning.value = false;
  mixedChannelWarningDetails.value = null;
  mixedChannelWarningRawMessage.value = "";
  mixedChannelWarningAction.value = null;
};

const applySharedEditCredentialsState = (
  credentials: Record<string, unknown>,
) => {
  return applySharedAccountCredentialsState(credentials, {
    interceptWarmupRequests: interceptWarmupRequests.value,
    tempUnschedEnabled: tempUnschedEnabled.value,
    tempUnschedRules: tempUnschedRules.value,
    showError: appStore.showError,
    t,
  });
};

const getAccountCredentials = () =>
  (props.account?.credentials as Record<string, unknown>) || {};

const getAccountExtra = () =>
  (props.account?.extra as Record<string, unknown>) || {};

const getPendingCredentials = (updatePayload: Record<string, unknown>) =>
  (updatePayload.credentials as Record<string, unknown>) ||
  getAccountCredentials();

const getPendingExtra = (updatePayload: Record<string, unknown>) =>
  (updatePayload.extra as Record<string, unknown>) || getAccountExtra();

const applyModelRestrictionState = (rawMapping: unknown) => {
  const nextState = deriveModelRestrictionStateFromMapping(rawMapping);
  modelRestrictionMode.value = nextState.mode;
  allowedModels.value = nextState.allowedModels;
  modelMappings.value = nextState.modelMappings;
};

const resetModelRestrictionState = () => {
  const nextState = createEmptyModelRestrictionState();
  modelRestrictionMode.value = nextState.mode;
  allowedModels.value = nextState.allowedModels;
  modelMappings.value = nextState.modelMappings;
};

const syncAntigravityModelRestrictionState = (
  credentials: Record<string, unknown> | undefined,
) => {
  antigravityModelRestrictionMode.value = "mapping";
  antigravityWhitelistModels.value = [];
  antigravityModelMappings.value = deriveAntigravityModelMappings(credentials);
};

const syncOpenAIExtraState = (
  accountType: AccountType,
  extra: Record<string, unknown> | undefined,
) => {
  const nextState = deriveOpenAIExtraState(accountType, extra);
  openaiPassthroughEnabled.value = nextState.openaiPassthroughEnabled;
  openaiOAuthResponsesWebSocketV2Mode.value =
    nextState.openaiOAuthResponsesWebSocketV2Mode;
  openaiAPIKeyResponsesWebSocketV2Mode.value =
    nextState.openaiAPIKeyResponsesWebSocketV2Mode;
  codexCLIOnlyEnabled.value = nextState.codexCLIOnlyEnabled;
};

const form = reactive<EditAccountForm>(createDefaultEditAccountForm());

const statusOptions = computed(() => {
  const options = [
    { value: "active", label: t("common.active") },
    { value: "inactive", label: t("common.inactive") },
  ];
  if (form.status === "error") {
    options.push({ value: "error", label: t("admin.accounts.status.error") });
  }
  return options;
});

const expiresAtInput = computed({
  get: () => formatDateTimeLocal(form.expires_at),
  set: (value: string) => {
    form.expires_at = parseDateTimeLocal(value);
  },
});

// Watchers
const syncFormFromAccount = (newAccount: Account | null) => {
  if (!newAccount) {
    return;
  }
  antigravityMixedChannelConfirmed.value = false;
  resetMixedChannelDialogState();
  hydrateEditAccountForm(form, newAccount);

  // Load intercept warmup requests setting (applies to all account types)
  const credentials = newAccount.credentials as
    | Record<string, unknown>
    | undefined;
  const platformDefaultUrl = getDefaultBaseURL(newAccount.platform);
  interceptWarmupRequests.value =
    credentials?.intercept_warmup_requests === true;
  autoPauseOnExpired.value = newAccount.auto_pause_on_expired === true;
  editBaseUrl.value = platformDefaultUrl;
  editApiKey.value = "";
  editSessionToken.value = "";
  resetModelRestrictionState();
  poolModeEnabled.value = false;
  poolModeRetryCount.value = DEFAULT_POOL_MODE_RETRY_COUNT;
  customErrorCodesEnabled.value = false;
  selectedErrorCodes.value = [];
  customErrorCodeInput.value = null;

  // Load mixed scheduling setting (only for antigravity accounts)
  mixedScheduling.value = false;
  allowOverages.value = false;
  const extra = newAccount.extra as Record<string, unknown> | undefined;
  mixedScheduling.value = extra?.mixed_scheduling === true;
  allowOverages.value = extra?.allow_overages === true;

  // Load OpenAI passthrough toggle (OpenAI OAuth/API Key)
  openaiPassthroughEnabled.value = false;
  openaiOAuthResponsesWebSocketV2Mode.value = OPENAI_WS_MODE_OFF;
  openaiAPIKeyResponsesWebSocketV2Mode.value = OPENAI_WS_MODE_OFF;
  codexCLIOnlyEnabled.value = false;
  anthropicPassthroughEnabled.value = false;
  if (
    newAccount.platform === "openai" &&
    (newAccount.type === "oauth" || newAccount.type === "apikey")
  ) {
    syncOpenAIExtraState(newAccount.type, extra);
  }
  if (newAccount.platform === "anthropic" && newAccount.type === "apikey") {
    anthropicPassthroughEnabled.value = extra?.anthropic_passthrough === true;
  }

  // Load quota limit for compatible key/upstream and bedrock accounts.
  if (
    newAccount.type === "apikey" ||
    newAccount.type === "upstream" ||
    newAccount.type === "bedrock"
  ) {
    const quotaVal = extra?.quota_limit as number | undefined;
    editQuotaLimit.value = quotaVal && quotaVal > 0 ? quotaVal : null;
    const dailyVal = extra?.quota_daily_limit as number | undefined;
    editQuotaDailyLimit.value = dailyVal && dailyVal > 0 ? dailyVal : null;
    const weeklyVal = extra?.quota_weekly_limit as number | undefined;
    editQuotaWeeklyLimit.value = weeklyVal && weeklyVal > 0 ? weeklyVal : null;
    // Load quota reset mode config
    editDailyResetMode.value =
      (extra?.quota_daily_reset_mode as "rolling" | "fixed") || null;
    editDailyResetHour.value =
      (extra?.quota_daily_reset_hour as number) ?? null;
    editWeeklyResetMode.value =
      (extra?.quota_weekly_reset_mode as "rolling" | "fixed") || null;
    editWeeklyResetDay.value =
      (extra?.quota_weekly_reset_day as number) ?? null;
    editWeeklyResetHour.value =
      (extra?.quota_weekly_reset_hour as number) ?? null;
    editResetTimezone.value = (extra?.quota_reset_timezone as string) || null;
    editQuotaNotifyDailyEnabled.value =
      (extra?.quota_notify_daily_enabled as boolean) || null;
    editQuotaNotifyDailyThreshold.value =
      (extra?.quota_notify_daily_threshold as number) ?? null;
    editQuotaNotifyDailyThresholdType.value =
      (extra?.quota_notify_daily_threshold_type as "fixed" | "percentage") ||
      null;
    editQuotaNotifyWeeklyEnabled.value =
      (extra?.quota_notify_weekly_enabled as boolean) || null;
    editQuotaNotifyWeeklyThreshold.value =
      (extra?.quota_notify_weekly_threshold as number) ?? null;
    editQuotaNotifyWeeklyThresholdType.value =
      (extra?.quota_notify_weekly_threshold_type as "fixed" | "percentage") ||
      null;
    editQuotaNotifyTotalEnabled.value =
      (extra?.quota_notify_total_enabled as boolean) || null;
    editQuotaNotifyTotalThreshold.value =
      (extra?.quota_notify_total_threshold as number) ?? null;
    editQuotaNotifyTotalThresholdType.value =
      (extra?.quota_notify_total_threshold_type as "fixed" | "percentage") ||
      null;
  } else {
    editQuotaLimit.value = null;
    editQuotaDailyLimit.value = null;
    editQuotaWeeklyLimit.value = null;
    editDailyResetMode.value = null;
    editDailyResetHour.value = null;
    editWeeklyResetMode.value = null;
    editWeeklyResetDay.value = null;
    editWeeklyResetHour.value = null;
    editResetTimezone.value = null;
    editQuotaNotifyDailyEnabled.value = null;
    editQuotaNotifyDailyThreshold.value = null;
    editQuotaNotifyDailyThresholdType.value = null;
    editQuotaNotifyWeeklyEnabled.value = null;
    editQuotaNotifyWeeklyThreshold.value = null;
    editQuotaNotifyWeeklyThresholdType.value = null;
    editQuotaNotifyTotalEnabled.value = null;
    editQuotaNotifyTotalThreshold.value = null;
    editQuotaNotifyTotalThresholdType.value = null;
  }

  // Load antigravity model mapping (Antigravity 只支持映射模式)
  if (newAccount.platform === "antigravity") {
    syncAntigravityModelRestrictionState(
      newAccount.credentials as Record<string, unknown> | undefined,
    );
  } else {
    antigravityModelRestrictionMode.value = "mapping";
    antigravityWhitelistModels.value = [];
    antigravityModelMappings.value = [];
  }

  // Load quota control settings (Anthropic OAuth/SetupToken only)
  loadQuotaControlSettings(newAccount);

  const tempUnschedState = loadTempUnschedRuleState(credentials);
  tempUnschedEnabled.value = tempUnschedState.enabled;
  tempUnschedRules.value = tempUnschedState.rules;

  // Initialize compatible API key/upstream fields.
  if (
    (newAccount.type === "apikey" || newAccount.type === "upstream") &&
    newAccount.credentials
  ) {
    const credentials = newAccount.credentials as Record<string, unknown>;
    editBaseUrl.value = (credentials.base_url as string) || platformDefaultUrl;

    applyModelRestrictionState(credentials.model_mapping);

    // Load pool mode
    poolModeEnabled.value = credentials.pool_mode === true;
    poolModeRetryCount.value = normalizePoolModeRetryCount(
      Number(
        credentials.pool_mode_retry_count ?? DEFAULT_POOL_MODE_RETRY_COUNT,
      ),
    );

    // Load custom error codes
    customErrorCodesEnabled.value =
      credentials.custom_error_codes_enabled === true;
    const existingErrorCodes = credentials.custom_error_codes as
      | number[]
      | undefined;
    if (existingErrorCodes && Array.isArray(existingErrorCodes)) {
      selectedErrorCodes.value = [...existingErrorCodes];
    } else {
      selectedErrorCodes.value = [];
    }
  } else if (newAccount.type === "bedrock" && newAccount.credentials) {
    const bedrockCreds = newAccount.credentials as Record<string, unknown>;
    const authMode = (bedrockCreds.auth_mode as string) || "sigv4";
    editBedrockRegion.value = (bedrockCreds.aws_region as string) || "";
    editBedrockForceGlobal.value =
      (bedrockCreds.aws_force_global as string) === "true";

    if (authMode === "apikey") {
      editBedrockApiKeyValue.value = "";
    } else {
      editBedrockAccessKeyId.value =
        (bedrockCreds.aws_access_key_id as string) || "";
      editBedrockSecretAccessKey.value = "";
      editBedrockSessionToken.value = "";
    }

    // Load pool mode for bedrock
    poolModeEnabled.value = bedrockCreds.pool_mode === true;
    const retryCount = bedrockCreds.pool_mode_retry_count;
    poolModeRetryCount.value =
      typeof retryCount === "number" && retryCount >= 0
        ? retryCount
        : DEFAULT_POOL_MODE_RETRY_COUNT;

    // Load quota limits for bedrock
    const bedrockExtra = (newAccount.extra as Record<string, unknown>) || {};
    editQuotaLimit.value =
      typeof bedrockExtra.quota_limit === "number"
        ? bedrockExtra.quota_limit
        : null;
    editQuotaDailyLimit.value =
      typeof bedrockExtra.quota_daily_limit === "number"
        ? bedrockExtra.quota_daily_limit
        : null;
    editQuotaWeeklyLimit.value =
      typeof bedrockExtra.quota_weekly_limit === "number"
        ? bedrockExtra.quota_weekly_limit
        : null;
    editQuotaNotifyDailyEnabled.value =
      (bedrockExtra.quota_notify_daily_enabled as boolean) || null;
    editQuotaNotifyDailyThreshold.value =
      (bedrockExtra.quota_notify_daily_threshold as number) ?? null;
    editQuotaNotifyDailyThresholdType.value =
      (bedrockExtra.quota_notify_daily_threshold_type as
        | "fixed"
        | "percentage") || null;
    editQuotaNotifyWeeklyEnabled.value =
      (bedrockExtra.quota_notify_weekly_enabled as boolean) || null;
    editQuotaNotifyWeeklyThreshold.value =
      (bedrockExtra.quota_notify_weekly_threshold as number) ?? null;
    editQuotaNotifyWeeklyThresholdType.value =
      (bedrockExtra.quota_notify_weekly_threshold_type as
        | "fixed"
        | "percentage") || null;
    editQuotaNotifyTotalEnabled.value =
      (bedrockExtra.quota_notify_total_enabled as boolean) || null;
    editQuotaNotifyTotalThreshold.value =
      (bedrockExtra.quota_notify_total_threshold as number) ?? null;
    editQuotaNotifyTotalThresholdType.value =
      (bedrockExtra.quota_notify_total_threshold_type as
        | "fixed"
        | "percentage") || null;

    applyModelRestrictionState(bedrockCreds.model_mapping);
  } else if (newAccount.type === "session") {
    editSessionToken.value = "";
  } else {
    editBaseUrl.value = platformDefaultUrl;

    // Load model mappings for OpenAI OAuth accounts
    if (newAccount.platform === "openai" && newAccount.credentials) {
      const oauthCredentials = newAccount.credentials as Record<
        string,
        unknown
      >;
      applyModelRestrictionState(oauthCredentials.model_mapping);
    } else {
      resetModelRestrictionState();
    }
    poolModeEnabled.value = false;
    poolModeRetryCount.value = DEFAULT_POOL_MODE_RETRY_COUNT;
    customErrorCodesEnabled.value = false;
    selectedErrorCodes.value = [];
  }
};

watch(
  [() => props.show, () => props.account],
  ([show, newAccount], [wasShow, previousAccount]) => {
    if (!show || !newAccount) {
      return;
    }
    if (!wasShow || newAccount !== previousAccount) {
      syncFormFromAccount(newAccount);
      loadTLSProfiles();
    }
  },
  { immediate: true },
);

async function loadTLSProfiles() {
  try {
    const profiles = await adminAPI.tlsFingerprintProfiles.list();
    tlsFingerprintProfiles.value = profiles.map((p) => ({
      id: p.id,
      name: p.name,
    }));
  } catch {
    tlsFingerprintProfiles.value = [];
  }
}

// Model mapping helpers
const addModelMapping = () => {
  appendEmptyModelMapping(modelMappings.value);
};

const removeModelMapping = (index: number) => {
  removeModelMappingAt(modelMappings.value, index);
};

const addPresetMapping = (from: string, to: string) => {
  appendPresetModelMapping(modelMappings.value, from, to, (model) => {
    appStore.showInfo(t("admin.accounts.mappingExists", { model }));
  });
};

const addAntigravityModelMapping = () => {
  appendEmptyModelMapping(antigravityModelMappings.value);
};

const removeAntigravityModelMapping = (index: number) => {
  removeModelMappingAt(antigravityModelMappings.value, index);
};

const addAntigravityPresetMapping = (from: string, to: string) => {
  appendPresetModelMapping(
    antigravityModelMappings.value,
    from,
    to,
    (model) => {
      appStore.showInfo(t("admin.accounts.mappingExists", { model }));
    },
  );
};

// Error code toggle helper
const toggleErrorCode = (code: number) => {
  const index = selectedErrorCodes.value.indexOf(code);
  if (index === -1) {
    if (!confirmCustomErrorCodeSelection(code, confirm, t)) {
      return;
    }
    selectedErrorCodes.value.push(code);
  } else {
    selectedErrorCodes.value.splice(index, 1);
  }
};

// Add custom error code from input
const addCustomErrorCode = () => {
  const code = customErrorCodeInput.value;
  if (code === null || code < 100 || code > 599) {
    appStore.showError(t("admin.accounts.invalidErrorCode"));
    return;
  }
  if (selectedErrorCodes.value.includes(code)) {
    appStore.showInfo(t("admin.accounts.errorCodeExists"));
    return;
  }
  if (!confirmCustomErrorCodeSelection(code, confirm, t)) {
    return;
  }
  selectedErrorCodes.value.push(code);
  customErrorCodeInput.value = null;
};

// Remove error code
const removeErrorCode = (code: number) => {
  const index = selectedErrorCodes.value.indexOf(code);
  if (index !== -1) {
    selectedErrorCodes.value.splice(index, 1);
  }
};

const addTempUnschedRule = (preset?: TempUnschedRuleForm) => {
  tempUnschedRules.value.push(createTempUnschedRule(preset));
};

const removeTempUnschedRule = (index: number) => {
  tempUnschedRules.value.splice(index, 1);
};

const moveTempUnschedRule = (index: number, direction: number) => {
  moveItemInPlace(tempUnschedRules.value, index, direction);
};

// Load quota control settings from account (Anthropic OAuth/SetupToken only)
function loadQuotaControlSettings(account: Account) {
  // Reset all quota control state first
  windowCostEnabled.value = false;
  windowCostLimit.value = null;
  windowCostStickyReserve.value = null;
  sessionLimitEnabled.value = false;
  maxSessions.value = null;
  sessionIdleTimeout.value = null;
  rpmLimitEnabled.value = false;
  baseRpm.value = null;
  rpmStrategy.value = "tiered";
  rpmStickyBuffer.value = null;
  userMsgQueueMode.value = "";
  tlsFingerprintEnabled.value = false;
  tlsFingerprintProfileId.value = null;
  sessionIdMaskingEnabled.value = false;
  cacheTTLOverrideEnabled.value = false;
  cacheTTLOverrideTarget.value = "5m";
  customBaseUrlEnabled.value = false;
  customBaseUrl.value = "";

  // Only applies to Anthropic OAuth/SetupToken accounts
  if (
    account.platform !== "anthropic" ||
    (account.type !== "oauth" && account.type !== "setup-token")
  ) {
    return;
  }

  // Load from extra field (via backend DTO fields)
  if (account.window_cost_limit != null && account.window_cost_limit > 0) {
    windowCostEnabled.value = true;
    windowCostLimit.value = account.window_cost_limit;
    windowCostStickyReserve.value = account.window_cost_sticky_reserve ?? 10;
  }

  if (account.max_sessions != null && account.max_sessions > 0) {
    sessionLimitEnabled.value = true;
    maxSessions.value = account.max_sessions;
    sessionIdleTimeout.value = account.session_idle_timeout_minutes ?? 5;
  }

  // RPM limit
  if (account.base_rpm != null && account.base_rpm > 0) {
    rpmLimitEnabled.value = true;
    baseRpm.value = account.base_rpm;
    rpmStrategy.value =
      (account.rpm_strategy as "tiered" | "sticky_exempt") || "tiered";
    rpmStickyBuffer.value = account.rpm_sticky_buffer ?? null;
  }

  // UMQ mode（独立于 RPM 加载，防止编辑无 RPM 账号时丢失已有配置）
  userMsgQueueMode.value = account.user_msg_queue_mode ?? "";

  // Load TLS fingerprint setting
  if (account.enable_tls_fingerprint === true) {
    tlsFingerprintEnabled.value = true;
  }
  tlsFingerprintProfileId.value = account.tls_fingerprint_profile_id ?? null;

  // Load session ID masking setting
  if (account.session_id_masking_enabled === true) {
    sessionIdMaskingEnabled.value = true;
  }

  // Load cache TTL override setting
  if (account.cache_ttl_override_enabled === true) {
    cacheTTLOverrideEnabled.value = true;
    cacheTTLOverrideTarget.value = account.cache_ttl_override_target || "5m";
  }

  // Load custom base URL setting
  if (account.custom_base_url_enabled === true) {
    customBaseUrlEnabled.value = true;
    customBaseUrl.value = account.custom_base_url || "";
  }
}

const clearMixedChannelDialog = () => {
  resetMixedChannelDialogState();
};

const openMixedChannelDialog = (opts: {
  response?: CheckMixedChannelResponse;
  message?: string;
  onConfirm: () => Promise<void>;
}) => {
  mixedChannelWarningDetails.value = buildMixedChannelDetails(opts.response);
  mixedChannelWarningRawMessage.value =
    opts.message ||
    opts.response?.message ||
    t("admin.accounts.failedToUpdate");
  mixedChannelWarningAction.value = opts.onConfirm;
  showMixedChannelWarning.value = true;
};

const withAntigravityConfirmFlag = (payload: Record<string, unknown>) => {
  if (
    props.account?.platform &&
    needsMixedChannelCheck(props.account.platform) &&
    antigravityMixedChannelConfirmed.value
  ) {
    return {
      ...payload,
      confirm_mixed_channel_risk: true,
    };
  }
  const cloned = { ...payload };
  delete cloned.confirm_mixed_channel_risk;
  return cloned;
};

const ensureAntigravityMixedChannelConfirmed = async (
  onConfirm: () => Promise<void>,
): Promise<boolean> => {
  if (
    !props.account?.platform ||
    !needsMixedChannelCheck(props.account.platform)
  ) {
    return true;
  }
  if (antigravityMixedChannelConfirmed.value) {
    return true;
  }
  if (!props.account) {
    return false;
  }

  try {
    const result = await adminAPI.accounts.checkMixedChannelRisk({
      platform: props.account.platform,
      group_ids: form.group_ids,
      account_id: props.account.id,
    });
    if (!result.has_risk) {
      return true;
    }
    openMixedChannelDialog({
      response: result,
      onConfirm: async () => {
        antigravityMixedChannelConfirmed.value = true;
        await onConfirm();
      },
    });
    return false;
  } catch (error: any) {
    appStore.showError(
      resolveRequestErrorMessage(error, t("admin.accounts.failedToUpdate")),
    );
    return false;
  }
};

const formatDateTimeLocal = formatDateTimeLocalInput;
const parseDateTimeLocal = parseDateTimeLocalInput;

// Methods
const handleClose = () => {
  antigravityMixedChannelConfirmed.value = false;
  clearMixedChannelDialog();
  emit("close");
};

const submitUpdateAccount = async (
  accountID: number,
  updatePayload: Record<string, unknown>,
) => {
  submitting.value = true;
  try {
    const updatedAccount = await adminAPI.accounts.update(
      accountID,
      withAntigravityConfirmFlag(updatePayload),
    );
    appStore.showSuccess(t("admin.accounts.accountUpdated"));
    emit("updated", updatedAccount);
    handleClose();
  } catch (error: any) {
    if (
      error.status === 409 &&
      error.error === "mixed_channel_warning" &&
      props.account?.platform &&
      needsMixedChannelCheck(props.account.platform)
    ) {
      openMixedChannelDialog({
        message: error.message,
        onConfirm: async () => {
          antigravityMixedChannelConfirmed.value = true;
          await submitUpdateAccount(accountID, updatePayload);
        },
      });
      return;
    }
    appStore.showError(
      resolveRequestErrorMessage(error, t("admin.accounts.failedToUpdate")),
    );
  } finally {
    submitting.value = false;
  }
};

const handleSubmit = async () => {
  if (!props.account) return;
  const accountID = props.account.id;

  if (
    form.status !== "active" &&
    form.status !== "inactive" &&
    form.status !== "error"
  ) {
    appStore.showError(t("admin.accounts.pleaseSelectStatus"));
    return;
  }

  const updatePayload = buildEditAccountBasePayload(
    form,
    autoPauseOnExpired.value,
  );
  try {
    // For apikey type, handle credentials update
    if (props.account.type === "apikey") {
      const currentCredentials = getAccountCredentials();
      const shouldApplyModelMapping = !(
        props.account.platform === "openai" && openaiPassthroughEnabled.value
      );
      const result = buildEditableCompatibleCredentials({
        allowedModels: allowedModels.value,
        apiKeyInput: editApiKey.value,
        baseUrlInput: editBaseUrl.value,
        currentCredentials,
        customErrorCodesEnabled: customErrorCodesEnabled.value,
        defaultBaseUrl: defaultBaseUrl.value,
        mode: modelRestrictionMode.value,
        modelMappings: modelMappings.value,
        poolModeEnabled: poolModeEnabled.value,
        poolModeRetryCount: poolModeRetryCount.value,
        preserveModelMappingWhenDisabled: true,
        selectedErrorCodes: selectedErrorCodes.value,
        shouldApplyModelMapping,
      });
      if (result.error === "api_key_required" || !result.credentials) {
        appStore.showError(t("admin.accounts.apiKeyIsRequired"));
        return;
      }
      const newCredentials = result.credentials;

      // Add intercept warmup requests setting
      if (!applySharedEditCredentialsState(newCredentials)) {
        return;
      }

      updatePayload.credentials = newCredentials;
    } else if (props.account.type === "upstream") {
      const currentCredentials = getAccountCredentials();
      const result = buildEditableCompatibleCredentials({
        allowedModels: allowedModels.value,
        apiKeyInput: editApiKey.value,
        baseUrlInput: editBaseUrl.value,
        currentCredentials,
        customErrorCodesEnabled: customErrorCodesEnabled.value,
        defaultBaseUrl: defaultBaseUrl.value,
        mode: modelRestrictionMode.value,
        modelMappings: modelMappings.value,
        poolModeEnabled: poolModeEnabled.value,
        poolModeRetryCount: poolModeRetryCount.value,
        preserveModelMappingWhenDisabled: false,
        selectedErrorCodes: selectedErrorCodes.value,
        shouldApplyModelMapping: true,
      });
      if (result.error === "api_key_required" || !result.credentials) {
        appStore.showError(t("admin.accounts.apiKeyIsRequired"));
        return;
      }
      const newCredentials = result.credentials;

      // Add intercept warmup requests setting
      if (!applySharedEditCredentialsState(newCredentials)) {
        return;
      }

      updatePayload.credentials = newCredentials;
    } else if (props.account.type === "session") {
      const currentCredentials = getAccountCredentials();
      const newCredentials: Record<string, unknown> = { ...currentCredentials };

      if (editSessionToken.value.trim()) {
        const normalizedSessionToken = normalizeGrokSessionToken(
          editSessionToken.value,
        );
        if (!normalizedSessionToken) {
          appStore.showError(
            t("admin.accounts.grok.sessionTokenInvalidFormat"),
          );
          return;
        }
        newCredentials.session_token = normalizedSessionToken;
      } else if (currentCredentials.session_token) {
        newCredentials.session_token = currentCredentials.session_token;
      } else {
        appStore.showError(t("admin.accounts.grok.sessionTokenRequired"));
        return;
      }

      if (!applySharedEditCredentialsState(newCredentials)) {
        return;
      }

      updatePayload.credentials = newCredentials;
    } else if (props.account.type === "bedrock") {
      const currentCredentials = getAccountCredentials();
      const newCredentials = buildEditableBedrockCredentials({
        accessKeyId: editBedrockAccessKeyId.value,
        allowedModels: allowedModels.value,
        apiKeyInput: editBedrockApiKeyValue.value,
        currentCredentials,
        forceGlobal: editBedrockForceGlobal.value,
        isApiKeyMode: isBedrockAPIKeyMode.value,
        mode: modelRestrictionMode.value,
        modelMappings: modelMappings.value,
        poolModeEnabled: poolModeEnabled.value,
        poolModeRetryCount: poolModeRetryCount.value,
        region: editBedrockRegion.value,
        secretAccessKey: editBedrockSecretAccessKey.value,
        sessionToken: editBedrockSessionToken.value,
      });

      if (!applySharedEditCredentialsState(newCredentials)) {
        return;
      }

      updatePayload.credentials = newCredentials;
    } else {
      // For oauth/setup-token types, only update intercept_warmup_requests if changed
      const currentCredentials = getAccountCredentials();
      const newCredentials: Record<string, unknown> = { ...currentCredentials };

      if (!applySharedEditCredentialsState(newCredentials)) {
        return;
      }

      updatePayload.credentials = newCredentials;
    }

    // OpenAI OAuth: persist model mapping to credentials
    if (props.account.platform === "openai" && props.account.type === "oauth") {
      const currentCredentials = getPendingCredentials(updatePayload);
      const newCredentials: Record<string, unknown> = { ...currentCredentials };
      const shouldApplyModelMapping = !openaiPassthroughEnabled.value;

      if (shouldApplyModelMapping) {
        replaceBuiltModelMapping(
          newCredentials,
          modelRestrictionMode.value,
          allowedModels.value,
          modelMappings.value,
        );
      } else if (currentCredentials.model_mapping) {
        // 透传模式保留现有映射
        newCredentials.model_mapping = currentCredentials.model_mapping;
      }

      updatePayload.credentials = newCredentials;
    }

    // Antigravity: persist model mapping to credentials (applies to all antigravity types)
    // Antigravity 只支持映射模式
    if (props.account.platform === "antigravity") {
      const currentCredentials = getPendingCredentials(updatePayload);
      const newCredentials: Record<string, unknown> = { ...currentCredentials };

      replaceAntigravityModelMapping(
        newCredentials,
        antigravityModelMappings.value,
      );

      updatePayload.credentials = newCredentials;
    }

    // For antigravity accounts, handle mixed_scheduling and allow_overages in extra
    if (props.account.platform === "antigravity") {
      const currentExtra = getAccountExtra();
      updatePayload.extra = buildUpdatedAntigravityExtra(currentExtra, {
        mixedScheduling: mixedScheduling.value,
        allowOverages: allowOverages.value,
      });
    }

    // For Anthropic OAuth/SetupToken accounts, handle quota control settings in extra
    if (
      props.account.platform === "anthropic" &&
      (props.account.type === "oauth" || props.account.type === "setup-token")
    ) {
      const currentExtra = getAccountExtra();
      updatePayload.extra = buildUpdatedAnthropicQuotaControlExtra(
        currentExtra,
        {
          baseRpm: baseRpm.value,
          cacheTTLOverrideEnabled: cacheTTLOverrideEnabled.value,
          cacheTTLOverrideTarget: cacheTTLOverrideTarget.value,
          customBaseUrl: customBaseUrl.value,
          customBaseUrlEnabled: customBaseUrlEnabled.value,
          maxSessions: maxSessions.value,
          rpmLimitEnabled: rpmLimitEnabled.value,
          rpmStickyBuffer: rpmStickyBuffer.value,
          rpmStrategy: rpmStrategy.value,
          sessionIdMaskingEnabled: sessionIdMaskingEnabled.value,
          sessionIdleTimeout: sessionIdleTimeout.value,
          sessionLimitEnabled: sessionLimitEnabled.value,
          tlsFingerprintEnabled: tlsFingerprintEnabled.value,
          tlsFingerprintProfileId: tlsFingerprintProfileId.value,
          userMsgQueueMode: userMsgQueueMode.value,
          windowCostEnabled: windowCostEnabled.value,
          windowCostLimit: windowCostLimit.value,
          windowCostStickyReserve: windowCostStickyReserve.value,
        },
      );
    }

    // For Anthropic API Key accounts, handle passthrough mode in extra
    if (
      props.account.platform === "anthropic" &&
      props.account.type === "apikey"
    ) {
      const currentExtra = getAccountExtra();
      updatePayload.extra = buildUpdatedAnthropicAPIKeyExtra(currentExtra, {
        anthropicPassthroughEnabled: anthropicPassthroughEnabled.value,
      });
    }

    // For OpenAI OAuth/API Key accounts, handle passthrough mode in extra
    if (
      props.account.platform === "openai" &&
      (props.account.type === "oauth" || props.account.type === "apikey")
    ) {
      const currentExtra = getAccountExtra();
      updatePayload.extra = buildUpdatedOpenAIExtra(currentExtra, {
        accountType: props.account.type,
        codexCLIOnlyEnabled: codexCLIOnlyEnabled.value,
        openaiAPIKeyResponsesWebSocketV2Mode:
          openaiAPIKeyResponsesWebSocketV2Mode.value,
        openaiOAuthResponsesWebSocketV2Mode:
          openaiOAuthResponsesWebSocketV2Mode.value,
        openaiPassthroughEnabled: openaiPassthroughEnabled.value,
      });
    }

    // For compatible key/upstream and bedrock accounts, handle quota_limit in extra.
    if (
      props.account.type === "apikey" ||
      props.account.type === "upstream" ||
      props.account.type === "bedrock"
    ) {
      const currentExtra = getPendingExtra(updatePayload);
      updatePayload.extra = buildAccountQuotaExtra(currentExtra, {
        dailyResetHour: editDailyResetHour.value,
        dailyResetMode: editDailyResetMode.value,
        quotaDailyLimit: editQuotaDailyLimit.value,
        quotaLimit: editQuotaLimit.value,
        quotaWeeklyLimit: editQuotaWeeklyLimit.value,
        quotaNotifyDailyEnabled: editQuotaNotifyDailyEnabled.value,
        quotaNotifyDailyThreshold: editQuotaNotifyDailyThreshold.value,
        quotaNotifyDailyThresholdType: editQuotaNotifyDailyThresholdType.value,
        quotaNotifyWeeklyEnabled: editQuotaNotifyWeeklyEnabled.value,
        quotaNotifyWeeklyThreshold: editQuotaNotifyWeeklyThreshold.value,
        quotaNotifyWeeklyThresholdType:
          editQuotaNotifyWeeklyThresholdType.value,
        quotaNotifyTotalEnabled: editQuotaNotifyTotalEnabled.value,
        quotaNotifyTotalThreshold: editQuotaNotifyTotalThreshold.value,
        quotaNotifyTotalThresholdType: editQuotaNotifyTotalThresholdType.value,
        resetTimezone: editResetTimezone.value,
        weeklyResetDay: editWeeklyResetDay.value,
        weeklyResetHour: editWeeklyResetHour.value,
        weeklyResetMode: editWeeklyResetMode.value,
      });
    }

    const canContinue = await ensureAntigravityMixedChannelConfirmed(
      async () => {
        await submitUpdateAccount(accountID, updatePayload);
      },
    );
    if (!canContinue) {
      return;
    }

    await submitUpdateAccount(accountID, updatePayload);
  } catch (error: any) {
    appStore.showError(
      resolveRequestErrorMessage(error, t("admin.accounts.failedToUpdate")),
    );
  }
};

// Handle mixed channel warning confirmation
const handleMixedChannelConfirm = async () => {
  const action = mixedChannelWarningAction.value;
  if (!action) {
    clearMixedChannelDialog();
    return;
  }
  clearMixedChannelDialog();
  submitting.value = true;
  try {
    await action();
  } finally {
    submitting.value = false;
  }
};

const handleMixedChannelCancel = () => {
  clearMixedChannelDialog();
};
</script>

<style scoped>
.form-section {
  border-top: 1px solid
    color-mix(in srgb, var(--theme-page-border) 76%, transparent);
  padding-top: 1rem;
}

.edit-account-modal__choice-text,
.edit-account-modal__table-heading,
.edit-account-modal__table-primary {
  color: var(--theme-page-text);
}

.edit-account-modal__muted,
.edit-account-modal__table-secondary {
  color: var(--theme-page-muted);
}

.edit-account-modal__config-card {
  border-radius: var(--theme-surface-radius);
  padding: var(--theme-markdown-block-padding);
  border: 1px solid
    color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  background: color-mix(
    in srgb,
    var(--theme-surface-soft) 90%,
    var(--theme-surface)
  );
}

.edit-account-modal__config-card--compact {
  padding: 0.75rem;
}

.edit-account-modal__notice {
  border-radius: var(--theme-auth-feedback-radius);
  border-color: color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.edit-account-modal__notice-card {
  padding: var(--theme-auth-callback-feedback-padding);
}

.edit-account-modal__notice--purple,
.edit-account-modal__tone-tag--purple {
  --edit-account-tone-rgb: var(--theme-brand-purple-rgb);
}

.edit-account-modal__notice--amber,
.edit-account-modal__tone-tag--amber {
  --edit-account-tone-rgb: var(--theme-warning-rgb);
}

.edit-account-modal__notice--blue,
.edit-account-modal__tone-tag--blue {
  --edit-account-tone-rgb: var(--theme-info-rgb);
}

.edit-account-modal__notice--danger,
.edit-account-modal__tone-tag--danger {
  --edit-account-tone-rgb: var(--theme-danger-rgb);
}

.edit-account-modal__notice--purple,
.edit-account-modal__notice--amber,
.edit-account-modal__notice--blue,
.edit-account-modal__notice--danger {
  background: color-mix(
    in srgb,
    rgb(var(--edit-account-tone-rgb)) 10%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--edit-account-tone-rgb)) 84%,
    var(--theme-page-text)
  );
}

.edit-account-modal__tone-tag {
  display: inline-flex;
  align-items: center;
}

.edit-account-modal__tone-tag--purple,
.edit-account-modal__tone-tag--amber,
.edit-account-modal__tone-tag--blue,
.edit-account-modal__tone-tag--danger {
  background: color-mix(
    in srgb,
    rgb(var(--edit-account-tone-rgb)) 16%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--edit-account-tone-rgb)) 88%,
    var(--theme-page-text)
  );
}

.edit-account-modal__mode-toggle--idle,
.edit-account-modal__status-chip--idle {
  background: color-mix(
    in srgb,
    var(--theme-surface-soft) 86%,
    var(--theme-surface)
  );
  color: var(--theme-page-muted);
}

.edit-account-modal__mode-toggle-control {
  border-radius: var(--theme-button-radius);
  padding: 0.5rem 1rem;
}

.edit-account-modal__compact-action {
  padding-inline: var(--theme-settings-action-padding-x);
}

.edit-account-modal__status-chip-control {
  border-radius: var(--theme-button-radius);
  padding: 0.375rem 0.75rem;
}

.edit-account-modal__summary-chip {
  border-radius: 999px;
  padding: 0.125rem 0.625rem;
}

.edit-account-modal__icon-button {
  border-radius: var(--theme-settings-inline-button-radius);
  padding: var(--theme-settings-inline-button-padding);
}

.edit-account-modal__mode-toggle--idle:hover,
.edit-account-modal__status-chip--idle:hover {
  background: color-mix(
    in srgb,
    var(--theme-page-border) 66%,
    var(--theme-surface)
  );
  color: var(--theme-page-text);
}

.edit-account-modal__mode-toggle--accent {
  background: color-mix(in srgb, var(--theme-accent) 14%, var(--theme-surface));
  color: color-mix(in srgb, var(--theme-accent) 90%, var(--theme-page-text));
}

.edit-account-modal__mode-toggle--purple,
.edit-account-modal__status-chip--purple {
  background: color-mix(
    in srgb,
    rgb(var(--theme-brand-purple-rgb)) 14%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--theme-brand-purple-rgb)) 88%,
    var(--theme-page-text)
  );
}

.edit-account-modal__mode-toggle--danger,
.edit-account-modal__status-chip--danger {
  background: color-mix(
    in srgb,
    rgb(var(--theme-danger-rgb)) 12%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--theme-danger-rgb)) 88%,
    var(--theme-page-text)
  );
}

.edit-account-modal__umq-option-control {
  border: 1px solid
    color-mix(in srgb, var(--theme-input-border) 82%, transparent);
  border-radius: calc(var(--theme-button-radius) - 2px);
  padding: 0.375rem 0.75rem;
}

.edit-account-modal__umq-option--selected {
  background: var(--theme-accent);
  color: var(--theme-accent-text);
  border-color: var(--theme-accent);
}

.edit-account-modal__umq-option--idle {
  background: var(--theme-surface);
  color: var(--theme-page-text);
}

.edit-account-modal__umq-option--idle:hover {
  background: color-mix(
    in srgb,
    var(--theme-surface-soft) 82%,
    var(--theme-surface)
  );
}

.edit-account-modal__switch {
  box-shadow: 0 0 0 1px
    color-mix(in srgb, var(--theme-page-border) 40%, transparent);
}

.edit-account-modal__switch:focus-visible {
  box-shadow:
    0 0 0 2px color-mix(in srgb, var(--theme-accent) 22%, transparent),
    0 0 0 4px color-mix(in srgb, var(--theme-accent) 12%, transparent);
}

.edit-account-modal__switch--enabled {
  background: var(--theme-accent);
}

.edit-account-modal__switch--disabled {
  background: color-mix(
    in srgb,
    var(--theme-page-border) 76%,
    var(--theme-surface)
  );
}

.edit-account-modal__switch-thumb {
  background: var(--theme-surface-contrast);
}

.edit-account-modal__checkbox {
  border-color: color-mix(in srgb, var(--theme-input-border) 82%, transparent);
  color: var(--theme-accent);
}

.edit-account-modal__checkbox:focus {
  outline: none;
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--theme-accent) 18%, transparent);
}

.edit-account-modal__input-error {
  border-color: rgb(var(--theme-danger-rgb));
}

.edit-account-modal__error-text {
  color: rgb(var(--theme-danger-rgb));
}

.edit-account-modal__tooltip {
  width: 18rem;
  border-radius: var(--theme-tooltip-radius);
  padding: 0.5rem 0.75rem;
  background: var(--theme-surface-contrast);
  color: var(--theme-filled-text);
}

.edit-account-modal__tooltip-arrow {
  border-bottom-color: var(--theme-surface-contrast);
}
</style>
