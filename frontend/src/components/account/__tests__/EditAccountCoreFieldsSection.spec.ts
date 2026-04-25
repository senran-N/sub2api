import { describe, expect, it, vi } from "vitest";
import { defineComponent } from "vue";
import { mount } from "@vue/test-utils";
import EditAccountCoreFieldsSection from "../EditAccountCoreFieldsSection.vue";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  };
});

const SelectStub = defineComponent({
  name: "SelectStub",
  props: {
    modelValue: {
      type: String,
      default: "",
    },
  },
  emits: ["update:modelValue"],
  template: `
    <button
      type="button"
      data-testid="status-select"
      :data-value="modelValue"
      @click="$emit('update:modelValue', 'error')"
    />
  `,
});

const ProxySelectorStub = defineComponent({
  name: "ProxySelector",
  emits: ["update:modelValue"],
  template: `
    <button
      type="button"
      data-testid="proxy-selector"
      @click="$emit('update:modelValue', 7)"
    />
  `,
});

const GroupSelectorStub = defineComponent({
  name: "GroupSelector",
  emits: ["update:modelValue"],
  template: `
    <button
      type="button"
      data-testid="group-selector"
      @click="$emit('update:modelValue', [2, 3])"
    />
  `,
});

function mountSection() {
  return mount(EditAccountCoreFieldsSection, {
    props: {
      allowOverages: false,
      concurrency: 2,
      expiresAt: "",
      groupIds: [],
      groups: [],
      loadFactor: null,
      mixedScheduling: true,
      name: "Account",
      notes: "",
      platform: "antigravity",
      priority: 1,
      proxies: [],
      proxyId: null,
      rateMultiplier: 1,
      simpleMode: false,
      status: "active",
      statusOptions: [
        { value: "active", label: "Active" },
        { value: "inactive", label: "Inactive" },
        { value: "error", label: "Error" },
      ],
    },
    global: {
      stubs: {
        Select: SelectStub,
        ProxySelector: ProxySelectorStub,
        GroupSelector: GroupSelectorStub,
      },
    },
  });
}

describe("EditAccountCoreFieldsSection", () => {
  it("emits core field, scheduling, status, and group updates", async () => {
    const wrapper = mountSection();

    await wrapper.get('[data-tour="edit-account-form-name"]').setValue("Next");
    await wrapper.get("textarea").setValue("note");
    await wrapper.get('[data-testid="proxy-selector"]').trigger("click");
    await wrapper.get('[data-testid="status-select"]').trigger("click");
    await wrapper.get('[data-testid="group-selector"]').trigger("click");

    const numberInputs = wrapper.findAll('input[type="number"]');
    await numberInputs[0].setValue("0");
    await numberInputs[1].setValue("0");
    await numberInputs[2].setValue("5");
    await numberInputs[3].setValue("1.5");

    const checkboxes = wrapper.findAll('input[type="checkbox"]');
    expect(checkboxes[0].attributes("disabled")).toBeDefined();
    await checkboxes[1].setValue(true);

    expect(wrapper.emitted("update:name")).toEqual([["Next"]]);
    expect(wrapper.emitted("update:notes")).toEqual([["note"]]);
    expect(wrapper.emitted("update:proxyId")).toEqual([[7]]);
    expect(wrapper.emitted("update:status")).toEqual([["error"]]);
    expect(wrapper.emitted("update:groupIds")).toEqual([[[2, 3]]]);
    expect(wrapper.emitted("update:concurrency")).toEqual([[1]]);
    expect(wrapper.emitted("update:loadFactor")).toEqual([[null]]);
    expect(wrapper.emitted("update:priority")).toEqual([[5]]);
    expect(wrapper.emitted("update:rateMultiplier")).toEqual([[1.5]]);
    expect(wrapper.emitted("update:allowOverages")).toEqual([[true]]);
  });
});
