<template>
  <div class="oauth-flow">
    <div class="flex items-start gap-4">
      <div class="oauth-flow__hero-icon flex h-10 w-10 flex-shrink-0 items-center justify-center">
        <Icon name="link" size="md" class="oauth-flow__hero-icon-symbol" />
      </div>
      <div class="flex-1">
        <h4 class="oauth-flow__title mb-3 font-semibold">{{ oauthText.title }}</h4>

        <div v-if="showMethodSelection" class="mb-4">
          <label class="oauth-flow__section-label mb-2 block text-sm font-medium">
            {{ methodLabel }}
          </label>
          <div class="flex flex-wrap gap-4">
            <label class="oauth-flow__radio-option flex cursor-pointer items-center gap-2">
              <input v-model="inputMethod" type="radio" value="manual" class="oauth-flow__radio" />
              <span class="oauth-flow__radio-label text-sm">
                {{ t('admin.accounts.oauth.manualAuth') }}
              </span>
            </label>
            <label
              v-if="showCookieOption"
              class="oauth-flow__radio-option flex cursor-pointer items-center gap-2"
            >
              <input v-model="inputMethod" type="radio" value="cookie" class="oauth-flow__radio" />
              <span class="oauth-flow__radio-label text-sm">
                {{ t('admin.accounts.oauth.cookieAutoAuth') }}
              </span>
            </label>
            <label
              v-if="showRefreshTokenOption"
              class="oauth-flow__radio-option flex cursor-pointer items-center gap-2"
            >
              <input
                v-model="inputMethod"
                type="radio"
                value="refresh_token"
                class="oauth-flow__radio"
              />
              <span class="oauth-flow__radio-label text-sm">
                {{ t(getOAuthKey('refreshTokenAuth')) }}
              </span>
            </label>
            <label
              v-if="showMobileRefreshTokenOption"
              class="oauth-flow__radio-option flex cursor-pointer items-center gap-2"
            >
              <input
                v-model="inputMethod"
                type="radio"
                value="mobile_refresh_token"
                class="oauth-flow__radio"
              />
              <span class="oauth-flow__radio-label text-sm">
                {{ t('admin.accounts.oauth.openai.mobileRefreshTokenAuth', '手动输入 Mobile RT') }}
              </span>
            </label>
            <label
              v-if="showSessionTokenOption"
              class="oauth-flow__radio-option flex cursor-pointer items-center gap-2"
            >
              <input
                v-model="inputMethod"
                type="radio"
                value="session_token"
                class="oauth-flow__radio"
              />
              <span class="oauth-flow__radio-label text-sm">
                {{ t(getOAuthKey('sessionTokenAuth')) }}
              </span>
            </label>
            <label
              v-if="showAccessTokenOption"
              class="oauth-flow__radio-option flex cursor-pointer items-center gap-2"
            >
              <input
                v-model="inputMethod"
                type="radio"
                value="access_token"
                class="oauth-flow__radio"
              />
              <span class="oauth-flow__radio-label text-sm">
                {{ t('admin.accounts.oauth.openai.accessTokenAuth', '手动输入 AT') }}
              </span>
            </label>
          </div>
        </div>

        <div
          v-if="inputMethod === 'refresh_token' || inputMethod === 'mobile_refresh_token'"
          class="space-y-4"
        >
          <div class="oauth-flow__panel">
            <p class="oauth-flow__panel-description mb-3 text-sm">
              {{ t(getOAuthKey('refreshTokenDesc')) }}
            </p>

            <div class="mb-4">
              <label class="oauth-flow__field-label mb-2 flex items-center gap-2 text-sm font-semibold">
                <Icon name="key" size="sm" class="oauth-flow__field-icon" />
                Refresh Token
                <span v-if="parsedRefreshTokenCount > 1" class="theme-chip theme-chip--compact theme-chip--info">
                  {{ t('admin.accounts.oauth.keysCount', { count: parsedRefreshTokenCount }) }}
                </span>
              </label>
              <textarea
                v-model="refreshTokenInput"
                rows="3"
                class="input w-full resize-y font-mono text-sm"
                :placeholder="t(getOAuthKey('refreshTokenPlaceholder'))"
              ></textarea>
              <p v-if="parsedRefreshTokenCount > 1" class="oauth-flow__note mt-1 text-xs">
                {{ t('admin.accounts.oauth.batchCreateAccounts', { count: parsedRefreshTokenCount }) }}
              </p>
            </div>

            <div v-if="error" class="oauth-flow__alert oauth-flow__alert--danger mb-4">
              <p class="oauth-flow__alert-text whitespace-pre-line text-sm">
                {{ error }}
              </p>
            </div>

            <button
              type="button"
              class="btn btn-primary w-full"
              :disabled="loading || !refreshTokenInput.trim()"
              @click="handleValidateRefreshToken"
            >
              <svg
                v-if="loading"
                class="-ml-1 mr-2 h-4 w-4 animate-spin"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
              <Icon v-else name="sparkles" size="sm" class="mr-2" />
              {{ loading ? t(getOAuthKey('validating')) : t(getOAuthKey('validateAndCreate')) }}
            </button>
          </div>
        </div>

        <div v-if="inputMethod === 'session_token'" class="space-y-4">
          <div class="oauth-flow__panel">
            <p class="oauth-flow__panel-description mb-3 text-sm">
              {{ t(getOAuthKey('sessionTokenDesc')) }}
            </p>

            <div class="mb-4">
              <label class="oauth-flow__field-label mb-2 flex items-center gap-2 text-sm font-semibold">
                <Icon name="key" size="sm" class="oauth-flow__field-icon" />
                {{ t(getOAuthKey('sessionTokenRawLabel')) }}
                <span
                  v-if="parsedSoraTokenState.sessionTokenCount > 1"
                  class="theme-chip theme-chip--compact theme-chip--info"
                >
                  {{ t('admin.accounts.oauth.keysCount', { count: parsedSoraTokenState.sessionTokenCount }) }}
                </span>
              </label>
              <textarea
                v-model="sessionTokenInput"
                rows="3"
                class="input w-full resize-y font-mono text-sm"
                :placeholder="t(getOAuthKey('sessionTokenRawPlaceholder'))"
              ></textarea>
              <p class="oauth-flow__note mt-1 text-xs">
                {{ t(getOAuthKey('sessionTokenRawHint')) }}
              </p>
              <div class="mt-2 flex flex-wrap items-center gap-2">
                <button
                  type="button"
                  class="oauth-flow__secondary-action btn btn-secondary text-xs"
                  @click="handleOpenSoraSessionUrl"
                >
                  {{ t(getOAuthKey('openSessionUrl')) }}
                </button>
                <button
                  type="button"
                  class="oauth-flow__secondary-action btn btn-secondary text-xs"
                  @click="handleCopySoraSessionUrl"
                >
                  {{ t(getOAuthKey('copySessionUrl')) }}
                </button>
              </div>
              <p class="oauth-flow__note oauth-flow__note--break mt-1 text-xs">
                {{ soraSessionUrl }}
              </p>
              <p class="oauth-flow__alert-copy oauth-flow__alert-copy--warning mt-1 text-xs">
                {{ t(getOAuthKey('sessionUrlHint')) }}
              </p>
              <p
                v-if="parsedSoraTokenState.sessionTokenCount > 1"
                class="oauth-flow__note mt-1 text-xs"
              >
                {{ t('admin.accounts.oauth.batchCreateAccounts', { count: parsedSoraTokenState.sessionTokenCount }) }}
              </p>
            </div>

            <div v-if="sessionTokenInput.trim()" class="mb-4 space-y-3">
              <div>
                <label class="oauth-flow__field-label mb-2 flex items-center gap-2 text-xs font-semibold">
                  {{ t(getOAuthKey('parsedSessionTokensLabel')) }}
                  <span
                    v-if="parsedSoraTokenState.sessionTokenCount > 0"
                    class="theme-chip theme-chip--compact theme-chip--success"
                  >
                    {{ parsedSoraTokenState.sessionTokenCount }}
                  </span>
                </label>
                <textarea
                  :value="parsedSoraTokenState.sessionTokensText"
                  rows="2"
                  readonly
                  class="input oauth-flow__readonly-input w-full resize-y font-mono text-xs"
                ></textarea>
                <p
                  v-if="parsedSoraTokenState.sessionTokenCount === 0"
                  class="oauth-flow__alert-copy oauth-flow__alert-copy--warning mt-1 text-xs"
                >
                  {{ t(getOAuthKey('parsedSessionTokensEmpty')) }}
                </p>
              </div>

              <div>
                <label class="oauth-flow__field-label mb-2 flex items-center gap-2 text-xs font-semibold">
                  {{ t(getOAuthKey('parsedAccessTokensLabel')) }}
                  <span
                    v-if="parsedSoraTokenState.accessTokenCount > 0"
                    class="theme-chip theme-chip--compact theme-chip--success"
                  >
                    {{ parsedSoraTokenState.accessTokenCount }}
                  </span>
                </label>
                <textarea
                  :value="parsedSoraTokenState.accessTokensText"
                  rows="2"
                  readonly
                  class="input oauth-flow__readonly-input w-full resize-y font-mono text-xs"
                ></textarea>
              </div>
            </div>

            <div v-if="error" class="oauth-flow__alert oauth-flow__alert--danger mb-4">
              <p class="oauth-flow__alert-text whitespace-pre-line text-sm">
                {{ error }}
              </p>
            </div>

            <button
              type="button"
              class="btn btn-primary w-full"
              :disabled="loading || parsedSoraTokenState.sessionTokenCount === 0"
              @click="handleValidateSessionToken"
            >
              <svg
                v-if="loading"
                class="-ml-1 mr-2 h-4 w-4 animate-spin"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
              <Icon v-else name="sparkles" size="sm" class="mr-2" />
              {{ loading ? t(getOAuthKey('validating')) : t(getOAuthKey('validateAndCreate')) }}
            </button>
          </div>
        </div>

        <div v-if="inputMethod === 'access_token'" class="space-y-4">
          <div class="oauth-flow__panel">
            <p class="oauth-flow__panel-description mb-3 text-sm">
              {{
                t(
                  'admin.accounts.oauth.openai.accessTokenDesc',
                  '直接粘贴 Access Token 创建账号，无需 OAuth 授权流程。支持批量导入（每行一个）。'
                )
              }}
            </p>

            <div class="mb-4">
              <label class="oauth-flow__field-label mb-2 flex items-center gap-2 text-sm font-semibold">
                <Icon name="key" size="sm" class="oauth-flow__field-icon" />
                Access Token
                <span v-if="parsedAccessTokenCount > 1" class="theme-chip theme-chip--compact theme-chip--info">
                  {{ t('admin.accounts.oauth.keysCount', { count: parsedAccessTokenCount }) }}
                </span>
              </label>
              <textarea
                v-model="accessTokenInput"
                rows="3"
                class="input w-full resize-y font-mono text-sm"
                :placeholder="t('admin.accounts.oauth.openai.accessTokenPlaceholder', '粘贴 Access Token，每行一个')"
              ></textarea>
              <p v-if="parsedAccessTokenCount > 1" class="oauth-flow__note mt-1 text-xs">
                {{ t('admin.accounts.oauth.batchCreateAccounts', { count: parsedAccessTokenCount }) }}
              </p>
            </div>

            <div v-if="error" class="oauth-flow__alert oauth-flow__alert--danger mb-4">
              <p class="oauth-flow__alert-text whitespace-pre-line text-sm">
                {{ error }}
              </p>
            </div>

            <button
              type="button"
              class="btn btn-primary w-full"
              :disabled="loading || !accessTokenInput.trim()"
              @click="handleImportAccessToken"
            >
              <Icon name="sparkles" size="sm" class="mr-2" />
              {{ t('admin.accounts.oauth.openai.importAccessToken', '导入 Access Token') }}
            </button>
          </div>
        </div>

        <div v-if="inputMethod === 'cookie'" class="space-y-4">
          <div class="oauth-flow__panel">
            <p class="oauth-flow__panel-description mb-3 text-sm">
              {{ t('admin.accounts.oauth.cookieAutoAuthDesc') }}
            </p>

            <div class="mb-4">
              <label class="oauth-flow__field-label mb-2 flex items-center gap-2 text-sm font-semibold">
                <Icon name="key" size="sm" class="oauth-flow__field-icon" />
                {{ t('admin.accounts.oauth.sessionKey') }}
                <span
                  v-if="parsedKeyCount > 1 && allowMultiple"
                  class="theme-chip theme-chip--compact theme-chip--info"
                >
                  {{ t('admin.accounts.oauth.keysCount', { count: parsedKeyCount }) }}
                </span>
                <button
                  v-if="showHelp"
                  type="button"
                  class="oauth-flow__inline-link"
                  @click="showHelpDialog = !showHelpDialog"
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
                      d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z"
                    />
                  </svg>
                </button>
              </label>
              <textarea
                v-model="sessionKeyInput"
                rows="3"
                class="input w-full resize-y font-mono text-sm"
                :placeholder="
                  allowMultiple
                    ? t('admin.accounts.oauth.sessionKeyPlaceholder')
                    : t('admin.accounts.oauth.sessionKeyPlaceholderSingle')
                "
              ></textarea>
              <p
                v-if="parsedKeyCount > 1 && allowMultiple"
                class="oauth-flow__note mt-1 text-xs"
              >
                {{ t('admin.accounts.oauth.batchCreateAccounts', { count: parsedKeyCount }) }}
              </p>
            </div>

            <div
              v-if="showHelpDialog && showHelp"
              class="oauth-flow__alert oauth-flow__alert--warning mb-4"
            >
              <h5 class="oauth-flow__alert-title mb-2 font-semibold">
                {{ t('admin.accounts.oauth.howToGetSessionKey') }}
              </h5>
              <ol class="oauth-flow__alert-list list-inside list-decimal space-y-1 text-xs">
                <li>{{ t('admin.accounts.oauth.step1') }}</li>
                <li>{{ t('admin.accounts.oauth.step2') }}</li>
                <li>{{ t('admin.accounts.oauth.step3') }}</li>
                <li>{{ t('admin.accounts.oauth.step4') }}</li>
                <li>{{ t('admin.accounts.oauth.step5') }}</li>
                <li>{{ t('admin.accounts.oauth.step6') }}</li>
              </ol>
              <p
                class="oauth-flow__alert-copy oauth-flow__alert-copy--warning mt-2 text-xs"
                v-text="t('admin.accounts.oauth.sessionKeyFormat')"
              ></p>
            </div>

            <div v-if="error" class="oauth-flow__alert oauth-flow__alert--danger mb-4">
              <p class="oauth-flow__alert-text whitespace-pre-line text-sm">
                {{ error }}
              </p>
            </div>

            <button
              type="button"
              class="btn btn-primary w-full"
              :disabled="loading || !sessionKeyInput.trim()"
              @click="handleCookieAuth"
            >
              <svg
                v-if="loading"
                class="-ml-1 mr-2 h-4 w-4 animate-spin"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
              <Icon v-else name="sparkles" size="sm" class="mr-2" />
              {{ loading ? t('admin.accounts.oauth.authorizing') : t('admin.accounts.oauth.startAutoAuth') }}
            </button>
          </div>
        </div>

        <div v-if="inputMethod === 'manual'" class="space-y-4">
          <p class="oauth-flow__section-label mb-4 text-sm">
            {{ oauthText.followSteps }}
          </p>

          <div class="oauth-flow__panel">
            <div class="flex items-start gap-3">
              <div class="oauth-flow__step-badge flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-xs font-bold">
                1
              </div>
              <div class="flex-1">
                <p class="oauth-flow__panel-title mb-2 font-medium">
                  {{ oauthText.step1GenerateUrl }}
                </p>
                <div v-if="showProjectId && platform === 'gemini'" class="mb-3">
                  <label class="input-label flex items-center gap-2">
                    {{ t('admin.accounts.oauth.gemini.projectIdLabel') }}
                    <a
                      href="https://console.cloud.google.com/"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="oauth-flow__inline-link inline-flex items-center gap-1 text-xs font-normal"
                    >
                      <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z" />
                      </svg>
                      {{ t('admin.accounts.oauth.gemini.howToGetProjectId') }}
                    </a>
                  </label>
                  <input
                    v-model="projectId"
                    type="text"
                    class="input w-full font-mono text-sm"
                    :placeholder="t('admin.accounts.oauth.gemini.projectIdPlaceholder')"
                  />
                  <p class="oauth-flow__muted-note mt-1 text-xs">
                    {{ t('admin.accounts.oauth.gemini.projectIdHint') }}
                  </p>
                </div>
                <button
                  v-if="!authUrl"
                  type="button"
                  :disabled="loading"
                  class="btn btn-primary text-sm"
                  @click="handleGenerateUrl"
                >
                  <svg
                    v-if="loading"
                    class="-ml-1 mr-2 h-4 w-4 animate-spin"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path
                      class="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    ></path>
                  </svg>
                  <Icon v-else name="link" size="sm" class="mr-2" />
                  {{ loading ? t('admin.accounts.oauth.generating') : oauthText.generateAuthUrl }}
                </button>
                <div v-else class="space-y-3">
                  <div class="flex items-center gap-2">
                    <input
                      :value="authUrl"
                      readonly
                      type="text"
                      class="input oauth-flow__readonly-input flex-1 font-mono text-xs"
                    />
                    <button
                      type="button"
                      class="oauth-flow__icon-button btn btn-secondary"
                      title="Copy URL"
                      @click="handleCopyUrl"
                    >
                      <svg
                        v-if="!copied"
                        class="h-4 w-4"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        stroke-width="1.5"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184"
                        />
                      </svg>
                      <Icon v-else name="check" size="sm" class="oauth-flow__success-icon" :stroke-width="2" />
                    </button>
                  </div>
                  <button type="button" class="oauth-flow__inline-link text-xs" @click="handleRegenerate">
                    <Icon name="refresh" size="xs" class="mr-1 inline" />
                    {{ t('admin.accounts.oauth.regenerate') }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div class="oauth-flow__panel">
            <div class="flex items-start gap-3">
              <div class="oauth-flow__step-badge flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-xs font-bold">
                2
              </div>
              <div class="flex-1">
                <p class="oauth-flow__panel-title mb-2 font-medium">
                  {{ oauthText.step2OpenUrl }}
                </p>
                <p class="oauth-flow__panel-description text-sm">
                  {{ oauthText.openUrlDesc }}
                </p>
                <div v-if="isOpenAI" class="oauth-flow__alert oauth-flow__alert--warning mt-2">
                  <p class="oauth-flow__alert-text text-xs" v-text="oauthText.importantNotice"></p>
                </div>
                <div
                  v-else-if="showProxyWarning"
                  class="oauth-flow__alert oauth-flow__alert--notice mt-2"
                >
                  <p class="oauth-flow__alert-text text-xs" v-text="t('admin.accounts.oauth.proxyWarning')"></p>
                </div>
              </div>
            </div>
          </div>

          <div class="oauth-flow__panel">
            <div class="flex items-start gap-3">
              <div class="oauth-flow__step-badge flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-xs font-bold">
                3
              </div>
              <div class="flex-1">
                <p class="oauth-flow__panel-title mb-2 font-medium">
                  {{ oauthText.step3EnterCode }}
                </p>
                <p class="oauth-flow__panel-description mb-3 text-sm" v-text="oauthText.authCodeDesc"></p>
                <div>
                  <label class="input-label">
                    <Icon name="key" size="sm" class="oauth-flow__field-icon mr-1 inline" />
                    {{ oauthText.authCode }}
                  </label>
                  <textarea
                    v-model="authCodeInput"
                    rows="3"
                    class="input w-full resize-none font-mono text-sm"
                    :placeholder="oauthText.authCodePlaceholder"
                  ></textarea>
                  <p class="oauth-flow__muted-note mt-2 text-xs">
                    <Icon name="infoCircle" size="xs" class="mr-1 inline" />
                    {{ oauthText.authCodeHint }}
                  </p>

                  <div
                    v-if="platform === 'gemini'"
                    class="oauth-flow__alert oauth-flow__alert--warning-strong mt-3"
                  >
                    <div class="flex items-start gap-2">
                      <Icon
                        name="exclamationTriangle"
                        size="md"
                        class="oauth-flow__warning-icon flex-shrink-0"
                        :stroke-width="2"
                      />
                      <div class="oauth-flow__alert-text text-sm">
                        <p class="font-semibold">{{ $t('admin.accounts.oauth.gemini.stateWarningTitle') }}</p>
                        <p class="mt-1">{{ $t('admin.accounts.oauth.gemini.stateWarningDesc') }}</p>
                      </div>
                    </div>
                  </div>
                </div>

                <div v-if="error" class="oauth-flow__alert oauth-flow__alert--danger mt-3">
                  <p class="oauth-flow__alert-text whitespace-pre-line text-sm">
                    {{ error }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useClipboard } from '@/composables/useClipboard'
import {
  countMultilineEntries,
  extractNormalizedOAuthCallback,
  resolveOAuthImportantNoticeKey,
  resolveOAuthKey
} from '@/components/account/oauthAuthorizationFlowHelpers'
import { parseSoraRawTokens } from '@/utils/soraTokenParser'
import Icon from '@/components/icons/Icon.vue'
import type { AddMethod, AuthInputMethod } from '@/composables/useAccountOAuth'
import type { AccountPlatform } from '@/types'

interface Props {
  addMethod: AddMethod
  authUrl?: string
  sessionId?: string
  loading?: boolean
  error?: string
  showHelp?: boolean
  showProxyWarning?: boolean
  allowMultiple?: boolean
  methodLabel?: string
  showCookieOption?: boolean
  showRefreshTokenOption?: boolean
  showMobileRefreshTokenOption?: boolean
  showSessionTokenOption?: boolean
  showAccessTokenOption?: boolean
  platform?: AccountPlatform
  showProjectId?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  authUrl: '',
  sessionId: '',
  loading: false,
  error: '',
  showHelp: true,
  showProxyWarning: true,
  allowMultiple: false,
  methodLabel: 'Authorization Method',
  showCookieOption: true,
  showRefreshTokenOption: false,
  showMobileRefreshTokenOption: false,
  showSessionTokenOption: false,
  showAccessTokenOption: false,
  platform: 'anthropic',
  showProjectId: true
})

