package service

import (
	"context"
	"errors"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/senran-N/sub2api/internal/config"
)

const (
	openAIWSConnMaxAge             = 60 * time.Minute
	openAIWSConnHealthCheckIdle    = 90 * time.Second
	openAIWSConnHealthCheckTO      = 2 * time.Second
	openAIWSConnPrewarmExtraDelay  = 2 * time.Second
	openAIWSAcquireCleanupInterval = 3 * time.Second
	openAIWSBackgroundPingInterval = 30 * time.Second
	openAIWSBackgroundSweepTicker  = 30 * time.Second

	openAIWSPrewarmFailureWindow   = 30 * time.Second
	openAIWSPrewarmFailureSuppress = 2
)

type openAIWSAcquireRequest struct {
	Account         *Account
	WSURL           string
	Headers         http.Header
	ProxyURL        string
	PreferredConnID string
	// ForceNewConn: 强制本次获取新连接（避免复用导致连接内续链状态互相污染）。
	ForceNewConn bool
	// ForcePreferredConn: 强制本次只使用 PreferredConnID，禁止漂移到其它连接。
	ForcePreferredConn bool
}

type openAIWSAccountPool struct {
	mu            sync.Mutex
	conns         map[string]*openAIWSConn
	pinnedConns   map[string]int
	creating      int
	lastCleanupAt time.Time
	lastAcquire   *openAIWSAcquireRequest
	prewarmActive bool
	prewarmUntil  time.Time
	prewarmFails  int
	prewarmFailAt time.Time
}

type OpenAIWSPoolMetricsSnapshot struct {
	AcquireTotal            int64
	AcquireReuseTotal       int64
	AcquireCreateTotal      int64
	AcquireQueueWaitTotal   int64
	AcquireQueueWaitMsTotal int64
	ConnPickTotal           int64
	ConnPickMsTotal         int64
	ScaleUpTotal            int64
	ScaleDownTotal          int64
}

type openAIWSPoolMetrics struct {
	acquireTotal          atomic.Int64
	acquireReuseTotal     atomic.Int64
	acquireCreateTotal    atomic.Int64
	acquireQueueWaitTotal atomic.Int64
	acquireQueueWaitMs    atomic.Int64
	connPickTotal         atomic.Int64
	connPickMs            atomic.Int64
	scaleUpTotal          atomic.Int64
	scaleDownTotal        atomic.Int64
}

type openAIWSConnPool struct {
	cfg *config.Config
	// 通过接口解耦底层 WS 客户端实现，默认使用 coder/websocket。
	clientDialer openAIWSClientDialer

	accounts sync.Map // key: int64(accountID), value: *openAIWSAccountPool
	seq      atomic.Uint64

	metrics openAIWSPoolMetrics

	workerStopCh chan struct{}
	workerWg     sync.WaitGroup
	closeOnce    sync.Once
}

func newOpenAIWSConnPool(cfg *config.Config) *openAIWSConnPool {
	pool := &openAIWSConnPool{
		cfg:          cfg,
		clientDialer: newDefaultOpenAIWSClientDialer(),
		workerStopCh: make(chan struct{}),
	}
	pool.startBackgroundWorkers()
	return pool
}

func (p *openAIWSConnPool) SnapshotMetrics() OpenAIWSPoolMetricsSnapshot {
	if p == nil {
		return OpenAIWSPoolMetricsSnapshot{}
	}
	return OpenAIWSPoolMetricsSnapshot{
		AcquireTotal:            p.metrics.acquireTotal.Load(),
		AcquireReuseTotal:       p.metrics.acquireReuseTotal.Load(),
		AcquireCreateTotal:      p.metrics.acquireCreateTotal.Load(),
		AcquireQueueWaitTotal:   p.metrics.acquireQueueWaitTotal.Load(),
		AcquireQueueWaitMsTotal: p.metrics.acquireQueueWaitMs.Load(),
		ConnPickTotal:           p.metrics.connPickTotal.Load(),
		ConnPickMsTotal:         p.metrics.connPickMs.Load(),
		ScaleUpTotal:            p.metrics.scaleUpTotal.Load(),
		ScaleDownTotal:          p.metrics.scaleDownTotal.Load(),
	}
}

