<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.createAccount')"
    width="wide"
    @close="handleClose"
  >
    <!-- Step Indicator for OAuth accounts -->
    <div v-if="isOAuthFlow" class="mb-6 flex items-center justify-center">
      <div class="create-account-modal__stepper flex items-center space-x-4">
        <div class="create-account-modal__step-group flex items-center">
          <div
            :class="[
              'create-account-modal__step-node flex h-8 w-8 items-center justify-center rounded-full text-sm font-semibold',
              step >= 1 ? 'create-account-modal__step-node--active' : 'create-account-modal__step-node--idle'
            ]"
          >
            1
          </div>
          <span class="create-account-modal__step-label ml-2 text-sm font-medium">{{
            t('admin.accounts.oauth.authMethod')
          }}</span>
        </div>
        <div class="create-account-modal__step-connector h-0.5 w-8" />
        <div class="create-account-modal__step-group flex items-center">
          <div
            :class="[
              'create-account-modal__step-node flex h-8 w-8 items-center justify-center rounded-full text-sm font-semibold',
              step >= 2 ? 'create-account-modal__step-node--active' : 'create-account-modal__step-node--idle'
            ]"
          >
            2
          </div>
          <span class="create-account-modal__step-label ml-2 text-sm font-medium">{{
            oauthStepTitle
          }}</span>
        </div>
      </div>
    </div>

    <!-- Step 1: Basic Info -->
    <form
      v-if="step === 1"
      id="create-account-form"
      @submit.prevent="handleSubmit"
      class="space-y-5"
    >
      <div>
        <label class="input-label">{{ t('admin.accounts.accountName') }}</label>
        <input
          v-model="form.name"
          type="text"
          required
          class="input"
          :placeholder="t('admin.accounts.enterAccountName')"
          data-tour="account-form-name"
        />
      </div>
      <div>
        <label class="input-label">{{ t('admin.accounts.notes') }}</label>
        <textarea
          v-model="form.notes"
          rows="3"
          class="input"
          :placeholder="t('admin.accounts.notesPlaceholder')"
        ></textarea>
        <p class="input-hint">{{ t('admin.accounts.notesHint') }}</p>
      </div>

      <!-- Platform Selection - Segmented Control Style -->
      <div>
        <label class="input-label">{{ t('admin.accounts.platform') }}</label>
        <div class="segmented-control mt-2 flex" data-tour="account-form-platform">
          <button
            type="button"
            @click="form.platform = 'anthropic'"
            :class="[
              'create-account-modal__platform-button create-account-modal__platform-button-control flex flex-1 items-center justify-center gap-2 text-sm font-medium transition-all',
              form.platform === 'anthropic'
                ? 'create-account-modal__platform-button--active create-account-modal__platform-button--anthropic'
                : 'create-account-modal__platform-button--idle'
            ]"
          >
            <Icon name="sparkles" size="sm" />
            Anthropic
          </button>
          <button
            type="button"
            @click="form.platform = 'openai'"
            :class="[
              'create-account-modal__platform-button create-account-modal__platform-button-control flex flex-1 items-center justify-center gap-2 text-sm font-medium transition-all',
              form.platform === 'openai'
                ? 'create-account-modal__platform-button--active create-account-modal__platform-button--openai'
                : 'create-account-modal__platform-button--idle'
            ]"
          >
            <svg
              class="h-4 w-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="1.5"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z"
              />
            </svg>
            OpenAI
          </button>
          <button
            type="button"
            @click="form.platform = 'gemini'"
            :class="[
              'create-account-modal__platform-button create-account-modal__platform-button-control flex flex-1 items-center justify-center gap-2 text-sm font-medium transition-all',
              form.platform === 'gemini'
                ? 'create-account-modal__platform-button--active create-account-modal__platform-button--gemini'
                : 'create-account-modal__platform-button--idle'
            ]"
          >
            <svg
              class="h-4 w-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="1.5"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M12 2l1.5 6.5L20 10l-6.5 1.5L12 18l-1.5-6.5L4 10l6.5-1.5L12 2z"
              />
            </svg>
            Gemini
          </button>
          <button
            type="button"
            @click="form.platform = 'antigravity'"
            :class="[
              'create-account-modal__platform-button create-account-modal__platform-button-control flex flex-1 items-center justify-center gap-2 text-sm font-medium transition-all',
              form.platform === 'antigravity'
                ? 'create-account-modal__platform-button--active create-account-modal__platform-button--antigravity'
                : 'create-account-modal__platform-button--idle'
            ]"
          >
            <Icon name="cloud" size="sm" />
            Antigravity
          </button>
        </div>
      </div>

      <!-- Account Type Selection (Anthropic) -->
      <div v-if="form.platform === 'anthropic'">
        <label class="input-label">{{ t('admin.accounts.accountType') }}</label>
        <div class="mt-2 grid grid-cols-3 gap-3" data-tour="account-form-type">
          <button
            type="button"
            @click="accountCategory = 'oauth-based'"
            :class="getChoiceCardClasses(accountCategory === 'oauth-based', 'orange')"
          >
            <div :class="getChoiceIconClasses(accountCategory === 'oauth-based', 'orange')">
              <Icon name="sparkles" size="sm" />
            </div>
            <div>
              <span class="create-account-modal__choice-title block text-sm font-medium">{{
                t('admin.accounts.claudeCode')
              }}</span>
              <span class="create-account-modal__choice-description text-xs">{{
                t('admin.accounts.oauthSetupToken')
              }}</span>
            </div>
          </button>

          <button
            type="button"
            @click="accountCategory = 'apikey'"
            :class="getChoiceCardClasses(accountCategory === 'apikey', 'purple')"
          >
            <div :class="getChoiceIconClasses(accountCategory === 'apikey', 'purple')">
              <Icon name="key" size="sm" />
            </div>
            <div>
              <span class="create-account-modal__choice-title block text-sm font-medium">{{
                t('admin.accounts.claudeConsole')
              }}</span>
              <span class="create-account-modal__choice-description text-xs">{{
                t('admin.accounts.apiKey')
              }}</span>
            </div>
          </button>

          <button
            type="button"
            @click="accountCategory = 'bedrock'"
            :class="getChoiceCardClasses(accountCategory === 'bedrock', 'amber')"
          >
            <div :class="getChoiceIconClasses(accountCategory === 'bedrock', 'amber')">
              <Icon name="cloud" size="sm" />
            </div>
            <div>
              <span class="create-account-modal__choice-title block text-sm font-medium">{{
                t('admin.accounts.bedrockLabel')
              }}</span>
              <span class="create-account-modal__choice-description text-xs">{{
                t('admin.accounts.bedrockDesc')
              }}</span>
            </div>
          </button>

        </div>
      </div>

      <!-- Account Type Selection (OpenAI) -->
      <div v-if="form.platform === 'openai'">
        <label class="input-label">{{ t('admin.accounts.accountType') }}</label>
        <div class="mt-2 grid grid-cols-2 gap-3" data-tour="account-form-type">
          <button
            type="button"
            @click="accountCategory = 'oauth-based'"
            :class="getChoiceCardClasses(accountCategory === 'oauth-based', 'green')"
          >
            <div :class="getChoiceIconClasses(accountCategory === 'oauth-based', 'green')">
              <Icon name="key" size="sm" />
            </div>
            <div>
              <span class="create-account-modal__choice-title block text-sm font-medium">OAuth</span>
              <span class="create-account-modal__choice-description text-xs">{{ t('admin.accounts.types.chatgptOauth') }}</span>
            </div>
          </button>

          <button
            type="button"
            @click="accountCategory = 'apikey'"
            :class="getChoiceCardClasses(accountCategory === 'apikey', 'purple')"
          >
            <div :class="getChoiceIconClasses(accountCategory === 'apikey', 'purple')">
              <Icon name="key" size="sm" />
            </div>
            <div>
              <span class="create-account-modal__choice-title block text-sm font-medium">API Key</span>
              <span class="create-account-modal__choice-description text-xs">{{ t('admin.accounts.types.responsesApi') }}</span>
            </div>
          </button>
        </div>
      </div>

      <!-- Account Type Selection (Gemini) -->
      <div v-if="form.platform === 'gemini'">
        <div class="flex items-center justify-between">
          <label class="input-label">{{ t('admin.accounts.accountType') }}</label>
          <button
            type="button"
            @click="showGeminiHelpDialog = true"
            class="create-account-modal__help-button"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z" />
            </svg>
            {{ t('admin.accounts.gemini.helpButton') }}
          </button>
        </div>
        <div class="mt-2 grid grid-cols-2 gap-3" data-tour="account-form-type">
          <button
            type="button"
            @click="accountCategory = 'oauth-based'"
            :class="getChoiceCardClasses(accountCategory === 'oauth-based', 'blue')"
          >
            <div :class="getChoiceIconClasses(accountCategory === 'oauth-based', 'blue')">
              <Icon name="key" size="sm" />
            </div>
            <div>
              <span class="create-account-modal__choice-title block text-sm font-medium">
                {{ t('admin.accounts.gemini.accountType.oauthTitle') }}
              </span>
              <span class="create-account-modal__choice-description text-xs">
                {{ t('admin.accounts.gemini.accountType.oauthDesc') }}
              </span>
            </div>
          </button>

          <button
            type="button"
            @click="accountCategory = 'apikey'"
            :class="getChoiceCardClasses(accountCategory === 'apikey', 'purple')"
          >
            <div :class="getChoiceIconClasses(accountCategory === 'apikey', 'purple')">
              <svg
                class="h-4 w-4"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M15.75 5.25a3 3 0 013 3m3 0a6 6 0 01-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1721.75 8.25z"
                />
              </svg>
            </div>
            <div>
              <span class="create-account-modal__choice-title block text-sm font-medium">
                {{ t('admin.accounts.gemini.accountType.apiKeyTitle') }}
              </span>
              <span class="create-account-modal__choice-description text-xs">
                {{ t('admin.accounts.gemini.accountType.apiKeyDesc') }}
              </span>
            </div>
          </button>
        </div>

        <div
          v-if="accountCategory === 'apikey'"
          class="create-account-modal__notice create-account-modal__notice--purple create-account-modal__notice-inline mt-3 text-xs"
        >
          <p>{{ t('admin.accounts.gemini.accountType.apiKeyNote') }}</p>
          <div class="mt-2 flex flex-wrap gap-2">
            <a
              :href="geminiHelpLinks.apiKey"
              class="create-account-modal__link font-medium"
              target="_blank"
              rel="noreferrer"
            >
              {{ t('admin.accounts.gemini.accountType.apiKeyLink') }}
            </a>
          </div>
        </div>

        <!-- OAuth Type Selection (only show when oauth-based is selected) -->
        <div v-if="accountCategory === 'oauth-based'" class="mt-4">
          <label class="input-label">{{ t('admin.accounts.oauth.gemini.oauthTypeLabel') }}</label>
          <div class="mt-2 grid grid-cols-2 gap-3">
            <!-- Google One OAuth -->
            <button
              type="button"
              @click="handleSelectGeminiOAuthType('google_one')"
              :class="getChoiceCardClasses(geminiOAuthType === 'google_one', 'purple')"
            >
              <div :class="getChoiceIconClasses(geminiOAuthType === 'google_one', 'purple')">
                <Icon name="user" size="sm" />
              </div>
              <div class="min-w-0">
                <span class="create-account-modal__choice-title block text-sm font-medium">
                  Google One
                </span>
                <span class="create-account-modal__choice-description text-xs">
                  个人账号，享受 Google One 订阅配额
                </span>
                <div class="mt-2 flex flex-wrap gap-1">
                  <span :class="getToneTagClasses('purple')">
                    推荐个人用户
                  </span>
                  <span :class="getToneTagClasses('emerald')">
                    无需 GCP
                  </span>
                </div>
              </div>
            </button>

            <!-- GCP Code Assist OAuth -->
            <button
              type="button"
              @click="handleSelectGeminiOAuthType('code_assist')"
              :class="getChoiceCardClasses(geminiOAuthType === 'code_assist', 'blue')"
            >
              <div :class="getChoiceIconClasses(geminiOAuthType === 'code_assist', 'blue')">
                <Icon name="cloud" size="sm" />
              </div>
              <div class="min-w-0">
                <span class="create-account-modal__choice-title block text-sm font-medium">
                  GCP Code Assist
                </span>
                <span class="create-account-modal__choice-description text-xs">
                  企业级，需要 GCP 项目
                </span>
                <div class="create-account-modal__choice-description mt-1 text-xs">
                  需要激活 GCP 项目并绑定信用卡
                  <a
                    :href="geminiHelpLinks.gcpProject"
                    class="create-account-modal__link ml-1"
                    target="_blank"
                    rel="noreferrer"
                  >
                    {{ t('admin.accounts.gemini.oauthType.gcpProjectLink') }}
                  </a>
                </div>
                <div class="mt-2 flex flex-wrap gap-1">
                  <span :class="getToneTagClasses('blue')">
                    企业用户
                  </span>
                  <span :class="getToneTagClasses('emerald')">
                    高并发
                  </span>
                </div>
              </div>
            </button>
          </div>

          <!-- Advanced Options Toggle -->
          <div class="mt-3">
            <button
              type="button"
              @click="showAdvancedOAuth = !showAdvancedOAuth"
              class="create-account-modal__inline-toggle flex items-center gap-2 text-sm"
            >
              <svg
                :class="['h-4 w-4 transition-transform', showAdvancedOAuth ? 'rotate-90' : '']"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="2"
              >
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
              </svg>
              <span>{{ showAdvancedOAuth ? '隐藏' : '显示' }}高级选项（自建 OAuth Client）</span>
            </button>
          </div>

          <!-- Custom OAuth Client (Advanced) -->
          <div v-if="showAdvancedOAuth" class="mt-3 group relative">
            <button
              type="button"
              :disabled="!geminiAIStudioOAuthEnabled"
              @click="handleSelectGeminiOAuthType('ai_studio')"
              :class="getChoiceCardClasses(geminiOAuthType === 'ai_studio', 'amber', !geminiAIStudioOAuthEnabled)"
            >
              <div :class="getChoiceIconClasses(geminiOAuthType === 'ai_studio', 'amber')">
                <svg
                  class="h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  stroke-width="1.5"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09z"
                  />
                </svg>
              </div>
              <div class="min-w-0">
                <span class="create-account-modal__choice-title block text-sm font-medium">
                  {{ t('admin.accounts.gemini.oauthType.customTitle') }}
                </span>
                <span class="create-account-modal__choice-description text-xs">
                  {{ t('admin.accounts.gemini.oauthType.customDesc') }}
                </span>
                <div class="create-account-modal__choice-description mt-1 text-xs">
                  {{ t('admin.accounts.gemini.oauthType.customRequirement') }}
                </div>
                <div class="mt-2 flex flex-wrap gap-1">
                  <span :class="getToneTagClasses('amber')">
                    {{ t('admin.accounts.gemini.oauthType.badges.orgManaged') }}
                  </span>
                  <span :class="getToneTagClasses('amber')">
                    {{ t('admin.accounts.gemini.oauthType.badges.adminRequired') }}
                  </span>
                </div>
              </div>
              <span
                v-if="!geminiAIStudioOAuthEnabled"
                :class="['create-account-modal__tone-tag-anchor', getToneTagClasses('amber')]"
              >
                {{ t('admin.accounts.oauth.gemini.aiStudioNotConfiguredShort') }}
              </span>
            </button>

            <div
              v-if="!geminiAIStudioOAuthEnabled"
              class="create-account-modal__notice create-account-modal__notice--amber create-account-modal__notice-tooltip pointer-events-none absolute right-0 top-full z-50 mt-2 text-xs opacity-0 shadow-lg transition-opacity group-hover:opacity-100"
            >
              {{ t('admin.accounts.oauth.gemini.aiStudioNotConfiguredTip') }}
            </div>
          </div>
        </div>

        <!-- Tier selection (used as fallback when auto-detection is unavailable/fails) -->
        <div class="mt-4">
          <label class="input-label">{{ t('admin.accounts.gemini.tier.label') }}</label>
          <div class="mt-2">
            <select
              v-if="geminiOAuthType === 'google_one'"
              v-model="geminiTierGoogleOne"
              class="input"
            >
              <option value="google_one_free">{{ t('admin.accounts.gemini.tier.googleOne.free') }}</option>
              <option value="google_ai_pro">{{ t('admin.accounts.gemini.tier.googleOne.pro') }}</option>
              <option value="google_ai_ultra">{{ t('admin.accounts.gemini.tier.googleOne.ultra') }}</option>
            </select>

            <select
              v-else-if="geminiOAuthType === 'code_assist'"
              v-model="geminiTierGcp"
              class="input"
            >
              <option value="gcp_standard">{{ t('admin.accounts.gemini.tier.gcp.standard') }}</option>
              <option value="gcp_enterprise">{{ t('admin.accounts.gemini.tier.gcp.enterprise') }}</option>
            </select>

            <select
              v-else
              v-model="geminiTierAIStudio"
              class="input"
            >
              <option value="aistudio_free">{{ t('admin.accounts.gemini.tier.aiStudio.free') }}</option>
              <option value="aistudio_paid">{{ t('admin.accounts.gemini.tier.aiStudio.paid') }}</option>
            </select>
          </div>
          <p class="input-hint">{{ t('admin.accounts.gemini.tier.hint') }}</p>
        </div>
      </div>

      <!-- Account Type Selection (Antigravity - OAuth or Upstream) -->
      <div v-if="form.platform === 'antigravity'">
        <label class="input-label">{{ t('admin.accounts.accountType') }}</label>
        <div class="mt-2 grid grid-cols-2 gap-3">
          <button
            type="button"
            @click="antigravityAccountType = 'oauth'"
            :class="getChoiceCardClasses(antigravityAccountType === 'oauth', 'purple')"
          >
            <div :class="getChoiceIconClasses(antigravityAccountType === 'oauth', 'purple')">
              <Icon name="key" size="sm" />
            </div>
            <div>
              <span class="create-account-modal__choice-title block text-sm font-medium">OAuth</span>
              <span class="create-account-modal__choice-description text-xs">{{ t('admin.accounts.types.antigravityOauth') }}</span>
            </div>
          </button>

          <button
            type="button"
            @click="antigravityAccountType = 'upstream'"
            :class="getChoiceCardClasses(antigravityAccountType === 'upstream', 'purple')"
          >
            <div :class="getChoiceIconClasses(antigravityAccountType === 'upstream', 'purple')">
              <Icon name="cloud" size="sm" />
            </div>
            <div>
              <span class="create-account-modal__choice-title block text-sm font-medium">API Key</span>
              <span class="create-account-modal__choice-description text-xs">{{ t('admin.accounts.types.antigravityApikey') }}</span>
            </div>
          </button>
        </div>
      </div>

      <!-- Upstream config (only for Antigravity upstream type) -->
      <div v-if="form.platform === 'antigravity' && antigravityAccountType === 'upstream'" class="space-y-4">
        <div>
          <label class="input-label">{{ t('admin.accounts.upstream.baseUrl') }}</label>
          <input
            v-model="upstreamBaseUrl"
            type="text"
            required
            class="input"
            placeholder="https://cloudcode-pa.googleapis.com"
          />
          <p class="input-hint">{{ t('admin.accounts.upstream.baseUrlHint') }}</p>
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.upstream.apiKey') }}</label>
          <input
            v-model="upstreamApiKey"
            type="password"
            required
            class="input font-mono"
            placeholder="sk-..."
          />
          <p class="input-hint">{{ t('admin.accounts.upstream.apiKeyHint') }}</p>
        </div>
      </div>

      <!-- Antigravity model restriction (applies to OAuth + Upstream) -->
      <!-- Antigravity 只支持模型映射模式，不支持白名单模式 -->
      <div v-if="form.platform === 'antigravity'" class="form-section">
        <label class="input-label">{{ t('admin.accounts.modelRestriction') }}</label>

        <!-- Mapping Mode Only (no toggle for Antigravity) -->
        <div>
          <div class="create-account-modal__notice create-account-modal__notice--purple create-account-modal__notice-block mb-3">
            <p class="text-xs">
              {{ t('admin.accounts.mapRequestModels') }}
            </p>
          </div>

          <div v-if="antigravityModelMappings.length > 0" class="mb-3 space-y-2">
            <div
              v-for="(mapping, index) in antigravityModelMappings"
              :key="getAntigravityModelMappingKey(mapping)"
              class="space-y-1"
            >
              <div class="flex items-center gap-2">
                <input
                  v-model="mapping.from"
                  type="text"
                  :class="getValidationInputClasses(!isValidWildcardPattern(mapping.from), 'flex-1')"
                  :placeholder="t('admin.accounts.requestModel')"
                />
                <svg class="create-account-modal__choice-description h-4 w-4 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
                </svg>
                <input
                  v-model="mapping.to"
                  type="text"
                  :class="getValidationInputClasses(mapping.to.includes('*'), 'flex-1')"
                  :placeholder="t('admin.accounts.actualModel')"
                />
                <button
                  type="button"
                  @click="removeAntigravityModelMapping(index)"
                  class="create-account-modal__status-chip create-account-modal__status-chip--danger create-account-modal__status-chip-action"
                >
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
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
              <p v-if="!isValidWildcardPattern(mapping.from)" class="create-account-modal__error-text">
                {{ t('admin.accounts.wildcardOnlyAtEnd') }}
              </p>
              <p v-if="mapping.to.includes('*')" class="create-account-modal__error-text">
                {{ t('admin.accounts.targetNoWildcard') }}
              </p>
            </div>
          </div>

          <button
            type="button"
            @click="addAntigravityModelMapping"
            class="btn btn-secondary mb-3 w-full border-2 border-dashed"
          >
            <svg class="mr-1 inline h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            {{ t('admin.accounts.addMapping') }}
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

      <!-- Add Method (only for Anthropic OAuth-based type) -->
      <div v-if="form.platform === 'anthropic' && isOAuthFlow">
        <label class="input-label">{{ t('admin.accounts.addMethod') }}</label>
        <div class="mt-2 flex gap-4">
          <label :class="getRadioOptionClasses(addMethod === 'oauth')">
            <input
              v-model="addMethod"
              type="radio"
              value="oauth"
              class="create-account-modal__radio-input"
            />
            <span class="create-account-modal__choice-title text-sm">{{ t('admin.accounts.types.oauth') }}</span>
          </label>
          <label :class="getRadioOptionClasses(addMethod === 'setup-token')">
            <input
              v-model="addMethod"
              type="radio"
              value="setup-token"
              class="create-account-modal__radio-input"
            />
            <span class="create-account-modal__choice-title text-sm">{{
              t('admin.accounts.setupTokenLongLived')
            }}</span>
          </label>
        </div>
      </div>

      <!-- API Key input (only for apikey type, excluding Antigravity which has its own fields) -->
      <div v-if="form.type === 'apikey' && form.platform !== 'antigravity'" class="space-y-4">
        <div>
          <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
            <label class="input-label mb-0">{{ t('admin.accounts.baseUrl') }}</label>
            <div v-if="form.platform === 'openai'" class="flex flex-wrap gap-2">
              <button
                v-for="preset in openAICompatibleBaseUrlPresets"
                :key="preset.value"
                type="button"
                :class="getPresetMappingChipClasses('success')"
                @click="apiKeyBaseUrl = preset.value"
              >
                {{ preset.label }}
              </button>
            </div>
          </div>
          <input
            v-model="apiKeyBaseUrl"
            type="text"
            class="input"
            :placeholder="baseUrlPlaceholder"
          />
          <p class="input-hint">{{ baseUrlHint }}</p>
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.apiKeyRequired') }}</label>
          <input
            v-model="apiKeyValue"
            type="password"
            required
            class="input font-mono"
            :placeholder="apiKeyPlaceholder"
          />
          <p class="input-hint">{{ apiKeyHint }}</p>
        </div>

        <!-- Gemini API Key tier selection -->
        <div v-if="form.platform === 'gemini'">
          <label class="input-label">{{ t('admin.accounts.gemini.tier.label') }}</label>
          <select v-model="geminiTierAIStudio" class="input">
            <option value="aistudio_free">{{ t('admin.accounts.gemini.tier.aiStudio.free') }}</option>
            <option value="aistudio_paid">{{ t('admin.accounts.gemini.tier.aiStudio.paid') }}</option>
          </select>
          <p class="input-hint">{{ t('admin.accounts.gemini.tier.aiStudioHint') }}</p>
        </div>

        <!-- Model Restriction Section (Antigravity 已在上层条件排除) -->
        <div class="form-section">
          <label class="input-label">{{ t('admin.accounts.modelRestriction') }}</label>

          <div
            v-if="isOpenAIModelRestrictionDisabled"
            class="create-account-modal__notice create-account-modal__notice--amber create-account-modal__notice-block mb-3"
          >
            <p class="text-xs">
              {{ t('admin.accounts.openai.modelRestrictionDisabledByPassthrough') }}
            </p>
          </div>

          <template v-else>
            <!-- Mode Toggle -->
            <div class="mb-4 flex gap-2">
              <button
                type="button"
                @click="modelRestrictionMode = 'whitelist'"
                :class="getModeToggleClasses(modelRestrictionMode === 'whitelist', 'accent')"
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
                {{ t('admin.accounts.modelWhitelist') }}
              </button>
              <button
                type="button"
                @click="modelRestrictionMode = 'mapping'"
                :class="getModeToggleClasses(modelRestrictionMode === 'mapping', 'purple')"
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
                {{ t('admin.accounts.modelMapping') }}
              </button>
            </div>

            <!-- Whitelist Mode -->
            <div v-if="modelRestrictionMode === 'whitelist'">
              <ModelWhitelistSelector v-model="allowedModels" :platform="form.platform" />
              <p class="create-account-modal__choice-description text-xs">
                {{ t('admin.accounts.selectedModels', { count: allowedModels.length }) }}
                <span v-if="allowedModels.length === 0">{{
                  t('admin.accounts.supportsAllModels')
                }}</span>
              </p>
            </div>

            <!-- Mapping Mode -->
            <div v-else>
              <div class="create-account-modal__notice create-account-modal__notice--purple create-account-modal__notice-block mb-3">
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
                  {{ t('admin.accounts.mapRequestModels') }}
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
                  class="create-account-modal__choice-description h-4 w-4 flex-shrink-0"
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
                  class="create-account-modal__status-chip create-account-modal__status-chip--danger create-account-modal__status-chip-action"
                >
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
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
              {{ t('admin.accounts.addMapping') }}
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
              <label class="input-label mb-0">{{ t('admin.accounts.poolMode') }}</label>
              <p class="create-account-modal__choice-description mt-1 text-xs">
                {{ t('admin.accounts.poolModeHint') }}
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
          <div v-if="poolModeEnabled" class="create-account-modal__notice create-account-modal__notice--blue create-account-modal__notice-block">
            <p class="text-xs">
              <Icon name="exclamationCircle" size="sm" class="mr-1 inline" :stroke-width="2" />
              {{ t('admin.accounts.poolModeInfo') }}
            </p>
          </div>
          <div v-if="poolModeEnabled" class="mt-3">
            <label class="input-label">{{ t('admin.accounts.poolModeRetryCount') }}</label>
            <input
              v-model.number="poolModeRetryCount"
              type="number"
              min="0"
              :max="MAX_POOL_MODE_RETRY_COUNT"
              step="1"
              class="input"
            />
            <p class="create-account-modal__choice-description mt-1 text-xs">
              {{
                t('admin.accounts.poolModeRetryCountHint', {
                  default: DEFAULT_POOL_MODE_RETRY_COUNT,
                  max: MAX_POOL_MODE_RETRY_COUNT
                })
              }}
            </p>
          </div>
        </div>

        <!-- Custom Error Codes Section -->
        <div class="form-section">
          <div class="mb-3 flex items-center justify-between">
            <div>
              <label class="input-label mb-0">{{ t('admin.accounts.customErrorCodes') }}</label>
              <p class="create-account-modal__choice-description mt-1 text-xs">
                {{ t('admin.accounts.customErrorCodesHint') }}
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
            <div class="create-account-modal__notice create-account-modal__notice--amber create-account-modal__notice-block">
              <p class="text-xs">
                <Icon name="exclamationTriangle" size="sm" class="mr-1 inline" :stroke-width="2" />
                {{ t('admin.accounts.customErrorCodesWarning') }}
              </p>
            </div>

            <!-- Error Code Buttons -->
            <div class="flex flex-wrap gap-2">
              <button
                v-for="code in commonErrorCodes"
                :key="code.value"
                type="button"
                @click="toggleErrorCode(code.value)"
                :class="getStatusChipClasses(selectedErrorCodes.includes(code.value), 'danger')"
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
              <button type="button" @click="addCustomErrorCode" class="btn btn-secondary create-account-modal__secondary-button">
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
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
                class="create-account-modal__status-chip create-account-modal__status-chip--danger create-account-modal__status-chip-inline"
              >
                {{ code }}
                <button
                  type="button"
                  @click="removeErrorCode(code)"
                  class="create-account-modal__choice-title"
                >
                  <Icon name="x" size="sm" :stroke-width="2" />
                </button>
              </span>
              <span v-if="selectedErrorCodes.length === 0" class="create-account-modal__choice-description text-xs">
                {{ t('admin.accounts.noneSelectedUsesDefault') }}
              </span>
            </div>
          </div>
        </div>

      </div>

      <!-- Bedrock credentials (only for Anthropic Bedrock type) -->
      <div v-if="form.platform === 'anthropic' && accountCategory === 'bedrock'" class="space-y-4">
        <!-- Auth Mode Radio -->
        <div>
          <label class="input-label">{{ t('admin.accounts.bedrockAuthMode') }}</label>
          <div class="mt-2 flex gap-4">
            <label :class="getRadioOptionClasses(bedrockAuthMode === 'sigv4')">
              <input
                v-model="bedrockAuthMode"
                type="radio"
                value="sigv4"
                class="create-account-modal__radio-input"
              />
              <span class="create-account-modal__choice-title text-sm">{{ t('admin.accounts.bedrockAuthModeSigv4') }}</span>
            </label>
            <label :class="getRadioOptionClasses(bedrockAuthMode === 'apikey')">
              <input
                v-model="bedrockAuthMode"
                type="radio"
                value="apikey"
                class="create-account-modal__radio-input"
              />
              <span class="create-account-modal__choice-title text-sm">{{ t('admin.accounts.bedrockAuthModeApikey') }}</span>
            </label>
          </div>
        </div>

        <!-- SigV4 fields -->
        <template v-if="bedrockAuthMode === 'sigv4'">
          <div>
            <label class="input-label">{{ t('admin.accounts.bedrockAccessKeyId') }}</label>
            <input
              v-model="bedrockAccessKeyId"
              type="text"
              required
              class="input font-mono"
              placeholder="AKIA..."
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.accounts.bedrockSecretAccessKey') }}</label>
            <input
              v-model="bedrockSecretAccessKey"
              type="password"
              required
              class="input font-mono"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.accounts.bedrockSessionToken') }}</label>
            <input
              v-model="bedrockSessionToken"
              type="password"
              class="input font-mono"
            />
            <p class="input-hint">{{ t('admin.accounts.bedrockSessionTokenHint') }}</p>
          </div>
        </template>

        <!-- API Key field -->
        <div v-if="bedrockAuthMode === 'apikey'">
          <label class="input-label">{{ t('admin.accounts.bedrockApiKeyInput') }}</label>
          <input
            v-model="bedrockApiKeyValue"
            type="password"
            required
            class="input font-mono"
          />
        </div>

        <!-- Shared: Region -->
        <div>
          <label class="input-label">{{ t('admin.accounts.bedrockRegion') }}</label>
          <select v-model="bedrockRegion" class="input">
            <optgroup label="US">
              <option value="us-east-1">us-east-1 (N. Virginia)</option>
              <option value="us-east-2">us-east-2 (Ohio)</option>
              <option value="us-west-1">us-west-1 (N. California)</option>
              <option value="us-west-2">us-west-2 (Oregon)</option>
              <option value="us-gov-east-1">us-gov-east-1 (GovCloud US-East)</option>
              <option value="us-gov-west-1">us-gov-west-1 (GovCloud US-West)</option>
            </optgroup>
            <optgroup label="Europe">
              <option value="eu-west-1">eu-west-1 (Ireland)</option>
              <option value="eu-west-2">eu-west-2 (London)</option>
              <option value="eu-west-3">eu-west-3 (Paris)</option>
              <option value="eu-central-1">eu-central-1 (Frankfurt)</option>
              <option value="eu-central-2">eu-central-2 (Zurich)</option>
              <option value="eu-south-1">eu-south-1 (Milan)</option>
              <option value="eu-south-2">eu-south-2 (Spain)</option>
              <option value="eu-north-1">eu-north-1 (Stockholm)</option>
            </optgroup>
            <optgroup label="Asia Pacific">
              <option value="ap-northeast-1">ap-northeast-1 (Tokyo)</option>
              <option value="ap-northeast-2">ap-northeast-2 (Seoul)</option>
              <option value="ap-northeast-3">ap-northeast-3 (Osaka)</option>
              <option value="ap-south-1">ap-south-1 (Mumbai)</option>
              <option value="ap-south-2">ap-south-2 (Hyderabad)</option>
              <option value="ap-southeast-1">ap-southeast-1 (Singapore)</option>
              <option value="ap-southeast-2">ap-southeast-2 (Sydney)</option>
            </optgroup>
            <optgroup label="Canada">
              <option value="ca-central-1">ca-central-1 (Canada)</option>
            </optgroup>
            <optgroup label="South America">
              <option value="sa-east-1">sa-east-1 (São Paulo)</option>
            </optgroup>
          </select>
          <p class="input-hint">{{ t('admin.accounts.bedrockRegionHint') }}</p>
        </div>

        <!-- Shared: Force Global -->
        <div>
          <label class="create-account-modal__checkbox">
            <input
              v-model="bedrockForceGlobal"
              type="checkbox"
              class="create-account-modal__checkbox-input"
            />
            <span class="create-account-modal__choice-title text-sm">{{ t('admin.accounts.bedrockForceGlobal') }}</span>
          </label>
          <p class="input-hint mt-1">{{ t('admin.accounts.bedrockForceGlobalHint') }}</p>
        </div>

        <!-- Model Restriction Section for Bedrock -->
        <div class="form-section">
          <label class="input-label">{{ t('admin.accounts.modelRestriction') }}</label>

          <!-- Mode Toggle -->
          <div class="mb-4 flex gap-2">
            <button
              type="button"
              @click="modelRestrictionMode = 'whitelist'"
              :class="getModeToggleClasses(modelRestrictionMode === 'whitelist', 'accent')"
            >
              {{ t('admin.accounts.modelWhitelist') }}
            </button>
            <button
              type="button"
              @click="modelRestrictionMode = 'mapping'"
              :class="getModeToggleClasses(modelRestrictionMode === 'mapping', 'purple')"
            >
              {{ t('admin.accounts.modelMapping') }}
            </button>
          </div>

          <!-- Whitelist Mode -->
          <div v-if="modelRestrictionMode === 'whitelist'">
            <ModelWhitelistSelector v-model="allowedModels" platform="anthropic" />
            <p class="create-account-modal__choice-description text-xs">
              {{ t('admin.accounts.selectedModels', { count: allowedModels.length }) }}
              <span v-if="allowedModels.length === 0">{{ t('admin.accounts.supportsAllModels') }}</span>
            </p>
          </div>

          <!-- Mapping Mode -->
          <div v-else class="space-y-3">
            <div v-for="(mapping, index) in modelMappings" :key="index" class="flex items-center gap-2">
              <input v-model="mapping.from" type="text" class="input flex-1" :placeholder="t('admin.accounts.fromModel')" />
              <span class="create-account-modal__choice-description">→</span>
              <input v-model="mapping.to" type="text" class="input flex-1" :placeholder="t('admin.accounts.toModel')" />
              <button type="button" @click="modelMappings.splice(index, 1)" class="create-account-modal__status-chip create-account-modal__status-chip--danger">
                <Icon name="trash" size="sm" />
              </button>
            </div>
            <button type="button" @click="modelMappings.push({ from: '', to: '' })" class="btn btn-secondary text-sm">
              + {{ t('admin.accounts.addMapping') }}
            </button>
            <!-- Bedrock Preset Mappings -->
            <div class="flex flex-wrap gap-2">
              <button
                v-for="preset in bedrockPresets"
                :key="preset.from"
                type="button"
                @click="addPresetMapping(preset.from, preset.to)"
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
              <label class="input-label mb-0">{{ t('admin.accounts.poolMode') }}</label>
              <p class="create-account-modal__choice-description mt-1 text-xs">
                {{ t('admin.accounts.poolModeHint') }}
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
          <div v-if="poolModeEnabled" class="create-account-modal__notice create-account-modal__notice--blue create-account-modal__notice-block">
            <p class="text-xs">
              <Icon name="exclamationCircle" size="sm" class="mr-1 inline" :stroke-width="2" />
              {{ t('admin.accounts.poolModeInfo') }}
            </p>
          </div>
          <div v-if="poolModeEnabled" class="mt-3">
            <label class="input-label">{{ t('admin.accounts.poolModeRetryCount') }}</label>
            <input
              v-model.number="poolModeRetryCount"
              type="number"
              min="0"
              :max="MAX_POOL_MODE_RETRY_COUNT"
              step="1"
              class="input"
            />
            <p class="create-account-modal__choice-description mt-1 text-xs">
              {{
                t('admin.accounts.poolModeRetryCountHint', {
                  default: DEFAULT_POOL_MODE_RETRY_COUNT,
                  max: MAX_POOL_MODE_RETRY_COUNT
                })
              }}
            </p>
          </div>
        </div>
      </div>

      <!-- API Key / Bedrock 账号配额限制 -->
      <div v-if="form.type === 'apikey' || form.type === 'bedrock'" class="form-section space-y-4">
        <div class="mb-3">
          <h3 class="input-label mb-0 text-base font-semibold">{{ t('admin.accounts.quotaLimit') }}</h3>
          <p class="create-account-modal__choice-description mt-1 text-xs">
            {{ t('admin.accounts.quotaLimitHint') }}
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
      </div>

      <!-- OpenAI OAuth Model Mapping (OAuth 类型没有 apikey 容器，需要独立的模型映射区域) -->
      <div
        v-if="form.platform === 'openai' && accountCategory === 'oauth-based'"
        class="form-section"
      >
        <label class="input-label">{{ t('admin.accounts.modelRestriction') }}</label>

        <div
          v-if="isOpenAIModelRestrictionDisabled"
          class="create-account-modal__notice create-account-modal__notice--amber create-account-modal__notice-block mb-3"
        >
          <p class="text-xs">
            {{ t('admin.accounts.openai.modelRestrictionDisabledByPassthrough') }}
          </p>
        </div>

        <template v-else>
          <!-- Mode Toggle -->
          <div class="mb-4 flex gap-2">
            <button
              type="button"
              @click="modelRestrictionMode = 'whitelist'"
              :class="getModeToggleClasses(modelRestrictionMode === 'whitelist', 'accent')"
            >
              {{ t('admin.accounts.modelWhitelist') }}
            </button>
            <button
              type="button"
              @click="modelRestrictionMode = 'mapping'"
              :class="getModeToggleClasses(modelRestrictionMode === 'mapping', 'purple')"
            >
              {{ t('admin.accounts.modelMapping') }}
            </button>
          </div>

          <!-- Whitelist Mode -->
          <div v-if="modelRestrictionMode === 'whitelist'">
            <ModelWhitelistSelector v-model="allowedModels" :platform="form.platform" />
            <p class="create-account-modal__choice-description text-xs">
              {{ t('admin.accounts.selectedModels', { count: allowedModels.length }) }}
              <span v-if="allowedModels.length === 0">{{
                t('admin.accounts.supportsAllModels')
              }}</span>
            </p>
          </div>

          <!-- Mapping Mode -->
          <div v-else>
            <div class="create-account-modal__notice create-account-modal__notice--purple create-account-modal__notice-block mb-3">
              <p class="text-xs">
                {{ t('admin.accounts.mapRequestModels') }}
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
                  class="create-account-modal__choice-description h-4 w-4 flex-shrink-0"
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
                  class="create-account-modal__status-chip create-account-modal__status-chip--danger create-account-modal__status-chip-action"
                >
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
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
              + {{ t('admin.accounts.addMapping') }}
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

      <!-- Temp Unschedulable Rules -->
      <div class="form-section space-y-4">
        <div class="mb-3 flex items-center justify-between">
          <div>
            <label class="input-label mb-0">{{ t('admin.accounts.tempUnschedulable.title') }}</label>
            <p class="create-account-modal__choice-description mt-1 text-xs">
              {{ t('admin.accounts.tempUnschedulable.hint') }}
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
          <div class="create-account-modal__notice create-account-modal__notice--blue create-account-modal__notice-block">
              <p class="text-xs">
                <Icon name="exclamationTriangle" size="sm" class="mr-1 inline" :stroke-width="2" />
                {{ t('admin.accounts.tempUnschedulable.notice') }}
              </p>
            </div>

          <div class="flex flex-wrap gap-2">
            <button
              v-for="preset in tempUnschedPresets"
              :key="preset.label"
              type="button"
              @click="addTempUnschedRule(preset.rule)"
              class="create-account-modal__tag-button"
            >
              + {{ preset.label }}
            </button>
          </div>

          <div v-if="tempUnschedRules.length > 0" class="space-y-3">
            <div
              v-for="(rule, index) in tempUnschedRules"
              :key="getTempUnschedRuleKey(rule)"
              class="create-account-modal__rule-card"
            >
              <div class="mb-2 flex items-center justify-between">
                <span class="create-account-modal__rule-index">
                  {{ t('admin.accounts.tempUnschedulable.ruleIndex', { index: index + 1 }) }}
                </span>
                <div class="flex items-center gap-2">
                  <button
                    type="button"
                    :disabled="index === 0"
                    @click="moveTempUnschedRule(index, -1)"
                    class="create-account-modal__icon-button"
                  >
                    <Icon name="chevronUp" size="sm" :stroke-width="2" />
                  </button>
                  <button
                    type="button"
                    :disabled="index === tempUnschedRules.length - 1"
                    @click="moveTempUnschedRule(index, 1)"
                    class="create-account-modal__icon-button"
                  >
                    <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                    </svg>
                  </button>
                  <button
                    type="button"
                    @click="removeTempUnschedRule(index)"
                    class="create-account-modal__icon-button create-account-modal__icon-button--danger"
                  >
                    <Icon name="x" size="sm" :stroke-width="2" />
                  </button>
                </div>
              </div>

              <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div>
                  <label class="input-label">{{ t('admin.accounts.tempUnschedulable.errorCode') }}</label>
                  <input
                    v-model.number="rule.error_code"
                    type="number"
                    min="100"
                    max="599"
                    class="input"
                    :placeholder="t('admin.accounts.tempUnschedulable.errorCodePlaceholder')"
                  />
                </div>
                <div>
                  <label class="input-label">{{ t('admin.accounts.tempUnschedulable.durationMinutes') }}</label>
                  <input
                    v-model.number="rule.duration_minutes"
                    type="number"
                    min="1"
                    class="input"
                    :placeholder="t('admin.accounts.tempUnschedulable.durationPlaceholder')"
                  />
                </div>
                <div class="sm:col-span-2">
                  <label class="input-label">{{ t('admin.accounts.tempUnschedulable.keywords') }}</label>
                  <input
                    v-model="rule.keywords"
                    type="text"
                    class="input"
                    :placeholder="t('admin.accounts.tempUnschedulable.keywordsPlaceholder')"
                  />
                  <p class="input-hint">{{ t('admin.accounts.tempUnschedulable.keywordsHint') }}</p>
                </div>
                <div class="sm:col-span-2">
                  <label class="input-label">{{ t('admin.accounts.tempUnschedulable.description') }}</label>
                  <input
                    v-model="rule.description"
                    type="text"
                    class="input"
                    :placeholder="t('admin.accounts.tempUnschedulable.descriptionPlaceholder')"
                  />
                </div>
              </div>
            </div>
          </div>

          <button
            type="button"
            @click="addTempUnschedRule()"
            class="create-account-modal__dashed-action w-full"
          >
            <svg
              class="mr-1 inline h-4 w-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            {{ t('admin.accounts.tempUnschedulable.addRule') }}
          </button>
        </div>
      </div>

      <!-- Intercept Warmup Requests (Anthropic/Antigravity) -->
      <div
        v-if="form.platform === 'anthropic' || form.platform === 'antigravity'"
        class="form-section"
      >
        <div class="flex items-center justify-between">
          <div>
            <label class="input-label mb-0">{{
              t('admin.accounts.interceptWarmupRequests')
            }}</label>
            <p class="create-account-modal__choice-description mt-1 text-xs">
              {{ t('admin.accounts.interceptWarmupRequestsDesc') }}
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

      <!-- Quota Control Section (Anthropic OAuth/SetupToken only) -->
      <div
        v-if="form.platform === 'anthropic' && accountCategory === 'oauth-based'"
        class="form-section space-y-4"
      >
        <div class="mb-3">
          <h3 class="input-label mb-0 text-base font-semibold">{{ t('admin.accounts.quotaControl.title') }}</h3>
          <p class="create-account-modal__choice-description mt-1 text-xs">
            {{ t('admin.accounts.quotaControl.hint') }}
          </p>
        </div>

        <!-- Window Cost Limit -->
        <div class="create-account-modal__config-card">
          <div class="mb-3 flex items-center justify-between">
            <div>
              <label class="input-label mb-0">{{ t('admin.accounts.quotaControl.windowCost.label') }}</label>
              <p class="create-account-modal__choice-description mt-1 text-xs">
                {{ t('admin.accounts.quotaControl.windowCost.hint') }}
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

          <div v-if="windowCostEnabled" class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4">
            <div>
              <label class="input-label">{{ t('admin.accounts.quotaControl.windowCost.limit') }}</label>
              <div class="relative">
                <span class="create-account-modal__choice-description absolute left-3 top-1/2 -translate-y-1/2">$</span>
                <input
                  v-model.number="windowCostLimit"
                  type="number"
                  min="0"
                  step="1"
                  class="input pl-7"
                  :placeholder="t('admin.accounts.quotaControl.windowCost.limitPlaceholder')"
                />
              </div>
              <p class="input-hint">{{ t('admin.accounts.quotaControl.windowCost.limitHint') }}</p>
            </div>
            <div>
              <label class="input-label">{{ t('admin.accounts.quotaControl.windowCost.stickyReserve') }}</label>
              <div class="relative">
                <span class="create-account-modal__choice-description absolute left-3 top-1/2 -translate-y-1/2">$</span>
                <input
                  v-model.number="windowCostStickyReserve"
                  type="number"
                  min="0"
                  step="1"
                  class="input pl-7"
                  :placeholder="t('admin.accounts.quotaControl.windowCost.stickyReservePlaceholder')"
                />
              </div>
              <p class="input-hint">{{ t('admin.accounts.quotaControl.windowCost.stickyReserveHint') }}</p>
            </div>
          </div>
        </div>

        <!-- Session Limit -->
        <div class="create-account-modal__config-card">
          <div class="mb-3 flex items-center justify-between">
            <div>
              <label class="input-label mb-0">{{ t('admin.accounts.quotaControl.sessionLimit.label') }}</label>
              <p class="create-account-modal__choice-description mt-1 text-xs">
                {{ t('admin.accounts.quotaControl.sessionLimit.hint') }}
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

          <div v-if="sessionLimitEnabled" class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4">
            <div>
              <label class="input-label">{{ t('admin.accounts.quotaControl.sessionLimit.maxSessions') }}</label>
              <input
                v-model.number="maxSessions"
                type="number"
                min="1"
                step="1"
                class="input"
                :placeholder="t('admin.accounts.quotaControl.sessionLimit.maxSessionsPlaceholder')"
              />
              <p class="input-hint">{{ t('admin.accounts.quotaControl.sessionLimit.maxSessionsHint') }}</p>
            </div>
            <div>
              <label class="input-label">{{ t('admin.accounts.quotaControl.sessionLimit.idleTimeout') }}</label>
              <div class="relative">
                <input
                  v-model.number="sessionIdleTimeout"
                  type="number"
                  min="1"
                  step="1"
                  class="input pr-12"
                  :placeholder="t('admin.accounts.quotaControl.sessionLimit.idleTimeoutPlaceholder')"
                />
                <span class="create-account-modal__choice-description absolute right-3 top-1/2 -translate-y-1/2">{{ t('common.minutes') }}</span>
              </div>
              <p class="input-hint">{{ t('admin.accounts.quotaControl.sessionLimit.idleTimeoutHint') }}</p>
            </div>
          </div>
        </div>

        <!-- RPM Limit -->
        <div class="create-account-modal__config-card">
          <div class="mb-3 flex items-center justify-between">
            <div>
              <label class="input-label mb-0">{{ t('admin.accounts.quotaControl.rpmLimit.label') }}</label>
              <p class="create-account-modal__choice-description mt-1 text-xs">
                {{ t('admin.accounts.quotaControl.rpmLimit.hint') }}
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
              <label class="input-label">{{ t('admin.accounts.quotaControl.rpmLimit.baseRpm') }}</label>
              <input
                v-model.number="baseRpm"
                type="number"
                min="1"
                max="1000"
                step="1"
                class="input"
                :placeholder="t('admin.accounts.quotaControl.rpmLimit.baseRpmPlaceholder')"
              />
              <p class="input-hint">{{ t('admin.accounts.quotaControl.rpmLimit.baseRpmHint') }}</p>
            </div>

            <div>
              <label class="input-label">{{ t('admin.accounts.quotaControl.rpmLimit.strategy') }}</label>
              <div class="flex gap-2">
                <button
                  type="button"
                  @click="rpmStrategy = 'tiered'"
                  :class="getModeToggleClasses(rpmStrategy === 'tiered', 'accent')"
                >
                  <div class="text-center">
                    <div>{{ t('admin.accounts.quotaControl.rpmLimit.strategyTiered') }}</div>
                    <div class="mt-0.5 text-[10px] opacity-70">{{ t('admin.accounts.quotaControl.rpmLimit.strategyTieredHint') }}</div>
                  </div>
                </button>
                <button
                  type="button"
                  @click="rpmStrategy = 'sticky_exempt'"
                  :class="getModeToggleClasses(rpmStrategy === 'sticky_exempt', 'accent')"
                >
                  <div class="text-center">
                    <div>{{ t('admin.accounts.quotaControl.rpmLimit.strategyStickyExempt') }}</div>
                    <div class="mt-0.5 text-[10px] opacity-70">{{ t('admin.accounts.quotaControl.rpmLimit.strategyStickyExemptHint') }}</div>
                  </div>
                </button>
              </div>
            </div>

            <div v-if="rpmStrategy === 'tiered'">
              <label class="input-label">{{ t('admin.accounts.quotaControl.rpmLimit.stickyBuffer') }}</label>
              <input
                v-model.number="rpmStickyBuffer"
                type="number"
                min="1"
                step="1"
                class="input"
                :placeholder="t('admin.accounts.quotaControl.rpmLimit.stickyBufferPlaceholder')"
              />
              <p class="input-hint">{{ t('admin.accounts.quotaControl.rpmLimit.stickyBufferHint') }}</p>
            </div>

          </div>

          <!-- 用户消息限速模式（独立于 RPM 开关，始终可见） -->
          <div class="mt-4">
            <label class="input-label">{{ t('admin.accounts.quotaControl.rpmLimit.userMsgQueue') }}</label>
            <p class="create-account-modal__choice-description mt-1 mb-2 text-xs">
              {{ t('admin.accounts.quotaControl.rpmLimit.userMsgQueueHint') }}
            </p>
            <div class="flex space-x-2">
              <button type="button" v-for="opt in umqModeOptions" :key="opt.value"
                @click="userMsgQueueMode = opt.value"
                :class="getSegmentOptionClasses(userMsgQueueMode === opt.value)">
                {{ opt.label }}
              </button>
            </div>
          </div>
        </div>

        <!-- TLS Fingerprint -->
        <div class="create-account-modal__config-card">
          <div class="flex items-center justify-between">
            <div>
              <label class="input-label mb-0">{{ t('admin.accounts.quotaControl.tlsFingerprint.label') }}</label>
              <p class="create-account-modal__choice-description mt-1 text-xs">
                {{ t('admin.accounts.quotaControl.tlsFingerprint.hint') }}
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
              <option :value="null">{{ t('admin.accounts.quotaControl.tlsFingerprint.defaultProfile') }}</option>
              <option v-if="tlsFingerprintProfiles.length > 0" :value="-1">{{ t('admin.accounts.quotaControl.tlsFingerprint.randomProfile') }}</option>
              <option v-for="p in tlsFingerprintProfiles" :key="p.id" :value="p.id">{{ p.name }}</option>
            </select>
          </div>
        </div>

        <!-- Session ID Masking -->
        <div class="create-account-modal__config-card">
          <div class="flex items-center justify-between">
            <div>
              <label class="input-label mb-0">{{ t('admin.accounts.quotaControl.sessionIdMasking.label') }}</label>
              <p class="create-account-modal__choice-description mt-1 text-xs">
                {{ t('admin.accounts.quotaControl.sessionIdMasking.hint') }}
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
        <div class="create-account-modal__config-card">
          <div class="flex items-center justify-between">
            <div>
              <label class="input-label mb-0">{{ t('admin.accounts.quotaControl.cacheTTLOverride.label') }}</label>
              <p class="create-account-modal__choice-description mt-1 text-xs">
                {{ t('admin.accounts.quotaControl.cacheTTLOverride.hint') }}
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
            <label class="input-label text-xs">{{ t('admin.accounts.quotaControl.cacheTTLOverride.target') }}</label>
            <select
              v-model="cacheTTLOverrideTarget"
              class="input mt-1"
            >
              <option value="5m">5m</option>
              <option value="1h">1h</option>
            </select>
            <p class="input-hint mt-1">
              {{ t('admin.accounts.quotaControl.cacheTTLOverride.targetHint') }}
            </p>
          </div>
        </div>

        <!-- Custom Base URL Relay -->
        <div class="create-account-modal__config-card">
          <div class="flex items-center justify-between">
            <div>
              <label class="input-label mb-0">{{ t('admin.accounts.quotaControl.customBaseUrl.label') }}</label>
              <p class="create-account-modal__choice-description mt-1 text-xs">
                {{ t('admin.accounts.quotaControl.customBaseUrl.hint') }}
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
              :placeholder="t('admin.accounts.quotaControl.customBaseUrl.urlHint')"
            />
          </div>
        </div>
      </div>

      <div>
        <label class="input-label">{{ t('admin.accounts.proxy') }}</label>
        <ProxySelector v-model="form.proxy_id" :proxies="proxies" />
      </div>

      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4 lg:grid-cols-4">
        <div>
          <label class="input-label">{{ t('admin.accounts.concurrency') }}</label>
          <input v-model.number="form.concurrency" type="number" min="1" class="input"
            @input="form.concurrency = Math.max(1, form.concurrency || 1)" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.loadFactor') }}</label>
          <input v-model.number="form.load_factor" type="number" min="1"
            class="input" :placeholder="String(form.concurrency || 1)"
            @input="form.load_factor = (form.load_factor &amp;&amp; form.load_factor >= 1) ? form.load_factor : null" />
          <p class="input-hint">{{ t('admin.accounts.loadFactorHint') }}</p>
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.priority') }}</label>
          <input
            v-model.number="form.priority"
            type="number"
            min="1"
            class="input"
            data-tour="account-form-priority"
          />
          <p class="input-hint">{{ t('admin.accounts.priorityHint') }}</p>
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.billingRateMultiplier') }}</label>
          <input v-model.number="form.rate_multiplier" type="number" min="0" step="0.001" class="input" />
          <p class="input-hint">{{ t('admin.accounts.billingRateMultiplierHint') }}</p>
        </div>
      </div>
      <div class="form-section">
        <label class="input-label">{{ t('admin.accounts.expiresAt') }}</label>
        <input v-model="expiresAtInput" type="datetime-local" class="input" />
        <p class="input-hint">{{ t('admin.accounts.expiresAtHint') }}</p>
      </div>

      <!-- OpenAI 自动透传开关（OAuth/API Key） -->
      <div
        v-if="form.platform === 'openai'"
        class="form-section"
      >
        <div class="flex items-center justify-between">
          <div>
            <label class="input-label mb-0">{{ t('admin.accounts.openai.oauthPassthrough') }}</label>
            <p class="create-account-modal__choice-description mt-1 text-xs">
              {{ t('admin.accounts.openai.oauthPassthroughDesc') }}
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
        v-if="form.platform === 'openai' && (accountCategory === 'oauth-based' || accountCategory === 'apikey')"
        class="form-section"
      >
        <div class="flex items-center justify-between">
          <div>
            <label class="input-label mb-0">{{ t('admin.accounts.openai.wsMode') }}</label>
            <p class="create-account-modal__choice-description mt-1 text-xs">
              {{ t('admin.accounts.openai.wsModeDesc') }}
            </p>
            <p class="create-account-modal__choice-description mt-1 text-xs">
              {{ t(openAIWSModeConcurrencyHintKey) }}
            </p>
          </div>
          <div class="w-52">
            <Select v-model="openaiResponsesWebSocketV2Mode" :options="openAIWSModeOptions" />
          </div>
        </div>
      </div>

      <!-- Anthropic API Key 自动透传开关 -->
      <div
        v-if="form.platform === 'anthropic' && accountCategory === 'apikey'"
        class="form-section"
      >
        <div class="flex items-center justify-between">
          <div>
            <label class="input-label mb-0">{{ t('admin.accounts.anthropic.apiKeyPassthrough') }}</label>
            <p class="create-account-modal__choice-description mt-1 text-xs">
              {{ t('admin.accounts.anthropic.apiKeyPassthroughDesc') }}
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

      <!-- OpenAI OAuth Codex 官方客户端限制开关 -->
      <div
        v-if="form.platform === 'openai' && accountCategory === 'oauth-based'"
        class="form-section"
      >
        <div class="flex items-center justify-between">
          <div>
            <label class="input-label mb-0">{{ t('admin.accounts.openai.codexCLIOnly') }}</label>
            <p class="create-account-modal__choice-description mt-1 text-xs">
              {{ t('admin.accounts.openai.codexCLIOnlyDesc') }}
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
              t('admin.accounts.autoPauseOnExpired')
            }}</label>
            <p class="create-account-modal__choice-description mt-1 text-xs">
              {{ t('admin.accounts.autoPauseOnExpiredDesc') }}
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

      <div class="create-account-modal__extra-options">
        <!-- Mixed Scheduling (only for antigravity accounts) -->
        <div v-if="form.platform === 'antigravity'" class="flex items-center gap-2">
          <label class="create-account-modal__checkbox">
            <input
              type="checkbox"
              v-model="mixedScheduling"
              class="create-account-modal__checkbox-input"
            />
            <span class="create-account-modal__choice-title text-sm">
              {{ t('admin.accounts.mixedScheduling') }}
            </span>
          </label>
          <div class="group relative">
            <span class="create-account-modal__tooltip-trigger">
              ?
            </span>
            <div class="create-account-modal__tooltip-panel">
              {{ t('admin.accounts.mixedSchedulingTooltip') }}
              <div class="create-account-modal__tooltip-arrow"></div>
            </div>
          </div>
        </div>
        <div v-if="form.platform === 'antigravity'" class="mt-3 flex items-center gap-2">
          <label class="create-account-modal__checkbox">
            <input
              type="checkbox"
              v-model="allowOverages"
              class="create-account-modal__checkbox-input"
            />
            <span class="create-account-modal__choice-title text-sm">
              {{ t('admin.accounts.allowOverages') }}
            </span>
          </label>
          <div class="group relative">
            <span class="create-account-modal__tooltip-trigger">
              ?
            </span>
            <div class="create-account-modal__tooltip-panel">
              {{ t('admin.accounts.allowOveragesTooltip') }}
              <div class="create-account-modal__tooltip-arrow"></div>
            </div>
          </div>
        </div>

        <!-- Group Selection - 仅标准模式显示 -->
        <GroupSelector
          v-if="!authStore.isSimpleMode"
          v-model="form.group_ids"
          :groups="groups"
          :platform="form.platform"
          :mixed-scheduling="mixedScheduling"
          data-tour="account-form-groups"
        />
      </div>

    </form>

    <!-- Step 2: OAuth Authorization -->
    <div v-else class="space-y-5">
      <OAuthAuthorizationFlow
        ref="oauthFlowRef"
        :add-method="form.platform === 'anthropic' ? addMethod : 'oauth'"
        :auth-url="currentOAuthState.authUrl"
        :session-id="currentOAuthState.sessionId"
        :loading="currentOAuthState.loading"
        :error="currentOAuthState.error"
        :show-help="form.platform === 'anthropic'"
        :show-proxy-warning="form.platform !== 'openai' && !!form.proxy_id"
        :allow-multiple="form.platform === 'anthropic'"
        :show-cookie-option="form.platform === 'anthropic'"
        :show-refresh-token-option="form.platform === 'openai' || form.platform === 'antigravity'"
        :show-mobile-refresh-token-option="form.platform === 'openai'"
        :platform="form.platform"
        :show-project-id="geminiOAuthType === 'code_assist'"
        @generate-url="handleGenerateUrl"
        @cookie-auth="handleCookieAuth"
        @validate-refresh-token="handleValidateRefreshToken"
        @validate-mobile-refresh-token="handleOpenAIValidateMobileRT"
      />

    </div>

    <template #footer>
      <div v-if="step === 1" class="flex justify-end gap-3">
        <button @click="handleClose" type="button" class="btn btn-secondary">
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          form="create-account-form"
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
          {{
            isOAuthFlow
              ? t('common.next')
              : submitting
                ? t('admin.accounts.creating')
                : t('common.create')
          }}
        </button>
      </div>
      <div v-else class="flex justify-between gap-3">
        <button type="button" class="btn btn-secondary" @click="goBackToBasicInfo">
          {{ t('common.back') }}
        </button>
        <button
          v-if="isManualInputMethod"
          type="button"
          :disabled="!canExchangeCode"
          class="btn btn-primary"
          @click="handleExchangeCode"
        >
          <svg
            v-if="currentOAuthState.loading"
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
          {{
            currentOAuthState.loading
              ? t('admin.accounts.oauth.verifying')
              : t('admin.accounts.oauth.completeAuth')
          }}
        </button>
      </div>
    </template>
  </BaseDialog>

  <!-- Gemini Help Dialog -->
  <BaseDialog
    :show="showGeminiHelpDialog"
    :title="t('admin.accounts.gemini.helpDialog.title')"
    @close="showGeminiHelpDialog = false"
    width="wide"
  >
    <div class="space-y-6">
      <!-- Setup Guide Section -->
      <div>
        <h3 class="create-account-modal__dialog-title">
          {{ t('admin.accounts.gemini.setupGuide.title') }}
        </h3>
        <div class="space-y-4">
          <div>
            <p class="create-account-modal__dialog-subtitle">
              {{ t('admin.accounts.gemini.setupGuide.checklistTitle') }}
            </p>
            <ul class="create-account-modal__dialog-list">
              <li>{{ t('admin.accounts.gemini.setupGuide.checklistItems.usIp') }}</li>
              <li>{{ t('admin.accounts.gemini.setupGuide.checklistItems.age') }}</li>
            </ul>
          </div>
          <div>
            <p class="create-account-modal__dialog-subtitle">
              {{ t('admin.accounts.gemini.setupGuide.activationTitle') }}
            </p>
            <ul class="create-account-modal__dialog-list">
              <li>{{ t('admin.accounts.gemini.setupGuide.activationItems.geminiWeb') }}</li>
              <li>{{ t('admin.accounts.gemini.setupGuide.activationItems.gcpProject') }}</li>
            </ul>
            <div class="mt-2 flex flex-wrap gap-2">
              <a
                href="https://policies.google.com/terms"
                target="_blank"
                rel="noreferrer"
                class="create-account-modal__link text-sm"
              >
                {{ t('admin.accounts.gemini.setupGuide.links.countryCheck') }}
              </a>
              <span class="create-account-modal__choice-description">·</span>
              <a
                href="https://policies.google.com/country-association-form"
                target="_blank"
                rel="noreferrer"
                class="create-account-modal__link text-sm"
              >
                修改归属地
              </a>
              <span class="create-account-modal__choice-description">·</span>
              <a
                href="https://gemini.google.com/gems/create?hl=en-US&pli=1"
                target="_blank"
                rel="noreferrer"
                class="create-account-modal__link text-sm"
              >
                {{ t('admin.accounts.gemini.setupGuide.links.geminiWebActivation') }}
              </a>
              <span class="create-account-modal__choice-description">·</span>
              <a
                href="https://console.cloud.google.com"
                target="_blank"
                rel="noreferrer"
                class="create-account-modal__link text-sm"
              >
                {{ t('admin.accounts.gemini.setupGuide.links.gcpProject') }}
              </a>
            </div>
          </div>
        </div>
      </div>

      <!-- Quota Policy Section -->
      <div class="form-section pt-6">
        <h3 class="create-account-modal__choice-title mb-3 text-sm font-semibold">
          {{ t('admin.accounts.gemini.quotaPolicy.title') }}
        </h3>
        <p class="create-account-modal__notice create-account-modal__notice--amber create-account-modal__notice-inline mb-4 text-xs">
          {{ t('admin.accounts.gemini.quotaPolicy.note') }}
        </p>
        <div class="overflow-x-auto">
          <table class="w-full text-xs">
            <thead class="create-account-modal__table-head">
              <tr>
                <th class="create-account-modal__table-heading">
                  {{ t('admin.accounts.gemini.quotaPolicy.columns.channel') }}
                </th>
                <th class="create-account-modal__table-heading">
                  {{ t('admin.accounts.gemini.quotaPolicy.columns.account') }}
                </th>
                <th class="create-account-modal__table-heading">
                  {{ t('admin.accounts.gemini.quotaPolicy.columns.limits') }}
                </th>
              </tr>
            </thead>
            <tbody class="create-account-modal__table-body">
              <tr>
                <td class="create-account-modal__table-primary">
                  {{ t('admin.accounts.gemini.quotaPolicy.rows.googleOne.channel') }}
                </td>
                <td class="create-account-modal__table-secondary">Free</td>
                <td class="create-account-modal__table-secondary">
                  {{ t('admin.accounts.gemini.quotaPolicy.rows.googleOne.limitsFree') }}
                </td>
              </tr>
              <tr>
                <td class="create-account-modal__table-primary"></td>
                <td class="create-account-modal__table-secondary">Pro</td>
                <td class="create-account-modal__table-secondary">
                  {{ t('admin.accounts.gemini.quotaPolicy.rows.googleOne.limitsPro') }}
                </td>
              </tr>
              <tr>
                <td class="create-account-modal__table-primary"></td>
                <td class="create-account-modal__table-secondary">Ultra</td>
                <td class="create-account-modal__table-secondary">
                  {{ t('admin.accounts.gemini.quotaPolicy.rows.googleOne.limitsUltra') }}
                </td>
              </tr>
              <tr>
                <td class="create-account-modal__table-primary">
                  {{ t('admin.accounts.gemini.quotaPolicy.rows.gcp.channel') }}
                </td>
                <td class="create-account-modal__table-secondary">Standard</td>
                <td class="create-account-modal__table-secondary">
                  {{ t('admin.accounts.gemini.quotaPolicy.rows.gcp.limitsStandard') }}
                </td>
              </tr>
              <tr>
                <td class="create-account-modal__table-primary"></td>
                <td class="create-account-modal__table-secondary">Enterprise</td>
                <td class="create-account-modal__table-secondary">
                  {{ t('admin.accounts.gemini.quotaPolicy.rows.gcp.limitsEnterprise') }}
                </td>
              </tr>
              <tr>
                <td class="create-account-modal__table-primary">
                  {{ t('admin.accounts.gemini.quotaPolicy.rows.aiStudio.channel') }}
                </td>
                <td class="create-account-modal__table-secondary">Free</td>
                <td class="create-account-modal__table-secondary">
                  {{ t('admin.accounts.gemini.quotaPolicy.rows.aiStudio.limitsFree') }}
                </td>
              </tr>
              <tr>
                <td class="create-account-modal__table-primary"></td>
                <td class="create-account-modal__table-secondary">Paid</td>
                <td class="create-account-modal__table-secondary">
                  {{ t('admin.accounts.gemini.quotaPolicy.rows.aiStudio.limitsPaid') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="mt-4 flex flex-wrap gap-3">
          <a
            :href="geminiQuotaDocs.codeAssist"
            target="_blank"
            rel="noreferrer"
            class="create-account-modal__link text-sm"
          >
            {{ t('admin.accounts.gemini.quotaPolicy.docs.codeAssist') }}
          </a>
          <a
            :href="geminiQuotaDocs.aiStudio"
            target="_blank"
            rel="noreferrer"
            class="create-account-modal__link text-sm"
          >
            {{ t('admin.accounts.gemini.quotaPolicy.docs.aiStudio') }}
          </a>
          <a
            :href="geminiQuotaDocs.vertex"
            target="_blank"
            rel="noreferrer"
            class="create-account-modal__link text-sm"
          >
            {{ t('admin.accounts.gemini.quotaPolicy.docs.vertex') }}
          </a>
        </div>
      </div>

      <!-- API Key Links Section -->
      <div class="form-section pt-6">
        <h3 class="create-account-modal__choice-title mb-3 text-sm font-semibold">
          {{ t('admin.accounts.gemini.helpDialog.apiKeySection') }}
        </h3>
        <div class="flex flex-wrap gap-3">
          <a
            :href="geminiHelpLinks.apiKey"
            target="_blank"
            rel="noreferrer"
            class="create-account-modal__link text-sm"
          >
            {{ t('admin.accounts.gemini.accountType.apiKeyLink') }}
          </a>
          <a
            :href="geminiHelpLinks.aiStudioPricing"
            target="_blank"
            rel="noreferrer"
            class="create-account-modal__link text-sm"
          >
            {{ t('admin.accounts.gemini.accountType.quotaLink') }}
          </a>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button @click="showGeminiHelpDialog = false" type="button" class="btn btn-primary">
          {{ t('common.close') }}
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
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import {
  claudeModels,
  getPresetMappingChipClasses,
  getPresetMappingsByPlatform,
  getModelsByPlatform,
  commonErrorCodes,
  fetchAntigravityDefaultMappings,
  isValidWildcardPattern
} from '@/composables/useModelWhitelist'
import { useAuthStore } from '@/stores/auth'
import { adminAPI } from '@/api/admin'
import {
  useAccountOAuth,
  type AddMethod,
  type AuthInputMethod
} from '@/composables/useAccountOAuth'
import { useOpenAIOAuth } from '@/composables/useOpenAIOAuth'
import { useGeminiOAuth } from '@/composables/useGeminiOAuth'
import { useAntigravityOAuth } from '@/composables/useAntigravityOAuth'
import type {
  Proxy,
  AdminGroup,
  AccountPlatform,
  AccountType,
  CheckMixedChannelResponse,
  CreateAccountRequest
} from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import ProxySelector from '@/components/common/ProxySelector.vue'
import GroupSelector from '@/components/common/GroupSelector.vue'
import ModelWhitelistSelector from '@/components/account/ModelWhitelistSelector.vue'
import QuotaLimitCard from '@/components/account/QuotaLimitCard.vue'
import {
  buildOpenAICompatibleBaseUrlPresets,
  buildAccountOpenAIWSModeOptions,
  buildAccountQuotaExtra,
  buildAccountTempUnschedPresets,
  buildAccountUmqModeOptions,
  buildMixedChannelDetails,
  createDefaultCreateAccountForm,
  geminiHelpLinks,
  geminiQuotaDocs,
  needsMixedChannelCheck,
  resetCreateAccountForm,
  resolveAccountApiKeyHint,
  resolveAccountApiKeyPlaceholder,
  resolveAccountBaseUrlHint,
  resolveAccountBaseUrlPlaceholder,
  resolveCreateAccountOAuthStepTitle,
  resolveMixedChannelWarningMessage,
  type CreateAccountForm
} from '@/components/account/accountModalShared'
import {
  buildCreateAccountRequest,
  buildCreateAccountSharedPayload,
  buildCreateApiKeyCredentials,
  buildCreateAnthropicOAuthAccountPayload,
  buildCreateBatchAccountName,
  buildCreateOpenAICompatOAuthTarget,
  buildCreateAnthropicExtra,
  buildCreateAnthropicQuotaControlExtra,
  buildCreateAntigravityOAuthCredentials,
  buildCreateAntigravityUpstreamCredentials,
  buildCreateAntigravityExtra,
  buildCreateBedrockCredentials,
  buildCreateOpenAIExtra,
  resolveBatchCreateOutcome,
  resolveCreateAccountGeminiSelectedTier,
  resolveCreateAccountOAuthFlow
} from '@/components/account/createAccountModalHelpers'
import {
  appendEmptyModelMapping,
  appendPresetModelMapping,
  applyTempUnschedCredentialsState,
  confirmCustomErrorCodeSelection,
  removeModelMappingAt
} from '@/components/account/accountModalInteractions'
import {
  assignBuiltModelMapping,
  buildTempUnschedRules,
  createTempUnschedRule,
  DEFAULT_POOL_MODE_RETRY_COUNT,
  getDefaultBaseURL,
  MAX_POOL_MODE_RETRY_COUNT,
  moveItemInPlace,
  type ModelMapping,
  type TempUnschedRuleForm
} from '@/components/account/credentialsBuilder'
import { formatDateTimeLocalInput, parseDateTimeLocalInput } from '@/utils/format'
import { createStableObjectKeyResolver } from '@/utils/stableObjectKey'
import {
  // OPENAI_WS_MODE_CTX_POOL,
  OPENAI_WS_MODE_OFF,
  resolveOpenAIWSModeConcurrencyHintKey,
  type OpenAIWSMode
} from '@/utils/openaiWsMode'
import {
  consumeValidationFailureMessage,
  resolveAnthropicExchangeEndpoint,
  resolveOAuthExchangeState,
  runBatchCreateFlow,
  runOAuthExchangeFlow
} from '@/components/account/oauthAuthorizationFlowHelpers'
import OAuthAuthorizationFlow from './OAuthAuthorizationFlow.vue'

// Type for exposed OAuthAuthorizationFlow component
// Note: defineExpose automatically unwraps refs, so we use the unwrapped types
interface OAuthFlowExposed {
  authCode: string
  oauthState: string
  projectId: string
  sessionKey: string
  refreshToken: string
  sessionToken: string
  inputMethod: AuthInputMethod
  reset: () => void
}

const { t } = useI18n()
const authStore = useAuthStore()

const oauthStepTitle = computed(() => {
  return resolveCreateAccountOAuthStepTitle(form.platform, t)
})

// Platform-specific hints for API Key type
const baseUrlHint = computed(() => {
  return resolveAccountBaseUrlHint(form.platform, t)
})

const baseUrlPlaceholder = computed(() => {
  return resolveAccountBaseUrlPlaceholder(form.platform, t)
})

const apiKeyHint = computed(() => {
  return resolveAccountApiKeyHint(form.platform, t)
})

const apiKeyPlaceholder = computed(() => {
  return resolveAccountApiKeyPlaceholder(form.platform, t)
})

const openAICompatibleBaseUrlPresets = computed(() => {
  return buildOpenAICompatibleBaseUrlPresets(t)
})

interface Props {
  show: boolean
  proxies: Proxy[]
  groups: AdminGroup[]
}

const props = defineProps<Props>()
const emit = defineEmits<{
  close: []
  created: []
}>()

const appStore = useAppStore()

// OAuth composables
const oauth = useAccountOAuth() // For Anthropic OAuth
const openaiOAuth = useOpenAIOAuth() // For OpenAI OAuth
const geminiOAuth = useGeminiOAuth() // For Gemini OAuth
const antigravityOAuth = useAntigravityOAuth() // For Antigravity OAuth

const currentOAuthState = computed(() => {
  if (form.platform === 'openai') {
    return {
      authUrl: openaiOAuth.authUrl.value,
      sessionId: openaiOAuth.sessionId.value,
      loading: openaiOAuth.loading.value,
      error: openaiOAuth.error.value
    }
  }
  if (form.platform === 'gemini') {
    return {
      authUrl: geminiOAuth.authUrl.value,
      sessionId: geminiOAuth.sessionId.value,
      loading: geminiOAuth.loading.value,
      error: geminiOAuth.error.value
    }
  }
  if (form.platform === 'antigravity') {
    return {
      authUrl: antigravityOAuth.authUrl.value,
      sessionId: antigravityOAuth.sessionId.value,
      loading: antigravityOAuth.loading.value,
      error: antigravityOAuth.error.value
    }
  }
  return {
    authUrl: oauth.authUrl.value,
    sessionId: oauth.sessionId.value,
    loading: oauth.loading.value,
    error: oauth.error.value
  }
})

// Refs
const oauthFlowRef = ref<OAuthFlowExposed | null>(null)

// State
const step = ref(1)
const submitting = ref(false)
const accountCategory = ref<'oauth-based' | 'apikey' | 'bedrock'>('oauth-based') // UI selection for account category
const addMethod = ref<AddMethod>('oauth') // For oauth-based: 'oauth' or 'setup-token'
const apiKeyBaseUrl = ref(getDefaultBaseURL('anthropic'))
const apiKeyValue = ref('')
const editQuotaLimit = ref<number | null>(null)
const editQuotaDailyLimit = ref<number | null>(null)
const editQuotaWeeklyLimit = ref<number | null>(null)
const editDailyResetMode = ref<'rolling' | 'fixed' | null>(null)
const editDailyResetHour = ref<number | null>(null)
const editWeeklyResetMode = ref<'rolling' | 'fixed' | null>(null)
const editWeeklyResetDay = ref<number | null>(null)
const editWeeklyResetHour = ref<number | null>(null)
const editResetTimezone = ref<string | null>(null)
const modelMappings = ref<ModelMapping[]>([])
const modelRestrictionMode = ref<'whitelist' | 'mapping'>('whitelist')
const allowedModels = ref<string[]>([])
const poolModeEnabled = ref(false)
const poolModeRetryCount = ref(DEFAULT_POOL_MODE_RETRY_COUNT)
const customErrorCodesEnabled = ref(false)
const selectedErrorCodes = ref<number[]>([])
const customErrorCodeInput = ref<number | null>(null)
const interceptWarmupRequests = ref(false)
const autoPauseOnExpired = ref(true)
const openaiPassthroughEnabled = ref(false)
const openaiOAuthResponsesWebSocketV2Mode = ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF)
const openaiAPIKeyResponsesWebSocketV2Mode = ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF)
const codexCLIOnlyEnabled = ref(false)
const anthropicPassthroughEnabled = ref(false)
const mixedScheduling = ref(false) // For antigravity accounts: enable mixed scheduling
const allowOverages = ref(false) // For antigravity accounts: enable AI Credits overages
const antigravityAccountType = ref<'oauth' | 'upstream'>('oauth') // For antigravity: oauth or upstream
const upstreamBaseUrl = ref('') // For upstream type: base URL
const upstreamApiKey = ref('') // For upstream type: API key
const antigravityModelRestrictionMode = ref<'whitelist' | 'mapping'>('whitelist')
const antigravityWhitelistModels = ref<string[]>([])
const antigravityModelMappings = ref<ModelMapping[]>([])
const antigravityPresetMappings = computed(() => getPresetMappingsByPlatform('antigravity'))
const bedrockPresets = computed(() => getPresetMappingsByPlatform('bedrock'))

