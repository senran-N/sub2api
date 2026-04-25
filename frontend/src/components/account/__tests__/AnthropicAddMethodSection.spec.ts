import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import AnthropicAddMethodSection from "../AnthropicAddMethodSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

describe("AnthropicAddMethodSection", () => {
  it("emits add method changes", async () => {
    const wrapper = mount(AnthropicAddMethodSection, {
      props: {
        modelValue: "oauth",
      },
    });

    await wrapper.find('input[value="setup-token"]').setValue();

    expect(wrapper.text()).toContain("admin.accounts.addMethod");
    expect(wrapper.text()).toContain("admin.accounts.setupTokenLongLived");
    expect(wrapper.emitted("update:modelValue")).toEqual([["setup-token"]]);
  });
});