const emit = defineEmits<{
  'generate-url': []
  'exchange-code': [code: string]
  'cookie-auth': [sessionKey: string]
  'validate-refresh-token': [refreshToken: string]
  'validate-mobile-refresh-token': [refreshToken: string]
  'validate-session-token': [sessionToken: string]
  'import-access-token': [accessToken: string]
  'update:inputMethod': [method: AuthInputMethod]
}>()

const { t } = useI18n()

const isOpenAI = computed(() => props.platform === 'openai' || props.platform === 'sora')

const getOAuthKey = (key: string) => resolveOAuthKey(props.platform, key)

const oauthText = computed(() => {
  const translateCurrentPlatformKey = (key: string) => t(getOAuthKey(key))
  const importantNoticeKey = resolveOAuthImportantNoticeKey(props.platform)
  return {
    title: translateCurrentPlatformKey('title'),
    followSteps: translateCurrentPlatformKey('followSteps'),
    step1GenerateUrl: translateCurrentPlatformKey('step1GenerateUrl'),
    generateAuthUrl: translateCurrentPlatformKey('generateAuthUrl'),
    step2OpenUrl: translateCurrentPlatformKey('step2OpenUrl'),
    openUrlDesc: translateCurrentPlatformKey('openUrlDesc'),
    step3EnterCode: translateCurrentPlatformKey('step3EnterCode'),
    authCodeDesc: translateCurrentPlatformKey('authCodeDesc'),
    authCode: translateCurrentPlatformKey('authCode'),
    authCodePlaceholder: translateCurrentPlatformKey('authCodePlaceholder'),
    authCodeHint: translateCurrentPlatformKey('authCodeHint'),
    importantNotice: importantNoticeKey ? t(importantNoticeKey) : ''
  }
})

