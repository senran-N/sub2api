package service

import (
	"container/heap"
	"context"
	"errors"
	"hash/fnv"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (s *defaultOpenAIAccountScheduler) selectBySessionHash(
	ctx context.Context,
	req OpenAIAccountScheduleRequest,
) (*AccountSelectionResult, bool, error) {
	sessionHash := strings.TrimSpace(req.SessionHash)
	if sessionHash == "" || s == nil || s.service == nil || s.service.cache == nil {
		return nil, false, nil
	}

	stickyAccountID := req.StickyAccountID
	if stickyAccountID <= 0 {
		var err error
		stickyAccountID, err = s.service.getStickySessionAccountID(ctx, req.GroupID, sessionHash)
		if err != nil || stickyAccountID <= 0 {
			return nil, false, nil
		}
	}
	if s.service.isOpenAITransportFallbackCooling(stickyAccountID, req.RequiredTransport) {
		recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindSession, newStickyBindingSoftMiss("transport_cooling"), stickyAccountID, sessionHash, "")
		return nil, true, nil
	}

	account, accountID := s.service.resolveOpenAIStickySessionAccount(
		ctx,
		req.GroupID,
		sessionHash,
		req.RequestedModel,
		req.ExcludedIDs,
		stickyAccountID,
		openAIStickySessionResolvePolicy{
			deleteOnLookupMiss: true,
			refreshTTLOnHit:    false,
			recheckOnResolve:   true,
		},
	)
	if account == nil {
		return nil, true, nil
	}
	if !s.isAccountTransportCompatible(account, req.RequiredTransport) {
		recordOpenAIStickyBindingDisposition(ctx, stickyBindingKindSession, newStickyBindingSoftMiss("transport_incompatible"), accountID, sessionHash, "")
		return nil, true, nil
	}

	cfg := gatewaySchedulingConfigOrDefault(s.service.cfg)
	if selection, ok := s.service.trySelectResolvedOpenAIStickyAccount(ctx, openAIStickyResolvedSelectionSpec{
		account:        account,
		accountID:      accountID,
		cfg:            cfg,
		stickyWaitPlan: s.service.buildOpenAIStickyWaitPlanAdapter(cfg),
		onSelected: s.service.buildOpenAIStickyTTLSelectionAdapter(
			ctx,
			req.GroupID,
			sessionHash,
			s.service.openAIWSSessionStickyTTL(),
		),
	}); ok {
		return selection, true, nil
	}
	return nil, true, nil
}

type openAIAccountCandidateScore struct {
	account   *Account
	loadInfo  *AccountLoadInfo
	cooling   bool
	score     float64
	errorRate float64
	ttft      float64
	hasTTFT   bool
}

type openAIAccountCandidateHeap []openAIAccountCandidateScore

func (h openAIAccountCandidateHeap) Len() int { return len(h) }

func (h openAIAccountCandidateHeap) Less(i, j int) bool {
	return isOpenAIAccountCandidateBetter(h[j], h[i])
}

func (h openAIAccountCandidateHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *openAIAccountCandidateHeap) Push(x any) {
	candidate, ok := x.(openAIAccountCandidateScore)
	if !ok {
		panic("openAIAccountCandidateHeap: invalid element type")
	}
	*h = append(*h, candidate)
}

func (h *openAIAccountCandidateHeap) Pop() any {
	old := *h
	last := old[len(old)-1]
	*h = old[:len(old)-1]
	return last
}

func isOpenAIAccountCandidateBetter(left openAIAccountCandidateScore, right openAIAccountCandidateScore) bool {
	if left.score != right.score {
		return left.score > right.score
	}
	if left.cooling != right.cooling {
		return !left.cooling
	}
	if left.account.Priority != right.account.Priority {
		return left.account.Priority < right.account.Priority
	}
	if left.loadInfo.LoadRate != right.loadInfo.LoadRate {
		return left.loadInfo.LoadRate < right.loadInfo.LoadRate
	}
	if left.loadInfo.WaitingCount != right.loadInfo.WaitingCount {
		return left.loadInfo.WaitingCount < right.loadInfo.WaitingCount
	}
	return left.account.ID < right.account.ID
}

