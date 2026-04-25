import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import CacheTtlOverrideSection from "../CacheTtlOverrideSection.vue";
import CustomBaseUrlControlSection from "../CustomBaseUrlControlSection.vue";
import SessionIdMaskingControlSection from "../SessionIdMaskingControlSection.vue";
import TlsFingerprintControlSection from "../TlsFingerprintControlSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

describe("Quota simple control sections", () => {
  it("emits TLS fingerprint profile selection", async () => {
    const wrapper = mount(TlsFingerprintControlSection, {
      props: {
        enabled: true,
        profileId: null,
        profiles: [{ id: 7, name: "Chrome" }],
      },
    });

    await wrapper.find("select").setValue("7");

    expect(wrapper.text()).toContain(
      "admin.accounts.quotaControl.tlsFingerprint.label",
    );
    expect(wrapper.text()).toContain("Chrome");
    expect(wrapper.emitted("update:profileId")).toEqual([[7]]);
  });

  it("emits session masking enabled changes", async () => {
    const wrapper = mount(SessionIdMaskingControlSection, {
      props: {
        enabled: false,
      },
    });

    await wrapper.find("button").trigger("click");

    expect(wrapper.text()).toContain(
      "admin.accounts.quotaControl.sessionIdMasking.label",
    );
    expect(wrapper.emitted("update:enabled")).toEqual([[true]]);
  });

  it("emits cache TTL target changes", async () => {
    const wrapper = mount(CacheTtlOverrideSection, {
      props: {
        enabled: true,
        target: "5m",
      },
    });

    await wrapper.find("select").setValue("1h");

    expect(wrapper.text()).toContain(
      "admin.accounts.quotaControl.cacheTTLOverride.target",
    );
    expect(wrapper.emitted("update:target")).toEqual([["1h"]]);
  });

  it("emits custom base URL changes", async () => {
    const wrapper = mount(CustomBaseUrlControlSection, {
      props: {
        enabled: true,
        baseUrl: "",
      },
    });

    await wrapper.find("input").setValue("https://relay.example.com");

    expect(wrapper.text()).toContain(
      "admin.accounts.quotaControl.customBaseUrl.label",
    );
    expect(wrapper.emitted("update:baseUrl")).toEqual([
      ["https://relay.example.com"],
    ]);
  });
});
