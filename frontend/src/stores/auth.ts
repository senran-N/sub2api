/**
 * Authentication Store
 * Manages user authentication state, login/logout, token refresh, and token persistence
 */

import { defineStore } from "pinia";
import { ref, computed, readonly } from "vue";
import type { LoginResponse } from "@/api/auth";
import type {
  User,
  LoginRequest,
  RegisterRequest,
  AuthResponse,
  TotpLoginResponse,
} from "@/types";
import {
  setAuthTokenRefreshHandler,
  type AuthTokenRefreshedDetail,
} from "./authSync";
import {
  AUTH_TOKEN_KEY,
  AUTH_USER_KEY,
  REFRESH_TOKEN_KEY,
  TOKEN_EXPIRES_AT_KEY,
} from "@/utils/authStorage";

const AUTO_REFRESH_INTERVAL = 60 * 1000; // 60 seconds for user data refresh
const TOKEN_REFRESH_BUFFER = 120 * 1000; // 120 seconds before expiry to refresh token

type AuthApiModule = typeof import("@/api/auth");

let authApiModulePromise: Promise<AuthApiModule> | null = null;

function loadAuthApiModule(): Promise<AuthApiModule> {
  if (!authApiModulePromise) {
    authApiModulePromise = import("@/api/auth");
  }

  return authApiModulePromise;
}

function requiresTotpLogin(
  response: LoginResponse,
): response is TotpLoginResponse {
  return "requires_2fa" in response && response.requires_2fa === true;
}

