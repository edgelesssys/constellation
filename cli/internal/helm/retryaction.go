/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/retry"
	"k8s.io/apimachinery/pkg/util/wait"
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
func retryApply(ctx context.Context, action retrieableApplier, log debugLog) error {
	var retries int
	retriable := func(err error) bool {
		// abort after maximumRetryAttempts tries.
		if retries >= maximumRetryAttempts {
			return false
		}
		retries++
		// only retry if atomic is set
		// otherwise helm doesn't uninstall
		// the release on failure
		if !action.IsAtomic() {
			return false
		}
		// check if error is retriable
		return wait.Interrupted(err) ||
			strings.Contains(err.Error(), "connection refused")
	}
	doer := applyDoer{
		action,
		log,
	}
	retrier := retry.NewIntervalRetrier(doer, 30*time.Second, retriable)

	retryLoopStartTime := time.Now()
	if err := retrier.Do(ctx); err != nil {
		return fmt.Errorf("helm install: %w", err)
	}
	retryLoopFinishDuration := time.Since(retryLoopStartTime)
	log.Debugf("Helm chart %q installation finished after %s", action.ReleaseName(), retryLoopFinishDuration)
	return nil
}

// applyDoer is a helper struct to enable retrying helm actions.
type applyDoer struct {
	Applier retrieableApplier
	log     debugLog
}

// Do tries to apply the action.
func (i applyDoer) Do(ctx context.Context) error {
	i.log.Debugf("Trying to apply Helm chart %s", i.Applier.ReleaseName())
	if err := i.Applier.apply(ctx); err != nil {
		i.log.Debugf("Helm chart installation %s failed: %v", i.Applier.ReleaseName(), err)
		return err
	}

	return nil
}
