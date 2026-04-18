package service

const openAIWSFallbackCoolingScorePenalty = 8.0

type openAIAccountScoreContext struct {
	minPriority   int
	maxPriority   int
	maxWaiting    int
	loadRateSum   float64
	loadRateSumSq float64
	minTTFT       float64
	maxTTFT       float64
	hasTTFTSample bool
}

func newOpenAIAccountScoreContext(firstPriority int) openAIAccountScoreContext {
	return openAIAccountScoreContext{
		minPriority: firstPriority,
		maxPriority: firstPriority,
		maxWaiting:  1,
	}
}

func (ctx *openAIAccountScoreContext) observe(
	account *Account,
	loadInfo *AccountLoadInfo,
	ttft float64,
	hasTTFT bool,
) {
	if account.Priority < ctx.minPriority {
		ctx.minPriority = account.Priority
	}
	if account.Priority > ctx.maxPriority {
		ctx.maxPriority = account.Priority
	}
	if loadInfo.WaitingCount > ctx.maxWaiting {
		ctx.maxWaiting = loadInfo.WaitingCount
	}
	if hasTTFT && ttft > 0 {
		if !ctx.hasTTFTSample {
			ctx.minTTFT = ttft
			ctx.maxTTFT = ttft
			ctx.hasTTFTSample = true
		} else {
			if ttft < ctx.minTTFT {
				ctx.minTTFT = ttft
			}
			if ttft > ctx.maxTTFT {
				ctx.maxTTFT = ttft
			}
		}
	}

	loadRate := float64(loadInfo.LoadRate)
	ctx.loadRateSum += loadRate
	ctx.loadRateSumSq += loadRate * loadRate
}

func (ctx openAIAccountScoreContext) loadSkew(candidateCount int) float64 {
	return calcLoadSkewByMoments(ctx.loadRateSum, ctx.loadRateSumSq, candidateCount)
}

func (s *defaultOpenAIAccountScheduler) buildOpenAILoadBalancedCandidates(
	filtered []*Account,
	loadMap map[int64]*AccountLoadInfo,
	requiredTransport OpenAIUpstreamTransport,
) ([]openAIAccountCandidateScore, float64) {
	if len(filtered) == 0 {
		return nil, 0
	}

	scoreCtx := newOpenAIAccountScoreContext(filtered[0].Priority)
	candidates := make([]openAIAccountCandidateScore, 0, len(filtered))
	for _, account := range filtered {
		loadInfo := loadMap[account.ID]
		if loadInfo == nil {
			loadInfo = &AccountLoadInfo{AccountID: account.ID}
		}

		errorRate := 0.0
		ttft := 0.0
		hasTTFT := false
		if s != nil && s.stats != nil {
			errorRate, ttft, hasTTFT = s.stats.snapshot(account.ID)
		}
		cooling := s != nil && s.service != nil && s.service.isOpenAITransportFallbackCooling(account.ID, requiredTransport)
		scoreCtx.observe(account, loadInfo, ttft, hasTTFT)
		candidates = append(candidates, openAIAccountCandidateScore{
			account:   account,
			loadInfo:  loadInfo,
			cooling:   cooling,
			errorRate: errorRate,
			ttft:      ttft,
			hasTTFT:   hasTTFT,
		})
	}

	s.scoreOpenAICandidates(candidates, scoreCtx)
	return candidates, scoreCtx.loadSkew(len(candidates))
}

func (s *defaultOpenAIAccountScheduler) scoreOpenAICandidates(
	candidates []openAIAccountCandidateScore,
	scoreCtx openAIAccountScoreContext,
) {
	weights := s.service.openAIWSSchedulerWeights()
	for i := range candidates {
		item := &candidates[i]
		priorityFactor := 1.0
		if scoreCtx.maxPriority > scoreCtx.minPriority {
			priorityFactor = 1 - float64(item.account.Priority-scoreCtx.minPriority)/
				float64(scoreCtx.maxPriority-scoreCtx.minPriority)
		}
		loadFactor := 1 - clamp01(float64(item.loadInfo.LoadRate)/100.0)
		queueFactor := 1 - clamp01(float64(item.loadInfo.WaitingCount)/float64(scoreCtx.maxWaiting))
		errorFactor := 1 - clamp01(item.errorRate)
		ttftFactor := 0.5
		if item.hasTTFT && scoreCtx.hasTTFTSample && scoreCtx.maxTTFT > scoreCtx.minTTFT {
			ttftFactor = 1 - clamp01((item.ttft-scoreCtx.minTTFT)/(scoreCtx.maxTTFT-scoreCtx.minTTFT))
		}

		item.score = weights.Priority*priorityFactor +
			weights.Load*loadFactor +
			weights.Queue*queueFactor +
			weights.ErrorRate*errorFactor +
			weights.TTFT*ttftFactor
		if item.cooling {
			item.score -= openAIWSFallbackCoolingScorePenalty
		}
	}
}