// Bedrock credentials
const bedrockAuthMode = ref<'sigv4' | 'apikey'>('sigv4')
const bedrockAccessKeyId = ref('')
const bedrockSecretAccessKey = ref('')
const bedrockSessionToken = ref('')
const bedrockRegion = ref('us-east-1')
const bedrockForceGlobal = ref(false)
const bedrockApiKeyValue = ref('')
const tempUnschedEnabled = ref(false)
const tempUnschedRules = ref<TempUnschedRuleForm[]>([])
const getModelMappingKey = createStableObjectKeyResolver<ModelMapping>('create-model-mapping')
const getAntigravityModelMappingKey = createStableObjectKeyResolver<ModelMapping>('create-antigravity-model-mapping')
const getTempUnschedRuleKey = createStableObjectKeyResolver<TempUnschedRuleForm>('create-temp-unsched-rule')
const geminiOAuthType = ref<'code_assist' | 'google_one' | 'ai_studio'>('google_one')
const geminiAIStudioOAuthEnabled = ref(false)

const showMixedChannelWarning = ref(false)
const mixedChannelWarningDetails = ref<{ groupName: string; currentPlatform: string; otherPlatform: string } | null>(
  null
)
const mixedChannelWarningRawMessage = ref('')
const mixedChannelWarningAction = ref<(() => Promise<void>) | null>(null)
const antigravityMixedChannelConfirmed = ref(false)
const showAdvancedOAuth = ref(false)
const showGeminiHelpDialog = ref(false)

