package service

import "sync"

func runGrokParallelAccounts(
	accounts []Account,
	workerCount int,
	fn func(*Account) error,
) error {
	if len(accounts) == 0 {
		return nil
	}

	accountPtrs := make([]*Account, 0, len(accounts))
	for i := range accounts {
		accountPtrs = append(accountPtrs, &accounts[i])
	}
	return runGrokParallelAccountPointers(accountPtrs, workerCount, fn)
}

func runGrokParallelAccountPointers(
	accounts []*Account,
	workerCount int,
	fn func(*Account) error,
) error {
	if len(accounts) == 0 || fn == nil {
		return nil
	}
	if workerCount <= 0 {
		workerCount = 1
	}
	if workerCount > len(accounts) {
		workerCount = len(accounts)
	}

	jobs := make(chan *Account)
	var (
		wg       sync.WaitGroup
		errMu    sync.Mutex
		firstErr error
	)

	recordError := func(err error) {
		if err == nil {
			return
		}
		errMu.Lock()
		if firstErr == nil {
			firstErr = err
		}
		errMu.Unlock()
	}

	for worker := 0; worker < workerCount; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for account := range jobs {
				if account == nil {
					continue
				}
				recordError(fn(account))
			}
		}()
	}

	for _, account := range accounts {
		jobs <- account
	}
	close(jobs)
	wg.Wait()
	return firstErr
}