func (p *openAIWSConnPool) SnapshotTransportMetrics() OpenAIWSTransportMetricsSnapshot {
	if p == nil {
		return OpenAIWSTransportMetricsSnapshot{}
	}
	if dialer, ok := p.clientDialer.(openAIWSTransportMetricsDialer); ok {
		return dialer.SnapshotTransportMetrics()
	}
	return OpenAIWSTransportMetricsSnapshot{}
}

func (p *openAIWSConnPool) setClientDialerForTest(dialer openAIWSClientDialer) {
	if p == nil || dialer == nil {
		return
	}
	p.clientDialer = dialer
}

func (p *openAIWSConnPool) Acquire(ctx context.Context, req openAIWSAcquireRequest) (*openAIWSConnLease, error) {
	if p != nil {
		p.metrics.acquireTotal.Add(1)
	}
	return p.acquire(ctx, cloneOpenAIWSAcquireRequest(req), 0)
}

func (p *openAIWSConnPool) acquire(ctx context.Context, req openAIWSAcquireRequest, retry int) (*openAIWSConnLease, error) {
	if p == nil || req.Account == nil || req.Account.ID <= 0 {
		return nil, errors.New("invalid ws acquire request")
	}
	if stringsTrim(req.WSURL) == "" {
		return nil, errors.New("ws url is empty")
	}

	accountID := req.Account.ID
	effectiveMaxConns := p.effectiveMaxConnsByAccount(req.Account)
	if effectiveMaxConns <= 0 {
		return nil, errOpenAIWSConnQueueFull
	}
	var evicted []*openAIWSConn
	ap := p.getOrCreateAccountPool(accountID)
	ap.mu.Lock()
	ap.lastAcquire = cloneOpenAIWSAcquireRequestPtr(&req)
	now := time.Now()
	if ap.lastCleanupAt.IsZero() || now.Sub(ap.lastCleanupAt) >= openAIWSAcquireCleanupInterval {
		evicted = p.cleanupAccountLocked(ap, now, effectiveMaxConns)
		ap.lastCleanupAt = now
	}
	pickStartedAt := time.Now()
	allowReuse := !req.ForceNewConn
	preferredConnID := stringsTrim(req.PreferredConnID)
	forcePreferredConn := allowReuse && req.ForcePreferredConn

	if allowReuse {
		if forcePreferredConn {
			if preferredConnID == "" {
				p.recordConnPickDuration(time.Since(pickStartedAt))
				ap.mu.Unlock()
				closeOpenAIWSConns(evicted)
				return nil, errOpenAIWSPreferredConnUnavailable
			}
			preferredConn, ok := ap.conns[preferredConnID]
			if !ok || preferredConn == nil {
				p.recordConnPickDuration(time.Since(pickStartedAt))
				ap.mu.Unlock()
				closeOpenAIWSConns(evicted)
				return nil, errOpenAIWSPreferredConnUnavailable
			}
			if preferredConn.tryAcquire() {
				connPick := time.Since(pickStartedAt)
				p.recordConnPickDuration(connPick)
				ap.mu.Unlock()
				closeOpenAIWSConns(evicted)
				if p.shouldHealthCheckConn(preferredConn) {
					if err := preferredConn.pingWithTimeout(openAIWSConnHealthCheckTO); err != nil {
						preferredConn.close()
						p.evictConn(accountID, preferredConn.id)
						if retry < 1 {
							return p.acquire(ctx, req, retry+1)
						}
						return nil, err
					}
				}
				lease := &openAIWSConnLease{
					pool:      p,
					accountID: accountID,
					conn:      preferredConn,
					connPick:  connPick,
					reused:    true,
				}
				p.metrics.acquireReuseTotal.Add(1)
				p.ensureTargetIdleAsync(accountID)
				return lease, nil
			}

			connPick := time.Since(pickStartedAt)
			p.recordConnPickDuration(connPick)
			if int(preferredConn.waiters.Load()) >= p.queueLimitPerConn() {
				ap.mu.Unlock()
				closeOpenAIWSConns(evicted)
				return nil, errOpenAIWSConnQueueFull
			}
			preferredConn.waiters.Add(1)
			ap.mu.Unlock()
			closeOpenAIWSConns(evicted)
			defer preferredConn.waiters.Add(-1)
			waitStart := time.Now()
			p.metrics.acquireQueueWaitTotal.Add(1)

			if err := preferredConn.acquire(ctx); err != nil {
				if errors.Is(err, errOpenAIWSConnClosed) && retry < 1 {
					return p.acquire(ctx, req, retry+1)
				}
				return nil, err
			}
			if p.shouldHealthCheckConn(preferredConn) {
				if err := preferredConn.pingWithTimeout(openAIWSConnHealthCheckTO); err != nil {
					preferredConn.release()
					preferredConn.close()
					p.evictConn(accountID, preferredConn.id)
					if retry < 1 {
						return p.acquire(ctx, req, retry+1)
					}
					return nil, err
				}
			}

			queueWait := time.Since(waitStart)
			p.metrics.acquireQueueWaitMs.Add(queueWait.Milliseconds())
			lease := &openAIWSConnLease{
				pool:      p,
				accountID: accountID,
				conn:      preferredConn,
				queueWait: queueWait,
				connPick:  connPick,
				reused:    true,
			}
			p.metrics.acquireReuseTotal.Add(1)
			p.ensureTargetIdleAsync(accountID)
			return lease, nil
		}

		if preferredConnID != "" {
			if conn, ok := ap.conns[preferredConnID]; ok && conn.tryAcquire() {
				connPick := time.Since(pickStartedAt)
				p.recordConnPickDuration(connPick)
				ap.mu.Unlock()
				closeOpenAIWSConns(evicted)
				if p.shouldHealthCheckConn(conn) {
					if err := conn.pingWithTimeout(openAIWSConnHealthCheckTO); err != nil {
						conn.close()
						p.evictConn(accountID, conn.id)
						if retry < 1 {
							return p.acquire(ctx, req, retry+1)
						}
						return nil, err
					}
				}
				lease := &openAIWSConnLease{pool: p, accountID: accountID, conn: conn, connPick: connPick, reused: true}
				p.metrics.acquireReuseTotal.Add(1)
				p.ensureTargetIdleAsync(accountID)
				return lease, nil
			}
		}

		best := p.pickLeastBusyConnLocked(ap, "")
		if best != nil && best.tryAcquire() {
			connPick := time.Since(pickStartedAt)
			p.recordConnPickDuration(connPick)
			ap.mu.Unlock()
			closeOpenAIWSConns(evicted)
			if p.shouldHealthCheckConn(best) {
				if err := best.pingWithTimeout(openAIWSConnHealthCheckTO); err != nil {
					best.close()
					p.evictConn(accountID, best.id)
					if retry < 1 {
						return p.acquire(ctx, req, retry+1)
					}
					return nil, err
				}
			}
			lease := &openAIWSConnLease{pool: p, accountID: accountID, conn: best, connPick: connPick, reused: true}
			p.metrics.acquireReuseTotal.Add(1)
			p.ensureTargetIdleAsync(accountID)
			return lease, nil
		}
		for _, conn := range ap.conns {
			if conn == nil || conn == best {
				continue
			}
			if conn.tryAcquire() {
				connPick := time.Since(pickStartedAt)
				p.recordConnPickDuration(connPick)
				ap.mu.Unlock()
				closeOpenAIWSConns(evicted)
				if p.shouldHealthCheckConn(conn) {
					if err := conn.pingWithTimeout(openAIWSConnHealthCheckTO); err != nil {
						conn.close()
						p.evictConn(accountID, conn.id)
						if retry < 1 {
							return p.acquire(ctx, req, retry+1)
						}
						return nil, err
					}
				}
				lease := &openAIWSConnLease{pool: p, accountID: accountID, conn: conn, connPick: connPick, reused: true}
				p.metrics.acquireReuseTotal.Add(1)
				p.ensureTargetIdleAsync(accountID)
				return lease, nil
			}
		}
	}

	if req.ForceNewConn && len(ap.conns)+ap.creating >= effectiveMaxConns {
		if idle := p.pickOldestIdleConnLocked(ap); idle != nil {
			delete(ap.conns, idle.id)
			evicted = append(evicted, idle)
			p.metrics.scaleDownTotal.Add(1)
		}
	}

	if len(ap.conns)+ap.creating < effectiveMaxConns {
		connPick := time.Since(pickStartedAt)
		p.recordConnPickDuration(connPick)
		ap.creating++
		ap.mu.Unlock()
		closeOpenAIWSConns(evicted)

		conn, dialErr := p.dialConn(ctx, req)

		ap = p.getOrCreateAccountPool(accountID)
		ap.mu.Lock()
		ap.creating--
		if dialErr != nil {
			ap.prewarmFails++
			ap.prewarmFailAt = time.Now()
			ap.mu.Unlock()
			return nil, dialErr
		}
		ap.conns[conn.id] = conn
		ap.prewarmFails = 0
		ap.prewarmFailAt = time.Time{}
		ap.mu.Unlock()
		p.metrics.acquireCreateTotal.Add(1)

		if !conn.tryAcquire() {
			if err := conn.acquire(ctx); err != nil {
				conn.close()
				p.evictConn(accountID, conn.id)
				return nil, err
			}
		}
		lease := &openAIWSConnLease{pool: p, accountID: accountID, conn: conn, connPick: connPick}
		p.ensureTargetIdleAsync(accountID)
		return lease, nil
	}

	if req.ForceNewConn {
		p.recordConnPickDuration(time.Since(pickStartedAt))
		ap.mu.Unlock()
		closeOpenAIWSConns(evicted)
		return nil, errOpenAIWSConnQueueFull
	}

	target := p.pickLeastBusyConnLocked(ap, req.PreferredConnID)
	connPick := time.Since(pickStartedAt)
	p.recordConnPickDuration(connPick)
	if target == nil {
		ap.mu.Unlock()
		closeOpenAIWSConns(evicted)
		return nil, errOpenAIWSConnClosed
	}
	if int(target.waiters.Load()) >= p.queueLimitPerConn() {
		ap.mu.Unlock()
		closeOpenAIWSConns(evicted)
		return nil, errOpenAIWSConnQueueFull
	}
	target.waiters.Add(1)
	ap.mu.Unlock()
	closeOpenAIWSConns(evicted)
	defer target.waiters.Add(-1)
	waitStart := time.Now()
	p.metrics.acquireQueueWaitTotal.Add(1)

	if err := target.acquire(ctx); err != nil {
		if errors.Is(err, errOpenAIWSConnClosed) && retry < 1 {
			return p.acquire(ctx, req, retry+1)
		}
		return nil, err
	}
	if p.shouldHealthCheckConn(target) {
		if err := target.pingWithTimeout(openAIWSConnHealthCheckTO); err != nil {
			target.release()
			target.close()
			p.evictConn(accountID, target.id)
			if retry < 1 {
				return p.acquire(ctx, req, retry+1)
			}
			return nil, err
		}
	}

	queueWait := time.Since(waitStart)
	p.metrics.acquireQueueWaitMs.Add(queueWait.Milliseconds())
	lease := &openAIWSConnLease{pool: p, accountID: accountID, conn: target, queueWait: queueWait, connPick: connPick, reused: true}
	p.metrics.acquireReuseTotal.Add(1)
	p.ensureTargetIdleAsync(accountID)
	return lease, nil
}

