package handler

import "github.com/senran-N/sub2api/internal/service"

func (h *GatewayHandler) nativeGatewayRuntime() *service.NativeGatewayRuntime {
	if h == nil {
		return service.NewNativeGatewayRuntime(nil, nil, nil)
	}
	return service.NewNativeGatewayRuntime(
		h.gatewayService,
		h.geminiCompatService,
		h.antigravityGatewayService,
	)
}
