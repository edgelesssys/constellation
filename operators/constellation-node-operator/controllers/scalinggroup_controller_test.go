package controllers

import (
	"context"
	"sync"
)

type fakeScalingGroupUpdater struct {
	sync.RWMutex
	scalingGroupImage map[string]string
}

func newFakeScalingGroupUpdater() *fakeScalingGroupUpdater {
	return &fakeScalingGroupUpdater{
		scalingGroupImage: make(map[string]string),
	}
}

func (u *fakeScalingGroupUpdater) GetScalingGroupImage(ctx context.Context, scalingGroupID string) (string, error) {
	u.RLock()
	defer u.RUnlock()
	return u.scalingGroupImage[scalingGroupID], nil
}

func (u *fakeScalingGroupUpdater) SetScalingGroupImage(ctx context.Context, scalingGroupID, imageURI string) error {
	u.Lock()
	defer u.Unlock()
	u.scalingGroupImage[scalingGroupID] = imageURI
	return nil
}