func selectTopKOpenAICandidates(candidates []openAIAccountCandidateScore, topK int) []openAIAccountCandidateScore {
	if len(candidates) == 0 {
		return nil
	}
	if topK <= 0 {
		topK = 1
	}
	if topK >= len(candidates) {
		ranked := append([]openAIAccountCandidateScore(nil), candidates...)
		sort.Slice(ranked, func(i, j int) bool {
			return isOpenAIAccountCandidateBetter(ranked[i], ranked[j])
		})
		return ranked
	}

	best := make(openAIAccountCandidateHeap, 0, topK)
	for _, candidate := range candidates {
		if len(best) < topK {
			heap.Push(&best, candidate)
			continue
		}
		if isOpenAIAccountCandidateBetter(candidate, best[0]) {
			best[0] = candidate
			heap.Fix(&best, 0)
		}
	}

	ranked := make([]openAIAccountCandidateScore, len(best))
	copy(ranked, best)
	sort.Slice(ranked, func(i, j int) bool {
		return isOpenAIAccountCandidateBetter(ranked[i], ranked[j])
	})
	return ranked
}

type openAISelectionRNG struct {
	state uint64
}

func newOpenAISelectionRNG(seed uint64) openAISelectionRNG {
	if seed == 0 {
		seed = 0x9e3779b97f4a7c15
	}
	return openAISelectionRNG{state: seed}
}

func (r *openAISelectionRNG) nextUint64() uint64 {
	x := r.state
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	r.state = x
	return x * 2685821657736338717
}

func (r *openAISelectionRNG) nextFloat64() float64 {
	return float64(r.nextUint64()>>11) / (1 << 53)
}

func deriveOpenAISelectionSeed(req OpenAIAccountScheduleRequest) uint64 {
	hasher := fnv.New64a()
	writeValue := func(value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return
		}
		_, _ = hasher.Write([]byte(trimmed))
		_, _ = hasher.Write([]byte{0})
	}

	writeValue(req.SessionHash)
	writeValue(req.PreviousResponseID)
	writeValue(req.RequestedModel)
	if req.GroupID != nil {
		_, _ = hasher.Write([]byte(strconv.FormatInt(*req.GroupID, 10)))
	}

	seed := hasher.Sum64()
	if strings.TrimSpace(req.SessionHash) == "" && strings.TrimSpace(req.PreviousResponseID) == "" {
		seed ^= uint64(time.Now().UnixNano())
	}
	if seed == 0 {
		seed = uint64(time.Now().UnixNano()) ^ 0x9e3779b97f4a7c15
	}
	return seed
}

func buildOpenAIWeightedSelectionOrder(
	candidates []openAIAccountCandidateScore,
	req OpenAIAccountScheduleRequest,
) []openAIAccountCandidateScore {
	if len(candidates) <= 1 {
		return append([]openAIAccountCandidateScore(nil), candidates...)
	}

	pool := append([]openAIAccountCandidateScore(nil), candidates...)
	weights := make([]float64, len(pool))
	minScore := pool[0].score
	for i := 1; i < len(pool); i++ {
		if pool[i].score < minScore {
			minScore = pool[i].score
		}
	}
	for i := range pool {
		weight := (pool[i].score - minScore) + 1.0
		if math.IsNaN(weight) || math.IsInf(weight, 0) || weight <= 0 {
			weight = 1.0
		}
		weights[i] = weight
	}

	order := make([]openAIAccountCandidateScore, 0, len(pool))
	rng := newOpenAISelectionRNG(deriveOpenAISelectionSeed(req))
	for len(pool) > 0 {
		total := 0.0
		for _, weight := range weights {
			total += weight
		}

		selectedIdx := 0
		if total > 0 {
			r := rng.nextFloat64() * total
			acc := 0.0
			for i, weight := range weights {
				acc += weight
				if r <= acc {
					selectedIdx = i
					break
				}
			}
		} else {
			selectedIdx = int(rng.nextUint64() % uint64(len(pool)))
		}

		order = append(order, pool[selectedIdx])
		pool = append(pool[:selectedIdx], pool[selectedIdx+1:]...)
		weights = append(weights[:selectedIdx], weights[selectedIdx+1:]...)
	}
	return order
}

