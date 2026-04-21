package service

import "fmt"

type grokRequestScopedAttempt[T any] struct {
	account *Account
	value   T
}

type grokRequestScopedErrorPhase string

const (
	grokRequestScopedErrorPhasePrepare grokRequestScopedErrorPhase = "prepare"
	grokRequestScopedErrorPhaseExecute grokRequestScopedErrorPhase = "execute"
)

type grokRequestScopedExecutionError struct {
	phase grokRequestScopedErrorPhase
	err   error
}

func (e *grokRequestScopedExecutionError) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *grokRequestScopedExecutionError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func newGrokRequestScopedExecutionError(phase grokRequestScopedErrorPhase, err error) error {
	if err == nil {
		return nil
	}
	return &grokRequestScopedExecutionError{
		phase: phase,
		err:   err,
	}
}

func executeGrokRequestScopedFailover[T any](
	prepare func(excludedIDs map[int64]struct{}) (*grokRequestScopedAttempt[T], error),
	execute func(*grokRequestScopedAttempt[T]) error,
	classifyFailover func(error) *UpstreamFailoverError,
	shouldReturnLastFailover func(error) bool,
) error {
	if prepare == nil || execute == nil {
		return nil
	}

	excludedIDs := make(map[int64]struct{})
	var lastFailoverErr *UpstreamFailoverError

	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		prepared, err := prepare(excludedIDs)
		if err != nil {
			if lastFailoverErr != nil && shouldReturnLastFailover != nil && shouldReturnLastFailover(err) {
				return lastFailoverErr
			}
			return newGrokRequestScopedExecutionError(grokRequestScopedErrorPhasePrepare, err)
		}
		if prepared == nil {
			return newGrokRequestScopedExecutionError(
				grokRequestScopedErrorPhasePrepare,
				fmt.Errorf("grok failover preparation returned nil attempt"),
			)
		}
		if prepared.account == nil {
			return newGrokRequestScopedExecutionError(
				grokRequestScopedErrorPhasePrepare,
				fmt.Errorf("grok failover preparation returned nil account"),
			)
		}

		err = execute(prepared)
		if err == nil {
			return nil
		}

		failoverErr := classifyFailover(err)
		if failoverErr == nil {
			return newGrokRequestScopedExecutionError(grokRequestScopedErrorPhaseExecute, err)
		}

		lastFailoverErr = failoverErr
		excludedIDs[prepared.account.ID] = struct{}{}
		if len(excludedIDs) >= maxRetryAttempts {
			break
		}
	}

	if lastFailoverErr != nil {
		return lastFailoverErr
	}
	return newGrokRequestScopedExecutionError(
		grokRequestScopedErrorPhasePrepare,
		fmt.Errorf("grok failover exhausted without a classified upstream error"),
	)
}
