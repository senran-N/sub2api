package service

import (
	"context"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type affiliateServiceTestSettingsRepo struct {
	values map[string]string
}

func (r *affiliateServiceTestSettingsRepo) Get(ctx context.Context, key string) (*Setting, error) {
	value, err := r.GetValue(ctx, key)
	if err != nil {
		return nil, err
	}
	return &Setting{Key: key, Value: value}, nil
}

func (r *affiliateServiceTestSettingsRepo) GetValue(_ context.Context, key string) (string, error) {
	if r != nil {
		if value, ok := r.values[key]; ok {
			return value, nil
		}
	}
	return "", ErrSettingNotFound
}

func (r *affiliateServiceTestSettingsRepo) Set(_ context.Context, key, value string) error {
	if r.values == nil {
		r.values = make(map[string]string)
	}
	r.values[key] = value
	return nil
}

func (r *affiliateServiceTestSettingsRepo) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := r.values[key]; ok {
			result[key] = value
		}
	}
	return result, nil
}

func (r *affiliateServiceTestSettingsRepo) SetMultiple(_ context.Context, settings map[string]string) error {
	if r.values == nil {
		r.values = make(map[string]string)
	}
	for key, value := range settings {
		r.values[key] = value
	}
	return nil
}

func (r *affiliateServiceTestSettingsRepo) GetAll(_ context.Context) (map[string]string, error) {
	result := make(map[string]string, len(r.values))
	for key, value := range r.values {
		result[key] = value
	}
	return result, nil
}

func (r *affiliateServiceTestSettingsRepo) Delete(_ context.Context, key string) error {
	delete(r.values, key)
	return nil
}

type affiliateServiceTestRepo struct {
	summaries        map[int64]*AffiliateSummary
	codeToUserID     map[string]int64
	accruedByPair    map[[2]int64]float64
	lastAccrueAmount float64
	lastFreezeHours  int
}

func newAffiliateServiceTestRepo() *affiliateServiceTestRepo {
	return &affiliateServiceTestRepo{
		summaries:     make(map[int64]*AffiliateSummary),
		codeToUserID:  make(map[string]int64),
		accruedByPair: make(map[[2]int64]float64),
	}
}

func (r *affiliateServiceTestRepo) put(summary AffiliateSummary) {
	clone := summary
	if clone.CreatedAt.IsZero() {
		clone.CreatedAt = time.Now()
	}
	clone.UpdatedAt = clone.CreatedAt
	r.summaries[clone.UserID] = &clone
	r.codeToUserID[strings.ToUpper(clone.AffCode)] = clone.UserID
}

func (r *affiliateServiceTestRepo) EnsureUserAffiliate(_ context.Context, userID int64) (*AffiliateSummary, error) {
	if summary, ok := r.summaries[userID]; ok {
		return cloneAffiliateSummary(summary), nil
	}
	r.put(AffiliateSummary{UserID: userID, AffCode: "USER" + strings.TrimPrefix(strings.TrimSpace(strings.ReplaceAll(time.Now().Format("150405.000"), ".", "")), "0")})
	return cloneAffiliateSummary(r.summaries[userID]), nil
}

func (r *affiliateServiceTestRepo) GetAffiliateByCode(_ context.Context, code string) (*AffiliateSummary, error) {
	userID, ok := r.codeToUserID[strings.ToUpper(code)]
	if !ok {
		return nil, ErrAffiliateProfileNotFound
	}
	return cloneAffiliateSummary(r.summaries[userID]), nil
}

func (r *affiliateServiceTestRepo) BindInviter(_ context.Context, userID, inviterID int64) (bool, error) {
	summary := r.summaries[userID]
	if summary.InviterID != nil {
		return false, nil
	}
	summary.InviterID = &inviterID
	if inviter := r.summaries[inviterID]; inviter != nil {
		inviter.AffCount++
	}
	return true, nil
}

func (r *affiliateServiceTestRepo) AccrueQuota(_ context.Context, inviterID, inviteeUserID int64, amount float64, freezeHours int) (bool, error) {
	r.lastAccrueAmount = amount
	r.lastFreezeHours = freezeHours
	pair := [2]int64{inviterID, inviteeUserID}
	r.accruedByPair[pair] += amount
	if summary := r.summaries[inviterID]; summary != nil {
		if freezeHours > 0 {
			summary.AffFrozenQuota += amount
		} else {
			summary.AffQuota += amount
		}
		summary.AffHistoryQuota += amount
	}
	return true, nil
}

func (r *affiliateServiceTestRepo) GetAccruedRebateFromInvitee(_ context.Context, inviterID, inviteeUserID int64) (float64, error) {
	return r.accruedByPair[[2]int64{inviterID, inviteeUserID}], nil
}

func (r *affiliateServiceTestRepo) ThawFrozenQuota(_ context.Context, _ int64) (float64, error) {
	return 0, nil
}

func (r *affiliateServiceTestRepo) TransferQuotaToBalance(_ context.Context, userID int64) (float64, float64, error) {
	summary := r.summaries[userID]
	if summary == nil || summary.AffQuota <= 0 {
		return 0, 0, nil
	}
	transferred := summary.AffQuota
	summary.AffQuota = 0
	return transferred, 100 + transferred, nil
}

