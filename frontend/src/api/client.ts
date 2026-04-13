/**
 * Axios HTTP Client Configuration
 * Base client with interceptors for authentication, token refresh, and error handling
 */

import axios, {
  AxiosInstance,
  AxiosError,
  InternalAxiosRequestConfig,
  AxiosResponse,
} from "axios";
import type { ApiResponse, JsonObject } from "@/types";
import { getLocale } from "@/i18n";
import { emitAuthTokenRefreshed } from "@/stores/authSync";
import {
  AUTH_TOKEN_KEY,
  REFRESH_TOKEN_KEY,
  TOKEN_EXPIRES_AT_KEY,
  AUTH_USER_KEY,
} from "@/utils/authStorage";

// ==================== Axios Instance Configuration ====================

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "/api/v1";

export const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    "Content-Type": "application/json",
  },
});

// ==================== Token Refresh State ====================

// Track if a token refresh is in progress to prevent multiple simultaneous refresh requests
let isRefreshing = false;
// Queue of requests waiting for token refresh
let refreshSubscribers: Array<(token: string) => void> = [];
let pendingRedirectTarget: string | null = null;

function redirectOnce(target: string): void {
  if (window.location.pathname === target || pendingRedirectTarget === target) {
    return;
  }

  pendingRedirectTarget = target;
  window.location.href = target;
}

/**
 * Subscribe to token refresh completion
 */
function subscribeTokenRefresh(callback: (token: string) => void): void {
  refreshSubscribers.push(callback);
}

/**
 * Notify all subscribers that token has been refreshed
 */
function onTokenRefreshed(token: string): void {
  refreshSubscribers.forEach((callback) => callback(token));
  refreshSubscribers = [];
}

// ==================== Request Interceptor ====================

// Get user's timezone
const getUserTimezone = (): string => {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone;
  } catch {
    return "UTC";
  }
};

apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // Attach token from localStorage
    const token = localStorage.getItem(AUTH_TOKEN_KEY);
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }

    // Attach locale for backend translations
    if (config.headers) {
      config.headers["Accept-Language"] = getLocale();
    }

    // Attach timezone for all GET requests (backend may use it for default date ranges)
    if (config.method === "get") {
      if (!config.params) {
        config.params = {};
      }
      config.params.timezone = getUserTimezone();
    }

    return config;
  },
  (error) => {
    return Promise.reject(error);
  },
);

// ==================== Response Interceptor ====================

apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    // Unwrap standard API response format { code, message, data }
    const apiResponse = response.data as ApiResponse<unknown>;
    if (
      apiResponse &&
      typeof apiResponse === "object" &&
      "code" in apiResponse
    ) {
      if (apiResponse.code === 0) {
        // Success - return the data portion
        response.data = apiResponse.data;
      } else {
        // API error
        return Promise.reject({
          status: response.status,
          code: apiResponse.code,
          message: apiResponse.message || "Unknown error",
        });
      }
    }
    return response;
  },
  async (error: AxiosError<ApiResponse<unknown>>) => {
    // Request cancellation: keep the original axios cancellation error so callers can ignore it.
    // Otherwise we'd misclassify it as a generic "network error".
    if (error.code === "ERR_CANCELED" || axios.isCancel(error)) {
      return Promise.reject(error);
    }

    const originalRequest = error.config as InternalAxiosRequestConfig & {
      _retry?: boolean;
    };

    // Handle common errors
    if (error.response) {
      const { status, data } = error.response;
      const url = String(error.config?.url || "");

      // Validate `data` shape to avoid HTML error pages breaking our error handling.
      const apiData = (typeof data === "object" && data !== null ? data : {}) as JsonObject;
      const apiCode = apiData.code;
      const apiError = apiData.error;
      const apiMessage = typeof apiData.message === "string" ? apiData.message : undefined;
      const apiDetail = typeof apiData.detail === "string" ? apiData.detail : undefined;

      // Ops monitoring disabled: treat as feature-flagged 404, and proactively redirect away
      // from ops pages to avoid broken UI states.
      if (status === 404 && apiMessage === "Ops monitoring is disabled") {
        try {
          localStorage.setItem("ops_monitoring_enabled_cached", "false");
        } catch {
          // ignore localStorage failures
        }
        try {
          window.dispatchEvent(new CustomEvent("ops-monitoring-disabled"));
        } catch {
          // ignore event failures
        }

        if (window.location.pathname.startsWith("/admin/ops")) {
          redirectOnce("/admin/settings");
        }

        return Promise.reject({
          status,
          code: "OPS_DISABLED",
          message: apiMessage || error.message,
          url,
        });
      }

      // 401: Try to refresh the token if we have a refresh token
      // This handles TOKEN_EXPIRED, INVALID_TOKEN, TOKEN_REVOKED, etc.
      if (status === 401 && !originalRequest._retry) {
        const refreshToken = localStorage.getItem(REFRESH_TOKEN_KEY);
        const isAuthEndpoint =
          url.includes("/auth/login") ||
          url.includes("/auth/register") ||
          url.includes("/auth/refresh");

        // If we have a refresh token and this is not an auth endpoint, try to refresh
        if (refreshToken && !isAuthEndpoint) {
          if (isRefreshing) {
            // Wait for the ongoing refresh to complete
            return new Promise((resolve, reject) => {
              subscribeTokenRefresh((newToken: string) => {
                if (newToken) {
                  // Mark as retried to prevent infinite loop if retry also returns 401
                  originalRequest._retry = true;
                  if (originalRequest.headers) {
                    originalRequest.headers.Authorization = `Bearer ${newToken}`;
                  }
                  resolve(apiClient(originalRequest));
                } else {
                  // Refresh failed, reject with original error
                  reject({
                    status,
                    code: apiCode,
                    message: apiMessage || apiDetail || error.message,
                  });
                }
              });
            });
          }

          originalRequest._retry = true;
          isRefreshing = true;

          try {
            // Call refresh endpoint directly to avoid circular dependency
            const refreshResponse = await axios.post(
              `${API_BASE_URL}/auth/refresh`,
              { refresh_token: refreshToken },
              { headers: { "Content-Type": "application/json" } },
            );

            const refreshData = refreshResponse.data as ApiResponse<{
              access_token: string;
              refresh_token: string;
              expires_in: number;
            }>;

            if (refreshData.code === 0 && refreshData.data) {
              const {
                access_token,
                refresh_token: newRefreshToken,
                expires_in,
              } = refreshData.data;
              const expiresAt = Date.now() + expires_in * 1000;

              // Update tokens in localStorage (convert expires_in to timestamp)
              localStorage.setItem(AUTH_TOKEN_KEY, access_token);
              localStorage.setItem(REFRESH_TOKEN_KEY, newRefreshToken);
              localStorage.setItem(TOKEN_EXPIRES_AT_KEY, String(expiresAt));
              emitAuthTokenRefreshed({
                access_token,
                refresh_token: newRefreshToken,
                expires_at: expiresAt,
              });

              // Notify subscribers with new token
              onTokenRefreshed(access_token);

              // Retry the original request with new token
              if (originalRequest.headers) {
                originalRequest.headers.Authorization = `Bearer ${access_token}`;
              }

              isRefreshing = false;
              return apiClient(originalRequest);
            }

            // Refresh response was not successful, fall through to clear auth
            throw new Error("Token refresh failed");
          } catch (refreshError) {
            // Refresh failed - notify subscribers with empty token
            onTokenRefreshed("");
            isRefreshing = false;

            // Clear tokens and redirect to login
            localStorage.removeItem(AUTH_TOKEN_KEY);
            localStorage.removeItem(REFRESH_TOKEN_KEY);
            localStorage.removeItem(AUTH_USER_KEY);
            localStorage.removeItem(TOKEN_EXPIRES_AT_KEY);
            sessionStorage.setItem("auth_expired", "1");

            redirectOnce("/login");

            return Promise.reject({
              status: 401,
              code: "TOKEN_REFRESH_FAILED",
              message: "Session expired. Please log in again.",
            });
          }
        }

        // No refresh token or is auth endpoint - clear auth and redirect
        const hasToken = !!localStorage.getItem(AUTH_TOKEN_KEY);
        const headers = error.config?.headers as
          | Record<string, unknown>
          | undefined;
        const authHeader = headers?.Authorization ?? headers?.authorization;
        const sentAuth =
          typeof authHeader === "string"
            ? authHeader.trim() !== ""
            : Array.isArray(authHeader)
              ? authHeader.length > 0
              : !!authHeader;

        localStorage.removeItem(AUTH_TOKEN_KEY);
        localStorage.removeItem(REFRESH_TOKEN_KEY);
        localStorage.removeItem(AUTH_USER_KEY);
        localStorage.removeItem(TOKEN_EXPIRES_AT_KEY);
        if ((hasToken || sentAuth) && !isAuthEndpoint) {
          sessionStorage.setItem("auth_expired", "1");
        }
        // Only redirect if not already on login page
        redirectOnce("/login");
      }

      // Return structured error
      return Promise.reject({
        status,
        code: apiCode,
        error: apiError,
        message: apiMessage || apiDetail || error.message,
      });
    }

    // Network error
    return Promise.reject({
      status: 0,
      message: "Network error. Please check your connection.",
    });
  },
);

export default apiClient;
