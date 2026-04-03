package service

import (
	"context"

	"github.com/senran-N/sub2api/internal/domain"
	infraerrors "github.com/senran-N/sub2api/internal/pkg/errors"
)

// Error definitions for user attribute operations
var (
	ErrAttributeDefinitionNotFound = infraerrors.NotFound("ATTRIBUTE_DEFINITION_NOT_FOUND", "attribute definition not found")
	ErrAttributeKeyExists          = infraerrors.Conflict("ATTRIBUTE_KEY_EXISTS", "attribute key already exists")
	ErrInvalidAttributeType        = infraerrors.BadRequest("INVALID_ATTRIBUTE_TYPE", "invalid attribute type")
	ErrAttributeValidationFailed   = infraerrors.BadRequest("ATTRIBUTE_VALIDATION_FAILED", "attribute value validation failed")
)

type UserAttributeType = domain.UserAttributeType

const (
	AttributeTypeText        = domain.AttributeTypeText
	AttributeTypeTextarea    = domain.AttributeTypeTextarea
	AttributeTypeNumber      = domain.AttributeTypeNumber
	AttributeTypeEmail       = domain.AttributeTypeEmail
	AttributeTypeURL         = domain.AttributeTypeURL
	AttributeTypeDate        = domain.AttributeTypeDate
	AttributeTypeSelect      = domain.AttributeTypeSelect
	AttributeTypeMultiSelect = domain.AttributeTypeMultiSelect
)

type UserAttributeOption = domain.UserAttributeOption
type UserAttributeValidation = domain.UserAttributeValidation
type UserAttributeDefinition = domain.UserAttributeDefinition
type UserAttributeValue = domain.UserAttributeValue
type CreateAttributeDefinitionInput = domain.CreateAttributeDefinitionInput
type UpdateAttributeDefinitionInput = domain.UpdateAttributeDefinitionInput
type UpdateUserAttributeInput = domain.UpdateUserAttributeInput

// UserAttributeDefinitionRepository interface for attribute definition persistence
type UserAttributeDefinitionRepository interface {
	Create(ctx context.Context, def *UserAttributeDefinition) error
	GetByID(ctx context.Context, id int64) (*UserAttributeDefinition, error)
	GetByKey(ctx context.Context, key string) (*UserAttributeDefinition, error)
	Update(ctx context.Context, def *UserAttributeDefinition) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, enabledOnly bool) ([]UserAttributeDefinition, error)
	UpdateDisplayOrders(ctx context.Context, orders map[int64]int) error
	ExistsByKey(ctx context.Context, key string) (bool, error)
}

// UserAttributeValueRepository interface for user attribute value persistence
type UserAttributeValueRepository interface {
	GetByUserID(ctx context.Context, userID int64) ([]UserAttributeValue, error)
	GetByUserIDs(ctx context.Context, userIDs []int64) ([]UserAttributeValue, error)
	UpsertBatch(ctx context.Context, userID int64, values []UpdateUserAttributeInput) error
	DeleteByAttributeID(ctx context.Context, attributeID int64) error
	DeleteByUserID(ctx context.Context, userID int64) error
}
