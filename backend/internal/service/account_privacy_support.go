package service

import "context"

func readAccountCredential(account *Account, key string) string {
	value, _ := account.Credentials[key].(string)
	return value
}

func setAccountPrivacyMode(account *Account, mode string) {
	if account.Extra == nil {
		account.Extra = make(map[string]any)
	}
	account.Extra["privacy_mode"] = mode
}

func resolveAccountProxyURL(ctx context.Context, proxyRepo ProxyRepository, proxyID *int64) string {
	if proxyID == nil || proxyRepo == nil {
		return ""
	}

	proxy, err := proxyRepo.GetByID(ctx, *proxyID)
	if err != nil || proxy == nil {
		return ""
	}

	return proxy.URL()
}

func persistAccountPrivacyMode(
	ctx context.Context,
	accountRepo AccountRepository,
	account *Account,
	mode string,
	applyInMemory func(*Account, string),
	onPersisted func(error),
) string {
	if mode == "" {
		return ""
	}

	err := accountRepo.UpdateExtra(ctx, account.ID, map[string]any{"privacy_mode": mode})
	if onPersisted != nil {
		onPersisted(err)
	}
	if err != nil {
		return mode
	}

	if applyInMemory != nil {
		applyInMemory(account, mode)
	}
	return mode
}