const inputMethod = ref<AuthInputMethod>('manual')
const authCodeInput = ref('')
const sessionKeyInput = ref('')
const refreshTokenInput = ref('')
const sessionTokenInput = ref('')
const accessTokenInput = ref('')
const showHelpDialog = ref(false)
const oauthState = ref('')
const projectId = ref('')

const showMethodSelection = computed(
  () =>
    props.showCookieOption ||
    props.showRefreshTokenOption ||
    props.showMobileRefreshTokenOption ||
    props.showSessionTokenOption ||
    props.showAccessTokenOption
)

const { copied, copyToClipboard } = useClipboard()

const parsedKeyCount = computed(() => countMultilineEntries(sessionKeyInput.value))
const parsedRefreshTokenCount = computed(() => countMultilineEntries(refreshTokenInput.value))

const parsedSoraTokenState = computed(() => {
  const parsed = parseSoraRawTokens(sessionTokenInput.value)
  return {
    sessionTokenCount: parsed.sessionTokens.length,
    sessionTokensText: parsed.sessionTokens.join('\n'),
    accessTokenCount: parsed.accessTokens.length,
    accessTokensText: parsed.accessTokens.join('\n')
  }
})

const soraSessionUrl = 'https://sora.chatgpt.com/api/auth/session'

const parsedAccessTokenCount = computed(() => countMultilineEntries(accessTokenInput.value))

