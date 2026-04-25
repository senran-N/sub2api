import { describe, expect, it } from "vitest";
import { computed, ref } from "vue";
import { OPENAI_WS_MODE_OFF, type OpenAIWSMode } from "@/utils/openaiWsMode";
import { useBulkAccountMutationPayload } from "../useBulkAccountMutationPayload";

function createBuilder() {
  const enableBaseUrl = ref(false);
  const baseUrl = ref("");
  const enableCustomErrorCodes = ref(false);
  const selectedErrorCodes = ref<number[]>([]);
  const enableGroups = ref(false);
  const groupIds = ref<number[]>([]);
  const enableInterceptWarmup = ref(false);
  const interceptWarmupRequests = ref(false);
  const enableLoadFactor = ref(false);
  const loadFactor = ref<number | null>(null);
  const enableModelRestriction = ref(false);
  const allowedModels = ref<string[]>([]);
  const isOpenAIModelRestrictionDisabled = computed(() => false);
  const modelRestrictionMode = ref<"whitelist" | "mapping">("whitelist");
  const modelMappings = ref<Array<{ from: string; to: string }>>([]);
  const enableOpenAIPassthrough = ref(false);
  const openaiPassthroughEnabled = ref(false);
  const enableOpenAIWSMode = ref(false);
  const openAIWSMode = ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF);
  const enableProxy = ref(false);
  const proxyId = ref<number | null>(null);
  const enableRpmLimit = ref(false);
  const rpmLimitEnabled = ref(false);
  const bulkBaseRpm = ref<number | null>(null);
  const bulkRpmStickyBuffer = ref<number | null>(null);
  const bulkRpmStrategy = ref<"tiered" | "sticky_exempt">("tiered");
  const enableConcurrency = ref(false);
  const enablePriority = ref(false);
  const enableRateMultiplier = ref(false);
  const enableStatus = ref(false);
  const concurrency = ref(1);
  const priority = ref(1);
  const rateMultiplier = ref(1);
  const status = ref<"active" | "inactive">("active");
  const userMsgQueueMode = ref<string | null>(null);
  const payload = useBulkAccountMutationPayload({
    baseUrl: {
      enabled: enableBaseUrl,
      value: baseUrl,
    },
    customErrorCodes: {
      enabled: enableCustomErrorCodes,
      selectedErrorCodes,
    },
    groups: {
      enabled: enableGroups,
      groupIds,
    },
    interceptWarmup: {
      enabled: enableInterceptWarmup,
      value: interceptWarmupRequests,
    },
    loadFactor: {
      enabled: enableLoadFactor,
      value: loadFactor,
    },
    modelRestriction: {
      allowedModels,
      disabledByOpenAIPassthrough: isOpenAIModelRestrictionDisabled,
      enabled: enableModelRestriction,
      mode: modelRestrictionMode,
      modelMappings,
    },
    openAI: {
      passthroughEnabled: enableOpenAIPassthrough,
      passthroughValue: openaiPassthroughEnabled,
      wsModeEnabled: enableOpenAIWSMode,
      wsModeValue: openAIWSMode,
    },
    proxy: {
      enabled: enableProxy,
      proxyId,
    },
    rpmLimit: {
      baseRpm: bulkBaseRpm,
      enabled: enableRpmLimit,
      rpmEnabled: rpmLimitEnabled,
      stickyBuffer: bulkRpmStickyBuffer,
      strategy: bulkRpmStrategy,
    },
    scalars: {
      concurrency,
      enableConcurrency,
      enablePriority,
      enableRateMultiplier,
      enableStatus,
      priority,
      rateMultiplier,
      status,
    },
    userMsgQueueMode,
  });

  return {
    ...payload,
    allowedModels,
    baseUrl,
    concurrency,
    enableBaseUrl,
    enableConcurrency,
    enableModelRestriction,
    userMsgQueueMode,
  };
}

describe("useBulkAccountMutationPayload", () => {
  it("reports no selected fields and builds no payload by default", () => {
    const builder = createBuilder();

    expect(builder.hasAnyBulkEditFieldEnabled()).toBe(false);
    expect(builder.buildBulkEditPayload()).toBeNull();
  });

  it("collects enabled field refs into a bulk update payload", () => {
    const builder = createBuilder();

    builder.enableBaseUrl.value = true;
    builder.baseUrl.value = "https://proxy.example";
    builder.enableModelRestriction.value = true;
    builder.allowedModels.value = ["gpt-5.4"];
    builder.enableConcurrency.value = true;
    builder.concurrency.value = 8;
    builder.userMsgQueueMode.value = "sticky";

    expect(builder.hasAnyBulkEditFieldEnabled()).toBe(true);
    expect(builder.buildBulkEditPayload()).toEqual({
      concurrency: 8,
      credentials: {
        base_url: "https://proxy.example",
        model_mapping: {
          "gpt-5.4": "gpt-5.4",
        },
      },
      extra: {
        user_msg_queue_enabled: false,
        user_msg_queue_mode: "sticky",
      },
    });
  });
});