func (p *openAIWSConnPool) pickOldestIdleConnLocked(ap *openAIWSAccountPool) *openAIWSConn {
	if ap == nil || len(ap.conns) == 0 {
		return nil
	}
	var oldest *openAIWSConn
	for _, conn := range ap.conns {
		if conn == nil || conn.isLeased() || conn.waiters.Load() > 0 || p.isConnPinnedLocked(ap, conn.id) {
			continue
		}
		if oldest == nil || conn.lastUsedAt().Before(oldest.lastUsedAt()) {
			oldest = conn
		}
	}
	return oldest
}

func (p *openAIWSConnPool) getOrCreateAccountPool(accountID int64) *openAIWSAccountPool {
	if p == nil || accountID <= 0 {
		return nil
	}
	if existing, ok := p.accounts.Load(accountID); ok {
		if ap, typed := existing.(*openAIWSAccountPool); typed && ap != nil {
			return ap
		}
	}
	ap := &openAIWSAccountPool{
		conns:       make(map[string]*openAIWSConn),
		pinnedConns: make(map[string]int),
	}
	actual, _ := p.accounts.LoadOrStore(accountID, ap)
	if typed, ok := actual.(*openAIWSAccountPool); ok && typed != nil {
		return typed
	}
	return ap
}

// ensureAccountPoolLocked 兼容旧调用。
func (p *openAIWSConnPool) ensureAccountPoolLocked(accountID int64) *openAIWSAccountPool {
	return p.getOrCreateAccountPool(accountID)
}

