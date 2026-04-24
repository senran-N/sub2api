//go:build integration

package repository

import (
	"context"
	"testing"

	dbent "github.com/senran-N/sub2api/ent"
	"github.com/senran-N/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUserAttributeValueRepository_UpsertBatch_DeduplicatesAttributeIDs(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	client := tx.Client()
	repo := NewUserAttributeValueRepository(client)

	user := mustCreateUser(t, client, &service.User{Email: "attr-upsert@example.com"})
	defA := mustCreateUserAttributeDefinition(t, client, "department")
	defB := mustCreateUserAttributeDefinition(t, client, "region")

	err := repo.UpsertBatch(ctx, user.ID, []service.UpdateUserAttributeInput{
		{AttributeID: defA.ID, Value: "ops"},
		{AttributeID: defB.ID, Value: "apac"},
		{AttributeID: defA.ID, Value: "finance"},
	})
	require.NoError(t, err)

	values, err := repo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, values, 2)

	valueByAttributeID := make(map[int64]string, len(values))
	for _, value := range values {
		valueByAttributeID[value.AttributeID] = value.Value
	}
	require.Equal(t, "finance", valueByAttributeID[defA.ID])
	require.Equal(t, "apac", valueByAttributeID[defB.ID])
}

func TestUserAttributeDefinitionRepository_UpdateDisplayOrders_MissingIDRollsBack(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	client := tx.Client()
	repo := NewUserAttributeDefinitionRepository(client)

	defA := mustCreateUserAttributeDefinition(t, client, "attr-order-a")
	defB := mustCreateUserAttributeDefinition(t, client, "attr-order-b")

	err := repo.UpdateDisplayOrders(ctx, map[int64]int{
		defA.ID:   10,
		999999999: 20,
		defB.ID:   30,
	})
	require.ErrorIs(t, err, service.ErrAttributeDefinitionNotFound)

	gotA, err := repo.GetByID(ctx, defA.ID)
	require.NoError(t, err)
	require.Equal(t, defA.DisplayOrder, gotA.DisplayOrder)

	gotB, err := repo.GetByID(ctx, defB.ID)
	require.NoError(t, err)
	require.Equal(t, defB.DisplayOrder, gotB.DisplayOrder)
}

func mustCreateUserAttributeDefinition(t *testing.T, client *dbent.Client, key string) *service.UserAttributeDefinition {
	t.Helper()
	ctx := context.Background()

	created, err := client.UserAttributeDefinition.Create().
		SetKey(key).
		SetName(key).
		SetType(string(service.AttributeTypeText)).
		SetOptions([]map[string]any{}).
		SetValidation(map[string]any{}).
		SetEnabled(true).
		Save(ctx)
	require.NoError(t, err, "create user attribute definition")

	return &service.UserAttributeDefinition{
		ID:           created.ID,
		Key:          created.Key,
		Name:         created.Name,
		Type:         service.UserAttributeType(created.Type),
		DisplayOrder: created.DisplayOrder,
		Enabled:      created.Enabled,
		CreatedAt:    created.CreatedAt,
		UpdatedAt:    created.UpdatedAt,
	}
}
