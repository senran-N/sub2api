import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import BedrockCredentialsSection from "../BedrockCredentialsSection.vue";
import EditBedrockCredentialsSection from "../EditBedrockCredentialsSection.vue";
import ModelRestrictionSection from "../ModelRestrictionSection.vue";
import PoolModeSection from "../PoolModeSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

function mountSection() {
  return mount(EditBedrockCredentialsSection, {
    props: {
      authMode: "sigv4",
      accessKeyId: "",
      secretAccessKey: "",
      sessionToken: "",
      apiKeyValue: "",
      region: "us-east-1",
      forceGlobal: false,
      modelRestrictionMode: "mapping",
      allowedModels: [],
      mappings: [{ from: "claude", to: "anthropic.claude" }],
      presetMappings: [
        {
          label: "Claude",
          from: "claude",
          to: "anthropic.claude",
        },
      ],
      mappingKey: (mapping: { from: string; to: string }) => mapping.from,
      poolModeEnabled: false,
      poolModeRetryCount: 2,
    },
  });
}

describe("EditBedrockCredentialsSection", () => {
  it("forwards credential, model restriction, and pool mode events", async () => {
    const wrapper = mountSection();

    expect(
      wrapper.findComponent(BedrockCredentialsSection).props(
        "allowAuthModeChange",
      ),
    ).toBe(false);

    await wrapper
      .findComponent(BedrockCredentialsSection)
      .vm.$emit("update:region", "eu-west-1");
    await wrapper
      .findComponent(ModelRestrictionSection)
      .vm.$emit("addPreset", "claude", "anthropic.claude");
    await wrapper
      .findComponent(PoolModeSection)
      .vm.$emit("update:retryCount", 3);

    expect(wrapper.emitted("update:region")).toEqual([["eu-west-1"]]);
    expect(wrapper.emitted("addPreset")).toEqual([
      ["claude", "anthropic.claude"],
    ]);
    expect(wrapper.emitted("update:poolModeRetryCount")).toEqual([[3]]);
  });
});
