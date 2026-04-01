package gemini

import "testing"

func TestDefaultModels_ContainsImageModels(t *testing.T) {
	t.Parallel()

	models := DefaultModels()
	byName := make(map[string]Model, len(models))
	for _, model := range models {
		byName[model.Name] = model
	}

	required := []string{
		"models/gemini-2.5-flash-image",
		"models/gemini-3.1-flash-image",
		"models/gemini-3.1-pro-preview-customtools",
	}

	for _, name := range required {
		model, ok := byName[name]
		if !ok {
			t.Fatalf("expected fallback model %q to exist", name)
		}
		if len(model.SupportedGenerationMethods) == 0 {
			t.Fatalf("expected fallback model %q to advertise generation methods", name)
		}
	}
}

func TestHasFallbackModel(t *testing.T) {
	t.Parallel()

	if !HasFallbackModel("gemini-3.1-pro-preview-customtools") {
		t.Fatalf("expected customtools model to be recognized without models/ prefix")
	}
	if !HasFallbackModel("models/gemini-3.1-pro-preview-customtools") {
		t.Fatalf("expected customtools model to be recognized with models/ prefix")
	}
	if HasFallbackModel("gemini-unknown-preview") {
		t.Fatalf("did not expect unknown model to be recognized as fallback")
	}
}