// Quota control state (Anthropic OAuth/SetupToken only)
const windowCostEnabled = ref(false)
const windowCostLimit = ref<number | null>(null)
const windowCostStickyReserve = ref<number | null>(null)
const sessionLimitEnabled = ref(false)
const maxSessions = ref<number | null>(null)
const sessionIdleTimeout = ref<number | null>(null)
const rpmLimitEnabled = ref(false)
const baseRpm = ref<number | null>(null)
const rpmStrategy = ref<'tiered' | 'sticky_exempt'>('tiered')
const rpmStickyBuffer = ref<number | null>(null)
const userMsgQueueMode = ref('')
const umqModeOptions = computed(() => buildAccountUmqModeOptions(t))
const tlsFingerprintEnabled = ref(false)
const tlsFingerprintProfileId = ref<number | null>(null)
const tlsFingerprintProfiles = ref<{ id: number; name: string }[]>([])
const sessionIdMaskingEnabled = ref(false)
const cacheTTLOverrideEnabled = ref(false)
const cacheTTLOverrideTarget = ref<string>('5m')
const customBaseUrlEnabled = ref(false)
const customBaseUrl = ref('')

// Gemini tier selection (used as fallback when auto-detection is unavailable/fails)
const geminiTierGoogleOne = ref<'google_one_free' | 'google_ai_pro' | 'google_ai_ultra'>('google_one_free')
const geminiTierGcp = ref<'gcp_standard' | 'gcp_enterprise'>('gcp_standard')
const geminiTierAIStudio = ref<'aistudio_free' | 'aistudio_paid'>('aistudio_free')