func buildOpenAIImmediateSelectionOrder(
	candidates []openAIAccountCandidateScore,
	topK int,
	req OpenAIAccountScheduleRequest,
) []openAIAccountCandidateScore {
	if len(candidates) == 0 {
		return nil
	}

	rankedTopK := selectTopKOpenAICandidates(candidates, topK)
	if len(rankedTopK) == 0 {
		return nil
	}
	if len(rankedTopK) == len(candidates) {
		return buildOpenAIWeightedSelectionOrder(rankedTopK, req)
	}

	selectionOrder := buildOpenAIWeightedSelectionOrder(rankedTopK, req)
	selected := make(map[int64]struct{}, len(selectionOrder))
	for _, candidate := range selectionOrder {
		if candidate.account != nil {
			selected[candidate.account.ID] = struct{}{}
		}
	}

	remaining := make([]openAIAccountCandidateScore, 0, len(candidates)-len(selectionOrder))
	for _, candidate := range candidates {
		if candidate.account == nil {
			continue
		}
		if _, exists := selected[candidate.account.ID]; exists {
			continue
		}
		remaining = append(remaining, candidate)
	}
	sort.Slice(remaining, func(i, j int) bool {
		return isOpenAIAccountCandidateBetter(remaining[i], remaining[j])
	})
	return append(selectionOrder, remaining...)
}

func chooseOpenAIWaitCandidate(candidates []openAIAccountCandidateScore) *openAIAccountCandidateScore {
	if len(candidates) == 0 {
		return nil
	}

	best := &candidates[0]
	for i := 1; i < len(candidates); i++ {
		candidate := &candidates[i]
		if candidate.cooling != best.cooling {
			if !candidate.cooling {
				best = candidate
			}
			continue
		}
		if candidate.loadInfo == nil {
			continue
		}
		if best == nil || best.loadInfo == nil {
			best = candidate
			continue
		}
		if candidate.loadInfo.WaitingCount != best.loadInfo.WaitingCount {
			if candidate.loadInfo.WaitingCount < best.loadInfo.WaitingCount {
				best = candidate
			}
			continue
		}
		if candidate.loadInfo.LoadRate != best.loadInfo.LoadRate {
			if candidate.loadInfo.LoadRate < best.loadInfo.LoadRate {
				best = candidate
			}
			continue
		}
		if candidate.account != nil && best.account != nil && candidate.account.Priority != best.account.Priority {
			if candidate.account.Priority < best.account.Priority {
				best = candidate
			}
			continue
		}
		if isOpenAIAccountCandidateBetter(*candidate, *best) {
			best = candidate
		}
	}
	return best
}

func (s *defaultOpenAIAccountScheduler) selectByLoadBalance(
	ctx context.Context,
	req OpenAIAccountScheduleRequest,
) (*AccountSelectionResult, int, int, float64, string, error) {
	if s != nil && s.service != nil && s.service.schedulerSnapshot != nil {
		return s.selectByLoadBalancePagedSnapshot(ctx, req)
	}
	return s.selectByLoadBalanceFullScan(ctx, req)
}

