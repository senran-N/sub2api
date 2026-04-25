import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import AnthropicQuotaControlsSection from "../AnthropicQuotaControlsSection.vue";
import CacheTtlOverrideSection from "../CacheTtlOverrideSection.vue";
import RpmLimitControlSection from "../RpmLimitControlSection.vue";
import TlsFingerprintControlSection from "../TlsFingerprintControlSection.vue";
import WindowCostControlSection from "../WindowCostControlSection.vue";

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
  return mount(AnthropicQuotaControlsSection, {
    props: {
      windowCostEnabled: false,
      windowCostLimit: null,
      windowCostStickyReserve: null,
      sessionLimitEnabled: false,
      maxSessions: null,
      sessionIdleTimeout: null,
      rpmLimitEnabled: false,
      baseRpm: null,
      rpmStrategy: "tiered",
      rpmStickyBuffer: null,
      userMsgQueueMode: "",
      userMsgQueueModeOptions: [
        { value: "", label: "Off" },
        { value: "throttle", label: "Throttle" },
      ],
      tlsFingerprintEnabled: false,
      tlsFingerprintProfileId: null,
      tlsFingerprintProfiles: [{ id: 9, name: "Chrome" }],
      sessionIdMaskingEnabled: false,
      cacheTtlOverrideEnabled: false,
      cacheTtlOverrideTarget: "5m",
      customBaseUrlEnabled: false,
      customBaseUrl: "",
    },
  });
}

describe("AnthropicQuotaControlsSection", () => {
  it("renders quota controls and forwards model updates", async () => {
    const wrapper = mountSection();

    expect(wrapper.text()).toContain("admin.accounts.quotaControl.title");
    expect(wrapper.text()).toContain("admin.accounts.quotaControl.hint");
    expect(
      wrapper.findComponent(RpmLimitControlSection).props(
        "userMsgQueueModeOptions",
      ),
    ).toEqual([
      { value: "", label: "Off" },
      { value: "throttle", label: "Throttle" },
    ]);
    expect(
      wrapper.findComponent(TlsFingerprintControlSection).props("profiles"),
    ).toEqual([{ id: 9, name: "Chrome" }]);

    await wrapper
      .findComponent(WindowCostControlSection)
      .vm.$emit("update:enabled", true);
    await wrapper
      .findComponent(CacheTtlOverrideSection)
      .vm.$emit("update:target", "1h");

    expect(wrapper.emitted("update:windowCostEnabled")).toEqual([[true]]);
    expect(wrapper.emitted("update:cacheTtlOverrideTarget")).toEqual([["1h"]]);
  });
});
