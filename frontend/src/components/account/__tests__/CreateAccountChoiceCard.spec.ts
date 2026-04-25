import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import CreateAccountChoiceCard from "../CreateAccountChoiceCard.vue";

function mountCard(selected = false) {
  return mount(CreateAccountChoiceCard, {
    props: {
      selected,
      tone: "purple",
      icon: "key",
      title: "OAuth",
      description: "Provider OAuth account",
    },
    global: {
      stubs: {
        Icon: true,
      },
    },
  });
}

describe("CreateAccountChoiceCard", () => {
  it("emits select when clicked", async () => {
    const wrapper = mountCard();

    await wrapper.find("button").trigger("click");

    expect(wrapper.emitted("select")).toEqual([[]]);
  });

  it("renders selected and idle states with shared card classes", () => {
    const selected = mountCard(true);
    const idle = mountCard(false);

    expect(selected.find("button").classes()).toContain(
      "create-account-choice-card--purple",
    );
    expect(idle.find("button").classes()).toContain(
      "create-account-choice-card--idle",
    );
  });

  it("renders title, description, and meta slot", () => {
    const wrapper = mount(CreateAccountChoiceCard, {
      props: {
        selected: true,
        tone: "blue",
        icon: "cloud",
        title: "API Key",
        description: "Provider upstream account",
      },
      slots: {
        meta: "<span>High concurrency</span>",
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    });

    expect(wrapper.text()).toContain("API Key");
    expect(wrapper.text()).toContain("Provider upstream account");
    expect(wrapper.text()).toContain("High concurrency");
  });
});