func (s *defaultOpenAIAccountScheduler) selectByLoadBalanceFullScan(
	ctx context.Context,
	req OpenAIAccountScheduleRequest,
) (*AccountSelectionResult, int, int, float64, string, error) {
	accounts, err := s.service.listSchedulableAccounts(ctx, req.GroupID)
	if err != nil {
		return nil, 0, 0, 0, openAIAccountScheduleLoadBalanceStrategyFullScan, err
	}
	if len(accounts) == 0 {
		return nil, 0, 0, 0, openAIAccountScheduleLoadBalanceStrategyFullScan, errors.New("no available OpenAI accounts")
	}

	var schedGroup *Group
	if req.GroupID != nil && s.service.schedulerSnapshot != nil {
		schedGroup, _ = s.service.schedulerSnapshot.GetGroupByID(ctx, *req.GroupID)
	}

	filtered, loadReq := s.prepareLoadBalanceCandidates(ctx, req, accounts, schedGroup)
	if len(filtered) == 0 {
		if req.RequestedModel != "" && !openAIRequestedModelAvailable(accounts, req.RequestedModel) {
			return nil, 0, 0, 0, openAIAccountScheduleLoadBalanceStrategyFullScan, newOpenAIRequestedModelUnavailableError(req.RequestedModel)
		}
		return nil, 0, 0, 0, openAIAccountScheduleLoadBalanceStrategyFullScan, errors.New("no available OpenAI accounts")
	}

	loadMap := s.loadSchedulerAccountLoads(ctx, loadReq)

	candidates, loadSkew := s.buildOpenAILoadBalancedCandidates(filtered, loadMap, req.RequiredTransport)

	topK := normalizeOpenAISchedulerTopK(s.service.openAIWSLBTopK(), len(candidates))
	selectionOrder := buildOpenAIImmediateSelectionOrder(candidates, topK, req)

	if result, acquireErr, ok := s.trySelectImmediateScheduledCandidate(ctx, req, selectionOrder); ok {
		if acquireErr != nil {
			return nil, len(candidates), topK, loadSkew, openAIAccountScheduleLoadBalanceStrategyFullScan, acquireErr
		}
		return result, len(candidates), topK, loadSkew, openAIAccountScheduleLoadBalanceStrategyFullScan, nil
	}

	if waitPlan, ok := s.tryBuildScheduledWaitPlan(ctx, req, candidates); ok {
		return waitPlan, len(candidates), topK, loadSkew, openAIAccountScheduleLoadBalanceStrategyFullScan, nil
	}

	return nil, len(candidates), topK, loadSkew, openAIAccountScheduleLoadBalanceStrategyFullScan, ErrNoAvailableAccounts
}

