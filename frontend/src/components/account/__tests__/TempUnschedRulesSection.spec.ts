import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import TempUnschedRulesSection from "../TempUnschedRulesSection.vue";
import type { TempUnschedRuleForm } from "../credentialsBuilder";

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, values?: Record<string, unknown>) =>
        values ? `${key}:${JSON.stringify(values)}` : key,
    }),
  };
});

const baseRules: TempUnschedRuleForm[] = [
  {
    error_code: 529,
    keywords: "overloaded",
    duration_minutes: 60,
    description: "Overload",
  },
  {
    error_code: 429,
    keywords: "rate limit",
    duration_minutes: 10,
    description: "Rate limit",
  },
];

const presets = [
  {
    label: "Preset overload",
    rule: baseRules[0],
  },
];

function mountSection(
  props: Partial<{
    enabled: boolean;
    rules: TempUnschedRuleForm[];
  }> = {},
) {
  return mount(TempUnschedRulesSection, {
    props: {
      enabled: false,
      presets,
      rules: [],
      ruleKey: (rule: TempUnschedRuleForm) =>
        `${rule.error_code}-${rule.keywords}`,
      ...props,
    },
  });
}

describe("TempUnschedRulesSection", () => {
  it("renders the switch header and emits enabled changes", async () => {
    const wrapper = mountSection();

    await wrapper.find("button").trigger("click");

    expect(wrapper.text()).toContain(
      "admin.accounts.tempUnschedulable.title",
    );
    expect(wrapper.text()).toContain("admin.accounts.tempUnschedulable.hint");
    expect(wrapper.text()).not.toContain(
      "admin.accounts.tempUnschedulable.notice",
    );
    expect(wrapper.emitted("update:enabled")).toEqual([[true]]);
  });

  it("emits add, move, remove, and field updates", async () => {
    const wrapper = mountSection({
      enabled: true,
      rules: baseRules,
    });
    const inputs = wrapper.findAll("input");
    const presetButton = wrapper
      .findAll("button")
      .find((button) => button.text().includes("Preset overload"));

    expect(presetButton).toBeTruthy();
    await presetButton!.trigger("click");
    await wrapper
      .findAll("button")
      .find((button) =>
        button.text().includes("admin.accounts.tempUnschedulable.addRule"),
      )!
      .trigger("click");
    await inputs[0].setValue("503");
    await inputs[1].setValue("30");
    await inputs[2].setValue("maintenance");
    await inputs[3].setValue("Maintenance window");
    await wrapper
      .findAll(".temp-unsched-rules-section__icon-button")
      .at(1)!
      .trigger("click");
    await wrapper
      .find(".temp-unsched-rules-section__icon-button--danger")
      .trigger("click");

    expect(wrapper.text()).toContain(
      "admin.accounts.tempUnschedulable.notice",
    );
    expect(wrapper.emitted("addRule")).toEqual([[presets[0].rule], []]);
    expect(wrapper.emitted("updateRule")).toEqual([
      [0, "error_code", 503],
      [0, "duration_minutes", 30],
      [0, "keywords", "maintenance"],
      [0, "description", "Maintenance window"],
    ]);
    expect(wrapper.emitted("moveRule")).toEqual([[0, 1]]);
    expect(wrapper.emitted("removeRule")).toEqual([[0]]);
  });
});
