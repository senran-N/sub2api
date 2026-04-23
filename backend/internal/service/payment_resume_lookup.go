package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	dbent "github.com/senran-N/sub2api/ent"
	infraerrors "github.com/senran-N/sub2api/internal/pkg/errors"
)

func (s *PaymentService) GetPublicOrderByResumeToken(ctx context.Context, token string) (*dbent.PaymentOrder, error) {
	claims, err := s.paymentResume().ParseToken(strings.TrimSpace(token))
	if err != nil {
		return nil, err
	}

	order, err := s.entClient.PaymentOrder.Get(ctx, claims.OrderID)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, infraerrors.NotFound("NOT_FOUND", "order not found")
		}
		return nil, fmt.Errorf("get order by resume token: %w", err)
	}
	if claims.UserID > 0 && order.UserID != claims.UserID {
		return nil, invalidResumeTokenMatchError()
	}

	orderProviderInstanceID := stringPtrValue(order.ProviderInstanceID)
	if claims.ProviderInstanceID != "" && orderProviderInstanceID != claims.ProviderInstanceID {
		return nil, invalidResumeTokenMatchError()
	}
	if claims.ProviderKey != "" {
		orderProviderKey, providerErr := s.lookupOrderProviderKey(ctx, orderProviderInstanceID)
		if providerErr != nil {
			return nil, providerErr
		}
		if !strings.EqualFold(orderProviderKey, claims.ProviderKey) {
			return nil, invalidResumeTokenMatchError()
		}
	}
	if claims.PaymentType != "" && NormalizeVisibleMethod(order.PaymentType) != NormalizeVisibleMethod(claims.PaymentType) {
		return nil, invalidResumeTokenMatchError()
	}

	if order.Status == OrderStatusPending || order.Status == OrderStatusExpired {
		result := s.checkPaid(ctx, order)
		if result == checkPaidResultAlreadyPaid {
			order, err = s.entClient.PaymentOrder.Get(ctx, order.ID)
			if err != nil {
				return nil, fmt.Errorf("reload order by resume token: %w", err)
			}
		}
	}

	return order, nil
}

func (s *PaymentService) ParseWeChatPaymentResumeToken(token string) (*WeChatPaymentResumeClaims, error) {
	return s.paymentResume().ParseWeChatPaymentResumeToken(strings.TrimSpace(token))
}

func (s *PaymentService) lookupOrderProviderKey(ctx context.Context, providerInstanceID string) (string, error) {
	providerInstanceID = strings.TrimSpace(providerInstanceID)
	if providerInstanceID == "" {
		return "", invalidResumeTokenMatchError()
	}
	instanceID, err := strconv.ParseInt(providerInstanceID, 10, 64)
	if err != nil {
		return "", invalidResumeTokenMatchError()
	}
	instance, err := s.entClient.PaymentProviderInstance.Get(ctx, instanceID)
	if err != nil {
		if dbent.IsNotFound(err) {
			return "", invalidResumeTokenMatchError()
		}
		return "", fmt.Errorf("lookup provider instance for resume token: %w", err)
	}
	return strings.TrimSpace(instance.ProviderKey), nil
}

func invalidResumeTokenMatchError() error {
	return infraerrors.BadRequest("INVALID_RESUME_TOKEN", "resume token does not match the payment order")
}

func stringPtrValue(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}