func (r *affiliateServiceTestRepo) ListInvitees(_ context.Context, inviterID int64, limit int) ([]AffiliateInvitee, error) {
	result := make([]AffiliateInvitee, 0)
	for pair, amount := range r.accruedByPair {
		if pair[0] == inviterID {
			result = append(result, AffiliateInvitee{UserID: pair[1], Email: "invitee@example.com", TotalRebate: amount})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].UserID < result[j].UserID })
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *affiliateServiceTestRepo) UpdateUserAffCode(_ context.Context, userID int64, newCode string) error {
	summary := r.summaries[userID]
	if summary == nil {
		return ErrAffiliateProfileNotFound
	}
	delete(r.codeToUserID, strings.ToUpper(summary.AffCode))
	summary.AffCode = strings.ToUpper(newCode)
	summary.AffCodeCustom = true
	r.codeToUserID[summary.AffCode] = userID
	return nil
}

func (r *affiliateServiceTestRepo) ResetUserAffCode(_ context.Context, userID int64) (string, error) {
	code := "RESET"
	return code, r.UpdateUserAffCode(context.Background(), userID, code)
}

func (r *affiliateServiceTestRepo) SetUserRebateRate(_ context.Context, userID int64, ratePercent *float64) error {
	if summary := r.summaries[userID]; summary != nil {
		summary.AffRebateRatePercent = ratePercent
	}
	return nil
}

func (r *affiliateServiceTestRepo) BatchSetUserRebateRate(ctx context.Context, userIDs []int64, ratePercent *float64) error {
	for _, userID := range userIDs {
		if err := r.SetUserRebateRate(ctx, userID, ratePercent); err != nil {
			return err
		}
	}
	return nil
}

func (r *affiliateServiceTestRepo) ListUsersWithCustomSettings(_ context.Context, _ AffiliateAdminFilter) ([]AffiliateAdminEntry, int64, error) {
	return nil, 0, nil
}

func cloneAffiliateSummary(summary *AffiliateSummary) *AffiliateSummary {
	if summary == nil {
		return nil
	}
	clone := *summary
	return &clone
}

func newAffiliateServiceForTest(repo *affiliateServiceTestRepo, settings map[string]string) *AffiliateService {
	return NewAffiliateService(repo, NewSettingService(&affiliateServiceTestSettingsRepo{values: settings}, nil), nil, nil)
}

func TestAffiliateBindInviterByCode(t *testing.T) {
	ctx := context.Background()
	repo := newAffiliateServiceTestRepo()
	repo.put(AffiliateSummary{UserID: 1, AffCode: "INVITER"})
	repo.put(AffiliateSummary{UserID: 2, AffCode: "INVITEE"})
	svc := newAffiliateServiceForTest(repo, map[string]string{
		SettingKeyAffiliateEnabled: "true",
	})

	require.NoError(t, svc.BindInviterByCode(ctx, 2, " inviter "))
	require.NotNil(t, repo.summaries[2].InviterID)
	require.Equal(t, int64(1), *repo.summaries[2].InviterID)
	require.Equal(t, 1, repo.summaries[1].AffCount)

	require.ErrorIs(t, svc.BindInviterByCode(ctx, 1, "INVITER"), ErrAffiliateCodeInvalid)
}

func TestAffiliateAccrueInviteRebateCapsAndFreezes(t *testing.T) {
	ctx := context.Background()
	repo := newAffiliateServiceTestRepo()
	repo.put(AffiliateSummary{UserID: 1, AffCode: "INVITER"})
	inviterID := int64(1)
	repo.put(AffiliateSummary{UserID: 2, AffCode: "INVITEE", InviterID: &inviterID, CreatedAt: time.Now()})
	repo.accruedByPair[[2]int64{1, 2}] = 5
	svc := newAffiliateServiceForTest(repo, map[string]string{
		SettingKeyAffiliateEnabled:             "true",
		SettingKeyAffiliateRebateRate:          "20",
		SettingKeyAffiliateRebateFreezeHours:   "12",
		SettingKeyAffiliateRebateDurationDays:  "30",
		SettingKeyAffiliateRebatePerInviteeCap: "7",
	})

	rebate, err := svc.AccrueInviteRebate(ctx, 2, 100)
	require.NoError(t, err)
	require.Equal(t, 2.0, rebate)
	require.Equal(t, 2.0, repo.lastAccrueAmount)
	require.Equal(t, 12, repo.lastFreezeHours)
	require.Equal(t, 2.0, repo.summaries[1].AffFrozenQuota)
	require.Equal(t, 2.0, repo.summaries[1].AffHistoryQuota)
}

func TestAffiliateTransferQuota(t *testing.T) {
	ctx := context.Background()
	repo := newAffiliateServiceTestRepo()
	repo.put(AffiliateSummary{UserID: 1, AffCode: "INVITER", AffQuota: 3.5})
	svc := newAffiliateServiceForTest(repo, map[string]string{
		SettingKeyAffiliateEnabled: "true",
	})

	transferred, balance, err := svc.TransferAffiliateQuota(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, 3.5, transferred)
	require.Equal(t, 103.5, balance)
	require.Zero(t, repo.summaries[1].AffQuota)
}