watch(inputMethod, (newVal) => {
  emit('update:inputMethod', newVal)
})

watch(authCodeInput, (newVal) => {
  const normalized = extractNormalizedOAuthCallback(props.platform, newVal)
  if (normalized.oauthState) {
    oauthState.value = normalized.oauthState
  }
  if (normalized.nextInputValue) {
    authCodeInput.value = normalized.nextInputValue
  }
})

const handleGenerateUrl = () => {
  emit('generate-url')
}

const handleCopyUrl = () => {
  if (props.authUrl) {
    copyToClipboard(props.authUrl, 'URL copied to clipboard')
  }
}

const handleRegenerate = () => {
  authCodeInput.value = ''
  emit('generate-url')
}

const handleCookieAuth = () => {
  if (sessionKeyInput.value.trim()) {
    emit('cookie-auth', sessionKeyInput.value)
  }
}

const handleValidateRefreshToken = () => {
  if (refreshTokenInput.value.trim()) {
    if (inputMethod.value === 'mobile_refresh_token') {
      emit('validate-mobile-refresh-token', refreshTokenInput.value.trim())
    } else {
      emit('validate-refresh-token', refreshTokenInput.value.trim())
    }
  }
}

const handleValidateSessionToken = () => {
  if (parsedSoraTokenState.value.sessionTokenCount > 0) {
    emit('validate-session-token', parsedSoraTokenState.value.sessionTokensText)
  }
}

