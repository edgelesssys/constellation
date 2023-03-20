/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package info

import (
	"testing"

	pb "github.com/edgelesssys/constellation/v2/debugd/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)

	i := NewMap()

	assert.NotNil(i)
	assert.NotNil(i.m)
	assert.False(i.received)
}

func TestGet(t *testing.T) {
	testCases := map[string]struct {
		infosMap map[string]string
		key      string
		want     string
		wantOk   bool
		wantErr  bool
	}{
		"empty map": {
			infosMap: map[string]string{},
			key:      "key",
		},
		"key not found": {
			infosMap: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			key: "key3",
		},
		"key found": {
			infosMap: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			key:    "key2",
			want:   "value2",
			wantOk: true,
		},
		"not received": {
			infosMap: nil,
			key:      "key",
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			infos := &Map{m: tc.infosMap}
			if infos.m != nil {
				infos.received = true
			}

			got, gotOk, err := infos.Get(tc.key)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantOk, gotOk)
				assert.Equal(tc.want, got)
			}
		})
	}
}

func TestGetCopy(t *testing.T) {
	testCases := map[string]struct {
		infosMap map[string]string
		received bool
		wantMap  map[string]string
		wantErr  bool
	}{
		"empty": {
			infosMap: map[string]string{},
			received: true,
			wantMap:  map[string]string{},
		},
		"one": {
			infosMap: map[string]string{
				"key1": "value1",
			},
			received: true,
			wantMap: map[string]string{
				"key1": "value1",
			},
		},
		"multiple": {
			infosMap: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			received: true,
			wantMap: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		"not received": {
			infosMap: nil,
			received: false,
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			i := &Map{m: tc.infosMap, received: tc.received}

			gotMap, err := i.GetCopy()

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantMap, gotMap)
			}
		})
	}
}

func TestSetProto(t *testing.T) {
	testCases := map[string]struct {
		infosPB  []*pb.Info
		received bool
		wantMap  map[string]string
		wantErr  bool
	}{
		"empty": {
			infosPB: []*pb.Info{},
			wantMap: map[string]string{},
		},
		"one": {
			infosPB: []*pb.Info{
				{Key: "foo", Value: "bar"},
			},
			wantMap: map[string]string{
				"foo": "bar",
			},
		},
		"multiple": {
			infosPB: []*pb.Info{
				{Key: "foo", Value: "bar"},
				{Key: "baz", Value: "qux"},
			},
			wantMap: map[string]string{
				"foo": "bar",
				"baz": "qux",
			},
		},
		"already received": {
			infosPB:  []*pb.Info{},
			received: true,
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			i := &Map{received: tc.received}
			err := i.SetProto(tc.infosPB)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantMap, i.m)
			}
		})
	}
}

func TestTrigger(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	m := NewMap()

	var tr1, tr2, tr3 bool
	m.RegisterOnReceiveTrigger(func(*Map) { tr1 = true })
	m.RegisterOnReceiveTrigger(func(*Map) { tr2 = true })
	m.RegisterOnReceiveTrigger(func(*Map) { tr3 = true })

	err := m.SetProto([]*pb.Info{})
	require.NoError(err)

	assert.True(tr1)
	assert.True(tr2)
	assert.True(tr3)
}

func TestGetProto(t *testing.T) {
	testCases := map[string]struct {
		infosMap map[string]string
		wantPB   []*pb.Info
		wantErr  bool
	}{
		"empty": {
			infosMap: map[string]string{},
			wantPB:   []*pb.Info{},
		},
		"one": {
			infosMap: map[string]string{
				"foo": "bar",
			},
			wantPB: []*pb.Info{
				{Key: "foo", Value: "bar"},
			},
		},
		"multiple": {
			infosMap: map[string]string{
				"foo": "bar",
				"baz": "qux",
			},
			wantPB: []*pb.Info{
				{Key: "foo", Value: "bar"},
				{Key: "baz", Value: "qux"},
			},
		},
		"not received": {
			infosMap: nil,
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			i := &Map{m: tc.infosMap}
			if i.m != nil {
				i.received = true
			}

			gotPB, err := i.GetProto()

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(len(tc.wantPB), len(gotPB))
			}
		})
	}
}

func TestConcurrency(_ *testing.T) {
	i := NewMap()

	get := func() {
		_, _, _ = i.Get("foo")
	}

	getCopy := func() {
		_, _ = i.GetCopy()
	}

	setProto := func() {
		_ = i.SetProto([]*pb.Info{{Key: "foo", Value: "bar"}})
	}

	getProto := func() {
		_, _ = i.GetProto()
	}

	received := func() {
		_ = i.Received()
	}

	go get()
	go get()
	go get()
	go get()
	go getCopy()
	go getCopy()
	go getCopy()
	go getCopy()
	go setProto()
	go setProto()
	go setProto()
	go setProto()
	go getProto()
	go getProto()
	go getProto()
	go getProto()
	go received()
	go received()
	go received()
	go received()
	// TODO(katexochen): fix this test, wait for routines to finish
}
