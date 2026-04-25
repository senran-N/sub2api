import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import AntigravityAccountTypeSection from "../AntigravityAccountTypeSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

function mountSection(accountType: "oauth" | "upstream" = "oauth") {
  return mount(AntigravityAccountTypeSection, {
    props: {
      accountType,
    },
    global: {
      stubs: {
        Icon: true,
      },
    },
  });
}

describe("AntigravityAccountTypeSection", () => {
  it("emits account type changes without owning parent state", async () => {
    const wrapper = mountSection();
    const buttons = wrapper.findAll("button");

    await buttons[1].trigger("click");
    await buttons[0].trigger("click");

    expect(wrapper.emitted("update:accountType")).toEqual([
      ["upstream"],
      ["oauth"],
    ]);
  });

  it("renders Antigravity OAuth and API key options", () => {
    const wrapper = mountSection("upstream");

    expect(wrapper.text()).toContain("OAuth");
    expect(wrapper.text()).toContain("API Key");
    expect(wrapper.text()).toContain("admin.accounts.types.antigravityOauth");
    expect(wrapper.text()).toContain("admin.accounts.types.antigravityApikey");
  });
});