const handleOpenSoraSessionUrl = () => {
  window.open(soraSessionUrl, '_blank', 'noopener,noreferrer')
}

const handleCopySoraSessionUrl = () => {
  copyToClipboard(soraSessionUrl, 'URL copied to clipboard')
}

const handleImportAccessToken = () => {
  if (accessTokenInput.value.trim()) {
    emit('import-access-token', accessTokenInput.value.trim())
  }
}

function resetFlowState() {
  authCodeInput.value = ''
  oauthState.value = ''
  projectId.value = ''
  sessionKeyInput.value = ''
  refreshTokenInput.value = ''
  sessionTokenInput.value = ''
  accessTokenInput.value = ''
  inputMethod.value = 'manual'
  showHelpDialog.value = false
}

defineExpose({
  authCode: authCodeInput,
  oauthState,
  projectId,
  sessionKey: sessionKeyInput,
  refreshToken: refreshTokenInput,
  sessionToken: sessionTokenInput,
  inputMethod,
  reset: resetFlowState
})
</script>

<style scoped>
.oauth-flow {
  padding: var(--theme-auth-callback-card-padding);
  border-radius: var(--theme-surface-radius);
  border: 1px solid color-mix(in srgb, var(--theme-accent) 22%, var(--theme-card-border));
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--theme-accent-soft) 76%, var(--theme-surface)),
      color-mix(in srgb, var(--theme-surface-soft) 90%, var(--theme-surface))
    );
}

