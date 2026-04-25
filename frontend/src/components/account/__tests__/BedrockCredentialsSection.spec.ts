import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import BedrockCredentialsSection from "../BedrockCredentialsSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

function mountSection(authMode: "sigv4" | "apikey" = "sigv4") {
  return mount(BedrockCredentialsSection, {
    props: {
      authMode,
      accessKeyId: "AKIAOLD",
      secretAccessKey: "secret-old",
      sessionToken: "session-old",
      apiKeyValue: "api-old",
      region: "us-east-1",
      forceGlobal: false,
    },
  });
}

describe("BedrockCredentialsSection", () => {
  it("renders SigV4 fields and emits credential updates", async () => {
    const wrapper = mountSection("sigv4");
    const inputs = wrapper.findAll("input");

    await inputs[2].setValue("AKIANEW");
    await inputs[3].setValue("secret-new");
    await inputs[4].setValue("session-new");

    expect(wrapper.text()).toContain("admin.accounts.bedrockAuthMode");
    expect(wrapper.text()).toContain("admin.accounts.bedrockAccessKeyId");
    expect(wrapper.text()).toContain("admin.accounts.bedrockSecretAccessKey");
    expect(wrapper.text()).toContain("admin.accounts.bedrockSessionToken");
    expect(wrapper.emitted("update:accessKeyId")).toEqual([["AKIANEW"]]);
    expect(wrapper.emitted("update:secretAccessKey")).toEqual([
      ["secret-new"],
    ]);
    expect(wrapper.emitted("update:sessionToken")).toEqual([["session-new"]]);
  });

  it("renders API key mode and emits API key updates", async () => {
    const wrapper = mountSection("apikey");
    const apiKeyInput = wrapper.find('input[type="password"]');

    await apiKeyInput.setValue("api-new");

    expect(wrapper.text()).toContain("admin.accounts.bedrockApiKeyInput");
    expect(wrapper.emitted("update:apiKeyValue")).toEqual([["api-new"]]);
  });

  it("emits auth mode, region, and force-global changes", async () => {
    const wrapper = mountSection("sigv4");

    await wrapper.find('input[value="apikey"]').setValue();
    await wrapper.find("select").setValue("eu-west-1");
    await wrapper.find('input[type="checkbox"]').setValue(true);

    expect(wrapper.emitted("update:authMode")).toEqual([["apikey"]]);
    expect(wrapper.emitted("update:region")).toEqual([["eu-west-1"]]);
    expect(wrapper.emitted("update:forceGlobal")).toEqual([[true]]);
  });

  it("supports edit mode without auth switching or required secrets", async () => {
    const wrapper = mount(BedrockCredentialsSection, {
      props: {
        authMode: "apikey",
        accessKeyId: "",
        secretAccessKey: "",
        sessionToken: "",
        apiKeyValue: "",
        region: "custom-region-1",
        forceGlobal: false,
        allowAuthModeChange: false,
        apiKeyHintKey: "admin.accounts.bedrockApiKeyLeaveEmpty",
        apiKeyPlaceholderKey: "admin.accounts.bedrockApiKeyLeaveEmpty",
        credentialsRequired: false,
        regionControl: "input",
      },
    });

    await wrapper.find('input[type="text"]').setValue("custom-region-2");

    expect(wrapper.text()).not.toContain("admin.accounts.bedrockAuthMode");
    expect(wrapper.find('input[type="password"]').attributes("required")).toBe(
      undefined,
    );
    expect(wrapper.text()).toContain("admin.accounts.bedrockApiKeyLeaveEmpty");
    expect(wrapper.emitted("update:region")).toEqual([["custom-region-2"]]);
  });
});
