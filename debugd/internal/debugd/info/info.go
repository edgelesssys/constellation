/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package info

import (
	"errors"
	"sync"

	servicepb "github.com/edgelesssys/constellation/v2/debugd/service"
)

// Map is a thread-safe map of info, with triggers that are run
// when the map is set.
type Map struct {
	m                map[string]string
	received         bool
	mux              sync.RWMutex
	onReceiveTrigger []func(*Map)
}

// NewMap creates a new Map object.
func NewMap() *Map {
	return &Map{
		m: make(map[string]string),
	}
}

// Get returns the value of the info with the given key.
func (i *Map) Get(key string) (string, bool, error) {
	i.mux.RLock()
	defer i.mux.RUnlock()

	if !i.received {
		return "", false, errors.New("info not set yet")
	}

	value, ok := i.m[key]
	return value, ok, nil
}

// GetCopy returns a copy of the info map.
func (i *Map) GetCopy() (map[string]string, error) {
	i.mux.RLock()
	defer i.mux.RUnlock()

	if !i.received {
		return nil, errors.New("info not set yet")
	}

	m := make(map[string]string)
	for k, v := range i.m {
		m[k] = v
	}

	return m, nil
}

// SetProto sets the info map to the given infos proto slice.
// It returns an error if the info map has already been set.
// Registered triggers are run after the info map is set.
func (i *Map) SetProto(infos []*servicepb.Info) error {
	i.mux.Lock()
	defer i.mux.Unlock()

	if i.received {
		return errors.New("info already set")
	}

	infoMap := make(map[string]string)
	for _, info := range infos {
		infoMap[info.Key] = info.Value
	}

	i.m = infoMap
	i.received = true

	for _, trigger := range i.onReceiveTrigger {
		trigger(i)
	}

	return nil
}

// RegisterOnReceiveTrigger registers a function that is called when the info map is set.
// The function mustn't block or be long-running.
func (i *Map) RegisterOnReceiveTrigger(f func(*Map)) {
	i.mux.Lock()
	defer i.mux.Unlock()

	if i.received {
		f(i)
		return
	}

	i.onReceiveTrigger = append(i.onReceiveTrigger, f)
}

// GetProto returns the info map as a slice of info proto.
func (i *Map) GetProto() ([]*servicepb.Info, error) {
	i.mux.RLock()
	defer i.mux.RUnlock()

	if !i.received {
		return nil, errors.New("info not set yet")
	}

	var infos []*servicepb.Info
	for key, value := range i.m {
		infos = append(infos, &servicepb.Info{Key: key, Value: value})
	}
	return infos, nil
}