.oauth-flow__hero-icon {
  border-radius: calc(var(--theme-button-radius) + 2px);
  color: var(--theme-filled-text);
  background: linear-gradient(
    135deg,
    var(--theme-accent),
    color-mix(in srgb, var(--theme-accent-strong) 20%, var(--theme-accent) 80%)
  );
  box-shadow: 0 12px 28px color-mix(in srgb, var(--theme-accent) 24%, transparent);
}

.oauth-flow__title,
.oauth-flow__panel-title,
.oauth-flow__radio-label,
.oauth-flow__field-label {
  color: var(--theme-page-text);
}

.oauth-flow__section-label,
.oauth-flow__panel-description,
.oauth-flow__note {
  color: color-mix(in srgb, var(--theme-accent) 72%, var(--theme-page-text));
}

.oauth-flow__note--break {
  word-break: break-all;
}

.oauth-flow__muted-note {
  color: var(--theme-page-muted);
}

.oauth-flow__panel {
  padding: var(--theme-markdown-block-padding);
  border-radius: var(--theme-surface-radius);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 76%, transparent);
  background: color-mix(in srgb, var(--theme-page-backdrop) 74%, var(--theme-surface));
  box-shadow: inset 0 1px 0 color-mix(in srgb, var(--theme-surface-contrast) 6%, transparent);
}

