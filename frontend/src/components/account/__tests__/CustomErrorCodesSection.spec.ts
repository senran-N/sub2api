import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import CustomErrorCodesSection from "../CustomErrorCodesSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

function mountSection(
  props: Partial<{
    enabled: boolean;
    selectedCodes: number[];
    inputValue: number | null;
  }> = {},
) {
  return mount(CustomErrorCodesSection, {
    props: {
      enabled: false,
      selectedCodes: [],
      inputValue: null,
      ...props,
    },
  });
}

describe("CustomErrorCodesSection", () => {
  it("renders the switch header and emits enabled changes", async () => {
    const wrapper = mountSection();

    await wrapper.find("button").trigger("click");

    expect(wrapper.text()).toContain("admin.accounts.customErrorCodes");
    expect(wrapper.text()).toContain("admin.accounts.customErrorCodesHint");
    expect(wrapper.text()).not.toContain(
      "admin.accounts.customErrorCodesWarning",
    );
    expect(wrapper.emitted("update:enabled")).toEqual([[true]]);
  });

  it("emits code actions while leaving validation to the parent", async () => {
    const wrapper = mountSection({
      enabled: true,
      selectedCodes: [503, 401],
    });

    const rateLimitButton = wrapper
      .findAll("button")
      .find((button) => button.text().includes("429 Rate Limit"));

    await wrapper.find("input").setValue("418");
    await wrapper.find("input").trigger("keyup.enter");
    expect(rateLimitButton).toBeTruthy();
    await rateLimitButton!.trigger("click");
    await wrapper
      .find(".custom-error-codes-section__remove-button")
      .trigger("click");

    expect(wrapper.text()).toContain("admin.accounts.customErrorCodesWarning");
    expect(wrapper.text()).toContain("401");
    expect(wrapper.text()).toContain("503");
    expect(wrapper.emitted("update:inputValue")).toEqual([[418]]);
    expect(wrapper.emitted("addCode")).toEqual([[]]);
    expect(wrapper.emitted("toggleCode")).toEqual([[429]]);
    expect(wrapper.emitted("removeCode")).toEqual([[401]]);
  });
});
