import {
  buildBulkAccountMutationPayload,
  type BuildBulkAccountMutationPayloadOptions,
} from "@/components/account/accountMutationPayload";

type ValueRef<T> = {
  value: T;
};

type BulkPayloadOptions = BuildBulkAccountMutationPayloadOptions;

interface UseBulkAccountMutationPayloadOptions {
  baseUrl: {
    enabled: ValueRef<boolean>;
    value: ValueRef<string>;
  };
  customErrorCodes: {
    enabled: ValueRef<boolean>;
    selectedErrorCodes: ValueRef<number[]>;
  };
  groups: {
    enabled: ValueRef<boolean>;
    groupIds: ValueRef<number[]>;
  };
  interceptWarmup: {
    enabled: ValueRef<boolean>;
    value: ValueRef<boolean>;
  };
  loadFactor: {
    enabled: ValueRef<boolean>;
    value: ValueRef<number | null>;
  };
  modelRestriction: {
    allowedModels: ValueRef<string[]>;
    disabledByOpenAIPassthrough: ValueRef<boolean>;
    enabled: ValueRef<boolean>;
    mode: ValueRef<BulkPayloadOptions["modelRestriction"]["mode"]>;
    modelMappings: ValueRef<BulkPayloadOptions["modelRestriction"]["modelMappings"]>;
  };
  openAI: {
    passthroughEnabled: ValueRef<boolean>;
    passthroughValue: ValueRef<boolean>;
    wsModeEnabled: ValueRef<boolean>;
    wsModeValue: ValueRef<BulkPayloadOptions["openAI"]["wsModeValue"]>;
  };
  proxy: {
    enabled: ValueRef<boolean>;
    proxyId: ValueRef<number | null>;
  };
  rpmLimit: {
    baseRpm: ValueRef<number | null>;
    enabled: ValueRef<boolean>;
    rpmEnabled: ValueRef<boolean>;
    stickyBuffer: ValueRef<number | null>;
    strategy: ValueRef<BulkPayloadOptions["rpmLimit"]["strategy"]>;
  };
  scalars: {
    concurrency: ValueRef<number>;
    enableConcurrency: ValueRef<boolean>;
    enablePriority: ValueRef<boolean>;
    enableRateMultiplier: ValueRef<boolean>;
    enableStatus: ValueRef<boolean>;
    priority: ValueRef<number>;
    rateMultiplier: ValueRef<number>;
    status: ValueRef<NonNullable<BulkPayloadOptions["scalars"]["status"]>>;
  };
  userMsgQueueMode: ValueRef<string | null>;
}

export function useBulkAccountMutationPayload(
  options: UseBulkAccountMutationPayloadOptions,
) {
  const hasAnyBulkEditFieldEnabled = () =>
    options.baseUrl.enabled.value ||
    options.openAI.passthroughEnabled.value ||
    options.modelRestriction.enabled.value ||
    options.customErrorCodes.enabled.value ||
    options.interceptWarmup.enabled.value ||
    options.proxy.enabled.value ||
    options.scalars.enableConcurrency.value ||
    options.loadFactor.enabled.value ||
    options.scalars.enablePriority.value ||
    options.scalars.enableRateMultiplier.value ||
    options.scalars.enableStatus.value ||
    options.groups.enabled.value ||
    options.openAI.wsModeEnabled.value ||
    options.rpmLimit.enabled.value ||
    options.userMsgQueueMode.value !== null;

  const buildBulkEditPayload = () =>
    buildBulkAccountMutationPayload({
      baseUrl: {
        enabled: options.baseUrl.enabled.value,
        value: options.baseUrl.value.value,
      },
      customErrorCodes: {
        enabled: options.customErrorCodes.enabled.value,
        selectedErrorCodes: options.customErrorCodes.selectedErrorCodes.value,
      },
      groups: {
        enabled: options.groups.enabled.value,
        groupIds: options.groups.groupIds.value,
      },
      interceptWarmup: {
        enabled: options.interceptWarmup.enabled.value,
        value: options.interceptWarmup.value.value,
      },
      loadFactor: {
        enabled: options.loadFactor.enabled.value,
        value: options.loadFactor.value.value,
      },
      modelRestriction: {
        allowedModels: options.modelRestriction.allowedModels.value,
        disabledByOpenAIPassthrough:
          options.modelRestriction.disabledByOpenAIPassthrough.value,
        enabled: options.modelRestriction.enabled.value,
        mode: options.modelRestriction.mode.value,
        modelMappings: options.modelRestriction.modelMappings.value,
      },
      openAI: {
        passthroughEnabled: options.openAI.passthroughEnabled.value,
        passthroughValue: options.openAI.passthroughValue.value,
        wsModeEnabled: options.openAI.wsModeEnabled.value,
        wsModeValue: options.openAI.wsModeValue.value,
      },
      proxy: {
        enabled: options.proxy.enabled.value,
        proxyId: options.proxy.proxyId.value,
      },
      rpmLimit: {
        baseRpm: options.rpmLimit.baseRpm.value,
        enabled: options.rpmLimit.enabled.value,
        rpmEnabled: options.rpmLimit.rpmEnabled.value,
        stickyBuffer: options.rpmLimit.stickyBuffer.value,
        strategy: options.rpmLimit.strategy.value,
      },
      scalars: {
        concurrency: options.scalars.concurrency.value,
        enableConcurrency: options.scalars.enableConcurrency.value,
        enablePriority: options.scalars.enablePriority.value,
        enableRateMultiplier: options.scalars.enableRateMultiplier.value,
        enableStatus: options.scalars.enableStatus.value,
        priority: options.scalars.priority.value,
        rateMultiplier: options.scalars.rateMultiplier.value,
        status: options.scalars.status.value,
      },
      userMsgQueueMode: options.userMsgQueueMode.value,
    });

  return {
    buildBulkEditPayload,
    hasAnyBulkEditFieldEnabled,
  };
}