.oauth-flow__radio {
  accent-color: var(--theme-accent);
}

.oauth-flow__field-icon,
.oauth-flow__inline-link {
  color: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-page-text));
}

.oauth-flow__inline-link:hover {
  color: color-mix(in srgb, var(--theme-accent-strong) 20%, var(--theme-accent) 80%);
}

.oauth-flow__readonly-input {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.oauth-flow__secondary-action {
  padding:
    var(--theme-account-usage-action-padding-y)
    var(--theme-settings-code-padding-x);
}

.oauth-flow__icon-button {
  padding: var(--theme-settings-code-padding-x);
}

.oauth-flow__step-badge {
  color: var(--theme-filled-text);
  background: color-mix(in srgb, var(--theme-accent) 82%, var(--theme-accent-strong));
}

.oauth-flow__success-icon {
  color: rgb(var(--theme-success-rgb));
}

.oauth-flow__warning-icon {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 82%, var(--theme-page-text));
}

.oauth-flow__alert {
  --oauth-flow-alert-rgb: var(--theme-warning-rgb);
  padding: var(--theme-auth-callback-feedback-padding);
  border-radius: var(--theme-auth-feedback-radius);
  border: 1px solid color-mix(in srgb, rgb(var(--oauth-flow-alert-rgb)) 20%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--oauth-flow-alert-rgb)) 10%, var(--theme-surface));
}

.oauth-flow__alert--danger {
  --oauth-flow-alert-rgb: var(--theme-danger-rgb);
}

.oauth-flow__alert--warning {
  --oauth-flow-alert-rgb: var(--theme-warning-rgb);
}

.oauth-flow__alert--notice {
  --oauth-flow-alert-rgb: var(--theme-warning-rgb);
}

.oauth-flow__alert--warning-strong {
  --oauth-flow-alert-rgb: var(--theme-warning-rgb);
  border-width: 2px;
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 12%, var(--theme-surface));
}

.oauth-flow__alert-title,
.oauth-flow__alert-text,
.oauth-flow__alert-list {
  color: color-mix(in srgb, rgb(var(--oauth-flow-alert-rgb)) 82%, var(--theme-page-text));
}

.oauth-flow__alert-copy {
  color: color-mix(in srgb, rgb(var(--oauth-flow-alert-rgb)) 78%, var(--theme-page-text));
}

.oauth-flow__alert-copy--warning {
  --oauth-flow-alert-rgb: var(--theme-warning-rgb);
}
</style>