func (p *openAIWSConnPool) getAccountPool(accountID int64) (*openAIWSAccountPool, bool) {
	if p == nil || accountID <= 0 {
		return nil, false
	}
	value, ok := p.accounts.Load(accountID)
	if !ok || value == nil {
		return nil, false
	}
	ap, typed := value.(*openAIWSAccountPool)
	return ap, typed && ap != nil
}

func (p *openAIWSConnPool) isConnPinnedLocked(ap *openAIWSAccountPool, connID string) bool {
	if ap == nil || connID == "" || len(ap.pinnedConns) == 0 {
		return false
	}
	return ap.pinnedConns[connID] > 0
}

func (p *openAIWSConnPool) pickLeastBusyConnLocked(ap *openAIWSAccountPool, preferredConnID string) *openAIWSConn {
	if ap == nil || len(ap.conns) == 0 {
		return nil
	}
	preferredConnID = stringsTrim(preferredConnID)
	if preferredConnID != "" {
		if conn, ok := ap.conns[preferredConnID]; ok {
			return conn
		}
	}
	var best *openAIWSConn
	var bestWaiters int32
	var bestLastUsed time.Time
	for _, conn := range ap.conns {
		if conn == nil {
			continue
		}
		waiters := conn.waiters.Load()
		lastUsed := conn.lastUsedAt()
		if best == nil ||
			waiters < bestWaiters ||
			(waiters == bestWaiters && lastUsed.Before(bestLastUsed)) {
			best = conn
			bestWaiters = waiters
			bestLastUsed = lastUsed
		}
	}
	return best
}

