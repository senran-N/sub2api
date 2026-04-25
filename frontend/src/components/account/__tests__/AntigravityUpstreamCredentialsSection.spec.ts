import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import AntigravityUpstreamCredentialsSection from "../AntigravityUpstreamCredentialsSection.vue";

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
  return mount(AntigravityUpstreamCredentialsSection, {
    props: {
      baseUrl: "https://cloudcode-pa.googleapis.com",
      apiKey: "sk-old",
    },
  });
}

describe("AntigravityUpstreamCredentialsSection", () => {
  it("emits credential field updates without owning parent state", async () => {
    const wrapper = mountSection();
    const inputs = wrapper.findAll("input");

    await inputs[0].setValue("https://example.test");
    await inputs[1].setValue("sk-new");

    expect(wrapper.emitted("update:baseUrl")).toEqual([
      ["https://example.test"],
    ]);
    expect(wrapper.emitted("update:apiKey")).toEqual([["sk-new"]]);
  });

  it("renders the upstream credential labels and hints", () => {
    const wrapper = mountSection();

    expect(wrapper.text()).toContain("admin.accounts.upstream.baseUrl");
    expect(wrapper.text()).toContain("admin.accounts.upstream.baseUrlHint");
    expect(wrapper.text()).toContain("admin.accounts.upstream.apiKey");
    expect(wrapper.text()).toContain("admin.accounts.upstream.apiKeyHint");
  });
});
