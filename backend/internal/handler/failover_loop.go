package handler

import "github.com/senran-N/sub2api/internal/service"

type TempUnscheduler = service.TempUnscheduler
type FailoverAction = service.RuntimeFailoverAction

const (
	FailoverContinue  = service.RuntimeFailoverContinue
	FailoverExhausted = service.RuntimeFailoverExhausted
	FailoverCanceled  = service.RuntimeFailoverCanceled
)

type FailoverState = service.RuntimeFailoverState

func NewFailoverState(maxSwitches int, hasBoundSession bool) *FailoverState {
	return service.NewRuntimeFailoverState(maxSwitches, hasBoundSession)
}