const geminiSelectedTier = computed(() => {
  return resolveCreateAccountGeminiSelectedTier({
    accountCategory: accountCategory.value,
    geminiOAuthType: geminiOAuthType.value,
    geminiTierAIStudio: geminiTierAIStudio.value,
    geminiTierGcp: geminiTierGcp.value,
    geminiTierGoogleOne: geminiTierGoogleOne.value,
    platform: form.platform
  })
})

const openAIWSModeOptions = computed(() => buildAccountOpenAIWSModeOptions(t))

const openaiResponsesWebSocketV2Mode = computed({
  get: () => {
    if (form.platform === 'openai' && accountCategory.value === 'apikey') {
      return openaiAPIKeyResponsesWebSocketV2Mode.value
    }
    return openaiOAuthResponsesWebSocketV2Mode.value
  },
  set: (mode: OpenAIWSMode) => {
    if (form.platform === 'openai' && accountCategory.value === 'apikey') {
      openaiAPIKeyResponsesWebSocketV2Mode.value = mode
      return
    }
    openaiOAuthResponsesWebSocketV2Mode.value = mode
  }
})

const openAIWSModeConcurrencyHintKey = computed(() =>
  resolveOpenAIWSModeConcurrencyHintKey(openaiResponsesWebSocketV2Mode.value)
)

