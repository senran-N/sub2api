import type { AccountPlatform } from "@/types";
import {
  createPlatformRequestGuard,
  createSequenceRequestGuard,
  type AccountModalPlatformRequestContext,
} from "@/components/account/accountModalRequestGuard";

export type CreateAccountRequestContext =
  AccountModalPlatformRequestContext<AccountPlatform>;

interface UseCreateAccountRequestGuardsOptions {
  getAccountCategory: () => string;
  getPlatform: () => AccountPlatform;
  isModalOpen: () => boolean;
}

export function useCreateAccountRequestGuards(
  options: UseCreateAccountRequestGuardsOptions,
) {
  let allowedModelsSyncSequence = 0;

  const createRequestGuard = createPlatformRequestGuard<AccountPlatform>(
    (platform) => options.isModalOpen() && options.getPlatform() === platform,
  );
  const tlsFingerprintProfilesRequestGuard = createSequenceRequestGuard(
    options.isModalOpen,
  );
  const antigravityDefaultMappingsRequestGuard = createSequenceRequestGuard(
    () => options.isModalOpen() && options.getPlatform() === "antigravity",
  );
  const geminiCapabilitiesRequestGuard = createSequenceRequestGuard(
    () =>
      options.isModalOpen() &&
      options.getPlatform() === "gemini" &&
      options.getAccountCategory() === "oauth-based",
  );

  const beginCreateRequestContext = (
    platform: AccountPlatform = options.getPlatform(),
  ): CreateAccountRequestContext => createRequestGuard.begin(platform);

  const invalidateCreateRequests = () => {
    createRequestGuard.invalidate();
  };

  const isActiveCreateRequest = (requestContext: CreateAccountRequestContext) =>
    createRequestGuard.isActive(requestContext);

  const isCurrentCreateRequestSequence = (
    requestContext: CreateAccountRequestContext,
  ) => createRequestGuard.isCurrentSequence(requestContext);

  const getCurrentCreateRequestSequence = () =>
    createRequestGuard.currentSequence();

  const beginTlsFingerprintProfilesRequest = () =>
    tlsFingerprintProfilesRequestGuard.begin();

  const invalidateTlsFingerprintProfilesRequests = () => {
    tlsFingerprintProfilesRequestGuard.invalidate();
  };

  const isActiveTlsFingerprintProfilesRequest = (requestSequence: number) =>
    tlsFingerprintProfilesRequestGuard.isActive(requestSequence);

  const beginAntigravityDefaultMappingsRequest = () =>
    antigravityDefaultMappingsRequestGuard.begin();

  const invalidateAntigravityDefaultMappingsRequests = () => {
    antigravityDefaultMappingsRequestGuard.invalidate();
  };

  const isActiveAntigravityDefaultMappingsRequest = (
    requestSequence: number,
  ) => antigravityDefaultMappingsRequestGuard.isActive(requestSequence);

  const beginGeminiCapabilitiesRequest = () =>
    geminiCapabilitiesRequestGuard.begin();

  const invalidateGeminiCapabilitiesRequests = () => {
    geminiCapabilitiesRequestGuard.invalidate();
  };

  const isActiveGeminiCapabilitiesRequest = (requestSequence: number) =>
    geminiCapabilitiesRequestGuard.isActive(requestSequence);

  const nextAllowedModelsSyncSequence = () => {
    allowedModelsSyncSequence += 1;
    return allowedModelsSyncSequence;
  };

  const invalidateAllowedModelsSync = () => {
    allowedModelsSyncSequence += 1;
  };

  const isCurrentAllowedModelsSync = (requestSequence: number) =>
    requestSequence === allowedModelsSyncSequence;

  const invalidateCreateModalAsyncLoads = () => {
    invalidateTlsFingerprintProfilesRequests();
    invalidateAntigravityDefaultMappingsRequests();
    invalidateGeminiCapabilitiesRequests();
    invalidateAllowedModelsSync();
  };

  return {
    beginAntigravityDefaultMappingsRequest,
    beginCreateRequestContext,
    beginGeminiCapabilitiesRequest,
    beginTlsFingerprintProfilesRequest,
    getCurrentCreateRequestSequence,
    invalidateAllowedModelsSync,
    invalidateAntigravityDefaultMappingsRequests,
    invalidateCreateModalAsyncLoads,
    invalidateCreateRequests,
    invalidateGeminiCapabilitiesRequests,
    invalidateTlsFingerprintProfilesRequests,
    isActiveAntigravityDefaultMappingsRequest,
    isActiveCreateRequest,
    isActiveGeminiCapabilitiesRequest,
    isActiveTlsFingerprintProfilesRequest,
    isCurrentAllowedModelsSync,
    isCurrentCreateRequestSequence,
    nextAllowedModelsSyncSequence,
  };
}