func accountPoolLoadLocked(ap *openAIWSAccountPool) (inflight int, waiters int) {
	if ap == nil {
		return 0, 0
	}
	for _, conn := range ap.conns {
		if conn == nil {
			continue
		}
		if conn.isLeased() {
			inflight++
		}
		waiters += int(conn.waiters.Load())
	}
	return inflight, waiters
}

// AccountPoolLoad 返回指定账号连接池的并发与排队快照。
func (p *openAIWSConnPool) AccountPoolLoad(accountID int64) (inflight int, waiters int, conns int) {
	if p == nil || accountID <= 0 {
		return 0, 0, 0
	}
	ap, ok := p.getAccountPool(accountID)
	if !ok || ap == nil {
		return 0, 0, 0
	}
	ap.mu.Lock()
	defer ap.mu.Unlock()
	inflight, waiters = accountPoolLoadLocked(ap)
	return inflight, waiters, len(ap.conns)
}

func (p *openAIWSConnPool) ensureTargetIdleAsync(accountID int64) {
	if p == nil || accountID <= 0 {
		return
	}

	var req openAIWSAcquireRequest
	need := 0
	ap, ok := p.getAccountPool(accountID)
	if !ok || ap == nil {
		return
	}
	ap.mu.Lock()
	defer ap.mu.Unlock()
	if ap.lastAcquire == nil {
		return
	}
	if ap.prewarmActive {
		return
	}
	now := time.Now()
	if !ap.prewarmUntil.IsZero() && now.Before(ap.prewarmUntil) {
		return
	}
	if p.shouldSuppressPrewarmLocked(ap, now) {
		return
	}
	effectiveMaxConns := p.maxConnsHardCap()
	if ap.lastAcquire != nil && ap.lastAcquire.Account != nil {
		effectiveMaxConns = p.effectiveMaxConnsByAccount(ap.lastAcquire.Account)
	}
	target := p.targetConnCountLocked(ap, effectiveMaxConns)
	current := len(ap.conns) + ap.creating
	if current >= target {
		return
	}
	need = target - current
	if need <= 0 {
		return
	}
	req = cloneOpenAIWSAcquireRequest(*ap.lastAcquire)
	ap.prewarmActive = true
	if cooldown := p.prewarmCooldown(); cooldown > 0 {
		ap.prewarmUntil = now.Add(cooldown)
	}
	ap.creating += need
	p.metrics.scaleUpTotal.Add(int64(need))

	go p.prewarmConns(accountID, req, need)
}

func (p *openAIWSConnPool) targetConnCountLocked(ap *openAIWSAccountPool, maxConns int) int {
	if ap == nil {
		return 0
	}

	if maxConns <= 0 {
		return 0
	}

	minIdle := p.minIdlePerAccount()
	if minIdle < 0 {
		minIdle = 0
	}
	if minIdle > maxConns {
		minIdle = maxConns
	}

	inflight, waiters := accountPoolLoadLocked(ap)
	utilization := p.targetUtilization()
	demand := inflight + waiters
	if demand <= 0 {
		return minIdle
	}

	target := 1
	if demand > 1 {
		target = int(math.Ceil(float64(demand) / utilization))
	}
	if waiters > 0 && target < len(ap.conns)+1 {
		target = len(ap.conns) + 1
	}
	if target < minIdle {
		target = minIdle
	}
	if target > maxConns {
		target = maxConns
	}
	return target
}

