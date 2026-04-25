import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import CreateAccountBasicInfoSection from "../CreateAccountBasicInfoSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

describe("CreateAccountBasicInfoSection", () => {
  it("emits name and notes updates", async () => {
    const wrapper = mount(CreateAccountBasicInfoSection, {
      props: {
        name: "",
        notes: "",
        nameLabel: "Name",
        namePlaceholder: "Name placeholder",
        nameHint: "Name hint",
        nameRequired: true,
      },
    });
    const input = wrapper.find("input");
    const textarea = wrapper.find("textarea");

    await input.setValue("Primary account");
    await textarea.setValue("Internal note");

    expect(wrapper.text()).toContain("Name");
    expect(wrapper.text()).toContain("Name hint");
    expect(input.attributes("required")).toBeDefined();
    expect(wrapper.emitted("update:name")).toEqual([["Primary account"]]);
    expect(wrapper.emitted("update:notes")).toEqual([["Internal note"]]);
  });
});
