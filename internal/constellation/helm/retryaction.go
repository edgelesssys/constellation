/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"fmt"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/retry"
)

const (
	// maximumRetryAttempts is the maximum number of attempts to retry a helm install.
	maximumRetryAttempts = 3
)

type retrieableApplier interface {
	apply(context.Context) error
	ReleaseName() string
	IsAtomic() bool
}

// retryApply retries the given retriable action.
func retryApply(ctx context.Context, action retrieableApplier, retryInterval time.Duration, log debugLog) error {
	var retries int
	retriable := func(_ error) bool {
		// abort after maximumRetryAttempts tries.
		if retries >= maximumRetryAttempts {
			return false
		}
		retries++
		// only retry if atomic is set
		// otherwise helm doesn't uninstall the release on failure
		return action.IsAtomic()
	}
	doer := applyDoer{
		applier: action,
		log:     log,
	}
	retrier := retry.NewIntervalRetrier(doer, retryInterval, retriable)

	retryLoopStartTime := time.Now()
	if err := retrier.Do(ctx); err != nil {
		return fmt.Errorf("helm install: %w", err)
	}
	retryLoopFinishDuration := time.Since(retryLoopStartTime)
	log.Debug(fmt.Sprintf("Helm chart %q installation finished after %q", action.ReleaseName(), retryLoopFinishDuration))
	return nil
}

// applyDoer is a helper struct to enable retrying helm actions.
type applyDoer struct {
	applier retrieableApplier
	log     debugLog
}

// Do tries to apply the action.
func (i applyDoer) Do(ctx context.Context) error {
	i.log.Debug(fmt.Sprintf("Trying to apply Helm chart %q", i.applier.ReleaseName()))
	if err := i.applier.apply(ctx); err != nil {
		i.log.Debug(fmt.Sprintf("Helm chart installation %q failed: %q", i.applier.ReleaseName(), err))
		return err
	}

	return nil
}