const isOpenAIModelRestrictionDisabled = computed(() =>
  form.platform === 'openai' && openaiPassthroughEnabled.value
)

const mixedChannelWarningMessageText = computed(() => {
  return resolveMixedChannelWarningMessage({
    details: mixedChannelWarningDetails.value,
    rawMessage: mixedChannelWarningRawMessage.value,
    t
  })
})

type CreateAccountTone = 'rose' | 'orange' | 'purple' | 'amber' | 'green' | 'blue' | 'emerald'
type CreateAccountModeTone = 'accent' | 'purple' | 'danger'

function joinClassNames(classNames: Array<string | false | null | undefined>) {
  return classNames.filter(Boolean).join(' ')
}

function getChoiceCardClasses(isSelected: boolean, tone: CreateAccountTone, isDisabled = false) {
  return joinClassNames([
    'create-account-modal__choice-card create-account-modal__choice-card-control flex items-center gap-3 border-2 text-left transition-all',
    isSelected ? `create-account-modal__choice-card--${tone}` : 'create-account-modal__choice-card--idle',
    isDisabled && 'create-account-modal__choice-card--disabled'
  ])
}

function getChoiceIconClasses(isSelected: boolean, tone: CreateAccountTone) {
  return joinClassNames([
    'create-account-modal__choice-icon create-account-modal__choice-icon-control flex h-8 w-8 shrink-0 items-center justify-center',
    isSelected ? `create-account-modal__choice-icon--${tone}` : 'create-account-modal__choice-icon--idle'
  ])
}

function getToneTagClasses(tone: CreateAccountTone) {
  return joinClassNames([
    'create-account-modal__tone-tag create-account-modal__tone-tag-control text-[10px] font-semibold',
    `create-account-modal__tone-tag--${tone}`
  ])
}

function getModeToggleClasses(isSelected: boolean, tone: CreateAccountModeTone) {
  return joinClassNames([
    'create-account-modal__mode-toggle create-account-modal__mode-toggle-control flex-1 text-sm font-medium transition-all',
    isSelected
      ? `create-account-modal__mode-toggle--${tone}`
      : 'create-account-modal__mode-toggle--idle'
  ])
}

function getSwitchTrackClasses(isEnabled: boolean) {
  return joinClassNames([
    'create-account-modal__switch relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
    isEnabled ? 'create-account-modal__switch--enabled' : 'create-account-modal__switch--disabled'
  ])
}

function getSwitchThumbClasses(isEnabled: boolean) {
  return joinClassNames([
    'create-account-modal__switch-thumb pointer-events-none inline-block h-5 w-5 transform rounded-full shadow ring-0 transition duration-200 ease-in-out',
    isEnabled ? 'translate-x-5' : 'translate-x-0'
  ])
}

function getStatusChipClasses(isSelected: boolean, tone: CreateAccountModeTone = 'danger') {
  return joinClassNames([
    'create-account-modal__status-chip create-account-modal__status-chip-control text-sm font-medium transition-colors',
    isSelected
      ? `create-account-modal__status-chip--${tone}`
      : 'create-account-modal__status-chip--idle'
  ])
}

function getValidationInputClasses(hasError: boolean, extraClassName = '') {
  return joinClassNames([
    'input',
    extraClassName,
    hasError && 'input-error'
  ])
}

function getRadioOptionClasses(isSelected: boolean) {
  return joinClassNames([
    'create-account-modal__radio-option',
    isSelected && 'create-account-modal__radio-option--active'
  ])
}

function getSegmentOptionClasses(isSelected: boolean) {
  return joinClassNames([
    'create-account-modal__segment-option',
    isSelected
      ? 'create-account-modal__segment-option--active'
      : 'create-account-modal__segment-option--idle'
  ])
}

// Computed: current preset mappings based on platform
const presetMappings = computed(() => getPresetMappingsByPlatform(form.platform))
const tempUnschedPresets = computed(() => buildAccountTempUnschedPresets(t))

const form = reactive<CreateAccountForm>(createDefaultCreateAccountForm())

// Helper to check if current type needs OAuth flow
const isOAuthFlow = computed(() => {
  return resolveCreateAccountOAuthFlow({
    accountCategory: accountCategory.value,
    antigravityAccountType: antigravityAccountType.value,
    platform: form.platform
  })
})

const isManualInputMethod = computed(() => {
  return oauthFlowRef.value?.inputMethod === 'manual'
})

const expiresAtInput = computed({
  get: () => formatDateTimeLocal(form.expires_at),
  set: (value: string) => {
    form.expires_at = parseDateTimeLocal(value)
  }
})

const canExchangeCode = computed(() => {
  const authCode = oauthFlowRef.value?.authCode || ''
  return Boolean(
    authCode.trim() &&
      currentOAuthState.value.sessionId &&
      !currentOAuthState.value.loading
  )
})

const loadAntigravityDefaultMappings = () =>
  fetchAntigravityDefaultMappings().then((mappings) => {
    antigravityModelMappings.value = [...mappings]
  })

const applyAntigravityModelDefaults = () => {
  antigravityModelRestrictionMode.value = 'mapping'
  antigravityWhitelistModels.value = []
  void loadAntigravityDefaultMappings()
}

const clearAntigravityModelState = () => {
  antigravityModelRestrictionMode.value = 'mapping'
  antigravityWhitelistModels.value = []
  antigravityModelMappings.value = []
}

const resetOAuthClientsState = (includeFlowState = false) => {
  oauth.resetState()
  openaiOAuth.resetState()
  geminiOAuth.resetState()
  antigravityOAuth.resetState()
  if (includeFlowState) {
    oauthFlowRef.value?.reset()
  }
}

