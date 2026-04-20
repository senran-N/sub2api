package grok

type Model struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	Created     int64  `json:"created"`
	OwnedBy     string `json:"owned_by"`
	Type        string `json:"type"`
	DisplayName string `json:"display_name"`
}

var DefaultModels = EnabledModels()

func DefaultModelIDs() []string {
	return EnabledModelIDs()
}

const DefaultTestModel = "grok-4.20-auto"