func (s *defaultOpenAIAccountScheduler) selectByLoadBalancePagedSnapshot(
	ctx context.Context,
	req OpenAIAccountScheduleRequest,
) (*AccountSelectionResult, int, int, float64, string, error) {
	if s == nil || s.service == nil || s.service.schedulerSnapshot == nil {
		return s.selectByLoadBalanceFullScan(ctx, req)
	}

	var schedGroup *Group
	if req.GroupID != nil {
		schedGroup, _ = s.service.schedulerSnapshot.GetGroupByID(ctx, *req.GroupID)
	}

	pageSize := s.service.openAISchedulerSnapshotPageSize()
	configuredTopK := s.service.openAIWSLBTopK()
	requestedModelAvailable := req.RequestedModel == ""
	candidateCount := 0
	loadRateSum := 0.0
	loadRateSumSq := 0.0
	var selectedResult *AccountSelectionResult
	pager, err := s.buildIndexedSnapshotPager(ctx, req, schedGroup)
	if err != nil {
		return nil, 0, 0, 0, openAIAccountScheduleLoadBalanceStrategyIndexedSnapshot, err
	}
	if pager == nil {
		return s.selectByLoadBalanceFullScan(ctx, req)
	}

	modelScopedFound, bestWaitCandidate, err := executeIndexedRuntimeSelection(
		ctx,
		pager,
		pageSize,
		func(accounts []*Account) (bool, *openAIAccountCandidateScore, error) {
			accounts = s.filterBatchByIndexedCapabilityPointers(ctx, req, accounts, schedGroup)
			prepared := s.prepareLoadBalanceCandidatePointers(req, accounts, schedGroup)
			requestedModelAvailable = requestedModelAvailable || prepared.requestedModelAvailable
			if len(prepared.filtered) == 0 {
				return false, nil, nil
			}

			loadMap := s.loadSchedulerAccountLoads(ctx, prepared.loadReq)
			candidates, _ := s.buildOpenAILoadBalancedCandidates(prepared.filtered, loadMap, req.RequiredTransport)
			candidateCount += len(candidates)
			for i := range candidates {
				loadRate := float64(candidates[i].loadInfo.LoadRate)
				loadRateSum += loadRate
				loadRateSumSq += loadRate * loadRate
			}

			pageTopK := normalizeOpenAISchedulerTopK(configuredTopK, len(candidates))
			selectionOrder := buildOpenAIImmediateSelectionOrder(candidates, pageTopK, req)
			if result, acquireErr, ok := s.trySelectImmediateScheduledCandidate(ctx, req, selectionOrder); ok {
				if acquireErr != nil {
					return false, nil, acquireErr
				}
				selectedResult = result
				return true, nil, nil
			}
			if len(candidates) == 0 {
				return false, nil, nil
			}
			return false, chooseOpenAIWaitCandidate(candidates), nil
		},
		isBetterOpenAIWaitCandidate,
	)
	if selectedResult != nil {
		loadSkew := calcLoadSkewByMoments(loadRateSum, loadRateSumSq, candidateCount)
		return selectedResult, candidateCount, selectionDecisionTopK(configuredTopK, candidateCount), loadSkew, openAIAccountScheduleLoadBalanceStrategyIndexedSnapshot, nil
	}
	if err != nil {
		return nil, candidateCount, selectionDecisionTopK(configuredTopK, candidateCount), calcLoadSkewByMoments(loadRateSum, loadRateSumSq, candidateCount), openAIAccountScheduleLoadBalanceStrategyIndexedSnapshot, err
	}

	if candidateCount == 0 {
		if req.RequestedModel != "" && !requestedModelAvailable && !modelScopedFound {
			return nil, 0, 0, 0, openAIAccountScheduleLoadBalanceStrategyIndexedSnapshot, newOpenAIRequestedModelUnavailableError(req.RequestedModel)
		}
		return nil, 0, 0, 0, openAIAccountScheduleLoadBalanceStrategyIndexedSnapshot, errors.New("no available OpenAI accounts")
	}

	loadSkew := calcLoadSkewByMoments(loadRateSum, loadRateSumSq, candidateCount)
	if bestWaitCandidate != nil {
		if waitPlan, ok := s.tryBuildScheduledWaitPlan(ctx, req, []openAIAccountCandidateScore{*bestWaitCandidate}); ok {
			return waitPlan, candidateCount, selectionDecisionTopK(configuredTopK, candidateCount), loadSkew, openAIAccountScheduleLoadBalanceStrategyIndexedSnapshot, nil
		}
	}

	return nil, candidateCount, selectionDecisionTopK(configuredTopK, candidateCount), loadSkew, openAIAccountScheduleLoadBalanceStrategyIndexedSnapshot, ErrNoAvailableAccounts
}

func (s *defaultOpenAIAccountScheduler) isAccountTransportCompatible(account *Account, requiredTransport OpenAIUpstreamTransport) bool {
	if requiredTransport == OpenAIUpstreamTransportAny || requiredTransport == OpenAIUpstreamTransportHTTPSSE {
		return true
	}
	if s == nil || s.service == nil || account == nil {
		return false
	}
	return s.service.getOpenAIWSProtocolResolver().Resolve(account).Transport == requiredTransport
}

func clamp01(value float64) float64 {
	switch {
	case value < 0:
		return 0
	case value > 1:
		return 1
	default:
		return value
	}
}

func calcLoadSkewByMoments(sum float64, sumSquares float64, count int) float64 {
	if count <= 1 {
		return 0
	}
	mean := sum / float64(count)
	variance := sumSquares/float64(count) - mean*mean
	if variance < 0 {
		variance = 0
	}
	return math.Sqrt(variance)
}