export const useAuthStore = defineStore("auth", () => {
  const user = ref<User | null>(null);
  const token = ref<string | null>(null);
  const refreshTokenValue = ref<string | null>(null);
  const tokenExpiresAt = ref<number | null>(null);
  const runMode = ref<"standard" | "simple">("standard");
  let refreshIntervalId: ReturnType<typeof setInterval> | null = null;
  let tokenRefreshTimeoutId: ReturnType<typeof setTimeout> | null = null;
  let authOperationQueue: Promise<unknown> = Promise.resolve();
  let tokenRefreshPromise: Promise<void> | null = null;

  function runSerializedAuthOperation<T>(
    operation: () => Promise<T> | T,
  ): Promise<T> {
    const next = authOperationQueue.then(operation, operation);
    authOperationQueue = next.then(
      () => undefined,
      () => undefined,
    );
    return next;
  }

  function stopAutoRefresh(): void {
    if (refreshIntervalId) {
      clearInterval(refreshIntervalId);
      refreshIntervalId = null;
    }
  }

  function stopTokenRefresh(): void {
    if (tokenRefreshTimeoutId) {
      clearTimeout(tokenRefreshTimeoutId);
      tokenRefreshTimeoutId = null;
    }
  }

  function clearAuth(): void {
    stopAutoRefresh();
    stopTokenRefresh();

    token.value = null;
    refreshTokenValue.value = null;
    tokenExpiresAt.value = null;
    user.value = null;
    runMode.value = "standard";
    localStorage.removeItem(AUTH_TOKEN_KEY);
    localStorage.removeItem(AUTH_USER_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
    localStorage.removeItem(TOKEN_EXPIRES_AT_KEY);
  }

  function scheduleTokenRefreshAt(expiresAtMs: number): void {
    stopTokenRefresh();

    const now = Date.now();
    const refreshInMs = Math.max(0, expiresAtMs - now - TOKEN_REFRESH_BUFFER);

    if (refreshInMs <= 0) {
      void performTokenRefresh();
      return;
    }

    tokenRefreshTimeoutId = setTimeout(() => {
      void performTokenRefresh();
    }, refreshInMs);
  }

  function applyTokenRefresh(detail: AuthTokenRefreshedDetail): void {
    token.value = detail.access_token;
    localStorage.setItem(AUTH_TOKEN_KEY, detail.access_token);

    if (detail.refresh_token) {
      refreshTokenValue.value = detail.refresh_token;
      localStorage.setItem(REFRESH_TOKEN_KEY, detail.refresh_token);
    }
    if (
      typeof detail.expires_at === "number" &&
      Number.isFinite(detail.expires_at)
    ) {
      tokenExpiresAt.value = detail.expires_at;
      localStorage.setItem(TOKEN_EXPIRES_AT_KEY, String(detail.expires_at));
      scheduleTokenRefreshAt(detail.expires_at);
    }
  }

  function bindTokenRefreshListener(): void {
    setAuthTokenRefreshHandler((detail: AuthTokenRefreshedDetail) => {
      if (!detail?.access_token) {
        return;
      }

      void runSerializedAuthOperation(() => {
        applyTokenRefresh(detail);
      });
    });
  }

  bindTokenRefreshListener();

  const isAuthenticated = computed(() => !!token.value && !!user.value);
  const isAdmin = computed(() => user.value?.role === "admin");
  const isSimpleMode = computed(() => runMode.value === "simple");

  function startAutoRefresh(): void {
    stopAutoRefresh();

    refreshIntervalId = setInterval(() => {
      if (token.value) {
        refreshUser().catch((error) => {
          console.error("Auto-refresh user failed:", error);
        });
      }
    }, AUTO_REFRESH_INTERVAL);
  }

  function scheduleTokenRefresh(expiresInSeconds: number): void {
    const expiresAtMs = Date.now() + expiresInSeconds * 1000;
    tokenExpiresAt.value = expiresAtMs;
    localStorage.setItem(TOKEN_EXPIRES_AT_KEY, String(expiresAtMs));
    scheduleTokenRefreshAt(expiresAtMs);
  }

  function setAuthFromResponse(response: AuthResponse): void {
    token.value = response.access_token;
    localStorage.setItem(AUTH_TOKEN_KEY, response.access_token);

    if (response.refresh_token) {
      refreshTokenValue.value = response.refresh_token;
      localStorage.setItem(REFRESH_TOKEN_KEY, response.refresh_token);
    }

    if (response.user.run_mode) {
      runMode.value = response.user.run_mode;
    }
    const { run_mode: _runMode, ...userData } = response.user;
    user.value = userData;
    localStorage.setItem(AUTH_USER_KEY, JSON.stringify(userData));

    startAutoRefresh();

    if (response.refresh_token && response.expires_in) {
      scheduleTokenRefresh(response.expires_in);
    }
  }

  async function refreshUserInternal(): Promise<User> {
    if (!token.value) {
      throw new Error("Not authenticated");
    }

    try {
      const { getCurrentUser } = await loadAuthApiModule();
      const response = await getCurrentUser();
      if (response.data.run_mode) {
        runMode.value = response.data.run_mode;
      }
      const { run_mode: _runMode, ...userData } = response.data;
      user.value = userData;
      localStorage.setItem(AUTH_USER_KEY, JSON.stringify(userData));
      return userData;
    } catch (error) {
      if ((error as { status?: number }).status === 401) {
        clearAuth();
      }
      throw error;
    }
  }

  function checkAuth(): void {
    const savedToken = localStorage.getItem(AUTH_TOKEN_KEY);
    const savedUser = localStorage.getItem(AUTH_USER_KEY);
    const savedRefreshToken = localStorage.getItem(REFRESH_TOKEN_KEY);
    const savedExpiresAt = localStorage.getItem(TOKEN_EXPIRES_AT_KEY);

    if (!savedToken || !savedUser) {
      return;
    }

    try {
      token.value = savedToken;
      user.value = JSON.parse(savedUser);
      refreshTokenValue.value = savedRefreshToken;
      tokenExpiresAt.value = savedExpiresAt
        ? parseInt(savedExpiresAt, 10)
        : null;

      refreshUser().catch((error) => {
        console.error("Failed to refresh user on init:", error);
      });

      startAutoRefresh();

      if (savedRefreshToken && tokenExpiresAt.value !== null) {
        scheduleTokenRefreshAt(tokenExpiresAt.value);
      }
    } catch (error) {
      console.error("Failed to parse saved user data:", error);
      clearAuth();
    }
  }

  async function performTokenRefresh(): Promise<void> {
    if (tokenRefreshPromise) {
      return tokenRefreshPromise;
    }

    tokenRefreshPromise = runSerializedAuthOperation(async () => {
      if (!refreshTokenValue.value) {
        return;
      }

      try {
        const { refreshToken } = await loadAuthApiModule();
        const response = await refreshToken();

        applyTokenRefresh({
          access_token: response.access_token,
          refresh_token: response.refresh_token,
          expires_at: Date.now() + response.expires_in * 1000,
        });
      } catch (error) {
        console.error("Token refresh failed:", error);
      }
    });

    try {
      await tokenRefreshPromise;
    } finally {
      tokenRefreshPromise = null;
    }
  }

  async function login(credentials: LoginRequest): Promise<LoginResponse> {
    return runSerializedAuthOperation(async () => {
      try {
        const { login } = await loadAuthApiModule();
        const response = await login(credentials);

        if (requiresTotpLogin(response)) {
          return response;
        }

        setAuthFromResponse(response);
        return response;
      } catch (error) {
        clearAuth();
        throw error;
      }
    });
  }

  async function login2FA(tempToken: string, totpCode: string): Promise<User> {
    return runSerializedAuthOperation(async () => {
      try {
        const { login2FA } = await loadAuthApiModule();
        const response = await login2FA({
          temp_token: tempToken,
          totp_code: totpCode,
        });
        setAuthFromResponse(response);
        return user.value!;
      } catch (error) {
        clearAuth();
        throw error;
      }
    });
  }

  async function register(userData: RegisterRequest): Promise<User> {
    return runSerializedAuthOperation(async () => {
      try {
        const { register } = await loadAuthApiModule();
        const response = await register(userData);
        setAuthFromResponse(response);
        return user.value!;
      } catch (error) {
        clearAuth();
        throw error;
      }
    });
  }

  async function setToken(newToken: string): Promise<User> {
    return runSerializedAuthOperation(async () => {
      stopAutoRefresh();
      stopTokenRefresh();
      token.value = null;
      user.value = null;

      token.value = newToken;
      localStorage.setItem(AUTH_TOKEN_KEY, newToken);

      const savedRefreshToken = localStorage.getItem(REFRESH_TOKEN_KEY);
      const savedExpiresAt = localStorage.getItem(TOKEN_EXPIRES_AT_KEY);

      refreshTokenValue.value = savedRefreshToken;
      tokenExpiresAt.value = savedExpiresAt
        ? parseInt(savedExpiresAt, 10)
        : null;

      try {
        const userData = await refreshUserInternal();
        startAutoRefresh();

        if (savedRefreshToken && tokenExpiresAt.value !== null) {
          scheduleTokenRefreshAt(tokenExpiresAt.value);
        }

        return userData;
      } catch (error) {
        clearAuth();
        throw error;
      }
    });
  }

  async function logout(): Promise<void> {
    await runSerializedAuthOperation(async () => {
      const { logout } = await loadAuthApiModule();
      await logout();
      clearAuth();
    });
  }

  async function refreshUser(): Promise<User> {
    return runSerializedAuthOperation(() => refreshUserInternal());
  }

  return {
    user,
    token,
    runMode: readonly(runMode),
    isAuthenticated,
    isAdmin,
    isSimpleMode,
    login,
    login2FA,
    register,
    setToken,
    logout,
    checkAuth,
    refreshUser,
  };
});