// Watchers
watch(
  () => props.show,
  (newVal) => {
    if (newVal) {
      // Load TLS fingerprint profiles
      adminAPI.tlsFingerprintProfiles.list()
        .then(profiles => { tlsFingerprintProfiles.value = profiles.map(p => ({ id: p.id, name: p.name })) })
        .catch(() => { tlsFingerprintProfiles.value = [] })
      // Modal opened - fill related models
      allowedModels.value = [...getModelsByPlatform(form.platform)]
      if (form.platform === 'antigravity') {
        applyAntigravityModelDefaults()
      } else {
        clearAntigravityModelState()
      }
    } else {
      resetForm()
    }
  }
)

// Sync form.type based on accountCategory, addMethod, and platform-specific type
watch(
  [accountCategory, addMethod, antigravityAccountType],
  ([category, method, agType]) => {
    // Antigravity upstream 类型（实际创建为 apikey）
    if (form.platform === 'antigravity' && agType === 'upstream') {
      form.type = 'apikey'
      return
    }
    // Bedrock 类型
    if (form.platform === 'anthropic' && category === 'bedrock') {
      form.type = 'bedrock' as AccountType
      return
    }
    if (category === 'oauth-based') {
      form.type = method as AccountType // 'oauth' or 'setup-token'
    } else {
      form.type = 'apikey'
    }
  },
  { immediate: true }
)

// Reset platform-specific settings when platform changes
watch(
  () => form.platform,
  (newPlatform) => {
    // Reset base URL based on platform
    apiKeyBaseUrl.value = getDefaultBaseURL(newPlatform)
    // Clear model-related settings
    allowedModels.value = []
    modelMappings.value = []
    if (newPlatform === 'antigravity') {
      applyAntigravityModelDefaults()
      accountCategory.value = 'oauth-based'
      antigravityAccountType.value = 'oauth'
    } else {
      clearAntigravityModelState()
      allowOverages.value = false
    }
    resetBedrockCredentialState()
    // Reset Anthropic/Antigravity-specific settings when switching to other platforms
    if (newPlatform !== 'anthropic' && newPlatform !== 'antigravity') {
      interceptWarmupRequests.value = false
    }
    if (newPlatform !== 'openai') {
      resetOpenAICreateState()
    }
    if (newPlatform !== 'anthropic') {
      anthropicPassthroughEnabled.value = false
    }
    resetOAuthClientsState()
  }
)

// Gemini AI Studio OAuth availability (requires operator-configured OAuth client)
watch(
  [accountCategory, () => form.platform],
  ([category, platform]) => {
    if (platform === 'openai' && category !== 'oauth-based') {
      codexCLIOnlyEnabled.value = false
    }
    if (platform !== 'anthropic' || category !== 'apikey') {
      anthropicPassthroughEnabled.value = false
    }
  }
)

watch(
  [() => props.show, () => form.platform, accountCategory],
  ([show, platform, category]) => {
    void syncGeminiAIStudioOAuthAvailability(show, platform, category)
  },
  { immediate: true }
)

const handleSelectGeminiOAuthType = (oauthType: 'code_assist' | 'google_one' | 'ai_studio') => {
  if (oauthType === 'ai_studio' && !geminiAIStudioOAuthEnabled.value) {
    appStore.showError(t('admin.accounts.oauth.gemini.aiStudioNotConfigured'))
    return
  }
  geminiOAuthType.value = oauthType
}

// Auto-fill related models when switching to whitelist mode or changing platform
watch(
  [modelRestrictionMode, () => form.platform],
  ([newMode]) => {
    if (newMode === 'whitelist') {
      allowedModels.value = [...getModelsByPlatform(form.platform)]
    }
  }
)

watch(
  [antigravityModelRestrictionMode, () => form.platform],
  ([, platform]) => {
    if (platform !== 'antigravity') return
    // Antigravity 默认不做限制：白名单留空表示允许所有（包含未来新增模型）。
    // 如果需要快速填充常用模型，可在组件内点“填充相关模型”。
  }
)

// Model mapping helpers
const addModelMapping = () => {
  appendEmptyModelMapping(modelMappings.value)
}

const removeModelMapping = (index: number) => {
  removeModelMappingAt(modelMappings.value, index)
}

const addPresetMapping = (from: string, to: string) => {
  appendPresetModelMapping(modelMappings.value, from, to, (model) => {
    appStore.showInfo(t('admin.accounts.mappingExists', { model }))
  })
}

const addAntigravityModelMapping = () => {
  appendEmptyModelMapping(antigravityModelMappings.value)
}

const removeAntigravityModelMapping = (index: number) => {
  removeModelMappingAt(antigravityModelMappings.value, index)
}

const addAntigravityPresetMapping = (from: string, to: string) => {
  appendPresetModelMapping(antigravityModelMappings.value, from, to, (model) => {
    appStore.showInfo(t('admin.accounts.mappingExists', { model }))
  })
}

// Error code toggle helper
const toggleErrorCode = (code: number) => {
  const index = selectedErrorCodes.value.indexOf(code)
  if (index === -1) {
    if (!confirmCustomErrorCodeSelection(code, confirm, t)) {
      return
    }
    selectedErrorCodes.value.push(code)
  } else {
    selectedErrorCodes.value.splice(index, 1)
  }
}

// Add custom error code from input
const addCustomErrorCode = () => {
  const code = customErrorCodeInput.value
  if (code === null || code < 100 || code > 599) {
    appStore.showError(t('admin.accounts.invalidErrorCode'))
    return
  }
  if (selectedErrorCodes.value.includes(code)) {
    appStore.showInfo(t('admin.accounts.errorCodeExists'))
    return
  }
  if (!confirmCustomErrorCodeSelection(code, confirm, t)) {
    return
  }
  selectedErrorCodes.value.push(code)
  customErrorCodeInput.value = null
}

// Remove error code
const removeErrorCode = (code: number) => {
  const index = selectedErrorCodes.value.indexOf(code)
  if (index !== -1) {
    selectedErrorCodes.value.splice(index, 1)
  }
}

const addTempUnschedRule = (preset?: TempUnschedRuleForm) => {
  tempUnschedRules.value.push(createTempUnschedRule(preset))
}

const removeTempUnschedRule = (index: number) => {
  tempUnschedRules.value.splice(index, 1)
}

const moveTempUnschedRule = (index: number, direction: number) => {
  moveItemInPlace(tempUnschedRules.value, index, direction)
}

const clearMixedChannelDialog = () => {
  showMixedChannelWarning.value = false
  mixedChannelWarningDetails.value = null
  mixedChannelWarningRawMessage.value = ''
  mixedChannelWarningAction.value = null
}

const resetMixedChannelState = () => {
  antigravityMixedChannelConfirmed.value = false
  clearMixedChannelDialog()
}

const resolveCreateAccountErrorMessage = (error: any) =>
  error.response?.data?.message || error.response?.data?.detail || t('admin.accounts.failedToCreate')

const resolveOAuthAuthErrorMessage = (error: any) =>
  error.response?.data?.detail || t('admin.accounts.oauth.authFailed')

const getCurrentProxyConfig = () => (form.proxy_id ? { proxy_id: form.proxy_id } : {})

const buildValidatedTempUnschedPayload = () => {
  if (!tempUnschedEnabled.value) {
    return []
  }

  const payload = buildTempUnschedRules(tempUnschedRules.value)
  if (payload.length > 0) {
    return payload
  }

  appStore.showError(t('admin.accounts.tempUnschedulable.rulesInvalid'))
  return null
}

const openMixedChannelDialog = (opts: {
  response?: CheckMixedChannelResponse
  message?: string
  onConfirm: () => Promise<void>
}) => {
  mixedChannelWarningDetails.value = buildMixedChannelDetails(opts.response)
  mixedChannelWarningRawMessage.value =
    opts.message || opts.response?.message || t('admin.accounts.failedToCreate')
  mixedChannelWarningAction.value = opts.onConfirm
  showMixedChannelWarning.value = true
}

const withAntigravityConfirmFlag = (payload: CreateAccountRequest): CreateAccountRequest => {
  if (needsMixedChannelCheck(payload.platform) && antigravityMixedChannelConfirmed.value) {
    return {
      ...payload,
      confirm_mixed_channel_risk: true
    }
  }
  const cloned = { ...payload }
  delete cloned.confirm_mixed_channel_risk
  return cloned
}

const ensureAntigravityMixedChannelConfirmed = async (onConfirm: () => Promise<void>): Promise<boolean> => {
  if (!needsMixedChannelCheck(form.platform)) {
    return true
  }
  if (antigravityMixedChannelConfirmed.value) {
    return true
  }

  try {
    const result = await adminAPI.accounts.checkMixedChannelRisk({
      platform: form.platform,
      group_ids: form.group_ids
    })
    if (!result.has_risk) {
      return true
    }
    openMixedChannelDialog({
      response: result,
      onConfirm: async () => {
        antigravityMixedChannelConfirmed.value = true
        await onConfirm()
      }
    })
    return false
  } catch (error: any) {
    appStore.showError(resolveCreateAccountErrorMessage(error))
    return false
  }
}

const submitCreateAccount = async (payload: CreateAccountRequest) => {
  submitting.value = true
  try {
    await adminAPI.accounts.create(withAntigravityConfirmFlag(payload))
    notifyAccountCreated()
    finalizeCreatedAndClose()
  } catch (error: any) {
    if (error.response?.status === 409 && error.response?.data?.error === 'mixed_channel_warning' && needsMixedChannelCheck(form.platform)) {
      openMixedChannelDialog({
        message: error.response?.data?.message,
        onConfirm: async () => {
          antigravityMixedChannelConfirmed.value = true
          await submitCreateAccount(payload)
        }
      })
      return
    }
    appStore.showError(resolveCreateAccountErrorMessage(error))
  } finally {
    submitting.value = false
  }
}

const resetBedrockCredentialState = () => {
  bedrockAccessKeyId.value = ''
  bedrockSecretAccessKey.value = ''
  bedrockSessionToken.value = ''
  bedrockRegion.value = 'us-east-1'
  bedrockForceGlobal.value = false
  bedrockAuthMode.value = 'sigv4'
  bedrockApiKeyValue.value = ''
}

const resetOpenAICreateState = () => {
  openaiPassthroughEnabled.value = false
  openaiOAuthResponsesWebSocketV2Mode.value = OPENAI_WS_MODE_OFF
  openaiAPIKeyResponsesWebSocketV2Mode.value = OPENAI_WS_MODE_OFF
  codexCLIOnlyEnabled.value = false
}

const resetAnthropicQuotaControlState = () => {
  windowCostEnabled.value = false
  windowCostLimit.value = null
  windowCostStickyReserve.value = null
  sessionLimitEnabled.value = false
  maxSessions.value = null
  sessionIdleTimeout.value = null
  rpmLimitEnabled.value = false
  baseRpm.value = null
  rpmStrategy.value = 'tiered'
  rpmStickyBuffer.value = null
  userMsgQueueMode.value = ''
  tlsFingerprintEnabled.value = false
  tlsFingerprintProfileId.value = null
  sessionIdMaskingEnabled.value = false
  cacheTTLOverrideEnabled.value = false
  cacheTTLOverrideTarget.value = '5m'
  customBaseUrlEnabled.value = false
  customBaseUrl.value = ''
}

const resetAntigravityCreateState = () => {
  allowOverages.value = false
  antigravityAccountType.value = 'oauth'
  upstreamBaseUrl.value = ''
  upstreamApiKey.value = ''
  clearAntigravityModelState()
}

const resetGeminiSelectionState = () => {
  geminiOAuthType.value = 'code_assist'
  geminiTierGoogleOne.value = 'google_one_free'
  geminiTierGcp.value = 'gcp_standard'
  geminiTierAIStudio.value = 'aistudio_free'
}

const resetCustomErrorCodeState = () => {
  customErrorCodesEnabled.value = false
  selectedErrorCodes.value = []
  customErrorCodeInput.value = null
}

const resetQuotaResetState = () => {
  editQuotaLimit.value = null
  editQuotaDailyLimit.value = null
  editQuotaWeeklyLimit.value = null
  editDailyResetMode.value = null
  editDailyResetHour.value = null
  editWeeklyResetMode.value = null
  editWeeklyResetDay.value = null
  editWeeklyResetHour.value = null
  editResetTimezone.value = null
}

async function syncGeminiAIStudioOAuthAvailability(
  show: boolean,
  platform: AccountPlatform,
  category: typeof accountCategory.value
) {
  if (!show || platform !== 'gemini' || category !== 'oauth-based') {
    geminiAIStudioOAuthEnabled.value = false
    return
  }

  const capabilities = await geminiOAuth.getCapabilities()
  geminiAIStudioOAuthEnabled.value = !!capabilities?.ai_studio_oauth_enabled
  if (!geminiAIStudioOAuthEnabled.value && geminiOAuthType.value === 'ai_studio') {
    geminiOAuthType.value = 'code_assist'
  }
}

// Methods
const resetForm = () => {
  step.value = 1
  resetCreateAccountForm(form)
  accountCategory.value = 'oauth-based'
  addMethod.value = 'oauth'
  apiKeyBaseUrl.value = getDefaultBaseURL('anthropic')
  apiKeyValue.value = ''
  resetQuotaResetState()
  modelMappings.value = []
  modelRestrictionMode.value = 'whitelist'
  allowedModels.value = [...claudeModels] // Default fill related models
  poolModeEnabled.value = false
  poolModeRetryCount.value = DEFAULT_POOL_MODE_RETRY_COUNT
  resetCustomErrorCodeState()
  interceptWarmupRequests.value = false
  autoPauseOnExpired.value = true
  resetOpenAICreateState()
  anthropicPassthroughEnabled.value = false
  resetAnthropicQuotaControlState()
  resetAntigravityCreateState()
  tempUnschedEnabled.value = false
  tempUnschedRules.value = []
  resetGeminiSelectionState()
  resetBedrockCredentialState()
  resetOAuthClientsState(true)
  resetMixedChannelState()
}

const handleClose = () => {
  resetMixedChannelState()
  emit('close')
}

// Helper function to create account with mixed channel warning handling
const doCreateAccount = async (payload: CreateAccountRequest) => {
  const canContinue = await ensureAntigravityMixedChannelConfirmed(async () => {
    await submitCreateAccount(payload)
  })
  if (!canContinue) {
    return
  }
  await submitCreateAccount(payload)
}

// Handle mixed channel warning confirmation
const handleMixedChannelConfirm = async () => {
  const action = mixedChannelWarningAction.value
  if (!action) {
    clearMixedChannelDialog()
    return
  }
  clearMixedChannelDialog()
  submitting.value = true
  try {
    await action()
  } finally {
    submitting.value = false
  }
}

const handleMixedChannelCancel = () => {
  clearMixedChannelDialog()
}

const finalizeCreatedAndClose = () => {
  emit('created')
  handleClose()
}

const notifyAccountCreated = () => {
  appStore.showSuccess(t('admin.accounts.accountCreated'))
}

const buildCurrentCreateSharedPayload = () =>
  buildCreateAccountSharedPayload({
    autoPauseOnExpired: autoPauseOnExpired.value,
    concurrency: form.concurrency,
    expiresAt: form.expires_at,
    groupIds: form.group_ids,
    loadFactor: form.load_factor,
    notes: form.notes,
    priority: form.priority,
    proxyId: form.proxy_id,
    rateMultiplier: form.rate_multiplier
  })

const buildCurrentOpenAIExtra = (base?: Record<string, unknown>) =>
  buildCreateOpenAIExtra({
    accountCategory: accountCategory.value,
    base,
    codexCLIOnlyEnabled: codexCLIOnlyEnabled.value,
    openaiAPIKeyResponsesWebSocketV2Mode: openaiAPIKeyResponsesWebSocketV2Mode.value,
    openaiOAuthResponsesWebSocketV2Mode: openaiOAuthResponsesWebSocketV2Mode.value,
    openaiPassthroughEnabled: openaiPassthroughEnabled.value,
    platform: form.platform
  })

const buildCurrentAnthropicQuotaExtra = (baseExtra?: Record<string, unknown>) =>
  buildCreateAnthropicQuotaControlExtra({
    baseExtra,
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
    windowCostStickyReserve: windowCostStickyReserve.value
  })

const buildCurrentAntigravityExtra = () =>
  buildCreateAntigravityExtra({
    allowOverages: allowOverages.value,
    mixedScheduling: mixedScheduling.value
  })

const applyOpenAIModelRestrictionIfNeeded = (
  credentials: Record<string, unknown>,
  shouldApply: boolean
) => {
  if (!shouldApply) {
    return
  }

  assignBuiltModelMapping(
    credentials,
    modelRestrictionMode.value,
    allowedModels.value,
    modelMappings.value
  )
}

const handleBatchCreateOutcome = (options: {
  failedCount: number
  successCount: number
  errors: string[]
  setError: (message: string) => void
}) => {
  const outcome = resolveBatchCreateOutcome({
    failedCount: options.failedCount,
    successCount: options.successCount,
    t
  })

  if (outcome.type === 'success') {
    appStore.showSuccess(outcome.message)
  } else if (outcome.type === 'warning') {
    appStore.showWarning(outcome.message)
    options.setError(options.errors.join('\n'))
  } else {
    options.setError(options.errors.join('\n'))
    appStore.showError(outcome.message)
  }

  if (outcome.shouldEmitCreated) {
    emit('created')
  }
  if (outcome.shouldClose) {
    handleClose()
  }
}

const resolveBatchCreateUnexpectedError = (error: any) =>
  error?.response?.data?.detail || error?.message || 'Unknown error'

const createOAuthAccount = async (options: {
  commonPayload: ReturnType<typeof buildCurrentCreateSharedPayload>
  name: string
  platform: AccountPlatform
  type: AccountType
  credentials: Record<string, unknown>
  extra?: Record<string, unknown>
}) =>
  adminAPI.accounts.create(
    buildCreateAccountRequest({
      common: options.commonPayload,
      name: options.name,
      platform: options.platform,
      type: options.type,
      credentials: options.credentials,
      extra: options.extra
    })
  )

const createBatchCompletionHandler = (errorRef: { value: string }) => (result: {
  failedCount: number
  successCount: number
  errors: string[]
}) => {
  handleBatchCreateOutcome({
    failedCount: result.failedCount,
    successCount: result.successCount,
    errors: result.errors,
    setError: (message) => {
      errorRef.value = message
    }
  })
}