func (p *openAIWSConnPool) prewarmConns(accountID int64, req openAIWSAcquireRequest, total int) {
	defer func() {
		if ap, ok := p.getAccountPool(accountID); ok && ap != nil {
			ap.mu.Lock()
			ap.prewarmActive = false
			ap.mu.Unlock()
		}
	}()

	for i := 0; i < total; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), p.dialTimeout()+openAIWSConnPrewarmExtraDelay)
		conn, err := p.dialConn(ctx, req)
		cancel()

		ap, ok := p.getAccountPool(accountID)
		if !ok || ap == nil {
			if conn != nil {
				conn.close()
			}
			return
		}
		ap.mu.Lock()
		if ap.creating > 0 {
			ap.creating--
		}
		if err != nil {
			ap.prewarmFails++
			ap.prewarmFailAt = time.Now()
			ap.mu.Unlock()
			continue
		}
		if len(ap.conns) >= p.effectiveMaxConnsByAccount(req.Account) {
			ap.mu.Unlock()
			conn.close()
			continue
		}
		ap.conns[conn.id] = conn
		ap.prewarmFails = 0
		ap.prewarmFailAt = time.Time{}
		ap.mu.Unlock()
	}
}

func (p *openAIWSConnPool) evictConn(accountID int64, connID string) {
	if p == nil || accountID <= 0 || stringsTrim(connID) == "" {
		return
	}
	var conn *openAIWSConn
	ap, ok := p.getAccountPool(accountID)
	if ok && ap != nil {
		ap.mu.Lock()
		if c, exists := ap.conns[connID]; exists {
			conn = c
			delete(ap.conns, connID)
			if len(ap.pinnedConns) > 0 {
				delete(ap.pinnedConns, connID)
			}
		}
		ap.mu.Unlock()
	}
	if conn != nil {
		conn.close()
	}
}

func (p *openAIWSConnPool) PinConn(accountID int64, connID string) bool {
	if p == nil || accountID <= 0 {
		return false
	}
	connID = stringsTrim(connID)
	if connID == "" {
		return false
	}
	ap, ok := p.getAccountPool(accountID)
	if !ok || ap == nil {
		return false
	}
	ap.mu.Lock()
	defer ap.mu.Unlock()
	if _, exists := ap.conns[connID]; !exists {
		return false
	}
	if ap.pinnedConns == nil {
		ap.pinnedConns = make(map[string]int)
	}
	ap.pinnedConns[connID]++
	return true
}

func (p *openAIWSConnPool) UnpinConn(accountID int64, connID string) {
	if p == nil || accountID <= 0 {
		return
	}
	connID = stringsTrim(connID)
	if connID == "" {
		return
	}
	ap, ok := p.getAccountPool(accountID)
	if !ok || ap == nil {
		return
	}
	ap.mu.Lock()
	defer ap.mu.Unlock()
	if len(ap.pinnedConns) == 0 {
		return
	}
	count := ap.pinnedConns[connID]
	if count <= 1 {
		delete(ap.pinnedConns, connID)
		return
	}
	ap.pinnedConns[connID] = count - 1
}

func (p *openAIWSConnPool) dialConn(ctx context.Context, req openAIWSAcquireRequest) (*openAIWSConn, error) {
	if p == nil || p.clientDialer == nil {
		return nil, errors.New("openai ws client dialer is nil")
	}
	conn, status, handshakeHeaders, err := p.clientDialer.Dial(ctx, req.WSURL, req.Headers, req.ProxyURL)
	if err != nil {
		return nil, &openAIWSDialError{
			StatusCode:      status,
			ResponseHeaders: cloneHeader(handshakeHeaders),
			Err:             err,
		}
	}
	if conn == nil {
		return nil, &openAIWSDialError{
			StatusCode:      status,
			ResponseHeaders: cloneHeader(handshakeHeaders),
			Err:             errors.New("openai ws dialer returned nil connection"),
		}
	}
	id := p.nextConnID(req.Account.ID)
	return newOpenAIWSConn(id, req.Account.ID, conn, handshakeHeaders), nil
}
