import { describe, expect, it, vi } from "vitest";
import { useEditAccountModelRestrictions } from "../useEditAccountModelRestrictions";

describe("useEditAccountModelRestrictions", () => {
  it("hydrates whitelist and mapping model restriction state", () => {
    const restrictions = useEditAccountModelRestrictions();

    restrictions.applyModelRestrictionState({
      "gpt-4.1": "gpt-4.1",
      "gpt-4o": "gpt-4o",
    });

    expect(restrictions.modelRestrictionMode.value).toBe("whitelist");
    expect(restrictions.allowedModels.value).toEqual(["gpt-4.1", "gpt-4o"]);
    expect(restrictions.modelMappings.value).toEqual([]);

    restrictions.applyModelRestrictionState({
      "gpt-4.1": "provider-gpt-4.1",
    });

    expect(restrictions.modelRestrictionMode.value).toBe("mapping");
    expect(restrictions.allowedModels.value).toEqual([]);
    expect(restrictions.modelMappings.value).toEqual([
      { from: "gpt-4.1", to: "provider-gpt-4.1" },
    ]);
  });

  it("hydrates Antigravity mappings from credentials", () => {
    const restrictions = useEditAccountModelRestrictions();

    restrictions.syncAntigravityModelRestrictionState({
      model_whitelist: ["auto", "pro"],
    });

    expect(restrictions.antigravityModelMappings.value).toEqual([
      { from: "auto", to: "auto" },
      { from: "pro", to: "pro" },
    ]);

    restrictions.resetAntigravityModelRestrictionState();
    expect(restrictions.antigravityModelMappings.value).toEqual([]);
  });

  it("updates mappings and reports duplicate presets", () => {
    const onMappingExists = vi.fn();
    const restrictions = useEditAccountModelRestrictions({ onMappingExists });

    restrictions.addPresetMapping("from-model", "to-model");
    restrictions.addPresetMapping("from-model", "to-model");
    restrictions.updateModelMapping(0, "to", "updated-model");
    restrictions.addAntigravityModelMapping();
    restrictions.updateAntigravityModelMapping(0, "from", "auto");
    restrictions.updateAntigravityModelMapping(0, "to", "auto");

    expect(restrictions.modelMappings.value).toEqual([
      { from: "from-model", to: "updated-model" },
    ]);
    expect(onMappingExists).toHaveBeenCalledWith("from-model");
    expect(restrictions.antigravityModelMappings.value).toEqual([
      { from: "auto", to: "auto" },
    ]);

    restrictions.removeModelMapping(0);
    restrictions.removeAntigravityModelMapping(0);
    expect(restrictions.modelMappings.value).toEqual([]);
    expect(restrictions.antigravityModelMappings.value).toEqual([]);
  });
});