const resolveCurrentOAuthState = (
  fallbackState: string | undefined,
  errorRef: { value: string }
) =>
  resolveOAuthExchangeState({
    fallbackState,
    inputState: oauthFlowRef.value?.oauthState,
    onMissingState: (message) => {
      errorRef.value = message
      appStore.showError(message)
    },
    authFailedMessage: t('admin.accounts.oauth.authFailed')
  })

const ensureCreateAccountName = () => {
  if (form.name.trim()) {
    return true
  }
  appStore.showError(t('admin.accounts.pleaseEnterAccountName'))
  return false
}

const buildBedrockCreateCredentials = () => {
  const result = buildCreateBedrockCredentials({
    accessKeyId: bedrockAccessKeyId.value,
    allowedModels: allowedModels.value,
    apiKey: bedrockApiKeyValue.value,
    authMode: bedrockAuthMode.value,
    forceGlobal: bedrockForceGlobal.value,
    interceptWarmupRequests: interceptWarmupRequests.value,
    mode: modelRestrictionMode.value,
    modelMappings: modelMappings.value,
    poolModeEnabled: poolModeEnabled.value,
    poolModeRetryCount: poolModeRetryCount.value,
    region: bedrockRegion.value,
    secretAccessKey: bedrockSecretAccessKey.value,
    sessionToken: bedrockSessionToken.value
  })

  if (result.errorMessageKey) {
    appStore.showError(t(result.errorMessageKey))
    return null
  }

  return result.credentials || null
}

const buildAntigravityUpstreamCreateCredentials = () => {
  const result = buildCreateAntigravityUpstreamCredentials({
    apiKey: upstreamApiKey.value,
    baseUrl: upstreamBaseUrl.value,
    interceptWarmupRequests: interceptWarmupRequests.value,
    modelMappings: antigravityModelMappings.value
  })

  if (result.errorMessageKey) {
    appStore.showError(t(result.errorMessageKey))
    return null
  }

  return result.credentials || null
}

const buildApiKeyCreateCredentials = () => {
  const result = buildCreateApiKeyCredentials({
    allowedModels: allowedModels.value,
    apiKey: apiKeyValue.value,
    baseUrl: apiKeyBaseUrl.value,
    customErrorCodesEnabled: customErrorCodesEnabled.value,
    geminiTierId: geminiTierAIStudio.value,
    interceptWarmupRequests: interceptWarmupRequests.value,
    isOpenAIModelRestrictionDisabled: isOpenAIModelRestrictionDisabled.value,
    mode: modelRestrictionMode.value,
    modelMappings: modelMappings.value,
    platform: form.platform,
    poolModeEnabled: poolModeEnabled.value,
    poolModeRetryCount: poolModeRetryCount.value,
    selectedErrorCodes: selectedErrorCodes.value
  })

  if (result.errorMessageKey) {
    appStore.showError(t(result.errorMessageKey))
    return null
  }

  const credentials = result.credentials
  if (!credentials) {
    return null
  }

  if (
    !applyTempUnschedCredentialsState(credentials, {
      tempUnschedEnabled: tempUnschedEnabled.value,
      tempUnschedRules: tempUnschedRules.value,
      showError: appStore.showError,
      t
    })
  ) {
    return null
  }

  return credentials
}

const handleSubmit = async () => {
  // For OAuth-based type, handle OAuth flow (goes to step 2)
  if (isOAuthFlow.value) {
    if (!ensureCreateAccountName()) {
      return
    }
    const canContinue = await ensureAntigravityMixedChannelConfirmed(async () => {
      step.value = 2
    })
    if (!canContinue) {
      return
    }
    step.value = 2
    return
  }

  if (!ensureCreateAccountName()) {
    return
  }

  // For Bedrock type, create directly
  if (form.platform === 'anthropic' && accountCategory.value === 'bedrock') {
    const credentials = buildBedrockCreateCredentials()
    if (!credentials) {
      return
    }
    await createAccountAndFinish('anthropic', 'bedrock' as AccountType, credentials)
    return
  }

  // For Antigravity upstream type, create directly
  if (form.platform === 'antigravity' && antigravityAccountType.value === 'upstream') {
    const credentials = buildAntigravityUpstreamCreateCredentials()
    if (!credentials) {
      return
    }
    await createAccountAndFinish(
      form.platform,
      'apikey',
      credentials,
      buildCurrentAntigravityExtra()
    )
    return
  }

  // For apikey type, create directly
  const credentials = buildApiKeyCreateCredentials()
  if (!credentials) {
    return
  }

  form.credentials = credentials
  const extra = buildCreateAnthropicExtra({
    accountCategory: accountCategory.value,
    anthropicPassthroughEnabled: anthropicPassthroughEnabled.value,
    base: buildCurrentOpenAIExtra(),
    platform: form.platform
  })

  await doCreateAccount(
    buildCreateAccountRequest({
      common: buildCurrentCreateSharedPayload(),
      name: form.name,
      platform: form.platform,
      type: form.type,
      credentials,
      extra
    })
  )
}

const goBackToBasicInfo = () => {
  step.value = 1
  resetOAuthClientsState(true)
}

const runPlatformOAuthGenerateUrl = async () => {
  switch (form.platform) {
    case 'openai':
      await openaiOAuth.generateAuthUrl(form.proxy_id)
      return
    case 'gemini':
      await geminiOAuth.generateAuthUrl(
        form.proxy_id,
        oauthFlowRef.value?.projectId,
        geminiOAuthType.value,
        geminiSelectedTier.value
      )
      return
    case 'antigravity':
      await antigravityOAuth.generateAuthUrl(form.proxy_id)
      return
    default:
      await oauth.generateAuthUrl(addMethod.value, form.proxy_id)
  }
}

const handleGenerateUrl = async () => {
  await runPlatformOAuthGenerateUrl()
}

const runPlatformRefreshTokenValidation = (refreshToken: string) => {
  if (form.platform === 'openai') {
    handleOpenAIValidateRT(refreshToken)
    return
  }
  if (form.platform === 'antigravity') {
    handleAntigravityValidateRT(refreshToken)
  }
}

const handleValidateRefreshToken = (rt: string) => {
  runPlatformRefreshTokenValidation(rt)
}

const formatDateTimeLocal = formatDateTimeLocalInput
const parseDateTimeLocal = parseDateTimeLocalInput

// Create account and handle success/failure
const createAccountAndFinish = async (
  platform: AccountPlatform,
  type: AccountType,
  credentials: Record<string, unknown>,
  extra?: Record<string, unknown>
) => {
  if (
    !applyTempUnschedCredentialsState(credentials, {
      tempUnschedEnabled: tempUnschedEnabled.value,
      tempUnschedRules: tempUnschedRules.value,
      showError: appStore.showError,
      t
    })
  ) {
    return
  }
  const finalExtra =
    type === 'apikey' || type === 'bedrock'
      ? (() => {
          const quotaExtra = buildAccountQuotaExtra(extra, {
            dailyResetHour: editDailyResetHour.value,
            dailyResetMode: editDailyResetMode.value,
            quotaDailyLimit: editQuotaDailyLimit.value,
            quotaLimit: editQuotaLimit.value,
            quotaWeeklyLimit: editQuotaWeeklyLimit.value,
            resetTimezone: editResetTimezone.value,
            weeklyResetDay: editWeeklyResetDay.value,
            weeklyResetHour: editWeeklyResetHour.value,
            weeklyResetMode: editWeeklyResetMode.value
          })
          return Object.keys(quotaExtra).length > 0 ? quotaExtra : undefined
        })()
      : extra

  await doCreateAccount(
    buildCreateAccountRequest({
      common: buildCurrentCreateSharedPayload(),
      name: form.name,
      platform,
      type,
      credentials,
      extra: finalExtra
    })
  )
}

const buildAnthropicOAuthExtra = (tokenInfo: Record<string, unknown>) =>
  buildCurrentAnthropicQuotaExtra(oauth.buildExtraInfo(tokenInfo) || {})

const createAnthropicOAuthAccountFromTokenInfo = async (options: {
  commonPayload: ReturnType<typeof buildCurrentCreateSharedPayload>
  index?: number
  tempUnschedPayload?: ReturnType<typeof buildTempUnschedRules>
  tokenInfo: Record<string, unknown>
  total?: number
}) => {
  await adminAPI.accounts.create(
    buildCreateAnthropicOAuthAccountPayload({
      common: options.commonPayload,
      name: buildCreateBatchAccountName(form.name, options.index ?? 0, options.total ?? 1),
      platform: form.platform,
      type: addMethod.value as AccountType,
      interceptWarmupRequests: interceptWarmupRequests.value,
      tempUnschedPayload: options.tempUnschedPayload,
      tokenInfo: options.tokenInfo,
      extra: buildAnthropicOAuthExtra(options.tokenInfo)
    })
  )
}

// OpenAI OAuth 授权码兑换
const handleOpenAIExchange = async (authCode: string) => {
  const oauthClient = openaiOAuth
  if (!authCode.trim() || !oauthClient.sessionId.value) return

  await runOAuthExchangeFlow(
    oauthClient,
    async () => {
      const stateToUse = resolveCurrentOAuthState(oauthClient.oauthState.value, oauthClient.error)
      if (!stateToUse) {
        return
      }

      const tokenInfo = await oauthClient.exchangeAuthCode(
        authCode.trim(),
        oauthClient.sessionId.value,
        stateToUse,
        form.proxy_id
      )
      if (!tokenInfo) return

      const credentials = oauthClient.buildCredentials(tokenInfo)
      const oauthExtra = oauthClient.buildExtraInfo(tokenInfo) as Record<string, unknown> | undefined
      const extra = buildCurrentOpenAIExtra(oauthExtra)

      applyOpenAIModelRestrictionIfNeeded(
        credentials,
        form.platform === 'openai' && !isOpenAIModelRestrictionDisabled.value
      )

      // 应用临时不可调度配置
      if (
        !applyTempUnschedCredentialsState(credentials, {
          tempUnschedEnabled: tempUnschedEnabled.value,
          tempUnschedRules: tempUnschedRules.value,
          showError: appStore.showError,
          t
        })
      ) {
        return
      }

      const commonPayload = buildCurrentCreateSharedPayload()
      const target = buildCreateOpenAICompatOAuthTarget({
        baseName: form.name,
        credentials,
        extra,
        platform: 'openai'
      })

      await createOAuthAccount({
        commonPayload,
        ...target
      })
      notifyAccountCreated()

      finalizeCreatedAndClose()
    },
    resolveOAuthAuthErrorMessage,
    appStore.showError
  )
}

// OpenAI 手动 RT 批量验证和创建
// OpenAI Mobile RT 使用的 client_id
const OPENAI_MOBILE_RT_CLIENT_ID = 'app_LlGpXReQgckcGGUo2JrYvtJK'

// OpenAI RT 批量验证和创建
const handleOpenAIBatchRT = async (refreshTokenInput: string, clientId?: string) => {
  const oauthClient = openaiOAuth
  const commonPayload = buildCurrentCreateSharedPayload()
  await runBatchCreateFlow({
    rawInput: refreshTokenInput,
    emptyInputMessage: t('admin.accounts.oauth.openai.pleaseEnterRefreshToken'),
    loadingRef: oauthClient.loading,
    errorRef: oauthClient.error,
    onComplete: createBatchCompletionHandler(oauthClient.error),
    processEntry: async (refreshToken, index, refreshTokens) => {
      const tokenInfo = await oauthClient.validateRefreshToken(
        refreshToken,
        form.proxy_id,
        clientId
      )
      if (!tokenInfo) {
        return consumeValidationFailureMessage(oauthClient.error)
      }

      const credentials = oauthClient.buildCredentials(tokenInfo)
      if (clientId) {
        credentials.client_id = clientId
      }
      const oauthExtra = oauthClient.buildExtraInfo(tokenInfo) as Record<string, unknown> | undefined
      const extra = buildCurrentOpenAIExtra(oauthExtra)

      applyOpenAIModelRestrictionIfNeeded(
        credentials,
        form.platform === 'openai' && !isOpenAIModelRestrictionDisabled.value
      )

      const target = buildCreateOpenAICompatOAuthTarget({
        baseName: form.name,
        credentials,
        extra,
        fallbackBaseName: tokenInfo.email || 'OpenAI OAuth Account',
        index,
        platform: 'openai',
        total: refreshTokens.length
      })

      await createOAuthAccount({
        commonPayload,
        ...target
      })

      return null
    },
    resolveUnexpectedError: resolveBatchCreateUnexpectedError
  })
}

// 手动输入 RT（Codex CLI client_id，默认）
const handleOpenAIValidateRT = (rt: string) => handleOpenAIBatchRT(rt)

// 手动输入 Mobile RT
const handleOpenAIValidateMobileRT = (rt: string) => handleOpenAIBatchRT(rt, OPENAI_MOBILE_RT_CLIENT_ID)

// Antigravity 手动 RT 批量验证和创建
const handleAntigravityValidateRT = async (refreshTokenInput: string) => {
  const commonPayload = buildCurrentCreateSharedPayload()
  await runBatchCreateFlow({
    rawInput: refreshTokenInput,
    emptyInputMessage: t('admin.accounts.oauth.antigravity.pleaseEnterRefreshToken'),
    loadingRef: antigravityOAuth.loading,
    errorRef: antigravityOAuth.error,
    onComplete: createBatchCompletionHandler(antigravityOAuth.error),
    processEntry: async (refreshToken, index, refreshTokens) => {
      const tokenInfo = await antigravityOAuth.validateRefreshToken(refreshToken, form.proxy_id)
      if (!tokenInfo) {
        return consumeValidationFailureMessage(antigravityOAuth.error)
      }

      const credentials = antigravityOAuth.buildCredentials(tokenInfo)
      const createPayload = withAntigravityConfirmFlag(
        buildCreateAccountRequest({
          common: commonPayload,
          name: buildCreateBatchAccountName(form.name, index, refreshTokens.length),
          platform: 'antigravity',
          type: 'oauth',
          credentials,
          extra: buildCurrentAntigravityExtra()
        })
      )
      await adminAPI.accounts.create(createPayload)
      return null
    },
    resolveUnexpectedError: resolveBatchCreateUnexpectedError
  })
}

// Gemini OAuth 授权码兑换
const handleGeminiExchange = async (authCode: string) => {
  if (!authCode.trim() || !geminiOAuth.sessionId.value) return

  await runOAuthExchangeFlow(
    geminiOAuth,
    async () => {
      const stateToUse = resolveCurrentOAuthState(geminiOAuth.state.value, geminiOAuth.error)
      if (!stateToUse) {
        return
      }

      const tokenInfo = await geminiOAuth.exchangeAuthCode({
        code: authCode.trim(),
        sessionId: geminiOAuth.sessionId.value,
        state: stateToUse,
        proxyId: form.proxy_id,
        oauthType: geminiOAuthType.value,
        tierId: geminiSelectedTier.value
      })
      if (!tokenInfo) return

      const credentials = geminiOAuth.buildCredentials(tokenInfo)
      const extra = geminiOAuth.buildExtraInfo(tokenInfo)
      await createAccountAndFinish('gemini', 'oauth', credentials, extra)
    },
    resolveOAuthAuthErrorMessage,
    appStore.showError
  )
}

// Antigravity OAuth 授权码兑换
const handleAntigravityExchange = async (authCode: string) => {
  if (!authCode.trim() || !antigravityOAuth.sessionId.value) return

  await runOAuthExchangeFlow(
    antigravityOAuth,
    async () => {
      const stateToUse = resolveCurrentOAuthState(antigravityOAuth.state.value, antigravityOAuth.error)
      if (!stateToUse) {
        return
      }

      const tokenInfo = await antigravityOAuth.exchangeAuthCode({
        code: authCode.trim(),
        sessionId: antigravityOAuth.sessionId.value,
        state: stateToUse,
        proxyId: form.proxy_id
      })
      if (!tokenInfo) return

      const credentials = buildCreateAntigravityOAuthCredentials({
        interceptWarmupRequests: interceptWarmupRequests.value,
        modelMappings: antigravityModelMappings.value,
        tokenInfo: antigravityOAuth.buildCredentials(tokenInfo)
      })
      const extra = buildCurrentAntigravityExtra()
      await createAccountAndFinish('antigravity', 'oauth', credentials, extra)
    },
    resolveOAuthAuthErrorMessage,
    appStore.showError
  )
}

// Anthropic OAuth 授权码兑换
const handleAnthropicExchange = async (authCode: string) => {
  if (!authCode.trim() || !oauth.sessionId.value) return

  await runOAuthExchangeFlow(
    oauth,
    async () => {
      const tokenInfo = await adminAPI.accounts.exchangeCode(
        resolveAnthropicExchangeEndpoint(addMethod.value as 'oauth' | 'setup-token', 'code'),
        {
          session_id: oauth.sessionId.value,
          code: authCode.trim(),
          ...getCurrentProxyConfig()
        }
      )

      await doCreateAccount(
        buildCreateAnthropicOAuthAccountPayload({
          common: buildCurrentCreateSharedPayload(),
          name: form.name,
          platform: form.platform,
          type: addMethod.value as AccountType,
          interceptWarmupRequests: interceptWarmupRequests.value,
          tokenInfo,
          extra: buildAnthropicOAuthExtra(tokenInfo)
        })
      )
    },
    resolveOAuthAuthErrorMessage,
    appStore.showError
  )
}

// 主入口：根据平台路由到对应处理函数
const runPlatformOAuthExchange = async (authCode: string) => {
  switch (form.platform) {
    case 'openai':
      return handleOpenAIExchange(authCode)
    case 'gemini':
      return handleGeminiExchange(authCode)
    case 'antigravity':
      return handleAntigravityExchange(authCode)
    default:
      return handleAnthropicExchange(authCode)
  }
}

const handleExchangeCode = async () => {
  await runPlatformOAuthExchange(oauthFlowRef.value?.authCode || '')
}

