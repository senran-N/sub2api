package domain

import "time"

type UserAttributeType string

const (
	AttributeTypeText        UserAttributeType = "text"
	AttributeTypeTextarea    UserAttributeType = "textarea"
	AttributeTypeNumber      UserAttributeType = "number"
	AttributeTypeEmail       UserAttributeType = "email"
	AttributeTypeURL         UserAttributeType = "url"
	AttributeTypeDate        UserAttributeType = "date"
	AttributeTypeSelect      UserAttributeType = "select"
	AttributeTypeMultiSelect UserAttributeType = "multi_select"
)

type UserAttributeOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type UserAttributeValidation struct {
	MinLength *int    `json:"min_length,omitempty"`
	MaxLength *int    `json:"max_length,omitempty"`
	Min       *int    `json:"min,omitempty"`
	Max       *int    `json:"max,omitempty"`
	Pattern   *string `json:"pattern,omitempty"`
	Message   *string `json:"message,omitempty"`
}

type UserAttributeDefinition struct {
	ID           int64
	Key          string
	Name         string
	Description  string
	Type         UserAttributeType
	Options      []UserAttributeOption
	Required     bool
	Validation   UserAttributeValidation
	Placeholder  string
	DisplayOrder int
	Enabled      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserAttributeValue struct {
	ID          int64
	UserID      int64
	AttributeID int64
	Value       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateAttributeDefinitionInput struct {
	Key         string
	Name        string
	Description string
	Type        UserAttributeType
	Options     []UserAttributeOption
	Required    bool
	Validation  UserAttributeValidation
	Placeholder string
	Enabled     bool
}

type UpdateAttributeDefinitionInput struct {
	Name        *string
	Description *string
	Type        *UserAttributeType
	Options     *[]UserAttributeOption
	Required    *bool
	Validation  *UserAttributeValidation
	Placeholder *string
	Enabled     *bool
}

type UpdateUserAttributeInput struct {
	AttributeID int64
	Value       string
}