const handleCookieAuth = async (sessionKey: string) => {
  try {
    const keys = oauth.parseSessionKeys(sessionKey)

    if (keys.length === 0) {
      oauth.error.value = t('admin.accounts.oauth.pleaseEnterSessionKey')
      return
    }

    const tempUnschedPayload = buildValidatedTempUnschedPayload()
    if (tempUnschedPayload == null) {
      return
    }

    const commonPayload = buildCurrentCreateSharedPayload()

    await runOAuthExchangeFlow(
      oauth,
      async () => {
        await runBatchCreateFlow({
          rawInput: keys.join('\n'),
          emptyInputMessage: t('admin.accounts.oauth.pleaseEnterSessionKey'),
          loadingRef: oauth.loading,
          errorRef: oauth.error,
          onComplete: ({ successCount, failedCount, errors }) => {
            if (successCount > 0) {
              appStore.showSuccess(t('admin.accounts.oauth.successCreated', { count: successCount }))
              emit('created')
              if (failedCount === 0) {
                handleClose()
              }
            }

            if (failedCount > 0) {
              oauth.error.value = errors.join('\n')
            }
          },
          processEntry: async (key, index, allKeys) => {
            try {
              const tokenInfo = await adminAPI.accounts.exchangeCode(
                resolveAnthropicExchangeEndpoint(addMethod.value as 'oauth' | 'setup-token', 'cookie'),
                {
                  session_id: '',
                  code: key,
                  ...getCurrentProxyConfig()
                }
              )

              await createAnthropicOAuthAccountFromTokenInfo({
                commonPayload,
                index,
                tempUnschedPayload,
                tokenInfo,
                total: allKeys.length
              })
              return null
            } catch (error: any) {
              return t('admin.accounts.oauth.keyAuthFailed', {
                index: index + 1,
                error: resolveOAuthAuthErrorMessage(error)
              })
            }
          }
        })
      },
      resolveOAuthAuthErrorMessage,
      appStore.showError
    )
  } catch (error: any) {
    oauth.error.value = error.response?.data?.detail || t('admin.accounts.oauth.cookieAuthFailed')
  }
}
</script>

<style scoped>
.create-account-modal__step-node--active {
  background: var(--theme-accent);
  color: var(--theme-filled-text);
}

.create-account-modal__step-node--idle {
  background: color-mix(in srgb, var(--theme-page-border) 84%, var(--theme-surface));
  color: var(--theme-page-muted);
}

.create-account-modal__step-label {
  color: var(--theme-page-text);
}

.create-account-modal__step-connector {
  background: color-mix(in srgb, var(--theme-page-border) 78%, transparent);
}

.create-account-modal__platform-button {
  color: var(--theme-page-muted);
}

.create-account-modal__platform-button-control {
  border-radius: calc(var(--theme-button-radius) - 2px);
  padding: 0.625rem 1rem;
}

.create-account-modal__platform-button--idle:hover {
  color: var(--theme-page-text);
}

.create-account-modal__platform-button--active {
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}

.create-account-modal__platform-button--anthropic {
  color: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 84%, var(--theme-page-text));
}

.create-account-modal__platform-button--openai {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.create-account-modal__platform-button--gemini {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.create-account-modal__platform-button--antigravity {
  color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 84%, var(--theme-page-text));
}

.create-account-modal__choice-card {
  border-radius: calc(var(--theme-button-radius) + 2px);
  padding: 0.75rem;
  border-color: color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 82%, var(--theme-surface));
}

.create-account-modal__choice-card:hover {
  border-color: color-mix(in srgb, var(--theme-page-border) 92%, var(--theme-accent));
}

.create-account-modal__choice-card--disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.create-account-modal__choice-card--idle {
  color: var(--theme-page-text);
}

.create-account-modal__choice-icon--idle {
  background: color-mix(in srgb, var(--theme-page-border) 86%, var(--theme-surface));
  color: var(--theme-page-muted);
}

.create-account-modal__choice-icon-control {
  border-radius: calc(var(--theme-button-radius) + 1px);
}

.create-account-modal__choice-title {
  color: var(--theme-page-text);
}

.create-account-modal__choice-description {
  color: var(--theme-page-muted);
}

.create-account-modal__help-button {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  border-radius: var(--theme-button-radius);
  padding: calc(var(--theme-button-padding-y) * 0.4) calc(var(--theme-button-padding-x) * 0.4);
  font-size: 0.75rem;
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 82%, var(--theme-page-text));
}

.create-account-modal__help-button:hover {
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface));
}

.create-account-modal__error-text {
  color: rgb(var(--theme-danger-rgb));
  font-size: 0.75rem;
}

.create-account-modal__link {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.create-account-modal__link:hover {
  text-decoration: underline;
}

.create-account-modal__inline-toggle {
  color: var(--theme-page-muted);
}

.create-account-modal__inline-toggle:hover {
  color: var(--theme-page-text);
}

.create-account-modal__notice {
  border-radius: var(--theme-auth-feedback-radius);
  border-color: color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.create-account-modal__notice-inline {
  padding: 0.5rem 0.75rem;
}

.create-account-modal__notice-block {
  padding: var(--theme-auth-callback-feedback-padding);
}

.create-account-modal__notice-tooltip {
  width: 20rem;
  padding: 0.5rem 0.75rem;
  border-radius: var(--theme-button-radius);
}

.form-section {
  border-top: 1px solid color-mix(in srgb, var(--theme-page-border) 76%, transparent);
  padding-top: 1rem;
}

.create-account-modal__config-card {
  border-radius: var(--theme-surface-radius);
  padding: var(--theme-markdown-block-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 90%, var(--theme-surface));
}

.create-account-modal__tone-tag {
  display: inline-flex;
  align-items: center;
}

.create-account-modal__tone-tag-control {
  border-radius: var(--theme-button-radius);
  padding: 0.125rem 0.5rem;
}

.create-account-modal__tone-tag-anchor {
  margin-left: auto;
  flex-shrink: 0;
  font-size: 0.75rem;
}

.create-account-modal__secondary-button {
  padding-inline: calc(var(--theme-button-padding-x) * 0.75);
}

.create-account-modal__choice-card--rose,
.create-account-modal__choice-icon--rose,
.create-account-modal__tone-tag--rose,
.create-account-modal__notice--rose {
  --create-account-tone-rgb: var(--theme-brand-rose-rgb);
}

.create-account-modal__choice-card--orange,
.create-account-modal__choice-icon--orange,
.create-account-modal__tone-tag--orange,
.create-account-modal__notice--orange {
  --create-account-tone-rgb: var(--theme-brand-orange-rgb);
}

.create-account-modal__choice-card--purple,
.create-account-modal__choice-icon--purple,
.create-account-modal__tone-tag--purple,
.create-account-modal__notice--purple {
  --create-account-tone-rgb: var(--theme-brand-purple-rgb);
}

.create-account-modal__choice-card--amber,
.create-account-modal__choice-icon--amber,
.create-account-modal__tone-tag--amber,
.create-account-modal__notice--amber {
  --create-account-tone-rgb: var(--theme-warning-rgb);
}

.create-account-modal__choice-card--green,
.create-account-modal__choice-icon--green,
.create-account-modal__tone-tag--green,
.create-account-modal__notice--green {
  --create-account-tone-rgb: var(--theme-success-rgb);
}

.create-account-modal__choice-card--blue,
.create-account-modal__choice-icon--blue,
.create-account-modal__tone-tag--blue,
.create-account-modal__notice--blue {
  --create-account-tone-rgb: var(--theme-info-rgb);
}

.create-account-modal__choice-card--emerald,
.create-account-modal__choice-icon--emerald,
.create-account-modal__tone-tag--emerald,
.create-account-modal__notice--emerald {
  --create-account-tone-rgb: var(--theme-success-rgb);
}

.create-account-modal__choice-card--rose,
.create-account-modal__choice-card--orange,
.create-account-modal__choice-card--purple,
.create-account-modal__choice-card--amber,
.create-account-modal__choice-card--green,
.create-account-modal__choice-card--blue,
.create-account-modal__choice-card--emerald {
  border-color: rgb(var(--create-account-tone-rgb));
  background: color-mix(in srgb, rgb(var(--create-account-tone-rgb)) 12%, var(--theme-surface));
}

.create-account-modal__choice-icon--rose,
.create-account-modal__choice-icon--orange,
.create-account-modal__choice-icon--purple,
.create-account-modal__choice-icon--amber,
.create-account-modal__choice-icon--green,
.create-account-modal__choice-icon--blue,
.create-account-modal__choice-icon--emerald {
  background: rgb(var(--create-account-tone-rgb));
  color: var(--theme-filled-text);
}

.create-account-modal__tone-tag--rose,
.create-account-modal__tone-tag--orange,
.create-account-modal__tone-tag--purple,
.create-account-modal__tone-tag--amber,
.create-account-modal__tone-tag--green,
.create-account-modal__tone-tag--blue,
.create-account-modal__tone-tag--emerald {
  background: color-mix(in srgb, rgb(var(--create-account-tone-rgb)) 16%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--create-account-tone-rgb)) 88%, var(--theme-page-text));
}

.create-account-modal__notice--rose,
.create-account-modal__notice--orange,
.create-account-modal__notice--purple,
.create-account-modal__notice--amber,
.create-account-modal__notice--green,
.create-account-modal__notice--blue,
.create-account-modal__notice--emerald {
  background: color-mix(in srgb, rgb(var(--create-account-tone-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--create-account-tone-rgb)) 84%, var(--theme-page-text));
}

.create-account-modal__radio-option,
.create-account-modal__checkbox {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}

.create-account-modal__radio-option {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  border-radius: calc(var(--theme-button-radius) + 2px);
  background: color-mix(in srgb, var(--theme-surface-soft) 84%, var(--theme-surface));
  padding: 0.55rem 0.8rem;
  transition:
    border-color 0.18s ease,
    background-color 0.18s ease,
    color 0.18s ease;
}

.create-account-modal__radio-option--active {
  border-color: color-mix(in srgb, var(--theme-accent) 64%, var(--theme-card-border));
  background: color-mix(in srgb, var(--theme-accent-soft) 78%, var(--theme-surface));
}

.create-account-modal__radio-input,
.create-account-modal__checkbox-input {
  accent-color: var(--theme-accent);
}

.create-account-modal__checkbox-input {
  border: 1px solid var(--theme-input-border);
  border-radius: 0.375rem;
}

.create-account-modal__mode-toggle--idle,
.create-account-modal__status-chip--idle {
  background: color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-surface));
  color: var(--theme-page-muted);
}

.create-account-modal__mode-toggle-control {
  border-radius: var(--theme-button-radius);
  padding: 0.5rem 1rem;
}

.create-account-modal__status-chip-control {
  border-radius: var(--theme-button-radius);
  padding: 0.375rem 0.75rem;
}

.create-account-modal__status-chip-action {
  border-radius: var(--theme-button-radius);
  padding: 0.5rem;
}

.create-account-modal__status-chip-inline {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  border-radius: 999px;
  padding: 0.125rem 0.625rem;
}

.create-account-modal__mode-toggle--idle:hover,
.create-account-modal__status-chip--idle:hover {
  background: color-mix(in srgb, var(--theme-page-border) 66%, var(--theme-surface));
  color: var(--theme-page-text);
}

.create-account-modal__mode-toggle--accent {
  background: color-mix(in srgb, var(--theme-accent) 14%, var(--theme-surface));
  color: color-mix(in srgb, var(--theme-accent) 90%, var(--theme-page-text));
}

.create-account-modal__mode-toggle--purple,
.create-account-modal__status-chip--purple {
  background: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 14%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 88%, var(--theme-page-text));
}

.create-account-modal__mode-toggle--danger,
.create-account-modal__status-chip--danger {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 12%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 88%, var(--theme-page-text));
}

.create-account-modal__switch {
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--theme-page-border) 40%, transparent);
}

.create-account-modal__switch:focus-visible {
  box-shadow:
    0 0 0 2px color-mix(in srgb, var(--theme-accent) 22%, transparent),
    0 0 0 4px color-mix(in srgb, var(--theme-accent) 12%, transparent);
}

.create-account-modal__switch--enabled {
  background: var(--theme-accent);
}

.create-account-modal__switch--disabled {
  background: color-mix(in srgb, var(--theme-page-border) 76%, var(--theme-surface));
}

.create-account-modal__switch-thumb {
  background: var(--theme-surface-contrast);
}

.create-account-modal__segment-option {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 82%, transparent);
  border-radius: calc(var(--theme-button-radius) - 2px);
  font-size: 0.875rem;
  padding: 0.375rem 0.75rem;
  transition:
    border-color 0.18s ease,
    background-color 0.18s ease,
    color 0.18s ease;
}

.create-account-modal__segment-option--active {
  border-color: var(--theme-accent);
  background: var(--theme-accent);
  color: var(--theme-filled-text);
}

.create-account-modal__segment-option--idle {
  background: var(--theme-surface);
  color: var(--theme-page-text);
}

.create-account-modal__segment-option--idle:hover {
  background: color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-surface));
}

.create-account-modal__tag-button {
  border-radius: calc(var(--theme-button-radius) + 2px);
  background: color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-surface));
  color: var(--theme-page-muted);
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.4rem 0.75rem;
  transition:
    background-color 0.18s ease,
    color 0.18s ease;
}

.create-account-modal__tag-button:hover,
.create-account-modal__tag-button:focus-visible {
  background: color-mix(in srgb, var(--theme-page-border) 72%, var(--theme-surface));
  color: var(--theme-page-text);
  outline: none;
}

.create-account-modal__rule-card {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 80%, transparent);
  border-radius: calc(var(--theme-button-radius) + 2px);
  background: var(--theme-surface);
  padding: 0.75rem;
}

.create-account-modal__rule-index {
  color: var(--theme-page-muted);
  font-size: 0.75rem;
  font-weight: 600;
}

.create-account-modal__icon-button {
  border-radius: calc(var(--theme-button-radius) - 4px);
  color: var(--theme-page-muted);
  padding: 0.25rem;
  transition: background-color 0.18s ease, color 0.18s ease;
}

.create-account-modal__icon-button:hover,
.create-account-modal__icon-button:focus-visible {
  background: color-mix(in srgb, var(--theme-button-ghost-hover-bg) 90%, transparent);
  color: var(--theme-page-text);
  outline: none;
}

.create-account-modal__icon-button:disabled {
  cursor: not-allowed;
  opacity: 0.4;
}

.create-account-modal__icon-button--danger {
  color: rgb(var(--theme-danger-rgb));
}

.create-account-modal__icon-button--danger:hover,
.create-account-modal__icon-button--danger:focus-visible {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, transparent);
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 88%, var(--theme-page-text));
}

.create-account-modal__dashed-action {
  border: 2px dashed color-mix(in srgb, var(--theme-card-border) 90%, transparent);
  border-radius: calc(var(--theme-button-radius) + 2px);
  color: var(--theme-page-muted);
  font-size: 0.875rem;
  padding: 0.625rem 1rem;
  transition:
    border-color 0.18s ease,
    color 0.18s ease,
    background-color 0.18s ease;
}

.create-account-modal__dashed-action:hover,
.create-account-modal__dashed-action:focus-visible {
  border-color: color-mix(in srgb, var(--theme-accent) 32%, var(--theme-card-border));
  background: color-mix(in srgb, var(--theme-accent-soft) 46%, var(--theme-surface));
  color: var(--theme-page-text);
  outline: none;
}

.create-account-modal__extra-options {
  border-top: 1px solid color-mix(in srgb, var(--theme-page-border) 76%, transparent);
  padding-top: 1rem;
}

.create-account-modal__tooltip-trigger {
  display: inline-flex;
  height: 1rem;
  width: 1rem;
  cursor: help;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: color-mix(in srgb, var(--theme-page-border) 82%, var(--theme-surface));
  color: var(--theme-page-muted);
  font-size: 0.75rem;
  transition: background-color 0.18s ease, color 0.18s ease;
}

.create-account-modal__tooltip-trigger:hover {
  background: color-mix(in srgb, var(--theme-page-border) 92%, var(--theme-surface));
  color: var(--theme-page-text);
}

.create-account-modal__tooltip-panel {
  pointer-events: none;
  position: absolute;
  left: 0;
  top: 100%;
  z-index: 100;
  width: 18rem;
  margin-top: 0.375rem;
  border-radius: calc(var(--theme-button-radius) + 2px);
  background: var(--theme-surface-contrast);
  color: var(--theme-surface-contrast-text);
  font-size: 0.75rem;
  opacity: 0;
  padding: 0.625rem 0.75rem;
  transition: opacity 0.18s ease;
}

.group:hover .create-account-modal__tooltip-panel {
  opacity: 1;
}

.create-account-modal__tooltip-arrow {
  position: absolute;
  bottom: 100%;
  left: 0.75rem;
  border: 4px solid transparent;
  border-bottom-color: var(--theme-surface-contrast);
}

.create-account-modal__dialog-title,
.create-account-modal__dialog-subtitle {
  color: var(--theme-page-text);
}

.create-account-modal__dialog-title {
  margin-bottom: 0.75rem;
  font-size: 0.875rem;
  font-weight: 700;
}

.create-account-modal__dialog-subtitle {
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  font-weight: 600;
}

.create-account-modal__dialog-list {
  list-style: disc;
  list-style-position: inside;
  color: var(--theme-page-muted);
  font-size: 0.875rem;
}

.create-account-modal__table-head {
  background: color-mix(in srgb, var(--theme-surface-soft) 92%, var(--theme-surface));
}

.create-account-modal__table-body {
  border-top: 1px solid color-mix(in srgb, var(--theme-page-border) 74%, transparent);
}

.create-account-modal__table-heading {
  padding: var(--theme-table-cell-padding-y) var(--theme-table-cell-padding-x);
  text-align: left;
  font-weight: 500;
  color: var(--theme-page-text);
}

.create-account-modal__table-primary {
  padding: var(--theme-table-cell-padding-y) var(--theme-table-cell-padding-x);
  color: var(--theme-page-text);
}

.create-account-modal__table-secondary {
  padding: var(--theme-table-cell-padding-y) var(--theme-table-cell-padding-x);
  color: var(--theme-page-muted);
}
</style>
